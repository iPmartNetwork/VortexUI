import { useState } from "react";
import { Outlet, useNavigate } from "react-router-dom";
import { AppSidebar } from "@/components/AppSidebar";
import { AppHeader } from "@/components/AppHeader";
import { CommandPalette } from "@/components/CommandPalette";
import { useAuth } from "@/auth/auth";
import { useLiveEvents } from "@/api/live";
import { useStopImpersonation } from "@/api/reseller-hooks";
import { setToken } from "@/api/client";

export function Layout() {
  const { logout, impersonating, session, refreshSession } = useAuth();
  const stopImpersonate = useStopImpersonation();
  const navigate = useNavigate();
  const [mobileOpen, setMobileOpen] = useState(false);
  useLiveEvents();

  async function exitImpersonation() {
    try {
      const res = await stopImpersonate.mutateAsync();
      setToken(res.token);
      await refreshSession();
    } catch {
      logout();
      navigate("/login");
    }
  }

  return (
    <div className="flex h-screen overflow-hidden bg-bg text-fg">
      <CommandPalette />
      <AppSidebar mobileOpen={mobileOpen} onMobileOpenChange={setMobileOpen} />

      <div className="flex-1 flex flex-col min-w-0 h-screen overflow-hidden">
        <AppHeader />
        <main className="flex-1 overflow-y-auto overscroll-contain">
          <div className="w-full px-4 py-5 md:px-6 md:py-6 lg:px-8">
            {impersonating && (
              <div className="mb-4 flex flex-wrap items-center justify-between gap-3 rounded-xl border border-amber-500/40 bg-amber-500/10 px-4 py-3 text-sm">
                <span>
                  Impersonating <strong>@{session?.admin.username}</strong> (support session)
                </span>
                <button
                  type="button"
                  onClick={() => void exitImpersonation()}
                  className="rounded-lg bg-amber-600 px-3 py-1.5 text-xs font-medium text-white hover:bg-amber-500"
                >
                  Exit impersonation
                </button>
              </div>
            )}
            <Outlet />
          </div>
        </main>
      </div>
    </div>
  );
}
