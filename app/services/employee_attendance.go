package services

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/maulanashalihin/laju-go/app/models"
	"github.com/maulanashalihin/laju-go/app/queries"
)

var (
	ErrEmployeeAlreadyLinked = errors.New("employee account already linked")
	ErrEmployeeInviteConflict = errors.New("email belongs to a non-employee account")
	ErrAttendanceSessionActive = errors.New("attendance session already active")
	ErrNoActiveAttendanceSession = errors.New("no active attendance session")
)

type EmployeeAttendanceValidationError struct {
	Message string
}

func (e *EmployeeAttendanceValidationError) Error() string { return e.Message }

type EmployeeAttendanceService struct {
	querier *queries.Querier
	mailer *MailerService
	location *time.Location
}

func NewEmployeeAttendanceService(querier *queries.Querier, mailer *MailerService) *EmployeeAttendanceService {
	return &EmployeeAttendanceService{
		querier: querier,
		mailer: mailer,
		location: time.FixedZone("WIB", 7*60*60),
	}
}

func (s *EmployeeAttendanceService) InviteEmployee(ownerUserID, memberID int64, email string) error {
	ctx := context.Background()
	email = strings.ToLower(strings.TrimSpace(email))
	if email == "" || !strings.Contains(email, "@") {
		return &EmployeeAttendanceValidationError{Message: "Email employee tidak valid."}
	}

	name, linkedUserID, linkedEmail, err := s.querier.GetAttendanceInviteTarget(ctx, ownerUserID, memberID)
	if err != nil {
		return err
	}

	if linkedUserID != nil {
		if !strings.EqualFold(linkedEmail, email) {
			return ErrEmployeeAlreadyLinked
		}
		return s.mailer.SendPasswordResetEmail(ctx, linkedEmail, *linkedUserID)
	}

	user, err := s.querier.GetUserByEmail(ctx, email)
	if err != nil && !errors.Is(err, queries.ErrUserNotFound) {
		return err
	}
	if errors.Is(err, queries.ErrUserNotFound) {
		user = &models.User{
			Email: email,
			Name: name,
			Password: sql.NullString{},
			Role: models.RoleEmployee,
			EmailVerified: false,
		}
		if err := s.querier.CreateUser(ctx, user); err != nil {
			return err
		}
	} else if user.Role != models.RoleEmployee {
		return ErrEmployeeInviteConflict
	}

	if err := s.querier.LinkEmployeeUser(ctx, ownerUserID, memberID, user.ID); err != nil {
		return err
	}

	return s.mailer.SendPasswordResetEmail(ctx, email, user.ID)
}

func (s *EmployeeAttendanceService) GetMyAttendance(userID int64) (*models.MyAttendanceSummary, error) {
	ctx := context.Background()
	identity, err := s.querier.GetEmployeeAttendanceIdentity(ctx, userID)
	if err != nil {
		return nil, err
	}

	now := time.Now().In(s.location)
	today := now.Format("2006-01-02")
	if err := s.querier.MarkStaleActiveSessionsIncomplete(ctx, identity.MemberID, today); err != nil {
		return nil, err
	}

	var active *models.AttendanceSession
	active, err = s.querier.GetActiveAttendanceSession(ctx, identity.MemberID)
	if err != nil && !errors.Is(err, queries.ErrActiveAttendanceSessionNotFound) {
		return nil, err
	}
	if errors.Is(err, queries.ErrActiveAttendanceSessionNotFound) {
		active = nil
	}

	weekday := int(now.Weekday())
	if weekday == 0 {
		weekday = 7
	}
	weekStartTime := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, s.location).AddDate(0, 0, -(weekday - 1))
	weekEndTime := weekStartTime.AddDate(0, 0, 6)
	weekStart := weekStartTime.Format("2006-01-02")
	weekEnd := weekEndTime.Format("2006-01-02")

	sessions, err := s.querier.ListAttendanceSessions(ctx, identity.MemberID, weekStart, weekEnd)
	if err != nil {
		return nil, err
	}
	todaySessions := make([]models.AttendanceSession, 0)
	weeklyMinutes := 0
	todayMinutes := 0
	for _, session := range sessions {
		if session.Status == "completed" {
			weeklyMinutes += session.DurationMinutes
			if session.WorkDate == today {
				todayMinutes += session.DurationMinutes
			}
		}
		if session.WorkDate == today {
			todaySessions = append(todaySessions, session)
		}
	}

	adjustments, err := s.querier.ListAttendanceAdjustmentsByMember(ctx, identity.MemberID, 10)
	if err != nil {
		return nil, err
	}
	overtime, err := s.querier.ListOvertimeRequestsByMember(ctx, identity.MemberID, 10)
	if err != nil {
		return nil, err
	}

	return &models.MyAttendanceSummary{
		Identity: identity,
		Today: today,
		ActiveSession: active,
		TodaySessions: todaySessions,
		TodayMinutes: todayMinutes,
		WeekStart: weekStart,
		WeekEnd: weekEnd,
		WeeklyMinutes: weeklyMinutes,
		WeeklyTargetMinutes: identity.WeeklyTargetMinutes,
		WeeklyBalanceMinutes: weeklyMinutes - identity.WeeklyTargetMinutes,
		Adjustments: adjustments,
		OvertimeRequests: overtime,
	}, nil
}

func (s *EmployeeAttendanceService) StartWork(userID int64, input models.StartWorkInput) error {
	ctx := context.Background()
	identity, err := s.querier.GetEmployeeAttendanceIdentity(ctx, userID)
	if err != nil {
		return err
	}
	locationMode := strings.ToLower(strings.TrimSpace(input.LocationMode))
	if locationMode != "office" && locationMode != "remote" {
		return &EmployeeAttendanceValidationError{Message: "Pilih lokasi kerja Kantor atau Remote."}
	}

	now := time.Now().In(s.location)
	today := now.Format("2006-01-02")
	if err := s.querier.MarkStaleActiveSessionsIncomplete(ctx, identity.MemberID, today); err != nil {
		return err
	}
	if _, err := s.querier.GetActiveAttendanceSession(ctx, identity.MemberID); err == nil {
		return ErrAttendanceSessionActive
	} else if !errors.Is(err, queries.ErrActiveAttendanceSessionNotFound) {
		return err
	}

	_, err = s.querier.CreateAttendanceSession(ctx, identity.OwnerUserID, identity.MemberID, userID, today, locationMode, now, "app")
	return err
}

func (s *EmployeeAttendanceService) StopWork(userID int64) error {
	ctx := context.Background()
	identity, err := s.querier.GetEmployeeAttendanceIdentity(ctx, userID)
	if err != nil {
		return err
	}
	now := time.Now().In(s.location)
	today := now.Format("2006-01-02")
	if err := s.querier.MarkStaleActiveSessionsIncomplete(ctx, identity.MemberID, today); err != nil {
		return err
	}
	active, err := s.querier.GetActiveAttendanceSession(ctx, identity.MemberID)
	if errors.Is(err, queries.ErrActiveAttendanceSessionNotFound) {
		return ErrNoActiveAttendanceSession
	}
	if err != nil {
		return err
	}
	if now.Before(active.CheckInAt) {
		return &EmployeeAttendanceValidationError{Message: "Waktu server tidak valid untuk menutup sesi."}
	}
	return s.querier.CompleteAttendanceSession(ctx, active.ID, identity.MemberID, now)
}

func (s *EmployeeAttendanceService) RequestAdjustment(userID int64, input models.AttendanceAdjustmentInput) error {
	ctx := context.Background()
	identity, err := s.querier.GetEmployeeAttendanceIdentity(ctx, userID)
	if err != nil {
		return err
	}

	input.Reason = strings.TrimSpace(input.Reason)
	if len(input.Reason) < 5 {
		return &EmployeeAttendanceValidationError{Message: "Alasan koreksi minimal 5 karakter."}
	}
	workDate, err := time.ParseInLocation("2006-01-02", strings.TrimSpace(input.WorkDate), s.location)
	if err != nil {
		return &EmployeeAttendanceValidationError{Message: "Tanggal koreksi tidak valid."}
	}
	today := time.Now().In(s.location)
	if workDate.After(time.Date(today.Year(), today.Month(), today.Day(), 23, 59, 59, 0, s.location)) {
		return &EmployeeAttendanceValidationError{Message: "Koreksi absensi tidak boleh untuk tanggal masa depan."}
	}
	startClock, err := time.Parse("15:04", strings.TrimSpace(input.StartTime))
	if err != nil {
		return &EmployeeAttendanceValidationError{Message: "Jam mulai tidak valid."}
	}
	endClock, err := time.Parse("15:04", strings.TrimSpace(input.EndTime))
	if err != nil {
		return &EmployeeAttendanceValidationError{Message: "Jam selesai tidak valid."}
	}
	checkIn := time.Date(workDate.Year(), workDate.Month(), workDate.Day(), startClock.Hour(), startClock.Minute(), 0, 0, s.location)
	checkOut := time.Date(workDate.Year(), workDate.Month(), workDate.Day(), endClock.Hour(), endClock.Minute(), 0, 0, s.location)
	if !checkOut.After(checkIn) {
		return &EmployeeAttendanceValidationError{Message: "Jam selesai harus setelah jam mulai."}
	}
	if checkOut.Sub(checkIn) > 16*time.Hour {
		return &EmployeeAttendanceValidationError{Message: "Durasi koreksi terlalu panjang. Maksimal 16 jam per sesi."}
	}

	return s.querier.CreateAttendanceAdjustmentRequest(ctx, identity.OwnerUserID, identity.MemberID, userID, input.WorkDate, checkIn, checkOut, input.Reason)
}

func (s *EmployeeAttendanceService) RequestOvertime(userID int64, input models.OvertimeRequestInput) error {
	ctx := context.Background()
	identity, err := s.querier.GetEmployeeAttendanceIdentity(ctx, userID)
	if err != nil {
		return err
	}
	input.Reason = strings.TrimSpace(input.Reason)
	if _, err := time.ParseInLocation("2006-01-02", strings.TrimSpace(input.WorkDate), s.location); err != nil {
		return &EmployeeAttendanceValidationError{Message: "Tanggal lembur tidak valid."}
	}
	if input.RequestedMinutes < 30 || input.RequestedMinutes > 720 {
		return &EmployeeAttendanceValidationError{Message: "Durasi lembur harus antara 30 menit dan 12 jam."}
	}
	if len(input.Reason) < 5 {
		return &EmployeeAttendanceValidationError{Message: "Alasan lembur minimal 5 karakter."}
	}
	return s.querier.CreateOvertimeRequest(ctx, identity.OwnerUserID, identity.MemberID, userID, input.WorkDate, input.RequestedMinutes, input.Reason)
}

func (s *EmployeeAttendanceService) ListPendingReviews(ownerUserID int64) ([]models.AttendanceAdjustmentRequest, []models.OvertimeRequest, error) {
	ctx := context.Background()
	adjustments, err := s.querier.ListPendingAttendanceAdjustmentsByOwner(ctx, ownerUserID)
	if err != nil {
		return nil, nil, err
	}
	overtime, err := s.querier.ListPendingOvertimeRequestsByOwner(ctx, ownerUserID)
	if err != nil {
		return nil, nil, err
	}
	return adjustments, overtime, nil
}

func (s *EmployeeAttendanceService) ReviewAdjustment(ownerUserID, requestID, reviewerUserID int64, input models.ReviewRequestInput) error {
	approve, err := reviewDecision(input.Action)
	if err != nil {
		return err
	}
	return s.querier.ReviewAttendanceAdjustment(context.Background(), ownerUserID, requestID, reviewerUserID, approve, strings.TrimSpace(input.ReviewNote))
}

func (s *EmployeeAttendanceService) ReviewOvertime(ownerUserID, requestID, reviewerUserID int64, input models.ReviewRequestInput) error {
	approve, err := reviewDecision(input.Action)
	if err != nil {
		return err
	}
	return s.querier.ReviewOvertimeRequest(context.Background(), ownerUserID, requestID, reviewerUserID, approve, strings.TrimSpace(input.ReviewNote))
}

func reviewDecision(action string) (bool, error) {
	switch strings.ToLower(strings.TrimSpace(action)) {
	case "approve":
		return true, nil
	case "reject":
		return false, nil
	default:
		return false, &EmployeeAttendanceValidationError{Message: fmt.Sprintf("Aksi review %q tidak valid.", action)}
	}
}
