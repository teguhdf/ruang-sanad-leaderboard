package services

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/maulanashalihin/laju-go/app/models"
	"github.com/maulanashalihin/laju-go/app/queries"
)

// RequestForgotCheckout releases a mistakenly-open session as incomplete
// evidence and submits the intended checkout time for SPV approval.
func (s *EmployeeAttendanceService) RequestForgotCheckout(userID int64, input models.ForgotCheckoutInput) error {
	ctx := context.Background()
	identity, err := s.querier.GetEmployeeAttendanceIdentity(ctx, userID)
	if err != nil {
		return err
	}

	active, err := s.querier.GetActiveAttendanceSession(ctx, identity.MemberID)
	if errors.Is(err, queries.ErrActiveAttendanceSessionNotFound) {
		return ErrNoActiveAttendanceSession
	}
	if err != nil {
		return err
	}

	reason := strings.TrimSpace(input.Reason)
	if len(reason) < 5 {
		return &EmployeeAttendanceValidationError{Message: "Alasan lupa checkout minimal 5 karakter."}
	}
	endClock, err := time.Parse("15:04", strings.TrimSpace(input.EndTime))
	if err != nil {
		return &EmployeeAttendanceValidationError{Message: "Jam selesai yang dimaksud tidak valid."}
	}
	workDate, err := time.ParseInLocation("2006-01-02", active.WorkDate, s.location)
	if err != nil {
		return &EmployeeAttendanceValidationError{Message: "Tanggal sesi aktif tidak valid."}
	}
	requestedCheckOut := time.Date(
		workDate.Year(), workDate.Month(), workDate.Day(),
		endClock.Hour(), endClock.Minute(), 0, 0, s.location,
	)
	if !requestedCheckOut.After(active.CheckInAt) {
		return &EmployeeAttendanceValidationError{Message: "Jam selesai harus setelah jam mulai sesi aktif."}
	}
	now := time.Now().In(s.location)
	if requestedCheckOut.After(now) {
		return &EmployeeAttendanceValidationError{Message: "Jam selesai tidak boleh berada di masa depan."}
	}
	if requestedCheckOut.Sub(active.CheckInAt) > 16*time.Hour {
		return &EmployeeAttendanceValidationError{Message: "Durasi sesi terlalu panjang. Maksimal 16 jam."}
	}

	return s.querier.CreateForgotCheckoutAdjustment(
		ctx,
		identity.OwnerUserID,
		identity.MemberID,
		userID,
		active.ID,
		active.WorkDate,
		active.LocationMode,
		active.CheckInAt,
		requestedCheckOut,
		reason,
	)
}
