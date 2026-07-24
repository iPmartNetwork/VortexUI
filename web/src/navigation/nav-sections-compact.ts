import {
  LayoutDashboard,
  Users as UsersIcon,
  Server,
  Globe,
  Route as RouteIcon,
  ShieldCheck,
  ShieldAlert,
  Fingerprint,
  Network,
  Lock,
  Package,
  LifeBuoy,
  Settings as SettingsIcon,
  Activity,
  ArrowRightLeft,
} from "lucide-react";
import type { NavSection } from "./nav-sections";

/** Compact sidebar groups matching the Veltrix / Arena command-tower mock. */
export function buildCompactNavSections(sudo: boolean): NavSection[] {
  return [
    {
      label: "nav.section.mainCommand",
      id: "main",
      items: [
        { to: "/overview", key: "nav.overview", icon: LayoutDashboard },
        { to: "/users", key: "nav.usersSubscriptions", icon: UsersIcon, badgeKey: "active_users" },
        { to: "/nodes", key: "nav.nodesFleet", icon: Server },
      ],
    },
    {
      label: "nav.section.networkProxy",
      id: "network",
      items: [
        { to: "/inbounds", key: "nav.inboundsSubhosts", icon: Globe },
        { to: "/inbounds?tab=groups", key: "nav.protocolGroups", icon: Network },
        { to: "/switch-analytics", key: "nav.switchAnalytics", icon: ArrowRightLeft },
        { to: "/routing", key: "nav.smartRoutingBalancers", icon: RouteIcon },
      ],
    },
    {
      label: "nav.section.security",
      id: "security",
      items: [
        { to: "/evasion", key: "nav.securityAntiDpi", icon: ShieldCheck, hotDot: true },
        { to: "/ip-limit", key: "nav.ipLimit", icon: ShieldAlert },
        { to: "/fingerprint", key: "nav.fingerprint", icon: Fingerprint },
        { to: "/doh", key: "nav.doh", icon: Network },
        { to: "/sni-manager", key: "nav.sniManager", icon: Lock },
        { to: "/connection-quality", key: "nav.connectionQuality", icon: Activity },
      ],
    },
    {
      label: "nav.section.commerce",
      id: "commerce",
      items: [
        {
          to: sudo ? "/wallet-billing" : "/reseller-account",
          key: "nav.resellerPlatform",
          icon: Package,
          badgeKey: "pending_orders",
        },
        { to: "/tickets", key: "nav.supportDesk", icon: LifeBuoy, badgeKey: "open_tickets" },
      ],
    },
    {
      label: "nav.section.systemConfig",
      id: "system",
      items: [
        { to: "/security", key: "nav.security", icon: ShieldAlert },
        { to: "/settings", key: "nav.systemSettings", icon: SettingsIcon },
      ],
    },
  ];
}
