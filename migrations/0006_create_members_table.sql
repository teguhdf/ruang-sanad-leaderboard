-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS members (
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

CREATE INDEX idx_members_owner_user_id ON members(owner_user_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_members_owner_user_id;
DROP TABLE IF EXISTS members;
-- +goose StatementEnd
