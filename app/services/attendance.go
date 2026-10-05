package services

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/maulanashalihin/laju-go/app/models"
	"github.com/maulanashalihin/laju-go/app/queries"
)

var ErrAttendanceProfileNotFound = errors.New("attendance profile not found")

type AttendanceValidationError struct {
	Message string
}

func (e *AttendanceValidationError) Error() string {
	return e.Message
}

type AttendanceService struct {
	querier *queries.Querier
}

func NewAttendanceService(querier *queries.Querier) *AttendanceService {
	return &AttendanceService{querier: querier}
}

func (s *AttendanceService) GetFoundation(ownerUserID int64) (*models.AttendanceFoundation, error) {
	return s.GetFoundationForWeek(ownerUserID, "")
}

func (s *AttendanceService) GetFoundationForWeek(ownerUserID int64, requestedWeek string) (*models.AttendanceFoundation, error) {
	ctx := context.Background()
	if err := s.querier.EnsureAttendanceFoundation(ctx, ownerUserID); err != nil {
		return nil, err
	}

	policy, err := s.querier.GetAttendancePolicy(ctx, ownerUserID)
	if err != nil {
		return nil, err
	}
	profiles, err := s.querier.ListMemberAttendanceProfiles(ctx, ownerUserID)
	if err != nil {
		return nil, err
	}
	shifts, err := s.querier.ListShiftTemplates(ctx, ownerUserID)
	if err != nil {
		return nil, err
	}

	weekStart, weekEnd, err := attendanceWeekRange(requestedWeek)
	if err != nil {
		return nil, err
	}
	weeks := make([]models.MemberAttendanceWeek, 0, len(profiles))
	for _, profile := range profiles {
		workDays, err := s.querier.GetMemberWorkDaysForLeave(ctx, ownerUserID, profile.MemberID)
		if err != nil {
			return nil, err
		}
		rosters, err := s.querier.ListRostersForMemberRange(ctx, ownerUserID, profile.MemberID, weekStart.Format("2006-01-02"), weekEnd.Format("2006-01-02"))
		if err != nil {
			return nil, err
		}
		leavePeriods, err := s.querier.ListApprovedLeavePeriodsForMemberRange(ctx, ownerUserID, profile.MemberID, weekStart.Format("2006-01-02"), weekEnd.Format("2006-01-02"))
		if err != nil {
			return nil, err
		}
		weeks = append(weeks, models.MemberAttendanceWeek{
			MemberID:   profile.MemberID,
			MemberName: profile.MemberName,
			Division:   profile.Division,
			WeekStart:  weekStart.Format("2006-01-02"),
			WeekEnd:    weekEnd.Format("2006-01-02"),
			Plan:       buildAttendanceWeekPlan(weekStart, weekEnd, policy, workDays, profile.WorkMode, rosters, leavePeriods),
		})
	}

	return &models.AttendanceFoundation{
		Policy:          policy,
		Profiles:        profiles,
		ShiftTemplates:  shifts,
		RosterWeekStart: weekStart.Format("2006-01-02"),
		RosterWeekEnd:   weekEnd.Format("2006-01-02"),
		RosterWeeks:     weeks,
	}, nil
}

func (s *AttendanceService) UpdateProfile(ownerUserID, memberID int64, input models.MemberAttendanceProfileInput) error {
	input.WorkMode = strings.ToLower(strings.TrimSpace(input.WorkMode))
	input.ScheduleType = strings.ToLower(strings.TrimSpace(input.ScheduleType))

	allowedWorkModes := map[string]bool{
		"onsite": true,
		"remote": true,
		"hybrid": true,
	}
	if !allowedWorkModes[input.WorkMode] {
		return &AttendanceValidationError{Message: "Mode kerja tidak valid."}
	}

	allowedScheduleTypes := map[string]bool{
		"shift":    true,
		"flexible": true,
	}
	if !allowedScheduleTypes[input.ScheduleType] {
		return &AttendanceValidationError{Message: "Tipe jadwal tidak valid."}
	}

	ctx := context.Background()
	if err := s.querier.EnsureAttendanceFoundation(ctx, ownerUserID); err != nil {
		return err
	}

	if err := s.querier.UpdateMemberAttendanceProfile(ctx, ownerUserID, memberID, input); err != nil {
		if errors.Is(err, queries.ErrAttendanceProfileNotFound) {
			return ErrAttendanceProfileNotFound
		}
		return err
	}
	return nil
}

func (s *AttendanceService) SaveRosterDay(ownerUserID, memberID int64, workDate string, input models.RosterDayInput) error {
	ctx := context.Background()
	if err := s.querier.EnsureAttendanceFoundation(ctx, ownerUserID); err != nil {
		return err
	}
	date, err := time.ParseInLocation("2006-01-02", strings.TrimSpace(workDate), time.FixedZone("WIB", 7*60*60))
	if err != nil {
		return &AttendanceValidationError{Message: "Tanggal roster tidak valid."}
	}

	leavePeriods, err := s.querier.ListApprovedLeavePeriodsForMemberRange(ctx, ownerUserID, memberID, workDate, workDate)
	if err != nil {
		return err
	}
	if len(leavePeriods) > 0 {
		return &AttendanceValidationError{Message: "Tanggal ini sudah memiliki cuti approved dan tidak dapat diubah lewat roster."}
	}

	input.DayType = normalizeAttendanceDayType(input.DayType)
	if input.DayType == "leave" {
		return &AttendanceValidationError{Message: "Status cuti hanya berasal dari pengajuan cuti yang approved."}
	}
	input.WorkMode = strings.ToLower(strings.TrimSpace(input.WorkMode))
	if input.WorkMode != "onsite" && input.WorkMode != "remote" && input.WorkMode != "hybrid" {
		return &AttendanceValidationError{Message: "Mode kerja roster tidak valid."}
	}

	if input.DayType != "work" {
		input.RequiredMinutes = 0
		input.ShiftTemplateID = nil
	} else {
		if input.ShiftTemplateID != nil {
			shifts, err := s.querier.ListShiftTemplates(ctx, ownerUserID)
			if err != nil {
				return err
			}
			validShift := false
			for _, shift := range shifts {
				if shift.ID == *input.ShiftTemplateID && shift.Active {
					validShift = true
					if input.RequiredMinutes <= 0 {
						input.RequiredMinutes = shift.RequiredMinutes
					}
					break
				}
			}
			if !validShift {
				return &AttendanceValidationError{Message: "Template shift tidak valid."}
			}
		}
		if input.RequiredMinutes <= 0 {
			policy, err := s.querier.GetAttendancePolicy(ctx, ownerUserID)
			if err != nil {
				return err
			}
			input.RequiredMinutes = defaultAttendanceMinutes(date.Weekday(), policy)
		}
		if input.RequiredMinutes > 960 {
			return &AttendanceValidationError{Message: "Target kerja roster maksimal 16 jam per hari."}
		}
	}

	workDays, err := s.querier.GetMemberWorkDaysForLeave(ctx, ownerUserID, memberID)
	if err != nil {
		return err
	}
	policy, err := s.querier.GetAttendancePolicy(ctx, ownerUserID)
	if err != nil {
		return err
	}
	workDaySet := make(map[string]bool, len(workDays))
	for _, day := range workDays {
		workDaySet[strings.ToLower(strings.TrimSpace(day))] = true
	}
	candidate := models.AttendanceRosterDay{
		WorkDate:        date.Format("2006-01-02"),
		DayType:         input.DayType,
		ShiftTemplateID: input.ShiftTemplateID,
		WorkMode:        input.WorkMode,
		RequiredMinutes: input.RequiredMinutes,
	}
	defaultWorkMode := "onsite"
	profiles, err := s.querier.ListMemberAttendanceProfiles(ctx, ownerUserID)
	if err != nil {
		return err
	}
	for _, profile := range profiles {
		if profile.MemberID == memberID {
			defaultWorkMode = profile.WorkMode
			break
		}
	}
	if rosterMatchesAttendanceDefault(candidate, date, policy, workDaySet[attendanceWeekdayKey(date.Weekday())], defaultWorkMode) {
		return s.querier.DeleteRosterDay(ctx, ownerUserID, memberID, date.Format("2006-01-02"))
	}

	return s.querier.UpsertRosterDay(ctx, ownerUserID, memberID, date.Format("2006-01-02"), input)
}

func (s *AttendanceService) ResetRosterDay(ownerUserID, memberID int64, workDate string) error {
	if _, err := time.Parse("2006-01-02", strings.TrimSpace(workDate)); err != nil {
		return &AttendanceValidationError{Message: "Tanggal roster tidak valid."}
	}
	return s.querier.DeleteRosterDay(context.Background(), ownerUserID, memberID, workDate)
}

func attendanceWeekRange(requested string) (time.Time, time.Time, error) {
	location := time.FixedZone("WIB", 7*60*60)
	var base time.Time
	var err error
	if strings.TrimSpace(requested) == "" {
		base = time.Now().In(location)
	} else {
		base, err = time.ParseInLocation("2006-01-02", strings.TrimSpace(requested), location)
		if err != nil {
			return time.Time{}, time.Time{}, &AttendanceValidationError{Message: "Minggu roster tidak valid."}
		}
	}
	weekday := int(base.Weekday())
	if weekday == 0 {
		weekday = 7
	}
	start := time.Date(base.Year(), base.Month(), base.Day(), 0, 0, 0, 0, location).AddDate(0, 0, -(weekday - 1))
	return start, start.AddDate(0, 0, 6), nil
}
