package queries

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/maulanashalihin/laju-go/app/models"
)

var (
	ErrEmployeeProfileNotFound = errors.New("employee attendance profile not found")
	ErrActiveAttendanceSessionNotFound = errors.New("active attendance session not found")
	ErrAdjustmentRequestNotFound = errors.New("attendance adjustment request not found")
	ErrOvertimeRequestNotFound = errors.New("overtime request not found")
)

func (q *Querier) GetEmployeeAttendanceIdentity(ctx context.Context, userID int64) (models.EmployeeAttendanceIdentity, error) {
	var identity models.EmployeeAttendanceIdentity
	err := q.db.QueryRowContext(ctx, `
		SELECT p.owner_user_id, p.member_id, m.name, m.division,
		       p.work_mode, p.schedule_type, p.weekly_target_minutes
		FROM member_attendance_profiles p
		JOIN members m ON m.id = p.member_id AND m.owner_user_id = p.owner_user_id
		WHERE p.linked_user_id = ? AND p.attendance_enabled = TRUE
	`, userID).Scan(
		&identity.OwnerUserID,
		&identity.MemberID,
		&identity.MemberName,
		&identity.Division,
		&identity.WorkMode,
		&identity.ScheduleType,
		&identity.WeeklyTargetMinutes,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return models.EmployeeAttendanceIdentity{}, ErrEmployeeProfileNotFound
	}
	return identity, err
}

func (q *Querier) GetAttendanceInviteTarget(ctx context.Context, ownerUserID, memberID int64) (string, *int64, string, error) {
	var name string
	var linked sql.NullInt64
	var linkedEmail sql.NullString
	err := q.db.QueryRowContext(ctx, `
		SELECT m.name, p.linked_user_id, u.email
		FROM member_attendance_profiles p
		JOIN members m ON m.id = p.member_id AND m.owner_user_id = p.owner_user_id
		LEFT JOIN users u ON u.id = p.linked_user_id
		WHERE p.owner_user_id = ? AND p.member_id = ?
	`, ownerUserID, memberID).Scan(&name, &linked, &linkedEmail)
	if errors.Is(err, sql.ErrNoRows) {
		return "", nil, "", ErrEmployeeProfileNotFound
	}
	if err != nil {
		return "", nil, "", err
	}
	var linkedID *int64
	if linked.Valid {
		v := linked.Int64
		linkedID = &v
	}
	return name, linkedID, linkedEmail.String, nil
}

func (q *Querier) LinkEmployeeUser(ctx context.Context, ownerUserID, memberID, userID int64) error {
	result, err := q.db.ExecContext(ctx, `
		UPDATE member_attendance_profiles
		SET linked_user_id = ?, updated_at = CURRENT_TIMESTAMP
		WHERE owner_user_id = ? AND member_id = ?
	`, userID, ownerUserID, memberID)
	if err != nil {
		return err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return ErrEmployeeProfileNotFound
	}
	return nil
}

func (q *Querier) MarkStaleActiveSessionsIncomplete(ctx context.Context, memberID int64, today string) error {
	_, err := q.db.ExecContext(ctx, `
		UPDATE attendance_sessions
		SET status = 'incomplete', updated_at = CURRENT_TIMESTAMP
		WHERE member_id = ? AND status = 'active' AND work_date < ?
	`, memberID, today)
	return err
}

func (q *Querier) GetActiveAttendanceSession(ctx context.Context, memberID int64) (*models.AttendanceSession, error) {
	var session models.AttendanceSession
	err := q.db.QueryRowContext(ctx, `
		SELECT id, member_id, work_date, location_mode, check_in_at,
		       check_out_at, status, source
		FROM attendance_sessions
		WHERE member_id = ? AND status = 'active'
		LIMIT 1
	`, memberID).Scan(
		&session.ID,
		&session.MemberID,
		&session.WorkDate,
		&session.LocationMode,
		&session.CheckInAt,
		&session.CheckOutAt,
		&session.Status,
		&session.Source,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrActiveAttendanceSessionNotFound
	}
	if err != nil {
		return nil, err
	}
	return &session, nil
}

func (q *Querier) CreateAttendanceSession(ctx context.Context, ownerUserID, memberID, userID int64, workDate, locationMode string, checkIn time.Time, source string) (*models.AttendanceSession, error) {
	var session models.AttendanceSession
	err := q.db.QueryRowContext(ctx, `
		INSERT INTO attendance_sessions (
			owner_user_id, member_id, user_id, work_date, location_mode,
			check_in_at, status, source
		) VALUES (?, ?, ?, ?, ?, ?, 'active', ?)
		RETURNING id, member_id, work_date, location_mode, check_in_at,
		          check_out_at, status, source
	`, ownerUserID, memberID, userID, workDate, locationMode, checkIn.UTC(), source).Scan(
		&session.ID,
		&session.MemberID,
		&session.WorkDate,
		&session.LocationMode,
		&session.CheckInAt,
		&session.CheckOutAt,
		&session.Status,
		&session.Source,
	)
	if err != nil {
		return nil, err
	}
	return &session, nil
}

func (q *Querier) CompleteAttendanceSession(ctx context.Context, sessionID, memberID int64, checkOut time.Time) error {
	result, err := q.db.ExecContext(ctx, `
		UPDATE attendance_sessions
		SET check_out_at = ?, status = 'completed', updated_at = CURRENT_TIMESTAMP
		WHERE id = ? AND member_id = ? AND status = 'active'
	`, checkOut.UTC(), sessionID, memberID)
	if err != nil {
		return err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return ErrActiveAttendanceSessionNotFound
	}
	return nil
}

func (q *Querier) ListAttendanceSessions(ctx context.Context, memberID int64, startDate, endDate string) ([]models.AttendanceSession, error) {
	rows, err := q.db.QueryContext(ctx, `
		SELECT id, member_id, work_date, location_mode, check_in_at,
		       check_out_at, status, source
		FROM attendance_sessions
		WHERE member_id = ? AND work_date >= ? AND work_date <= ?
		ORDER BY check_in_at ASC
	`, memberID, startDate, endDate)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]models.AttendanceSession, 0)
	for rows.Next() {
		var item models.AttendanceSession
		var checkOut sql.NullTime
		if err := rows.Scan(
			&item.ID,
			&item.MemberID,
			&item.WorkDate,
			&item.LocationMode,
			&item.CheckInAt,
			&checkOut,
			&item.Status,
			&item.Source,
		); err != nil {
			return nil, err
		}
		if checkOut.Valid {
			v := checkOut.Time
			item.CheckOutAt = &v
			if item.Status == "completed" && v.After(item.CheckInAt) {
				item.DurationMinutes = int(v.Sub(item.CheckInAt).Minutes())
			}
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (q *Querier) CreateAttendanceAdjustmentRequest(ctx context.Context, ownerUserID, memberID, userID int64, workDate string, checkIn, checkOut time.Time, reason string) error {
	_, err := q.db.ExecContext(ctx, `
		INSERT INTO attendance_adjustment_requests (
			owner_user_id, member_id, requested_by_user_id, work_date,
			requested_check_in, requested_check_out, reason
		) VALUES (?, ?, ?, ?, ?, ?, ?)
	`, ownerUserID, memberID, userID, workDate, checkIn.UTC(), checkOut.UTC(), reason)
	return err
}

func (q *Querier) ListAttendanceAdjustmentsByMember(ctx context.Context, memberID int64, limit int) ([]models.AttendanceAdjustmentRequest, error) {
	rows, err := q.db.QueryContext(ctx, `
		SELECT r.id, r.member_id, m.name, r.work_date, r.requested_check_in,
		       r.requested_check_out, r.reason, r.status, r.review_note,
		       r.reviewed_at, r.created_at
		FROM attendance_adjustment_requests r
		JOIN members m ON m.id = r.member_id
		WHERE r.member_id = ?
		ORDER BY r.created_at DESC
		LIMIT ?
	`, memberID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanAdjustmentRows(rows)
}

func (q *Querier) ListPendingAttendanceAdjustmentsByOwner(ctx context.Context, ownerUserID int64) ([]models.AttendanceAdjustmentRequest, error) {
	rows, err := q.db.QueryContext(ctx, `
		SELECT r.id, r.member_id, m.name, r.work_date, r.requested_check_in,
		       r.requested_check_out, r.reason, r.status, r.review_note,
		       r.reviewed_at, r.created_at
		FROM attendance_adjustment_requests r
		JOIN members m ON m.id = r.member_id AND m.owner_user_id = r.owner_user_id
		WHERE r.owner_user_id = ? AND r.status = 'pending'
		ORDER BY r.created_at ASC
	`, ownerUserID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanAdjustmentRows(rows)
}

func scanAdjustmentRows(rows *sql.Rows) ([]models.AttendanceAdjustmentRequest, error) {
	items := make([]models.AttendanceAdjustmentRequest, 0)
	for rows.Next() {
		var item models.AttendanceAdjustmentRequest
		var reviewedAt sql.NullTime
		if err := rows.Scan(
			&item.ID,
			&item.MemberID,
			&item.MemberName,
			&item.WorkDate,
			&item.RequestedCheckIn,
			&item.RequestedCheckOut,
			&item.Reason,
			&item.Status,
			&item.ReviewNote,
			&reviewedAt,
			&item.CreatedAt,
		); err != nil {
			return nil, err
		}
		if reviewedAt.Valid {
			v := reviewedAt.Time
			item.ReviewedAt = &v
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (q *Querier) ReviewAttendanceAdjustment(ctx context.Context, ownerUserID, requestID, reviewerUserID int64, approve bool, reviewNote string) error {
	var memberID, requestedByUserID int64
	var workDate, reason string
	var checkIn, checkOut time.Time
	var status string
	if err := q.db.QueryRowContext(ctx, `
		SELECT member_id, requested_by_user_id, work_date, requested_check_in,
		       requested_check_out, reason, status
		FROM attendance_adjustment_requests
		WHERE id = ? AND owner_user_id = ?
	`, requestID, ownerUserID).Scan(
		&memberID, &requestedByUserID, &workDate, &checkIn, &checkOut, &reason, &status,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrAdjustmentRequestNotFound
		}
		return err
	}
	if status != "pending" {
		return ErrAdjustmentRequestNotFound
	}

	newStatus := "rejected"
	var createdSessionID interface{}
	if approve {
		newStatus = "approved"
		result, err := q.db.ExecContext(ctx, `
			INSERT INTO attendance_sessions (
				owner_user_id, member_id, user_id, work_date, location_mode,
				check_in_at, check_out_at, status, source
			) VALUES (?, ?, ?, ?, 'remote', ?, ?, 'completed', 'manual_approved')
		`, ownerUserID, memberID, requestedByUserID, workDate, checkIn.UTC(), checkOut.UTC())
		if err != nil {
			return err
		}
		id, err := result.LastInsertId()
		if err != nil {
			return err
		}
		createdSessionID = id
	} else {
		createdSessionID = nil
	}

	result, err := q.db.ExecContext(ctx, `
		UPDATE attendance_adjustment_requests
		SET status = ?, reviewed_by_user_id = ?, reviewed_at = CURRENT_TIMESTAMP,
		    review_note = ?, created_session_id = ?, updated_at = CURRENT_TIMESTAMP
		WHERE id = ? AND owner_user_id = ? AND status = 'pending'
	`, newStatus, reviewerUserID, reviewNote, createdSessionID, requestID, ownerUserID)
	if err != nil {
		return err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return ErrAdjustmentRequestNotFound
	}
	_ = reason
	return nil
}

func (q *Querier) CreateOvertimeRequest(ctx context.Context, ownerUserID, memberID, userID int64, workDate string, requestedMinutes int, reason string) error {
	_, err := q.db.ExecContext(ctx, `
		INSERT INTO overtime_requests (
			owner_user_id, member_id, work_date, requested_minutes,
			reason, status, created_by_user_id
		) VALUES (?, ?, ?, ?, ?, 'pending', ?)
	`, ownerUserID, memberID, workDate, requestedMinutes, reason, userID)
	return err
}

func (q *Querier) ListOvertimeRequestsByMember(ctx context.Context, memberID int64, limit int) ([]models.OvertimeRequest, error) {
	rows, err := q.db.QueryContext(ctx, `
		SELECT id, member_id, work_date, requested_minutes, reason, status,
		       approved_by_user_id, approved_at, review_note, created_at
		FROM overtime_requests
		WHERE member_id = ?
		ORDER BY created_at DESC
		LIMIT ?
	`, memberID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanOvertimeRows(rows)
}

func (q *Querier) ListPendingOvertimeRequestsByOwner(ctx context.Context, ownerUserID int64) ([]models.OvertimeRequest, error) {
	rows, err := q.db.QueryContext(ctx, `
		SELECT id, member_id, work_date, requested_minutes, reason, status,
		       approved_by_user_id, approved_at, review_note, created_at
		FROM overtime_requests
		WHERE owner_user_id = ? AND status = 'pending'
		ORDER BY created_at ASC
	`, ownerUserID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanOvertimeRows(rows)
}

func scanOvertimeRows(rows *sql.Rows) ([]models.OvertimeRequest, error) {
	items := make([]models.OvertimeRequest, 0)
	for rows.Next() {
		var item models.OvertimeRequest
		var approvedBy sql.NullInt64
		var approvedAt sql.NullTime
		if err := rows.Scan(
			&item.ID,
			&item.MemberID,
			&item.WorkDate,
			&item.RequestedMinutes,
			&item.Reason,
			&item.Status,
			&approvedBy,
			&approvedAt,
			&item.ReviewNote,
			&item.CreatedAt,
		); err != nil {
			return nil, err
		}
		if approvedBy.Valid {
			v := approvedBy.Int64
			item.ApprovedByUserID = &v
		}
		if approvedAt.Valid {
			v := approvedAt.Time
			item.ApprovedAt = &v
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (q *Querier) ReviewOvertimeRequest(ctx context.Context, ownerUserID, requestID, reviewerUserID int64, approve bool, reviewNote string) error {
	status := "rejected"
	if approve {
		status = "approved"
	}
	result, err := q.db.ExecContext(ctx, `
		UPDATE overtime_requests
		SET status = ?, approved_by_user_id = ?, approved_at = CURRENT_TIMESTAMP,
		    review_note = ?, updated_at = CURRENT_TIMESTAMP
		WHERE id = ? AND owner_user_id = ? AND status = 'pending'
	`, status, reviewerUserID, reviewNote, requestID, ownerUserID)
	if err != nil {
		return err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return ErrOvertimeRequestNotFound
	}
	return nil
}
