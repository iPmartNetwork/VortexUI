import { NavLink, Outlet, useNavigate } from "react-router-dom";
import { cn } from "@/lib/utils";
import { useAuth } from "@/auth/auth";
import { Button } from "./ui";

const nav = [
  { to: "/users", label: "Users" },
  { to: "/nodes", label: "Nodes" },
  { to: "/admins", label: "Admins" },
  { to: "/settings", label: "Settings" },
];

export function Layout() {
  const { logout } = useAuth();
  const navigate = useNavigate();

  return (
    <div className="flex min-h-screen">
      <aside className="flex w-56 flex-col border-r bg-card">
        <div className="px-5 py-5 text-lg font-bold tracking-tight">
          Vortex<span className="text-primary">UI</span>
        </div>
        <nav className="flex-1 space-y-1 px-3">
          {nav.map((n) => (
            <NavLink
              key={n.to}
              to={n.to}
              className={({ isActive }) =>
                cn(
                  "block rounded-md px-3 py-2 text-sm font-medium transition",
                  isActive ? "bg-primary/10 text-primary" : "text-muted-foreground hover:bg-muted",
                )
              }
            >
              {n.label}
            </NavLink>
          ))}
        </nav>
        <div className="p-3">
          <Button
            variant="ghost"
            className="w-full text-muted-foreground"
            onClick={() => {
              logout();
              navigate("/login");
            }}
          >
            Sign out
          </Button>
        </div>
      </aside>
      <main className="flex-1 overflow-auto p-8">
        <Outlet />
      </main>
    </div>
  );
}
