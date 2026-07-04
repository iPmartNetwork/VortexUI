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
  Users,
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

/** Renders a 2-letter ISO country code as a flag emoji via regional indicator
 * symbols; falls back to a globe icon when no code is set. */
function flagEmoji(code?: string): string {
  if (!code || code.length !== 2) return "🌐";
  const cc = code.toUpperCase();
  const points = [...cc].map((c) => 127397 + c.charCodeAt(0));
  return String.fromCodePoint(...points);
}

function metricTone(v: number): { bar: string; text: string } {
  if (v > 85) return { bar: "bg-danger", text: "text-danger" };
  if (v > 65) return { bar: "bg-warning", text: "text-warning" };
  return { bar: "bg-success", text: "text-success" };
}

function MetricBar({ icon, label, value }: { icon: React.ReactNode; label: string; value: number }) {
  const v = Math.min(100, Math.max(0, value || 0));
  const tone = metricTone(v);
  return (
    <div className="flex items-center gap-2">
      <span className="flex items-center gap-1 text-[10px] text-fg-subtle w-9 flex-shrink-0">
        {icon}
        {label}
      </span>
      <div className="flex-1 h-1.5 rounded-full bg-surface-3 overflow-hidden">
        <div
          className={cn("h-full rounded-full transition-all duration-500", tone.bar)}
          style={{ width: `${v}%` }}
        />
      </div>
      <span className={cn("text-[10px] font-bold tabular-nums w-8 text-end flex-shrink-0", tone.text)}>
        {v.toFixed(0)}%
      </span>
    </div>
  );
}

function SummaryStat({ value, label, color }: { value: number; label: string; color: string }) {
  return (
    <GlassCard hover={false} className="!p-3.5 text-center">
      <p className={cn("text-2xl font-black tabular-nums leading-none", color)}>{value}</p>
      <p className="text-[9px] font-bold uppercase tracking-widest text-fg-subtle mt-1.5">{label}</p>
    </GlassCard>
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

  const summary = useMemo(() => {
    let active = 0;
    let warning = 0;
    let offline = 0;
    let users = 0;
    for (const n of nodes) {
      const st = nodeDisplayStatus(n);
      if (st === "active") active++;
      else if (st === "warning") warning++;
      else offline++;
      users += n.users_count ?? 0;
    }
    return { active, warning, offline, users };
  }, [nodes]);

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

      {/* Fleet summary — active / warning / offline / total users at a glance */}
      <div className="grid grid-cols-2 sm:grid-cols-4 gap-3">
        <SummaryStat value={summary.active} label={t("nodes.filterOnline")} color="text-success" />
        <SummaryStat value={summary.warning} label={t("nodes.filterWarning")} color="text-warning" />
        <SummaryStat value={summary.offline} label={t("nodes.filterOffline")} color="text-danger" />
        <SummaryStat value={summary.users} label={t("nodes.users")} color="text-primary" />
      </div>

      <div className="flex flex-wrap items-center gap-1">
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
        <GlassCard hover={false} className="flex flex-col items-center gap-3 py-16 text-center">
          <Server size={32} className="text-fg-subtle" />
          <p className="text-sm text-fg-muted">
            {nodes.length === 0
              ? canManage
                ? t("nodes.none")
                : "No nodes assigned to your account yet — ask the main admin to allow nodes for you."
              : t("users.none")}
          </p>
        </GlassCard>
      )}

      {!isLoading && filteredNodes.length > 0 && (
        <div className="grid grid-cols-1 md:grid-cols-2 xl:grid-cols-3 gap-4">
          {filteredNodes.map((n) => {
            const online = isNodeOnline(n);
            const status = nodeDisplayStatus(n);
            const location = n.location || n.region || n.name;
            const isMenuOpen = menuNodeId === n.id;
            return (
              <GlassCard key={n.id} hover className="!p-4 flex flex-col gap-3.5">
                <div className="flex items-start justify-between gap-2">
                  <div className="flex items-center gap-2.5 min-w-0">
                    <span className="text-xl leading-none flex-shrink-0">{flagEmoji(n.country_code)}</span>
                    <div className="min-w-0">
                      <p className="font-semibold text-fg text-sm truncate">{n.name}</p>
                      <p className="text-[10px] text-fg-subtle truncate">{location}</p>
                    </div>
                  </div>
                  <div className="flex items-center gap-1 flex-shrink-0">
                    <StatusBadge status={status} label={statusLabel(status, online)} pulse={online && status === "active"} />
                    <div className="relative" ref={isMenuOpen ? menuRef : undefined}>
                      <button
                        type="button"
                        onClick={() => setMenuNodeId(isMenuOpen ? null : n.id)}
                        className="p-1.5 rounded-lg text-fg-muted hover:text-fg hover:bg-surface-2/80 transition"
                        aria-label={t("common.actions")}
                      >
                        <MoreVertical size={15} />
                      </button>
                      {isMenuOpen && (
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
                  </div>
                </div>

                {(n.core_version || n.enrollment_phase) && (
                  <div className="flex flex-wrap items-center gap-1 -mt-2">
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

                <div className="space-y-1.5">
                  <MetricBar icon={<Cpu size={10} />} label="CPU" value={n.health.cpu_percent} />
                  <MetricBar icon={<MemoryStick size={10} />} label="RAM" value={n.health.mem_percent} />
                  <MetricBar icon={<HardDrive size={10} />} label="Disk" value={n.health.disk_percent} />
                </div>

                <div className="grid grid-cols-2 gap-x-3 gap-y-2.5 pt-3 border-t border-border/50">
                  <div>
                    <p className="text-[9px] font-semibold uppercase tracking-wide text-fg-subtle flex items-center gap-1">
                      <Signal size={9} /> Latency
                    </p>
                    <p className="text-xs font-bold text-fg tabular-nums mt-0.5">
                      {n.ping_ms && n.ping_ms > 0 ? `${n.ping_ms}ms` : "—"}
                    </p>
                  </div>
                  <div>
                    <p className="text-[9px] font-semibold uppercase tracking-wide text-fg-subtle flex items-center gap-1">
                      <Users size={9} /> {t("nodes.users")}
                    </p>
                    <p className="text-xs font-bold text-fg tabular-nums mt-0.5">{n.users_count ?? 0}</p>
                  </div>
                  <div>
                    <p className="text-[9px] font-semibold uppercase tracking-wide text-fg-subtle">
                      {t("nodes.connections")}
                    </p>
                    <p className="text-xs font-bold text-fg tabular-nums mt-0.5">{n.health.connections}</p>
                  </div>
                  <div>
                    <p className="text-[9px] font-semibold uppercase tracking-wide text-fg-subtle">
                      {t("nodes.core")}
                    </p>
                    <div className="mt-0.5"><CoreBadge core={n.core} /></div>
                  </div>
                </div>

                <div className="flex items-center justify-between pt-3 border-t border-border/50 text-[10px] text-fg-subtle">
                  <span className="font-mono truncate" dir="ltr">{n.address}</span>
                  <span className="flex-shrink-0">
                    {online ? "· live" : `${t("nodes.lastSeen")} ${timeAgoShort(n.last_seen)}`}
                  </span>
                </div>
              </GlassCard>
            );
          })}
        </div>
      )}

      {!isLoading && filteredNodes.length > 0 && (
        <p className="text-xs text-fg-subtle text-center">
          {t("users.showingOf")
            .replace("{count}", String(filteredNodes.length))
            .replace("{total}", String(nodes.length))}
        </p>
      )}
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
