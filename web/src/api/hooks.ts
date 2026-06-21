import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { api } from "./client";
import type { CreateUserInput, ListUsersResponse, Node, User } from "./types";

// --- users ---

export function useUsers(params: { search?: string; status?: string; limit?: number; offset?: number }) {
  return useQuery({
    queryKey: ["users", params],
    queryFn: () => api<ListUsersResponse>("/api/users", { query: params }),
  });
}

export function useCreateUser() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (input: CreateUserInput) =>
      api<{ user: User; warning?: string }>("/api/users", { method: "POST", body: input }),
    onSuccess: () => qc.invalidateQueries({ queryKey: ["users"] }),
  });
}

export interface BulkCreateInput {
  prefix: string;
  count: number;
  start?: number;
  pad?: number;
  note?: string;
  data_limit?: number;
  expire_at?: string | null;
  device_limit?: number;
  reset_strategy?: string;
  inbound_ids?: string[];
  on_hold?: boolean;
}

export function useBulkCreateUsers() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (input: BulkCreateInput) =>
      api<{ created: User[]; created_count: number; failures: { username: string; error: string }[] }>(
        "/api/users/bulk",
        { method: "POST", body: input },
      ),
    onSuccess: () => qc.invalidateQueries({ queryKey: ["users"] }),
  });
}

export interface ImportUsersInput {
  source: "3xui" | "marzban";
  data: unknown;
  inbound_ids?: string[];
}

export function useImportUsers() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (input: ImportUsersInput) =>
      api<{ parsed: number; created: User[]; created_count: number; failures: { username: string; error: string }[] }>(
        "/api/users/import",
        { method: "POST", body: input },
      ),
    onSuccess: () => qc.invalidateQueries({ queryKey: ["users"] }),
  });
}

export interface UpdateUserInput {
  note?: string;
  status?: string;
  data_limit?: number;
  expire_at?: string | null;
  device_limit?: number;
  reset_strategy?: string;
  inbound_ids?: string[];
}

export function useUpdateUser() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ id, input }: { id: string; input: UpdateUserInput }) =>
      api<{ user: User; warning?: string }>(`/api/users/${id}`, { method: "PUT", body: input }),
    onSuccess: () => qc.invalidateQueries({ queryKey: ["users"] }),
  });
}

export interface UsagePoint {
  time: string;
  up: number;
  down: number;
}

export function useUserUsage(id: string | null) {
  return useQuery({
    queryKey: ["usage", id],
    enabled: !!id,
    queryFn: () => api<{ points: UsagePoint[] }>(`/api/users/${id}/usage`, { query: { bucket: "1d" } }),
  });
}

export function useDeleteUser() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => api<void>(`/api/users/${id}`, { method: "DELETE" }),
    onSuccess: () => qc.invalidateQueries({ queryKey: ["users"] }),
  });
}

// --- nodes ---

export function useNodes() {
  return useQuery({
    queryKey: ["nodes"],
    queryFn: () => api<{ nodes: Node[] }>("/api/nodes"),
  });
}

export function useCreateNode() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (input: { name: string; address: string; core: string; usage_ratio?: number; endpoint?: string }) =>
      api<{ node: Node; warning?: string }>("/api/nodes", { method: "POST", body: input }),
    onSuccess: () => qc.invalidateQueries({ queryKey: ["nodes"] }),
  });
}

export function useUpdateNode() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ id, input }: { id: string; input: { name?: string; address?: string; usage_ratio?: number; endpoint?: string; speed_limit?: number; geo_block?: string[] } }) =>
      api<{ node: Node; warning?: string }>(`/api/nodes/${id}`, { method: "PUT", body: input }),
    onSuccess: () => qc.invalidateQueries({ queryKey: ["nodes"] }),
  });
}

export function useDeleteNode() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => api<void>(`/api/nodes/${id}`, { method: "DELETE" }),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["nodes"] });
      qc.invalidateQueries({ queryKey: ["inbounds-all"] });
    },
  });
}

// --- capabilities (per-core option matrix) ---

export interface CoreCapability {
  protocols: string[];
  transports: string[];
  securities: string[];
  udp_native: string[];
  no_transport: string[];
  protocol_securities?: Record<string, string[]>;
}

export interface Capabilities {
  xray: CoreCapability;
  singbox: CoreCapability;
}

// useCapabilities fetches the authoritative per-core capability matrix. The data
// is effectively static (it only changes when the binary changes), so we cache
// it indefinitely and never refetch within a session.
export function useCapabilities() {
  return useQuery({
    queryKey: ["capabilities"],
    queryFn: () => api<Capabilities>("/api/capabilities"),
    staleTime: Infinity,
    gcTime: Infinity,
    refetchOnWindowFocus: false,
    refetchOnMount: false,
    refetchOnReconnect: false,
  });
}

// --- inbounds (per node) ---

export interface Inbound {
  id: string;
  node_id: string;
  tag: string;
  protocol: string;
  listen: string;
  port: number;
  network: string;
  security: string;
  sni?: string[];
  path?: string;
  host?: string[];
  flow?: string;
  enabled: boolean;
  geo_policy?: { allowed_countries?: string[]; blocked_countries?: string[] } | null;
}

export interface CreateInboundInput {
  node_id: string;
  tag: string;
  protocol: string;
  port: number;
  network?: string;
  security?: string;
  sni?: string[];
  path?: string;
  host?: string[];
  flow?: string;
  raw?: Record<string, unknown>;
  enabled: boolean;
  geo_policy?: { allowed_countries?: string[]; blocked_countries?: string[] } | null;
}

export function useNodeInbounds(nodeId: string | null) {
  return useQuery({
    queryKey: ["inbounds", nodeId],
    enabled: !!nodeId,
    queryFn: () => api<{ inbounds: Inbound[] }>("/api/inbounds", { query: { node_id: nodeId! } }),
  });
}

export function useCreateInbound() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (input: CreateInboundInput) =>
      api<{ inbound: Inbound; warning?: string }>("/api/inbounds", { method: "POST", body: input }),
    onSuccess: (_d, v) => {
      qc.invalidateQueries({ queryKey: ["inbounds", v.node_id] });
      qc.invalidateQueries({ queryKey: ["inbounds-all"] });
    },
  });
}

export interface UpdateInboundInput {
  listen?: string;
  port: number;
  network?: string;
  security?: string;
  sni?: string[];
  path?: string;
  host?: string[];
  flow?: string;
  raw?: Record<string, unknown>;
  enabled: boolean;
  geo_policy?: { allowed_countries?: string[]; blocked_countries?: string[] } | null;
}

export function useUpdateInbound() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ id, input }: { id: string; input: UpdateInboundInput }) =>
      api<{ inbound: Inbound; warning?: string }>(`/api/inbounds/${id}`, { method: "PUT", body: input }),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["inbounds"] });
      qc.invalidateQueries({ queryKey: ["inbounds-all"] });
    },
  });
}

export function useDeleteInbound() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => api<void>(`/api/inbounds/${id}`, { method: "DELETE" }),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["inbounds"] });
      qc.invalidateQueries({ queryKey: ["inbounds-all"] });
    },
  });
}

// --- inbounds (flattened across all nodes, for binding pickers) ---

export interface InboundOption {
  id: string;
  tag: string;
  protocol: string;
  nodeName: string;
}

export function useAllInbounds() {
  return useQuery({
    queryKey: ["inbounds-all"],
    queryFn: async (): Promise<InboundOption[]> => {
      const { nodes } = await api<{ nodes: Node[] }>("/api/nodes");
      const lists = await Promise.all(
        nodes.map((n) => api<{ inbounds: { id: string; tag: string; protocol: string }[] }>("/api/inbounds", { query: { node_id: n.id } })),
      );
      return lists.flatMap((l, i) =>
        (l.inbounds ?? []).map((ib) => ({ id: ib.id, tag: ib.tag, protocol: ib.protocol, nodeName: nodes[i].name })),
      );
    },
  });
}

// --- subscription hosts (per inbound) ---

export type HostSecurity = "inbound_default" | "none" | "tls" | "reality";

// SubHost mirrors the backend domain.SubHost JSON shape. A host overrides the
// inbound's endpoint in subscription links; remark/address support template
// variables ({USERNAME}, {SERVER_IP}, {SERVER_IPV6}).
export interface SubHost {
  id: string;
  inbound_id: string;
  remark: string;
  address: string;
  port: number | null;
  sni: string;
  host: string;
  path: string;
  alpn: string;
  fingerprint: string;
  security: HostSecurity;
  allow_insecure: boolean;
  mux_enable: boolean;
  fragment: string;
  priority: number;
  enabled: boolean;
  created_at: string;
}

// SubHostBody is the create/update payload. On create the inbound_id is added
// by the mutation hook; on update the backend ignores it.
export interface SubHostBody {
  inbound_id?: string;
  remark: string;
  address: string;
  port: number | null;
  sni: string;
  host: string;
  path: string;
  alpn: string;
  fingerprint: string;
  security: HostSecurity;
  allow_insecure: boolean;
  mux_enable: boolean;
  fragment: string;
  enabled: boolean;
}

export function useSubHosts(inboundId: string | null) {
  return useQuery({
    queryKey: ["sub-hosts", inboundId],
    enabled: !!inboundId,
    queryFn: () => api<{ hosts: SubHost[] }>("/api/sub-hosts", { query: { inbound_id: inboundId! } }),
  });
}

export function useCreateSubHost(inboundId: string | null) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (body: SubHostBody) =>
      api<{ host: SubHost }>("/api/sub-hosts", { method: "POST", body: { ...body, inbound_id: inboundId } }),
    onSuccess: () => qc.invalidateQueries({ queryKey: ["sub-hosts", inboundId] }),
  });
}

export function useUpdateSubHost(inboundId: string | null) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ id, body }: { id: string; body: SubHostBody }) =>
      api<{ host: SubHost }>(`/api/sub-hosts/${id}`, { method: "PUT", body }),
    onSuccess: () => qc.invalidateQueries({ queryKey: ["sub-hosts", inboundId] }),
  });
}

export function useDeleteSubHost(inboundId: string | null) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => api<void>(`/api/sub-hosts/${id}`, { method: "DELETE" }),
    onSuccess: () => qc.invalidateQueries({ queryKey: ["sub-hosts", inboundId] }),
  });
}

export function useReorderSubHosts(inboundId: string | null) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (ids: string[]) => api<void>("/api/sub-hosts/reorder", { method: "POST", body: { ids } }),
    onSuccess: () => qc.invalidateQueries({ queryKey: ["sub-hosts", inboundId] }),
  });
}
