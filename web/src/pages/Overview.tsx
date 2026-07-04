import {
  Users, Wifi, Activity, Cpu, MemoryStick,
  Zap, Clock, TrendingUp, MonitorSmartphone, Layers, Timer, Box,
  Power, RotateCcw, Tag, Gauge, ChevronRight, Server, ArrowUpRight,
} from "lucide-react";
import { Link } from "react-router-dom";
import { motion } from "framer-motion";
import { useOverview, useSystem, useTrafficSamples, useTrafficSeries, useRestartCore, useStopCore } from "@/api/policy-hooks";
import { useAccountQuota } from "@/api/quota-hooks";
import { useAllInbounds, useNodes, useUsers, useVersion } from "@/api/hooks";
import { useAuth } from "@/auth/auth";
import { Card } from "@/components/ui";
import { TrafficSeriesChart } from "@/components/TrafficSeriesChart";
import { GlassCard, StatsCard, StatusBadge } from "@/components/veltrix";
import { useI18n } from "@/i18n/i18n";
import { useTitle } from "@/lib/useTitle";
import { cn, formatBytes } from "@/lib/utils";
import type { Overview as OverviewData } from "@/api/types";

/* ═══════ Mini Sparkline (SVG) ═══════ */
function Sparkline({ data, className }: { data: number[]; className?: string }) {
  if (data.length < 2) return null;
  const max = Math.max(...data, 1);
  const w = 100, h = 28;
  const pts = data.map((v, i) => `${(i / (data.length - 1)) * w},${h - (v / max) * h}`).join(" ");
  return (
    <svg viewBox={`0 0 ${w} ${h}`} className={cn("w-full", className)} preserveAspectRatio="none">
      <defs>
        <linearGradient id="sparkGrad" x1="0" y1="0" x2="0" y2="1">
          <stop offset="0%" stopColor="hsl(var(--accent))" stopOpacity="0.4" />
          <stop offset="100%" stopColor="hsl(var(--accent))" stopOpacity="0" />
        </linearGradient>
      </defs>
      <polygon points={`0,${h} ${pts} ${w},${h}`} fill="url(#sparkGrad)" />
      <polyline points={pts} fill="none" stroke="hsl(var(--accent))" strokeWidth="1.5" strokeLinejoin="round" strokeLinecap="round" />
    </svg>
  );
}

/* ═══════ Status colors ═══════ */
const STATUS_META: Record<string, { color: string; label: string }> = {
  active: { color: "bg-success", label: "Active" },
  limited: { color: "bg-warning", label: "Limited" },
  expired: { color: "bg-danger", label: "Expired" },
  disabled: { color: "bg-fg-subtle", label: "Disabled" },
  on_hold: { color: "bg-accent", label: "On Hold" },
};

function fmtUptime(sec: number): string {
  const d = Math.floor(sec / 86400);
  const h = Math.floor((sec % 86400) / 3600);
  const m = Math.floor((sec % 3600) / 60);
  if (d > 0) return `${d}d ${h}h ${m}m`;
  if (h > 0) return `${h}h ${m}m`;
  return `${m}m`;
}

function daysUntil(iso: string | null): string {
  if (!iso) return "∞";
  const d = Math.ceil((new Date(iso).getTime() - Date.now()) / 86400000);
  if (d < 0) return "expired";
  return `${d}d`;
}

function nodeFleetStatus(item: OverviewData["nodes"]["items"][number]): "active" | "warning" | "inactive" {
  if (!item.online || !item.health?.core_running) return "inactive";
  const load = Math.max(item.health.cpu_percent ?? 0, item.health.mem_percent ?? 0);
  if (load > 75) return "warning";
  return "active";
}

/* ═══════ Core Engine Card ═══════ */
function CoreCard({ name, version, running, onStop, onRestart }: {
  name: string; version: string; running: boolean;
  onStop: () => void; onRestart: () => void;
}) {
  return (
    <Card className="flex items-center justify-between p-4">
      <div className="flex items-center gap-3">
        <div className={cn("h-3 w-3 rounded-full", running ? "bg-success shadow-[0_0_6px_1px] shadow-success/50" : "bg-fg-subtle/50")} />
        <span className="text-sm font-bold text-fg">{name}</span>
      </div>
      <div className="flex items-center gap-3">
        <button onClick={onStop} className="flex items-center gap-1.5 rounded-lg px-2.5 py-1.5 text-xs font-medium text-fg-muted transition hover:bg-surface-2/60 hover:text-danger" title="Stop">
          <Power size={13} /> Stop
        </button>
        <button onClick={onRestart} className="flex items-center gap-1.5 rounded-lg px-2.5 py-1.5 text-xs font-medium text-fg-muted transition hover:bg-surface-2/60 hover:text-accent" title="Restart">
          <RotateCcw size={13} /> Restart
        </button>
        {version && (
          <span className="flex items-center gap-1 rounded-md bg-surface-2/50 px-2 py-1 text-[11px] font-mono text-fg-muted">
            <Tag size={11} /> {version}
          </span>
        )}
      </div>
    </Card>
  );
}

/* ═══════════════════════════════════════════════════════════════════════
   OVERVIEW PAGE
   ═══════════════════════════════════════════════════════════════════════ */
export function Overview() {
  useTitle("Overview");
  const { sudo } = useAuth();
  const accountQuota = useAccountQuota();
  const { data, dataUpdatedAt, isLoading: overviewLoading } = useOverview();
  const sys = useSystem();
  const inbounds = useAllInbounds();
  const nodesQ = useNodes();
  const recentUsersQ = useUsers({ limit: 20, status: "active" });
  const panelVersion = useVersion().data;
  const { t } = useI18n();

  const u = data?.users;
  const onlineCount = data?.nodes.online ?? 0;
  const totalNodes = data?.nodes.total ?? 0;
  const byStatus = u?.by_status ?? {};
  const totalUsers = u?.total ?? 0;
  const totalUsed = u?.total_used ?? 0;
  const trafficSamples = useTrafficSamples(totalUsed);
  const trafficSeries = useTrafficSeries();

  const s = sys.data;
  const inboundCount = inbounds.data?.length ?? 0;
  const fleetItems = data?.nodes.items ?? [];
  const totalConnections = fleetItems.reduce((sum, n) => sum + (n.health?.connections ?? 0), 0);
  const trafficPoints = trafficSeries.data?.points ?? [];
  const peakBucket = trafficPoints.length
    ? Math.max(...trafficPoints.map((p) => p.up + p.down))
    : 0;

  const nodesList = nodesQ.data?.nodes ?? [];
  const xrayNode = nodesList.find((n) => n.core === "xray");
  const singboxNode = nodesList.find((n) => n.core === "singbox");
  const xrayVer = xrayNode?.core_version || "—";
  const singboxVer = singboxNode?.core_version || "—";
  const xrayRunning = xrayNode?.health.core_running ?? false;
  const singboxRunning = singboxNode?.health.core_running ?? false;
  const restartCore = useRestartCore();
  const stopCore = useStopCore();

  const topUsers = [...(recentUsersQ.data?.users ?? [])]
    .sort((a, b) => b.used_traffic - a.used_traffic)
    .slice(0, 5);

  const coreLabel = [xrayVer !== "—" ? `Xray ${xrayVer}` : null, singboxVer !== "—" ? `sing-box ${singboxVer}` : null]
    .filter(Boolean)
    .join(" · ");

  return (
    <div className="space-y-6 animate-page-enter">
      <motion.div
        initial={{ opacity: 0, y: 10 }}
        animate={{ opacity: 1, y: 0 }}
        className="relative overflow-hidden rounded-3xl bg-gradient-to-r from-bg-elevated via-surface to-bg-elevated border border-border/80 p-6 md:p-8 shadow-xl"
      >
        <div className="absolute top-0 end-0 w-80 h-80 rounded-full bg-primary/10 blur-3xl pointer-events-none" />
        <div className="absolute bottom-0 start-1/4 w-64 h-64 rounded-full bg-accent/10 blur-3xl pointer-events-none" />
        <div className="absolute inset-0 bg-grid-pattern opacity-40 pointer-events-none" />

        <div className="relative z-10 flex flex-col lg:flex-row lg:items-center justify-between gap-6">
          <div className="space-y-2 max-w-2xl">
            <div className="flex flex-wrap items-center gap-2">
              <StatusBadge
                status={onlineCount === totalNodes && totalNodes > 0 ? "optimal" : onlineCount > 0 ? "warning" : "inactive"}
                label={
                  totalNodes === 0
                    ? t("overview.noNodes")
                    : `${onlineCount}/${totalNodes} ${t("overview.online")}`
                }
              />
              {panelVersion && (
                <span className="px-2.5 py-0.5 rounded-full bg-primary/15 text-primary border border-primary/30 text-[10px] font-semibold">
                  v{panelVersion}
                </span>
              )}
              {coreLabel && (
                <span className="px-2.5 py-0.5 rounded-full bg-surface-2/80 text-fg-muted border border-border/60 text-[10px] font-semibold truncate max-w-xs">
                  {coreLabel}
                </span>
              )}
            </div>
            <h1 className="text-2xl md:text-3xl font-black text-fg tracking-tight">{t("nav.overview")}</h1>
            <p className="text-xs md:text-sm text-fg-muted leading-relaxed">
              {overviewLoading || !s ? (
                t("overview.loadingTelemetry")
              ) : (
                <>
                  {s.hostname} · {t("overview.uptime")} {fmtUptime(s.uptime_seconds)} ·{" "}
                  {totalConnections} {t("overview.liveConnections")}
                </>
              )}
            </p>
          </div>

          <div className="flex items-center gap-2 rounded-full bg-surface/60 px-3 py-1.5 text-[11px] text-fg-subtle ring-1 ring-border/50 flex-shrink-0">
            <span className="relative flex h-2 w-2">
              <span className="absolute inline-flex h-full w-full animate-ping rounded-full bg-success/60" />
              <span className="inline-flex h-2 w-2 rounded-full bg-success" />
            </span>
            {t("overview.live")}
            {dataUpdatedAt > 0 && (
              <span className="flex items-center gap-1">
                <Clock size={10} />
                {new Date(dataUpdatedAt).toLocaleTimeString()}
              </span>
            )}
          </div>
        </div>
      </motion.div>

      {!sudo && accountQuota.data?.usage && (
        <Card className="p-4">
          <div className="mb-3 flex flex-wrap items-center justify-between gap-2">
            <div className="text-sm font-semibold">{t("reseller.overview.pool")}</div>
            <Link to="/reseller-dashboard" className="inline-flex items-center gap-1 text-xs font-medium text-primary hover:underline">
              {t("reseller.overview.viewDashboard")} <ChevronRight size={14} />
            </Link>
          </div>
          <div className="grid gap-3 sm:grid-cols-3">
            <div className="flex items-center gap-2 rounded-lg bg-muted/40 px-3 py-2 text-sm">
              <Users size={16} className="text-primary" />
              <span>{t("reseller.dashboard.accounts")}: {accountQuota.data.usage.user_count}{accountQuota.data.usage.user_quota > 0 ? ` / ${accountQuota.data.usage.user_quota}` : ""}</span>
            </div>
            <div className="flex items-center gap-2 rounded-lg bg-muted/40 px-3 py-2 text-sm">
              <Gauge size={16} className="text-accent" />
              <span>{t("reseller.dashboard.assigned")}: {formatBytes(accountQuota.data.usage.traffic_allocated, false)}</span>
            </div>
            <div className="flex items-center gap-2 rounded-lg bg-muted/40 px-3 py-2 text-sm">
              <TrendingUp size={16} className="text-success" />
              <span>{t("reseller.dashboard.consumed")}: {formatBytes(accountQuota.data.usage.traffic_used, false)}</span>
            </div>
          </div>
        </Card>
      )}

      <div className="grid grid-cols-1 sm:grid-cols-2 xl:grid-cols-4 gap-4">
        <StatsCard
          title={t("overview.totalUsers")}
          value={overviewLoading ? "—" : totalUsers}
          icon={<Users size={18} />}
          color="cyan"
          delay={0.05}
          subLabel={`${byStatus.active ?? 0} ${t("overview.activeShort")}`}
        />
        <StatsCard
          title={t("overview.nodesOnline")}
          value={overviewLoading ? "—" : onlineCount}
          suffix={totalNodes > 0 ? `/ ${totalNodes}` : undefined}
          icon={<Server size={18} />}
          color="green"
          delay={0.1}
        />
        <StatsCard
          title={t("overview.trafficUsed")}
          value={overviewLoading ? "—" : formatBytes(totalUsed, false)}
          icon={<Zap size={18} />}
          color="purple"
          delay={0.15}
          subLabel={
            peakBucket > 0
              ? `${t("overview.peak")} ${formatBytes(peakBucket, false)}/min`
              : undefined
          }
        />
        <StatsCard
          title={t("overview.liveConnectionsTitle")}
          value={overviewLoading ? "—" : totalConnections}
          icon={<Wifi size={18} />}
          color="blue"
          delay={0.2}
          subLabel={`${byStatus.active ?? 0} ${t("overview.activeAccounts")}`}
        />
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        <GlassCard className="space-y-4">
          <div className="flex items-center justify-between border-b border-border/60 pb-3">
            <div>
              <h3 className="text-base font-bold text-fg flex items-center gap-2">
                <Server size={18} className="text-success" />
                {t("overview.nodeFleet")}
              </h3>
              <p className="text-xs text-fg-subtle mt-0.5">
                {t("overview.liveApiHint")}
              </p>
            </div>
            <Link to="/nodes" className="text-xs text-primary hover:underline font-medium flex items-center gap-1">
              {t("overview.allNodes")} <ArrowUpRight size={13} />
            </Link>
          </div>
          {overviewLoading ? (
            <div className="h-32 animate-pulse rounded-lg bg-surface-2/50" />
          ) : fleetItems.length === 0 ? (
            <p className="text-sm text-fg-muted py-6 text-center">{t("overview.noNodesEnrolled")}</p>
          ) : (
            <div className="space-y-3">
              {fleetItems.map((node) => {
                const status = nodeFleetStatus(node);
                const load = Math.max(node.health?.cpu_percent ?? 0, node.health?.mem_percent ?? 0);
                return (
                  <div
                    key={node.id}
                    className="p-3.5 rounded-2xl bg-surface-2/60 border border-border/80 hover:border-primary/30 transition-all space-y-2.5"
                  >
                    <div className="flex items-center justify-between gap-2">
                      <div className="min-w-0">
                        <p className="text-xs font-semibold text-fg truncate">{node.name}</p>
                        <p className="text-[10px] text-fg-subtle">
                          {node.core} · {node.health?.connections ?? 0} conn
                        </p>
                      </div>
                      <StatusBadge status={status} label={status} pulse={status === "active"} />
                    </div>
                    <div className="space-y-1">
                      <div className="flex justify-between text-[10px]">
                        <span className="text-fg-subtle">CPU / RAM</span>
                        <span
                          className={cn(
                            "font-bold",
                            load > 75 ? "text-danger" : load > 50 ? "text-warning" : "text-success",
                          )}
                        >
                          {load.toFixed(0)}%
                        </span>
                      </div>
                      <div className="w-full bg-surface-3 h-1.5 rounded-full overflow-hidden">
                        <div
                          className={cn(
                            "h-full rounded-full transition-all duration-500",
                            load > 75 ? "bg-danger" : load > 50 ? "bg-warning" : "bg-success",
                          )}
                          style={{ width: `${Math.min(100, load)}%` }}
                        />
                      </div>
                    </div>
                  </div>
                );
              })}
            </div>
          )}
        </GlassCard>

        <GlassCard className="space-y-4">
          <div className="flex items-center justify-between border-b border-border/60 pb-3">
            <div>
              <h3 className="text-base font-bold text-fg flex items-center gap-2">
                <Users size={18} className="text-accent" />
                {t("overview.topUsers")}
              </h3>
              <p className="text-xs text-fg-subtle mt-0.5">
                {t("overview.sortedByTraffic")}
              </p>
            </div>
            <Link to="/users" className="text-xs text-primary hover:underline font-medium flex items-center gap-1">
              {t("overview.allUsers")} <ArrowUpRight size={13} />
            </Link>
          </div>
          {recentUsersQ.isLoading ? (
            <div className="h-32 animate-pulse rounded-lg bg-surface-2/50" />
          ) : topUsers.length === 0 ? (
            <p className="text-sm text-fg-muted py-6 text-center">{t("overview.noUsersYet")}</p>
          ) : (
            <div className="space-y-3">
              {topUsers.map((user) => {
                const usedPct =
                  user.data_limit > 0 ? Math.min(100, (user.used_traffic / user.data_limit) * 100) : 0;
                return (
                  <div
                    key={user.id}
                    className="p-3.5 rounded-2xl bg-surface-2/60 border border-border/80 hover:border-border-strong transition-all flex flex-col sm:flex-row sm:items-center justify-between gap-3"
                  >
                    <div className="flex items-center gap-3 min-w-0">
                      <div className="h-9 w-9 rounded-xl bg-surface-3 flex items-center justify-center text-fg-subtle font-semibold text-xs flex-shrink-0">
                        {user.username.slice(0, 2).toUpperCase()}
                      </div>
                      <div className="min-w-0">
                        <p className="text-xs font-semibold text-fg truncate">{user.username}</p>
                        <p className="text-[10px] text-fg-muted">
                          {t("overview.expiresShort")}: {daysUntil(user.expire_at)}
                        </p>
                      </div>
                    </div>
                    <div className="flex items-center gap-3 self-end sm:self-center">
                      <div className="text-end text-xs">
                        <span className="font-bold text-fg">{formatBytes(user.used_traffic, false)}</span>
                        {user.data_limit > 0 && (
                          <span className="text-fg-subtle text-[10px]"> / {formatBytes(user.data_limit, false)}</span>
                        )}
                        {user.data_limit > 0 && (
                          <div className="w-24 bg-surface-3 h-1 rounded-full overflow-hidden mt-1 ms-auto">
                            <div
                              className={cn("h-full rounded-full", usedPct > 80 ? "bg-warning" : "bg-primary")}
                              style={{ width: `${usedPct}%` }}
                            />
                          </div>
                        )}
                      </div>
                      <StatusBadge status={user.status} label={user.status} pulse={user.status === "active"} />
                    </div>
                  </div>
                );
              })}
            </div>
          )}
        </GlassCard>
      </div>

      {/* ── System Info + Traffic Chart row ── */}
      <div className="grid grid-cols-1 gap-5 lg:grid-cols-2">
        {/* System Info */}
        <Card className="space-y-4">
          <h3 className="flex items-center gap-2 text-xs font-semibold uppercase tracking-wider text-fg-subtle"><Box size={13} /> System</h3>
          {s ? (
            <div className="grid grid-cols-2 gap-y-3 gap-x-6 text-sm">
              <div className="flex items-center gap-2"><Timer size={13} className="text-accent" /><span className="text-fg-muted">Uptime</span></div>
              <span className="font-semibold text-fg">{fmtUptime(s.uptime_seconds)}</span>
              <div className="flex items-center gap-2"><MonitorSmartphone size={13} className="text-primary" /><span className="text-fg-muted">Host</span></div>
              <span className="font-semibold text-fg">{s.hostname}</span>
              <div className="flex items-center gap-2"><Cpu size={13} className="text-warning" /><span className="text-fg-muted">Platform</span></div>
              <span className="font-semibold text-fg">{s.os}/{s.arch}</span>
              <div className="flex items-center gap-2"><MemoryStick size={13} className="text-success" /><span className="text-fg-muted">Memory</span></div>
              <span className="font-semibold text-fg">{formatBytes(s.mem_alloc_bytes, false)} / {formatBytes(s.mem_sys_bytes, false)}</span>
              <div className="flex items-center gap-2"><Layers size={13} className="text-fg-subtle" /><span className="text-fg-muted">Inbounds</span></div>
              <span className="font-semibold text-fg">{inboundCount}</span>
              <div className="flex items-center gap-2"><Activity size={13} className="text-fg-subtle" /><span className="text-fg-muted">Goroutines</span></div>
              <span className="font-semibold text-fg">{s.goroutines}</span>
            </div>
          ) : <div className="h-32 animate-pulse rounded-lg bg-surface-2/50" />}
        </Card>

        {/* Real-time traffic sparkline */}
        <Card className="flex flex-col justify-between">
          <h3 className="flex items-center gap-2 text-xs font-semibold uppercase tracking-wider text-fg-subtle"><TrendingUp size={13} /> Real-time Bandwidth</h3>
          <div className="mt-4 flex-1">
            {trafficSamples.length > 1 ? (
              <Sparkline data={trafficSamples} className="h-20" />
            ) : (
              <div className="flex h-20 items-center justify-center text-xs text-fg-subtle">Collecting samples…</div>
            )}
          </div>
          <div className="mt-3 flex items-center justify-between text-xs text-fg-muted">
            <span>Last {trafficSamples.length} intervals</span>
            <span className="font-semibold text-fg">{formatBytes(trafficSamples[trafficSamples.length - 1] ?? 0, false)}/tick</span>
          </div>
        </Card>
      </div>

      {/* ── Fleet-wide traffic time-series ── */}
      <Card className="space-y-3">
        <div className="flex items-center justify-between">
          <h3 className="flex items-center gap-2 text-xs font-semibold uppercase tracking-wider text-fg-subtle"><TrendingUp size={13} /> Traffic — last hour</h3>
          <span className="text-[11px] text-fg-subtle">1-minute buckets</span>
        </div>
        {trafficSeries.isLoading ? (
          <div className="h-32 animate-pulse rounded-lg bg-surface-2/50" />
        ) : (
          <TrafficSeriesChart points={trafficSeries.data?.points ?? []} />
        )}
      </Card>

      {/* ── Status + Traffic ── */}
      <div className="grid grid-cols-1 gap-6 lg:grid-cols-3">
        {/* Status breakdown — takes 2/3 on large screens */}
        <Card className="space-y-5 lg:col-span-2">
          <h3 className="text-xs font-semibold uppercase tracking-wider text-fg-subtle">User Status</h3>
          <div className="flex h-2.5 overflow-hidden rounded-full bg-border/30 dark:bg-surface-2/60">
            {Object.entries(byStatus).map(([s, v]) => v ? <div key={s} className={cn("transition-all duration-700", STATUS_META[s]?.color ?? "bg-fg-subtle")} style={{ width: `${(v / (totalUsers || 1)) * 100}%` }} title={`${STATUS_META[s]?.label}: ${v}`} /> : null)}
          </div>
          <div className="grid grid-cols-2 gap-x-8 gap-y-2.5 sm:grid-cols-3 md:grid-cols-5">
            {(["active", "limited", "expired", "disabled", "on_hold"] as const).map((st) => {
              const meta = STATUS_META[st]; const count = byStatus[st] ?? 0;
              const percent = totalUsers > 0 ? ((count / totalUsers) * 100).toFixed(0) : "0";
              return (
                <div key={st} className="flex items-center justify-between gap-2">
                  <div className="flex items-center gap-2"><div className={cn("h-2.5 w-2.5 rounded-[3px]", meta.color)} /><span className="text-xs text-fg-muted">{meta.label}</span></div>
                  <span className="text-xs font-semibold tabular-nums text-fg">{count} <span className="text-fg-subtle">({percent}%)</span></span>
                </div>
              );
            })}
          </div>
          <div className="flex items-center justify-between border-t border-border/40 pt-3"><span className="text-xs font-medium text-fg-subtle">Total</span><span className="text-sm font-bold text-fg">{totalUsers}</span></div>
        </Card>

        {/* Traffic summary */}
        <Card className="space-y-4">
          <h3 className="text-xs font-semibold uppercase tracking-wider text-fg-subtle">Traffic</h3>
          <div className="space-y-3">
            <div className="flex items-center justify-between"><div className="flex items-center gap-2 text-xs text-fg-muted"><TrendingUp size={13} className="text-warning" />Consumed</div><span className="text-sm font-bold tabular-nums text-fg">{formatBytes(totalUsed, false)}</span></div>
            <div className="flex items-center justify-between"><div className="flex items-center gap-2 text-xs text-fg-muted"><Activity size={13} className="text-success" />Active</div><span className="text-sm font-bold tabular-nums text-fg">{byStatus.active ?? 0}</span></div>
            <div className="flex items-center justify-between"><div className="flex items-center gap-2 text-xs text-fg-muted"><Layers size={13} className="text-accent" />Inbounds</div><span className="text-sm font-bold tabular-nums text-fg">{inboundCount}</span></div>
          </div>
        </Card>
      </div>

      {/* ── Core Engines ── */}
      <div className="grid grid-cols-1 gap-5 md:grid-cols-2">
        <CoreCard name="Xray" version={xrayVer} running={xrayRunning}
          onRestart={() => xrayNode && restartCore.mutate(xrayNode.id)}
          onStop={() => xrayNode && stopCore.mutate(xrayNode.id)} />
        <CoreCard name="Sing-Box" version={singboxVer} running={singboxRunning}
          onRestart={() => singboxNode && restartCore.mutate(singboxNode.id)}
          onStop={() => singboxNode && stopCore.mutate(singboxNode.id)} />
      </div>

      {/* ── Charts Panel ── */}
      <ChartsPanel />
    </div>
  );
}

/* ═══════════════════════════════════════════════════════════════════════
   CHARTS PANEL — System History / Xray Metrics / Sing-Box Metrics
   Live data sampled from the system/overview polling.
   ═══════════════════════════════════════════════════════════════════════ */
import { useState, useEffect, useRef } from "react";
import { BarChart3, X as XIcon } from "lucide-react";

type ChartTab = "system" | "xray" | "singbox";

const SYSTEM_METRICS = ["CPU", "RAM", "Bandwidth", "Connections", "Online"] as const;
const XRAY_METRICS = ["Heap", "Sys", "Objects", "GC Count", "Connections"] as const;
const SINGBOX_METRICS = ["Heap", "Sys", "Objects", "GC Count", "Connections"] as const;
const TIME_RANGES = ["2m", "5m", "30m", "1h", "2h"] as const;

type SystemMetric = (typeof SYSTEM_METRICS)[number];
type CoreMetric = (typeof XRAY_METRICS)[number];

function ChartsPanel() {
  const [open, setOpen] = useState(false);
  const [mainTab, setMainTab] = useState<ChartTab>("system");
  const [systemMetric, setSystemMetric] = useState<SystemMetric>("CPU");
  const [xrayMetric, setXrayMetric] = useState<CoreMetric>("Heap");
  const [singboxMetric, setSingboxMetric] = useState<CoreMetric>("Heap");
  const [timeRange, setTimeRange] = useState<string>("2m");

  // Hooks must run unconditionally (React rules of hooks).
  const sys = useSystem().data;
  const ov = useOverview().data;

  if (!open) {
    return (
      <Card className="p-4">
        <div className="flex items-center justify-between">
          <h3 className="flex items-center gap-2 text-sm font-semibold text-fg">
            <BarChart3 size={15} /> Charts
          </h3>
        </div>
        <div className="mt-3 grid grid-cols-3 gap-3 border-t border-border/40 pt-4">
          <button onClick={() => { setMainTab("system"); setOpen(true); }} className="flex items-center justify-center gap-2 rounded-xl bg-surface-2/40 py-3 text-sm font-medium text-fg-muted transition hover:bg-primary/10 hover:text-primary">
            <BarChart3 size={15} /> System History
          </button>
          <button onClick={() => { setMainTab("xray"); setOpen(true); }} className="flex items-center justify-center gap-2 rounded-xl bg-surface-2/40 py-3 text-sm font-medium text-fg-muted transition hover:bg-primary/10 hover:text-primary">
            <BarChart3 size={15} /> Xray Metrics
          </button>
          <button onClick={() => { setMainTab("singbox"); setOpen(true); }} className="flex items-center justify-center gap-2 rounded-xl bg-surface-2/40 py-3 text-sm font-medium text-fg-muted transition hover:bg-primary/10 hover:text-primary">
            <BarChart3 size={15} /> Sing-Box Metrics
          </button>
        </div>
      </Card>
    );
  }

  const subTabs = mainTab === "system" ? SYSTEM_METRICS : mainTab === "xray" ? XRAY_METRICS : SINGBOX_METRICS;
  const activeSubTab = mainTab === "system" ? systemMetric : mainTab === "xray" ? xrayMetric : singboxMetric;
  const setSubTab = mainTab === "system" ? setSystemMetric : mainTab === "xray" ? setXrayMetric : setSingboxMetric;
  const title = mainTab === "system" ? "System History" : mainTab === "xray" ? "Xray Metrics" : "Sing-Box Metrics";

  // Real value for the active metric. System metrics come from the live host
  // stats / overview; core metrics come from Go runtime (shared process).
  let liveValue: number | null = null;
  let liveUnit = "";

  if (mainTab === "system") {
    switch (systemMetric) {
      case "CPU": liveValue = sys?.cpu_percent ?? null; liveUnit = "%"; break;
      case "RAM": liveValue = sys?.mem_percent ?? null; liveUnit = "%"; break;
      case "Bandwidth": liveValue = null; liveUnit = ""; break; // see traffic chart above
      case "Connections": liveValue = ov?.nodes.items?.reduce((s: number, n) => s + (n.health?.connections ?? 0), 0) ?? null; break;
      case "Online": liveValue = ov?.nodes.online ?? null; break;
    }
  } else {
    // Xray / Sing-Box — derive from Go runtime stats (same process hosts the core)
    const allocMB = sys ? sys.mem_alloc_bytes / (1024 * 1024) : null;
    const sysMB = sys ? sys.mem_sys_bytes / (1024 * 1024) : null;
    const goroutines = sys?.goroutines ?? null;
    const metric = mainTab === "xray" ? xrayMetric : singboxMetric;
    switch (metric) {
      case "Heap": liveValue = allocMB; liveUnit = " MB"; break;
      case "Sys": liveValue = sysMB; liveUnit = " MB"; break;
      case "Objects": liveValue = goroutines; liveUnit = ""; break;
      case "GC Count": liveValue = null; liveUnit = ""; break;
      case "Connections": liveValue = ov?.nodes.items?.reduce((s: number, n) => s + (n.health?.connections ?? 0), 0) ?? null; break;
    }
  }

  return (
    <Card className="space-y-0 p-0 animate-slide-up">
      {/* Header */}
      <div className="flex items-center justify-between border-b border-border/40 px-5 py-3">
        <div className="flex items-center gap-3">
          <h3 className="text-sm font-bold text-fg">{title}</h3>
          <select
            value={timeRange}
            onChange={(e) => setTimeRange(e.target.value)}
            className="rounded-lg border border-border bg-surface-2/50 px-2 py-1 text-xs font-medium text-fg-muted outline-none"
          >
            {TIME_RANGES.map((r) => <option key={r} value={r}>{r}</option>)}
          </select>
        </div>
        <div className="flex items-center gap-2">
          {/* Main tab switcher */}
          {(["system", "xray", "singbox"] as ChartTab[]).map((tab) => (
            <button
              key={tab}
              onClick={() => setMainTab(tab)}
              className={cn(
                "rounded-lg px-2.5 py-1 text-xs font-medium transition",
                mainTab === tab ? "bg-primary/10 text-primary" : "text-fg-subtle hover:text-fg-muted",
              )}
            >
              {tab === "system" ? "System" : tab === "xray" ? "Xray" : "Sing-Box"}
            </button>
          ))}
          <button onClick={() => setOpen(false)} className="ms-2 grid h-7 w-7 place-items-center rounded-lg text-fg-subtle transition hover:bg-surface-2 hover:text-fg">
            <XIcon size={14} />
          </button>
        </div>
      </div>

      {/* Sub-tabs */}
      <div className="flex gap-1 border-b border-border/30 px-5 py-2">
        {subTabs.map((tab) => (
          <button
            key={tab}
            onClick={() => setSubTab(tab as any)}
            className={cn(
              "rounded-lg px-3 py-1.5 text-xs font-medium transition",
              activeSubTab === tab ? "bg-accent/10 text-accent border-b-2 border-accent" : "text-fg-muted hover:text-fg",
            )}
          >
            {tab}
          </button>
        ))}
      </div>

      {/* Chart area */}
      <div className="p-5">
        <LiveChart value={liveValue} unit={liveUnit} label={`${activeSubTab}`} timeRange={timeRange} />
      </div>
    </Card>
  );
}

/* ═══════ Live Chart — SVG sampling a real metric value over time ═══════ */
function LiveChart({ value, unit, label, timeRange }: { value: number | null; unit: string; label: string; timeRange: string }) {
  const maxSamples = timeRange === "2m" ? 24 : 60;
  const [samples, setSamples] = useState<number[]>([]);
  const valRef = useRef<number | null>(value);
  valRef.current = value;

  useEffect(() => {
    setSamples([]);
    const interval = timeRange === "2m" || timeRange === "5m" ? 5000 : 10000;
    const id = setInterval(() => {
      const v = valRef.current;
      if (v == null || Number.isNaN(v)) return;
      setSamples((prev) => [...prev.slice(-(maxSamples - 1)), v]);
    }, interval);
    return () => clearInterval(id);
  }, [timeRange, maxSamples]);

  if (value == null || Number.isNaN(value)) {
    return <div className="flex h-40 items-center justify-center text-xs text-fg-subtle">No live data for this metric.</div>;
  }
  if (samples.length < 2) return <div className="flex h-40 items-center justify-center text-xs text-fg-subtle">Collecting data…</div>;

  const max = Math.max(...samples, 1);
  const min = Math.min(...samples, 0);
  const range = max - min || 1;
  const w = 600, h = 160;
  const pts = samples.map((v, i) => {
    const x = (i / (samples.length - 1)) * w;
    const y = h - ((v - min) / range) * (h - 20) - 10;
    return `${x},${y}`;
  }).join(" ");

  const maxVal = max;
  const minVal = min;
  const currentVal = samples[samples.length - 1];

  // Y-axis labels
  const yLabels = [max, max * 0.75, max * 0.5, max * 0.25, min];

  return (
    <div>
      {/* Title + live values */}
      <div className="mb-3 flex items-center justify-between">
        <span className="text-xs font-medium text-fg-muted">{label}</span>
        <div className="flex items-center gap-3 text-[11px]">
          <span className="text-danger">▲ {maxVal.toFixed(1)}{unit}</span>
          <span className="text-success">▼ {minVal.toFixed(1)}{unit}</span>
        </div>
      </div>

      {/* Chart */}
      <div className="relative rounded-xl bg-surface-2/30 p-3 dark:bg-surface/40">
        {/* Y-axis grid */}
        <div className="absolute inset-y-3 start-3 flex flex-col justify-between text-[9px] text-fg-subtle/60">
          {yLabels.map((v, i) => <span key={i}>{v.toFixed(0)}{unit}</span>)}
        </div>

        <svg viewBox={`0 0 ${w} ${h}`} className="ms-8 w-full" preserveAspectRatio="none" style={{ height: 160 }}>
          {/* Grid lines */}
          {[0, 0.25, 0.5, 0.75, 1].map((pct) => (
            <line key={pct} x1="0" y1={10 + pct * (h - 20)} x2={w} y2={10 + pct * (h - 20)} stroke="hsl(var(--border))" strokeWidth="0.5" strokeDasharray="4 4" opacity="0.4" />
          ))}
          {/* Fill */}
          <polygon points={`0,${h} ${pts} ${w},${h}`} fill="url(#chartGrad)" />
          {/* Line */}
          <polyline points={pts} fill="none" stroke="hsl(var(--accent))" strokeWidth="2" strokeLinejoin="round" strokeLinecap="round" />
          {/* Current value dot */}
          <circle cx={w} cy={h - ((currentVal - min) / range) * (h - 20) - 10} r="4" fill="hsl(var(--accent))" />
          <defs>
            <linearGradient id="chartGrad" x1="0" y1="0" x2="0" y2="1">
              <stop offset="0%" stopColor="hsl(var(--accent))" stopOpacity="0.25" />
              <stop offset="100%" stopColor="hsl(var(--accent))" stopOpacity="0.02" />
            </linearGradient>
          </defs>
        </svg>

        {/* X-axis time labels */}
        <div className="ms-8 mt-1 flex justify-between text-[9px] text-fg-subtle/60">
          <span>{new Date(Date.now() - (samples.length * 5000)).toLocaleTimeString([], { hour: "2-digit", minute: "2-digit", second: "2-digit" })}</span>
          <span>{new Date().toLocaleTimeString([], { hour: "2-digit", minute: "2-digit", second: "2-digit" })}</span>
        </div>
      </div>
    </div>
  );
}
