import { useState } from "react";
import { NavLink, Outlet, useNavigate, useLocation } from "react-router-dom";
import {
  LayoutDashboard,
  Users as UsersIcon,
  Server,
  Network,
  Route as RouteIcon,
  Scale,
  ShieldCheck,
  ScrollText,
  History,
  Settings as SettingsIcon,
  LogOut,
  Moon,
  Sun,
  PanelLeftClose,
  PanelLeftOpen,
  Menu,
  X,
  Globe,
  Scan,
  Gauge,
  Link2,
  EyeOff,
  BarChart3,
  LifeBuoy,
  ArrowRightLeft,
  Shield,
  Users2,
  Gift,
  Wifi,
  Lock,
  Fingerprint as FingerprintIcon,
  Layers,
  QrCode,
  Bell,
  Unplug,
  Ban,
  ChevronDown,
  Wallet,
} from "lucide-react";
import { cn } from "@/lib/utils";
import { useAuth } from "@/auth/auth";
import { canAccessRoute } from "@/auth/permissions";
import { useTheme } from "@/theme/theme";
import { useI18n } from "@/i18n/i18n";
import { useLiveEvents } from "@/api/live";
import { useVersion } from "@/api/hooks";
import { useStopImpersonation } from "@/api/reseller-hooks";
import { setToken } from "@/api/client";
import type { TKey, Lang } from "@/i18n/dict";

// PANEL_VERSION is the fallback shown until the backend version is fetched.
const PANEL_VERSION = "1.2.5";

// Grouped navigation with collapsible sections
interface NavItem { to: string; key: TKey; icon: React.ElementType }
interface NavSection { label: string; id: string; items: NavItem[] }

function buildNavSections(sudo: boolean): NavSection[] {
  const resellerSection: NavSection | null = sudo ? null : {
    label: "nav.section.reseller",
    id: "reseller",
    items: [
      { to: "/reseller-dashboard", key: "nav.resellerDashboard", icon: Gauge },
      { to: "/reseller-account", key: "nav.resellerAccount", icon: Wallet },
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
        { to: "/plans", key: "nav.plans", icon: Network },
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
      { to: "/routing", key: "nav.routing", icon: RouteIcon },
      { to: "/routing-packs", key: "nav.routingPacks", icon: RouteIcon },
      { to: "/balancers", key: "nav.balancers", icon: Scale },
      { to: "/relay-chains", key: "nav.relayChains", icon: Link2 },
      { to: "/migration", key: "nav.migration", icon: ArrowRightLeft },
      { to: "/federation", key: "nav.federation", icon: Layers },
    ],
  },
  {
    label: "nav.section.security",
    id: "security",
    items: [
      { to: "/evasion", key: "nav.evasion", icon: ShieldCheck },
      { to: "/tls-tricks", key: "nav.tlsTricks", icon: Unplug },
      { to: "/sni-manager", key: "nav.sniManager", icon: Lock },
      { to: "/reality-scanner", key: "nav.realityScanner", icon: Scan },
      { to: "/clean-ip", key: "nav.cleanIpScanner", icon: Globe },
      { to: "/decoy-website", key: "nav.decoyWebsite", icon: EyeOff },
      { to: "/probing-protection", key: "nav.probingProtection", icon: Shield },
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

function IconButton({ onClick, label, children, className }: { onClick: () => void; label: string; children: React.ReactNode; className?: string }) {
  return (
    <button onClick={onClick} aria-label={label} title={label} className={cn("grid h-9 w-9 place-items-center rounded-xl text-fg-muted transition hover:bg-surface-2/60 hover:text-fg", className)}>
      {children}
    </button>
  );
}

export function Layout() {
  const { logout, sudo, permissions, impersonating, session, refreshSession } = useAuth();
  const stopImpersonate = useStopImpersonation();
  const navigate = useNavigate();
  const location = useLocation();
  const { resolved, toggle } = useTheme();
  const { t, lang, setLang } = useI18n();
  const [collapsed, setCollapsed] = useState(false);
  const [mobileOpen, setMobileOpen] = useState(false);
  const [openSections, setOpenSections] = useState<Record<string, boolean>>(() => {
    const init: Record<string, boolean> = {};
    const sections = buildNavSections(sudo);
    sections.forEach(s => {
      init[s.id] = s.items.some(item => location.pathname.startsWith(item.to));
    });
    init["dashboard"] = true;
    if (!sudo) init["reseller"] = true;
    return init;
  });
  useLiveEvents();
  const version = useVersion().data ?? PANEL_VERSION;

  function toggleSection(id: string) {
    setOpenSections(prev => ({ ...prev, [id]: !prev[id] }));
  }

  const visibleSections = buildNavSections(sudo)
    .map((section) => ({
      ...section,
      items: section.items.filter((item) => canAccessRoute(item.to, sudo, permissions, session?.admin.reseller_settings)),
    }))
    .filter((section) => section.items.length > 0);

  async function exitImpersonation() {
    try {
      const res = await stopImpersonate.mutateAsync();
      setToken(res.token);
      await refreshSession();
    } catch {
      logout();
      navigate("/login");
    }
  }

  const sidebarContent = (
    <>
      {/* Logo */}
      <div className={cn("flex items-center gap-3 px-4 py-5", collapsed && "justify-center px-2")}>
        {!collapsed && (
          <div className="min-w-0">
            <div className="text-base font-black tracking-wider" style={{ fontFamily: "'Orbitron', sans-serif" }}>VortexUI</div>
          </div>
        )}
        {collapsed && (
          <div className="text-sm font-black text-fg" style={{ fontFamily: "'Orbitron', sans-serif" }}>V</div>
        )}
      </div>

      {/* Navigation — collapsible sections */}
      <nav className="flex-1 overflow-y-auto px-2 py-2 space-y-1">
        {visibleSections.map((section) => {
          const isOpen = openSections[section.id] ?? false;
          const hasActive = section.items.some(item => location.pathname.startsWith(item.to));

          return (
            <div key={section.id}>
              {/* Section header */}
              {!collapsed && (
                <button
                  onClick={() => toggleSection(section.id)}
                  className={cn(
                    "flex w-full items-center justify-between rounded-lg px-3 py-2 text-[11px] font-semibold uppercase tracking-wider transition",
                    hasActive ? "text-primary/80" : "text-fg-subtle hover:text-fg-muted",
                  )}
                >
                  <span>{t(section.label as TKey)}</span>
                  <ChevronDown size={12} className={cn("transition-transform duration-200", isOpen && "rotate-180")} />
                </button>
              )}

              {/* Section items */}
              {(collapsed || isOpen) && (
                <div className={cn("space-y-0.5", !collapsed && "pb-2")}>
                  {section.items.map((n) => {
                    const Icon = n.icon;
                    return (
                      <NavLink
                        key={n.to}
                        to={n.to}
                        title={collapsed ? t(n.key) : undefined}
                        onClick={() => setMobileOpen(false)}
                        className={({ isActive }) =>
                          cn(
                            "group relative flex items-center rounded-xl text-sm font-medium transition-all duration-200",
                            collapsed ? "justify-center px-0 py-2.5" : "gap-3 px-3 py-2",
                            isActive ? "nav-glow bg-primary/[0.08] text-primary" : "text-fg-muted hover:bg-surface-2/50 hover:text-fg",
                          )
                        }
                      >
                        <Icon size={16} strokeWidth={2} />
                        {!collapsed && <span className="truncate">{t(n.key)}</span>}
                      </NavLink>
                    );
                  })}
                </div>
              )}
            </div>
          );
        })}
      </nav>

      {/* Footer */}
      <div className="border-t border-border/50 px-2 pb-3 pt-2">
        <div className={cn("flex items-center", collapsed ? "flex-col gap-1" : "justify-between")}>
          <div className={cn("flex", collapsed ? "flex-col gap-1" : "gap-0.5")}>
            <IconButton onClick={toggle} label={t("theme.label")}>
              {resolved === "dark" ? <Sun size={16} /> : <Moon size={16} />}
            </IconButton>
            {!collapsed && (
              <LangMenu lang={lang as Lang} setLang={setLang} />
            )}
          </div>
          <div className={cn("flex", collapsed ? "flex-col gap-1" : "gap-0.5")}>
            <IconButton onClick={() => setCollapsed(!collapsed)} label="Toggle sidebar" className="hidden lg:grid">
              {collapsed ? <PanelLeftOpen size={16} /> : <PanelLeftClose size={16} />}
            </IconButton>
            {!collapsed && (
              <IconButton onClick={() => { logout(); navigate("/login"); }} label={t("nav.signout")}>
                <LogOut size={16} />
              </IconButton>
            )}
          </div>
        </div>
        {!collapsed && (
          <div className="mt-3 rounded-lg bg-surface-2/40 px-3 py-2 text-center">
            <div className="text-[10px] font-semibold text-fg-subtle">VortexUI v{version}</div>
            <div className="mt-0.5 text-[9px] text-fg-subtle/70">© {new Date().getFullYear()} iPmart Network. All rights reserved.</div>
          </div>
        )}
        {collapsed && (
          <div className="mt-2 text-center text-[9px] text-fg-subtle/60">v{version}</div>
        )}
      </div>
    </>
  );

  return (
    <div className="flex min-h-screen">
      {/* Mobile overlay */}
      {mobileOpen && (
        <div className="fixed inset-0 z-40 bg-black/50 backdrop-blur-sm lg:hidden" onClick={() => setMobileOpen(false)} />
      )}

      {/* Sidebar — desktop */}
      <aside className={cn(
        "hidden lg:flex flex-col border-e border-border/60 bg-bg-elevated/50 backdrop-blur-2xl transition-all duration-300 dark:bg-bg-elevated/40",
        collapsed ? "w-[68px]" : "w-64",
      )}>
        {sidebarContent}
      </aside>

      {/* Sidebar — mobile drawer */}
      <aside className={cn(
        "fixed inset-y-0 start-0 z-50 flex w-64 flex-col border-e border-border/60 bg-bg-elevated/90 backdrop-blur-2xl transition-transform duration-300 lg:hidden dark:bg-bg-elevated/95",
        mobileOpen ? "translate-x-0" : "-translate-x-full rtl:translate-x-full",
      )}>
        <button onClick={() => setMobileOpen(false)} className="absolute end-3 top-3 grid h-8 w-8 place-items-center rounded-lg text-fg-muted hover:bg-surface-2 hover:text-fg">
          <X size={18} />
        </button>
        {sidebarContent}
      </aside>

      {/* Main content */}
      <main className="flex-1 overflow-auto">
        {/* Mobile top bar */}
        <div className="sticky top-0 z-30 flex items-center gap-3 border-b border-border/40 bg-bg/80 px-4 py-3 backdrop-blur-lg lg:hidden">
          <button onClick={() => setMobileOpen(true)} className="grid h-9 w-9 place-items-center rounded-xl text-fg-muted hover:bg-surface-2/60 hover:text-fg">
            <Menu size={20} />
          </button>
          <span className="text-sm font-black tracking-wider" style={{ fontFamily: "'Orbitron', sans-serif" }}>VortexUI</span>
        </div>
        <div className="mx-auto max-w-7xl px-4 py-6 lg:px-8 lg:py-8 animate-fade-in">
          {impersonating && (
            <div className="mb-4 flex flex-wrap items-center justify-between gap-3 rounded-xl border border-amber-500/40 bg-amber-500/10 px-4 py-3 text-sm">
              <span>
                Impersonating <strong>@{session?.admin.username}</strong> (support session)
              </span>
              <button type="button" onClick={() => void exitImpersonation()} className="rounded-lg bg-amber-600 px-3 py-1.5 text-xs font-medium text-white hover:bg-amber-500">
                Exit impersonation
              </button>
            </div>
          )}
          <Outlet />
        </div>
      </main>
    </div>
  );
}

const LANG_OPTIONS: { code: Lang; label: string }[] = [
  { code: "en", label: "English" },
  { code: "fa", label: "فارسی" },
  { code: "tr", label: "Türkçe" },
  { code: "ar", label: "العربية" },
  { code: "ru", label: "Русский" },
  { code: "zh", label: "中文" },
  { code: "ja", label: "日本語" },
  { code: "es", label: "Español" },
];

function LangMenu({ lang, setLang }: { lang: Lang; setLang: (l: Lang) => void }) {
  const [open, setOpen] = useState(false);
  return (
    <div className="relative">
      <button
        onClick={() => setOpen(!open)}
        className="grid h-9 w-9 place-items-center rounded-xl text-fg-muted transition hover:bg-surface-2/60 hover:text-fg"
        title="Language"
      >
        <Globe size={16} />
      </button>
      {open && (
        <>
          <div className="fixed inset-0 z-40" onClick={() => setOpen(false)} />
          <div className="absolute bottom-full start-0 z-50 mb-2 w-36 rounded-xl border border-border/60 bg-bg-elevated/95 p-1 shadow-xl backdrop-blur-xl animate-scale-in">
            {LANG_OPTIONS.map((l) => (
              <button
                key={l.code}
                onClick={() => { setLang(l.code); setOpen(false); }}
                className={cn(
                  "flex w-full items-center gap-2 rounded-lg px-3 py-2 text-xs font-medium transition",
                  lang === l.code ? "bg-primary/10 text-primary" : "text-fg-muted hover:bg-surface-2/60 hover:text-fg",
                )}
              >
                <span className="w-5 text-[10px] font-bold text-fg-subtle">{l.code.toUpperCase()}</span>
                {l.label}
              </button>
            ))}
          </div>
        </>
      )}
    </div>
  );
}
