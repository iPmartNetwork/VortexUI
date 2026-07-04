import { useEffect, useMemo, useRef, useState } from "react";
import { useNavigate, useSearchParams } from "react-router-dom";
import {
  Copy,
  Cpu,
  HardDrive,
  MemoryStick,
  MoreVertical,
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
import { CoreBadge, GlassCard, StatusBadge } from "@/components/veltrix";
import { useConfirm } from "@/components/confirm";
import { useToast } from "@/components/toast";
import { useI18n } from "@/i18n/i18n";
import { useTitle } from "@/lib/useTitle";
import { useAuth } from "@/auth/auth";
import { cn } from "@/lib/utils";

type FleetFilter = "" | "online" | "warning" | "offline";

const FLEET_FILTERS: { value: FleetFilter; labelKey: "users.filterAll" | "nodes.filterOnline" | "nodes.filterWarning" | "nodes.filterOffline" }[] = [
  { value: "", labelKey: "users.filterAll" },
  { value: "online", labelKey: "nodes.filterOnline" },
  { value: "warning", labelKey: "nodes.filterWarning" },
  { value: "offline", labelKey: "nodes.filterOffline" },
];

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

function statusLabel(status: "active" | "warning" | "inactive", online: boolean): string {
  if (!online) return "OFFLINE";
  if (status === "warning") return "WARNING";
  return "ONLINE";
}

function nodeLoad(n: Node): number {
  return Math.max(n.health.cpu_percent ?? 0, n.health.mem_percent ?? 0);
}

function matchesFleetFilter(n: Node, filter: FleetFilter): boolean {
  const st = nodeDisplayStatus(n);
  switch (filter) {
    case "online":
      return st === "active";
    case "warning":
      return st === "warning";
    case "offline":
      return st === "inactive";
    default:
      return true;
  }
}

function LoadBar({ value }: { value: number }) {
  const v = Math.min(100, Math.max(0, value));
  return (
    <div className="space-y-1 min-w-[100px] max-w-[140px]">
      <div className="flex justify-between text-[10px]">
        <span className="text-fg-subtle">CPU / RAM</span>
        <span
          className={cn(
            "font-semibold tabular-nums",
            v > 75 ? "text-danger" : v > 50 ? "text-warning" : "text-success",
          )}
        >
          {v.toFixed(0)}%
        </span>
      </div>
      <div className="h-1.5 rounded-full bg-surface-3 overflow-hidden">
        <div
          className={cn(
            "h-full rounded-full transition-all duration-500",
            v > 75 ? "bg-danger" : v > 50 ? "bg-warning" : "bg-success",
          )}
          style={{ width: `${v}%` }}
        />
      </div>
    </div>
  );
}

export function Nodes() {
  useTitle("Nodes");
  const { can } = useAuth();
  const canManage = can("node:write");
  const navigate = useNavigate();
  const [searchParams] = useSearchParams();
  const [fleetFilter, setFleetFilter] = useState<FleetFilter>("");
  const [menuNodeId, setMenuNodeId] = useState<string | null>(null);
  const menuRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    if (searchParams.get("tab") === "inbounds") {
      navigate("/inbounds", { replace: true });
    }
  }, [searchParams, navigate]);

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

  const filteredNodes = useMemo(
    () => nodes.filter((n) => matchesFleetFilter(n, fleetFilter)),
    [nodes, fleetFilter],
  );

  useEffect(() => {
    function onDocClick(e: MouseEvent) {
      if (menuRef.current && !menuRef.current.contains(e.target as globalThis.Node)) {
        setMenuNodeId(null);
      }
    }
    document.addEventListener("mousedown", onDocClick);
    return () => document.removeEventListener("mousedown", onDocClick);
  }, []);

  async function remove(n: Node) {
    setMenuNodeId(null);
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
    setMenuNodeId(null);
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
    setMenuNodeId(null);
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
    setMenuNodeId(null);
    try {
      const res = await debug.mutateAsync(n.id);
      await navigator.clipboard.writeText(res.debug_text);
      toast.success("Debug bundle copied");
    } catch {
      toast.error("Could not copy debug bundle");
    }
  }

  return (
    <div className="space-y-5 animate-page-enter">
      <CreateNodeModal open={createOpen} onClose={() => setCreateOpen(false)} />
      <EditNodeModal node={editing} onClose={() => setEditing(null)} />
      <NodeInboundsModal node={managing} onClose={() => setManaging(null)} />
      <NodeLogsModal node={logging} onClose={() => setLogging(null)} />

      <div className="flex flex-col sm:flex-row sm:items-start justify-between gap-4">
        <div>
          <h1 className="text-2xl font-bold text-fg tracking-tight">{t("nodes.managementTitle")}</h1>
          <p className="text-sm text-fg-muted mt-1">
            {nodes.length} {t("nodes.registeredNodes")}
          </p>
        </div>
        {canManage && (
          <Button onClick={() => setCreateOpen(true)}>
            <Plus size={15} /> {t("nodes.new")}
          </Button>
        )}
      </div>

      <GlassCard hover={false} className="!p-0 overflow-hidden">
          <div className="flex flex-wrap items-center gap-1 p-4 border-b border-border/40">
            {FLEET_FILTERS.map((f) => (
              <button
                key={f.value || "all"}
                type="button"
                onClick={() => setFleetFilter(f.value)}
                className={cn(
                  "px-3.5 py-1.5 rounded-lg text-xs font-medium transition-all",
                  fleetFilter === f.value
                    ? "bg-primary text-primary-fg shadow-sm"
                    : "text-fg-muted hover:text-fg hover:bg-surface/60",
                )}
              >
                {t(f.labelKey)}
              </button>
            ))}
          </div>

          {isLoading && (
            <div className="p-8 text-sm text-fg-muted text-center">{t("common.loading")}</div>
          )}

          {!isLoading && filteredNodes.length === 0 && (
            <div className="flex flex-col items-center gap-3 py-16 text-center">
              <Server size={32} className="text-fg-subtle" />
              <p className="text-sm text-fg-muted">
                {nodes.length === 0
                  ? canManage
                    ? t("nodes.none")
                    : "No nodes assigned to your account yet — ask the main admin to allow nodes for you."
                  : t("users.none")}
              </p>
            </div>
          )}

          {!isLoading && filteredNodes.length > 0 && (
            <div className="overflow-x-auto">
              <table className="w-full text-sm">
                <thead>
                  <tr className="border-b border-border/40 bg-surface/30">
                    <th className="text-start py-3 px-4 text-xs font-semibold text-fg-subtle uppercase tracking-wide">
                      {t("nodes.title")}
                    </th>
                    <th className="text-start py-3 px-4 text-xs font-semibold text-fg-subtle uppercase tracking-wide hidden md:table-cell">
                      {t("nodes.core")}
                    </th>
                    <th className="text-start py-3 px-4 text-xs font-semibold text-fg-subtle uppercase tracking-wide hidden lg:table-cell">
                      {t("nodes.location")}
                    </th>
                    <th className="text-start py-3 px-4 text-xs font-semibold text-fg-subtle uppercase tracking-wide">
                      {t("nodes.load")}
                    </th>
                    <th className="text-start py-3 px-4 text-xs font-semibold text-fg-subtle uppercase tracking-wide hidden sm:table-cell">
                      {t("nodes.users")}
                    </th>
                    <th className="text-start py-3 px-4 text-xs font-semibold text-fg-subtle uppercase tracking-wide hidden sm:table-cell">
                      {t("nodes.connections")}
                    </th>
                    <th className="text-start py-3 px-4 text-xs font-semibold text-fg-subtle uppercase tracking-wide">
                      {t("common.status")}
                    </th>
                    <th className="text-end py-3 px-4 text-xs font-semibold text-fg-subtle uppercase tracking-wide w-12">
                      {t("common.actions")}
                    </th>
                  </tr>
                </thead>
                <tbody>
                  {filteredNodes.map((n) => {
                    const online = isNodeOnline(n);
                    const status = nodeDisplayStatus(n);
                    const load = nodeLoad(n);
                    const location = n.location || n.region || n.name;
                    return (
                      <tr
                        key={n.id}
                        className="border-b border-border/20 hover:bg-surface/40 transition-colors"
                      >
                        <td className="py-3.5 px-4">
                          <div className="flex items-center gap-3 min-w-[160px]">
                            <div className="h-9 w-9 rounded-full bg-primary/10 flex items-center justify-center text-primary flex-shrink-0">
                              <Server size={16} />
                            </div>
                            <div className="min-w-0">
                              <p className="font-semibold text-fg text-sm truncate">{n.name}</p>
                              <p className="text-[11px] text-fg-subtle font-mono truncate" dir="ltr">
                                {n.address}
                              </p>
                              {(n.core_version || n.enrollment_phase) && (
                                <div className="flex flex-wrap items-center gap-1 mt-1">
                                  {n.enrollment_phase && n.enrollment_phase !== "synced" && (
                                    <Badge color={n.enrollment_phase === "connected" ? "on_hold" : "muted"}>
                                      {phaseLabel(n.enrollment_phase)}
                                    </Badge>
                                  )}
                                  {!online && n.diagnostics && n.diagnostics.code !== "ok" && (
                                    <Badge color={diagColor(n.diagnostics.code)}>
                                      {diagLabel(n.diagnostics.code)}
                                    </Badge>
                                  )}
                                </div>
                              )}
                            </div>
                          </div>
                        </td>
                        <td className="py-3.5 px-4 hidden md:table-cell">
                          <CoreBadge core={n.core} />
                          {n.core_version && (
                            <p className="text-[10px] text-fg-subtle mt-1 font-mono truncate max-w-[120px]" title={n.core_version}>
                              {n.core_version.split(" ")[0]}
                            </p>
                          )}
                        </td>
                        <td className="py-3.5 px-4 hidden lg:table-cell">
                          <div className="text-xs text-fg-muted">
                            <p className="font-medium text-fg">{location}</p>
                            <p className="text-[10px] text-fg-subtle mt-0.5">
                              {n.ping_ms && n.ping_ms > 0 ? `${n.ping_ms}ms` : "—"}
                              <span className="mx-1">·</span>
                              {t("nodes.lastSeen")} {timeAgoShort(n.last_seen)}
                            </p>
                          </div>
                        </td>
                        <td className="py-3.5 px-4">
                          <LoadBar value={load} />
                          <p className="text-[10px] text-fg-subtle mt-1 flex items-center gap-2">
                            <span className="inline-flex items-center gap-0.5">
                              <Cpu size={9} /> {n.health.cpu_percent.toFixed(0)}%
                            </span>
                            <span className="inline-flex items-center gap-0.5">
                              <MemoryStick size={9} /> {n.health.mem_percent.toFixed(0)}%
                            </span>
                            <span className="inline-flex items-center gap-0.5 hidden xl:inline-flex">
                              <HardDrive size={9} /> {n.health.disk_percent.toFixed(0)}%
                            </span>
                          </p>
                        </td>
                        <td className="py-3.5 px-4 hidden sm:table-cell tabular-nums text-fg-muted">
                          {n.users_count ?? 0}
                        </td>
                        <td className="py-3.5 px-4 hidden sm:table-cell tabular-nums text-fg-muted">
                          <span className="inline-flex items-center gap-1">
                            <Signal size={12} className="text-fg-subtle" />
                            {n.health.connections}
                          </span>
                        </td>
                        <td className="py-3.5 px-4">
                          <StatusBadge
                            status={status}
                            label={statusLabel(status, online)}
                            pulse={online && status === "active"}
                          />
                        </td>
                        <td className="py-3.5 px-4 text-end">
                          <div className="relative inline-block" ref={menuNodeId === n.id ? menuRef : undefined}>
                            <button
                              type="button"
                              onClick={() => setMenuNodeId(menuNodeId === n.id ? null : n.id)}
                              className="p-1.5 rounded-lg text-fg-muted hover:text-fg hover:bg-surface-2/80 transition"
                              aria-label={t("common.actions")}
                            >
                              <MoreVertical size={16} />
                            </button>
                            {menuNodeId === n.id && (
                              <div className="absolute end-0 top-full mt-1 z-20 min-w-[170px] rounded-lg border border-border/60 bg-surface shadow-lg py-1 text-xs">
                                <MenuAction onClick={() => { setManaging(n); setMenuNodeId(null); }}>
                                  {t("nodes.inbounds")}
                                </MenuAction>
                                <MenuAction onClick={() => { setLogging(n); setMenuNodeId(null); }}>
                                  Logs
                                </MenuAction>
                                {!online && (
                                  <MenuAction onClick={() => copyDebug(n)}>
                                    <Copy size={13} className="inline me-1" />
                                    Debug
                                  </MenuAction>
                                )}
                                {canManage && (
                                  <>
                                    {online ? (
                                      <MenuAction className="text-warning" onClick={() => doStop(n)}>
                                        Stop
                                      </MenuAction>
                                    ) : (
                                      <MenuAction
                                        className="text-success"
                                        onClick={() => {
                                          setMenuNodeId(null);
                                          restart
                                            .mutateAsync(n.id)
                                            .then(() => toast.success("Core started"))
                                            .catch(() => toast.error("Start failed"));
                                        }}
                                      >
                                        Start
                                      </MenuAction>
                                    )}
                                    <MenuAction
                                      onClick={() => {
                                        setMenuNodeId(null);
                                        restart
                                          .mutateAsync(n.id)
                                          .then(() => toast.success("Core restarted"))
                                          .catch(() => toast.error("Restart failed"));
                                      }}
                                    >
                                      Restart
                                    </MenuAction>
                                    <MenuAction onClick={() => doUpdateGeo(n)}>Update Geo</MenuAction>
                                    <MenuAction onClick={() => { setEditing(n); setMenuNodeId(null); }}>
                                      {t("common.edit")}
                                    </MenuAction>
                                    <MenuAction className="text-danger hover:bg-danger/10" onClick={() => remove(n)}>
                                      {t("common.delete")}
                                    </MenuAction>
                                  </>
                                )}
                              </div>
                            )}
                          </div>
                        </td>
                      </tr>
                    );
                  })}
                </tbody>
              </table>
            </div>
          )}

          {!isLoading && filteredNodes.length > 0 && (
            <div className="border-t border-border/40 px-4 py-3 text-sm text-fg-muted">
              {t("users.showingOf")
                .replace("{count}", String(filteredNodes.length))
                .replace("{total}", String(nodes.length))}
            </div>
          )}
        </GlassCard>
    </div>
  );
}

function MenuAction({
  children,
  onClick,
  className,
}: {
  children: React.ReactNode;
  onClick: () => void;
  className?: string;
}) {
  return (
    <button
      type="button"
      onClick={onClick}
      className={cn(
        "w-full px-3 py-2 text-start text-fg hover:bg-surface-2/60 transition",
        className,
      )}
    >
      {children}
    </button>
  );
}
