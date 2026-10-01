-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS amaliyah_snapshots (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    owner_user_id INTEGER NOT NULL,
    member_id INTEGER NOT NULL,
    period_month TEXT NOT NULL,
    source_file_name TEXT NOT NULL,
    source_file_sha256 TEXT NOT NULL,
    overall_score REAL NOT NULL DEFAULT 0,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (owner_user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (member_id) REFERENCES members(id) ON DELETE CASCADE,
    UNIQUE(owner_user_id, member_id, period_month)
);

CREATE INDEX IF NOT EXISTS idx_amaliyah_snapshots_owner_period
ON amaliyah_snapshots(owner_user_id, period_month);

CREATE INDEX IF NOT EXISTS idx_amaliyah_snapshots_member_period
ON amaliyah_snapshots(member_id, period_month);

CREATE TABLE IF NOT EXISTS amaliyah_metric_scores (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    snapshot_id INTEGER NOT NULL,
    metric_key TEXT NOT NULL CHECK (metric_key IN ('tilawah_tadabur', 'dzikir_pagi', 'dzikir_petang', 'tahajud')),
    completed_count INTEGER NOT NULL DEFAULT 0,
    available_count INTEGER NOT NULL DEFAULT 0,
    score REAL NOT NULL DEFAULT 0,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (snapshot_id) REFERENCES amaliyah_snapshots(id) ON DELETE CASCADE,
    UNIQUE(snapshot_id, metric_key)
);

CREATE INDEX IF NOT EXISTS idx_amaliyah_metric_scores_snapshot
ON amaliyah_metric_scores(snapshot_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_amaliyah_metric_scores_snapshot;
DROP TABLE IF EXISTS amaliyah_metric_scores;
DROP INDEX IF EXISTS idx_amaliyah_snapshots_member_period;
DROP INDEX IF EXISTS idx_amaliyah_snapshots_owner_period;
DROP TABLE IF EXISTS amaliyah_snapshots;
-- +goose StatementEnd
