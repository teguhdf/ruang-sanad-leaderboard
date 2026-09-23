-- +goose Up
-- +goose StatementBegin
-- The `id` column stores SHA-256(raw_session_id), NOT the raw session ID.
-- The raw ID (64 hex chars from crypto/rand) is sent to the client via cookie.
-- If the database leaks, the attacker cannot reconstruct valid cookies from
-- the hashed IDs — they would need to find a preimage for each SHA-256 hash.
CREATE TABLE IF NOT EXISTS sessions (
    id TEXT PRIMARY KEY,
    user_id INTEGER NOT NULL,
    data TEXT NOT NULL,
    expires_at DATETIME NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE INDEX idx_sessions_user_id ON sessions(user_id);
CREATE INDEX idx_sessions_expires_at ON sessions(expires_at);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_sessions_expires_at;
DROP INDEX IF EXISTS idx_sessions_user_id;
DROP TABLE IF EXISTS sessions;
-- +goose StatementEnd
