package queries

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/maulanashalihin/laju-go/app/models"
	"github.com/stretchr/testify/require"

	_ "github.com/mattn/go-sqlite3"
)

func setupLeaveTestDB(t *testing.T) (*sql.DB, *Querier) {
	t.Helper()
	db, err := sql.Open("sqlite3", "file:leave-test?mode=memory&cache=shared")
	require.NoError(t, err)
	db.SetMaxOpenConns(1)

	_, err = db.Exec(`
		PRAGMA foreign_keys = ON;
		CREATE TABLE users (id INTEGER PRIMARY KEY);
		CREATE TABLE members (
			id INTEGER PRIMARY KEY,
			owner_user_id INTEGER NOT NULL,
			name TEXT NOT NULL,
			division TEXT NOT NULL,
			work_days TEXT NOT NULL DEFAULT 'mon,tue,wed,thu,fri,sat'
		);
		CREATE TABLE leave_entitlements (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			owner_user_id INTEGER NOT NULL,
			member_id INTEGER NOT NULL,
			year INTEGER NOT NULL,
			entitled_days INTEGER NOT NULL DEFAULT 0,
			carryover_days INTEGER NOT NULL DEFAULT 0,
			adjustment_days INTEGER NOT NULL DEFAULT 0,
			note TEXT NOT NULL DEFAULT '',
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(owner_user_id, member_id, year)
		);
		CREATE TABLE leave_requests (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			owner_user_id INTEGER NOT NULL,
			member_id INTEGER NOT NULL,
			requested_by_user_id INTEGER NOT NULL,
			entitlement_year INTEGER NOT NULL,
			start_date TEXT NOT NULL,
			end_date TEXT NOT NULL,
			chargeable_days INTEGER NOT NULL,
			reason TEXT NOT NULL,
			status TEXT NOT NULL DEFAULT 'pending',
			reviewed_by_user_id INTEGER,
			reviewed_at DATETIME,
			review_note TEXT NOT NULL DEFAULT '',
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
		);
		INSERT INTO users(id) VALUES (1), (2);
		INSERT INTO members(id, owner_user_id, name, division) VALUES (10, 1, 'Fathya', 'Multimedia');
	`)
	require.NoError(t, err)

	t.Cleanup(func() { _ = db.Close() })
	return db, NewQuerier(db)
}

func TestAnnualLeaveBalanceApprovalAndConflicts(t *testing.T) {
	_, q := setupLeaveTestDB(t)
	ctx := context.Background()

	require.NoError(t, q.UpsertLeaveEntitlement(ctx, 1, 10, models.LeaveEntitlementInput{
		Year: 2026, EntitledDays: 12,
	}))

	balance, err := q.GetLeaveBalance(ctx, 1, 10, 2026)
	require.NoError(t, err)
	require.Equal(t, 12, balance.AvailableDays)
	require.Equal(t, 0, balance.ApprovedDays)

	require.NoError(t, q.CreateLeaveRequest(ctx, 1, 10, 2, 2026, "2026-10-06", "2026-10-07", 2, "Keperluan keluarga"))
	balance, err = q.GetLeaveBalance(ctx, 1, 10, 2026)
	require.NoError(t, err)
	require.Equal(t, 12, balance.AvailableDays, "pending leave must not reduce available balance")
	require.Equal(t, 2, balance.PendingDays)

	pending, err := q.ListPendingLeaveRequestsByOwner(ctx, 1)
	require.NoError(t, err)
	require.Len(t, pending, 1)
	require.NoError(t, q.ReviewLeaveRequest(ctx, 1, pending[0].ID, 1, true, "Disetujui"))

	balance, err = q.GetLeaveBalance(ctx, 1, 10, 2026)
	require.NoError(t, err)
	require.Equal(t, 2, balance.ApprovedDays)
	require.Equal(t, 10, balance.AvailableDays)
	require.Equal(t, 0, balance.PendingDays)

	err = q.CreateLeaveRequest(ctx, 1, 10, 2, 2026, "2026-10-07", "2026-10-08", 1, "Bentrok cuti")
	require.ErrorIs(t, err, ErrLeaveRequestConflict)

	require.NoError(t, q.CreateLeaveRequest(ctx, 1, 10, 2, 2026, "2026-10-12", "2026-10-22", 11, "Cuti panjang"))
	pending, err = q.ListPendingLeaveRequestsByOwner(ctx, 1)
	require.NoError(t, err)
	require.Len(t, pending, 1)
	err = q.ReviewLeaveRequest(ctx, 1, pending[0].ID, 1, true, "Saldo tidak cukup")
	require.ErrorIs(t, err, ErrLeaveInsufficientBalance)

	var status string
	require.NoError(t, q.db.QueryRowContext(ctx, `SELECT status FROM leave_requests WHERE id = ?`, pending[0].ID).Scan(&status))
	require.Equal(t, "pending", status, "failed approval must not mutate request")

	require.NoError(t, q.ReviewLeaveRequest(ctx, 1, pending[0].ID, 1, false, "Ditolak"))
	balance, err = q.GetLeaveBalance(ctx, 1, 10, 2026)
	require.NoError(t, err)
	require.Equal(t, 10, balance.AvailableDays)
}

func TestCancelLeaveOnlyPending(t *testing.T) {
	_, q := setupLeaveTestDB(t)
	ctx := context.Background()
	require.NoError(t, q.UpsertLeaveEntitlement(ctx, 1, 10, models.LeaveEntitlementInput{Year: 2026, EntitledDays: 12}))
	require.NoError(t, q.CreateLeaveRequest(ctx, 1, 10, 2, 2026, "2026-11-02", "2026-11-02", 1, "Acara keluarga"))

	pending, err := q.ListPendingLeaveRequestsByOwner(ctx, 1)
	require.NoError(t, err)
	require.Len(t, pending, 1)
	require.NoError(t, q.CancelLeaveRequest(ctx, 10, pending[0].ID))

	err = q.CancelLeaveRequest(ctx, 10, pending[0].ID)
	require.True(t, errors.Is(err, ErrLeaveRequestNotFound))

	balance, err := q.GetLeaveBalance(ctx, 1, 10, 2026)
	require.NoError(t, err)
	require.Equal(t, 12, balance.AvailableDays)
	require.Equal(t, 0, balance.PendingDays)
}
