import { useState, useEffect, useRef, useMemo } from "react";
import { useNavigate } from "react-router-dom";
import { Search, Command } from "lucide-react";
import { cn } from "@/lib/utils";
import { useI18n } from "@/i18n/i18n";
import type { TKey } from "@/i18n/dict";

interface PaletteItem {
  id: string;
  label: string;
  sublabel?: string;
  action: () => void;
  keywords?: string[];
}

const ROUTES: { path: string; labelKey: TKey; keywords: string[] }[] = [
  { path: "/overview", labelKey: "nav.overview", keywords: ["home", "overview", "stats"] },
  { path: "/users", labelKey: "nav.users", keywords: ["accounts", "subscribers"] },
  { path: "/nodes", labelKey: "nav.nodes", keywords: ["servers", "agents"] },
  { path: "/routing?tab=outbounds", labelKey: "nav.outbounds", keywords: ["egress", "proxy", "outbound"] },
  { path: "/routing", labelKey: "nav.smartRoutingBalancers", keywords: ["routes", "policy", "packs", "load", "balance", "balancers", "outbounds"] },
  { path: "/wallet-billing", labelKey: "nav.resellerPlatform", keywords: ["wallet", "billing", "plans", "orders", "deposit", "reseller"] },
  { path: "/orders", labelKey: "nav.orders", keywords: ["payments", "purchases"] },
  { path: "/analytics", labelKey: "nav.analytics", keywords: ["stats", "traffic", "geo"] },
  { path: "/tickets", labelKey: "nav.supportDesk", keywords: ["help", "support", "tickets"] },
  { path: "/monitor", labelKey: "nav.monitor", keywords: ["realtime", "watch"] },
  { path: "/evasion", labelKey: "nav.securityAntiDpi", keywords: ["reality", "clean-ip", "tls", "decoy", "probing", "anti-dpi", "fragment", "gfw"] },
  { path: "/ip-limit", labelKey: "nav.ipLimit", keywords: ["share", "device", "ip", "limit", "shareguard"] },
  { path: "/security", labelKey: "nav.security", keywords: ["threat", "hardening", "waf", "score"] },
  { path: "/smart-quota", labelKey: "nav.smartQuota", keywords: ["fair use", "throttle", "speed"] },
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
  { path: "/logs", labelKey: "nav.logs", keywords: ["debug", "error"] },
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

  const items: PaletteItem[] = useMemo(() => ROUTES.map(r => ({
    id: r.path,
    label: t(r.labelKey),
    sublabel: r.path,
    action: () => { navigate(r.path); setOpen(false); },
    keywords: r.keywords,
  })), [t, navigate]);

  const filtered = query.trim()
    ? items.filter(item => {
        const q = query.toLowerCase();
        return item.label.toLowerCase().includes(q)
          || item.sublabel?.toLowerCase().includes(q)
          || item.keywords?.some(k => k.includes(q));
      })
    : items.slice(0, 10);

  function handleKeyDown(e: React.KeyboardEvent) {
    if (e.key === "ArrowDown") { e.preventDefault(); setSelected(s => Math.min(s + 1, filtered.length - 1)); }
    if (e.key === "ArrowUp") { e.preventDefault(); setSelected(s => Math.max(s - 1, 0)); }
    if (e.key === "Enter" && filtered[selected]) { filtered[selected].action(); }
  }

  if (!open) return null;

  return (
    <div className="fixed inset-0 z-[100] flex items-start justify-center pt-[20vh] animate-fade-in" onClick={() => setOpen(false)}>
      <div className="fixed inset-0 bg-black/50 backdrop-blur-sm" />
      <div className="relative w-full max-w-lg rounded-2xl border border-border/60 bg-bg-elevated/95 shadow-2xl backdrop-blur-xl animate-scale-in" onClick={e => e.stopPropagation()}>
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

        <div className="max-h-[320px] overflow-y-auto p-2">
          {filtered.length === 0 && (
            <div className="py-8 text-center text-sm text-fg-muted">{t("cmdPalette.noResults")}</div>
          )}
          {filtered.map((item, i) => (
            <button
              key={item.id}
              onClick={item.action}
              className={cn(
                "flex w-full items-center gap-3 rounded-xl px-3 py-2.5 text-left transition",
                i === selected ? "bg-primary/10 text-primary" : "text-fg-muted hover:bg-surface-2/60 hover:text-fg",
              )}
            >
              <div className="flex-1">
                <div className="text-sm font-medium">{item.label}</div>
                <div className="text-xs text-fg-subtle">{item.sublabel}</div>
              </div>
              {i === selected && <span className="text-[10px] text-fg-subtle">↵</span>}
            </button>
          ))}
        </div>

        <div className="flex items-center justify-between border-t border-border/40 px-4 py-2 text-[10px] text-fg-subtle">
          <span>{t("cmdPalette.navigate")}</span>
          <span>{t("cmdPalette.select")}</span>
          <span className="flex items-center gap-1"><Command size={10} />{t("cmdPalette.toggle")}</span>
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
