import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { api } from "./client";
import type { Admin, Role } from "./types";

// --- admins ---

export function useAdmins() {
  return useQuery({ queryKey: ["admins"], queryFn: () => api<{ admins: Admin[] }>("/api/admins") });
}

export function useCreateAdmin() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (input: { username: string; password: string; sudo: boolean; role_id?: string | null; user_quota?: number; traffic_quota?: number; traffic_quota_mode?: string; inbound_ids?: string[]; node_ids?: string[]; plan_ids?: string[] }) =>
      api<{ admin: Admin }>("/api/admins", { method: "POST", body: input }),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["admins"] });
      qc.invalidateQueries({ queryKey: ["inbounds-all"] });
    },
  });
}

export function useDeleteAdmin() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => api<void>(`/api/admins/${id}`, { method: "DELETE" }),
    onSuccess: () => qc.invalidateQueries({ queryKey: ["admins"] }),
  });
}

export function useUnsuspendAdmin() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => api<{ admin: Admin }>(`/api/admins/${id}/unsuspend`, { method: "POST" }),
    onSuccess: () => qc.invalidateQueries({ queryKey: ["admins"] }),
  });
}

export function useAdjustAdminQuota() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (args: { id: string; user_quota_delta?: number; traffic_quota_delta?: number }) =>
      api<{ admin: Admin }>(`/api/admins/${args.id}/quota-adjust`, {
        method: "POST",
        body: { user_quota_delta: args.user_quota_delta ?? 0, traffic_quota_delta: args.traffic_quota_delta ?? 0 },
      }),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["admins"] });
      qc.invalidateQueries({ queryKey: ["reseller-quota-usage"] });
    },
  });
}

export function useUpdateAdmin() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (args: { id: string; input: {
      password?: string; sudo: boolean; role_id?: string | null; user_quota?: number; traffic_quota?: number;
      traffic_quota_mode?: string; inbound_ids?: string[]; node_ids?: string[]; plan_ids?: string[];
      policy_max_data_limit?: number; policy_max_expire_days?: number;
      policy_allow_bulk_delete?: boolean; policy_allow_bulk_create?: boolean;
      auto_suspend_enabled?: boolean; ip_violation_suspend_threshold?: number; suspend_grace_minutes?: number;
      allow_sub_resellers?: boolean; allow_user_backup?: boolean; reseller_settings?: Record<string, boolean>;
    } }) =>
      api<{ admin: Admin }>(`/api/admins/${args.id}`, { method: "PUT", body: args.input }),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["admins"] });
      qc.invalidateQueries({ queryKey: ["account"] });
      qc.invalidateQueries({ queryKey: ["inbounds-all"] });
    },
  });
}

export function useAdminInbounds(adminId: string | null) {
  return useQuery({
    queryKey: ["admin-inbounds", adminId],
    queryFn: () => api<{ inbound_ids: string[] }>(`/api/admins/${adminId}/inbounds`),
    enabled: !!adminId,
  });
}

export function useAdminNodes(adminId: string | null) {
  return useQuery({
    queryKey: ["admin-nodes", adminId],
    queryFn: () => api<{ node_ids: string[] }>(`/api/admins/${adminId}/nodes`),
    enabled: !!adminId,
  });
}

export function useAdminPlans(adminId: string | null) {
  return useQuery({
    queryKey: ["admin-plans", adminId],
    queryFn: () => api<{ plan_ids: string[] }>(`/api/admins/${adminId}/plans`),
    enabled: !!adminId,
  });
}

export function useAccount() {
  return useQuery({
    queryKey: ["account"],
    queryFn: () => api<{ admin: Admin; permissions: string[] }>("/api/account"),
    retry: false,
  });
}

// --- roles ---

export function useRoles() {
  return useQuery({ queryKey: ["roles"], queryFn: () => api<{ roles: Role[] }>("/api/roles") });
}

export function useCreateRole() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (input: { name: string; permissions: string[] }) =>
      api<{ role: Role }>("/api/roles", { method: "POST", body: input }),
    onSuccess: () => qc.invalidateQueries({ queryKey: ["roles"] }),
  });
}

export function useUpdateRole() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (args: { id: string; input: { name: string; permissions: string[] } }) =>
      api<{ role: Role }>(`/api/roles/${args.id}`, { method: "PUT", body: args.input }),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["roles"] });
      qc.invalidateQueries({ queryKey: ["admins"] });
      qc.invalidateQueries({ queryKey: ["account"] });
    },
  });
}

export function useDeleteRole() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => api<void>(`/api/roles/${id}`, { method: "DELETE" }),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["roles"] });
      qc.invalidateQueries({ queryKey: ["admins"] });
    },
  });
}

// --- 2FA self-enrollment ---

export function useSetupTOTP() {
  return useMutation({ mutationFn: () => api<{ secret: string; url: string }>("/api/account/2fa/setup", { method: "POST" }) });
}

export function useConfirmTOTP() {
  return useMutation({ mutationFn: (code: string) => api<{ enabled: boolean }>("/api/account/2fa/confirm", { method: "POST", body: { code } }) });
}

export function useDisableTOTP() {
  return useMutation({ mutationFn: (code: string) => api<{ enabled: boolean }>("/api/account/2fa/disable", { method: "POST", body: { code } }) });
}
