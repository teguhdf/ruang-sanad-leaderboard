package queries

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"

	"github.com/maulanashalihin/laju-go/app/models"
)

var (
	ErrLeaveRequestNotFound    = errors.New("leave request not found")
	ErrLeaveRequestConflict    = errors.New("leave request overlaps existing request")
	ErrLeaveInsufficientBalance = errors.New("leave balance is insufficient")
)

func (q *Querier) UpsertLeaveEntitlement(ctx context.Context, ownerUserID, memberID int64, input models.LeaveEntitlementInput) error {
	var exists int
	if err := q.db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM members WHERE id = ? AND owner_user_id = ?
	`, memberID, ownerUserID).Scan(&exists); err != nil {
		return err
	}
	if exists != 1 {
		return ErrMemberNotFound
	}

	_, err := q.db.ExecContext(ctx, `
		INSERT INTO leave_entitlements (
			owner_user_id, member_id, year, entitled_days, note
		) VALUES (?, ?, ?, ?, ?)
		ON CONFLICT(owner_user_id, member_id, year) DO UPDATE SET
			entitled_days = excluded.entitled_days,
			note = excluded.note,
			updated_at = CURRENT_TIMESTAMP
	`, ownerUserID, memberID, input.Year, input.EntitledDays, strings.TrimSpace(input.Note))
	return err
}

func (q *Querier) GetLeaveBalance(ctx context.Context, ownerUserID, memberID int64, year int) (models.LeaveBalance, error) {
	var item models.LeaveBalance
	err := q.db.QueryRowContext(ctx, `
		SELECT m.id, m.name, m.division, ?,
		       COALESCE(e.entitled_days, 0),
		       COALESCE(e.carryover_days, 0),
		       COALESCE(e.adjustment_days, 0),
		       COALESCE((SELECT SUM(r.chargeable_days) FROM leave_requests r
		                 WHERE r.owner_user_id = m.owner_user_id AND r.member_id = m.id
		                   AND r.entitlement_year = ? AND r.status = 'approved'), 0),
		       COALESCE((SELECT SUM(r.chargeable_days) FROM leave_requests r
		                 WHERE r.owner_user_id = m.owner_user_id AND r.member_id = m.id
		                   AND r.entitlement_year = ? AND r.status = 'pending'), 0)
		FROM members m
		LEFT JOIN leave_entitlements e
		  ON e.owner_user_id = m.owner_user_id AND e.member_id = m.id AND e.year = ?
		WHERE m.owner_user_id = ? AND m.id = ?
	`, year, year, year, year, ownerUserID, memberID).Scan(
		&item.MemberID,
		&item.MemberName,
		&item.Division,
		&item.Year,
		&item.EntitledDays,
		&item.CarryoverDays,
		&item.AdjustmentDays,
		&item.ApprovedDays,
		&item.PendingDays,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return models.LeaveBalance{}, ErrMemberNotFound
	}
	if err != nil {
		return models.LeaveBalance{}, err
	}
	item.AvailableDays = item.EntitledDays + item.CarryoverDays + item.AdjustmentDays - item.ApprovedDays
	return item, nil
}

func (q *Querier) ListLeaveBalancesByOwner(ctx context.Context, ownerUserID int64, year int) ([]models.LeaveBalance, error) {
	rows, err := q.db.QueryContext(ctx, `
		SELECT m.id, m.name, m.division, ?,
		       COALESCE(e.entitled_days, 0),
		       COALESCE(e.carryover_days, 0),
		       COALESCE(e.adjustment_days, 0),
		       COALESCE((SELECT SUM(r.chargeable_days) FROM leave_requests r
		                 WHERE r.owner_user_id = m.owner_user_id AND r.member_id = m.id
		                   AND r.entitlement_year = ? AND r.status = 'approved'), 0),
		       COALESCE((SELECT SUM(r.chargeable_days) FROM leave_requests r
		                 WHERE r.owner_user_id = m.owner_user_id AND r.member_id = m.id
		                   AND r.entitlement_year = ? AND r.status = 'pending'), 0)
		FROM members m
		LEFT JOIN leave_entitlements e
		  ON e.owner_user_id = m.owner_user_id AND e.member_id = m.id AND e.year = ?
		WHERE m.owner_user_id = ?
		ORDER BY m.name ASC
	`, year, year, year, year, ownerUserID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]models.LeaveBalance, 0)
	for rows.Next() {
		var item models.LeaveBalance
		if err := rows.Scan(
			&item.MemberID,
			&item.MemberName,
			&item.Division,
			&item.Year,
			&item.EntitledDays,
			&item.CarryoverDays,
			&item.AdjustmentDays,
			&item.ApprovedDays,
			&item.PendingDays,
		); err != nil {
			return nil, err
		}
		item.AvailableDays = item.EntitledDays + item.CarryoverDays + item.AdjustmentDays - item.ApprovedDays
		items = append(items, item)
	}
	return items, rows.Err()
}

func (q *Querier) GetMemberWorkDaysForLeave(ctx context.Context, ownerUserID, memberID int64) ([]string, error) {
	var raw string
	err := q.db.QueryRowContext(ctx, `
		SELECT work_days FROM members WHERE owner_user_id = ? AND id = ?
	`, ownerUserID, memberID).Scan(&raw)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrMemberNotFound
	}
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(raw) == "" {
		return []string{}, nil
	}
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.ToLower(strings.TrimSpace(part))
		if part != "" {
			out = append(out, part)
		}
	}
	return out, nil
}

// GetRosterChargeability lets leave use explicit roster when one exists.
// When found=false the service falls back to the member's fixed work_days.
func (q *Querier) GetRosterChargeability(ctx context.Context, ownerUserID, memberID int64, workDate string) (chargeable bool, found bool, err error) {
	var dayType string
	var requiredMinutes int
	err = q.db.QueryRowContext(ctx, `
		SELECT day_type, required_minutes
		FROM rosters
		WHERE owner_user_id = ? AND member_id = ? AND work_date = ?
	`, ownerUserID, memberID, workDate).Scan(&dayType, &requiredMinutes)
	if errors.Is(err, sql.ErrNoRows) {
		return false, false, nil
	}
	if err != nil {
		return false, false, err
	}
	return strings.EqualFold(dayType, "work") && requiredMinutes > 0, true, nil
}

func (q *Querier) CreateLeaveRequest(ctx context.Context, ownerUserID, memberID, userID int64, year int, startDate, endDate string, chargeableDays int, reason string) error {
	beginner, ok := q.db.(attendanceTxBeginner)
	if !ok {
		return errors.New("database does not support transactions")
	}
	tx, err := beginner.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var conflict int
	if err := tx.QueryRowContext(ctx, `
		SELECT COUNT(*)
		FROM leave_requests
		WHERE owner_user_id = ? AND member_id = ?
		  AND status IN ('pending', 'approved')
		  AND start_date <= ? AND end_date >= ?
	`, ownerUserID, memberID, endDate, startDate).Scan(&conflict); err != nil {
		return err
	}
	if conflict > 0 {
		return ErrLeaveRequestConflict
	}

	if _, err := tx.ExecContext(ctx, `
		INSERT INTO leave_requests (
			owner_user_id, member_id, requested_by_user_id, entitlement_year,
			start_date, end_date, chargeable_days, reason
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`, ownerUserID, memberID, userID, year, startDate, endDate, chargeableDays, strings.TrimSpace(reason)); err != nil {
		return err
	}
	return tx.Commit()
}

func (q *Querier) ListLeaveRequestsByMember(ctx context.Context, memberID int64, limit int) ([]models.LeaveRequest, error) {
	if limit <= 0 || limit > 50 {
		limit = 10
	}
	rows, err := q.db.QueryContext(ctx, `
		SELECT r.id, r.member_id, m.name, m.division, r.entitlement_year,
		       r.start_date, r.end_date, r.chargeable_days, r.reason, r.status,
		       r.review_note, r.reviewed_at, r.created_at
		FROM leave_requests r
		JOIN members m ON m.id = r.member_id AND m.owner_user_id = r.owner_user_id
		WHERE r.member_id = ?
		ORDER BY r.created_at DESC, r.id DESC
		LIMIT ?
	`, memberID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanLeaveRows(rows)
}

func (q *Querier) ListPendingLeaveRequestsByOwner(ctx context.Context, ownerUserID int64) ([]models.LeaveRequest, error) {
	rows, err := q.db.QueryContext(ctx, `
		SELECT r.id, r.member_id, m.name, m.division, r.entitlement_year,
		       r.start_date, r.end_date, r.chargeable_days, r.reason, r.status,
		       r.review_note, r.reviewed_at, r.created_at
		FROM leave_requests r
		JOIN members m ON m.id = r.member_id AND m.owner_user_id = r.owner_user_id
		WHERE r.owner_user_id = ? AND r.status = 'pending'
		ORDER BY r.start_date ASC, r.created_at ASC
	`, ownerUserID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanLeaveRows(rows)
}

func scanLeaveRows(rows *sql.Rows) ([]models.LeaveRequest, error) {
	items := make([]models.LeaveRequest, 0)
	for rows.Next() {
		var item models.LeaveRequest
		var reviewedAt sql.NullTime
		if err := rows.Scan(
			&item.ID,
			&item.MemberID,
			&item.MemberName,
			&item.Division,
			&item.EntitlementYear,
			&item.StartDate,
			&item.EndDate,
			&item.ChargeableDays,
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

func (q *Querier) CancelLeaveRequest(ctx context.Context, memberID, requestID int64) error {
	result, err := q.db.ExecContext(ctx, `
		UPDATE leave_requests
		SET status = 'cancelled', updated_at = CURRENT_TIMESTAMP
		WHERE id = ? AND member_id = ? AND status = 'pending'
	`, requestID, memberID)
	if err != nil {
		return err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected != 1 {
		return ErrLeaveRequestNotFound
	}
	return nil
}

func (q *Querier) ReviewLeaveRequest(ctx context.Context, ownerUserID, requestID, reviewerUserID int64, approve bool, reviewNote string) error {
	beginner, ok := q.db.(attendanceTxBeginner)
	if !ok {
		return errors.New("database does not support transactions")
	}
	tx, err := beginner.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var memberID int64
	var year, chargeableDays int
	var startDate, endDate, status string
	if err := tx.QueryRowContext(ctx, `
		SELECT member_id, entitlement_year, start_date, end_date, chargeable_days, status
		FROM leave_requests
		WHERE id = ? AND owner_user_id = ?
	`, requestID, ownerUserID).Scan(&memberID, &year, &startDate, &endDate, &chargeableDays, &status); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrLeaveRequestNotFound
		}
		return err
	}
	if status != "pending" {
		return ErrLeaveRequestNotFound
	}

	newStatus := "rejected"
	if approve {
		var overlap int
		if err := tx.QueryRowContext(ctx, `
			SELECT COUNT(*) FROM leave_requests
			WHERE owner_user_id = ? AND member_id = ? AND id <> ?
			  AND status = 'approved'
			  AND start_date <= ? AND end_date >= ?
		`, ownerUserID, memberID, requestID, endDate, startDate).Scan(&overlap); err != nil {
			return err
		}
		if overlap > 0 {
			return ErrLeaveRequestConflict
		}

		var total int
		err := tx.QueryRowContext(ctx, `
			SELECT entitled_days + carryover_days + adjustment_days
			FROM leave_entitlements
			WHERE owner_user_id = ? AND member_id = ? AND year = ?
		`, ownerUserID, memberID, year).Scan(&total)
		if errors.Is(err, sql.ErrNoRows) {
			total = 0
		} else if err != nil {
			return err
		}

		var approved int
		if err := tx.QueryRowContext(ctx, `
			SELECT COALESCE(SUM(chargeable_days), 0)
			FROM leave_requests
			WHERE owner_user_id = ? AND member_id = ? AND entitlement_year = ?
			  AND status = 'approved' AND id <> ?
		`, ownerUserID, memberID, year, requestID).Scan(&approved); err != nil {
			return err
		}
		if approved+chargeableDays > total {
			return ErrLeaveInsufficientBalance
		}
		newStatus = "approved"
	}

	result, err := tx.ExecContext(ctx, `
		UPDATE leave_requests
		SET status = ?, reviewed_by_user_id = ?, reviewed_at = ?, review_note = ?, updated_at = CURRENT_TIMESTAMP
		WHERE id = ? AND owner_user_id = ? AND status = 'pending'
	`, newStatus, reviewerUserID, time.Now().UTC(), strings.TrimSpace(reviewNote), requestID, ownerUserID)
	if err != nil {
		return err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected != 1 {
		return ErrLeaveRequestNotFound
	}
	return tx.Commit()
}
