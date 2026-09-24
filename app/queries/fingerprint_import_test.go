package queries

import (
	"context"
	"database/sql"
	"testing"

	"github.com/maulanashalihin/laju-go/app/models"
	"github.com/stretchr/testify/require"

	_ "github.com/mattn/go-sqlite3"
)

func setupFingerprintImportTestDB(t *testing.T) (*sql.DB, *Querier) {
	t.Helper()

	db, err := sql.Open("sqlite3", "file:fingerprint-import-test?mode=memory&cache=shared")
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
			FOREIGN KEY (owner_user_id) REFERENCES users(id) ON DELETE CASCADE,
			UNIQUE(owner_user_id, fingerprint_id)
		);

		CREATE TABLE fingerprint_imports (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			owner_user_id INTEGER NOT NULL,
			file_name TEXT NOT NULL,
			file_sha256 TEXT NOT NULL,
			period_start TEXT NOT NULL,
			period_end TEXT NOT NULL,
			file_people INTEGER NOT NULL DEFAULT 0,
			matched INTEGER NOT NULL DEFAULT 0,
			unmatched INTEGER NOT NULL DEFAULT 0,
			total_slots INTEGER NOT NULL DEFAULT 0,
			two_times INTEGER NOT NULL DEFAULT 0,
			one_time INTEGER NOT NULL DEFAULT 0,
			empty INTEGER NOT NULL DEFAULT 0,
			multiple_times INTEGER NOT NULL DEFAULT 0,
			corrected_count INTEGER NOT NULL DEFAULT 0,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (owner_user_id) REFERENCES users(id) ON DELETE CASCADE,
			UNIQUE(owner_user_id, file_sha256)
		);

		CREATE TABLE fingerprint_records (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			import_id INTEGER NOT NULL,
			owner_user_id INTEGER NOT NULL,
			member_id INTEGER,
			fingerprint_id TEXT NOT NULL,
			file_name TEXT NOT NULL,
			department TEXT NOT NULL DEFAULT '',
			attendance_date TEXT NOT NULL,
			day INTEGER NOT NULL,
			raw_value TEXT NOT NULL DEFAULT '',
			raw_check_in TEXT NOT NULL DEFAULT '',
			raw_check_out TEXT NOT NULL DEFAULT '',
			corrected_check_in TEXT NOT NULL DEFAULT '',
			corrected_check_out TEXT NOT NULL DEFAULT '',
			correction_reason TEXT NOT NULL DEFAULT '',
			is_corrected INTEGER NOT NULL DEFAULT 0,
			read_state TEXT NOT NULL,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (import_id) REFERENCES fingerprint_imports(id) ON DELETE CASCADE,
			FOREIGN KEY (owner_user_id) REFERENCES users(id) ON DELETE CASCADE,
			FOREIGN KEY (member_id) REFERENCES members(id) ON DELETE SET NULL,
			UNIQUE(import_id, fingerprint_id, attendance_date)
		);

		INSERT INTO users(id) VALUES (1);
		INSERT INTO users(id) VALUES (2);

		INSERT INTO members(
			id, owner_user_id, name, division, work_days, start_time,
			fingerprint_id, google_sheet_url
		) VALUES (
			10, 1, 'Teh Icha', 'Koordinator Guru', 'mon,tue,wed,thu,fri',
			'08:00', '6', 'https://docs.google.com/spreadsheets/d/test/edit'
		);
	`)
	require.NoError(t, err)

	t.Cleanup(func() {
		_ = db.Close()
	})

	return db, NewQuerier(db)
}

func TestFingerprintImportPersistenceAndDuplicateProtection(t *testing.T) {
	db, q := setupFingerprintImportTestDB(t)
	ctx := context.Background()
	memberID := int64(10)

	save := models.FingerprintImportSave{
		OwnerUserID: 1,
		Import: models.FingerprintImport{
			FileName:       "agustus.xls",
			FileSHA256:     "same-hash",
			PeriodStart:    "2026-08-01",
			PeriodEnd:      "2026-08-31",
			FilePeople:     1,
			Matched:        1,
			TotalSlots:     2,
			TwoTimes:       1,
			OneTime:        1,
			CorrectedCount: 1,
		},
		Records: []models.FingerprintRecord{
			{
				MemberID:        &memberID,
				FingerprintID:    "6",
				FileName:         "Icha",
				AttendanceDate:   "2026-08-01",
				Day:              1,
				RawValue:         "08:03\n16:27",
				RawCheckIn:       "08:03",
				RawCheckOut:      "16:27",
				ReadState:        "two_times",
			},
			{
				MemberID:          &memberID,
				FingerprintID:      "6",
				FileName:           "Icha",
				AttendanceDate:     "2026-08-02",
				Day:                2,
				RawValue:           "08:04",
				RawCheckIn:         "08:04",
				CorrectedCheckIn:   "08:04",
				CorrectedCheckOut:  "16:05",
				CorrectionReason:   "Lupa fingerprint pulang",
				IsCorrected:        true,
				ReadState:          "one_time",
			},
		},
	}

	first, err := q.SaveFingerprintImport(ctx, save)
	require.NoError(t, err)
	require.NotZero(t, first.ID)
	require.Equal(t, 1, first.CorrectedCount)

	var importCount int
	require.NoError(t, db.QueryRow(`SELECT COUNT(*) FROM fingerprint_imports WHERE owner_user_id = 1`).Scan(&importCount))
	require.Equal(t, 1, importCount)

	var recordCount int
	require.NoError(t, db.QueryRow(`SELECT COUNT(*) FROM fingerprint_records WHERE import_id = ?`, first.ID).Scan(&recordCount))
	require.Equal(t, 2, recordCount)

	var rawCheckOut, correctedCheckOut, reason string
	var corrected bool
	require.NoError(t, db.QueryRow(`
		SELECT raw_check_out, corrected_check_out, correction_reason, is_corrected
		FROM fingerprint_records
		WHERE import_id = ? AND attendance_date = '2026-08-02'
	`, first.ID).Scan(&rawCheckOut, &correctedCheckOut, &reason, &corrected))
	require.Equal(t, "", rawCheckOut)
	require.Equal(t, "16:05", correctedCheckOut)
	require.Equal(t, "Lupa fingerprint pulang", reason)
	require.True(t, corrected)

	_, err = q.SaveFingerprintImport(ctx, save)
	require.Error(t, err, "same owner and same file hash must not create a duplicate import")

	require.NoError(t, db.QueryRow(`SELECT COUNT(*) FROM fingerprint_imports WHERE owner_user_id = 1`).Scan(&importCount))
	require.Equal(t, 1, importCount)

	list, err := q.ListFingerprintImportsByOwner(ctx, 1, 10)
	require.NoError(t, err)
	require.Len(t, list, 1)
	require.Equal(t, "agustus.xls", list[0].FileName)

	otherOwner, err := q.ListFingerprintImportsByOwner(ctx, 2, 10)
	require.NoError(t, err)
	require.Empty(t, otherOwner)
}
