-- +goose Up
-- HWID & Device Management: track user devices by hardware ID.

CREATE TABLE IF NOT EXISTS user_devices (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id     UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    hwid        TEXT NOT NULL,
    os          TEXT NOT NULL DEFAULT 'unknown',
    last_seen   TIMESTAMPTZ NOT NULL DEFAULT now(),
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE(user_id, hwid)
);

CREATE INDEX IF NOT EXISTS idx_user_devices_user_id ON user_devices(user_id);
CREATE INDEX IF NOT EXISTS idx_user_devices_hwid ON user_devices(hwid);

ALTER TABLE users ADD COLUMN IF NOT EXISTS device_lock BOOLEAN NOT NULL DEFAULT false;

-- +goose Down
ALTER TABLE users DROP COLUMN IF EXISTS device_lock;
DROP TABLE IF EXISTS user_devices;
