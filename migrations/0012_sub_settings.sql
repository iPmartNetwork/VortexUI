-- +goose Up
CREATE TABLE IF NOT EXISTS sub_settings (
    id              INTEGER PRIMARY KEY DEFAULT 1,
    update_interval INTEGER NOT NULL DEFAULT 12,
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT sub_settings_singleton CHECK (id = 1)
);
INSERT INTO sub_settings (id, update_interval) VALUES (1, 12) ON CONFLICT (id) DO NOTHING;

-- +goose Down
DROP TABLE IF EXISTS sub_settings;
