import {
  LayoutDashboard,
  Users as UsersIcon,
  Server,
  Network,
  Route as RouteIcon,
  ShieldCheck,
  Wallet,
  LifeBuoy,
  Settings as SettingsIcon,
  Scale,
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
        { to: "/nodes?tab=inbounds", key: "nav.inboundsSubhosts", icon: Network },
        { to: "/routing", key: "nav.smartRoutingBalancers", icon: RouteIcon },
        { to: "/balancers", key: "nav.balancersShort", icon: Scale },
        { to: "/evasion", key: "nav.securityAntiDpi", icon: ShieldCheck },
      ],
    },
    {
      label: "nav.section.commerce",
      id: "commerce",
      items: [
        {
          to: sudo ? "/wallet-billing" : "/reseller-account",
          key: "nav.resellerWallet",
          icon: Wallet,
          badgeKey: "pending_orders",
        },
        { to: "/tickets", key: "nav.supportDesk", icon: LifeBuoy, badgeKey: "open_tickets" },
      ],
    },
    {
      label: "nav.section.systemConfig",
      id: "system",
      items: [{ to: "/settings", key: "nav.systemSettings", icon: SettingsIcon }],
    },
  ];
}
