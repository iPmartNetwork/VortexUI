import { useMemo, useState } from "react";
import { motion } from "framer-motion";
import {
  Cpu,
  Copy,
  Globe,
  HardDrive,
  MemoryStick,
  Plus,
  Server,
  Signal,
} from "lucide-react";
import { useDeleteNode, useNodeDebugBundle, useNodes } from "@/api/hooks";
import { useRestartCore, useStopCore, useUpdateGeo } from "@/api/policy-hooks";
import type { Node } from "@/api/types";
import { Badge, Button } from "@/components/ui";
import { CreateNodeModal, diagColor, diagLabel, phaseLabel } from "@/components/CreateNodeModal";
import { EditNodeModal } from "@/components/EditNodeModal";
import { NodeInboundsModal } from "@/components/NodeInboundsModal";
import { NodeLogsModal } from "@/components/NodeLogsModal";
import { GlassCard, StatsCard, StatusBadge } from "@/components/veltrix";
import { useConfirm } from "@/components/confirm";
import { useToast } from "@/components/toast";
import { useI18n } from "@/i18n/i18n";
import { useTitle } from "@/lib/useTitle";
import { useAuth } from "@/auth/auth";
import { cn } from "@/lib/utils";

function MetricBar({ value, label, icon }: { value: number; label: string; icon: React.ReactNode }) {
  const v = Math.min(100, Math.max(0, value));
  return (
    <div className="space-y-1">
      <div className="flex items-center justify-between text-[10px]">
        <span className="flex items-center gap-1 text-fg-subtle">
          {icon} {label}
        </span>
        <span
          className={cn(
            "font-medium",
            v > 80 ? "text-danger" : v > 60 ? "text-warning" : "text-success",
          )}
        >
          {v.toFixed(0)}%
        </span>
      </div>
      <div className="h-1 rounded-full bg-surface-3 overflow-hidden">
        <div
          className={cn(
            "h-full rounded-full transition-all duration-500",
            v > 80 ? "bg-danger" : v > 60 ? "bg-warning" : "bg-success",
          )}
          style={{ width: `${v}%` }}
        />
      </div>
    </div>
  );
}

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

function isNodeOnline(n: Node): boolean {
  const lastSeenFresh =
    n.last_seen != null && Date.now() - new Date(n.last_seen).getTime() < 90_000;
  return lastSeenFresh && n.health.core_running;
}

function nodeDisplayStatus(n: Node): "active" | "warning" | "inactive" {
  if (!isNodeOnline(n)) return "inactive";
  const load = Math.max(n.health.cpu_percent ?? 0, n.health.mem_percent ?? 0);
  if (load > 75) return "warning";
  return "active";
}

function NodeCard({
  n,
  onInbounds,
  onLogs,
  onEdit,
  onDelete,
  onRestart,
  onStop,
  onStart,
  onUpdateGeo,
  onCopyDebug,
  canManage,
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
  const online = isNodeOnline(n);
  const status = nodeDisplayStatus(n);

  return (
    <GlassCard hover className="!p-0 overflow-hidden flex flex-col">
      <div className="p-5 flex-1">
        <div className="flex items-start justify-between gap-2 mb-4">
          <div className="flex items-center gap-3 min-w-0">
            <div className="grid h-10 w-10 place-items-center rounded-xl bg-surface-2/80 text-primary flex-shrink-0">
              <Server size={18} />
            </div>
            <div className="min-w-0">
              <h3 className="text-sm font-semibold text-fg truncate">{n.name}</h3>
              <div className="flex items-center gap-1 text-[10px] text-fg-subtle mt-0.5" dir="ltr">
                <Globe size={9} />
                <span className="truncate font-mono">{n.address}</span>
              </div>
            </div>
          </div>
          <StatusBadge status={status} label={online ? "online" : "offline"} pulse={online} />
        </div>

        <div className="space-y-2.5 mb-4">
          <MetricBar value={n.health.cpu_percent} label="CPU" icon={<Cpu size={9} />} />
          <MetricBar value={n.health.mem_percent} label="RAM" icon={<MemoryStick size={9} />} />
          <MetricBar value={n.health.disk_percent} label="Disk" icon={<HardDrive size={9} />} />
        </div>

        <div className="grid grid-cols-2 gap-2 pt-3 border-t border-border/40">
          <div className="text-center p-2 rounded-lg bg-surface/40">
            <div className="flex items-center justify-center gap-1 text-[10px] text-fg-subtle mb-0.5">
              <Signal size={9} /> Connections
            </div>
            <p className="text-xs font-semibold text-fg tabular-nums">{n.health.connections}</p>
          </div>
          <div className="text-center p-2 rounded-lg bg-surface/40">
            <div className="text-[10px] text-fg-subtle mb-0.5">Last seen</div>
            <p className="text-xs font-semibold text-fg">{timeAgoShort(n.last_seen)}</p>
          </div>
        </div>

        <div className="flex flex-wrap items-center gap-x-2 gap-y-1.5 mt-3 pt-3 border-t border-border/40 text-[10px]">
          <Badge color={online ? "running" : "down"}>{n.core}</Badge>
          {n.enrollment_phase && n.enrollment_phase !== "synced" && (
            <Badge color={n.enrollment_phase === "connected" ? "on_hold" : "muted"}>
              {phaseLabel(n.enrollment_phase)}
            </Badge>
          )}
          {!online && n.diagnostics && n.diagnostics.code !== "ok" && (
            <span title={n.diagnostics.message}>
              <Badge color={diagColor(n.diagnostics.code)}>{diagLabel(n.diagnostics.code)}</Badge>
            </span>
          )}
          {n.core_version && (
            <span className="rounded-md bg-surface-2/60 px-2 py-0.5 font-mono text-fg-muted">
              {n.core_version}
            </span>
          )}
          {n.agent_version && <span className="text-fg-subtle">agent {n.agent_version}</span>}
        </div>
      </div>

      {canManage && (
        <div className="flex flex-wrap items-center gap-1 border-t border-border/40 px-3 py-2.5 bg-surface/20">
          <Button variant="ghost" size="sm" onClick={onInbounds}>
            Inbounds
          </Button>
          <Button variant="ghost" size="sm" onClick={onLogs}>
            Logs
          </Button>
          {!online && (
            <Button variant="ghost" size="sm" onClick={onCopyDebug} title="Copy debug bundle for support">
              <Copy size={12} className="me-1" />
              Debug
            </Button>
          )}
          {online ? (
            <Button variant="ghost" size="sm" className="text-warning" onClick={onStop}>
              Stop
            </Button>
          ) : (
            <Button variant="ghost" size="sm" className="text-success" onClick={onStart}>
              Start
            </Button>
          )}
          <Button variant="ghost" size="sm" onClick={onRestart}>
            Restart
          </Button>
          <Button variant="ghost" size="sm" onClick={onUpdateGeo} title="Refresh Iran geoip/geosite routing data">
            Update Geo
          </Button>
          <div className="ms-auto flex gap-1">
            <Button variant="ghost" size="sm" onClick={onEdit}>
              Edit
            </Button>
            <Button variant="ghost" size="sm" className="text-danger" onClick={onDelete}>
              Delete
            </Button>
          </div>
        </div>
      )}
    </GlassCard>
  );
}

export function Nodes() {
  useTitle("Nodes");
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

  const nodes = data?.nodes ?? [];

  const fleetStats = useMemo(() => {
    let active = 0;
    let warning = 0;
    let offline = 0;
    let connections = 0;
    for (const n of nodes) {
      connections += n.health.connections ?? 0;
      const st = nodeDisplayStatus(n);
      if (st === "active") active++;
      else if (st === "warning") warning++;
      else offline++;
    }
    return { active, warning, offline, connections, total: nodes.length };
  }, [nodes]);

  async function remove(n: Node) {
    if (
      await confirm({
        title: `Delete node ${n.name}?`,
        message: "Its inbounds are removed and the agent is deregistered.",
        confirmLabel: "Delete",
        destructive: true,
      })
    ) {
      del
        .mutateAsync(n.id)
        .then(() => toast.success(`Deleted ${n.name}`))
        .catch(() => toast.error("Delete failed"));
    }
  }

  async function doStop(n: Node) {
    if (
      await confirm({
        title: `Stop core on ${n.name}?`,
        message: "The proxy engine will shut down. Users on this node will disconnect.",
        confirmLabel: "Stop",
        destructive: true,
      })
    ) {
      stop
        .mutateAsync(n.id)
        .then(() => toast.success("Core stopped"))
        .catch(() => toast.error("Stop failed"));
    }
  }

  async function doUpdateGeo(n: Node) {
    if (
      await confirm({
        title: `Update geo data on ${n.name}?`,
        message: "Downloads the latest Iran geoip/geosite databases and restarts the core (brief reconnect).",
        confirmLabel: "Update",
      })
    ) {
      toast.info("Updating geo data…");
      updateGeo
        .mutateAsync(n.id)
        .then((r) =>
          toast.success(`Geo updated (${Math.round((r.geoip_bytes + r.geosite_bytes) / 1024)} KB)`),
        )
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
    <div className="space-y-6 animate-page-enter">
      <CreateNodeModal open={createOpen} onClose={() => setCreateOpen(false)} />
      <EditNodeModal node={editing} onClose={() => setEditing(null)} />
      <NodeInboundsModal node={managing} onClose={() => setManaging(null)} />
      <NodeLogsModal node={logging} onClose={() => setLogging(null)} />

      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
        <div>
          <h1 className="text-xl font-bold text-fg">{t("nodes.title")}</h1>
          <p className="text-sm text-fg-muted mt-0.5">
            {fleetStats.total} {t("nodes.registered")}
          </p>
        </div>
        {canManage && (
          <Button onClick={() => setCreateOpen(true)}>
            <Plus size={15} /> {t("nodes.new")}
          </Button>
        )}
      </div>

      <div className="grid grid-cols-2 lg:grid-cols-4 gap-3">
        <StatsCard
          title={t("nodes.statOnline")}
          value={isLoading ? "—" : fleetStats.active}
          icon={<Server size={18} />}
          color="green"
          delay={0.05}
        />
        <StatsCard
          title={t("nodes.statHighLoad")}
          value={isLoading ? "—" : fleetStats.warning}
          icon={<Cpu size={18} />}
          color="orange"
          delay={0.1}
        />
        <StatsCard
          title={t("nodes.statOffline")}
          value={isLoading ? "—" : fleetStats.offline}
          icon={<Server size={18} />}
          color="red"
          delay={0.15}
        />
        <StatsCard
          title={t("nodes.statConnections")}
          value={isLoading ? "—" : fleetStats.connections}
          icon={<Signal size={18} />}
          color="cyan"
          delay={0.2}
        />
      </div>

      {isLoading && <div className="text-sm text-fg-muted">{t("common.loading")}</div>}

      <div className="grid grid-cols-1 gap-4 md:grid-cols-2 xl:grid-cols-3">
        {nodes.map((n, i) => (
          <motion.div
            key={n.id}
            initial={{ opacity: 0, y: 10 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ delay: i * 0.04 }}
          >
            <NodeCard
              n={n}
              onInbounds={() => setManaging(n)}
              onLogs={() => setLogging(n)}
              onEdit={() => setEditing(n)}
              onDelete={() => remove(n)}
              onRestart={() =>
                restart
                  .mutateAsync(n.id)
                  .then(() => toast.success("Core restarted"))
                  .catch(() => toast.error("Restart failed"))
              }
              onStart={() =>
                restart
                  .mutateAsync(n.id)
                  .then(() => toast.success("Core started"))
                  .catch(() => toast.error("Start failed"))
              }
              onStop={() => doStop(n)}
              onUpdateGeo={() => doUpdateGeo(n)}
              onCopyDebug={() => copyDebug(n)}
              canManage={canManage}
            />
          </motion.div>
        ))}
      </div>

      {!isLoading && nodes.length === 0 && (
        <GlassCard className="flex flex-col items-center gap-3 py-16 text-center">
          <Server size={32} className="text-fg-subtle" />
          <p className="text-sm text-fg-muted">
            {canManage
              ? t("nodes.none")
              : "No nodes assigned to your account yet — ask the main admin to allow nodes for you."}
          </p>
        </GlassCard>
      )}
    </div>
  );
}
