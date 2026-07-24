-- +goose Up
-- Format Variables & Subscription Intelligence: add template columns to sub_settings
-- for configurable subscription output formatting.

ALTER TABLE sub_settings
    ADD COLUMN IF NOT EXISTS profile_title_template TEXT NOT NULL DEFAULT '{USERNAME}',
    ADD COLUMN IF NOT EXISTS remark_template TEXT NOT NULL DEFAULT '{PROTOCOL} - {NODE_NAME}',
    ADD COLUMN IF NOT EXISTS address_template TEXT NOT NULL DEFAULT '';

-- +goose Down
ALTER TABLE sub_settings
    DROP COLUMN IF EXISTS address_template,
    DROP COLUMN IF EXISTS remark_template,
    DROP COLUMN IF EXISTS profile_title_template;
