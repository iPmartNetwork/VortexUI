-- +goose Up
-- Reseller wallet packages, manual/online deposits, and billing settings.

CREATE TABLE IF NOT EXISTS wallet_packages (
    id              UUID PRIMARY KEY,
    name            TEXT NOT NULL,
    description     TEXT NOT NULL DEFAULT '',
    traffic_bytes   BIGINT NOT NULL DEFAULT 0,
    user_credits    INT NOT NULL DEFAULT 0,
    price_amount    BIGINT NOT NULL DEFAULT 0,
    currency        TEXT NOT NULL DEFAULT 'IRR',
    methods         JSONB NOT NULL DEFAULT '[]',
    enabled         BOOLEAN NOT NULL DEFAULT TRUE,
    sort_order      INT NOT NULL DEFAULT 0,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_wallet_packages_enabled ON wallet_packages (enabled, sort_order, created_at DESC);

CREATE TABLE IF NOT EXISTS billing_settings (
    id                  INT PRIMARY KEY DEFAULT 1,
    card_number         TEXT NOT NULL DEFAULT '',
    card_holder         TEXT NOT NULL DEFAULT '',
    card_bank           TEXT NOT NULL DEFAULT '',
    crypto_addresses    JSONB NOT NULL DEFAULT '{}',
    manual_instructions TEXT NOT NULL DEFAULT '',
    CONSTRAINT billing_settings_singleton CHECK (id = 1)
);

INSERT INTO billing_settings (id) VALUES (1) ON CONFLICT (id) DO NOTHING;

CREATE TABLE IF NOT EXISTS wallet_deposits (
    id              UUID PRIMARY KEY,
    admin_id        UUID NOT NULL REFERENCES admins(id) ON DELETE CASCADE,
    package_id      UUID REFERENCES wallet_packages(id) ON DELETE SET NULL,
    method          TEXT NOT NULL,
    status          TEXT NOT NULL DEFAULT 'pending',
    amount          BIGINT NOT NULL DEFAULT 0,
    currency        TEXT NOT NULL DEFAULT 'IRR',
    traffic_bytes   BIGINT NOT NULL DEFAULT 0,
    user_credits    INT NOT NULL DEFAULT 0,
    gateway_id      TEXT NOT NULL DEFAULT '',
    tx_id           TEXT NOT NULL DEFAULT '',
    proof_image     TEXT NOT NULL DEFAULT '',
    reseller_note   TEXT NOT NULL DEFAULT '',
    admin_note      TEXT NOT NULL DEFAULT '',
    reviewer_id     UUID REFERENCES admins(id) ON DELETE SET NULL,
    reviewed_at     TIMESTAMPTZ,
    paid_at         TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_wallet_deposits_admin ON wallet_deposits (admin_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_wallet_deposits_status ON wallet_deposits (status, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_wallet_deposits_gateway ON wallet_deposits (gateway_id) WHERE gateway_id <> '';

-- +goose Down
DROP INDEX IF EXISTS idx_wallet_deposits_gateway;
DROP INDEX IF EXISTS idx_wallet_deposits_status;
DROP INDEX IF EXISTS idx_wallet_deposits_admin;
DROP TABLE IF EXISTS wallet_deposits;
DROP TABLE IF EXISTS billing_settings;
DROP INDEX IF EXISTS idx_wallet_packages_enabled;
DROP TABLE IF EXISTS wallet_packages;
