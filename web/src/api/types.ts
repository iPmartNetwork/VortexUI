// Domain types mirroring the panel API (docs/openapi.yaml). For a larger surface
// these can be generated: `npx openapi-typescript docs/openapi.yaml`.

export type UserStatus = "active" | "limited" | "expired" | "disabled" | "on_hold";

export interface Credentials {
  vmess_uuid: string;
  vless_uuid: string;
  trojan_password: string;
  ss_password: string;
  ss_method: string;
}

export interface User {
  id: string;
  username: string;
  status: UserStatus;
  note: string;
  data_limit: number;
  used_traffic: number;
  expire_at: string | null;
  reset_strategy: "no_reset" | "daily" | "weekly" | "monthly";
  device_limit: number;
  proxies: Credentials;
  sub_token: string;
  created_at: string;
  updated_at: string;
}

export interface NodeHealth {
  cpu_percent: number;
  mem_percent: number;
  disk_percent: number;
  core_running: boolean;
  connections: number;
}

export type NodeDiagCode = "ok" | "unreachable" | "mtls_fail" | "core_down" | "unknown";

export interface NodeDiagnostics {
  code: NodeDiagCode;
  message?: string;
  network_reachable?: boolean;
  ca_match?: boolean;
  checked_at?: string | null;
}

export type NodeEnrollmentPhase = "pending" | "connected" | "synced";

export interface EnrollmentBundle {
  bundle: string;
  ca_fingerprint: string;
  cert_dir: string;
}

export interface Node {
  id: string;
  name: string;
  address: string;
  core: "xray" | "singbox";
  status: "connected" | "disconnected" | "error" | "disabled";
  usage_ratio: number;
  endpoint?: string;
  speed_limit?: number;
  geo_block?: string[];
  last_seen: string | null;
  health: NodeHealth;
  diagnostics?: NodeDiagnostics;
  enrollment_phase?: NodeEnrollmentPhase;
  core_version?: string;
  agent_version?: string;
  created_at: string;
}

export interface Outbound {
  id: string;
  node_id: string;
  tag: string;
  protocol: string;
  address: string;
  port: number;
  uuid: string;
  password: string;
  username: string;
  method: string;
  flow: string;
  network: string;
  security: string;
  sni: string;
  raw?: Record<string, unknown>;
  enabled: boolean;
}

export interface RoutingRule {
  id: string;
  node_id: string;
  priority: number;
  name: string;
  inbound_tags: string[];
  domains: string[];
  ip: string[];
  port: string;
  network: string;
  outbound_tag: string;
  balancer_tag: string;
  enabled: boolean;
}

export interface Balancer {
  id: string;
  node_id: string;
  tag: string;
  selectors: string[];
  strategy: string;
  observe: boolean;
  probe_url: string;
  probe_interval: string;
  enabled: boolean;
}

export interface Overview {
  users: { total: number; by_status: Record<string, number>; total_used: number };
  nodes: {
    total: number;
    online: number;
    items: {
      id: string;
      name: string;
      core: string;
      online: boolean;
      health: NodeHealth;
    }[];
  };
  widgets: DashboardWidgets;
}

export interface DashboardWidgets {
  nav_badges: {
    active_users: number;
    open_tickets: number;
    pending_orders: number;
  };
  trends: {
    users_pct: number;
    bandwidth_pct: number;
    sessions_pct: number;
  };
  probing: {
    enabled: boolean;
    blocked_scanners: number;
    events_24h: number;
  };
  routing: {
    active_rules: number;
    routing_packs: number;
    balancers: number;
    inbounds: number;
  };
  protocols: { label: string; count: number; percent: number }[];
  telemetry?: {
    node_name: string;
    core: string;
    connections: number;
    cpu_percent: number;
    online: boolean;
  };
}

export interface AuditEntry {
  id: string;
  time: string;
  admin_id?: string;
  username: string;
  method: string;
  path: string;
  status: number;
  ip: string;
}

export interface LogEntry {
  time: string;
  level: number;
  message: string;
  attrs?: Record<string, unknown>;
}

export interface UserSub {
  token: string;
  subscription_url: string;
  formats: { auto: string; clash: string; singbox: string; base64: string };
  links: string[];
}

export interface ListUsersResponse {
  users: User[];
  total: number;
}

export interface Admin {
  id: string;
  username: string;
  sudo: boolean;
  role_id: string | null;
  totp_enabled: boolean;
  user_quota: number;
  traffic_quota: number;
  traffic_quota_mode?: string;
  policy_max_data_limit?: number;
  policy_max_expire_days?: number;
  policy_allow_bulk_delete?: boolean;
  policy_allow_bulk_create?: boolean;
  suspended?: boolean;
  suspended_at?: string | null;
  suspend_reason?: string;
  auto_suspend_enabled?: boolean;
  ip_violation_suspend_threshold?: number;
  suspend_grace_minutes?: number;
  allow_sub_resellers?: boolean;
  allow_user_backup?: boolean;
  reseller_settings?: Record<string, boolean>;
  last_login: string | null;
  created_at: string;
}

export interface Role {
  id: string;
  name: string;
  permissions: string[];
}

export const ALL_PERMISSIONS = [
  "user:read",
  "user:write",
  "node:read",
  "node:write",
  "inbound:read",
  "inbound:write",
  "admin:manage",
  "system:read",
] as const;

export interface CreateUserInput {
  username: string;
  note?: string;
  data_limit?: number;
  expire_at?: string | null;
  device_limit?: number;
  reset_strategy?: string;
  inbound_ids?: string[];
  on_hold?: boolean;
}
