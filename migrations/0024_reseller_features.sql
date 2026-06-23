-- +goose Up
-- Per-reseller feature toggles controlled by sudo admins.

ALTER TABLE admins
    ADD COLUMN IF NOT EXISTS allow_sub_resellers BOOLEAN NOT NULL DEFAULT FALSE,
    ADD COLUMN IF NOT EXISTS allow_user_backup BOOLEAN NOT NULL DEFAULT FALSE,
    ADD COLUMN IF NOT EXISTS reseller_settings JSONB NOT NULL DEFAULT '{}';

-- +goose Down
ALTER TABLE admins
    DROP COLUMN IF EXISTS reseller_settings,
    DROP COLUMN IF EXISTS allow_user_backup,
    DROP COLUMN IF EXISTS allow_sub_resellers;
