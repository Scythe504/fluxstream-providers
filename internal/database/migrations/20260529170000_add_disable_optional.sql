-- +goose Up
-- +goose StatementBegin
ALTER TABLE providers ADD COLUMN disable_optional BOOLEAN DEFAULT FALSE;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
-- SQLite doesn't easily support dropping columns in older versions, but if needed, we can define a recreation, or just leave it.
-- +goose StatementEnd
