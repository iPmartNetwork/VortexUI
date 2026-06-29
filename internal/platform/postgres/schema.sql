-- Schema mirror for sqlc's static type inference only. The runtime source of
-- truth is migrations/ (which also adds TimescaleDB hypertables). Keep the table
-- shapes here in sync with migrations/0001_init.sql.

CREATE TABLE roles (
    id          UUID PRIMARY KEY,
    name        TEXT NOT NULL UNIQUE,
    permissions JSONB NOT NULL DEFAULT '[]'
);

CREATE TABLE admins (
    id            UUID PRIMARY KEY,
    username      TEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    sudo          BOOLEAN NOT NULL DEFAULT FALSE,
    role_id       UUID REFERENCES roles(id) ON DELETE SET NULL,
    totp_secret   TEXT NOT NULL DEFAULT '',
    totp_enabled  BOOLEAN NOT NULL DEFAULT FALSE,
    user_quota          INTEGER NOT NULL DEFAULT 0,
    traffic_quota       BIGINT  NOT NULL DEFAULT 0,
    traffic_quota_mode  TEXT NOT NULL DEFAULT 'allocated',
    parent_admin_id     UUID REFERENCES admins(id) ON DELETE SET NULL,
    wallet_traffic_bytes BIGINT NOT NULL DEFAULT 0,
    wallet_user_credits  INT NOT NULL DEFAULT 0,
    webhook_url          TEXT NOT NULL DEFAULT '',
    webhook_secret       TEXT NOT NULL DEFAULT '',
    webhook_enabled      BOOLEAN NOT NULL DEFAULT FALSE,
    policy_max_data_limit BIGINT NOT NULL DEFAULT 0,
    policy_max_expire_days INT NOT NULL DEFAULT 0,
    policy_allow_bulk_delete BOOLEAN NOT NULL DEFAULT TRUE,
    policy_allow_bulk_create BOOLEAN NOT NULL DEFAULT TRUE,
    suspended            BOOLEAN NOT NULL DEFAULT FALSE,
    suspended_at         TIMESTAMPTZ,
    suspend_reason       TEXT NOT NULL DEFAULT '',
    auto_suspend_enabled BOOLEAN NOT NULL DEFAULT TRUE,
    ip_violation_suspend_threshold INT NOT NULL DEFAULT 0,
    suspend_grace_minutes INT NOT NULL DEFAULT 60,
    quota_breached_at    TIMESTAMPTZ,
    allow_sub_resellers  BOOLEAN NOT NULL DEFAULT FALSE,
    allow_user_backup    BOOLEAN NOT NULL DEFAULT FALSE,
    reseller_settings    JSONB NOT NULL DEFAULT '{}',
    last_login          TIMESTAMPTZ,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE nodes (
    id            UUID PRIMARY KEY,
    name          TEXT NOT NULL UNIQUE,
    address       TEXT NOT NULL,
    core          TEXT NOT NULL DEFAULT 'xray',
    status        TEXT NOT NULL DEFAULT 'disconnected',
    usage_ratio   DOUBLE PRECISION NOT NULL DEFAULT 1,
    endpoint      TEXT NOT NULL DEFAULT '',
    last_seen     TIMESTAMPTZ,
    cpu_percent   DOUBLE PRECISION NOT NULL DEFAULT 0,
    mem_percent   DOUBLE PRECISION NOT NULL DEFAULT 0,
    disk_percent  DOUBLE PRECISION NOT NULL DEFAULT 0,
    core_running  BOOLEAN NOT NULL DEFAULT FALSE,
    connections   INTEGER NOT NULL DEFAULT 0,
    core_version  TEXT NOT NULL DEFAULT '',
    agent_version TEXT NOT NULL DEFAULT '',
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE inbounds (
    id                 UUID PRIMARY KEY,
    node_id            UUID NOT NULL REFERENCES nodes(id) ON DELETE CASCADE,
    tag                TEXT NOT NULL,
    protocol           TEXT NOT NULL,
    listen             TEXT NOT NULL DEFAULT '',
    port               INTEGER NOT NULL,
    network            TEXT NOT NULL DEFAULT 'tcp',
    security           TEXT NOT NULL DEFAULT 'none',
    sni                JSONB NOT NULL DEFAULT '[]',
    path               TEXT NOT NULL DEFAULT '',
    host               JSONB NOT NULL DEFAULT '[]',
    flow               TEXT NOT NULL DEFAULT '',
    evasion_profile_id UUID,
    raw                JSONB NOT NULL DEFAULT '{}',
    enabled            BOOLEAN NOT NULL DEFAULT TRUE,
    speed_limit        BIGINT NOT NULL DEFAULT 0,
    geo_policy         JSONB,
    UNIQUE (node_id, tag)
);

CREATE TABLE users (
    id             UUID PRIMARY KEY,
    username       TEXT NOT NULL UNIQUE,
    status         TEXT NOT NULL DEFAULT 'active',
    note           TEXT NOT NULL DEFAULT '',
    data_limit     BIGINT NOT NULL DEFAULT 0,
    used_traffic   BIGINT NOT NULL DEFAULT 0,
    expire_at      TIMESTAMPTZ,
    on_hold_expire BIGINT,
    reset_strategy TEXT NOT NULL DEFAULT 'no_reset',
    last_reset     TIMESTAMPTZ,
    device_limit   INTEGER NOT NULL DEFAULT 0,
    allowed_hwids  JSONB NOT NULL DEFAULT '[]',
    vmess_uuid     UUID NOT NULL,
    vless_uuid     UUID NOT NULL,
    trojan_pass    TEXT NOT NULL DEFAULT '',
    ss_password    TEXT NOT NULL DEFAULT '',
    ss_method      TEXT NOT NULL DEFAULT 'aes-128-gcm',
    sub_token      TEXT NOT NULL UNIQUE,
    routing_pack_id TEXT NOT NULL DEFAULT '',
    admin_id       UUID REFERENCES admins(id) ON DELETE SET NULL,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_users_admin_id ON users (admin_id);

CREATE TABLE admin_inbounds (
    admin_id   UUID NOT NULL REFERENCES admins(id) ON DELETE CASCADE,
    inbound_id UUID NOT NULL REFERENCES inbounds(id) ON DELETE CASCADE,
    PRIMARY KEY (admin_id, inbound_id)
);

CREATE INDEX idx_admin_inbounds_inbound ON admin_inbounds (inbound_id);

CREATE TABLE plans (
    id             UUID PRIMARY KEY,
    admin_id       UUID REFERENCES admins(id) ON DELETE CASCADE,
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

CREATE INDEX idx_plans_admin ON plans(admin_id);

CREATE TABLE admin_plans (
    admin_id UUID NOT NULL REFERENCES admins(id) ON DELETE CASCADE,
    plan_id  UUID NOT NULL REFERENCES plans(id) ON DELETE CASCADE,
    PRIMARY KEY (admin_id, plan_id)
);

CREATE TABLE admin_nodes (
    admin_id UUID NOT NULL REFERENCES admins(id) ON DELETE CASCADE,
    node_id  UUID NOT NULL REFERENCES nodes(id) ON DELETE CASCADE,
    PRIMARY KEY (admin_id, node_id)
);

CREATE TABLE admin_quota_notify_config (
    id               INT PRIMARY KEY DEFAULT 1,
    enabled          BOOLEAN NOT NULL DEFAULT FALSE,
    threshold_pct    INT[] NOT NULL DEFAULT '{80,90,100}',
    notify_telegram  BOOLEAN NOT NULL DEFAULT TRUE,
    webhook_url      TEXT NOT NULL DEFAULT '',
    cooldown_minutes INT NOT NULL DEFAULT 1440
);

CREATE TABLE admin_quota_notify_events (
    id         UUID PRIMARY KEY,
    admin_id   UUID NOT NULL REFERENCES admins(id) ON DELETE CASCADE,
    threshold  INT NOT NULL,
    metric     TEXT NOT NULL,
    usage_pct  INT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE admin_wallet_ledger (
    id             UUID PRIMARY KEY,
    admin_id       UUID NOT NULL REFERENCES admins(id) ON DELETE CASCADE,
    delta_traffic  BIGINT NOT NULL DEFAULT 0,
    delta_users    INT NOT NULL DEFAULT 0,
    reason         TEXT NOT NULL,
    actor_admin_id UUID REFERENCES admins(id) ON DELETE SET NULL,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE portal_branding (
    admin_id      UUID PRIMARY KEY REFERENCES admins(id) ON DELETE CASCADE,
    panel_title   TEXT NOT NULL DEFAULT '',
    logo_url      TEXT NOT NULL DEFAULT '',
    accent_color  TEXT NOT NULL DEFAULT '#6366f1',
    footer_text   TEXT NOT NULL DEFAULT '',
    portal_slug   TEXT UNIQUE,
    custom_domain TEXT NOT NULL DEFAULT ''
);

CREATE TABLE user_inbounds (
    user_id    UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    inbound_id UUID NOT NULL REFERENCES inbounds(id) ON DELETE CASCADE,
    PRIMARY KEY (user_id, inbound_id)
);

CREATE INDEX idx_user_inbounds_inbound ON user_inbounds (inbound_id);

CREATE TABLE traffic_points (
    time    TIMESTAMPTZ NOT NULL,
    user_id UUID NOT NULL,
    node_id UUID NOT NULL,
    up      BIGINT NOT NULL,
    down    BIGINT NOT NULL
);

CREATE TABLE audit_log (
    id              UUID PRIMARY KEY,
    time            TIMESTAMPTZ NOT NULL DEFAULT now(),
    admin_id        UUID,
    impersonator_id UUID,
    method          TEXT NOT NULL,
    path            TEXT NOT NULL,
    status          INTEGER NOT NULL,
    ip              TEXT NOT NULL DEFAULT ''
);

CREATE TABLE api_tokens (
    id           UUID PRIMARY KEY,
    name         TEXT NOT NULL,
    token_hash   TEXT NOT NULL UNIQUE,
    admin_id     UUID NOT NULL REFERENCES admins(id) ON DELETE CASCADE,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    last_used_at TIMESTAMPTZ
);

CREATE INDEX idx_api_tokens_admin ON api_tokens (admin_id);


CREATE TABLE outbounds (
    id        UUID PRIMARY KEY,
    node_id   UUID NOT NULL REFERENCES nodes(id) ON DELETE CASCADE,
    tag       TEXT NOT NULL,
    protocol  TEXT NOT NULL,
    address   TEXT NOT NULL DEFAULT '',
    port      INTEGER NOT NULL DEFAULT 0,
    uuid      TEXT NOT NULL DEFAULT '',
    password  TEXT NOT NULL DEFAULT '',
    username  TEXT NOT NULL DEFAULT '',
    method    TEXT NOT NULL DEFAULT '',
    flow      TEXT NOT NULL DEFAULT '',
    network   TEXT NOT NULL DEFAULT '',
    security  TEXT NOT NULL DEFAULT '',
    sni       TEXT NOT NULL DEFAULT '',
    path      TEXT NOT NULL DEFAULT '',
    host      TEXT NOT NULL DEFAULT '',
    raw       JSONB NOT NULL DEFAULT '{}',
    enabled   BOOLEAN NOT NULL DEFAULT TRUE,
    UNIQUE (node_id, tag)
);

CREATE TABLE routing_rules (
    id           UUID PRIMARY KEY,
    node_id      UUID NOT NULL REFERENCES nodes(id) ON DELETE CASCADE,
    priority     INTEGER NOT NULL DEFAULT 0,
    name         TEXT NOT NULL DEFAULT '',
    inbound_tags JSONB NOT NULL DEFAULT '[]',
    domains      JSONB NOT NULL DEFAULT '[]',
    ip           JSONB NOT NULL DEFAULT '[]',
    port         TEXT NOT NULL DEFAULT '',
    protocols    JSONB NOT NULL DEFAULT '[]',
    network      TEXT NOT NULL DEFAULT '',
    outbound_tag TEXT NOT NULL DEFAULT '',
    balancer_tag TEXT NOT NULL DEFAULT '',
    enabled      BOOLEAN NOT NULL DEFAULT TRUE
);

CREATE INDEX idx_routing_rules_node ON routing_rules (node_id, priority);

CREATE TABLE balancers (
    id             UUID PRIMARY KEY,
    node_id        UUID NOT NULL REFERENCES nodes(id) ON DELETE CASCADE,
    tag            TEXT NOT NULL,
    selectors      JSONB NOT NULL DEFAULT '[]',
    strategy       TEXT NOT NULL DEFAULT 'random',
    observe        BOOLEAN NOT NULL DEFAULT FALSE,
    probe_url      TEXT NOT NULL DEFAULT '',
    probe_interval TEXT NOT NULL DEFAULT '',
    enabled        BOOLEAN NOT NULL DEFAULT TRUE,
    UNIQUE (node_id, tag)
);

-- v1.2.0: Self-service portal tickets
CREATE TABLE tickets (
    id         UUID PRIMARY KEY,
    user_id    UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    subject    TEXT NOT NULL,
    status     TEXT NOT NULL DEFAULT 'open',
    priority   TEXT NOT NULL DEFAULT 'medium',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_tickets_user ON tickets (user_id);
CREATE INDEX idx_tickets_status ON tickets (status);

CREATE TABLE ticket_messages (
    id         UUID PRIMARY KEY,
    ticket_id  UUID NOT NULL REFERENCES tickets(id) ON DELETE CASCADE,
    sender     TEXT NOT NULL,
    sender_id  UUID NOT NULL,
    body       TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_ticket_messages_ticket ON ticket_messages (ticket_id);

-- v1.2.0: Reality Scanner results cache
CREATE TABLE reality_scans (
    id         UUID PRIMARY KEY,
    node_id    UUID NOT NULL REFERENCES nodes(id) ON DELETE CASCADE,
    sni        TEXT NOT NULL,
    latency_ms INTEGER NOT NULL DEFAULT 0,
    score      INTEGER NOT NULL DEFAULT 0,
    valid      BOOLEAN NOT NULL DEFAULT TRUE,
    scanned_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_reality_scans_node ON reality_scans (node_id, score DESC);

-- panel-feature-parity: Subscription Hosts (Marzban-style per-inbound overrides)
CREATE TABLE sub_hosts (
    id             UUID PRIMARY KEY,
    inbound_id     UUID NOT NULL,
    remark         TEXT NOT NULL DEFAULT '',
    address        TEXT NOT NULL DEFAULT '',
    port           INTEGER,
    sni            TEXT NOT NULL DEFAULT '',
    host_header    TEXT NOT NULL DEFAULT '',
    path           TEXT NOT NULL DEFAULT '',
    alpn           TEXT NOT NULL DEFAULT '',
    fingerprint    TEXT NOT NULL DEFAULT '',
    security       TEXT NOT NULL DEFAULT 'inbound_default',
    allow_insecure BOOLEAN NOT NULL DEFAULT FALSE,
    mux_enable     BOOLEAN NOT NULL DEFAULT FALSE,
    fragment       TEXT NOT NULL DEFAULT '',
    priority       INTEGER NOT NULL DEFAULT 0,
    enabled        BOOLEAN NOT NULL DEFAULT TRUE,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_sub_hosts_inbound ON sub_hosts (inbound_id, priority);

-- panel-feature-parity: Smart-routing rule packs (reusable routing rule sets)
CREATE TABLE routing_packs (
    id          UUID PRIMARY KEY,
    name        TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    category    TEXT NOT NULL DEFAULT '',
    rules       JSONB NOT NULL DEFAULT '[]',
    outbounds   JSONB NOT NULL DEFAULT '[]',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE TABLE routing_pack_selection (
    id      INTEGER PRIMARY KEY DEFAULT 1,
    pack_id TEXT NOT NULL DEFAULT ''
);

-- v1.2.0: Smart Quota tiers
CREATE TABLE quota_policies (
    id              UUID PRIMARY KEY,
    name            TEXT NOT NULL UNIQUE,
    tiers           JSONB NOT NULL DEFAULT '[]',
    enabled         BOOLEAN NOT NULL DEFAULT TRUE,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- v1.2.0: CDN/Relay chains
CREATE TABLE relay_chains (
    id         UUID PRIMARY KEY,
    name       TEXT NOT NULL,
    node_id    UUID NOT NULL REFERENCES nodes(id) ON DELETE CASCADE,
    hops       JSONB NOT NULL DEFAULT '[]',
    enabled    BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_relay_chains_node ON relay_chains (node_id);

-- v1.2.0: Decoy website configuration
CREATE TABLE decoy_sites (
    id          UUID PRIMARY KEY,
    node_id     UUID REFERENCES nodes(id) ON DELETE CASCADE,
    mode        TEXT NOT NULL DEFAULT 'proxy',
    target_url  TEXT NOT NULL DEFAULT '',
    static_html TEXT NOT NULL DEFAULT '',
    enabled     BOOLEAN NOT NULL DEFAULT TRUE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- v1.2.0: Analytics geo breakdown cache
CREATE TABLE traffic_geo (
    time       TIMESTAMPTZ NOT NULL,
    node_id    UUID NOT NULL,
    country    TEXT NOT NULL DEFAULT '',
    connections INTEGER NOT NULL DEFAULT 0,
    bytes_up   BIGINT NOT NULL DEFAULT 0,
    bytes_down BIGINT NOT NULL DEFAULT 0
);

-- v1.2.0: Per-user geo (country) resolved from subscription-fetch IPs, used to
-- compute the "Traffic by Country" breakdown by joining with traffic_points.
CREATE TABLE user_geo (
    user_id    UUID PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    country    TEXT NOT NULL DEFAULT '',
    ip         TEXT NOT NULL DEFAULT '',
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_user_geo_country ON user_geo(country);

-- v1.2.0: Node auto-migration
CREATE TABLE migration_events (
    id           UUID PRIMARY KEY,
    user_id      UUID,
    username     TEXT NOT NULL DEFAULT '',
    from_node_id UUID NOT NULL,
    to_node_id   UUID NOT NULL,
    reason       TEXT NOT NULL DEFAULT '',
    status       TEXT NOT NULL DEFAULT 'pending',
    error        TEXT NOT NULL DEFAULT '',
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE migration_policy (
    id                    INTEGER PRIMARY KEY DEFAULT 1,
    enabled               BOOLEAN NOT NULL DEFAULT FALSE,
    health_check_interval INTEGER NOT NULL DEFAULT 30,
    unhealthy_threshold   INTEGER NOT NULL DEFAULT 3,
    cpu_threshold         DOUBLE PRECISION NOT NULL DEFAULT 90,
    mem_threshold         DOUBLE PRECISION NOT NULL DEFAULT 90,
    packet_loss_max       DOUBLE PRECISION NOT NULL DEFAULT 10,
    migrate_back          BOOLEAN NOT NULL DEFAULT TRUE
);

-- v1.2.0: Active probing protection
CREATE TABLE probing_policy (
    id               INTEGER PRIMARY KEY DEFAULT 1,
    enabled          BOOLEAN NOT NULL DEFAULT FALSE,
    action           TEXT NOT NULL DEFAULT 'block',
    block_duration   INTEGER NOT NULL DEFAULT 3600,
    max_probe_per_min INTEGER NOT NULL DEFAULT 5,
    whitelisted_ips  JSONB NOT NULL DEFAULT '[]',
    honeypot_html    TEXT NOT NULL DEFAULT '',
    notify_telegram  BOOLEAN NOT NULL DEFAULT TRUE
);

CREATE TABLE probe_events (
    id          UUID PRIMARY KEY,
    source_ip   TEXT NOT NULL,
    port        INTEGER NOT NULL DEFAULT 0,
    method      TEXT NOT NULL DEFAULT '',
    fingerprint TEXT NOT NULL DEFAULT '',
    action      TEXT NOT NULL DEFAULT 'log',
    node_id     UUID,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_probe_events_ip ON probe_events (source_ip);

CREATE TABLE blocked_ips (
    ip         TEXT PRIMARY KEY,
    reason     TEXT NOT NULL DEFAULT '',
    blocked_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    expires_at TIMESTAMPTZ NOT NULL
);

-- v1.2.0: Family/Group subscriptions
CREATE TABLE family_groups (
    id           UUID PRIMARY KEY,
    name         TEXT NOT NULL,
    owner_id     UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    data_limit   BIGINT NOT NULL DEFAULT 0,
    used_traffic BIGINT NOT NULL DEFAULT 0,
    max_members  INTEGER NOT NULL DEFAULT 5,
    member_quota BIGINT NOT NULL DEFAULT 0,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_family_groups_owner ON family_groups (owner_id);

CREATE TABLE family_members (
    id           UUID PRIMARY KEY,
    group_id     UUID NOT NULL REFERENCES family_groups(id) ON DELETE CASCADE,
    user_id      UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    used_traffic BIGINT NOT NULL DEFAULT 0,
    label        TEXT NOT NULL DEFAULT '',
    joined_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (group_id, user_id)
);

CREATE INDEX idx_family_members_group ON family_members (group_id);

-- v1.2.0: Invite/Referral system
CREATE TABLE referral_config (
    id            INTEGER PRIMARY KEY DEFAULT 1,
    enabled       BOOLEAN NOT NULL DEFAULT FALSE,
    reward_type   TEXT NOT NULL DEFAULT 'data',
    reward_amount BIGINT NOT NULL DEFAULT 1073741824,
    max_referrals INTEGER NOT NULL DEFAULT 0,
    require_paid  BOOLEAN NOT NULL DEFAULT FALSE
);

CREATE TABLE referral_codes (
    id         UUID PRIMARY KEY,
    user_id    UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    code       TEXT NOT NULL UNIQUE,
    uses       INTEGER NOT NULL DEFAULT 0,
    max_uses   INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_referral_codes_user ON referral_codes (user_id);

CREATE TABLE referral_events (
    id             UUID PRIMARY KEY,
    referrer_id    UUID NOT NULL,
    referred_id    UUID NOT NULL,
    code_used      TEXT NOT NULL,
    reward_type    TEXT NOT NULL DEFAULT 'data',
    reward_amount  BIGINT NOT NULL DEFAULT 0,
    reward_applied BOOLEAN NOT NULL DEFAULT FALSE,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- v1.2.0: DNS-over-HTTPS
CREATE TABLE doh_config (
    id               INTEGER PRIMARY KEY DEFAULT 1,
    enabled          BOOLEAN NOT NULL DEFAULT FALSE,
    listen_addr      TEXT NOT NULL DEFAULT ':8053',
    upstream_dns     JSONB NOT NULL DEFAULT '["1.1.1.1","8.8.8.8"]',
    block_ads        BOOLEAN NOT NULL DEFAULT FALSE,
    block_malware    BOOLEAN NOT NULL DEFAULT TRUE,
    custom_blocklist JSONB NOT NULL DEFAULT '[]',
    log_queries      BOOLEAN NOT NULL DEFAULT FALSE,
    cache_ttl        INTEGER NOT NULL DEFAULT 300
);

CREATE TABLE doh_query_logs (
    domain     TEXT NOT NULL,
    type       TEXT NOT NULL DEFAULT 'A',
    client_ip  TEXT NOT NULL DEFAULT '',
    blocked    BOOLEAN NOT NULL DEFAULT FALSE,
    latency_ms INTEGER NOT NULL DEFAULT 0,
    timestamp  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_doh_query_logs_time ON doh_query_logs (timestamp DESC);

-- v1.2.0: SNI Domains
CREATE TABLE sni_domains (
    id          UUID PRIMARY KEY,
    inbound_id  UUID NOT NULL,
    domain      TEXT NOT NULL,
    auto_cert   BOOLEAN NOT NULL DEFAULT TRUE,
    cert_status TEXT NOT NULL DEFAULT 'pending',
    expires_at  TIMESTAMPTZ,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_sni_domains_inbound ON sni_domains (inbound_id);

-- v1.2.0: SSL Certificates
CREATE TABLE ssl_certificates (
    id         UUID PRIMARY KEY,
    domain     TEXT NOT NULL,
    wildcard   BOOLEAN NOT NULL DEFAULT FALSE,
    issuer     TEXT NOT NULL DEFAULT 'letsencrypt',
    status     TEXT NOT NULL DEFAULT 'pending',
    auto_renew BOOLEAN NOT NULL DEFAULT TRUE,
    expires_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- v1.2.0: SNI Routes
CREATE TABLE sni_routes (
    id          UUID PRIMARY KEY,
    inbound_id  UUID NOT NULL,
    sni         TEXT NOT NULL,
    action      TEXT NOT NULL DEFAULT 'proxy',
    target_tag  TEXT NOT NULL DEFAULT '',
    priority    INTEGER NOT NULL DEFAULT 0,
    enabled     BOOLEAN NOT NULL DEFAULT TRUE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_sni_routes_inbound ON sni_routes (inbound_id);

-- v1.2.0: TLS Trick Profiles
CREATE TABLE tls_trick_profiles (
    id                 UUID PRIMARY KEY,
    name               TEXT NOT NULL,
    description        TEXT NOT NULL DEFAULT '',
    fragment_enabled   BOOLEAN NOT NULL DEFAULT FALSE,
    fragment_length    TEXT NOT NULL DEFAULT '',
    fragment_interval  TEXT NOT NULL DEFAULT '',
    fingerprint        TEXT NOT NULL DEFAULT 'chrome',
    mux_enabled        BOOLEAN NOT NULL DEFAULT FALSE,
    mux_protocol       TEXT NOT NULL DEFAULT '',
    enabled            BOOLEAN NOT NULL DEFAULT TRUE,
    created_at         TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- v1.2.0: Fingerprint Policy (singleton)
CREATE TABLE fingerprint_policy (
    id              INTEGER PRIMARY KEY DEFAULT 1,
    enabled         BOOLEAN NOT NULL DEFAULT FALSE,
    default_action  TEXT NOT NULL DEFAULT 'allow',
    log_unknown     BOOLEAN NOT NULL DEFAULT TRUE
);

-- v1.2.0: Fingerprint Rules
CREATE TABLE fingerprint_rules (
    id          UUID PRIMARY KEY,
    name        TEXT NOT NULL,
    fingerprint TEXT NOT NULL DEFAULT '',
    ja3_hash    TEXT NOT NULL DEFAULT '',
    action      TEXT NOT NULL DEFAULT 'allow',
    priority    INTEGER NOT NULL DEFAULT 0,
    enabled     BOOLEAN NOT NULL DEFAULT TRUE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- v1.2.0: Fingerprint Events
CREATE TABLE fingerprint_events (
    id          UUID PRIMARY KEY,
    client_ip   TEXT NOT NULL,
    fingerprint TEXT NOT NULL DEFAULT '',
    user_agent  TEXT NOT NULL DEFAULT '',
    action      TEXT NOT NULL DEFAULT 'allow',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_fp_events_time ON fingerprint_events (created_at DESC);

-- v1.2.0: Federation Config (singleton)
CREATE TABLE federation_config (
    id             INTEGER PRIMARY KEY DEFAULT 1,
    enabled        BOOLEAN NOT NULL DEFAULT FALSE,
    cluster_name   TEXT NOT NULL DEFAULT '',
    sso_enabled    BOOLEAN NOT NULL DEFAULT FALSE,
    sync_interval  INTEGER NOT NULL DEFAULT 60
);

-- v1.2.0: Federation Peers
CREATE TABLE federation_peers (
    id          UUID PRIMARY KEY,
    name        TEXT NOT NULL,
    endpoint    TEXT NOT NULL,
    api_key     TEXT NOT NULL DEFAULT '',
    status      TEXT NOT NULL DEFAULT 'disconnected',
    sync_users  BOOLEAN NOT NULL DEFAULT TRUE,
    sync_nodes  BOOLEAN NOT NULL DEFAULT TRUE,
    last_sync   TIMESTAMPTZ,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- v1.2.0: Federation Sync Events
CREATE TABLE federation_sync_events (
    id          UUID PRIMARY KEY,
    peer_name   TEXT NOT NULL,
    direction   TEXT NOT NULL DEFAULT 'push',
    entity_type TEXT NOT NULL DEFAULT '',
    count       INTEGER NOT NULL DEFAULT 0,
    status      TEXT NOT NULL DEFAULT 'success',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_fed_sync_time ON federation_sync_events (created_at DESC);

-- v1.2.0: Deep Link Config (singleton)
CREATE TABLE deeplink_config (
    id              INTEGER PRIMARY KEY DEFAULT 1,
    base_url        TEXT NOT NULL DEFAULT '',
    app_scheme      TEXT NOT NULL DEFAULT 'vortex',
    include_name    BOOLEAN NOT NULL DEFAULT TRUE,
    qr_logo_url     TEXT NOT NULL DEFAULT ''
);

-- v1.2.0: Quota Notification Config (singleton)
CREATE TABLE quota_notify_config (
    id              INTEGER PRIMARY KEY DEFAULT 1,
    enabled         BOOLEAN NOT NULL DEFAULT FALSE,
    threshold_pct   INTEGER NOT NULL DEFAULT 80,
    notify_telegram BOOLEAN NOT NULL DEFAULT TRUE,
    notify_email    BOOLEAN NOT NULL DEFAULT FALSE,
    message_template TEXT NOT NULL DEFAULT ''
);

-- v1.2.0: Quota Notification Events
CREATE TABLE quota_notify_events (
    id         UUID PRIMARY KEY,
    user_id    UUID NOT NULL,
    username   TEXT NOT NULL DEFAULT '',
    threshold  INTEGER NOT NULL DEFAULT 0,
    usage_pct  INTEGER NOT NULL DEFAULT 0,
    notified   BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_quota_notify_time ON quota_notify_events (created_at DESC);

-- v1.2.0: Subscription Auto-Update Settings (singleton)
CREATE TABLE sub_settings (
    id              INTEGER PRIMARY KEY DEFAULT 1,
    update_interval INTEGER NOT NULL DEFAULT 12,
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT sub_settings_singleton CHECK (id = 1)
);
INSERT INTO sub_settings (id, update_interval) VALUES (1, 12) ON CONFLICT (id) DO NOTHING;

-- Stage 1: WireGuard server inbound per-user peers (keypair + tunnel IP).
CREATE TABLE wireguard_peers (
    inbound_id  UUID NOT NULL REFERENCES inbounds(id) ON DELETE CASCADE,
    user_id     UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    private_key TEXT NOT NULL,
    public_key  TEXT NOT NULL,
    address     TEXT NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (inbound_id, user_id)
);

-- panel-feature-parity: IP-limit enforcement policy (singleton)
CREATE TABLE ip_limit_policy (
    id             INTEGER PRIMARY KEY DEFAULT 1,
    enabled        BOOLEAN NOT NULL DEFAULT FALSE,
    action         TEXT NOT NULL DEFAULT 'warn',
    alert_cooldown INTEGER NOT NULL DEFAULT 900,
    restore_after  INTEGER NOT NULL DEFAULT 900
);

-- panel-feature-parity: IP-limit enforcement events
CREATE TABLE ip_limit_events (
    id         UUID PRIMARY KEY,
    user_id    UUID NOT NULL,
    username   TEXT NOT NULL DEFAULT '',
    online_ips INTEGER NOT NULL DEFAULT 0,
    limit_val  INTEGER NOT NULL DEFAULT 0,
    action     TEXT NOT NULL DEFAULT 'warn',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_ip_limit_events_time ON ip_limit_events (created_at DESC);

-- panel-feature-parity: Clean-IP scanner results cache
CREATE TABLE clean_ip_scans (
    id         UUID PRIMARY KEY,
    ip         TEXT NOT NULL,
    latency_ms INTEGER NOT NULL DEFAULT 0,
    loss_pct   INTEGER NOT NULL DEFAULT 0,
    score      INTEGER NOT NULL DEFAULT 0,
    reachable  BOOLEAN NOT NULL DEFAULT FALSE,
    scanned_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_clean_ip_score ON clean_ip_scans (score DESC);

-- reseller per-admin payment configuration
CREATE TABLE reseller_payment_config (
    admin_id            UUID PRIMARY KEY REFERENCES admins(id) ON DELETE CASCADE,
    card_number         TEXT NOT NULL DEFAULT '',
    card_holder         TEXT NOT NULL DEFAULT '',
    card_bank           TEXT NOT NULL DEFAULT '',
    crypto_addresses    JSONB NOT NULL DEFAULT '{}',
    zarinpal_merchant_id TEXT NOT NULL DEFAULT '',
    manual_instructions TEXT NOT NULL DEFAULT '',
    enabled_methods     JSONB NOT NULL DEFAULT '[]'
);

-- Orders table (purchase orders for plans)
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
    paid_at     TIMESTAMPTZ,
    proof_image TEXT NOT NULL DEFAULT ''
);

CREATE INDEX idx_orders_user ON orders(user_id);
CREATE INDEX idx_orders_status ON orders(status);
