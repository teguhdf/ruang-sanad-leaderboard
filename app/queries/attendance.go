package queries

import (
	"context"
	"database/sql"
	"errors"

	"github.com/maulanashalihin/laju-go/app/models"
)

var ErrAttendanceProfileNotFound = errors.New("attendance profile not found")

type defaultShiftSegment struct {
	SortOrder     int
	LocationMode string
	StartTime    *string
	EndTime      *string
	TargetMinutes int
	Flexible     bool
}

func (q *Querier) EnsureAttendanceFoundation(ctx context.Context, ownerUserID int64) error {
	if _, err := q.db.ExecContext(ctx, `
		INSERT OR IGNORE INTO attendance_policies (owner_user_id)
		VALUES (?)
	`, ownerUserID); err != nil {
		return err
	}

	if _, err := q.db.ExecContext(ctx, `
		INSERT OR IGNORE INTO member_attendance_profiles (
			owner_user_id, member_id, work_mode, schedule_type,
			required_daily_minutes, weekly_target_minutes, attendance_enabled
		)
		SELECT owner_user_id, id, 'onsite', 'shift', 480, 2640, TRUE
		FROM members
		WHERE owner_user_id = ?
	`, ownerUserID); err != nil {
		return err
	}

	if err := q.ensureDefaultShift(ctx, ownerUserID, "morning", "Shift Pagi", 480,
		"1 jam remote pagi, 6 jam kantor, lalu 1 jam remote fleksibel.", []defaultShiftSegment{
			{SortOrder: 1, LocationMode: "remote", StartTime: stringPtr("06:30"), EndTime: stringPtr("07:30"), TargetMinutes: 60, Flexible: false},
			{SortOrder: 2, LocationMode: "office", StartTime: stringPtr("08:00"), EndTime: stringPtr("14:00"), TargetMinutes: 360, Flexible: false},
			{SortOrder: 3, LocationMode: "remote", TargetMinutes: 60, Flexible: true},
		}); err != nil {
		return err
	}

	if err := q.ensureDefaultShift(ctx, ownerUserID, "afternoon", "Shift Siang", 480,
		"4 jam kantor, lalu 4 jam remote fleksibel.", []defaultShiftSegment{
			{SortOrder: 1, LocationMode: "office", StartTime: stringPtr("12:00"), EndTime: stringPtr("16:00"), TargetMinutes: 240, Flexible: false},
			{SortOrder: 2, LocationMode: "remote", TargetMinutes: 240, Flexible: true},
		}); err != nil {
		return err
	}

	return nil
}

func (q *Querier) GetAttendancePolicy(ctx context.Context, ownerUserID int64) (models.AttendancePolicy, error) {
	var policy models.AttendancePolicy
	err := q.db.QueryRowContext(ctx, `
		SELECT id, owner_user_id, weekday_target_minutes, saturday_target_minutes,
		       weekly_target_minutes, late_tolerance_minutes, early_leave_tolerance_minutes,
		       allow_weekly_carryover, overtime_requires_approval, week_starts_on,
		       created_at, updated_at
		FROM attendance_policies
		WHERE owner_user_id = ?
	`, ownerUserID).Scan(
		&policy.ID,
		&policy.OwnerUserID,
		&policy.WeekdayTargetMinutes,
		&policy.SaturdayTargetMinutes,
		&policy.WeeklyTargetMinutes,
		&policy.LateToleranceMinutes,
		&policy.EarlyLeaveToleranceMinutes,
		&policy.AllowWeeklyCarryover,
		&policy.OvertimeRequiresApproval,
		&policy.WeekStartsOn,
		&policy.CreatedAt,
		&policy.UpdatedAt,
	)
	return policy, err
}

func (q *Querier) ListMemberAttendanceProfiles(ctx context.Context, ownerUserID int64) ([]models.MemberAttendanceProfile, error) {
	rows, err := q.db.QueryContext(ctx, `
		SELECT p.id, p.member_id, m.name, m.division, p.linked_user_id,
		       p.work_mode, p.schedule_type, p.required_daily_minutes,
		       p.weekly_target_minutes, p.attendance_enabled
		FROM member_attendance_profiles p
		JOIN members m ON m.id = p.member_id
		WHERE p.owner_user_id = ? AND m.owner_user_id = ?
		ORDER BY m.name ASC
	`, ownerUserID, ownerUserID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	profiles := make([]models.MemberAttendanceProfile, 0)
	for rows.Next() {
		var profile models.MemberAttendanceProfile
		var linkedUserID sql.NullInt64
		if err := rows.Scan(
			&profile.ID,
			&profile.MemberID,
			&profile.MemberName,
			&profile.Division,
			&linkedUserID,
			&profile.WorkMode,
			&profile.ScheduleType,
			&profile.RequiredDailyMinutes,
			&profile.WeeklyTargetMinutes,
			&profile.AttendanceEnabled,
		); err != nil {
			return nil, err
		}
		if linkedUserID.Valid {
			value := linkedUserID.Int64
			profile.LinkedUserID = &value
		}
		profiles = append(profiles, profile)
	}
	return profiles, rows.Err()
}

func (q *Querier) UpdateMemberAttendanceProfile(ctx context.Context, ownerUserID, memberID int64, input models.MemberAttendanceProfileInput) error {
	result, err := q.db.ExecContext(ctx, `
		UPDATE member_attendance_profiles
		SET work_mode = ?, schedule_type = ?, updated_at = CURRENT_TIMESTAMP
		WHERE owner_user_id = ? AND member_id = ?
	`, input.WorkMode, input.ScheduleType, ownerUserID, memberID)
	if err != nil {
		return err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return ErrAttendanceProfileNotFound
	}
	return nil
}

func (q *Querier) ListShiftTemplates(ctx context.Context, ownerUserID int64) ([]models.ShiftTemplate, error) {
	rows, err := q.db.QueryContext(ctx, `
		SELECT id, code, name, required_minutes, description, is_active
		FROM shift_templates
		WHERE owner_user_id = ?
		ORDER BY id ASC
	`, ownerUserID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	shifts := make([]models.ShiftTemplate, 0)
	for rows.Next() {
		var shift models.ShiftTemplate
		if err := rows.Scan(&shift.ID, &shift.Code, &shift.Name, &shift.RequiredMinutes, &shift.Description, &shift.Active); err != nil {
			return nil, err
		}
		shift.Segments = make([]models.ShiftSegment, 0)
		shifts = append(shifts, shift)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	for i := range shifts {
		segments, err := q.listShiftSegments(ctx, shifts[i].ID)
		if err != nil {
			return nil, err
		}
		shifts[i].Segments = segments
	}
	return shifts, nil
}

func (q *Querier) ensureDefaultShift(ctx context.Context, ownerUserID int64, code, name string, requiredMinutes int, description string, segments []defaultShiftSegment) error {
	if _, err := q.db.ExecContext(ctx, `
		INSERT OR IGNORE INTO shift_templates (
			owner_user_id, code, name, required_minutes, description, is_active
		) VALUES (?, ?, ?, ?, ?, TRUE)
	`, ownerUserID, code, name, requiredMinutes, description); err != nil {
		return err
	}

	var shiftID int64
	if err := q.db.QueryRowContext(ctx, `
		SELECT id FROM shift_templates WHERE owner_user_id = ? AND code = ?
	`, ownerUserID, code).Scan(&shiftID); err != nil {
		return err
	}

	for _, segment := range segments {
		if _, err := q.db.ExecContext(ctx, `
			INSERT OR IGNORE INTO shift_segments (
				shift_template_id, sort_order, location_mode, start_time,
				end_time, target_minutes, is_flexible
			) VALUES (?, ?, ?, ?, ?, ?, ?)
		`, shiftID, segment.SortOrder, segment.LocationMode, segment.StartTime, segment.EndTime, segment.TargetMinutes, segment.Flexible); err != nil {
			return err
		}
	}
	return nil
}

func (q *Querier) listShiftSegments(ctx context.Context, shiftID int64) ([]models.ShiftSegment, error) {
	rows, err := q.db.QueryContext(ctx, `
		SELECT id, sort_order, location_mode, start_time, end_time, target_minutes, is_flexible
		FROM shift_segments
		WHERE shift_template_id = ?
		ORDER BY sort_order ASC
	`, shiftID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	segments := make([]models.ShiftSegment, 0)
	for rows.Next() {
		var segment models.ShiftSegment
		var startTime, endTime sql.NullString
		if err := rows.Scan(
			&segment.ID,
			&segment.SortOrder,
			&segment.LocationMode,
			&startTime,
			&endTime,
			&segment.TargetMinutes,
			&segment.Flexible,
		); err != nil {
			return nil, err
		}
		if startTime.Valid {
			value := startTime.String
			segment.StartTime = &value
		}
		if endTime.Valid {
			value := endTime.String
			segment.EndTime = &value
		}
		segments = append(segments, segment)
	}
	return segments, rows.Err()
}

func stringPtr(value string) *string {
	return &value
}
