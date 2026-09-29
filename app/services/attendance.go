package services

import (
	"context"
	"errors"
	"strings"

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

	return &models.AttendanceFoundation{
		Policy:         policy,
		Profiles:       profiles,
		ShiftTemplates: shifts,
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

	if input.WorkMode == "remote" && input.ScheduleType == "shift" {
		// Remote may still use a shift, so this is allowed intentionally.
	}

	if err := s.querier.UpdateMemberAttendanceProfile(context.Background(), ownerUserID, memberID, input); err != nil {
		if errors.Is(err, queries.ErrAttendanceProfileNotFound) {
			return ErrAttendanceProfileNotFound
		}
		return err
	}
	return nil
}
