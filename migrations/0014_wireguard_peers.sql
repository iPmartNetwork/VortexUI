-- +goose Up
CREATE TABLE IF NOT EXISTS wireguard_peers (
    inbound_id  UUID NOT NULL REFERENCES inbounds(id) ON DELETE CASCADE,
    user_id     UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    private_key TEXT NOT NULL,
    public_key  TEXT NOT NULL,
    address     TEXT NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (inbound_id, user_id)
);

-- +goose Down
DROP TABLE IF EXISTS wireguard_peers;
