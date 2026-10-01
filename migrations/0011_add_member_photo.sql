-- +goose Up
-- +goose StatementBegin
ALTER TABLE members ADD COLUMN photo_url TEXT NOT NULL DEFAULT '';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE members DROP COLUMN photo_url;
-- +goose StatementEnd
