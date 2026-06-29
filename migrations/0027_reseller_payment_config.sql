-- +goose Up
CREATE TABLE IF NOT EXISTS reseller_payment_config (
    admin_id            UUID PRIMARY KEY REFERENCES admins(id) ON DELETE CASCADE,
    card_number         TEXT NOT NULL DEFAULT '',
    card_holder         TEXT NOT NULL DEFAULT '',
    card_bank           TEXT NOT NULL DEFAULT '',
    crypto_addresses    JSONB NOT NULL DEFAULT '{}',
    zarinpal_merchant_id TEXT NOT NULL DEFAULT '',
    manual_instructions TEXT NOT NULL DEFAULT '',
    enabled_methods     JSONB NOT NULL DEFAULT '[]'
);

-- +goose Down
DROP TABLE IF EXISTS reseller_payment_config;
