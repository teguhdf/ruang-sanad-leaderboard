package queries

import (
	"context"
	"database/sql"
	"testing"

	"github.com/maulanashalihin/laju-go/app/models"
	"github.com/stretchr/testify/require"

	_ "github.com/mattn/go-sqlite3"
)

func setupAttendanceScheduleTestDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite3", "file:attendance-schedule-test?mode=memory&cache=shared")
	require.NoError(t, err)
	db.SetMaxOpenConns(1)
	t.Cleanup(func() { _ = db.Close() })
	_, err = db.Exec(`
		CREATE TABLE members (
			id INTEGER PRIMARY KEY,
			owner_user_id INTEGER NOT NULL,
			name TEXT NOT NULL DEFAULT '',
			division TEXT NOT NULL DEFAULT ''
		);
		CREATE TABLE rosters (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			owner_user_id INTEGER NOT NULL,
			member_id INTEGER NOT NULL,
			work_date TEXT NOT NULL,
			day_type TEXT NOT NULL,
			shift_template_id INTEGER,
			work_mode TEXT NOT NULL,
			required_minutes INTEGER NOT NULL,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(member_id, work_date)
		);
		CREATE TABLE leave_requests (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			owner_user_id INTEGER NOT NULL,
			member_id INTEGER NOT NULL,
			start_date TEXT NOT NULL,
			end_date TEXT NOT NULL,
			status TEXT NOT NULL
		);
		INSERT INTO members(id,owner_user_id,name,division) VALUES
			(10,1,'Rizki','Multimedia'),
			(20,2,'Other','Other');
	`)
	require.NoError(t, err)
	return db
}

func TestAttendanceScheduleQueriesUseOwnerRangeAndApprovedLeave(t *testing.T) {
	db := setupAttendanceScheduleTestDB(t)
	_, err := db.Exec(`
		INSERT INTO rosters(owner_user_id,member_id,work_date,day_type,work_mode,required_minutes)
		VALUES
			(1,10,'2026-10-07','replacement_off','onsite',0),
			(1,10,'2026-10-11','work','remote',480),
			(2,20,'2026-10-08','work','onsite',480);
		INSERT INTO leave_requests(owner_user_id,member_id,start_date,end_date,status)
		VALUES
			(1,10,'2026-10-06','2026-10-06','approved'),
			(1,10,'2026-10-08','2026-10-08','rejected'),
			(2,20,'2026-10-09','2026-10-09','approved');
	`)
	require.NoError(t, err)

	q := NewQuerier(db)
	ctx := context.Background()

	rosters, err := q.ListRostersForMemberRange(ctx, 1, 10, "2026-10-05", "2026-10-11")
	require.NoError(t, err)
	require.Len(t, rosters, 2)
	require.Equal(t, "replacement_off", rosters[0].DayType)
	require.Equal(t, "remote", rosters[1].WorkMode)

	leave, err := q.ListApprovedLeavePeriodsForMemberRange(ctx, 1, 10, "2026-10-05", "2026-10-11")
	require.NoError(t, err)
	require.Len(t, leave, 1)
	require.Equal(t, "2026-10-06", leave[0].StartDate)
}

func TestUpsertAndResetRosterDay(t *testing.T) {
	db := setupAttendanceScheduleTestDB(t)
	q := NewQuerier(db)
	ctx := context.Background()
	shiftID := int64(7)

	err := q.UpsertRosterDay(ctx, 1, 10, "2026-10-11", models.RosterDayInput{
		DayType: "work", ShiftTemplateID: &shiftID, WorkMode: "remote", RequiredMinutes: 480,
	})
	require.NoError(t, err)

	items, err := q.ListRostersForMemberRange(ctx, 1, 10, "2026-10-11", "2026-10-11")
	require.NoError(t, err)
	require.Len(t, items, 1)
	require.Equal(t, "work", items[0].DayType)
	require.Equal(t, "remote", items[0].WorkMode)
	require.Equal(t, 480, items[0].RequiredMinutes)
	require.NotNil(t, items[0].ShiftTemplateID)
	require.Equal(t, shiftID, *items[0].ShiftTemplateID)

	err = q.UpsertRosterDay(ctx, 1, 10, "2026-10-11", models.RosterDayInput{
		DayType: "replacement_off", WorkMode: "onsite", RequiredMinutes: 0,
	})
	require.NoError(t, err)
	items, err = q.ListRostersForMemberRange(ctx, 1, 10, "2026-10-11", "2026-10-11")
	require.NoError(t, err)
	require.Len(t, items, 1)
	require.Equal(t, "replacement_off", items[0].DayType)
	require.Zero(t, items[0].RequiredMinutes)

	require.NoError(t, q.DeleteRosterDay(ctx, 1, 10, "2026-10-11"))
	items, err = q.ListRostersForMemberRange(ctx, 1, 10, "2026-10-11", "2026-10-11")
	require.NoError(t, err)
	require.Empty(t, items)

	err = q.UpsertRosterDay(ctx, 1, 999, "2026-10-11", models.RosterDayInput{DayType: "work", WorkMode: "onsite", RequiredMinutes: 480})
	require.ErrorIs(t, err, ErrMemberNotFound)
}
