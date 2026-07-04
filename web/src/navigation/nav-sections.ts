import {
  LayoutDashboard,
  Users as UsersIcon,
  Server,
  Network,
  Route as RouteIcon,
  ShieldCheck,
  ScrollText,
  History,
  Settings as SettingsIcon,
  Gauge,
  Link2,
  BarChart3,
  LifeBuoy,
  ArrowRightLeft,
  Users2,
  Gift,
  Wifi,
  Lock,
  Fingerprint as FingerprintIcon,
  Layers,
  QrCode,
  Bell,
  Ban,
  Wallet,
  ClipboardList,
} from "lucide-react";
import type { TKey } from "@/i18n/dict";

export interface NavItem {
  to: string;
  key: TKey;
  icon: React.ElementType;
  badgeKey?: "active_users" | "open_tickets" | "pending_orders";
  /** Green pulse dot (no count) — e.g. Security suite */
  hotDot?: boolean;
}

export interface NavSection {
  label: TKey;
  id: string;
  items: NavItem[];
}

export function buildNavSections(sudo: boolean): NavSection[] {
  const resellerSection: NavSection | null = sudo
    ? null
    : {
        label: "nav.section.reseller",
        id: "reseller",
        items: [
          { to: "/reseller-dashboard", key: "nav.resellerDashboard", icon: Gauge },
          { to: "/reseller-account", key: "nav.resellerAccount", icon: Wallet },
          { to: "/pending-orders", key: "nav.pendingOrders", icon: ClipboardList, badgeKey: "pending_orders" },
        ],
      };

  const sections: NavSection[] = [
    {
      label: "nav.section.dashboard",
      id: "dashboard",
      items: [
        { to: "/overview", key: "nav.overview", icon: LayoutDashboard },
        { to: "/monitor", key: "nav.monitor", icon: Server },
        { to: "/analytics", key: "nav.analytics", icon: BarChart3 },
      ],
    },
    {
      label: "nav.section.users",
      id: "users",
      items: [
        { to: "/users", key: "nav.users", icon: UsersIcon },
        { to: "/family-groups", key: "nav.familyGroups", icon: Users2 },
        { to: "/wallet-billing", key: "nav.resellerPlatform", icon: Wallet, badgeKey: "pending_orders" },
        { to: "/orders", key: "nav.orders", icon: History },
        { to: "/smart-quota", key: "nav.smartQuota", icon: Gauge },
        { to: "/quota-notifications", key: "nav.quotaNotify", icon: Bell },
        { to: "/reseller-quota-alerts", key: "nav.resellerQuotaAlerts", icon: Bell },
        { to: "/referrals", key: "nav.referrals", icon: Gift },
        { to: "/tickets", key: "nav.tickets", icon: LifeBuoy },
      ],
    },
    {
      label: "nav.section.network",
      id: "network",
      items: [
        { to: "/nodes", key: "nav.nodes", icon: Server },
        { to: "/outbounds", key: "nav.outbounds", icon: Network },
        { to: "/routing", key: "nav.smartRoutingBalancers", icon: RouteIcon },
        { to: "/relay-chains", key: "nav.relayChains", icon: Link2 },
        { to: "/migration", key: "nav.migration", icon: ArrowRightLeft },
        { to: "/federation", key: "nav.federation", icon: Layers },
      ],
    },
    {
      label: "nav.section.security",
      id: "security",
      items: [
        { to: "/evasion", key: "nav.securityAntiDpi", icon: ShieldCheck },
        { to: "/sni-manager", key: "nav.sniManager", icon: Lock },
        { to: "/ip-limit", key: "nav.ipLimit", icon: Ban },
        { to: "/fingerprint", key: "nav.fingerprint", icon: FingerprintIcon },
        { to: "/doh", key: "nav.doh", icon: Wifi },
      ],
    },
    {
      label: "nav.section.system",
      id: "system",
      items: [
        { to: "/admins", key: "nav.admins", icon: ShieldCheck },
        { to: "/deep-links", key: "nav.deepLinks", icon: QrCode },
        { to: "/audit", key: "nav.audit", icon: History },
        { to: "/logs", key: "nav.logs", icon: ScrollText },
        { to: "/settings", key: "nav.settings", icon: SettingsIcon },
      ],
    },
  ];

  return resellerSection ? [resellerSection, ...sections] : sections;
}
