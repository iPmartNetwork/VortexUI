import { Navigate } from "react-router-dom";
import { AdminsPanel } from "@/pages/admins/AdminsPanel";

/** Embedded in Settings → Admins tab (sudo only). */
export function AdminsTab({ embedded = false }: { embedded?: boolean }) {
  return <AdminsPanel embedded={embedded} />;
}

/** @deprecated — use /settings?tab=admins */
export function Admins() {
  return <Navigate to="/settings?tab=admins" replace />;
}

export { ResellerDetail } from "@/pages/admins/ResellerDetail";
