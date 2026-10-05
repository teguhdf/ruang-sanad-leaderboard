package queries

import (
	"context"
	"database/sql"
	"testing"

	"github.com/stretchr/testify/require"

	_ "github.com/mattn/go-sqlite3"
)

func TestAttendanceScheduleQueriesUseOwnerRangeAndApprovedLeave(t *testing.T) {
	db, err := sql.Open("sqlite3", "file:attendance-schedule-test?mode=memory&cache=shared")
	require.NoError(t, err)
	db.SetMaxOpenConns(1)
	t.Cleanup(func() { _ = db.Close() })

	_, err = db.Exec(`
		CREATE TABLE rosters (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			owner_user_id INTEGER NOT NULL,
			member_id INTEGER NOT NULL,
			work_date TEXT NOT NULL,
			day_type TEXT NOT NULL,
			shift_template_id INTEGER,
			work_mode TEXT NOT NULL,
			required_minutes INTEGER NOT NULL
		);
		CREATE TABLE leave_requests (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			owner_user_id INTEGER NOT NULL,
			member_id INTEGER NOT NULL,
			start_date TEXT NOT NULL,
			end_date TEXT NOT NULL,
			status TEXT NOT NULL
		);
		INSERT INTO rosters(owner_user_id,member_id,work_date,day_type,work_mode,required_minutes)
		VALUES
			(1,10,'2026-10-07','replacement_off','onsite',0),
			(1,10,'2026-10-11','work','remote',480),
			(2,10,'2026-10-08','work','onsite',480);
		INSERT INTO leave_requests(owner_user_id,member_id,start_date,end_date,status)
		VALUES
			(1,10,'2026-10-06','2026-10-06','approved'),
			(1,10,'2026-10-08','2026-10-08','rejected'),
			(2,10,'2026-10-09','2026-10-09','approved');
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
