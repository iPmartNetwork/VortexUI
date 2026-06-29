-- +goose Up
ALTER TABLE orders ADD COLUMN IF NOT EXISTS proof_image TEXT NOT NULL DEFAULT '';

-- +goose Down
ALTER TABLE orders DROP COLUMN IF EXISTS proof_image;
