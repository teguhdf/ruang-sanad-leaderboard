package queries

import (
	"context"
	"database/sql"
	"testing"

	"github.com/maulanashalihin/laju-go/app/models"
	"github.com/stretchr/testify/require"

	_ "github.com/mattn/go-sqlite3"
)

func setupAttendanceTestDB(t *testing.T) (*sql.DB, *Querier) {
	t.Helper()

	db, err := sql.Open("sqlite3", "file:attendance-foundation-test?mode=memory&cache=shared")
	require.NoError(t, err)
	db.SetMaxOpenConns(1)

	_, err = db.Exec(`
		PRAGMA foreign_keys = ON;

		CREATE TABLE users (
			id INTEGER PRIMARY KEY AUTOINCREMENT
		);

		CREATE TABLE members (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			owner_user_id INTEGER NOT NULL,
			name TEXT NOT NULL,
			division TEXT NOT NULL,
			work_days TEXT NOT NULL,
			start_time TEXT NOT NULL,
			fingerprint_id TEXT NOT NULL,
			google_sheet_url TEXT NOT NULL,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (owner_user_id) REFERENCES users(id) ON DELETE CASCADE
		);

		CREATE TABLE attendance_policies (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			owner_user_id INTEGER NOT NULL UNIQUE,
			weekday_target_minutes INTEGER NOT NULL DEFAULT 480,
			saturday_target_minutes INTEGER NOT NULL DEFAULT 240,
			weekly_target_minutes INTEGER NOT NULL DEFAULT 2640,
			late_tolerance_minutes INTEGER NOT NULL DEFAULT 5,
			early_leave_tolerance_minutes INTEGER NOT NULL DEFAULT 5,
			allow_weekly_carryover BOOLEAN NOT NULL DEFAULT TRUE,
			overtime_requires_approval BOOLEAN NOT NULL DEFAULT TRUE,
			week_starts_on INTEGER NOT NULL DEFAULT 1,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (owner_user_id) REFERENCES users(id) ON DELETE CASCADE
		);

		CREATE TABLE member_attendance_profiles (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			owner_user_id INTEGER NOT NULL,
			member_id INTEGER NOT NULL UNIQUE,
			linked_user_id INTEGER,
			work_mode TEXT NOT NULL DEFAULT 'onsite',
			schedule_type TEXT NOT NULL DEFAULT 'shift',
			required_daily_minutes INTEGER NOT NULL DEFAULT 480,
			weekly_target_minutes INTEGER NOT NULL DEFAULT 2640,
			attendance_enabled BOOLEAN NOT NULL DEFAULT TRUE,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (owner_user_id) REFERENCES users(id) ON DELETE CASCADE,
			FOREIGN KEY (member_id) REFERENCES members(id) ON DELETE CASCADE,
			FOREIGN KEY (linked_user_id) REFERENCES users(id) ON DELETE SET NULL
		);

		CREATE TABLE shift_templates (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			owner_user_id INTEGER NOT NULL,
			code TEXT NOT NULL,
			name TEXT NOT NULL,
			required_minutes INTEGER NOT NULL DEFAULT 480,
			description TEXT NOT NULL DEFAULT '',
			is_active BOOLEAN NOT NULL DEFAULT TRUE,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (owner_user_id) REFERENCES users(id) ON DELETE CASCADE,
			UNIQUE(owner_user_id, code)
		);

		CREATE TABLE shift_segments (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			shift_template_id INTEGER NOT NULL,
			sort_order INTEGER NOT NULL,
			location_mode TEXT NOT NULL,
			start_time TEXT,
			end_time TEXT,
			target_minutes INTEGER NOT NULL,
			is_flexible BOOLEAN NOT NULL DEFAULT FALSE,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (shift_template_id) REFERENCES shift_templates(id) ON DELETE CASCADE,
			UNIQUE(shift_template_id, sort_order)
		);

		INSERT INTO users(id) VALUES (1), (2);
		INSERT INTO members(id, owner_user_id, name, division, work_days, start_time, fingerprint_id, google_sheet_url)
		VALUES
			(10, 1, 'Fathiyah', 'SMM', 'mon,tue,wed,thu,fri,sat', '08:00', '9', 'https://docs.google.com/spreadsheets/d/a/edit'),
			(20, 2, 'Owner Dua', 'Ops', 'mon,tue,wed,thu,fri,sat', '08:00', '20', 'https://docs.google.com/spreadsheets/d/b/edit');
	`)
	require.NoError(t, err)

	t.Cleanup(func() { _ = db.Close() })
	return db, NewQuerier(db)
}

func TestAttendanceFoundationDefaultsAndOwnerIsolation(t *testing.T) {
	_, q := setupAttendanceTestDB(t)
	ctx := context.Background()

	require.NoError(t, q.EnsureAttendanceFoundation(ctx, 1))

	policy, err := q.GetAttendancePolicy(ctx, 1)
	require.NoError(t, err)
	require.Equal(t, 480, policy.WeekdayTargetMinutes)
	require.Equal(t, 240, policy.SaturdayTargetMinutes)
	require.Equal(t, 2640, policy.WeeklyTargetMinutes)
	require.Equal(t, 5, policy.LateToleranceMinutes)
	require.Equal(t, 5, policy.EarlyLeaveToleranceMinutes)
	require.True(t, policy.AllowWeeklyCarryover)
	require.True(t, policy.OvertimeRequiresApproval)

	profiles, err := q.ListMemberAttendanceProfiles(ctx, 1)
	require.NoError(t, err)
	require.Len(t, profiles, 1)
	require.Equal(t, int64(10), profiles[0].MemberID)
	require.Equal(t, "onsite", profiles[0].WorkMode)
	require.Equal(t, "shift", profiles[0].ScheduleType)

	shifts, err := q.ListShiftTemplates(ctx, 1)
	require.NoError(t, err)
	require.Len(t, shifts, 2)
	require.Equal(t, "morning", shifts[0].Code)
	require.Len(t, shifts[0].Segments, 3)
	require.Equal(t, 60, shifts[0].Segments[0].TargetMinutes)
	require.Equal(t, 360, shifts[0].Segments[1].TargetMinutes)
	require.True(t, shifts[0].Segments[2].Flexible)
	require.Equal(t, 60, shifts[0].Segments[2].TargetMinutes)

	require.NoError(t, q.UpdateMemberAttendanceProfile(ctx, 1, 10, models.MemberAttendanceProfileInput{
		WorkMode:     "hybrid",
		ScheduleType: "flexible",
	}))

	profiles, err = q.ListMemberAttendanceProfiles(ctx, 1)
	require.NoError(t, err)
	require.Equal(t, "hybrid", profiles[0].WorkMode)
	require.Equal(t, "flexible", profiles[0].ScheduleType)

	err = q.UpdateMemberAttendanceProfile(ctx, 2, 10, models.MemberAttendanceProfileInput{
		WorkMode:     "remote",
		ScheduleType: "flexible",
	})
	require.ErrorIs(t, err, ErrAttendanceProfileNotFound)

	otherOwner, err := q.ListMemberAttendanceProfiles(ctx, 2)
	require.NoError(t, err)
	require.Empty(t, otherOwner)
}
