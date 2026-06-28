-- +goose Up
ALTER TABLE wallet_deposits ADD COLUMN IF NOT EXISTS crypto_coin TEXT NOT NULL DEFAULT '';

-- +goose Down
ALTER TABLE wallet_deposits DROP COLUMN IF EXISTS crypto_coin;
