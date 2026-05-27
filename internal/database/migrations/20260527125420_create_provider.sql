-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS providers (
  id TEXT PRIMARY KEY,
  provider_name TEXT UNIQUE NOT NULL,
  provider_url TEXT UNIQUE NOT NULL,
  verification_pending BOOLEAN DEFAULT TRUE,
  version TEXT,
  verified_at INTEGER,
  provider_type TEXT,
  created_at INTEGER NOT NULL
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS providers;
-- +goose StatementEnd
