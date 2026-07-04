import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { api, getToken } from "./client";

export interface WalletLedgerEntry {
  id: string;
  admin_id: string;
  delta_traffic: number;
  delta_users: number;
  reason: string;
  actor_admin_id?: string;
  created_at: string;
}

export interface AdminWallet {
  admin_id: string;
  traffic_bytes: number;
  user_credits: number;
}

export interface PortalBranding {
  admin_id: string;
  panel_title: string;
  logo_url: string;
  accent_color: string;
  footer_text: string;
  portal_slug?: string;
  custom_domain?: string;
}

export function useAccountWallet() {
  return useQuery({
    queryKey: ["account-wallet"],
    queryFn: () => api<{ wallet: AdminWallet; ledger: WalletLedgerEntry[] }>("/api/account/wallet"),
  });
}

export function useSubAdmins() {
  return useQuery({
    queryKey: ["sub-admins"],
    queryFn: () => api<{ admins: import("./types").Admin[] }>("/api/account/sub-admins"),
  });
}

export function useAccountBranding() {
  return useQuery({
    queryKey: ["account-branding"],
    queryFn: () => api<{ branding: PortalBranding }>("/api/account/branding"),
  });
}

export function useAccountWebhook() {
  return useQuery({
    queryKey: ["account-webhook"],
    queryFn: () => api<{ url: string; enabled: boolean; has_secret: boolean }>("/api/account/webhook"),
  });
}

export function useSaveBranding() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (branding: PortalBranding) =>
      api("/api/account/branding", { method: "PUT", body: branding }),
    onSuccess: () => qc.invalidateQueries({ queryKey: ["account-branding"] }),
  });
}

export function useSaveWebhook() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (body: { url: string; secret: string; enabled: boolean }) =>
      api("/api/account/webhook", { method: "PUT", body }),
    onSuccess: () => qc.invalidateQueries({ queryKey: ["account-webhook"] }),
  });
}

export function useCreateSubAdmin() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (body: {
      username: string;
      password: string;
      role_id: string;
      user_quota: number;
      traffic_quota: number;
      traffic_quota_mode?: string;
    }) => api("/api/account/sub-admins", { method: "POST", body }),
    onSuccess: () => qc.invalidateQueries({ queryKey: ["sub-admins"] }),
  });
}

export function useTopUpAdminWallet() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (args: { adminId: string; traffic_bytes: number; user_credits: number; reason?: string }) =>
      api(`/api/admins/${args.adminId}/wallet`, {
        method: "POST",
        body: { traffic_bytes: args.traffic_bytes, user_credits: args.user_credits, reason: args.reason },
      }),
    onSuccess: (_data, vars) => {
      qc.invalidateQueries({ queryKey: ["account-wallet"] });
      qc.invalidateQueries({ queryKey: ["reseller-quota-usage"] });
      qc.invalidateQueries({ queryKey: ["admin-quota", vars.adminId] });
      qc.invalidateQueries({ queryKey: ["admin-wallet", vars.adminId] });
    },
  });
}

export function useAdminWallet(adminId: string | null) {
  return useQuery({
    queryKey: ["admin-wallet", adminId],
    queryFn: () => api<{ wallet: AdminWallet; ledger: WalletLedgerEntry[] }>(`/api/admins/${adminId}/wallet`),
    enabled: !!adminId,
  });
}

export function useImpersonateAdmin() {
  return useMutation({
    mutationFn: (adminId: string) =>
      api<{ token: string; impersonating: string }>(`/api/admins/${adminId}/impersonate`, { method: "POST" }),
  });
}

export function useStopImpersonation() {
  return useMutation({
    mutationFn: () => api<{ token: string }>("/api/account/stop-impersonate", { method: "POST" }),
  });
}

export function exportAccountWalletStatement() {
  const token = getToken();
  if (!token) return;
  window.open(`/api/account/wallet/export?access_token=${encodeURIComponent(token)}`, "_blank");
}
