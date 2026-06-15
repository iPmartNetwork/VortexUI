-- +goose Up
ALTER TABLE nodes ADD COLUMN IF NOT EXISTS endpoint TEXT NOT NULL DEFAULT '';

-- +goose Down
ALTER TABLE nodes DROP COLUMN IF EXISTS endpoint;
