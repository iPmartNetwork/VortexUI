-- +goose Up
CREATE TABLE IF NOT EXISTS routing_packs (
    id          UUID PRIMARY KEY,
    name        TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    category    TEXT NOT NULL DEFAULT '',
    rules       JSONB NOT NULL DEFAULT '[]',  -- []domain.RoutingRule
    outbounds   JSONB NOT NULL DEFAULT '[]',  -- []domain.Outbound
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE TABLE IF NOT EXISTS routing_pack_selection (
    id      INTEGER PRIMARY KEY DEFAULT 1,    -- singleton for the global default
    pack_id TEXT NOT NULL DEFAULT ''
);
-- Per-subscription selection lives on the user; add a nullable-defaulted column.
ALTER TABLE users ADD COLUMN IF NOT EXISTS routing_pack_id TEXT NOT NULL DEFAULT '';

-- +goose Down
ALTER TABLE users DROP COLUMN IF EXISTS routing_pack_id;
DROP TABLE IF EXISTS routing_pack_selection;
DROP TABLE IF EXISTS routing_packs;
