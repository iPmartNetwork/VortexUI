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

export interface Node {
  id: string;
  name: string;
  address: string;
  core: "xray" | "singbox";
  status: "connected" | "disconnected" | "error" | "disabled";
  usage_ratio: number;
  last_seen: string | null;
  health: NodeHealth;
  created_at: string;
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
