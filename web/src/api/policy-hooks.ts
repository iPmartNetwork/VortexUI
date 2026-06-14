import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { api } from "./client";
import type { Balancer, LogEntry, Outbound, Overview, RoutingRule, UserSub } from "./types";

// --- overview / logs ---

export function useOverview() {
  return useQuery({ queryKey: ["overview"], queryFn: () => api<Overview>("/api/overview"), refetchInterval: 10000 });
}

export function useLogs(level: string) {
  return useQuery({
    queryKey: ["logs", level],
    queryFn: () => api<{ entries: LogEntry[] }>("/api/logs", { query: { level, limit: 300 } }),
    refetchInterval: 5000,
  });
}

// --- user extras ---

export function useUserSub(id: string | null) {
  return useQuery({ queryKey: ["usersub", id], enabled: !!id, queryFn: () => api<UserSub>(`/api/users/${id}/sub`) });
}

export function useResetUser() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => api(`/api/users/${id}/reset`, { method: "POST" }),
    onSuccess: () => qc.invalidateQueries({ queryKey: ["users"] }),
  });
}

export function useRevokeSub() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => api(`/api/users/${id}/revoke-sub`, { method: "POST" }),
    onSuccess: () => qc.invalidateQueries({ queryKey: ["users"] }),
  });
}

export function useReality() {
  return useMutation({ mutationFn: () => api<{ private_key: string; public_key: string; short_id: string }>("/api/reality/keypair") });
}

// --- generic per-node policy CRUD factory ---

function makePolicyHooks<T>(path: string, key: string) {
  return {
    useList(nodeId: string | null) {
      return useQuery({
        queryKey: [key, nodeId],
        enabled: !!nodeId,
        queryFn: () => api<Record<string, T[]>>(`/api/${path}`, { query: { node_id: nodeId! } }),
      });
    },
    useCreate() {
      const qc = useQueryClient();
      return useMutation({
        mutationFn: (body: Record<string, unknown>) => api(`/api/${path}`, { method: "POST", body }),
        onSuccess: () => qc.invalidateQueries({ queryKey: [key] }),
      });
    },
    useDelete() {
      const qc = useQueryClient();
      return useMutation({
        mutationFn: (id: string) => api<void>(`/api/${path}/${id}`, { method: "DELETE" }),
        onSuccess: () => qc.invalidateQueries({ queryKey: [key] }),
      });
    },
  };
}

export const outboundHooks = makePolicyHooks<Outbound>("outbounds", "outbounds");
export const routingHooks = makePolicyHooks<RoutingRule>("routing", "routing");
export const balancerHooks = makePolicyHooks<Balancer>("balancers", "balancers");
