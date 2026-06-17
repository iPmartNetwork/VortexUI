-- +goose Up
-- Adds ownership: which admin created each user (reseller model).
ALTER TABLE users ADD COLUMN IF NOT EXISTS admin_id UUID REFERENCES admins(id) ON DELETE SET NULL;
CREATE INDEX IF NOT EXISTS idx_users_admin_id ON users(admin_id);

-- +goose Down
DROP INDEX IF EXISTS idx_users_admin_id;
ALTER TABLE users DROP COLUMN IF EXISTS admin_id;
