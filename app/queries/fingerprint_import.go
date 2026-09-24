package queries

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/maulanashalihin/laju-go/app/models"
)

var ErrFingerprintImportNotFound = errors.New("fingerprint import not found")

type txBeginner interface {
	BeginTx(context.Context, *sql.TxOptions) (*sql.Tx, error)
}

func (q *Querier) FindFingerprintImportByHash(ctx context.Context, ownerUserID int64, fileSHA256 string) (*models.FingerprintImport, error) {
	row := q.db.QueryRowContext(ctx, `
		SELECT id, file_name, file_sha256, period_start, period_end,
		       file_people, matched, unmatched, total_slots, two_times,
		       one_time, empty, multiple_times, corrected_count, created_at
		FROM fingerprint_imports
		WHERE owner_user_id = ? AND file_sha256 = ?
	`, ownerUserID, fileSHA256)

	var item models.FingerprintImport
	if err := row.Scan(
		&item.ID,
		&item.FileName,
		&item.FileSHA256,
		&item.PeriodStart,
		&item.PeriodEnd,
		&item.FilePeople,
		&item.Matched,
		&item.Unmatched,
		&item.TotalSlots,
		&item.TwoTimes,
		&item.OneTime,
		&item.Empty,
		&item.MultipleTimes,
		&item.CorrectedCount,
		&item.CreatedAt,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrFingerprintImportNotFound
		}
		return nil, err
	}

	return &item, nil
}

func (q *Querier) SaveFingerprintImport(ctx context.Context, save models.FingerprintImportSave) (*models.FingerprintImport, error) {
	beginner, ok := q.db.(txBeginner)
	if !ok {
		return nil, fmt.Errorf("fingerprint import: database does not support transactions")
	}

	tx, err := beginner.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	result, err := tx.ExecContext(ctx, `
		INSERT INTO fingerprint_imports (
			owner_user_id, file_name, file_sha256, period_start, period_end,
			file_people, matched, unmatched, total_slots, two_times,
			one_time, empty, multiple_times, corrected_count
		)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`,
		save.OwnerUserID,
		save.Import.FileName,
		save.Import.FileSHA256,
		save.Import.PeriodStart,
		save.Import.PeriodEnd,
		save.Import.FilePeople,
		save.Import.Matched,
		save.Import.Unmatched,
		save.Import.TotalSlots,
		save.Import.TwoTimes,
		save.Import.OneTime,
		save.Import.Empty,
		save.Import.MultipleTimes,
		save.Import.CorrectedCount,
	)
	if err != nil {
		return nil, err
	}

	importID, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	stmt, err := tx.PrepareContext(ctx, `
		INSERT INTO fingerprint_records (
			import_id, owner_user_id, member_id, fingerprint_id, file_name,
			department, attendance_date, day, raw_value, raw_check_in,
			raw_check_out, corrected_check_in, corrected_check_out,
			correction_reason, is_corrected, read_state
		)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	for _, record := range save.Records {
		var memberID any
		if record.MemberID != nil {
			memberID = *record.MemberID
		}

		if _, err := stmt.ExecContext(
			ctx,
			importID,
			save.OwnerUserID,
			memberID,
			record.FingerprintID,
			record.FileName,
			record.Department,
			record.AttendanceDate,
			record.Day,
			record.RawValue,
			record.RawCheckIn,
			record.RawCheckOut,
			record.CorrectedCheckIn,
			record.CorrectedCheckOut,
			record.CorrectionReason,
			record.IsCorrected,
			record.ReadState,
		); err != nil {
			return nil, err
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	saved, err := q.FindFingerprintImportByHash(ctx, save.OwnerUserID, save.Import.FileSHA256)
	if err != nil {
		return nil, err
	}
	return saved, nil
}

func (q *Querier) ListFingerprintImportsByOwner(ctx context.Context, ownerUserID int64, limit int) ([]models.FingerprintImport, error) {
	if limit <= 0 || limit > 50 {
		limit = 10
	}

	rows, err := q.db.QueryContext(ctx, `
		SELECT id, file_name, file_sha256, period_start, period_end,
		       file_people, matched, unmatched, total_slots, two_times,
		       one_time, empty, multiple_times, corrected_count, created_at
		FROM fingerprint_imports
		WHERE owner_user_id = ?
		ORDER BY created_at DESC, id DESC
		LIMIT ?
	`, ownerUserID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]models.FingerprintImport, 0)
	for rows.Next() {
		var item models.FingerprintImport
		if err := rows.Scan(
			&item.ID,
			&item.FileName,
			&item.FileSHA256,
			&item.PeriodStart,
			&item.PeriodEnd,
			&item.FilePeople,
			&item.Matched,
			&item.Unmatched,
			&item.TotalSlots,
			&item.TwoTimes,
			&item.OneTime,
			&item.Empty,
			&item.MultipleTimes,
			&item.CorrectedCount,
			&item.CreatedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, item)
	}

	return items, rows.Err()
}
