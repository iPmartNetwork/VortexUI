import { Navigate, NavLink, Outlet } from "react-router-dom";
import { getPortalToken, clearPortalToken } from "./PortalLogin";
import { LayoutDashboard, CreditCard, LifeBuoy, LogOut } from "lucide-react";
import { cn } from "@/lib/utils";
import { useI18n } from "@/i18n/i18n";

function PortalProtected({ children }: { children: React.ReactNode }) {
  const token = getPortalToken();
  return token ? <>{children}</> : <Navigate to="/portal/login" replace />;
}

/** Mobile-first portal layout with bottom navigation. */
export function PortalMobileLayout() {
  const { t } = useI18n();

  return (
    <PortalProtected>
      <div className="flex min-h-screen flex-col bg-bg text-fg">
        <header className="sticky top-0 z-30 flex items-center justify-between border-b border-border/60 bg-bg/90 px-4 py-3 backdrop-blur-xl safe-top">
          <h1 className="text-sm font-bold">
            Vortex<span className="text-primary">UI</span>
            <span className="text-fg-subtle font-normal ms-1">{t("shell.portal")}</span>
          </h1>
          <button
            type="button"
            onClick={() => {
              clearPortalToken();
              window.location.href = "/portal/login";
            }}
            className="grid h-9 w-9 place-items-center rounded-xl text-fg-muted hover:bg-surface-2 hover:text-fg transition"
            aria-label={t("portal.logout")}
          >
            <LogOut size={18} />
          </button>
        </header>

        <main className="flex-1 overflow-auto px-4 py-4 pb-24 animate-page-enter">
          <Outlet />
        </main>

        <nav className="fixed bottom-0 inset-x-0 z-40 border-t border-border/60 bg-bg-elevated/95 backdrop-blur-xl safe-bottom">
          <div className="flex items-center justify-around py-2">
            <BottomNavLink to="/portal/dashboard" icon={<LayoutDashboard size={20} />} label={t("portal.nav.home")} />
            <BottomNavLink to="/portal/plans" icon={<CreditCard size={20} />} label={t("portal.nav.plans")} />
            <BottomNavLink to="/portal/tickets" icon={<LifeBuoy size={20} />} label={t("portal.nav.support")} />
          </div>
        </nav>
      </div>
    </PortalProtected>
  );
}

function BottomNavLink({
  to,
  icon,
  label,
}: {
  to: string;
  icon: React.ReactNode;
  label: string;
}) {
  return (
    <NavLink
      to={to}
      className={({ isActive }) =>
        cn(
          "flex flex-col items-center gap-0.5 px-3 py-1.5 rounded-xl min-w-[56px] transition-all",
          isActive ? "text-primary" : "text-fg-muted",
        )
      }
    >
      {icon}
      <span className="text-[10px] font-medium">{label}</span>
    </NavLink>
  );
}
