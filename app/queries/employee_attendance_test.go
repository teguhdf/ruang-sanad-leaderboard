package queries

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	_ "github.com/mattn/go-sqlite3"
)

func setupEmployeeAttendanceTestDB(t *testing.T) (*sql.DB, *Querier) {
	t.Helper()
	db, err := sql.Open("sqlite3", "file:employee-attendance-test?mode=memory&cache=shared")
	require.NoError(t, err)
	db.SetMaxOpenConns(1)

	_, err = db.Exec(`
		PRAGMA foreign_keys = ON;
		CREATE TABLE users (id INTEGER PRIMARY KEY, email TEXT, name TEXT, role TEXT);
		CREATE TABLE members (id INTEGER PRIMARY KEY, owner_user_id INTEGER NOT NULL, name TEXT NOT NULL, division TEXT NOT NULL);
		CREATE TABLE member_attendance_profiles (
			id INTEGER PRIMARY KEY, owner_user_id INTEGER NOT NULL, member_id INTEGER NOT NULL UNIQUE,
			linked_user_id INTEGER, work_mode TEXT NOT NULL, schedule_type TEXT NOT NULL,
			required_daily_minutes INTEGER NOT NULL, weekly_target_minutes INTEGER NOT NULL,
			attendance_enabled BOOLEAN NOT NULL
		);
		CREATE TABLE attendance_sessions (
			id INTEGER PRIMARY KEY AUTOINCREMENT, owner_user_id INTEGER NOT NULL, member_id INTEGER NOT NULL,
			user_id INTEGER NOT NULL, work_date TEXT NOT NULL, location_mode TEXT NOT NULL,
			check_in_at DATETIME NOT NULL, check_out_at DATETIME, status TEXT NOT NULL,
			source TEXT NOT NULL, created_at DATETIME DEFAULT CURRENT_TIMESTAMP, updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);
		CREATE UNIQUE INDEX idx_test_one_active ON attendance_sessions(member_id) WHERE status = 'active';
		CREATE TABLE attendance_adjustment_requests (
			id INTEGER PRIMARY KEY AUTOINCREMENT, owner_user_id INTEGER NOT NULL, member_id INTEGER NOT NULL,
			requested_by_user_id INTEGER NOT NULL, related_session_id INTEGER, work_date TEXT NOT NULL,
			location_mode TEXT NOT NULL, requested_check_in DATETIME NOT NULL, requested_check_out DATETIME NOT NULL,
			reason TEXT NOT NULL, status TEXT NOT NULL DEFAULT 'pending', reviewed_by_user_id INTEGER,
			reviewed_at DATETIME, review_note TEXT NOT NULL DEFAULT '', created_session_id INTEGER,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP, updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);
		CREATE TABLE overtime_requests (
			id INTEGER PRIMARY KEY AUTOINCREMENT, owner_user_id INTEGER NOT NULL, member_id INTEGER NOT NULL,
			work_date TEXT NOT NULL, requested_minutes INTEGER NOT NULL, reason TEXT NOT NULL,
			status TEXT NOT NULL DEFAULT 'pending', created_by_user_id INTEGER, approved_by_user_id INTEGER,
			approved_at DATETIME, review_note TEXT NOT NULL DEFAULT '', created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);
		INSERT INTO users(id,email,name,role) VALUES (1,'spv@test.local','SPV','spv');
		INSERT INTO users(id,email,name,role) VALUES (2,'employee@test.local','Employee','employee');
		INSERT INTO users(id,email,name,role) VALUES (3,'other@test.local','Other','employee');
		INSERT INTO members(id,owner_user_id,name,division) VALUES (10,1,'Fathiyah','Operasional');
		INSERT INTO member_attendance_profiles(id,owner_user_id,member_id,linked_user_id,work_mode,schedule_type,required_daily_minutes,weekly_target_minutes,attendance_enabled)
		VALUES (100,1,10,2,'hybrid','shift',480,2640,TRUE);
	`)
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })
	return db, NewQuerier(db)
}

func TestEmployeeAttendanceSessionAndApprovalFlow(t *testing.T) {
	db, q := setupEmployeeAttendanceTestDB(t)
	ctx := context.Background()

	identity, err := q.GetEmployeeAttendanceIdentity(ctx, 2)
	require.NoError(t, err)
	require.Equal(t, int64(10), identity.MemberID)
	_, err = q.GetEmployeeAttendanceIdentity(ctx, 3)
	require.ErrorIs(t, err, ErrEmployeeProfileNotFound)

	checkIn := time.Date(2026, 9, 29, 1, 0, 0, 0, time.UTC)
	session, err := q.CreateAttendanceSession(ctx, 1, 10, 2, "2026-09-29", "remote", checkIn, "app")
	require.NoError(t, err)
	require.Equal(t, "active", session.Status)

	_, err = q.CreateAttendanceSession(ctx, 1, 10, 2, "2026-09-29", "office", checkIn.Add(time.Hour), "app")
	require.Error(t, err, "only one active session per member is allowed")

	require.NoError(t, q.CompleteAttendanceSession(ctx, session.ID, 10, checkIn.Add(2*time.Hour)))
	sessions, err := q.ListAttendanceSessions(ctx, 10, "2026-09-29", "2026-09-29")
	require.NoError(t, err)
	require.Len(t, sessions, 1)
	require.Equal(t, 120, sessions[0].DurationMinutes)

	requestedIn := checkIn.Add(3 * time.Hour)
	requestedOut := checkIn.Add(4 * time.Hour)
	require.NoError(t, q.CreateAttendanceAdjustmentRequest(ctx, 1, 10, 2, "2026-09-29", "office", requestedIn, requestedOut, "Lupa check-in sesi kantor"))
	pending, err := q.ListPendingAttendanceAdjustmentsByOwner(ctx, 1)
	require.NoError(t, err)
	require.Len(t, pending, 1)
	require.Equal(t, "Fathiyah", pending[0].MemberName)
	require.Equal(t, "office", pending[0].LocationMode)
	require.NoError(t, q.ReviewAttendanceAdjustment(ctx, 1, pending[0].ID, 1, true, "OK"))

	var approvedCount int
	require.NoError(t, db.QueryRow(`SELECT COUNT(*) FROM attendance_adjustment_requests WHERE status='approved'`).Scan(&approvedCount))
	require.Equal(t, 1, approvedCount)
	var manualCount int
	require.NoError(t, db.QueryRow(`SELECT COUNT(*) FROM attendance_sessions WHERE source='manual_approved' AND status='completed' AND location_mode='office'`).Scan(&manualCount))
	require.Equal(t, 1, manualCount)

	forgotSession, err := q.CreateAttendanceSession(ctx, 1, 10, 2, "2026-09-29", "remote", checkIn.Add(5*time.Hour), "app")
	require.NoError(t, err)
	require.NoError(t, q.CreateForgotCheckoutAdjustment(
		ctx, 1, 10, 2, forgotSession.ID, "2026-09-29", "remote",
		forgotSession.CheckInAt, checkIn.Add(6*time.Hour), "Lupa checkout sesi remote",
	))
	_, err = q.GetActiveAttendanceSession(ctx, 10)
	require.ErrorIs(t, err, ErrActiveAttendanceSessionNotFound)

	var incompleteCount int
	require.NoError(t, db.QueryRow(`SELECT COUNT(*) FROM attendance_sessions WHERE id = ? AND status='incomplete'`, forgotSession.ID).Scan(&incompleteCount))
	require.Equal(t, 1, incompleteCount)

	pending, err = q.ListPendingAttendanceAdjustmentsByOwner(ctx, 1)
	require.NoError(t, err)
	require.Len(t, pending, 1)
	require.Equal(t, "remote", pending[0].LocationMode)
	require.NoError(t, q.ReviewAttendanceAdjustment(ctx, 1, pending[0].ID, 1, true, "Verified"))

	var replacementCount int
	require.NoError(t, db.QueryRow(`SELECT COUNT(*) FROM attendance_sessions WHERE source='manual_approved' AND status='completed' AND location_mode='remote'`).Scan(&replacementCount))
	require.Equal(t, 1, replacementCount)

	require.NoError(t, q.CreateOvertimeRequest(ctx, 1, 10, 2, "2026-09-30", 120, "Closing program"))
	overtime, err := q.ListPendingOvertimeRequestsByOwner(ctx, 1)
	require.NoError(t, err)
	require.Len(t, overtime, 1)
	require.NoError(t, q.ReviewOvertimeRequest(ctx, 1, overtime[0].ID, 1, true, "Approved"))
	var overtimeStatus string
	require.NoError(t, db.QueryRow(`SELECT status FROM overtime_requests WHERE id = ?`, overtime[0].ID).Scan(&overtimeStatus))
	require.Equal(t, "approved", overtimeStatus)
}
