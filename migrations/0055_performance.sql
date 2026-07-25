-- +goose Up

-- Convert traffic_logs to hypertable (requires TimescaleDB extension).
-- +goose StatementBegin
DO $body$
BEGIN
    IF EXISTS (SELECT 1 FROM pg_extension WHERE extname = 'timescaledb') THEN
        IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'traffic_logs') THEN
            PERFORM create_hypertable('traffic_logs', 'created_at', if_not_exists => TRUE, migrate_data => TRUE);
            PERFORM add_retention_policy('traffic_logs', INTERVAL '90 days', if_not_exists => TRUE);
        END IF;
    END IF;
END $body$;
-- +goose StatementEnd

-- Cache invalidation log for subscription cache management.
CREATE TABLE IF NOT EXISTS cache_invalidation_log (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    cache_key   TEXT NOT NULL,
    reason      TEXT NOT NULL DEFAULT '',
    invalidated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_cache_invalidation_key ON cache_invalidation_log (cache_key, invalidated_at DESC);

-- Background job queue for async operations.
CREATE TABLE IF NOT EXISTS background_jobs (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    job_type    TEXT NOT NULL,
    payload     JSONB NOT NULL DEFAULT '{}',
    status      TEXT NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'running', 'completed', 'failed')),
    result      JSONB,
    error       TEXT,
    attempts    INT NOT NULL DEFAULT 0,
    max_retries INT NOT NULL DEFAULT 3,
    run_after   TIMESTAMPTZ NOT NULL DEFAULT now(),
    started_at  TIMESTAMPTZ,
    completed_at TIMESTAMPTZ,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_background_jobs_pending ON background_jobs (run_after) WHERE status = 'pending';
CREATE INDEX IF NOT EXISTS idx_background_jobs_status ON background_jobs (status, created_at DESC);

-- +goose Down
DROP TABLE IF EXISTS background_jobs;
DROP TABLE IF EXISTS cache_invalidation_log;
