import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { api } from "./client";
import type { Balancer, LogEntry, Outbound, Overview, RoutingRule, UserSub } from "./types";

// --- overview / logs / system ---

export function useOverview() {
  return useQuery({ queryKey: ["overview"], queryFn: () => api<Overview>("/api/overview"), refetchInterval: 5000 });
}

export function useSystem() {
  return useQuery({
    queryKey: ["system"],
    queryFn: () => api<{
      uptime_seconds: number;
      os: string;
      arch: string;
      go_version: string;
      goroutines: number;
      mem_alloc_bytes: number;
      mem_sys_bytes: number;
      hostname: string;
    }>("/api/system"),
    refetchInterval: 5000,
  });
}

export function useLogs(level: string) {
  return useQuery({
    queryKey: ["logs", level],
    queryFn: () => api<{ entries: LogEntry[] }>("/api/logs", { query: { level, limit: 300 } }),
    refetchInterval: 5000,
  });
}

// useTrafficSamples collects bandwidth snapshots from the overview polling to
// render a real-time mini chart. The hook keeps a fixed-size ring in state.
import { useEffect, useRef, useState as useStateReact } from "react";

export function useTrafficSamples(currentTotal: number, maxSamples = 30) {
  const prev = useRef(currentTotal);
  const [samples, setSamples] = useStateReact<number[]>([]);

  useEffect(() => {
    if (currentTotal === 0) return;
    const delta = currentTotal - prev.current;
    prev.current = currentTotal;
    if (delta < 0) return; // reset / first load
    setSamples((s) => [...s.slice(-(maxSamples - 1)), delta]);
  }, [currentTotal, maxSamples]);

  return samples;
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
    useUpdate() {
      const qc = useQueryClient();
      return useMutation({
        mutationFn: ({ id, body }: { id: string; body: Record<string, unknown> }) =>
          api(`/api/${path}/${id}`, { method: "PUT", body }),
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

// --- backup ---

export function useExportBackup() {
  return useMutation({
    mutationFn: async () => {
      const data = await api<unknown>("/api/backup");
      const blob = new Blob([JSON.stringify(data, null, 2)], { type: "application/json" });
      const url = URL.createObjectURL(blob);
      const a = document.createElement("a");
      a.href = url;
      a.download = `vortexui-backup-${new Date().toISOString().slice(0, 10)}.json`;
      a.click();
      URL.revokeObjectURL(url);
    },
  });
}

export function useRestoreBackup() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async (file: File) => {
      const text = await file.text();
      const body = JSON.parse(text);
      return api<{ restored: Record<string, number> }>("/api/backup/restore", { method: "POST", body });
    },
    onSuccess: () => {
      qc.invalidateQueries();
    },
  });
}

// --- node logs ---

export function useNodeLogs(nodeId: string | null, limit = 200) {
  return useQuery({
    queryKey: ["node-logs", nodeId],
    enabled: !!nodeId,
    queryFn: () => api<{ lines: string[] }>(`/api/nodes/${nodeId}/logs`, { query: { limit } }),
    refetchInterval: 5000,
  });
}

export function useNodeStatus(nodeId: string | null) {
  return useQuery({
    queryKey: ["node-status", nodeId],
    enabled: !!nodeId,
    queryFn: () => api<{
      id: string; name: string; address: string; core: string;
      core_version: string; agent_version: string; status: string;
      last_seen: string | null; health: { cpu_percent: number; mem_percent: number; disk_percent: number; core_running: boolean; connections: number };
    }>(`/api/nodes/${nodeId}/status`),
    refetchInterval: 5000,
  });
}

export function useRestartCore() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (nodeId: string) => api(`/api/nodes/${nodeId}/restart`, { method: "POST" }),
    onSuccess: () => { qc.invalidateQueries({ queryKey: ["nodes"] }); qc.invalidateQueries({ queryKey: ["overview"] }); },
  });
}

export function useStopCore() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (nodeId: string) => api(`/api/nodes/${nodeId}/stop`, { method: "POST" }),
    onSuccess: () => { qc.invalidateQueries({ queryKey: ["nodes"] }); qc.invalidateQueries({ queryKey: ["overview"] }); },
  });
}

// --- user online ---

export function useUserOnline(id: string | null) {
  return useQuery({
    queryKey: ["user-online", id],
    enabled: !!id,
    queryFn: () =>
      api<{ live_connections: number; live_tracking: boolean; active_devices: number; device_tracking: boolean }>(`/api/users/${id}/online`),
    refetchInterval: 10000,
  });
}
