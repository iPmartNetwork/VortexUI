-- +goose Up
ALTER TABLE inbounds ADD COLUMN IF NOT EXISTS speed_limit BIGINT NOT NULL DEFAULT 0;
ALTER TABLE inbounds ADD COLUMN IF NOT EXISTS geo_policy JSONB;

-- +goose Down
ALTER TABLE inbounds DROP COLUMN IF EXISTS geo_policy;
ALTER TABLE inbounds DROP COLUMN IF EXISTS speed_limit;
