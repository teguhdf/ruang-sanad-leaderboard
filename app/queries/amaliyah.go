package queries

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/maulanashalihin/laju-go/app/models"
)

var ErrAmaliyahMemberNotFound = errors.New("amaliyah member not found")

func (q *Querier) SaveAmaliyahSnapshot(
	ctx context.Context,
	ownerUserID int64,
	memberID int64,
	periodMonth string,
	fileName string,
	fileSHA256 string,
	overallScore float64,
	metrics []models.AmaliyahMetricScore,
) error {
	var exists int
	if err := q.db.QueryRowContext(ctx, `
		SELECT COUNT(*)
		FROM members
		WHERE id = ? AND owner_user_id = ?
	`, memberID, ownerUserID).Scan(&exists); err != nil {
		return err
	}
	if exists == 0 {
		return ErrAmaliyahMemberNotFound
	}

	beginner, ok := q.db.(txBeginner)
	if !ok {
		return fmt.Errorf("amaliyah: database does not support transactions")
	}

	tx, err := beginner.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var snapshotID int64
	if err := tx.QueryRowContext(ctx, `
		INSERT INTO amaliyah_snapshots (
			owner_user_id, member_id, period_month, source_file_name,
			source_file_sha256, overall_score, created_at, updated_at
		)
		VALUES (?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
		ON CONFLICT(owner_user_id, member_id, period_month)
		DO UPDATE SET
			source_file_name = excluded.source_file_name,
			source_file_sha256 = excluded.source_file_sha256,
			overall_score = excluded.overall_score,
			updated_at = CURRENT_TIMESTAMP
		RETURNING id
	`, ownerUserID, memberID, periodMonth, fileName, fileSHA256, overallScore).Scan(&snapshotID); err != nil {
		return err
	}

	if _, err := tx.ExecContext(ctx, `DELETE FROM amaliyah_metric_scores WHERE snapshot_id = ?`, snapshotID); err != nil {
		return err
	}

	stmt, err := tx.PrepareContext(ctx, `
		INSERT INTO amaliyah_metric_scores (
			snapshot_id, metric_key, completed_count, available_count, score
		)
		VALUES (?, ?, ?, ?, ?)
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, metric := range metrics {
		if _, err := stmt.ExecContext(
			ctx,
			snapshotID,
			metric.Key,
			metric.Completed,
			metric.Available,
			metric.Score,
		); err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (q *Querier) ListAmaliyahMemberSnapshots(ctx context.Context, ownerUserID int64, periodMonth string) ([]models.AmaliyahMemberSnapshot, error) {
	rows, err := q.db.QueryContext(ctx, `
		SELECT
			s.id,
			s.member_id,
			m.name,
			m.division,
			s.period_month,
			s.source_file_name,
			s.overall_score,
			s.updated_at,
			COALESCE(ms.metric_key, ''),
			COALESCE(ms.completed_count, 0),
			COALESCE(ms.available_count, 0),
			COALESCE(ms.score, 0)
		FROM amaliyah_snapshots s
		JOIN members m ON m.id = s.member_id AND m.owner_user_id = s.owner_user_id
		LEFT JOIN amaliyah_metric_scores ms ON ms.snapshot_id = s.id
		WHERE s.owner_user_id = ? AND s.period_month = ?
		ORDER BY m.division COLLATE NOCASE, m.name COLLATE NOCASE, ms.metric_key
	`, ownerUserID, periodMonth)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]models.AmaliyahMemberSnapshot, 0)
	indexByID := make(map[int64]int)

	for rows.Next() {
		var (
			id             int64
			memberID       int64
			memberName     string
			division       string
			period         string
			sourceFileName string
			overallScore   float64
			updatedAt      time.Time
			metricKey      string
			completed      int
			available      int
			metricScore    float64
		)

		if err := rows.Scan(
			&id,
			&memberID,
			&memberName,
			&division,
			&period,
			&sourceFileName,
			&overallScore,
			&updatedAt,
			&metricKey,
			&completed,
			&available,
			&metricScore,
		); err != nil {
			return nil, err
		}

		idx, ok := indexByID[id]
		if !ok {
			idx = len(items)
			indexByID[id] = idx
			items = append(items, models.AmaliyahMemberSnapshot{
				ID:             id,
				MemberID:       memberID,
				MemberName:     memberName,
				Division:       division,
				PeriodMonth:    period,
				SourceFileName: sourceFileName,
				OverallScore:   overallScore,
				Metrics:        make([]models.AmaliyahMetricScore, 0, 4),
				UpdatedAt:      updatedAt,
			})
		}

		if metricKey != "" {
			items[idx].Metrics = append(items[idx].Metrics, models.AmaliyahMetricScore{
				Key:       metricKey,
				Completed: completed,
				Available: available,
				Score:     metricScore,
			})
		}
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

func (q *Querier) ListAmaliyahPeriods(ctx context.Context, ownerUserID int64) ([]string, error) {
	rows, err := q.db.QueryContext(ctx, `
		SELECT DISTINCT period_month
		FROM amaliyah_snapshots
		WHERE owner_user_id = ?
		ORDER BY period_month DESC
	`, ownerUserID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	periods := make([]string, 0)
	for rows.Next() {
		var period string
		if err := rows.Scan(&period); err != nil {
			return nil, err
		}
		periods = append(periods, period)
	}
	return periods, rows.Err()
}
