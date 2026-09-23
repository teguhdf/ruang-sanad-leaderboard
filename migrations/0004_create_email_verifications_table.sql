-- +goose Up
-- +goose StatementBegin
-- The `token` column stores SHA-256(raw_token), NOT the raw token.
-- Same pattern as sessions and password_resets — a database leak cannot
-- be used to reconstruct valid verification URLs.
CREATE TABLE IF NOT EXISTS email_verifications (
    token      TEXT PRIMARY KEY,
    user_id    INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    email      TEXT NOT NULL,
    expires_at DATETIME NOT NULL,
    used       INTEGER NOT NULL DEFAULT 0,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_email_verifications_user_id ON email_verifications(user_id);
CREATE INDEX IF NOT EXISTS idx_email_verifications_expires_at ON email_verifications(expires_at);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_email_verifications_expires_at;
DROP INDEX IF EXISTS idx_email_verifications_user_id;
DROP TABLE IF EXISTS email_verifications;
-- +goose StatementEnd
