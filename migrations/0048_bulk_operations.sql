-- +goose Up
CREATE TABLE bulk_operation_history (
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    admin_id       UUID NOT NULL,
    operation_type TEXT NOT NULL,
    parameters     JSONB NOT NULL DEFAULT '{}',
    filters        JSONB NOT NULL DEFAULT '{}',
    affected_count INT NOT NULL DEFAULT 0,
    status         TEXT NOT NULL DEFAULT 'completed',
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_bulk_ops_admin ON bulk_operation_history(admin_id);
CREATE INDEX idx_bulk_ops_created ON bulk_operation_history(created_at DESC);

-- +goose Down
DROP TABLE IF EXISTS bulk_operation_history;
