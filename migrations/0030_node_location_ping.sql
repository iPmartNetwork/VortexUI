-- +goose Up
-- Node geography and latency for fleet telemetry on the command-tower dashboard.
ALTER TABLE nodes ADD COLUMN IF NOT EXISTS region TEXT NOT NULL DEFAULT '';
ALTER TABLE nodes ADD COLUMN IF NOT EXISTS country_code CHAR(2) NOT NULL DEFAULT '';
ALTER TABLE nodes ADD COLUMN IF NOT EXISTS ping_ms INTEGER NOT NULL DEFAULT 0;
ALTER TABLE nodes ADD COLUMN IF NOT EXISTS location_auto BOOLEAN NOT NULL DEFAULT TRUE;

-- +goose Down
ALTER TABLE nodes DROP COLUMN IF EXISTS location_auto;
ALTER TABLE nodes DROP COLUMN IF EXISTS ping_ms;
ALTER TABLE nodes DROP COLUMN IF EXISTS country_code;
ALTER TABLE nodes DROP COLUMN IF EXISTS region;
