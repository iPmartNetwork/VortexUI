import { useState } from "react";
import { NavLink, Outlet, useNavigate } from "react-router-dom";
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
} from "lucide-react";
import { cn } from "@/lib/utils";
import { useAuth } from "@/auth/auth";
import { useTheme } from "@/theme/theme";
import { useI18n } from "@/i18n/i18n";
import { useLiveEvents } from "@/api/live";
import type { TKey, Lang } from "@/i18n/dict";

const PANEL_VERSION = "1.0.1";

const nav: { to: string; key: TKey; icon: React.ElementType }[] = [
  { to: "/overview", key: "nav.overview", icon: LayoutDashboard },
  { to: "/users", key: "nav.users", icon: UsersIcon },
  { to: "/nodes", key: "nav.nodes", icon: Server },
  { to: "/outbounds", key: "nav.outbounds", icon: Network },
  { to: "/routing", key: "nav.routing", icon: RouteIcon },
  { to: "/balancers", key: "nav.balancers", icon: Scale },
  { to: "/admins", key: "nav.admins", icon: ShieldCheck },
  { to: "/plans", key: "nav.plans" as any, icon: Network },
  { to: "/orders", key: "nav.orders" as any, icon: History },
  { to: "/evasion", key: "nav.evasion" as any, icon: ShieldCheck },
  { to: "/monitor", key: "nav.monitor" as any, icon: Server },
  { to: "/audit", key: "nav.audit", icon: History },
  { to: "/logs", key: "nav.logs", icon: ScrollText },
  { to: "/settings", key: "nav.settings", icon: SettingsIcon },
];

function IconButton({ onClick, label, children, className }: { onClick: () => void; label: string; children: React.ReactNode; className?: string }) {
  return (
    <button onClick={onClick} aria-label={label} title={label} className={cn("grid h-9 w-9 place-items-center rounded-xl text-fg-muted transition hover:bg-surface-2/60 hover:text-fg", className)}>
      {children}
    </button>
  );
}

export function Layout() {
  const { logout } = useAuth();
  const navigate = useNavigate();
  const { resolved, toggle } = useTheme();
  const { t, lang, setLang } = useI18n();
  const [collapsed, setCollapsed] = useState(false);
  const [mobileOpen, setMobileOpen] = useState(false);
  useLiveEvents(); // subscribe to the SSE event stream for live updates

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

      {/* Navigation */}
      <nav className="flex-1 space-y-1 overflow-y-auto px-2 py-2">
        {nav.map((n) => {
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
                  collapsed ? "justify-center px-0 py-2.5" : "gap-3 px-3 py-2.5",
                  isActive ? "nav-glow bg-primary/[0.08] text-primary" : "text-fg-muted hover:bg-surface-2/50 hover:text-fg",
                )
              }
            >
              <Icon size={18} strokeWidth={2} />
              {!collapsed && <span>{t(n.key)}</span>}
            </NavLink>
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
            <div className="text-[10px] font-semibold text-fg-subtle">VortexUI v{PANEL_VERSION}</div>
            <div className="mt-0.5 text-[9px] text-fg-subtle/70">© {new Date().getFullYear()} iPmart Network. All rights reserved.</div>
          </div>
        )}
        {collapsed && (
          <div className="mt-2 text-center text-[9px] text-fg-subtle/60">v{PANEL_VERSION}</div>
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
