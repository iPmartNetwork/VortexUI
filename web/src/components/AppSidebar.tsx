import { useState } from "react";
import { NavLink, useLocation, useNavigate } from "react-router-dom";
import { motion, AnimatePresence } from "framer-motion";
import {
  Zap,
  Search,
  Menu,
  X,
  LogOut,
  Smartphone,
  ChevronsLeft,
  ChevronsRight,
  ExternalLink,
  Moon,
  Sun,
  Globe,
} from "lucide-react";
import { cn } from "@/lib/utils";
import { useAuth } from "@/auth/auth";
import { canAccessRoute } from "@/auth/permissions";
import { useTheme } from "@/theme/theme";
import { useI18n } from "@/i18n/i18n";
import { useVersion, useNodes } from "@/api/hooks";
import { buildCompactNavSections } from "@/navigation/nav-sections-compact";
import { useOverview } from "@/api/policy-hooks";
import type { Overview } from "@/api/types";
import { openCommandPalette } from "@/lib/commandPalette";
import { LANG_OPTIONS } from "@/i18n/lang-options";

const PANEL_VERSION = "1.2.8";

function navBadgeCount(badges: Overview["widgets"]["nav_badges"] | undefined, key?: string): number {
  if (!badges || !key) return 0;
  switch (key) {
    case "active_users":
      return badges.active_users;
    case "open_tickets":
      return badges.open_tickets;
    case "pending_orders":
      return badges.pending_orders;
    default:
      return 0;
  }
}

interface AppSidebarProps {
  mobileOpen: boolean;
  onMobileOpenChange: (open: boolean) => void;
}

export function AppSidebar({ mobileOpen, onMobileOpenChange }: AppSidebarProps) {
  const { logout, sudo, permissions, session } = useAuth();
  const navigate = useNavigate();
  const location = useLocation();
  const { resolved, toggle: toggleTheme } = useTheme();
  const { t, lang, setLang } = useI18n();
  const version = useVersion().data ?? PANEL_VERSION;
  const nodesQ = useNodes();
  const overviewQ = useOverview();
  const navBadges = overviewQ.data?.widgets?.nav_badges;
  const primaryCore = nodesQ.data?.nodes?.find((n) => n.core === "xray") ?? nodesQ.data?.nodes?.[0];
  const coreOnline = primaryCore?.health?.core_running ?? false;
  const coreVer = primaryCore?.core_version || "—";
  const [collapsed, setCollapsed] = useState(false);
  const [hovered, setHovered] = useState<string | null>(null);
  const [langOpen, setLangOpen] = useState(false);

  const visibleSections = buildCompactNavSections(sudo)
    .map((section) => ({
      ...section,
      items: section.items.filter((item) =>
        canAccessRoute(item.to, sudo, permissions, session?.admin.reseller_settings),
      ),
    }))
    .filter((section) => section.items.length > 0);

  const username = session?.admin.username ?? "Admin";
  const initials = username.slice(0, 2).toUpperCase();

  function signOut() {
    logout();
    navigate("/login");
  }

  function inner(mobile: boolean) {
    const mini = collapsed && !mobile;

    return (
      <div className="flex flex-col h-full select-none">
        <div className={cn("flex items-center flex-shrink-0 h-14", mini ? "justify-center" : "px-4 gap-3")}>
          {mini ? (
            <div className="h-8 w-8 rounded-[10px] grad-bg flex items-center justify-center">
              <Zap size={15} className="text-primary-fg" />
            </div>
          ) : (
            <>
              <div className="h-8 w-8 rounded-[10px] grad-bg flex items-center justify-center flex-shrink-0">
                <Zap size={15} className="text-primary-fg" />
              </div>
              <div className="min-w-0">
                <p className="text-sm font-semibold text-fg leading-none tracking-tight">
                  Vortex<span className="text-primary">UI</span>
                </p>
                <p className="text-[10px] text-fg-subtle mt-0.5 leading-none">{t("app.taglineVeltrix")}</p>
                <p className="text-[10px] text-primary/80 mt-0.5 leading-none font-medium">v{version}</p>
              </div>
            </>
          )}
        </div>

        {!mini && (
          <div className="px-3 pb-3">
            <button
              type="button"
              onClick={openCommandPalette}
              className="w-full flex items-center gap-2 h-8 px-2.5 rounded-lg bg-surface-2/50 border border-border/60 text-fg-subtle text-xs hover:border-border-strong transition-colors"
            >
              <Search size={13} className="flex-shrink-0 opacity-50" />
              <span className="flex-1 text-start truncate opacity-60">{t("shell.search")}</span>
              <kbd className="text-[9px] opacity-40 font-mono border border-border rounded px-1 py-px">⌘K</kbd>
            </button>
          </div>
        )}

        <nav className={cn("flex-1 overflow-y-auto", mini ? "px-1.5 py-2" : "px-2.5 py-1")}>
          <div className={mini ? "space-y-3" : "space-y-5"}>
            {visibleSections.map((sec, si) => (
              <div key={sec.id}>
                {!mini ? (
                  <p className="px-2 mb-1 text-[10px] font-semibold uppercase tracking-[0.08em] text-fg-subtle/60">
                    {t(sec.label)}
                  </p>
                ) : si > 0 ? (
                  <div className="h-px bg-border/40 mx-2 mb-1.5" />
                ) : null}

                <div className={mini ? "space-y-1 flex flex-col items-center" : "space-y-px"}>
                  {sec.items.map((item) => {
                    const Icon = item.icon;
                    const [itemPath, itemQuery] = item.to.split("?");
                    const active =
                      location.pathname.startsWith(itemPath) &&
                      (itemQuery
                        ? location.search.includes(itemQuery)
                        : !location.search.includes("tab=inbounds"));
                    const label = t(item.key);
                    const badge = navBadgeCount(navBadges, item.badgeKey);

                    return (
                      <div key={item.to} className="relative w-full">
                        <NavLink
                          to={item.to}
                          title={mini ? label : undefined}
                          onClick={() => onMobileOpenChange(false)}
                          onMouseEnter={() => setHovered(item.to)}
                          onMouseLeave={() => setHovered(null)}
                          className={cn(
                            "relative flex w-full items-center transition-all duration-[120ms]",
                            mini ? "h-9 w-9 mx-auto justify-center rounded-[10px]" : "h-9 gap-2.5 px-2.5 rounded-[10px]",
                            active
                              ? "bg-primary/[0.12] text-fg font-semibold"
                              : "text-fg-muted hover:text-fg hover:bg-surface-2",
                          )}
                        >
                          {active && !mini && (
                            <motion.div
                              layoutId={mobile ? "m-bar" : "d-bar"}
                              className="absolute start-0 inset-y-1 w-[2.5px] rounded-full bg-primary"
                              transition={{ type: "spring", stiffness: 500, damping: 32 }}
                            />
                          )}
                          {active && mini && (
                            <motion.div
                              layoutId="c-ring"
                              className="absolute inset-0 rounded-[10px] ring-[1.5px] ring-primary/50 pointer-events-none"
                              transition={{ type: "spring", stiffness: 500, damping: 32 }}
                            />
                          )}
                          <span className={cn("flex-shrink-0 transition-colors duration-100", active && "text-primary")}>
                            <Icon size={18} strokeWidth={2} />
                          </span>
                          {!mini && (
                            <span className="flex-1 text-start text-[13px] truncate">{label}</span>
                          )}
                          {!mini && badge > 0 && (
                            <span className="min-w-[18px] h-[18px] px-1 rounded-full bg-primary text-primary-fg text-[10px] font-bold flex items-center justify-center tabular-nums">
                              {badge > 99 ? "99+" : badge}
                            </span>
                          )}
                        </NavLink>

                        {mini && (
                          <AnimatePresence>
                            {hovered === item.to && (
                              <motion.div
                                initial={{ opacity: 0, x: lang === "fa" || lang === "ar" ? 4 : -4 }}
                                animate={{ opacity: 1, x: 0 }}
                                exit={{ opacity: 0, x: lang === "fa" || lang === "ar" ? 4 : -4 }}
                                transition={{ duration: 0.1 }}
                                className="absolute start-full ms-2.5 top-1/2 -translate-y-1/2 z-[60] pointer-events-none"
                              >
                                <div className="flex items-center gap-2 rounded-lg bg-fg text-bg px-2.5 py-1.5 shadow-xl text-[11px] font-medium whitespace-nowrap">
                                  {label}
                                </div>
                              </motion.div>
                            )}
                          </AnimatePresence>
                        )}
                      </div>
                    );
                  })}
                </div>
              </div>
            ))}
          </div>
        </nav>

        {!mini && (
          <div className="px-3 pb-2 flex-shrink-0 space-y-2">
            <button
              type="button"
              onClick={() => navigate("/portal/login")}
              className="w-full h-9 rounded-xl text-[12px] font-semibold text-primary flex items-center justify-center gap-1.5 border border-primary/20 bg-primary/[0.06] hover:bg-primary/10 transition-all"
            >
              <Smartphone size={14} />
              {t("shell.selfServicePortal")}
              <ExternalLink size={10} className="opacity-50" />
            </button>
            {primaryCore && (
              <div className="rounded-xl border border-border/70 bg-surface-2/50 px-3 py-2.5">
                <div className="flex items-center gap-2">
                  <span className={`h-2 w-2 rounded-full ${coreOnline ? "bg-success shadow-[0_0_6px] shadow-success/50" : "bg-fg-subtle"}`} />
                  <span className="text-xs font-bold text-fg truncate">
                    {primaryCore.core === "singbox" ? "sing-box" : "Xray"} {coreVer}
                  </span>
                </div>
                <p className="text-[10px] text-fg-subtle mt-1">
                  {coreOnline ? t("overview.coreOnline") : t("overview.coreOffline")}
                </p>
              </div>
            )}
          </div>
        )}

        <div
          className={cn(
            "flex-shrink-0 border-t border-border/40",
            mini ? "py-3 flex flex-col items-center gap-1.5" : "p-2.5 space-y-1",
          )}
        >
          {mini ? (
            <>
              <div className="relative">
                <div className="h-8 w-8 rounded-[10px] bg-gradient-to-br from-violet-500 to-indigo-500 flex items-center justify-center text-white text-[10px] font-bold">
                  {initials}
                </div>
                <span className="absolute -bottom-px -end-px h-2 w-2 rounded-full bg-success ring-[1.5px] ring-[var(--sidebar-bg)]" />
              </div>
              <button
                type="button"
                onClick={toggleTheme}
                className="h-7 w-7 rounded-lg flex items-center justify-center text-fg-subtle hover:text-fg hover:bg-surface-2 transition-colors"
                title={t("theme.label")}
              >
                {resolved === "dark" ? <Sun size={13} /> : <Moon size={13} />}
              </button>
              <button
                type="button"
                onClick={signOut}
                className="h-7 w-7 rounded-lg flex items-center justify-center text-fg-subtle hover:text-danger hover:bg-danger/10 transition-colors"
                title={t("nav.signout")}
              >
                <LogOut size={13} />
              </button>
            </>
          ) : (
            <>
              <div className="flex items-center gap-2 px-2 py-1.5">
                <div className="relative flex-shrink-0">
                  <div className="h-8 w-8 rounded-[10px] bg-gradient-to-br from-violet-500 to-indigo-500 flex items-center justify-center text-white text-[10px] font-bold shadow-sm">
                    {initials}
                  </div>
                  <span className="absolute -bottom-px -end-px h-2.5 w-2.5 rounded-full bg-success ring-[1.5px] ring-[var(--sidebar-bg)]" />
                </div>
                <div className="flex-1 min-w-0 text-start">
                  <p className="text-[13px] font-semibold text-fg leading-tight truncate">{username}</p>
                  <p className="text-[10px] text-fg-subtle leading-tight truncate">
                    {sudo ? t("shell.superAdmin") : t("shell.reseller")}
                  </p>
                </div>
              </div>
              <div className="flex items-center gap-1 px-1">
                <button
                  type="button"
                  onClick={toggleTheme}
                  className="h-8 flex-1 rounded-lg flex items-center justify-center text-fg-subtle hover:text-fg hover:bg-surface-2 transition-colors"
                  title={t("theme.label")}
                >
                  {resolved === "dark" ? <Sun size={14} /> : <Moon size={14} />}
                </button>
                <div className="relative flex-1">
                  <button
                    type="button"
                    onClick={() => setLangOpen(!langOpen)}
                    className="h-8 w-full rounded-lg flex items-center justify-center text-fg-subtle hover:text-fg hover:bg-surface-2 transition-colors"
                    title="Language"
                  >
                    <Globe size={14} />
                  </button>
                  {langOpen && (
                    <>
                      <div className="fixed inset-0 z-40" onClick={() => setLangOpen(false)} />
                      <div className="absolute bottom-full start-0 z-50 mb-2 w-36 rounded-xl border border-border/60 bg-bg-elevated/95 p-1 shadow-xl backdrop-blur-xl animate-scale-in">
                        {LANG_OPTIONS.map((l) => (
                          <button
                            key={l.code}
                            type="button"
                            onClick={() => {
                              setLang(l.code);
                              setLangOpen(false);
                            }}
                            className={cn(
                              "flex w-full items-center gap-2 rounded-lg px-3 py-2 text-xs font-medium transition",
                              lang === l.code
                                ? "bg-primary/10 text-primary"
                                : "text-fg-muted hover:bg-surface-2/60 hover:text-fg",
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
                <button
                  type="button"
                  onClick={signOut}
                  className="h-8 flex-1 rounded-lg flex items-center justify-center text-fg-subtle hover:text-danger hover:bg-danger/10 transition-colors"
                  title={t("nav.signout")}
                >
                  <LogOut size={14} />
                </button>
              </div>
            </>
          )}
        </div>

        {!mobile && (
          <button
            type="button"
            onClick={() => setCollapsed((c) => !c)}
            className={cn(
              "hidden md:flex absolute z-10 items-center justify-center h-5 w-5 rounded-full",
              "bg-bg-elevated border border-border/80 shadow text-fg-subtle/70 hover:text-fg hover:border-border-strong",
              "transition-all hover:scale-110 -end-2.5 top-[58px]",
            )}
          >
            {collapsed ? <ChevronsRight size={10} strokeWidth={2.5} /> : <ChevronsLeft size={10} strokeWidth={2.5} />}
          </button>
        )}
      </div>
    );
  }

  return (
    <>
      <button
        type="button"
        onClick={() => onMobileOpenChange(true)}
        className="md:hidden fixed top-3 start-3 z-50 h-10 w-10 rounded-xl bg-bg-elevated/90 backdrop-blur-lg border border-border flex items-center justify-center text-fg shadow-lg active:scale-95 transition-transform"
      >
        <Menu size={18} />
      </button>

      <AnimatePresence>
        {mobileOpen && (
          <>
            <motion.div
              initial={{ opacity: 0 }}
              animate={{ opacity: 1 }}
              exit={{ opacity: 0 }}
              className="md:hidden fixed inset-0 bg-black/50 backdrop-blur-[2px] z-40"
              onClick={() => onMobileOpenChange(false)}
            />
            <motion.aside
              initial={{ x: lang === "fa" || lang === "ar" ? 260 : -260 }}
              animate={{ x: 0 }}
              exit={{ x: lang === "fa" || lang === "ar" ? 260 : -260 }}
              transition={{ type: "spring", stiffness: 400, damping: 32 }}
              className={cn(
                "md:hidden fixed top-0 bottom-0 w-[252px] z-50 overflow-hidden shadow-2xl border-border/60",
                lang === "fa" || lang === "ar" ? "end-0 border-s" : "start-0 border-e",
              )}
              style={{ background: "var(--sidebar-bg)" }}
            >
              <button
                type="button"
                onClick={() => onMobileOpenChange(false)}
                className="absolute top-3 end-3 h-7 w-7 rounded-lg flex items-center justify-center text-fg-subtle hover:text-fg hover:bg-surface-2 z-10 transition-colors"
              >
                <X size={14} />
              </button>
              {inner(true)}
            </motion.aside>
          </>
        )}
      </AnimatePresence>

      <aside
        className={cn(
          "hidden md:flex relative flex-col z-20 transition-[width] duration-200 ease-[cubic-bezier(0.16,1,0.3,1)] border-e border-border/40 flex-shrink-0",
          collapsed ? "w-[58px]" : "w-[278px]",
        )}
        style={{ background: "var(--sidebar-bg)" }}
      >
        {inner(false)}
      </aside>
    </>
  );
}
