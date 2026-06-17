import { Navigate, NavLink, Outlet } from "react-router-dom";
import { getPortalToken, clearPortalToken } from "./PortalLogin";
import { LayoutDashboard, CreditCard, LifeBuoy, LogOut } from "lucide-react";

function PortalProtected({ children }: { children: React.ReactNode }) {
  const token = getPortalToken();
  return token ? <>{children}</> : <Navigate to="/portal/login" replace />;
}

export function PortalLayout() {
  return (
    <PortalProtected>
      <div className="flex min-h-screen bg-bg">
        <aside className="hidden md:flex w-56 flex-col border-r border-border/40 bg-surface/50 p-4">
          <div className="mb-6">
            <h2 className="text-sm font-bold text-fg">VortexUI Portal</h2>
          </div>
          <nav className="flex-1 space-y-1">
            <PortalNavLink to="/portal/dashboard" icon={<LayoutDashboard size={16} />}>Dashboard</PortalNavLink>
            <PortalNavLink to="/portal/plans" icon={<CreditCard size={16} />}>Plans</PortalNavLink>
            <PortalNavLink to="/portal/tickets" icon={<LifeBuoy size={16} />}>Support</PortalNavLink>
          </nav>
          <button
            onClick={() => { clearPortalToken(); window.location.href = "/portal/login"; }}
            className="flex items-center gap-2 rounded-lg px-3 py-2 text-xs text-fg-muted hover:bg-surface-2 hover:text-fg transition"
          >
            <LogOut size={14} /> Logout
          </button>
        </aside>
        <main className="flex-1 p-6 lg:p-8">
          <Outlet />
        </main>
      </div>
    </PortalProtected>
  );
}

function PortalNavLink({ to, icon, children }: { to: string; icon: React.ReactNode; children: React.ReactNode }) {
  return (
    <NavLink
      to={to}
      className={({ isActive }) =>
        `flex items-center gap-2 rounded-lg px-3 py-2 text-sm transition ${isActive ? "bg-primary/10 text-primary font-medium" : "text-fg-muted hover:bg-surface-2 hover:text-fg"}`
      }
    >
      {icon}
      {children}
    </NavLink>
  );
}
