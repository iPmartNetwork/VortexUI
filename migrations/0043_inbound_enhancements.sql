-- +goose Up
-- Inbound enhancements: notes, port ranges, per-inbound traffic stats.

ALTER TABLE inbounds
    ADD COLUMN IF NOT EXISTS notes TEXT NOT NULL DEFAULT '';

ALTER TABLE inbounds
    ADD COLUMN IF NOT EXISTS port_end INTEGER NOT NULL DEFAULT 0;

CREATE TABLE IF NOT EXISTS inbound_traffic_stats (
    inbound_id UUID NOT NULL REFERENCES inbounds(id) ON DELETE CASCADE,
    date       DATE NOT NULL DEFAULT CURRENT_DATE,
    upload     BIGINT NOT NULL DEFAULT 0,
    download   BIGINT NOT NULL DEFAULT 0,
    PRIMARY KEY (inbound_id, date)
);

CREATE INDEX IF NOT EXISTS idx_inbound_traffic_stats_date
    ON inbound_traffic_stats (inbound_id, date DESC);

-- +goose Down
DROP INDEX IF EXISTS idx_inbound_traffic_stats_date;
DROP TABLE IF EXISTS inbound_traffic_stats;
ALTER TABLE inbounds DROP COLUMN IF EXISTS port_end;
ALTER TABLE inbounds DROP COLUMN IF EXISTS notes;
