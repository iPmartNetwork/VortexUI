import { useState } from "react";
import { NavLink, useLocation, useNavigate } from "react-router-dom";
import { motion, AnimatePresence } from "framer-motion";
import {
  Menu,
  X,
  LogOut,
  Smartphone,
  ChevronsLeft,
  ChevronsRight,
  ExternalLink,
  Terminal,
} from "lucide-react";
import { cn } from "@/lib/utils";
import { useAuth } from "@/auth/auth";
import { canAccessRoute } from "@/auth/permissions";
import { useI18n } from "@/i18n/i18n";
import { useVersion, useNodes } from "@/api/hooks";
import { buildCompactNavSections } from "@/navigation/nav-sections-compact";
import { useOverview } from "@/api/policy-hooks";
import type { Overview } from "@/api/types";

const PANEL_VERSION = "1.3.1";

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
  const { t, lang } = useI18n();
  const version = useVersion().data ?? PANEL_VERSION;
  const nodesQ = useNodes();
  const overviewQ = useOverview();
  const navBadges = overviewQ.data?.widgets?.nav_badges;
  const primaryCore = nodesQ.data?.nodes?.find((n) => n.core === "xray") ?? nodesQ.data?.nodes?.[0];
  const coreOnline = primaryCore?.health?.core_running ?? false;
  const coreVer = primaryCore?.core_version || "—";
  const coreLabel = primaryCore?.core === "singbox" ? "sing-box" : "Xray";
  const [collapsed, setCollapsed] = useState(false);
  const [hovered, setHovered] = useState<string | null>(null);
  const rtl = lang === "fa" || lang === "ar";

  const visibleSections = buildCompactNavSections(sudo)
    .map((section) => ({
      ...section,
      items: section.items.filter((item) =>
        canAccessRoute(item.to, sudo, permissions, session?.admin.reseller_settings),
      ),
    }))
    .filter((section) => section.items.length > 0);

  function signOut() {
    logout();
    navigate("/login");
  }

  function inner(mobile: boolean) {
    const mini = collapsed && !mobile;

    return (
      <div className="flex flex-col h-full min-h-0 select-none">
        {/* Brand — text only (no logo icon) */}
        <div className={cn("flex-shrink-0 px-4 pt-4 pb-3", mini && "px-2 pt-3 pb-2 flex justify-center")}>
          {mini ? (
            <span className="text-[11px] font-black tracking-tight text-fg">VU</span>
          ) : (
            <>
              <div className="flex flex-wrap items-center gap-2">
                <span className="text-sm font-black tracking-[0.12em] text-fg">VORTEXUI</span>
                <span className="rounded-md bg-primary/10 px-1.5 py-0.5 text-[10px] font-bold text-primary tabular-nums">
                  {version}
                </span>
              </div>
              <p className="text-[10px] text-fg-subtle mt-1 leading-none">{t("app.taglineVeltrix")}</p>
            </>
          )}
        </div>

        {/* Nav — scrolls inside fixed sidebar */}
        <nav className={cn("flex-1 min-h-0 overflow-y-auto overscroll-contain", mini ? "px-1.5 py-1" : "px-2.5 py-1")}>
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
                    const itemPath = item.to.split("?")[0];
                    const active = location.pathname.startsWith(itemPath);
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
                            <>
                              <span className="flex-1 text-start text-[13px] truncate">{label}</span>
                              {item.hotDot && badge === 0 && (
                                <span className="h-1.5 w-1.5 rounded-full bg-success flex-shrink-0 shadow-[0_0_6px] shadow-success/50" />
                              )}
                              {badge > 0 && (
                                <span
                                  className={cn(
                                    "text-[10px] font-semibold leading-none min-w-[18px] h-[18px]",
                                    "flex items-center justify-center rounded-[6px] tabular-nums px-1",
                                    active
                                      ? "bg-primary/25 text-primary"
                                      : "bg-surface-3/70 text-fg-subtle",
                                  )}
                                >
                                  {badge > 99 ? "99+" : badge}
                                </span>
                              )}
                            </>
                          )}
                        </NavLink>

                        {mini && (
                          <AnimatePresence>
                            {hovered === item.to && (
                              <motion.div
                                initial={{ opacity: 0, x: rtl ? 4 : -4 }}
                                animate={{ opacity: 1, x: 0 }}
                                exit={{ opacity: 0, x: rtl ? 4 : -4 }}
                                transition={{ duration: 0.1 }}
                                className="absolute start-full ms-2.5 top-1/2 -translate-y-1/2 z-[60] pointer-events-none"
                              >
                                <div className="flex items-center gap-2 rounded-lg bg-fg text-bg px-2.5 py-1.5 shadow-xl text-[11px] font-medium whitespace-nowrap">
                                  {label}
                                  {badge > 0 && (
                                    <span className="bg-primary/20 text-primary rounded px-1 text-[9px] font-bold">
                                      {badge > 99 ? "99+" : badge}
                                    </span>
                                  )}
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

        {/* Footer — pinned at bottom */}
        <div className={cn("flex-shrink-0 border-t border-border/40", mini ? "p-2 space-y-2" : "p-3 space-y-2")}>
          {!mini && (
            <button
              type="button"
              onClick={() => navigate("/portal/login")}
              className="w-full h-9 rounded-xl text-[12px] font-semibold text-accent flex items-center justify-center gap-1.5 border border-accent/20 bg-accent/[0.06] hover:bg-accent/10 transition-all"
            >
              <Smartphone size={14} />
              {t("shell.selfServicePortal")}
              <ExternalLink size={10} className="opacity-50" />
            </button>
          )}

          {primaryCore && !mini && (
            <div className="flex items-center gap-2 rounded-xl border border-border/70 bg-surface-2/50 px-3 py-2.5">
              <div className="h-8 w-8 rounded-lg bg-surface-3/80 flex items-center justify-center flex-shrink-0 text-fg-subtle">
                <Terminal size={14} />
              </div>
              <div className="flex-1 min-w-0">
                <p className="text-xs font-bold text-fg truncate">
                  {coreLabel} v{coreVer}
                </p>
                <p className="text-[10px] text-fg-subtle flex items-center gap-1.5 mt-0.5">
                  <span className={cn("h-1.5 w-1.5 rounded-full", coreOnline ? "bg-success" : "bg-fg-subtle")} />
                  {coreOnline ? t("overview.coreOnline") : t("overview.coreOffline")}
                </p>
              </div>
              <button
                type="button"
                onClick={signOut}
                className="h-8 w-8 rounded-lg flex items-center justify-center text-fg-subtle hover:text-danger hover:bg-danger/10 transition-colors flex-shrink-0"
                title={t("nav.signout")}
              >
                <LogOut size={15} />
              </button>
            </div>
          )}

          {mini && (
            <button
              type="button"
              onClick={signOut}
              className="h-8 w-8 mx-auto rounded-lg flex items-center justify-center text-fg-subtle hover:text-danger hover:bg-danger/10 transition-colors"
              title={t("nav.signout")}
            >
              <LogOut size={14} />
            </button>
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
              initial={{ x: rtl ? 260 : -260 }}
              animate={{ x: 0 }}
              exit={{ x: rtl ? 260 : -260 }}
              transition={{ type: "spring", stiffness: 400, damping: 32 }}
              className={cn(
                "md:hidden fixed top-0 bottom-0 w-[252px] z-50 overflow-hidden shadow-2xl border-border/60",
                rtl ? "end-0 border-s" : "start-0 border-e",
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
          "hidden md:flex relative flex-col z-20 h-screen sticky top-0 flex-shrink-0",
          "transition-[width] duration-200 ease-[cubic-bezier(0.16,1,0.3,1)] border-e border-border/40",
          collapsed ? "w-[58px]" : "w-[236px]",
        )}
        style={{ background: "var(--sidebar-bg)" }}
      >
        {inner(false)}
      </aside>
    </>
  );
}
