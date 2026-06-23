import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { api } from "./client";
import type { AuditEntry, Balancer, LogEntry, Outbound, Overview, RoutingRule, UserSub } from "./types";

export function useAudit() {
  return useQuery({
    queryKey: ["audit"],
    queryFn: () => api<{ entries: AuditEntry[] }>("/api/audit", { query: { limit: 200 } }),
    refetchInterval: 10000,
  });
}

// --- overview / logs / system ---

export function useOverview() {
  return useQuery({ queryKey: ["overview"], queryFn: () => api<Overview>("/api/overview"), refetchInterval: 5000 });
}

// useTrafficSeries fetches fleet-wide bucketed throughput for the dashboard
// chart. Defaults to the last hour in 1-minute buckets (server-side).
export function useTrafficSeries() {
  return useQuery({
    queryKey: ["traffic-series"],
    queryFn: () => api<{ points: { time: string; up: number; down: number }[] }>("/api/traffic/series"),
    refetchInterval: 15000,
  });
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
      cpu_percent: number;
      mem_percent: number;
      disk_percent: number;
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
    onSuccess: (_d, id) => {
      qc.invalidateQueries({ queryKey: ["users"] });
      // Refresh the subscription link/QR so the UI shows the new token, not the
      // revoked one.
      qc.invalidateQueries({ queryKey: ["usersub", id] });
    },
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

export function useExportUserBackup() {
  return useMutation({
    mutationFn: async () => {
      const data = await api<unknown>("/api/account/backup/users");
      const blob = new Blob([JSON.stringify(data, null, 2)], { type: "application/json" });
      const url = URL.createObjectURL(blob);
      const a = document.createElement("a");
      a.href = url;
      a.download = `my-users-backup-${new Date().toISOString().slice(0, 10)}.json`;
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

// useUpdateGeo refreshes a node's geoip/geosite routing databases. With no body
// the server uses the Iran-focused defaults (geoip:ir / geosite:ir).
export function useUpdateGeo() {
  return useMutation({
    mutationFn: (nodeId: string) =>
      api<{ ok: boolean; geoip_bytes: number; geosite_bytes: number }>(`/api/nodes/${nodeId}/geo-update`, { method: "POST", body: {} }),
  });
}

// --- api tokens ---

export function useAPITokens() {
  return useQuery({ queryKey: ["api-tokens"], queryFn: () => api<{ tokens: { id: string; name: string; admin_id: string; created_at: string; last_used_at: string | null }[] }>("/api/tokens") });
}

export function useCreateAPIToken() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (name: string) => api<{ token: { id: string; name: string }; raw: string }>("/api/tokens", { method: "POST", body: { name } }),
    onSuccess: () => qc.invalidateQueries({ queryKey: ["api-tokens"] }),
  });
}

export function useDeleteAPIToken() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => api<void>(`/api/tokens/${id}`, { method: "DELETE" }),
    onSuccess: () => qc.invalidateQueries({ queryKey: ["api-tokens"] }),
  });
}


export function useUserOnline(id: string | null) {
  return useQuery({
    queryKey: ["user-online", id],
    enabled: !!id,
    queryFn: () =>
      api<{ live_connections: number; live_tracking: boolean; active_devices: number; device_tracking: boolean }>(`/api/users/${id}/online`),
    refetchInterval: 10000,
  });
}

// useUserOnlineIPs lists the distinct source IPs a user is currently connected
// from (account-sharing detection), refreshed periodically.
export function useUserOnlineIPs(id: string | null) {
  return useQuery({
    queryKey: ["user-online-ips", id],
    enabled: !!id,
    queryFn: () =>
      api<{ ips: { ip: string; last_seen: string }[]; count: number; tracking: boolean }>(`/api/users/${id}/online-ips`),
    refetchInterval: 10000,
  });
}
