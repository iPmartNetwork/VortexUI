-- +goose Up
-- Real download-throughput measurement for clean-IP scan results, run
-- on-demand per IP (separate from the bulk latency/loss scan).
ALTER TABLE clean_ip_scans ADD COLUMN IF NOT EXISTS throughput_mbps DOUBLE PRECISION NOT NULL DEFAULT 0;

-- +goose Down
ALTER TABLE clean_ip_scans DROP COLUMN IF EXISTS throughput_mbps;
