import { Navigate, NavLink, Outlet } from "react-router-dom";
import { getPortalToken, clearPortalToken } from "./PortalLogin";
import { LayoutDashboard, CreditCard, LifeBuoy, Gift, LogOut } from "lucide-react";
import { cn } from "@/lib/utils";

function PortalProtected({ children }: { children: React.ReactNode }) {
  const token = getPortalToken();
  return token ? <>{children}</> : <Navigate to="/portal/login" replace />;
}

/**
 * Mobile-first portal layout with bottom navigation bar,
 * safe-area padding, and large touch targets.
 */
export function PortalMobileLayout() {
  return (
    <PortalProtected>
      <div className="flex min-h-screen flex-col bg-bg">
        {/* Top bar (minimal) */}
        <header className="sticky top-0 z-30 flex items-center justify-between border-b border-border/30 bg-bg/80 px-4 py-3 backdrop-blur-lg safe-top">
          <h1 className="text-sm font-bold text-fg" style={{ fontFamily: "'Orbitron', sans-serif" }}>VortexUI</h1>
          <button
            onClick={() => { clearPortalToken(); window.location.href = "/portal/login"; }}
            className="grid h-9 w-9 place-items-center rounded-xl text-fg-muted hover:bg-surface-2/60 hover:text-fg transition"
            aria-label="Logout"
          >
            <LogOut size={18} />
          </button>
        </header>

        {/* Content area — scrollable, padded for bottom nav */}
        <main className="flex-1 overflow-auto px-4 py-4 pb-24">
          <Outlet />
        </main>

        {/* Bottom navigation bar */}
        <nav className="fixed bottom-0 inset-x-0 z-40 border-t border-border/40 bg-bg-elevated/95 backdrop-blur-xl safe-bottom">
          <div className="flex items-center justify-around py-2">
            <BottomNavLink to="/portal/dashboard" icon={<LayoutDashboard size={20} />} label="Home" />
            <BottomNavLink to="/portal/plans" icon={<CreditCard size={20} />} label="Plans" />
            <BottomNavLink to="/portal/tickets" icon={<LifeBuoy size={20} />} label="Support" />
            <BottomNavLink to="/portal/referral" icon={<Gift size={20} />} label="Invite" />
          </div>
        </nav>
      </div>
    </PortalProtected>
  );
}

function BottomNavLink({ to, icon, label }: { to: string; icon: React.ReactNode; label: string }) {
  return (
    <NavLink
      to={to}
      className={({ isActive }) => cn(
        "flex flex-col items-center gap-0.5 px-3 py-1.5 rounded-xl min-w-[56px] transition-all",
        isActive ? "text-primary" : "text-fg-muted",
      )}
    >
      {icon}
      <span className="text-[10px] font-medium">{label}</span>
    </NavLink>
  );
}
