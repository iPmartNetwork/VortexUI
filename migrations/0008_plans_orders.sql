-- +goose Up
CREATE TABLE plans (
    id             UUID PRIMARY KEY,
    name           TEXT NOT NULL,
    description    TEXT NOT NULL DEFAULT '',
    data_limit     BIGINT NOT NULL DEFAULT 0,
    duration_days  INTEGER NOT NULL DEFAULT 30,
    device_limit   INTEGER NOT NULL DEFAULT 0,
    reset_strategy TEXT NOT NULL DEFAULT 'monthly',
    inbound_ids    JSONB NOT NULL DEFAULT '[]',
    price_toman    BIGINT NOT NULL DEFAULT 0,
    price_usd      DOUBLE PRECISION NOT NULL DEFAULT 0,
    max_users      INTEGER NOT NULL DEFAULT 0,
    enabled        BOOLEAN NOT NULL DEFAULT TRUE,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE orders (
    id          UUID PRIMARY KEY,
    user_id     UUID REFERENCES users(id) ON DELETE SET NULL,
    admin_id    UUID REFERENCES admins(id) ON DELETE SET NULL,
    plan_id     UUID NOT NULL REFERENCES plans(id),
    username    TEXT NOT NULL DEFAULT '',
    status      TEXT NOT NULL DEFAULT 'pending',
    gateway     TEXT NOT NULL DEFAULT '',
    gateway_id  TEXT NOT NULL DEFAULT '',
    amount      BIGINT NOT NULL DEFAULT 0,
    currency    TEXT NOT NULL DEFAULT 'IRR',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    paid_at     TIMESTAMPTZ
);

CREATE INDEX idx_orders_user ON orders(user_id);
CREATE INDEX idx_orders_status ON orders(status);

-- +goose Down
DROP TABLE IF EXISTS orders;
DROP TABLE IF EXISTS plans;
