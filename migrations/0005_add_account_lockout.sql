-- +goose Up
-- +goose StatementBegin
-- Account lockout: track failed login attempts and lock accounts after
-- repeated failures. locked_until is NULL when the account is not locked.
ALTER TABLE users ADD COLUMN failed_login_attempts INTEGER NOT NULL DEFAULT 0;
ALTER TABLE users ADD COLUMN locked_until DATETIME;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE users DROP COLUMN locked_until;
ALTER TABLE users DROP COLUMN failed_login_attempts;
-- +goose StatementEnd
