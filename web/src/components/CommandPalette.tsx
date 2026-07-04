import { useState, useEffect, useRef } from "react";
import { useNavigate } from "react-router-dom";
import { Search, Command } from "lucide-react";
import { cn } from "@/lib/utils";

interface PaletteItem {
  id: string;
  label: string;
  sublabel?: string;
  icon?: React.ReactNode;
  action: () => void;
  keywords?: string[];
}

const ROUTES: { path: string; label: string; keywords: string[] }[] = [
  { path: "/overview", label: "Dashboard", keywords: ["home", "overview", "stats"] },
  { path: "/users", label: "Users", keywords: ["accounts", "subscribers"] },
  { path: "/nodes", label: "Nodes", keywords: ["servers", "agents"] },
  { path: "/routing?tab=outbounds", label: "Outbounds", keywords: ["egress", "proxy", "outbound"] },
  { path: "/routing", label: "Routing & Load Balancers", keywords: ["routes", "policy", "packs", "load", "balance", "balancers", "outbounds"] },
  { path: "/wallet-billing", label: "Reseller Platform & Shop", keywords: ["wallet", "billing", "plans", "orders", "deposit", "reseller"] },
  { path: "/orders", label: "Orders", keywords: ["payments", "purchases"] },
  { path: "/analytics", label: "Analytics", keywords: ["stats", "traffic", "geo"] },
  { path: "/tickets", label: "Support Tickets", keywords: ["help", "support"] },
  { path: "/monitor", label: "Live Monitor", keywords: ["realtime", "watch"] },
  { path: "/evasion", label: "Security & Anti-Censorship", keywords: ["reality", "clean-ip", "tls", "decoy", "probing", "anti-dpi", "fragment", "gfw"] },
  { path: "/smart-quota", label: "Smart Quota", keywords: ["fair use", "throttle", "speed"] },
  { path: "/relay-chains", label: "Relay Chains", keywords: ["cdn", "relay", "chain"] },
  { path: "/migration", label: "Auto Migration", keywords: ["failover", "health"] },
  { path: "/family-groups", label: "Family Groups", keywords: ["shared", "pool", "group"] },
  { path: "/referrals", label: "Referral System", keywords: ["invite", "reward", "code"] },
  { path: "/doh", label: "DNS-over-HTTPS", keywords: ["dns", "privacy", "doh"] },
  { path: "/sni-manager", label: "SNI & SSL Manager", keywords: ["domain", "cert", "tls"] },
  { path: "/fingerprint", label: "Fingerprint Validator", keywords: ["ja3", "client", "tls"] },
  { path: "/federation", label: "Panel Federation", keywords: ["multi", "sync", "peer"] },
  { path: "/deep-links", label: "Deep Links & QR", keywords: ["qr", "import", "app"] },
  { path: "/quota-notifications", label: "Quota Notifications", keywords: ["alert", "notify"] },
  { path: "/settings?tab=admins", label: "Admin Management", keywords: ["roles", "permissions", "admins"] },
  { path: "/logs", label: "System Logs", keywords: ["debug", "error"] },
  { path: "/settings", label: "Settings", keywords: ["config", "branding", "admins"] },
];

export function CommandPalette() {
  const [open, setOpen] = useState(false);
  const [query, setQuery] = useState("");
  const [selected, setSelected] = useState(0);
  const inputRef = useRef<HTMLInputElement>(null);
  const navigate = useNavigate();

  // ⌘K / Ctrl+K to open
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

  const items: PaletteItem[] = ROUTES.map(r => ({
    id: r.path,
    label: r.label,
    sublabel: r.path,
    action: () => { navigate(r.path); setOpen(false); },
    keywords: r.keywords,
  }));

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
        {/* Search input */}
        <div className="flex items-center gap-3 border-b border-border/40 px-4 py-3">
          <Search size={18} className="text-fg-subtle" />
          <input
            ref={inputRef}
            value={query}
            onChange={e => { setQuery(e.target.value); setSelected(0); }}
            onKeyDown={handleKeyDown}
            placeholder="Search pages, users, settings..."
            className="flex-1 bg-transparent text-sm text-fg outline-none placeholder:text-fg-subtle"
          />
          <kbd className="hidden sm:inline-flex items-center gap-0.5 rounded-md border border-border/60 bg-surface-2/50 px-1.5 py-0.5 text-[10px] text-fg-subtle">
            ESC
          </kbd>
        </div>

        {/* Results */}
        <div className="max-h-[320px] overflow-y-auto p-2">
          {filtered.length === 0 && (
            <div className="py-8 text-center text-sm text-fg-muted">No results found</div>
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

        {/* Footer */}
        <div className="flex items-center justify-between border-t border-border/40 px-4 py-2 text-[10px] text-fg-subtle">
          <span>↑↓ navigate</span>
          <span>↵ select</span>
          <span className="flex items-center gap-1"><Command size={10} />K to toggle</span>
        </div>
      </div>
    </div>
  );
}

// Trigger button for the sidebar/header
export function CommandPaletteTrigger() {
  return (
    <button
      onClick={() => window.dispatchEvent(new KeyboardEvent("keydown", { key: "k", metaKey: true }))}
      className="flex items-center gap-2 rounded-xl border border-border/50 bg-surface-2/30 px-3 py-1.5 text-xs text-fg-muted transition hover:bg-surface-2/60 hover:text-fg"
    >
      <Search size={13} />
      <span className="hidden md:inline">Search...</span>
      <kbd className="hidden md:inline-flex items-center rounded border border-border/50 bg-surface/50 px-1 text-[10px]">⌘K</kbd>
    </button>
  );
}
