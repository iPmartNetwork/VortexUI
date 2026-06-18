-- +goose Up
CREATE TABLE IF NOT EXISTS user_geo (
    user_id    UUID PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    country    TEXT NOT NULL DEFAULT '',
    ip         TEXT NOT NULL DEFAULT '',
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_user_geo_country ON user_geo(country);

-- +goose Down
DROP TABLE IF EXISTS user_geo;
