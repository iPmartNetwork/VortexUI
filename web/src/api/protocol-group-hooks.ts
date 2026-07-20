import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { api } from "./client";

// --- Types ---

export interface ProtocolGroup {
  id: string;
  node_id: string;
  name: string;
  inbound_ids: string[];
  probe_url: string;
  probe_interval: number;
  probe_timeout: number;
  max_retries: number;
  created_at: string;
  updated_at: string;
}

export interface ISPProfile {
  id: string;
  group_id: string;
  isp_identifier: string;
  country_code: string;
  preferred_protocols: string[];
  created_at: string;
}

export interface SwitchEvent {
  id: string;
  user_id: string;
  node_id: string;
  group_id?: string;
  source_protocol: string;
  target_protocol: string;
  isp?: string;
  timestamp: string;
}

export interface SwitchSummary {
  total_switches: number;
  by_protocol: Record<string, number>;
  by_node: Record<string, number>;
  by_isp: Record<string, number>;
  top_switch_pairs?: { source: string; target: string; count: number }[];
}

// --- Protocol Group Hooks ---

export function useProtocolGroups(nodeId: string | null) {
  return useQuery({
    queryKey: ["protocol-groups", nodeId],
    queryFn: () =>
      api<{ groups: ProtocolGroup[] }>("/api/protocol-groups", {
        query: { node_id: nodeId ?? "" },
      }),
    enabled: !!nodeId,
    select: (d) => d.groups ?? [],
  });
}

export function useProtocolGroup(id: string | null) {
  return useQuery({
    queryKey: ["protocol-group", id],
    queryFn: () => api<ProtocolGroup>(`/api/protocol-groups/${id}`),
    enabled: !!id,
  });
}

export interface CreateProtocolGroupInput {
  node_id: string;
  name: string;
  inbound_ids: string[];
  probe_url?: string;
  probe_interval?: number;
  probe_timeout?: number;
  max_retries?: number;
}

export function useCreateProtocolGroup() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (input: CreateProtocolGroupInput) =>
      api<{ group: ProtocolGroup }>("/api/protocol-groups", { method: "POST", body: input }),
    onSuccess: () => qc.invalidateQueries({ queryKey: ["protocol-groups"] }),
  });
}

export interface UpdateProtocolGroupInput {
  id: string;
  name: string;
  inbound_ids: string[];
  probe_url?: string;
  probe_interval?: number;
  probe_timeout?: number;
  max_retries?: number;
}

export function useUpdateProtocolGroup() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ id, ...body }: UpdateProtocolGroupInput) =>
      api<{ group: ProtocolGroup }>(`/api/protocol-groups/${id}`, { method: "PUT", body }),
    onSuccess: () => qc.invalidateQueries({ queryKey: ["protocol-groups"] }),
  });
}

export function useDeleteProtocolGroup() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: string) =>
      api<void>(`/api/protocol-groups/${id}`, { method: "DELETE" }),
    onSuccess: () => qc.invalidateQueries({ queryKey: ["protocol-groups"] }),
  });
}

export function useReorderGroupInbounds() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ groupId, inboundIds }: { groupId: string; inboundIds: string[] }) =>
      api<void>(`/api/protocol-groups/${groupId}/reorder`, {
        method: "POST",
        body: { inbound_ids: inboundIds },
      }),
    onSuccess: () => qc.invalidateQueries({ queryKey: ["protocol-groups"] }),
  });
}

// --- ISP Profile Hooks ---

export function useISPProfiles(groupId: string | null) {
  return useQuery({
    queryKey: ["isp-profiles", groupId],
    queryFn: () =>
      api<{ profiles: ISPProfile[] }>("/api/isp-profiles", {
        query: { group_id: groupId ?? "" },
      }),
    enabled: !!groupId,
    select: (d) => d.profiles ?? [],
  });
}

export interface CreateISPProfileInput {
  group_id: string;
  isp_identifier: string;
  country_code: string;
  preferred_protocols: string[];
}

export function useCreateISPProfile() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (input: CreateISPProfileInput) =>
      api<{ profile: ISPProfile }>("/api/isp-profiles", { method: "POST", body: input }),
    onSuccess: () => qc.invalidateQueries({ queryKey: ["isp-profiles"] }),
  });
}

export interface UpdateISPProfileInput {
  id: string;
  isp_identifier: string;
  country_code: string;
  preferred_protocols: string[];
}

export function useUpdateISPProfile() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ id, ...body }: UpdateISPProfileInput) =>
      api<{ profile: ISPProfile }>(`/api/isp-profiles/${id}`, { method: "PUT", body }),
    onSuccess: () => qc.invalidateQueries({ queryKey: ["isp-profiles"] }),
  });
}

export function useDeleteISPProfile() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: string) =>
      api<void>(`/api/isp-profiles/${id}`, { method: "DELETE" }),
    onSuccess: () => qc.invalidateQueries({ queryKey: ["isp-profiles"] }),
  });
}

// --- Switch Event Hooks ---

export interface RecordSwitchEventInput {
  user_id: string;
  node_id: string;
  group_id?: string;
  source_protocol: string;
  target_protocol: string;
  isp?: string;
}

export function useRecordSwitchEvent() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (input: RecordSwitchEventInput) =>
      api<{ event: SwitchEvent }>("/api/switch-events", { method: "POST", body: input }),
    onSuccess: () => qc.invalidateQueries({ queryKey: ["switch-summary"] }),
  });
}

export interface SwitchSummaryParams {
  node_id?: string;
  user_id?: string;
  isp?: string;
  from?: string;
  to?: string;
}

export function useSwitchSummary(params: SwitchSummaryParams = {}) {
  return useQuery({
    queryKey: ["switch-summary", params],
    queryFn: () => api<SwitchSummary>("/api/switch-events/summary", { query: params as Record<string, string> }),
    refetchInterval: 30000,
  });
}
