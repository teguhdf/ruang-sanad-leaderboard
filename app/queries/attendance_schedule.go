package queries

import (
	"context"
	"database/sql"

	"github.com/maulanashalihin/laju-go/app/models"
)

func (q *Querier) ListRostersForMemberRange(ctx context.Context, ownerUserID, memberID int64, startDate, endDate string) ([]models.AttendanceRosterDay, error) {
	rows, err := q.db.QueryContext(ctx, `
		SELECT work_date, day_type, shift_template_id, work_mode, required_minutes
		FROM rosters
		WHERE owner_user_id = ? AND member_id = ?
		  AND work_date >= ? AND work_date <= ?
		ORDER BY work_date ASC
	`, ownerUserID, memberID, startDate, endDate)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]models.AttendanceRosterDay, 0)
	for rows.Next() {
		var item models.AttendanceRosterDay
		var shiftID sql.NullInt64
		if err := rows.Scan(
			&item.WorkDate,
			&item.DayType,
			&shiftID,
			&item.WorkMode,
			&item.RequiredMinutes,
		); err != nil {
			return nil, err
		}
		if shiftID.Valid {
			value := shiftID.Int64
			item.ShiftTemplateID = &value
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (q *Querier) ListApprovedLeavePeriodsForMemberRange(ctx context.Context, ownerUserID, memberID int64, startDate, endDate string) ([]models.ApprovedLeavePeriod, error) {
	rows, err := q.db.QueryContext(ctx, `
		SELECT start_date, end_date
		FROM leave_requests
		WHERE owner_user_id = ? AND member_id = ? AND status = 'approved'
		  AND start_date <= ? AND end_date >= ?
		ORDER BY start_date ASC
	`, ownerUserID, memberID, endDate, startDate)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]models.ApprovedLeavePeriod, 0)
	for rows.Next() {
		var item models.ApprovedLeavePeriod
		if err := rows.Scan(&item.StartDate, &item.EndDate); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (q *Querier) UpsertRosterDay(ctx context.Context, ownerUserID, memberID int64, workDate string, input models.RosterDayInput) error {
	var exists int
	if err := q.db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM members WHERE owner_user_id = ? AND id = ?
	`, ownerUserID, memberID).Scan(&exists); err != nil {
		return err
	}
	if exists != 1 {
		return ErrMemberNotFound
	}

	_, err := q.db.ExecContext(ctx, `
		INSERT INTO rosters (
			owner_user_id, member_id, work_date, day_type,
			shift_template_id, work_mode, required_minutes
		) VALUES (?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(member_id, work_date) DO UPDATE SET
			day_type = excluded.day_type,
			shift_template_id = excluded.shift_template_id,
			work_mode = excluded.work_mode,
			required_minutes = excluded.required_minutes,
			updated_at = CURRENT_TIMESTAMP
	`, ownerUserID, memberID, workDate, input.DayType, input.ShiftTemplateID, input.WorkMode, input.RequiredMinutes)
	return err
}

func (q *Querier) DeleteRosterDay(ctx context.Context, ownerUserID, memberID int64, workDate string) error {
	_, err := q.db.ExecContext(ctx, `
		DELETE FROM rosters
		WHERE owner_user_id = ? AND member_id = ? AND work_date = ?
	`, ownerUserID, memberID, workDate)
	return err
}
