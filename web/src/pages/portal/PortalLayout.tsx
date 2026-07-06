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

  return (
    <PortalProtected>
      <div className="flex min-h-screen bg-bg text-fg overflow-hidden">
        <aside
          className="hidden md:flex w-[236px] flex-col border-e border-border/40 flex-shrink-0"
          style={{ background: "var(--sidebar-bg)" }}
        >
          <div className="flex items-center gap-3 px-4 h-14 flex-shrink-0">
            {logoURL ? (
              <img src={logoURL} alt="" className="h-8 w-8 rounded-[10px] object-contain" />
            ) : (
              <div className="h-8 w-8 rounded-[10px] grad-bg flex items-center justify-center">
                <Zap size={15} className="text-primary-fg" />
              </div>
            )}
            <div>
              <p className="text-sm font-semibold leading-none">{title}</p>
              <p className="text-[10px] text-fg-subtle mt-0.5">{t("portal.brand")}</p>
            </div>
          </div>
          <nav className="flex-1 px-2.5 py-2 space-y-px">
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
          <div className="p-2.5 border-t border-border/40">
            <button
              type="button"
              onClick={() => {
                clearPortalToken();
                window.location.href = "/portal/login";
              }}
              className="w-full flex items-center gap-2.5 h-9 px-2.5 rounded-[10px] text-[13px] text-fg-muted hover:text-fg hover:bg-surface-2 transition-colors"
            >
              <LogOut size={16} /> {t("portal.logout")}
            </button>
          </div>
        </aside>
        <main className="flex-1 overflow-y-auto overscroll-contain">
          <div className="w-full px-4 py-5 md:px-6 md:py-6 lg:px-8 animate-page-enter">
            <Outlet />
          </div>
        </main>
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
          "relative flex h-9 items-center gap-2.5 px-2.5 rounded-[10px] text-[13px] font-medium transition-all duration-[120ms]",
          isActive
            ? "bg-primary/[0.12] text-fg font-semibold"
            : "text-fg-muted hover:text-fg hover:bg-surface-2",
        )
      }
    >
      {({ isActive }) => (
        <>
          {isActive && (
            <span className="absolute start-0 inset-y-1 w-[2.5px] rounded-full bg-primary" />
          )}
          <span className={cn("flex-shrink-0", isActive && "text-primary")}>{icon}</span>
          {children}
        </>
      )}
    </NavLink>
  );
}
