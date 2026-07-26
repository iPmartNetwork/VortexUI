import { Navigate, NavLink, Outlet } from "react-router-dom";
import { useQuery } from "@tanstack/react-query";
import { getPortalToken, clearPortalToken } from "./PortalLogin";
import { LayoutDashboard, CreditCard, LifeBuoy, LogOut, Zap, Gift } from "lucide-react";
import { cn } from "@/lib/utils";
import { useI18n } from "@/i18n/i18n";

function PortalProtected({ children }: { children: React.ReactNode }) {
  const token = getPortalToken();
  return token ? <>{children}</> : <Navigate to="/portal/login" replace />;
}

export function PortalLayout() {
  const { t } = useI18n();
  const slug = new URLSearchParams(window.location.search).get("slug")
    || sessionStorage.getItem("portal_slug")
    || "";

  const { data: branding } = useQuery({
    queryKey: ["portal-branding", slug],
    queryFn: async () => {
      if (!slug) return null;
      const res = await fetch(`/api/portal/branding?slug=${encodeURIComponent(slug)}`);
      if (!res.ok) return null;
      return res.json() as Promise<{ branding: { panel_title?: string; logo_url?: string; accent_color?: string } }>;
    },
    staleTime: 60_000,
  });

  const title = branding?.branding?.panel_title || "VortexUI";
  const logoURL = branding?.branding?.logo_url;
  const accentColor = branding?.branding?.accent_color;

  return (
    <PortalProtected>
      <div className="flex h-screen overflow-hidden bg-bg text-fg" style={accentColor ? { "--primary": accentColor } as React.CSSProperties : undefined}>
        {/* Decorative background blobs + dot pattern */}
        <div className="fixed inset-0 pointer-events-none overflow-hidden z-0">
          <div className="absolute inset-0 bg-dot-pattern animate-pattern-drift opacity-15" />
          <div className="absolute top-[-20%] start-[5%] w-[500px] h-[500px] rounded-full bg-primary/[0.04] blur-[120px]" />
          <div className="absolute bottom-[-15%] end-[10%] w-[400px] h-[400px] rounded-full bg-accent/[0.03] blur-[100px]" />
        </div>

        {/* ── Sidebar ── */}
        <aside className="relative z-10 hidden md:flex w-[236px] flex-col border-e border-border/40 flex-shrink-0 h-screen bg-bg-elevated/80 backdrop-blur-xl">
          {/* Logo */}
          <div className="flex items-center gap-3 px-4 h-14 flex-shrink-0 border-b border-border/30">
            {logoURL ? (
              <img src={logoURL} alt="" className="h-8 w-8 rounded-[10px] object-contain" />
            ) : (
              <div className="h-8 w-8 rounded-[10px] bg-gradient-to-br from-primary to-accent flex items-center justify-center shadow-lg shadow-primary/30">
                <Zap size={15} className="text-white" />
              </div>
            )}
            <div>
              <p className="text-sm font-bold leading-none text-fg">{title}</p>
              <p className="text-[9px] text-fg-subtle mt-0.5 uppercase tracking-wider font-semibold">
                {t("portal.brand")}
              </p>
            </div>
          </div>

          {/* Nav */}
          <nav className="flex-1 px-2.5 py-3 space-y-0.5">
            <p className="px-2.5 pb-2 text-[9px] font-bold uppercase tracking-widest text-fg-subtle/60">
              {t("portal.nav.dashboard")}
            </p>
            <PortalNavLink to="/portal/dashboard" icon={<LayoutDashboard size={18} />}>
              {t("portal.nav.dashboard")}
            </PortalNavLink>
            <PortalNavLink to="/portal/plans" icon={<CreditCard size={18} />}>
              {t("portal.nav.plans")}
            </PortalNavLink>
            <PortalNavLink to="/portal/tickets" icon={<LifeBuoy size={18} />}>
              {t("portal.nav.support")}
            </PortalNavLink>
            <PortalNavLink to="/portal/referral" icon={<Gift size={18} />}>
              {t("portal.nav.referral")}
            </PortalNavLink>
          </nav>

          {/* Logout */}
          <div className="p-2.5 border-t border-border/30">
            <button
              type="button"
              onClick={() => {
                clearPortalToken();
                window.location.href = "/portal/login";
              }}
              className="w-full flex items-center gap-2.5 h-9 px-2.5 rounded-[10px] text-[13px] text-fg-muted hover:text-danger hover:bg-danger/10 transition-all duration-200 group"
            >
              <LogOut size={16} className="transition-transform duration-200 group-hover:-translate-x-0.5" />
              <span>{t("portal.logout")}</span>
            </button>
          </div>
        </aside>

        {/* ── Main content ── */}
        <div className="relative z-10 flex-1 flex flex-col min-w-0 h-screen overflow-hidden">
          <main className="flex-1 overflow-y-auto overscroll-contain">
            <div className="w-full px-4 py-5 md:px-6 md:py-6 lg:px-8 animate-page-enter">
              <Outlet />
            </div>
          </main>
        </div>
      </div>
    </PortalProtected>
  );
}

function PortalNavLink({
  to,
  icon,
  children,
}: {
  to: string;
  icon: React.ReactNode;
  children: React.ReactNode;
}) {
  return (
    <NavLink
      to={to}
      className={({ isActive }) =>
        cn(
          "relative flex h-9 items-center gap-2.5 px-2.5 rounded-[10px] text-[13px] font-medium transition-all duration-150 group",
          isActive
            ? "bg-primary/[0.12] text-fg font-semibold shadow-sm"
            : "text-fg-muted hover:text-fg hover:bg-surface-2/80",
        )
      }
    >
      {({ isActive }) => (
        <>
          {isActive && (
            <span className="absolute start-0 inset-y-1 w-[3px] rounded-full bg-gradient-to-b from-primary to-accent shadow-sm shadow-primary/40" />
          )}
          <span
            className={cn(
              "flex-shrink-0 transition-colors duration-150",
              isActive ? "text-primary" : "group-hover:text-primary/70",
            )}
          >
            {icon}
          </span>
          {children}
        </>
      )}
    </NavLink>
  );
}
