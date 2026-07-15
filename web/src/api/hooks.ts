import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { ensureArray } from "@/lib/utils";
import { api } from "./client";
import type { CoreType, CreateUserInput, EnrollmentBundle, ListUsersResponse, Node, NodeDiagnostics, NodeEnrollmentPhase, User } from "./types";

// --- panel version ---

// useVersion fetches the actually running panel build version from the backend
// (GET /api/version). The value only changes when the binary changes, so it is
// cached for the whole session.
export function useVersion() {
  return useQuery({
    queryKey: ["version"],
    queryFn: () => api<{ version: string }>("/api/version"),
    staleTime: Infinity,
    gcTime: Infinity,
    refetchOnWindowFocus: false,
    refetchOnMount: false,
    refetchOnReconnect: false,
    select: (d) => d.version,
  });
}

// --- users ---

export function useUsers(params: {
  search?: string;
  status?: string;
  status_group?: string;
  limit?: number;
  offset?: number;
}) {
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

export function useBulkDeleteUsers() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (ids: string[]) =>
      api<{ deleted: number; failures: { id: string; error: string }[] }>("/api/users/bulk-delete", {
        method: "POST",
        body: { ids },
      }),
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
    mutationFn: (input: {
      name: string;
      address: string;
      core: string;
      enabled_cores?: string[];
      usage_ratio?: number;
      endpoint?: string;
      region?: string;
      country_code?: string;
      location_auto?: boolean;
    }) =>
      api<{ node: Node; warning?: string }>("/api/nodes", { method: "POST", body: input }),
    onSuccess: () => qc.invalidateQueries({ queryKey: ["nodes"] }),
  });
}

export function useUpdateNode() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ id, input }: { id: string; input: {
      name?: string;
      address?: string;
      core?: string;
      enabled_cores?: string[];
      usage_ratio?: number;
      endpoint?: string;
      region?: string;
      country_code?: string;
      location_auto?: boolean;
      speed_limit?: number;
      geo_block?: string[];
    } }) =>
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
      qc.invalidateQueries({ queryKey: ["inbounds-fleet"] });
    },
  });
}

export function useNodeEnrollment() {
  return useQuery({
    queryKey: ["nodes", "enrollment"],
    queryFn: () => api<EnrollmentBundle>("/api/nodes/enrollment"),
    enabled: false,
  });
}

export function useTestNodeConnection() {
  return useMutation({
    mutationFn: (id: string) =>
      api<{
        diagnostics: NodeDiagnostics;
        panel_ca_fingerprint?: string;
        ca_match?: boolean;
        enrollment_phase?: NodeEnrollmentPhase;
      }>(`/api/nodes/${id}/test`, { method: "POST" }),
  });
}

export function useNodeDebugBundle() {
  return useMutation({
    mutationFn: (id: string) => api<{ debug_text: string }>(`/api/nodes/${id}/debug`),
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
  core?: CoreType | "";
  protocol: string;
  listen: string;
  port: number;
  network: string;
  security: string;
  sni?: string[];
  path?: string;
  host?: string[];
  flow?: string;
  evasion_profile_id?: string;
  enabled: boolean;
  geo_policy?: { allowed_countries?: string[]; blocked_countries?: string[] } | null;
  raw?: Record<string, unknown>;
}

export interface CreateInboundInput {
  node_id: string;
  tag: string;
  core?: CoreType | "";
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
      qc.invalidateQueries({ queryKey: ["inbounds-fleet"] });
    },
  });
}

export interface UpdateInboundInput {
  listen?: string;
  port: number;
  network?: string;
  security?: string;
  core?: CoreType | "";
  sni?: string[];
  path?: string;
  host?: string[];
  flow?: string;
  raw?: Record<string, unknown>;
  enabled: boolean;
  geo_policy?: { allowed_countries?: string[]; blocked_countries?: string[] } | null;
}

/** Build a full PUT body from an inbound row so partial toggles never wipe host/raw. */
export function inboundToUpdateInput(ib: Inbound, overrides: Partial<UpdateInboundInput> = {}): UpdateInboundInput {
  return {
    listen: ib.listen,
    port: ib.port,
    network: ib.network,
    security: ib.security,
    core: ib.core ?? "",
    sni: ib.sni ?? [],
    path: ib.path ?? "",
    host: ib.host ?? [],
    flow: ib.flow ?? "",
    raw: ib.raw,
    enabled: ib.enabled,
    geo_policy: ib.geo_policy ?? null,
    ...overrides,
  };
}

export function useUpdateInbound() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ id, input }: { id: string; input: UpdateInboundInput }) =>
      api<{ inbound: Inbound; warning?: string }>(`/api/inbounds/${id}`, { method: "PUT", body: input }),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["inbounds"] });
      qc.invalidateQueries({ queryKey: ["inbounds-all"] });
      qc.invalidateQueries({ queryKey: ["inbounds-fleet"] });
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
      qc.invalidateQueries({ queryKey: ["inbounds-fleet"] });
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

export interface InboundFleetRow extends Inbound {
  node_name: string;
}

export function useInboundsFleet() {
  return useQuery({
    queryKey: ["inbounds-fleet"],
    queryFn: () => api<{ inbounds: InboundFleetRow[] }>("/api/inbounds"),
  });
}

export function useAllInbounds() {
  return useQuery({
    queryKey: ["inbounds-all"],
    queryFn: async (): Promise<InboundOption[]> => {
      const { inbounds } = await api<{ inbounds: InboundFleetRow[] }>("/api/inbounds");
      return ensureArray(inbounds).map((ib) => ({
        id: ib.id,
        tag: ib.tag,
        protocol: ib.protocol,
        nodeName: ib.node_name,
      }));
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

export function useCreateSubHost() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (body: SubHostBody & { inbound_id: string }) =>
      api<{ host: SubHost }>("/api/sub-hosts", { method: "POST", body }),
    onSuccess: (_d, v) => qc.invalidateQueries({ queryKey: ["sub-hosts", v.inbound_id] }),
  });
}

export function useUpdateSubHost() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ id, body }: { id: string; inbound_id: string; body: SubHostBody }) =>
      api<{ host: SubHost }>(`/api/sub-hosts/${id}`, { method: "PUT", body }),
    onSuccess: (_d, v) => qc.invalidateQueries({ queryKey: ["sub-hosts", v.inbound_id] }),
  });
}

export function useDeleteSubHost() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ id }: { id: string; inbound_id: string }) =>
      api<void>(`/api/sub-hosts/${id}`, { method: "DELETE" }),
    onSuccess: (_d, v) => qc.invalidateQueries({ queryKey: ["sub-hosts", v.inbound_id] }),
  });
}

export function useReorderSubHosts() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ ids }: { ids: string[]; inbound_id: string }) =>
      api<void>("/api/sub-hosts/reorder", { method: "POST", body: { ids } }),
    onSuccess: (_d, v) => qc.invalidateQueries({ queryKey: ["sub-hosts", v.inbound_id] }),
  });
}

// --- routing rule packs (smart-routing) ---

// PackRoutingRule mirrors the engine-neutral domain.RoutingRule subset a pack
// carries. Node-specific fields (id/node_id) are unset on a pack; they are
// assigned fresh when the pack is applied to a node.
export interface PackRoutingRule {
  priority: number;
  name?: string;
  inbound_tags?: string[];
  domains?: string[];
  ip?: string[];
  port?: string;
  protocols?: string[];
  network?: string;
  outbound_tag?: string;
  balancer_tag?: string;
}

import type { BalancerFleetRow } from "./types";

export function useBalancersFleet() {
  return useQuery({
    queryKey: ["balancers-fleet"],
    queryFn: () => api<{ balancers: BalancerFleetRow[] }>("/api/balancers"),
  });
}

// RoutingPack mirrors the backend domain.RoutingPack. Built-in packs use their
// name as id and are flagged builtin (not editable/deletable); custom packs use
// a uuid id.
export interface RoutingPack {
  id: string;
  name: string;
  description: string;
  category: string;
  builtin: boolean;
  rules: PackRoutingRule[];
  outbounds?: unknown[];
}

// RoutingPackBody is the create/update payload for a custom pack.
export interface RoutingPackBody {
  name: string;
  description: string;
  category: string;
  rules: PackRoutingRule[];
  outbounds?: unknown[];
}

export function useRoutingPacks() {
  return useQuery({
    queryKey: ["routing-packs"],
    queryFn: () => api<{ packs: RoutingPack[] }>("/api/routing-packs"),
  });
}

export function useCreateRoutingPack() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (body: RoutingPackBody) =>
      api<{ pack: RoutingPack }>("/api/routing-packs", { method: "POST", body }),
    onSuccess: () => qc.invalidateQueries({ queryKey: ["routing-packs"] }),
  });
}

export function useUpdateRoutingPack() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ id, body }: { id: string; body: RoutingPackBody }) =>
      api<{ pack: RoutingPack }>(`/api/routing-packs/${id}`, { method: "PUT", body }),
    onSuccess: () => qc.invalidateQueries({ queryKey: ["routing-packs"] }),
  });
}

export function useDeleteRoutingPack() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => api<void>(`/api/routing-packs/${id}`, { method: "DELETE" }),
    onSuccess: () => qc.invalidateQueries({ queryKey: ["routing-packs"] }),
  });
}

// useApplyRoutingPack applies a pack to a node. The backend may respond with
// { status: "saved", warning } when the rules persist but the node resync
// failed; the caller surfaces that warning.
export function useApplyRoutingPack() {
  return useMutation({
    mutationFn: (body: { node_id: string; pack_id: string }) =>
      api<{ status: string; warning?: string }>("/api/routing-packs/apply", { method: "POST", body }),
  });
}

export function useDefaultRoutingPack() {
  return useQuery({
    queryKey: ["routing-packs-default"],
    queryFn: () => api<{ pack_id: string }>("/api/routing-packs/default"),
  });
}

export function useSetDefaultRoutingPack() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (packId: string) =>
      api<{ pack_id: string }>("/api/routing-packs/default", { method: "PUT", body: { pack_id: packId } }),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["routing-packs"] });
      qc.invalidateQueries({ queryKey: ["routing-packs-default"] });
    },
  });
}

// useUserRoutingPack reads a user's per-subscription pack selection ("" = none).
export function useUserRoutingPack(userId: string | null) {
  return useQuery({
    queryKey: ["user-routing-pack", userId],
    enabled: !!userId,
    queryFn: () => api<{ pack_id: string }>(`/api/routing-packs/user/${userId}`),
  });
}

export function useSetUserRoutingPack(userId: string | null) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (packId: string) =>
      api<{ pack_id: string }>(`/api/routing-packs/user/${userId}`, { method: "PUT", body: { pack_id: packId } }),
    onSuccess: () => qc.invalidateQueries({ queryKey: ["user-routing-pack", userId] }),
  });
}

// --- clean-IP scanner (Cloudflare) ---

// CleanIPScan mirrors the backend domain.CleanIPScan JSON shape. Results are
// returned scored best-first (score DESC); lower loss_pct and latency_ms are
// better, higher score is better.
export interface CleanIPScan {
  id: string;
  ip: string;
  latency_ms: number;
  loss_pct: number;
  score: number;
  reachable: boolean;
  throughput_mbps: number;
  scanned_at: string;
}

// useCleanIPResults reads the last scan's cached, scored results.
export function useCleanIPResults() {
  return useQuery({
    queryKey: ["clean-ip-results"],
    queryFn: () => api<{ results: CleanIPScan[] }>("/api/clean-ip/results"),
  });
}

// useScanCleanIP probes a set of candidate IPs and replaces the cached results.
// The backend caps the candidate list at 256 and rejects internal ranges.
export function useScanCleanIP() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (input: { ips: string[]; port?: number }) =>
      api<{ results: CleanIPScan[] }>("/api/clean-ip/scan", { method: "POST", body: input }),
    onSuccess: () => qc.invalidateQueries({ queryKey: ["clean-ip-results"] }),
  });
}

// useMeasureThroughput runs a real download-speed test against one cached
// candidate (identified by its result ID; the backend resolves the IP from
// the cached row) and persists the measured Mbps. This is a real network
// transfer (not a simulation), so it's slower than a latency probe and is
// invoked per-IP, on demand, rather than for the whole list at once.
export function useMeasureThroughput() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (input: { id: string; port?: number }) =>
      api<{ throughput_mbps: number }>("/api/clean-ip/throughput", { method: "POST", body: input }),
    onSuccess: () => qc.invalidateQueries({ queryKey: ["clean-ip-results"] }),
  });
}

// useMeasureAllThroughput runs the download-speed test against every
// reachable cached candidate and returns the refreshed result set. Slower
// than a single measurement since it performs one real transfer per IP
// (bounded by the same worker pool as a scan), but avoids clicking "test
// speed" one card at a time.
export function useMeasureAllThroughput() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (input: { port?: number }) =>
      api<{ results: CleanIPScan[] }>("/api/clean-ip/throughput/all", { method: "POST", body: input }),
    onSuccess: () => qc.invalidateQueries({ queryKey: ["clean-ip-results"] }),
  });
}

// CleanIPSchedule mirrors the backend domain.CleanIPSchedule JSON shape: a
// single-row config that lets the panel re-scan the saved candidate list
// unattended, on a cadence, instead of only on demand.
export interface CleanIPSchedule {
  enabled: boolean;
  interval_minutes: number;
  port: number;
  ips: string[];
  last_run_at?: string | null;
  updated_at: string;
}

// useCleanIPSchedule reads the current recurring-scan configuration.
export function useCleanIPSchedule() {
  return useQuery({
    queryKey: ["clean-ip-schedule"],
    queryFn: () => api<CleanIPSchedule>("/api/clean-ip/schedule"),
  });
}

// useUpdateCleanIPSchedule persists the recurring-scan configuration
// (enabled flag, cadence, port, candidate IPs).
export function useUpdateCleanIPSchedule() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (input: { enabled: boolean; interval_minutes: number; port?: number; ips: string[] }) =>
      api<{ ok: boolean }>("/api/clean-ip/schedule", { method: "PUT", body: input }),
    onSuccess: () => qc.invalidateQueries({ queryKey: ["clean-ip-schedule"] }),
  });
}

// --- IP-limit enforcement ---

// IPLimitAction is the enforcement action taken when a user exceeds its device
// (online-IP) limit. "kill_connections" only works on Xray nodes; sing-box
// nodes have no runtime connection API and degrade to "disable_temporarily".
export type IPLimitAction = "warn" | "disable_temporarily" | "kill_connections";

// IPLimitPolicy mirrors the backend domain.IPLimitPolicy singleton. When
// enabled is false, ShareGuard keeps its prior detection-only behavior.
export interface IPLimitPolicy {
  enabled: boolean;
  action: IPLimitAction;
  alert_cooldown: number; // seconds; per-user alert/action dedup
  restore_after: number; // seconds; auto-undo disable_temporarily
}

// IPLimitEvent is the audit record of a detected violation / enforcement action.
export interface IPLimitEvent {
  id: string;
  user_id: string;
  username: string;
  online_ips: number;
  limit: number;
  action: string;
  created_at: string;
}

export function useIPLimitPolicy() {
  return useQuery({
    queryKey: ["ip-limit-policy"],
    queryFn: () => api<{ policy: IPLimitPolicy }>("/api/ip-limit/policy"),
  });
}

export function useUpdateIPLimitPolicy() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (policy: IPLimitPolicy) =>
      api<{ policy: IPLimitPolicy }>("/api/ip-limit/policy", { method: "PUT", body: policy }),
    onSuccess: () => qc.invalidateQueries({ queryKey: ["ip-limit-policy"] }),
  });
}

export function useIPLimitEvents() {
  return useQuery({
    queryKey: ["ip-limit-events"],
    queryFn: () => api<{ events: IPLimitEvent[] }>("/api/ip-limit/events"),
  });
}
