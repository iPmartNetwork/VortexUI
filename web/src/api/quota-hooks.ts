import { useQuery } from "@tanstack/react-query";
import { ensureArray } from "@/lib/utils";
import { api, getToken } from "./client";

export interface AdminQuotaUsage {
  admin_id: string;
  username: string;
  user_quota: number;
  user_count: number;
  users_remaining: number | null;
  traffic_quota: number;
  traffic_quota_mode?: string;
  traffic_used: number;
  traffic_allocated: number;
  traffic_remaining: number | null;
  wallet_traffic_bytes?: number;
  wallet_user_credits?: number;
}

export function useAccountQuota() {
  return useQuery({
    queryKey: ["account-quota"],
    queryFn: () => api<{ usage: AdminQuotaUsage }>("/api/account/quota"),
  });
}

export function useResellerQuotaUsage() {
  return useQuery({
    queryKey: ["reseller-quota-usage"],
    queryFn: async () => {
      const res = await api<{ usage?: AdminQuotaUsage[] | null }>("/api/admins/usage");
      return { usage: ensureArray(res.usage) };
    },
  });
}

export function useAdminQuotaUsage(adminId: string | null) {
  return useQuery({
    queryKey: ["admin-quota", adminId],
    queryFn: () => api<{ usage: AdminQuotaUsage }>(`/api/admins/${adminId}/quota`),
    enabled: !!adminId,
  });
}

export interface ResellerDashboardData {
  quota: AdminQuotaUsage;
  users_by_status: Record<string, number>;
  top_users: { id: string; username: string; used_traffic: number; data_limit: number; status: string }[];
  expiring_soon: number;
  new_users_7d: number;
  new_users_30d: number;
}

export function useResellerDashboard() {
  return useQuery({
    queryKey: ["reseller-dashboard"],
    queryFn: () => api<{ dashboard: ResellerDashboardData }>("/api/account/dashboard"),
  });
}

export function exportResellerUsersCsv() {
  const token = getToken();
  if (!token) throw new Error("not authenticated");
  window.open(`/api/account/export/users?access_token=${encodeURIComponent(token)}`, "_blank");
}
