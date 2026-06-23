-- +goose Up
-- Reseller ownership on users + per-admin inbound allowlist.
ALTER TABLE users ADD COLUMN IF NOT EXISTS admin_id UUID REFERENCES admins(id) ON DELETE SET NULL;
CREATE INDEX IF NOT EXISTS idx_users_admin_id ON users(admin_id);

CREATE TABLE IF NOT EXISTS admin_inbounds (
    admin_id   UUID NOT NULL REFERENCES admins(id) ON DELETE CASCADE,
    inbound_id UUID NOT NULL REFERENCES inbounds(id) ON DELETE CASCADE,
    PRIMARY KEY (admin_id, inbound_id)
);
CREATE INDEX IF NOT EXISTS idx_admin_inbounds_inbound ON admin_inbounds (inbound_id);

-- +goose Down
DROP INDEX IF EXISTS idx_admin_inbounds_inbound;
DROP TABLE IF EXISTS admin_inbounds;
DROP INDEX IF EXISTS idx_users_admin_id;
ALTER TABLE users DROP COLUMN IF EXISTS admin_id;
