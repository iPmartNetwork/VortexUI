import { useQuery } from "@tanstack/react-query";
import { api } from "./client";

export interface AdminQuotaUsage {
  admin_id: string;
  username: string;
  user_quota: number;
  user_count: number;
  users_remaining: number | null;
  traffic_quota: number;
  traffic_used: number;
  traffic_allocated: number;
  traffic_remaining: number | null;
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
    queryFn: () => api<{ usage: AdminQuotaUsage[] }>("/api/admins/usage"),
  });
}
