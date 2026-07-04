import {
  Users, Wifi, Zap, Clock, ChevronRight, Server, ArrowUpRight,
  Power, RotateCcw, Tag, Shield, Radio, Gauge, TrendingUp,
} from "lucide-react";
import { Link } from "react-router-dom";
import { motion } from "framer-motion";
import { useOverview, useSystem, useTrafficSeries, useRestartCore, useStopCore } from "@/api/policy-hooks";
import { useAccountQuota } from "@/api/quota-hooks";
import { useAllInbounds, useNodes, useUsers, useVersion } from "@/api/hooks";
import { useAuth } from "@/auth/auth";
import { Card } from "@/components/ui";
import { TrafficSeriesChart } from "@/components/TrafficSeriesChart";
import { GlassCard, StatsCard, StatusBadge, ProtocolDonutChart, formatDailyBandwidth } from "@/components/veltrix";
import { useI18n } from "@/i18n/i18n";
import { useTitle } from "@/lib/useTitle";
import { cn, formatBytes } from "@/lib/utils";
import type { Overview as OverviewData } from "@/api/types";


/* ═══════ Status colors ═══════ */
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
  const trafficSeries = useTrafficSeries();

  const s = sys.data;
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

  const widgets = data?.widgets;
  const trends = widgets?.trends;
  const protocolSlices = (widgets?.protocols ?? []).map((p, i) => ({
    label: p.label,
    value: p.count,
    color: ["#3B82F6", "#8B5CF6", "#14B8A6", "#64748B", "#F59E0B", "#EC4899"][i % 6],
  }));
  const allHealthy = totalNodes > 0 && onlineCount === totalNodes;
  const standbyNodes = totalNodes - onlineCount;
  const inboundCount = widgets?.routing?.inbounds ?? inbounds.data?.length ?? 0;

  return (
    <div className="space-y-6 animate-page-enter">
      <motion.div
        initial={{ opacity: 0, y: 10 }}
        animate={{ opacity: 1, y: 0 }}
        className="relative overflow-hidden rounded-3xl bg-gradient-to-br from-bg-elevated via-surface to-primary/[0.04] border border-border/80 p-6 md:p-8 shadow-xl"
      >
        <div className="absolute top-0 end-0 w-96 h-96 rounded-full bg-primary/10 blur-3xl pointer-events-none" />
        <div className="absolute bottom-0 start-0 w-72 h-72 rounded-full bg-accent/10 blur-3xl pointer-events-none" />
        <div className="absolute inset-0 bg-grid-pattern opacity-30 pointer-events-none" />

        <div className="relative z-10 flex flex-col xl:flex-row xl:items-start justify-between gap-6">
          <div className="space-y-3 max-w-2xl">
            <div className="flex flex-wrap items-center gap-2">
              <StatusBadge
                status={allHealthy ? "optimal" : onlineCount > 0 ? "warning" : "inactive"}
                label={allHealthy ? t("overview.allNodesHealthy") : `${onlineCount}/${totalNodes} ${t("overview.online")}`}
              />
              {coreLabel && (
                <span className="px-2.5 py-0.5 rounded-full bg-primary/15 text-primary border border-primary/30 text-[10px] font-semibold truncate max-w-xs">
                  {coreLabel}
                </span>
              )}
              {panelVersion && (
                <span className="px-2.5 py-0.5 rounded-full bg-surface-2/80 text-fg-muted border border-border/60 text-[10px] font-semibold">
                  v{panelVersion}
                </span>
              )}
            </div>
            <h1 className="text-2xl md:text-4xl font-black text-fg tracking-tight">
              {t("overview.commandTower")}
              {panelVersion ? ` v${panelVersion}` : ""}
            </h1>
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

          <div className="grid grid-cols-1 sm:grid-cols-2 gap-3 w-full xl:max-w-md flex-shrink-0">
            <div className="rounded-2xl border border-border/70 bg-surface/70 backdrop-blur-sm p-4">
              <div className="flex items-center gap-2 text-primary mb-2">
                <Shield size={16} />
                <span className="text-[10px] font-bold uppercase tracking-wider">{t("overview.activeProbingShield")}</span>
              </div>
              <p className="text-xl font-black text-fg tabular-nums">
                {widgets?.probing?.blocked_scanners ?? 0}
              </p>
              <p className="text-[10px] text-fg-subtle mt-1">
                {widgets?.probing?.enabled ? t("overview.probingBlocked") : t("overview.probingDisabled")}
              </p>
            </div>
            <div className="rounded-2xl border border-border/70 bg-surface/70 backdrop-blur-sm p-4">
              <div className="flex items-center gap-2 text-accent mb-2">
                <Radio size={16} />
                <span className="text-[10px] font-bold uppercase tracking-wider">{t("overview.smartRoutingRules")}</span>
              </div>
              <p className="text-xl font-black text-fg tabular-nums">
                {widgets?.routing?.active_rules ?? 0}
              </p>
              <p className="text-[10px] text-fg-subtle mt-1">
                {widgets?.routing?.routing_packs ?? 0} {t("overview.routingPacksShort")}
              </p>
            </div>
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
          title={t("overview.totalSubscriptions")}
          value={overviewLoading ? "—" : totalUsers}
          change={trends?.users_pct}
          icon={<Users size={18} />}
          color="cyan"
          delay={0.05}
          subLabel={`${byStatus.active ?? 0} ${t("overview.activeShort")}`}
        />
        <StatsCard
          title={t("overview.nodeFleetOnline")}
          value={overviewLoading ? "—" : onlineCount}
          suffix={totalNodes > 0 ? `/ ${totalNodes}` : undefined}
          change={0}
          icon={<Server size={18} />}
          color="green"
          delay={0.1}
          subLabel={standbyNodes > 0 ? `${standbyNodes} ${t("overview.standby")}` : undefined}
        />
        <StatsCard
          title={t("overview.dailyBandwidth")}
          value={overviewLoading ? "—" : formatDailyBandwidth(totalUsed)}
          change={trends?.bandwidth_pct}
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
          title={t("overview.activeSessions")}
          value={overviewLoading ? "—" : totalConnections}
          change={trends?.sessions_pct}
          icon={<Wifi size={18} />}
          color="blue"
          delay={0.2}
          subLabel={`${byStatus.active ?? 0} ${t("overview.acrossAccounts")}`}
        />
      </div>

      <div className="grid grid-cols-1 xl:grid-cols-3 gap-6">
        <GlassCard className="xl:col-span-2 space-y-4">
          <div className="flex items-center justify-between border-b border-border/60 pb-3">
            <h3 className="text-base font-bold text-fg">{t("overview.liveTrafficStream")}</h3>
            <span className="text-[10px] text-fg-subtle flex items-center gap-1">
              <Clock size={11} />
              {dataUpdatedAt > 0 ? new Date(dataUpdatedAt).toLocaleTimeString() : t("overview.live")}
            </span>
          </div>
          {trafficSeries.isLoading ? (
            <div className="h-48 animate-pulse rounded-lg bg-surface-2/50" />
          ) : (
            <TrafficSeriesChart points={trafficSeries.data?.points ?? []} />
          )}
        </GlassCard>

        <GlassCard className="space-y-4">
          <h3 className="text-base font-bold text-fg border-b border-border/60 pb-3">{t("overview.protocolBreakdown")}</h3>
          <ProtocolDonutChart
            slices={protocolSlices}
            centerValue={totalConnections || byStatus.active || 0}
            centerLabel={t("overview.sessionsCenter")}
          />
        </GlassCard>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        <GlassCard className="space-y-4">
          <div className="flex items-center justify-between border-b border-border/60 pb-3">
            <div>
              <h3 className="text-base font-bold text-fg flex items-center gap-2">
                <Server size={18} className="text-success" />
                {t("overview.nodeFleetTelemetry")}
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
                {t("overview.activeUsersPool")}
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

      {(xrayNode || singboxNode) && (
        <div className="grid grid-cols-1 gap-5 md:grid-cols-2">
          {xrayNode && (
            <CoreCard name="Xray" version={xrayVer} running={xrayRunning}
              onRestart={() => restartCore.mutate(xrayNode.id)}
              onStop={() => stopCore.mutate(xrayNode.id)} />
          )}
          {singboxNode && (
            <CoreCard name="Sing-Box" version={singboxVer} running={singboxRunning}
              onRestart={() => restartCore.mutate(singboxNode.id)}
              onStop={() => stopCore.mutate(singboxNode.id)} />
          )}
        </div>
      )}
    </div>
  );
}
