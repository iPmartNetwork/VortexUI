import { useMemo, useState } from "react";
import { ChevronDown, Globe, Info, Network, Plus, Server, Settings, Search, Copy, CheckCircle } from "lucide-react";
import { useMutation, useQueryClient } from "@tanstack/react-query";
import {
  useInboundsFleet,
  useNodes,
  useSubHosts,
  type HostSecurity,
  type Inbound,
  type InboundFleetRow,
  type SubHost,
} from "@/api/hooks";
import type { Node } from "@/api/types";
import { api } from "@/api/client";
import { Badge, Button, Input } from "@/components/ui";
import { NodeInboundsModal } from "@/components/NodeInboundsModal";
import { SubHostsModal } from "@/components/SubHostsModal";
import { InboundBulkBar } from "@/components/InboundBulkBar";
import { InboundExpandedPanel } from "@/components/InboundExpandedPanel";
import { GlassCard, ProtocolBadge, StatusBadge } from "@/components/veltrix";
import { StaggerContainer } from "@/components/StaggerContainer";
import { EmptyState } from "@/components/EmptyState";
import { useI18n } from "@/i18n/i18n";
import { useTitle } from "@/lib/useTitle";
import { useAuth } from "@/auth/auth";
import { cn } from "@/lib/utils";

function transportLabel(ib: InboundFleetRow): string {
  const net = ib.network || "tcp";
  return net.toUpperCase();
}

function securityColor(s?: string): string {
  switch (s) {
    case "reality": return "on_hold";
    case "tls": return "active";
    case "none": case "": case undefined: return "disabled";
    default: return "muted";
  }
}

function securityLabel(s?: string): string {
  if (!s || s === "none") return "none";
  if (s === "inbound_default") return "default";
  return s;
}

function StatBox({ value, label, color }: { value: number | string; label: string; color: string }) {
  return (
    <div className="rounded-2xl bg-bg-elevated border border-border p-4 transition-all duration-200 hover:shadow-md">
      <p className={cn("text-2xl font-black tabular-nums leading-none", color)}>{value}</p>
      <p className="text-[9px] font-bold uppercase tracking-widest text-fg-subtle mt-1.5">{label}</p>
    </div>
  );
}

function ProtocolBar({ proto, count, total }: { proto: string; count: number; total: number }) {
  const colors: Record<string, string> = {
    vless: "bg-cyan-500", vmess: "bg-blue-500", trojan: "bg-green-500",
    shadowsocks: "bg-yellow-500", hysteria2: "bg-purple-500", tuic: "bg-pink-500",
    wireguard: "bg-emerald-500", socks: "bg-orange-500", http: "bg-red-500",
    naive: "bg-indigo-500", dokodemo: "bg-gray-500",
  };
  const barColor = colors[proto] ?? "bg-primary";
  const pct = total > 0 ? (count / total) * 100 : 0;

  return (
    <div className="flex items-center gap-3 group">
      <div className={cn(
        "flex items-center gap-1.5 rounded-lg border border-border/50 bg-surface-2/30 px-2.5 py-1 min-w-[120px]",
      )}>
        <div className={cn("h-2 w-2 rounded-full", barColor)} />
        <span className="text-[11px] font-semibold text-fg uppercase">{proto}</span>
      </div>
      <div className="flex-1 h-2 rounded-full bg-surface-2 overflow-hidden">
        <div className={cn("h-full rounded-full transition-all duration-500", barColor)} style={{ width: `${pct}%` }} />
      </div>
      <span className="text-xs font-bold text-fg tabular-nums w-8 text-end">{count}</span>
    </div>
  );
}

export function Inbounds() {
  useTitle("Inbounds");
  const { can } = useAuth();
  const canWrite = can("inbound:write");
  const fleet = useInboundsFleet();
  const nodes = useNodes();
  const [nodeFilter, setNodeFilter] = useState("");
  const [searchQuery, setSearchQuery] = useState("");
  const [managing, setManaging] = useState<Node | null>(null);
  const [pendingEdit, setPendingEdit] = useState<Inbound | null>(null);
  const [subHostsFor, setSubHostsFor] = useState<Inbound | null>(null);
  const [expandedId, setExpandedId] = useState<string | null>(null);
  const [copiedId, setCopiedId] = useState<string | null>(null);
  const [selected, setSelected] = useState<Set<string>>(new Set());

  const inbounds = fleet.data?.inbounds ?? [];
  const nodeList = nodes.data?.nodes ?? [];

  const filtered = useMemo(() => {
    let result = inbounds;
    if (nodeFilter) result = result.filter((ib) => ib.node_name === nodeFilter);
    if (searchQuery) {
      const q = searchQuery.toLowerCase();
      result = result.filter((ib) =>
        ib.tag?.toLowerCase().includes(q) ||
        ib.node_name.toLowerCase().includes(q) ||
        ib.protocol?.toLowerCase().includes(q)
      );
    }
    return result;
  }, [inbounds, nodeFilter, searchQuery]);

  const nodeNames = useMemo(
    () => [...new Set(inbounds.map((ib) => ib.node_name))].sort(),
    [inbounds],
  );

  const stats = useMemo(() => ({
    total: inbounds.length,
    active: inbounds.filter((ib) => ib.enabled !== false).length,
    nodes: nodeNames.length,
    protocols: Object.keys(inbounds.reduce((acc, ib) => {
      acc[ib.protocol || "unknown"] = (acc[ib.protocol || "unknown"] || 0) + 1;
      return acc;
    }, {} as Record<string, number>)).length,
    byProtocol: inbounds.reduce((acc, ib) => {
      const proto = ib.protocol || "unknown";
      acc[proto] = (acc[proto] || 0) + 1;
      return acc;
    }, {} as Record<string, number>),
    byNode: inbounds.reduce((acc, ib) => {
      acc[ib.node_name] = (acc[ib.node_name] || 0) + 1;
      return acc;
    }, {} as Record<string, number>),
  }), [inbounds, nodeNames]);

  function openNodeManager(nodeName: string, edit?: Inbound) {
    const node = nodeList.find((n) => n.name === nodeName);
    if (node) {
      setPendingEdit(edit ?? null);
      setManaging(node);
    }
  }

  function toggleSelect(id: string) {
    setSelected(prev => {
      const next = new Set(prev);
      if (next.has(id)) next.delete(id);
      else next.add(id);
      return next;
    });
  }

  function selectAll() {
    setSelected(new Set(filtered.map(ib => ib.id)));
  }

  function clearSelection() {
    setSelected(new Set());
  }

  const qc = useQueryClient();
  const bulkMutation = useMutation({
    mutationFn: (input: { ids: string[]; action: string }) =>
      api<{ affected: number }>("/api/inbounds/bulk", { method: "POST", body: input }),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["inbounds-fleet"] });
      clearSelection();
    },
  });

  function bulkAction(action: string) {
    bulkMutation.mutate({ ids: [...selected], action });
  }

  const protocolEntries = useMemo(
    () => Object.entries(stats.byProtocol).sort((a, b) => b[1] - a[1]),
    [stats.byProtocol],
  );

  return (
    <div className="space-y-5 animate-page-enter">
      <NodeInboundsModal
        node={managing}
        initialEdit={pendingEdit}
        onClose={() => { setManaging(null); setPendingEdit(null); }}
      />
      <SubHostsModal inbound={subHostsFor} onClose={() => setSubHostsFor(null)} />

      {/* Header */}
      <div className="flex flex-col sm:flex-row sm:items-start justify-between gap-4">
        <div>
          <h1 className="text-2xl font-bold text-fg tracking-tight">Inbounds & Hosts</h1>
          <p className="text-sm text-fg-muted mt-1">
            {stats.total} inbound{stats.total !== 1 ? "s" : ""} across {stats.nodes} node{stats.nodes !== 1 ? "s" : ""}
          </p>
        </div>
        {canWrite && nodeList.length > 0 && (
          <Button onClick={() => {
            const target = nodeFilter
              ? nodeList.find((n) => n.name === nodeFilter)
              : nodeList[0];
            if (target) setManaging(target);
          }}>
            <Plus size={15} /> New Inbound
          </Button>
        )}
      </div>

      {/* Stats cards */}
      <div className="grid grid-cols-2 sm:grid-cols-4 gap-3">
        <StatBox value={stats.total} label="Total" color="text-fg" />
        <StatBox value={stats.active} label="Active" color="text-success" />
        <StatBox value={stats.nodes} label="Nodes" color="text-primary" />
        <StatBox value={stats.protocols} label="Protocols" color="text-accent" />
      </div>

      {/* Info box */}
      <GlassCard hover={false} className="!p-3.5 bg-primary/5 border-primary/20">
        <div className="flex items-start gap-2.5">
          <div className="h-7 w-7 rounded-lg bg-primary/15 flex items-center justify-center text-primary flex-shrink-0">
            <Info size={14} />
          </div>
          <div className="min-w-0">
            <p className="text-xs font-semibold text-fg">Subscription Hosts</p>
            <p className="text-xs text-fg-muted mt-0.5 leading-relaxed">
              Create alternate domain/SNI combinations for your inbounds. Users can choose which host to use in their subscriptions.
            </p>
          </div>
        </div>
      </GlassCard>

      {/* Search + Filter */}
      <GlassCard hover={false} className="!p-0 overflow-x-hidden">
        <div className="p-4 border-b border-border/40 space-y-3">
          <div className="relative">
            <Search size={14} className="absolute start-3 top-1/2 -translate-y-1/2 text-fg-subtle pointer-events-none" />
            <Input
              placeholder="Search by name, node, protocol..."
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              className="ps-9"
            />
          </div>
          <div className="flex flex-wrap gap-1.5">
            <button
              type="button"
              onClick={() => setNodeFilter("")}
              className={cn(
                "px-3 py-1.5 rounded-lg text-xs font-medium transition-all",
                !nodeFilter
                  ? "bg-primary text-primary-fg shadow-sm"
                  : "text-fg-muted hover:text-fg hover:bg-surface/60 border border-border/50",
              )}
            >
              All ({inbounds.length})
            </button>
            {nodeNames.map((name) => (
              <button
                key={name}
                type="button"
                onClick={() => setNodeFilter(name)}
                className={cn(
                  "px-3 py-1.5 rounded-lg text-xs font-medium transition-all",
                  nodeFilter === name
                    ? "bg-primary text-primary-fg shadow-sm"
                    : "text-fg-muted hover:text-fg hover:bg-surface/60 border border-border/50",
                )}
              >
                {name} ({stats.byNode?.[name] ?? 0})
              </button>
            ))}
          </div>
          {canWrite && filtered.length > 0 && (
            <div className="flex items-center gap-2 pt-2 border-t border-border/20">
              <input
                type="checkbox"
                className="h-3.5 w-3.5 accent-primary rounded"
                checked={selected.size === filtered.length && filtered.length > 0}
                onChange={(e) => e.target.checked ? selectAll() : clearSelection()}
              />
              <span className="text-[10px] text-fg-subtle font-medium">
                Select All ({filtered.length})
              </span>
            </div>
          )}
        </div>

        {/* Loading */}
        {fleet.isLoading && (
          <div className="p-8 space-y-4">
            {Array.from({ length: 3 }).map((_, i) => (
              <div key={i} className="flex items-center gap-4 animate-shimmer bg-gradient-to-r from-surface-2/40 via-surface-2/80 to-surface-2/40 bg-[length:200%_100%] rounded-xl h-16" />
            ))}
          </div>
        )}

        {/* Empty */}
        {!fleet.isLoading && filtered.length === 0 && (
          <div className="py-12">
            <EmptyState
              icon={Globe}
              title={searchQuery || nodeFilter ? "No inbounds match your search" : "No inbounds configured"}
              description={searchQuery || nodeFilter ? "Try adjusting your search or filter." : "Add your first inbound to start routing traffic."}
              action={!searchQuery && !nodeFilter && canWrite && nodeList.length > 0
                ? { label: "Create Inbound", onClick: () => setManaging(nodeList[0]) }
                : undefined}
            />
          </div>
        )}

        {/* Inbound list */}
        {!fleet.isLoading && filtered.length > 0 && (
          <div className="divide-y divide-border/30">
            <StaggerContainer staggerDelay={0.03} yOffset={8}>
              {filtered.map((ib) => (
                <InboundRow
                  key={ib.id}
                  ib={ib}
                  showNode={!nodeFilter}
                  canWrite={canWrite}
                  selected={selected.has(ib.id)}
                  onToggleSelect={() => toggleSelect(ib.id)}
                  expanded={expandedId === ib.id}
                  onToggleExpand={() => setExpandedId((cur) => (cur === ib.id ? null : ib.id))}
                  onAddSubHost={() => setSubHostsFor(ib)}
                  onEdit={() => openNodeManager(ib.node_name, ib)}
                  copiedId={copiedId}
                  onCopy={(text, id) => {
                    navigator.clipboard.writeText(text);
                    setCopiedId(id);
                    setTimeout(() => setCopiedId(null), 2000);
                  }}
                />
              ))}
            </StaggerContainer>
          </div>
        )}

        {/* Footer */}
        {!fleet.isLoading && filtered.length > 0 && (
          <div className="border-t border-border/40 px-4 py-3 text-sm text-fg-muted flex items-center gap-2">
            <Network size={14} className="text-fg-subtle" />
            Showing {filtered.length} of {inbounds.length} inbounds
          </div>
        )}
      </GlassCard>

      {/* Protocol Distribution */}
      {protocolEntries.length > 0 && (
        <GlassCard hover={false} className="!p-4">
          <h3 className="text-sm font-bold text-fg mb-3">Protocol Distribution</h3>
          <div className="space-y-2">
            {protocolEntries.map(([proto, count]) => (
              <ProtocolBar key={proto} proto={proto} count={count} total={stats.total} />
            ))}
          </div>
        </GlassCard>
      )}

      <InboundBulkBar
        selectedCount={selected.size}
        onEnable={() => bulkAction("enable")}
        onDisable={() => bulkAction("disable")}
        onDelete={() => bulkAction("delete")}
        onClearSelection={clearSelection}
        isPending={bulkMutation.isPending}
      />
    </div>
  );
}

function InboundRow({
  ib, showNode, canWrite, selected, onToggleSelect, expanded, onToggleExpand, onAddSubHost, onEdit, copiedId, onCopy,
}: {
  ib: InboundFleetRow;
  showNode: boolean;
  canWrite: boolean;
  selected: boolean;
  onToggleSelect: () => void;
  expanded: boolean;
  onToggleExpand: () => void;
  onAddSubHost: () => void;
  onEdit: () => void;
  copiedId: string | null;
  onCopy: (text: string, id: string) => void;
}) {
  const sniList = (ib.sni ?? []).filter(Boolean);

  return (
    <div className="p-4 hover:bg-surface/30 transition-colors group">
      <div className="flex flex-wrap items-center justify-between gap-3">
        <div className="flex items-center gap-3 min-w-0 flex-1">
          {/* Checkbox */}
          {canWrite && (
            <input
              type="checkbox"
              className="h-3.5 w-3.5 accent-primary rounded flex-shrink-0"
              checked={selected}
              onChange={onToggleSelect}
              onClick={(e) => e.stopPropagation()}
            />
          )}

          {/* Icon */}
          <div className={cn(
            "h-10 w-10 rounded-xl flex items-center justify-center flex-shrink-0 border",
            ib.enabled
              ? "bg-primary/10 text-primary border-primary/20"
              : "bg-surface-2/60 text-fg-subtle border-border/50",
          )}>
            <Globe size={17} />
          </div>

          {/* Info */}
          <div className="min-w-0 flex-1">
            <div className="flex flex-wrap items-center gap-1.5">
              <span className={cn(
                "font-semibold text-sm truncate max-w-[300px]",
                ib.enabled ? "text-fg" : "text-fg-muted",
              )}>
                {ib.tag}
              </span>
              <ProtocolBadge label={ib.protocol} />
              <span className="inline-flex items-center rounded-md border border-border/60 bg-surface-2/70 px-2 py-0.5 text-[10px] font-semibold uppercase tracking-wide text-fg-muted">
                {transportLabel(ib)}
              </span>
              <Badge color={securityColor(ib.security)}>{securityLabel(ib.security)}</Badge>
              {(ib.speed_limit ?? 0) > 0 && (
                <span className="inline-flex items-center gap-0.5 rounded-md bg-amber-500/10 border border-amber-500/20 px-1.5 py-0.5 text-[10px] font-semibold text-amber-400">
                  ⚡ {Math.round((ib.speed_limit ?? 0) / 125000)} Mbps
                </span>
              )}
              {ib.health && ib.health !== "" && (
                <span className={cn(
                  "inline-flex items-center gap-1 rounded-md px-1.5 py-0.5 text-[10px] font-semibold border",
                  ib.health === "healthy" ? "bg-success/10 border-success/20 text-success" :
                  ib.health === "degraded" ? "bg-amber-500/10 border-amber-500/20 text-amber-400" :
                  "bg-danger/10 border-danger/20 text-danger"
                )}>
                  <span className={cn(
                    "h-1.5 w-1.5 rounded-full",
                    ib.health === "healthy" ? "bg-success" :
                    ib.health === "degraded" ? "bg-amber-500" :
                    "bg-danger"
                  )} />
                  {ib.health}
                </span>
              )}
              {ib.notes && (
                <span className="text-fg-subtle" title={ib.notes}>📝</span>
              )}
              {showNode && (
                <span className="inline-flex items-center gap-1 rounded-md bg-surface-2/70 px-2 py-0.5 text-[10px] font-medium text-fg-subtle">
                  <Server size={10} /> {ib.node_name}
                </span>
              )}
            </div>
            <div className="mt-1 flex flex-wrap items-center gap-x-3 gap-y-1 text-xs text-fg-muted">
              <span className="font-mono text-fg-subtle">
                :{ib.port}{ib.port_end ? `-${ib.port_end}` : ""}
              </span>
              {ib.listen && ib.listen !== "" && ib.listen !== "0.0.0.0" && (
                <span className="text-fg-subtle font-mono text-[10px]" title="Listen address">
                  {ib.listen}
                </span>
              )}
              {sniList.length > 0 && (
                <span className="truncate max-w-[240px]" title={sniList.join(", ")}>
                  SNI: <span className="text-fg-subtle">{sniList.join(", ")}</span>
                </span>
              )}
            </div>
          </div>
        </div>

        {/* Actions */}
        <div className="flex items-center gap-1 flex-shrink-0">
          <StatusBadge
            status={ib.enabled ? "active" : "inactive"}
            label={ib.enabled ? "ON" : "OFF"}
            pulse={false}
          />

          {/* Copy address */}
          <button
            type="button"
            onClick={() => onCopy(`0.0.0.0:${ib.port}`, ib.id)}
            className="h-8 w-8 rounded-lg flex items-center justify-center text-fg-subtle hover:text-fg hover:bg-surface-2/60 transition-all"
            title="Copy address"
          >
            {copiedId === ib.id ? <CheckCircle size={14} className="text-success" /> : <Copy size={14} />}
          </button>

          {/* Sub hosts */}
          <button
            type="button"
            onClick={onAddSubHost}
            className="h-8 w-8 rounded-lg flex items-center justify-center text-fg-subtle hover:text-fg hover:bg-surface-2/60 transition-all"
            title="Subscription hosts"
          >
            <Plus size={14} />
          </button>

          {/* Edit */}
          {canWrite && (
            <button
              type="button"
              onClick={onEdit}
              className="h-8 w-8 rounded-lg flex items-center justify-center text-fg-subtle hover:text-fg hover:bg-surface-2/60 transition-all"
              title="Edit inbound"
            >
              <Settings size={14} />
            </button>
          )}

          {/* Expand */}
          <button
            type="button"
            onClick={onToggleExpand}
            className="h-8 w-8 rounded-lg flex items-center justify-center text-fg-subtle hover:text-fg hover:bg-surface-2/60 transition-all"
            aria-label="Toggle details"
          >
            <ChevronDown size={14} className={cn("transition-transform duration-200", expanded && "rotate-180")} />
          </button>
        </div>
      </div>

      {expanded && (
        <>
          <InboundExpandedPanel inboundId={ib.id} notes={ib.notes} />
          <SubHostsSection inboundId={ib.id} onOpen={onAddSubHost} />
        </>
      )}
    </div>
  );
}

function SubHostsSection({ inboundId, onOpen }: { inboundId: string; onOpen: () => void }) {
  const { t } = useI18n();
  const { data, isLoading } = useSubHosts(inboundId);
  const hosts = data?.hosts ?? [];

  return (
    <div className="mt-3 ms-12 space-y-1.5 border-s-2 border-border/30 ps-4">
      <p className="text-[10px] font-semibold uppercase tracking-wider text-fg-subtle">Subscription Hosts</p>
      {isLoading && <p className="text-xs text-fg-subtle">{t("common.loading")}</p>}
      {!isLoading && hosts.length === 0 && (
        <button type="button" onClick={onOpen} className="text-xs text-fg-subtle hover:text-primary transition-colors">
          {t("hosts.empty")} — click to add
        </button>
      )}
      {hosts.map((h: SubHost) => (
        <button
          key={h.id}
          type="button"
          onClick={onOpen}
          className="w-full flex items-center justify-between gap-2 rounded-lg bg-surface/60 border border-border/30 px-3 py-2 text-xs hover:bg-surface-2/60 hover:border-primary/20 transition-colors text-start group/host"
        >
          <div className="flex items-center gap-2 min-w-0">
            <span className={cn("h-1.5 w-1.5 rounded-full flex-shrink-0", h.enabled ? "bg-success" : "bg-fg-subtle")} />
            <span className="font-medium text-fg truncate">{h.remark || "—"}</span>
            <span className="text-fg-subtle truncate font-mono" dir="ltr">
              {h.address}{h.port ? `:${h.port}` : ""}
            </span>
          </div>
          <Badge color={securityColor(h.security as HostSecurity)}>{securityLabel(h.security)}</Badge>
        </button>
      ))}
      {hosts.length > 0 && (
        <button type="button" onClick={onOpen} className="text-[10px] text-primary hover:underline transition-colors">
          + Manage hosts
        </button>
      )}
    </div>
  );
}
