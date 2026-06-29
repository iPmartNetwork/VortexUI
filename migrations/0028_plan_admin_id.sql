-- +goose Up
ALTER TABLE plans ADD COLUMN IF NOT EXISTS admin_id UUID REFERENCES admins(id) ON DELETE CASCADE;
CREATE INDEX IF NOT EXISTS idx_plans_admin ON plans(admin_id);
-- Backfill legacy global plans to the primary sudo admin so they stay visible in the main panel shop.
UPDATE plans SET admin_id = (SELECT id FROM admins WHERE sudo = TRUE ORDER BY created_at LIMIT 1) WHERE admin_id IS NULL;

-- +goose Down
DROP INDEX IF EXISTS idx_plans_admin;
ALTER TABLE plans DROP COLUMN IF EXISTS admin_id;
