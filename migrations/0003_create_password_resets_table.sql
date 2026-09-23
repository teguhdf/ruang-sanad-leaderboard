-- +goose Up
-- +goose StatementBegin
-- The `token` column stores SHA-256(raw_token), NOT the raw token.
-- The raw token (64 hex chars from crypto/rand) is embedded in the reset URL
-- sent via email. If the database leaks, the attacker cannot reconstruct
-- valid reset URLs from the hashed tokens.
CREATE TABLE IF NOT EXISTS password_resets (
    token      TEXT PRIMARY KEY,
    user_id    INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    email      TEXT NOT NULL,
    expires_at DATETIME NOT NULL,
    used       INTEGER NOT NULL DEFAULT 0,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_password_resets_user_id ON password_resets(user_id);
CREATE INDEX IF NOT EXISTS idx_password_resets_expires_at ON password_resets(expires_at);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_password_resets_expires_at;
DROP INDEX IF EXISTS idx_password_resets_user_id;
DROP TABLE IF EXISTS password_resets;
-- +goose StatementEnd
