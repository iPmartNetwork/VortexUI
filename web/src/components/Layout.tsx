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
  Settings as SettingsIcon,
  LogOut,
  Moon,
  Sun,
  Languages,
} from "lucide-react";
import { cn } from "@/lib/utils";
import { useAuth } from "@/auth/auth";
import { useTheme } from "@/theme/theme";
import { useI18n } from "@/i18n/i18n";
import type { TKey } from "@/i18n/dict";

const nav: { to: string; key: TKey; icon: React.ElementType }[] = [
  { to: "/overview", key: "nav.overview", icon: LayoutDashboard },
  { to: "/users", key: "nav.users", icon: UsersIcon },
  { to: "/nodes", key: "nav.nodes", icon: Server },
  { to: "/outbounds", key: "nav.outbounds", icon: Network },
  { to: "/routing", key: "nav.routing", icon: RouteIcon },
  { to: "/balancers", key: "nav.balancers", icon: Scale },
  { to: "/admins", key: "nav.admins", icon: ShieldCheck },
  { to: "/logs", key: "nav.logs", icon: ScrollText },
  { to: "/settings", key: "nav.settings", icon: SettingsIcon },
];

function IconButton({ onClick, label, children }: { onClick: () => void; label: string; children: React.ReactNode }) {
  return (
    <button
      onClick={onClick}
      aria-label={label}
      title={label}
      className="grid h-9 w-9 place-items-center rounded-lg text-fg-muted transition hover:bg-surface-2 hover:text-fg"
    >
      {children}
    </button>
  );
}

export function Layout() {
  const { logout } = useAuth();
  const navigate = useNavigate();
  const { resolved, toggle } = useTheme();
  const { t, lang, setLang } = useI18n();

  return (
    <div className="flex min-h-screen">
      <aside className="flex w-60 flex-col border-e bg-bg-elevated/60 backdrop-blur">
        <div className="flex items-center gap-2.5 px-5 py-5">
          <div className="grid h-9 w-9 place-items-center rounded-xl bg-gradient-to-br from-primary to-accent text-white shadow-lg shadow-primary/30">
            <Network size={18} strokeWidth={2.5} />
          </div>
          <div>
            <div className="text-[15px] font-bold leading-none tracking-tight">VortexUI</div>
            <div className="mt-1 text-[11px] text-fg-subtle">{t("app.tagline")}</div>
          </div>
        </div>

        <nav className="flex-1 space-y-0.5 overflow-y-auto px-3 py-2">
          {nav.map((n) => {
            const Icon = n.icon;
            return (
              <NavLink
                key={n.to}
                to={n.to}
                className={({ isActive }) =>
                  cn(
                    "group relative flex items-center gap-3 rounded-lg px-3 py-2 text-sm font-medium transition",
                    isActive
                      ? "bg-primary/10 text-primary"
                      : "text-fg-muted hover:bg-surface-2 hover:text-fg",
                  )
                }
              >
                {({ isActive }) => (
                  <>
                    {isActive && (
                      <span className="absolute inset-y-1.5 start-0 w-1 rounded-full bg-primary" />
                    )}
                    <Icon size={18} strokeWidth={2} />
                    {t(n.key)}
                  </>
                )}
              </NavLink>
            );
          })}
        </nav>

        <div className="flex items-center justify-between border-t px-4 py-3">
          <div className="flex gap-1">
            <IconButton onClick={toggle} label={t("theme.label")}>
              {resolved === "dark" ? <Sun size={18} /> : <Moon size={18} />}
            </IconButton>
            <IconButton onClick={() => setLang(lang === "en" ? "fa" : "en")} label="Language">
              <span className="flex items-center gap-1 text-xs font-semibold">
                <Languages size={16} />
                {lang.toUpperCase()}
              </span>
            </IconButton>
          </div>
          <IconButton
            onClick={() => {
              logout();
              navigate("/login");
            }}
            label={t("nav.signout")}
          >
            <LogOut size={18} />
          </IconButton>
        </div>
      </aside>

      <main className="flex-1 overflow-auto">
        <div className="mx-auto max-w-6xl px-8 py-8 animate-fade-in">
          <Outlet />
        </div>
      </main>
    </div>
  );
}
