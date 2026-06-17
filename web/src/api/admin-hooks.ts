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
    mutationFn: (input: { username: string; password: string; sudo: boolean; role_id?: string | null; user_quota?: number; traffic_quota?: number }) =>
      api<{ admin: Admin }>("/api/admins", { method: "POST", body: input }),
    onSuccess: () => qc.invalidateQueries({ queryKey: ["admins"] }),
  });
}

export function useDeleteAdmin() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => api<void>(`/api/admins/${id}`, { method: "DELETE" }),
    onSuccess: () => qc.invalidateQueries({ queryKey: ["admins"] }),
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
