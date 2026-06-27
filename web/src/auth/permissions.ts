/** Default permission bundle for a reseller (own users + read-only nodes). */
export const RESELLER_PERMISSIONS = [
  "user:read",
  "user:write",
  "node:read",
  "inbound:read",
  "system:read",
] as const;

/** Settings section keys mirrored from the panel API. */
export const RESELLER_SETTING_KEYS = [
  "appearance",
  "password",
  "totp",
  "api_tokens",
  "backup",
  "config_template",
  "sub_update",
  "ip_guard",
  "branding",
  "auto_backup",
  "update",
] as const;

export type ResellerSettingKey = (typeof RESELLER_SETTING_KEYS)[number];

export const DEFAULT_RESELLER_SETTINGS: Record<ResellerSettingKey, boolean> = {
  appearance: true,
  password: true,
  totp: true,
  api_tokens: false,
  backup: false,
  config_template: false,
  sub_update: false,
  ip_guard: false,
  branding: false,
  auto_backup: false,
  update: false,
};

export function mergeResellerSettings(stored?: Record<string, boolean> | null): Record<ResellerSettingKey, boolean> {
  return { ...DEFAULT_RESELLER_SETTINGS, ...(stored ?? {}) };
}

export function resellerSettingEnabled(stored: Record<string, boolean> | null | undefined, key: ResellerSettingKey): boolean {
  const merged = mergeResellerSettings(stored);
  return merged[key];
}

export function anyResellerSettingEnabled(stored?: Record<string, boolean> | null): boolean {
  return Object.values(mergeResellerSettings(stored)).some(Boolean);
}

/** Minimum permission to show a sidebar route (any listed permission is enough). */
export const ROUTE_PERMISSIONS: Record<string, readonly string[]> = {
  "/overview": ["system:read"],
  "/reseller-dashboard": ["system:read"],
  "/reseller-account": ["system:read"],
  "/my-quota": ["system:read"],
  "/monitor": ["system:read"],
  "/analytics": ["system:read"],
  "/users": ["user:read"],
  "/plans": ["user:read"],
  "/wallet-billing": ["admin:manage"],
  "/orders": ["user:read"],
  "/family-groups": ["user:write"],
  "/smart-quota": ["admin:manage"],
  "/quota-notifications": ["admin:manage"],
  "/reseller-quota-alerts": ["admin:manage"],
  "/referrals": ["admin:manage"],
  "/tickets": ["user:write"],
  "/nodes": ["node:read"],
  "/outbounds": ["inbound:write"],
  "/routing": ["inbound:write"],
  "/routing-packs": ["inbound:write"],
  "/balancers": ["inbound:write"],
  "/relay-chains": ["node:write"],
  "/migration": ["admin:manage"],
  "/federation": ["admin:manage"],
  "/evasion": ["inbound:write"],
  "/tls-tricks": ["inbound:write"],
  "/sni-manager": ["inbound:write"],
  "/reality-scanner": ["inbound:read"],
  "/clean-ip": ["inbound:read"],
  "/decoy-website": ["node:write"],
  "/probing-protection": ["admin:manage"],
  "/ip-limit": ["admin:manage"],
  "/fingerprint": ["admin:manage"],
  "/doh": ["admin:manage"],
  "/admins": ["admin:manage"],
  "/deep-links": ["admin:manage"],
  "/audit": ["system:read"],
  "/logs": ["system:read"],
  "/settings": ["system:read"],
};

export function canAccessRoute(
  path: string,
  sudo: boolean,
  permissions: Set<string>,
  resellerSettings?: Record<string, boolean> | null,
): boolean {
  if (sudo) return true;
  if (path === "/settings" && !anyResellerSettingEnabled(resellerSettings)) {
    return false;
  }
  const need = ROUTE_PERMISSIONS[path];
  if (!need || need.length === 0) return true;
  return need.some((p) => permissions.has(p));
}

export function hasPermission(sudo: boolean, permissions: Set<string>, perm: string): boolean {
  return sudo || permissions.has(perm);
}
