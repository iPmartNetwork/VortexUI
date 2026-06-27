import { useState } from "react";
import { Cpu, Copy, Globe, HardDrive, MemoryStick, Server, Signal } from "lucide-react";
import { useDeleteNode, useNodeDebugBundle, useNodes } from "@/api/hooks";
import { useRestartCore, useStopCore, useUpdateGeo } from "@/api/policy-hooks";
import type { Node } from "@/api/types";
import { Badge, Button, Card, PageHeader } from "@/components/ui";
import { CreateNodeModal, diagColor, diagLabel } from "@/components/CreateNodeModal";
import { EditNodeModal } from "@/components/EditNodeModal";
import { NodeInboundsModal } from "@/components/NodeInboundsModal";
import { NodeLogsModal } from "@/components/NodeLogsModal";
import { useConfirm } from "@/components/confirm";
import { useToast } from "@/components/toast";
import { useI18n } from "@/i18n/i18n";
import { useAuth } from "@/auth/auth";
import { cn } from "@/lib/utils";

/* ─── Gauge ─── */
function Gauge({ label, value, icon }: { label: string; value: number; icon: React.ReactNode }) {
  const v = Math.min(100, Math.max(0, value));
  const bar = v > 85 ? "from-danger to-danger/60" : v > 60 ? "from-warning to-warning/60" : "from-accent to-primary/50";
  return (
    <div className="group">
      <div className="mb-1 flex items-center justify-between">
        <span className="flex items-center gap-1.5 text-[11px] font-medium text-fg-subtle group-hover:text-fg-muted transition">{icon}{label}</span>
        <span className="text-[11px] font-bold tabular-nums text-fg">{v.toFixed(0)}%</span>
      </div>
      <div className="h-[5px] rounded-full bg-border/40 dark:bg-surface-2/80">
        <div className={cn("h-full rounded-full bg-gradient-to-r transition-all duration-1000 ease-out", bar)} style={{ width: `${v}%` }} />
      </div>
    </div>
  );
}

/* ─── Time ago helper ─── */
function timeAgoShort(iso: string | null): string {
  if (!iso) return "—";
  const diff = Date.now() - new Date(iso).getTime();
  const sec = Math.floor(diff / 1000);
  if (sec < 0) return "now";
  if (sec < 60) return `${sec}s`;
  const min = Math.floor(sec / 60);
  if (min < 60) return `${min}m`;
  const hrs = Math.floor(min / 60);
  if (hrs < 24) return `${hrs}h`;
  return `${Math.floor(hrs / 24)}d`;
}

/* ─── Node Card (rich + actions) ─── */
function NodeCard({
  n, onInbounds, onLogs, onEdit, onDelete, onRestart, onStop, onStart, onUpdateGeo, onCopyDebug, canManage,
}: {
  n: Node;
  onInbounds: () => void;
  onLogs: () => void;
  onEdit: () => void;
  onDelete: () => void;
  onRestart: () => void;
  onStop: () => void;
  onStart: () => void;
  onUpdateGeo: () => void;
  onCopyDebug: () => void;
  canManage: boolean;
}) {
  // Online = fresh heartbeat AND core running, matching the Overview's NODES
  // ONLINE count. A node the panel has never reached (no last_seen) stays
  // offline even if a stale health snapshot is present.
  const lastSeenFresh = n.last_seen != null && Date.now() - new Date(n.last_seen).getTime() < 90_000;
  const online = lastSeenFresh && n.health.core_running;
  return (
    <Card className="group relative overflow-hidden p-0 transition-all duration-300 hover:ring-1 hover:ring-primary/20 hover:shadow-lg">
      {/* Content area */}
      <div className="p-5">
        {/* Online dot */}
        <div className={cn("absolute end-4 top-4 h-2.5 w-2.5 rounded-full", online ? "bg-success shadow-[0_0_8px_2px] shadow-success/50" : "bg-fg-subtle/40")} />

        {/* Header */}
        <div className="flex items-center gap-3">
          <div className="grid h-11 w-11 place-items-center rounded-xl bg-surface-2/60 text-fg-muted transition group-hover:bg-primary/10 group-hover:text-primary">
            <Server size={18} />
          </div>
          <div className="min-w-0 flex-1">
            <span className="text-[15px] font-bold text-fg">{n.name}</span>
            <div className="mt-0.5 flex items-center gap-1.5 text-xs text-fg-subtle" dir="ltr">
              <Globe size={11} />
              <span className="truncate font-mono">{n.address}</span>
            </div>
          </div>
        </div>

        {/* Gauges */}
        <div className="mt-5 space-y-2.5">
          <Gauge label="CPU" value={n.health.cpu_percent} icon={<Cpu size={11} />} />
          <Gauge label="RAM" value={n.health.mem_percent} icon={<MemoryStick size={11} />} />
          <Gauge label="Disk" value={n.health.disk_percent} icon={<HardDrive size={11} />} />
        </div>

        {/* Stats: connections + last seen */}
        <div className="mt-5 grid grid-cols-2 gap-3 text-center">
          <div className="rounded-xl bg-surface-2/40 py-2.5">
            <div className="text-xl font-bold text-fg">{n.health.connections}</div>
            <div className="text-[10px] font-medium text-fg-subtle">Connections</div>
          </div>
          <div className="rounded-xl bg-surface-2/40 py-2.5">
            <div className="text-base font-bold text-fg">{timeAgoShort(n.last_seen)}</div>
            <div className="text-[10px] font-medium text-fg-subtle">Last seen</div>
          </div>
        </div>

        {/* Version footer */}
        <div className="mt-4 flex flex-wrap items-center gap-x-3 gap-y-1.5 text-[10px]">
          <Badge color={online ? "running" : "down"}>{n.core}</Badge>
          {!online && n.diagnostics && n.diagnostics.code !== "ok" && (
            <span title={n.diagnostics.message}>
              <Badge color={diagColor(n.diagnostics.code)}>
                {diagLabel(n.diagnostics.code)}
              </Badge>
            </span>
          )}
          {n.core_version && <span className="rounded-md bg-surface-2/60 px-2 py-0.5 font-mono text-fg-muted">{n.core_version}</span>}
          {n.agent_version && <span className="text-fg-subtle">agent {n.agent_version}</span>}
          <span className="ms-auto flex items-center gap-1 text-[11px] font-semibold text-fg">
            <Signal size={10} className="text-accent" />{n.health.connections}
          </span>
        </div>
      </div>

      {/* Action buttons — read-only resellers see stats only */}
      {canManage && (
      <div className="flex flex-wrap items-center gap-1 border-t border-border/40 px-3 py-2.5">
        <Button variant="ghost" size="sm" onClick={onInbounds}>Inbounds</Button>
        <Button variant="ghost" size="sm" onClick={onLogs}>Logs</Button>
        {!online && (
          <Button variant="ghost" size="sm" onClick={onCopyDebug} title="Copy debug bundle for support">
            <Copy size={12} className="me-1" />Debug
          </Button>
        )}
        {online ? (
          <Button variant="ghost" size="sm" className="text-warning" onClick={onStop}>Stop</Button>
        ) : (
          <Button variant="ghost" size="sm" className="text-success" onClick={onStart}>Start</Button>
        )}
        <Button variant="ghost" size="sm" onClick={onRestart}>Restart</Button>
        <Button variant="ghost" size="sm" onClick={onUpdateGeo} title="Refresh Iran geoip/geosite routing data">Update Geo</Button>
        <div className="ms-auto flex gap-1">
          <Button variant="ghost" size="sm" onClick={onEdit}>Edit</Button>
          <Button variant="ghost" size="sm" className="text-danger" onClick={onDelete}>Delete</Button>
        </div>
      </div>
      )}
    </Card>
  );
}

/* ═══════ Nodes Page ═══════ */
export function Nodes() {
  const { can } = useAuth();
  const canManage = can("node:write");
  const { data, isLoading } = useNodes();
  const del = useDeleteNode();
  const restart = useRestartCore();
  const stop = useStopCore();
  const updateGeo = useUpdateGeo();
  const debug = useNodeDebugBundle();
  const confirm = useConfirm();
  const toast = useToast();
  const { t } = useI18n();
  const [createOpen, setCreateOpen] = useState(false);
  const [editing, setEditing] = useState<Node | null>(null);
  const [managing, setManaging] = useState<Node | null>(null);
  const [logging, setLogging] = useState<Node | null>(null);

  async function remove(n: Node) {
    if (await confirm({ title: `Delete node ${n.name}?`, message: "Its inbounds are removed and the agent is deregistered.", confirmLabel: "Delete", destructive: true }))
      del.mutateAsync(n.id).then(() => toast.success(`Deleted ${n.name}`)).catch(() => toast.error("Delete failed"));
  }

  async function doStop(n: Node) {
    if (await confirm({ title: `Stop core on ${n.name}?`, message: "The proxy engine will shut down. Users on this node will disconnect.", confirmLabel: "Stop", destructive: true }))
      stop.mutateAsync(n.id).then(() => toast.success("Core stopped")).catch(() => toast.error("Stop failed"));
  }

  async function doUpdateGeo(n: Node) {
    if (await confirm({ title: `Update geo data on ${n.name}?`, message: "Downloads the latest Iran geoip/geosite databases and restarts the core (brief reconnect).", confirmLabel: "Update" })) {
      toast.info("Updating geo data…");
      updateGeo.mutateAsync(n.id)
        .then((r) => toast.success(`Geo updated (${Math.round((r.geoip_bytes + r.geosite_bytes) / 1024)} KB)`))
        .catch(() => toast.error("Geo update failed"));
    }
  }

  async function copyDebug(n: Node) {
    try {
      const res = await debug.mutateAsync(n.id);
      await navigator.clipboard.writeText(res.debug_text);
      toast.success("Debug bundle copied");
    } catch {
      toast.error("Could not copy debug bundle");
    }
  }

  return (
    <div className="space-y-6">
      <CreateNodeModal open={createOpen} onClose={() => setCreateOpen(false)} />
      <EditNodeModal node={editing} onClose={() => setEditing(null)} />
      <NodeInboundsModal node={managing} onClose={() => setManaging(null)} />
      <NodeLogsModal node={logging} onClose={() => setLogging(null)} />

      <PageHeader title={t("nodes.title")} subtitle={`${data?.nodes.length ?? 0} registered`}>
        {canManage && <Button onClick={() => setCreateOpen(true)}>{t("nodes.new")}</Button>}
      </PageHeader>

      {isLoading && <div className="text-sm text-fg-muted">{t("common.loading")}</div>}

      <div className="grid grid-cols-1 gap-5 md:grid-cols-2 xl:grid-cols-3">
        {data?.nodes.map((n) => (
          <NodeCard
            key={n.id}
            n={n}
            onInbounds={() => setManaging(n)}
            onLogs={() => setLogging(n)}
            onEdit={() => setEditing(n)}
            onDelete={() => remove(n)}
            onRestart={() => restart.mutateAsync(n.id).then(() => toast.success("Core restarted")).catch(() => toast.error("Restart failed"))}
            onStart={() => restart.mutateAsync(n.id).then(() => toast.success("Core started")).catch(() => toast.error("Start failed"))}
            onStop={() => doStop(n)}
            onUpdateGeo={() => doUpdateGeo(n)}
            onCopyDebug={() => copyDebug(n)}
            canManage={canManage}
          />
        ))}
      </div>

      {data?.nodes.length === 0 && (
        <Card className="flex flex-col items-center gap-3 py-16 text-center">
          <Server size={32} className="text-fg-subtle" />
          <p className="text-sm text-fg-muted">
            {canManage ? t("nodes.none") : "No nodes assigned to your account yet — ask the main admin to allow nodes for you."}
          </p>
        </Card>
      )}
    </div>
  );
}
