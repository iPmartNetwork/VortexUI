import { Navigate, NavLink, Outlet } from "react-router-dom";
import { getPortalToken, clearPortalToken } from "./PortalLogin";
import { LayoutDashboard, CreditCard, LifeBuoy, LogOut, Gift, Zap } from "lucide-react";
import { cn } from "@/lib/utils";
import { useI18n } from "@/i18n/i18n";

function PortalProtected({ children }: { children: React.ReactNode }) {
  const token = getPortalToken();
  return token ? <>{children}</> : <Navigate to="/portal/login" replace />;
}

/** Mobile-first portal layout with enhanced bottom navigation. */
export function PortalMobileLayout() {
  const { t } = useI18n();

  return (
    <PortalProtected>
      <div className="flex min-h-screen flex-col bg-bg text-fg">
        {/* ── Header ── */}
        <header className="sticky top-0 z-30 flex items-center justify-between border-b border-border/50 bg-bg/80 px-4 py-3 backdrop-blur-xl safe-top">
          <h1 className="text-sm font-bold flex items-center gap-1.5">
            <div className="h-6 w-6 rounded-lg bg-gradient-to-br from-primary to-accent flex items-center justify-center">
              <Zap size={12} className="text-white" />
            </div>
            Vortex<span className="text-primary">UI</span>
            <span className="text-fg-subtle font-normal ms-1 text-[10px]">{t("shell.portal")}</span>
          </h1>
          <button
            type="button"
            onClick={() => {
              clearPortalToken();
              window.location.href = "/portal/login";
            }}
            className="grid h-9 w-9 place-items-center rounded-xl text-fg-muted hover:bg-danger/10 hover:text-danger transition-all duration-200"
            aria-label={t("portal.logout")}
          >
            <LogOut size={18} />
          </button>
        </header>

        {/* ── Main Content ── */}
        <main className="flex-1 overflow-auto px-4 py-4 pb-28 animate-page-enter">
          <Outlet />
        </main>

        {/* ── Bottom Navigation ── */}
        <nav className="fixed bottom-0 inset-x-0 z-40 border-t border-border/50 bg-bg-elevated/90 backdrop-blur-xl safe-bottom shadow-lg shadow-bg/50">
          <div className="flex items-center justify-around py-1.5">
            <BottomNavLink to="/portal/dashboard" icon={<LayoutDashboard size={20} />} label={t("portal.nav.home")} />
            <BottomNavLink to="/portal/plans" icon={<CreditCard size={20} />} label={t("portal.nav.plans")} />
            <BottomNavLink to="/portal/referral" icon={<Gift size={20} />} label="Referral" />
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
          "relative flex flex-col items-center gap-0.5 px-4 py-1.5 rounded-xl min-w-[60px] transition-all duration-150 group",
          isActive ? "text-primary" : "text-fg-muted hover:text-fg",
        )
      }
    >
      {({ isActive }) => (
        <>
          {isActive && (
            <span className="absolute -top-1.5 inset-x-4 h-[2px] rounded-full bg-gradient-to-r from-primary to-accent" />
          )}
          <span
            className={cn(
              "transition-transform duration-150",
              isActive && "scale-110",
            )}
          >
            {icon}
          </span>
          <span className="text-[9px] font-semibold">{label}</span>
        </>
      )}
    </NavLink>
  );
}
