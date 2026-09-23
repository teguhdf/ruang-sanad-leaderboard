package queries

import (
	"context"
	"errors"
	"strings"

	"github.com/maulanashalihin/laju-go/app/models"
)

var ErrMemberNotFound = errors.New("member not found")

func (q *Querier) ListMembersByOwner(ctx context.Context, ownerUserID int64) ([]models.Member, error) {
	rows, err := q.db.QueryContext(ctx, `
		SELECT id, owner_user_id, name, division, work_days, start_time,
		       fingerprint_id, google_sheet_url, created_at, updated_at
		FROM members
		WHERE owner_user_id = ?
		ORDER BY name ASC
	`, ownerUserID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	members := make([]models.Member, 0)
	for rows.Next() {
		var member models.Member
		var workDays string
		if err := rows.Scan(
			&member.ID,
			&member.OwnerUserID,
			&member.Name,
			&member.Division,
			&workDays,
			&member.StartTime,
			&member.FingerprintID,
			&member.GoogleSheetURL,
			&member.CreatedAt,
			&member.UpdatedAt,
		); err != nil {
			return nil, err
		}
		if workDays != "" {
			member.WorkDays = strings.Split(workDays, ",")
		} else {
			member.WorkDays = []string{}
		}
		members = append(members, member)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return members, nil
}

func (q *Querier) CreateMemberRecord(ctx context.Context, member models.Member) (*models.Member, error) {
	workDays := strings.Join(member.WorkDays, ",")

	row := q.db.QueryRowContext(ctx, `
		INSERT INTO members (
			owner_user_id, name, division, work_days, start_time,
			fingerprint_id, google_sheet_url
		)
		VALUES (?, ?, ?, ?, ?, ?, ?)
		RETURNING id, created_at, updated_at
	`,
		member.OwnerUserID,
		member.Name,
		member.Division,
		workDays,
		member.StartTime,
		member.FingerprintID,
		member.GoogleSheetURL,
	)

	if err := row.Scan(&member.ID, &member.CreatedAt, &member.UpdatedAt); err != nil {
		return nil, err
	}

	return &member, nil
}

func (q *Querier) UpdateMemberRecord(ctx context.Context, ownerUserID, memberID int64, member models.Member) error {
	result, err := q.db.ExecContext(ctx, `
		UPDATE members
		SET name = ?,
		    division = ?,
		    work_days = ?,
		    start_time = ?,
		    fingerprint_id = ?,
		    google_sheet_url = ?,
		    updated_at = CURRENT_TIMESTAMP
		WHERE id = ? AND owner_user_id = ?
	`,
		member.Name,
		member.Division,
		strings.Join(member.WorkDays, ","),
		member.StartTime,
		member.FingerprintID,
		member.GoogleSheetURL,
		memberID,
		ownerUserID,
	)
	if err != nil {
		return err
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return ErrMemberNotFound
	}

	return nil
}
