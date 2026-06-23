-- +goose Up
-- Wallet, sub-reseller hierarchy, portal branding, reseller webhooks, impersonation audit.

ALTER TABLE admins
    ADD COLUMN IF NOT EXISTS parent_admin_id UUID REFERENCES admins(id) ON DELETE SET NULL,
    ADD COLUMN IF NOT EXISTS wallet_traffic_bytes BIGINT NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS wallet_user_credits INT NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS webhook_url TEXT NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS webhook_secret TEXT NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS webhook_enabled BOOLEAN NOT NULL DEFAULT FALSE;

CREATE INDEX IF NOT EXISTS idx_admins_parent ON admins (parent_admin_id);

CREATE TABLE IF NOT EXISTS admin_wallet_ledger (
    id              UUID PRIMARY KEY,
    admin_id        UUID NOT NULL REFERENCES admins(id) ON DELETE CASCADE,
    delta_traffic   BIGINT NOT NULL DEFAULT 0,
    delta_users     INT NOT NULL DEFAULT 0,
    reason          TEXT NOT NULL,
    actor_admin_id  UUID REFERENCES admins(id) ON DELETE SET NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_wallet_ledger_admin ON admin_wallet_ledger (admin_id, created_at DESC);

CREATE TABLE IF NOT EXISTS portal_branding (
    admin_id      UUID PRIMARY KEY REFERENCES admins(id) ON DELETE CASCADE,
    panel_title   TEXT NOT NULL DEFAULT '',
    logo_url      TEXT NOT NULL DEFAULT '',
    accent_color  TEXT NOT NULL DEFAULT '#6366f1',
    footer_text   TEXT NOT NULL DEFAULT '',
    portal_slug   TEXT UNIQUE,
    custom_domain TEXT NOT NULL DEFAULT ''
);

ALTER TABLE audit_log
    ADD COLUMN IF NOT EXISTS impersonator_id UUID REFERENCES admins(id) ON DELETE SET NULL;

-- +goose Down
ALTER TABLE audit_log DROP COLUMN IF EXISTS impersonator_id;
DROP TABLE IF EXISTS portal_branding;
DROP INDEX IF EXISTS idx_wallet_ledger_admin;
DROP TABLE IF EXISTS admin_wallet_ledger;
DROP INDEX IF EXISTS idx_admins_parent;
ALTER TABLE admins
    DROP COLUMN IF EXISTS webhook_enabled,
    DROP COLUMN IF EXISTS webhook_secret,
    DROP COLUMN IF EXISTS webhook_url,
    DROP COLUMN IF EXISTS wallet_user_credits,
    DROP COLUMN IF EXISTS wallet_traffic_bytes,
    DROP COLUMN IF EXISTS parent_admin_id;
