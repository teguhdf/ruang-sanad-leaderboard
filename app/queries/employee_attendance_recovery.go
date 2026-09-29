package queries

import (
	"context"
	"errors"
	"time"
)

// CreateForgotCheckoutAdjustment converts an active session into immutable
// incomplete evidence and creates a pending replacement request in one tx.
// This frees the employee to start another session without counting guessed time.
func (q *Querier) CreateForgotCheckoutAdjustment(
	ctx context.Context,
	ownerUserID, memberID, userID, sessionID int64,
	workDate, locationMode string,
	checkIn, requestedCheckOut time.Time,
	reason string,
) error {
	beginner, ok := q.db.(attendanceTxBeginner)
	if !ok {
		return errors.New("database does not support transactions")
	}
	tx, err := beginner.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	result, err := tx.ExecContext(ctx, `
		UPDATE attendance_sessions
		SET status = 'incomplete', updated_at = CURRENT_TIMESTAMP
		WHERE id = ? AND owner_user_id = ? AND member_id = ? AND user_id = ? AND status = 'active'
	`, sessionID, ownerUserID, memberID, userID)
	if err != nil {
		return err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected != 1 {
		return ErrActiveAttendanceSessionNotFound
	}

	if _, err := tx.ExecContext(ctx, `
		INSERT INTO attendance_adjustment_requests (
			owner_user_id, member_id, requested_by_user_id, related_session_id,
			work_date, location_mode, requested_check_in, requested_check_out, reason
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, ownerUserID, memberID, userID, sessionID, workDate, locationMode, checkIn.UTC(), requestedCheckOut.UTC(), reason); err != nil {
		return err
	}

	return tx.Commit()
}
