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

func setupMemberTestDB(t *testing.T) (*sql.DB, *Querier) {
	t.Helper()

	db, err := sql.Open("sqlite3", "file:member-test?mode=memory&cache=shared")
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
			photo_url TEXT NOT NULL DEFAULT '',
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (owner_user_id) REFERENCES users(id) ON DELETE CASCADE,
			UNIQUE(owner_user_id, fingerprint_id)
		);

		INSERT INTO users(id) VALUES (1);
		INSERT INTO users(id) VALUES (2);
	`)
	require.NoError(t, err)

	t.Cleanup(func() {
		_ = db.Close()
	})

	return db, NewQuerier(db)
}

func TestMemberPersistenceAndOwnerIsolation(t *testing.T) {
	_, q := setupMemberTestDB(t)
	ctx := context.Background()

	first, err := q.CreateMemberRecord(ctx, models.Member{
		OwnerUserID:    1,
		Name:           "Fathiyah",
		Division:       "SMM",
		WorkDays:       []string{"mon", "tue", "wed", "thu", "fri"},
		StartTime:      "08:00",
		FingerprintID:  "FP001",
		GoogleSheetURL: "https://docs.google.com/spreadsheets/d/example-1/edit",
	})
	require.NoError(t, err)
	require.NotZero(t, first.ID)

	second, err := q.CreateMemberRecord(ctx, models.Member{
		OwnerUserID:    2,
		Name:           "Owner Dua",
		Division:       "Admin",
		WorkDays:       []string{"mon"},
		StartTime:      "09:00",
		FingerprintID:  "FP001",
		GoogleSheetURL: "https://docs.google.com/spreadsheets/d/example-2/edit",
	})
	require.NoError(t, err, "fingerprint id may repeat across different owners")

	ownerOneMembers, err := q.ListMembersByOwner(ctx, 1)
	require.NoError(t, err)
	require.Len(t, ownerOneMembers, 1)
	require.Equal(t, "Fathiyah", ownerOneMembers[0].Name)
	require.Empty(t, ownerOneMembers[0].PhotoURL)

	_, err = q.CreateMemberRecord(ctx, models.Member{
		OwnerUserID:    1,
		Name:           "Duplicate",
		Division:       "SMM",
		WorkDays:       []string{"mon"},
		StartTime:      "08:00",
		FingerprintID:  "FP001",
		GoogleSheetURL: "https://docs.google.com/spreadsheets/d/example-3/edit",
	})
	require.Error(t, err, "fingerprint id must be unique for one owner")

	err = q.UpdateMemberRecord(ctx, 1, second.ID, models.Member{
		Name:           "Cross Owner",
		Division:       "SMM",
		WorkDays:       []string{"mon"},
		StartTime:      "08:00",
		FingerprintID:  "FP999",
		GoogleSheetURL: "https://docs.google.com/spreadsheets/d/example-4/edit",
	})
	require.True(t, errors.Is(err, ErrMemberNotFound), "owner must not update another owner's member")

	err = q.UpdateMemberRecord(ctx, 1, first.ID, models.Member{
		Name:           "Fathiyah Updated",
		Division:       "SMM",
		WorkDays:       []string{"mon", "wed", "fri"},
		StartTime:      "08:30",
		FingerprintID:  "FP002",
		GoogleSheetURL: "https://docs.google.com/spreadsheets/d/example-1/edit",
	})
	require.NoError(t, err)

	err = q.UpdateMemberPhotoURL(ctx, 1, first.ID, "/storage/member-photos/member_1.webp")
	require.NoError(t, err)

	err = q.UpdateMemberPhotoURL(ctx, 1, second.ID, "/storage/member-photos/cross-owner.webp")
	require.True(t, errors.Is(err, ErrMemberNotFound), "owner must not update another owner's photo")

	ownerOneMembers, err = q.ListMembersByOwner(ctx, 1)
	require.NoError(t, err)
	require.Len(t, ownerOneMembers, 1)
	require.Equal(t, "Fathiyah Updated", ownerOneMembers[0].Name)
	require.Equal(t, []string{"mon", "wed", "fri"}, ownerOneMembers[0].WorkDays)
	require.Equal(t, "FP002", ownerOneMembers[0].FingerprintID)
	require.Equal(t, "/storage/member-photos/member_1.webp", ownerOneMembers[0].PhotoURL)
}
