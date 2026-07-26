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
  FileText,
  Layers,
  Bell,
  BarChart3,
  Cpu,
  Shield,
  BookOpen,
  Smartphone,
  Gauge,
  Cable,
  RefreshCw,
} from "lucide-react";
import type { NavSection } from "./nav-sections";

/** Compact sidebar groups matching the VortexUI / Arena command-tower mock. */
export function buildCompactNavSections(sudo: boolean): NavSection[] {
  return [
    {
      label: "nav.section.mainCommand",
      id: "main",
      items: [
        { to: "/overview", key: "nav.overview", icon: LayoutDashboard },
        { to: "/dashboard-pro", key: "nav.dashboardPro", icon: BarChart3 },
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
        { to: "/config-management", key: "nav.configManagement", icon: Cpu },
        { to: "/wireguard", key: "nav.wireguard", icon: Lock },
        { to: "/switch-analytics", key: "nav.switchAnalytics", icon: ArrowRightLeft },
        { to: "/routing", key: "nav.smartRoutingBalancers", icon: RouteIcon },
        { to: "/tunnels", key: "nav.tunnels", icon: Cable },
      ],
    },
    {
      label: "nav.section.security",
      id: "security",
      items: [
        { to: "/evasion", key: "nav.securityAntiDpi", icon: ShieldCheck, hotDot: true },
        { to: "/advanced-security", key: "nav.advancedSecurity", icon: Shield },
        { to: "/ip-limit", key: "nav.ipLimit", icon: ShieldAlert },
        { to: "/fingerprint", key: "nav.fingerprint", icon: Fingerprint },
        { to: "/doh", key: "nav.doh", icon: Network },
        { to: "/sni-manager", key: "nav.sniManager", icon: Lock },
        { to: "/connection-quality", key: "nav.connectionQuality", icon: Activity },
        { to: "/geo-exits", key: "nav.geoExits", icon: Globe },
        { to: "/ip-rotation", key: "nav.ipRotation", icon: RefreshCw },
      ],
    },
    {
      label: "nav.section.templates",
      id: "templates",
      items: [
        { to: "/templates", key: "nav.userTemplates", icon: FileText },
        { to: "/bulk-operations", key: "nav.bulkOperations", icon: Layers },
        { to: "/client-templates", key: "nav.clientTemplates", icon: Smartphone },
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
        { to: "/notifications", key: "nav.notifications", icon: Bell },
        { to: "/performance", key: "nav.performance", icon: Gauge },
        { to: "/api-docs", key: "nav.apiDocs", icon: BookOpen },
        { to: "/settings", key: "nav.systemSettings", icon: SettingsIcon },
      ],
    },
  ];
}
