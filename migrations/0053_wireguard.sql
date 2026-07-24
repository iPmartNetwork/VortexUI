-- +goose Up

-- Extend wireguard_peers with per-peer settings and handshake/traffic stats.
ALTER TABLE wireguard_peers ADD COLUMN IF NOT EXISTS mtu INT NOT NULL DEFAULT 1420;
ALTER TABLE wireguard_peers ADD COLUMN IF NOT EXISTS dns TEXT NOT NULL DEFAULT '1.1.1.1';
ALTER TABLE wireguard_peers ADD COLUMN IF NOT EXISTS last_handshake TIMESTAMPTZ;
ALTER TABLE wireguard_peers ADD COLUMN IF NOT EXISTS tx_bytes BIGINT NOT NULL DEFAULT 0;
ALTER TABLE wireguard_peers ADD COLUMN IF NOT EXISTS rx_bytes BIGINT NOT NULL DEFAULT 0;

-- WireGuard mesh network for site-to-site tunnels between nodes.
CREATE TABLE wireguard_mesh (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name        TEXT NOT NULL,
    cidr        TEXT NOT NULL DEFAULT '10.10.0.0/16',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE wireguard_mesh_peers (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    mesh_id     UUID NOT NULL REFERENCES wireguard_mesh(id) ON DELETE CASCADE,
    node_id     UUID NOT NULL REFERENCES nodes(id) ON DELETE CASCADE,
    public_key  TEXT NOT NULL,
    private_key TEXT NOT NULL,
    endpoint    TEXT NOT NULL DEFAULT '',
    address     TEXT NOT NULL,
    keepalive   INT NOT NULL DEFAULT 25,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (mesh_id, node_id)
);

CREATE INDEX idx_wireguard_mesh_peers_mesh ON wireguard_mesh_peers (mesh_id);

-- +goose Down
DROP TABLE IF EXISTS wireguard_mesh_peers;
DROP TABLE IF EXISTS wireguard_mesh;
ALTER TABLE wireguard_peers DROP COLUMN IF EXISTS mtu;
ALTER TABLE wireguard_peers DROP COLUMN IF EXISTS dns;
ALTER TABLE wireguard_peers DROP COLUMN IF EXISTS last_handshake;
ALTER TABLE wireguard_peers DROP COLUMN IF EXISTS tx_bytes;
ALTER TABLE wireguard_peers DROP COLUMN IF EXISTS rx_bytes;
