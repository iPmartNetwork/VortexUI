/** Default permission bundle for a reseller (own users + read-only nodes/inbounds). */
export const RESELLER_PERMISSIONS = [
  "user:read",
  "user:write",
  "node:read",
  "inbound:read",
  "system:read",
] as const;

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
  "/orders": ["user:read"],
  "/family-groups": ["user:write"],
  "/smart-quota": ["admin:manage"],
  "/quota-notifications": ["admin:manage"],
  "/reseller-quota-alerts": ["admin:manage"],
  "/referrals": ["admin:manage"],
  "/tickets": ["user:write"],
  "/nodes": ["node:read"],
  "/outbounds": ["inbound:read"],
  "/routing": ["inbound:read"],
  "/routing-packs": ["inbound:read"],
  "/balancers": ["inbound:read"],
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
};

export function canAccessRoute(path: string, sudo: boolean, permissions: Set<string>): boolean {
  if (sudo) return true;
  const need = ROUTE_PERMISSIONS[path];
  if (!need || need.length === 0) return true;
  return need.some((p) => permissions.has(p));
}

export function hasPermission(sudo: boolean, permissions: Set<string>, perm: string): boolean {
  return sudo || permissions.has(perm);
}
