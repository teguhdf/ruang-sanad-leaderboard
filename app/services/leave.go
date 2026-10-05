package services

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/maulanashalihin/laju-go/app/models"
	"github.com/maulanashalihin/laju-go/app/queries"
)

func (s *EmployeeAttendanceService) RequestLeave(userID int64, input models.LeaveRequestInput) error {
	ctx := context.Background()
	identity, err := s.querier.GetEmployeeAttendanceIdentity(ctx, userID)
	if err != nil {
		return err
	}

	reason := strings.TrimSpace(input.Reason)
	if len(reason) < 5 {
		return &EmployeeAttendanceValidationError{Message: "Alasan cuti minimal 5 karakter."}
	}

	start, err := time.ParseInLocation("2006-01-02", strings.TrimSpace(input.StartDate), s.location)
	if err != nil {
		return &EmployeeAttendanceValidationError{Message: "Tanggal mulai cuti tidak valid."}
	}
	end, err := time.ParseInLocation("2006-01-02", strings.TrimSpace(input.EndDate), s.location)
	if err != nil {
		return &EmployeeAttendanceValidationError{Message: "Tanggal selesai cuti tidak valid."}
	}
	if end.Before(start) {
		return &EmployeeAttendanceValidationError{Message: "Tanggal selesai cuti harus sama atau setelah tanggal mulai."}
	}
	if start.Year() != end.Year() {
		return &EmployeeAttendanceValidationError{Message: "Cuti lintas tahun harus dibuat sebagai dua pengajuan terpisah."}
	}
	if end.Sub(start) > 31*24*time.Hour {
		return &EmployeeAttendanceValidationError{Message: "Rentang satu pengajuan cuti maksimal 32 hari kalender."}
	}

	now := time.Now().In(s.location)
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, s.location)
	if start.Before(today) {
		return &EmployeeAttendanceValidationError{Message: "Pengajuan cuti tidak boleh dimulai pada tanggal yang sudah lewat."}
	}

	chargeableDays, err := s.countLeaveChargeableDays(ctx, identity.OwnerUserID, identity.MemberID, start, end)
	if err != nil {
		return err
	}
	if chargeableDays == 0 {
		return &EmployeeAttendanceValidationError{Message: "Rentang cuti tidak memiliki hari kerja yang dapat dipotong."}
	}

	balance, err := s.querier.GetLeaveBalance(ctx, identity.OwnerUserID, identity.MemberID, start.Year())
	if err != nil {
		return err
	}
	if chargeableDays > balance.AvailableDays {
		return &EmployeeAttendanceValidationError{Message: "Sisa cuti tahunan tidak mencukupi untuk pengajuan ini."}
	}

	return s.querier.CreateLeaveRequest(
		ctx,
		identity.OwnerUserID,
		identity.MemberID,
		userID,
		start.Year(),
		start.Format("2006-01-02"),
		end.Format("2006-01-02"),
		chargeableDays,
		reason,
	)
}

func (s *EmployeeAttendanceService) CancelLeave(userID, requestID int64) error {
	ctx := context.Background()
	identity, err := s.querier.GetEmployeeAttendanceIdentity(ctx, userID)
	if err != nil {
		return err
	}
	return s.querier.CancelLeaveRequest(ctx, identity.MemberID, requestID)
}

func (s *EmployeeAttendanceService) GetEmployeeLeave(userID int64, year int) (models.LeaveBalance, []models.LeaveRequest, error) {
	ctx := context.Background()
	identity, err := s.querier.GetEmployeeAttendanceIdentity(ctx, userID)
	if err != nil {
		return models.LeaveBalance{}, nil, err
	}
	if year == 0 {
		year = time.Now().In(s.location).Year()
	}
	balance, err := s.querier.GetLeaveBalance(ctx, identity.OwnerUserID, identity.MemberID, year)
	if err != nil {
		return models.LeaveBalance{}, nil, err
	}
	requests, err := s.querier.ListLeaveRequestsByMember(ctx, identity.MemberID, 10)
	if err != nil {
		return models.LeaveBalance{}, nil, err
	}
	return balance, requests, nil
}

func (s *EmployeeAttendanceService) SetLeaveEntitlement(ownerUserID, memberID int64, input models.LeaveEntitlementInput) error {
	currentYear := time.Now().In(s.location).Year()
	if input.Year == 0 {
		input.Year = currentYear
	}
	if input.Year < currentYear-1 || input.Year > currentYear+1 {
		return &EmployeeAttendanceValidationError{Message: "Tahun jatah cuti harus tahun lalu, tahun berjalan, atau tahun depan."}
	}
	if input.EntitledDays < 0 || input.EntitledDays > 365 {
		return &EmployeeAttendanceValidationError{Message: "Jatah cuti tahunan harus antara 0 dan 365 hari."}
	}
	input.Note = strings.TrimSpace(input.Note)
	return s.querier.UpsertLeaveEntitlement(context.Background(), ownerUserID, memberID, input)
}

func (s *EmployeeAttendanceService) ListLeaveAdministration(ownerUserID int64, year int) ([]models.LeaveBalance, []models.LeaveRequest, error) {
	if year == 0 {
		year = time.Now().In(s.location).Year()
	}
	ctx := context.Background()
	balances, err := s.querier.ListLeaveBalancesByOwner(ctx, ownerUserID, year)
	if err != nil {
		return nil, nil, err
	}
	pending, err := s.querier.ListPendingLeaveRequestsByOwner(ctx, ownerUserID)
	if err != nil {
		return nil, nil, err
	}
	return balances, pending, nil
}

func (s *EmployeeAttendanceService) ReviewLeave(ownerUserID, requestID, reviewerUserID int64, input models.ReviewRequestInput) error {
	approve, err := reviewDecision(input.Action)
	if err != nil {
		return err
	}
	return s.querier.ReviewLeaveRequest(
		context.Background(),
		ownerUserID,
		requestID,
		reviewerUserID,
		approve,
		strings.TrimSpace(input.ReviewNote),
	)
}

func (s *EmployeeAttendanceService) countLeaveChargeableDays(ctx context.Context, ownerUserID, memberID int64, start, end time.Time) (int, error) {
	workDays, err := s.querier.GetMemberWorkDaysForLeave(ctx, ownerUserID, memberID)
	if err != nil {
		return 0, err
	}
	workDaySet := make(map[string]bool, len(workDays))
	for _, day := range workDays {
		workDaySet[strings.ToLower(strings.TrimSpace(day))] = true
	}

	count := 0
	for day := start; !day.After(end); day = day.AddDate(0, 0, 1) {
		date := day.Format("2006-01-02")
		chargeable, found, err := s.querier.GetRosterChargeability(ctx, ownerUserID, memberID, date)
		if err != nil {
			return 0, err
		}
		if found {
			if chargeable {
				count++
			}
			continue
		}
		if workDaySet[leaveWeekdayKey(day.Weekday())] {
			count++
		}
	}
	return count, nil
}

func leaveWeekdayKey(day time.Weekday) string {
	switch day {
	case time.Monday:
		return "mon"
	case time.Tuesday:
		return "tue"
	case time.Wednesday:
		return "wed"
	case time.Thursday:
		return "thu"
	case time.Friday:
		return "fri"
	case time.Saturday:
		return "sat"
	default:
		return "sun"
	}
}

func leaveErrorMessage(err error) string {
	switch {
	case errors.Is(err, queries.ErrLeaveRequestConflict):
		return "Tanggal cuti bertabrakan dengan pengajuan cuti lain yang masih pending atau sudah disetujui."
	case errors.Is(err, queries.ErrLeaveInsufficientBalance):
		return "Sisa cuti berubah dan tidak lagi mencukupi. Muat ulang saldo cuti lalu periksa pengajuan."
	case errors.Is(err, queries.ErrLeaveRequestNotFound):
		return "Pengajuan cuti tidak ditemukan atau sudah diproses."
	default:
		return ""
	}
}
