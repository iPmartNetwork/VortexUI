import { useState, useEffect, useRef, useMemo } from "react";
import { useNavigate } from "react-router-dom";
import {
  Search, Command, Users, Server,
  User, Activity,
} from "lucide-react";
import { cn } from "@/lib/utils";
import { useI18n } from "@/i18n/i18n";
import type { TKey } from "@/i18n/dict";
import { useQuery } from "@tanstack/react-query";
import { api } from "@/api/client";

interface PaletteItem {
  id: string;
  label: string;
  sublabel?: string;
  action: () => void;
  keywords?: string[];
  icon?: React.ReactNode;
  group?: string;
}

const ROUTES: { path: string; labelKey: TKey; keywords: string[]; icon?: React.ReactNode }[] = [
  { path: "/overview", labelKey: "nav.overview", keywords: ["home", "overview", "stats", "dashboard"] },
  { path: "/dashboard-pro", labelKey: "nav.dashboardPro", keywords: ["advanced", "monitor", "geo", "pro"] },
  { path: "/users", labelKey: "nav.users", keywords: ["accounts", "subscribers", "people"] },
  { path: "/nodes", labelKey: "nav.nodes", keywords: ["servers", "agents", "machines"] },
  // { path: "/inbounds", labelKey: "nav.inbounds", keywords: ["ports", "listen", "entries"] },
  { path: "/routing?tab=outbounds", labelKey: "nav.outbounds", keywords: ["egress", "proxy", "outbound"] },
  { path: "/routing", labelKey: "nav.smartRoutingBalancers", keywords: ["routes", "policy", "packs", "load", "balance", "balancers", "outbounds"] },
  { path: "/wallet-billing", labelKey: "nav.resellerPlatform", keywords: ["wallet", "billing", "plans", "orders", "deposit", "reseller"] },
  { path: "/orders", labelKey: "nav.orders", keywords: ["payments", "purchases"] },
  { path: "/analytics", labelKey: "nav.analytics", keywords: ["stats", "traffic", "geo", "charts"] },
  { path: "/tickets", labelKey: "nav.supportDesk", keywords: ["help", "support", "tickets", "issue"] },
  { path: "/monitor", labelKey: "nav.monitor", keywords: ["realtime", "watch", "live"] },
  { path: "/evasion", labelKey: "nav.securityAntiDpi", keywords: ["reality", "clean-ip", "tls", "decoy", "probing", "anti-dpi", "fragment", "gfw"] },
  { path: "/ip-limit", labelKey: "nav.ipLimit", keywords: ["share", "device", "ip", "limit", "shareguard"] },
  { path: "/security", labelKey: "nav.security", keywords: ["threat", "hardening", "waf", "score"] },
  { path: "/smart-quota", labelKey: "nav.smartQuota", keywords: ["fair use", "throttle", "speed", "traffic"] },
  { path: "/relay-chains", labelKey: "nav.relayChains", keywords: ["cdn", "relay", "chain"] },
  { path: "/migration", labelKey: "nav.migration", keywords: ["failover", "health"] },
  { path: "/family-groups", labelKey: "nav.familyGroups", keywords: ["shared", "pool", "group"] },
  { path: "/referrals", labelKey: "referral.title", keywords: ["invite", "reward", "code"] },
  { path: "/doh", labelKey: "nav.doh", keywords: ["dns", "privacy", "doh"] },
  { path: "/sni-manager", labelKey: "nav.sniManager", keywords: ["domain", "cert", "tls"] },
  { path: "/fingerprint", labelKey: "nav.fingerprint", keywords: ["ja3", "client", "tls"] },
  { path: "/federation", labelKey: "nav.federation", keywords: ["multi", "sync", "peer"] },
  { path: "/deep-links", labelKey: "deepLink.title", keywords: ["qr", "import", "app"] },
  { path: "/quota-notifications", labelKey: "nav.quotaNotify", keywords: ["alert", "notify"] },
  { path: "/settings?tab=admins", labelKey: "nav.admins", keywords: ["roles", "permissions", "admins"] },
  { path: "/logs", labelKey: "nav.logs", keywords: ["debug", "error", "system"] },
  { path: "/settings", labelKey: "nav.settings", keywords: ["config", "branding", "admins"] },
];

export function CommandPalette() {
  const { t } = useI18n();
  const [open, setOpen] = useState(false);
  const [query, setQuery] = useState("");
  const [selected, setSelected] = useState(0);
  const inputRef = useRef<HTMLInputElement>(null);
  const navigate = useNavigate();

  useEffect(() => {
    function onKeyDown(e: KeyboardEvent) {
      if ((e.metaKey || e.ctrlKey) && e.key === "k") {
        e.preventDefault();
        setOpen(prev => !prev);
      }
      if (e.key === "Escape") setOpen(false);
    }
    window.addEventListener("keydown", onKeyDown);
    return () => window.removeEventListener("keydown", onKeyDown);
  }, []);

  useEffect(() => {
    if (open) {
      setQuery("");
      setSelected(0);
      setTimeout(() => inputRef.current?.focus(), 50);
    }
  }, [open]);

  // ── Fetch users & nodes for dynamic search ──
  const { data: usersData } = useQuery({
    queryKey: ["cmd-palette-users"],
    queryFn: () => api<{ users: { id: string; username: string; status: string }[] }>("/api/users", { query: { limit: 50 } }),
    staleTime: 60_000,
  });

  const { data: nodesData } = useQuery({
    queryKey: ["cmd-palette-nodes"],
    queryFn: () => api<{ nodes: { id: string; name: string; core: string; location: string; status: string }[] }>("/api/nodes"),
    staleTime: 60_000,
  });

  // ── Build items from routes + dynamic data ──
  const routeItems: PaletteItem[] = useMemo(() =>
    ROUTES.map(r => ({
      id: `route:${r.path}`,
      label: t(r.labelKey),
      sublabel: r.path,
      action: () => { navigate(r.path); setOpen(false); },
      keywords: r.keywords,
      group: "routes",
    })), [t, navigate]);

  const userItems: PaletteItem[] = useMemo(() =>
    (usersData?.users ?? []).slice(0, 10).map(u => ({
      id: `user:${u.id}`,
      label: u.username,
      sublabel: u.status,
      action: () => { navigate(`/users/${u.id}`); setOpen(false); },
      keywords: [u.username.toLowerCase(), "user", "account", u.id],
      icon: <User size={14} />,
      group: "users",
    })), [usersData, navigate]);

  const nodeItems: PaletteItem[] = useMemo(() =>
    (nodesData?.nodes ?? []).slice(0, 10).map(n => ({
      id: `node:${n.id}`,
      label: n.name,
      sublabel: `${n.core} · ${n.location}`,
      action: () => { navigate(`/nodes/${n.id}`); setOpen(false); },
      keywords: [n.name.toLowerCase(), "node", "server", n.core, n.location.toLowerCase()],
      icon: <Server size={14} />,
      group: "nodes",
    })), [nodesData, navigate]);

  // Quick actions that can be performed directly from palette
  const actionItems: PaletteItem[] = [
    {
      id: "action:create-user",
      label: "Create User",
      sublabel: "Add a new subscription",
      action: () => { /* could dispatch modal open */ setOpen(false); },
      keywords: ["new user", "add", "create account"],
      icon: <Users size={14} />,
      group: "actions",
    },
    {
      id: "action:filter-active",
      label: "View Active Users",
      sublabel: "/users?status=active",
      action: () => { navigate("/users?status=active"); setOpen(false); },
      keywords: ["active", "online", "live"],
      icon: <Activity size={14} />,
      group: "actions",
    },
  ];

  const allItems = useMemo(() =>
    [...routeItems, ...actionItems, ...userItems, ...nodeItems],
    [routeItems, actionItems, userItems, nodeItems],
  );

  // ── Filter items across all groups ──
  const filtered = useMemo(() => {
    if (!query.trim()) return allItems.slice(0, 15);
    const q = query.toLowerCase();
    return allItems.filter(item =>
      item.label.toLowerCase().includes(q)
      || item.sublabel?.toLowerCase().includes(q)
      || item.keywords?.some(k => k.includes(q))
      || item.group?.includes(q)
    ).slice(0, 25);
  }, [query, allItems]);

  // Group filtered items by category
  const grouped = useMemo(() => {
    const groups: Record<string, PaletteItem[]> = {};
    for (const item of filtered) {
      const g = item.group || "other";
      if (!groups[g]) groups[g] = [];
      groups[g].push(item);
    }
    return groups;
  }, [filtered]);

  // ── Flatten for keyboard navigation ──
  const flatFiltered = useMemo(() => Object.values(grouped).flat(), [grouped]);

  // Reset selected index when filtered changes
  useEffect(() => { setSelected(0); }, [filtered.length]);

  function handleKeyDown(e: React.KeyboardEvent) {
    if (e.key === "ArrowDown") { e.preventDefault(); setSelected(s => Math.min(s + 1, flatFiltered.length - 1)); }
    if (e.key === "ArrowUp") { e.preventDefault(); setSelected(s => Math.max(s - 1, 0)); }
    if (e.key === "Enter" && flatFiltered[selected]) { flatFiltered[selected].action(); }
  }

  if (!open) return null;

  // Map groups to labels
  const groupLabels: Record<string, string> = {
    routes: "Navigation",
    actions: "Quick Actions",
    users: "Users",
    nodes: "Nodes",
  };

  return (
    <div className="fixed inset-0 z-[100] flex items-start justify-center pt-[15vh] animate-fade-in" onClick={() => setOpen(false)}>
      <div className="fixed inset-0 bg-black/50 backdrop-blur-sm" />
      <div className="relative w-full max-w-xl rounded-2xl border border-border/60 bg-bg-elevated/95 shadow-2xl backdrop-blur-xl animate-scale-in overflow-hidden" onClick={e => e.stopPropagation()}>
        <div className="flex items-center gap-3 border-b border-border/40 px-4 py-3">
          <Search size={18} className="text-fg-subtle" />
          <input
            ref={inputRef}
            value={query}
            onChange={e => { setQuery(e.target.value); setSelected(0); }}
            onKeyDown={handleKeyDown}
            placeholder={t("cmdPalette.searchPlaceholder")}
            className="flex-1 bg-transparent text-sm text-fg outline-none placeholder:text-fg-subtle"
          />
          <kbd className="hidden sm:inline-flex items-center gap-0.5 rounded-md border border-border/60 bg-surface-2/50 px-1.5 py-0.5 text-[10px] text-fg-subtle">
            ESC
          </kbd>
        </div>

        <div className="max-h-[400px] overflow-y-auto p-2 space-y-1">
          {flatFiltered.length === 0 && (
            <div className="py-12 text-center">
              <Search size={28} className="mx-auto text-fg-subtle/30 mb-2" />
              <p className="text-sm text-fg-muted">{t("cmdPalette.noResults")}</p>
              <p className="text-xs text-fg-subtle mt-1">Try a different search term</p>
            </div>
          )}
          {Object.entries(grouped).map(([group, items]) => {
            const globalStartIndex = Object.entries(grouped)
              .slice(0, Object.keys(grouped).indexOf(group))
              .reduce((sum, [, g]) => sum + g.length, 0);

            return (
              <div key={group}>
                <p className="px-3 py-1.5 text-[9px] font-bold uppercase tracking-wider text-fg-subtle/50">
                  {groupLabels[group] || group}
                </p>
                {items.map((item, localIdx) => {
                  const globalIdx = globalStartIndex + localIdx;
                  return (
                    <button
                      key={item.id}
                      onClick={item.action}
                      className={cn(
                        "flex w-full items-center gap-3 rounded-xl px-3 py-2.5 text-left transition-all duration-100",
                        globalIdx === selected
                          ? "bg-primary/12 text-primary border border-primary/20 shadow-sm"
                          : "text-fg-muted hover:bg-surface-2/60 hover:text-fg border border-transparent",
                      )}
                    >
                      {item.icon && (
                        <span className="flex-shrink-0 w-5 flex justify-center text-fg-subtle">
                          {item.icon}
                        </span>
                      )}
                      <div className="flex-1 min-w-0">
                        <div className="text-sm font-medium truncate">{item.label}</div>
                        <div className="text-[10px] text-fg-subtle/70 truncate">{item.sublabel}</div>
                      </div>
                      {globalIdx === selected && (
                        <span className="flex-shrink-0 text-[9px] text-fg-subtle bg-surface-2/60 rounded px-1.5 py-0.5 border border-border/40">
                          ↵
                        </span>
                      )}
                    </button>
                  );
                })}
              </div>
            );
          })}
        </div>

        <div className="flex items-center justify-between border-t border-border/40 px-4 py-2 text-[10px] text-fg-subtle">
          <div className="flex items-center gap-3">
            <span className="flex items-center gap-1"><kbd className="rounded border border-border/40 bg-surface-2/40 px-1">↑↓</kbd> Navigate</span>
            <span className="flex items-center gap-1"><kbd className="rounded border border-border/40 bg-surface-2/40 px-1">↵</kbd> Select</span>
          </div>
          <span className="flex items-center gap-1"><Command size={10} />K toggle</span>
        </div>
      </div>
    </div>
  );
}

export function CommandPaletteTrigger() {
  const { t } = useI18n();
  return (
    <button
      onClick={() => window.dispatchEvent(new KeyboardEvent("keydown", { key: "k", metaKey: true }))}
      className="flex items-center gap-2 rounded-xl border border-border/50 bg-surface-2/30 px-3 py-1.5 text-xs text-fg-muted transition hover:bg-surface-2/60 hover:text-fg"
    >
      <Search size={13} />
      <span className="hidden md:inline">{t("cmdPalette.searchTrigger")}</span>
      <kbd className="hidden md:inline-flex items-center rounded border border-border/50 bg-surface/50 px-1 text-[10px]">⌘K</kbd>
    </button>
  );
}
