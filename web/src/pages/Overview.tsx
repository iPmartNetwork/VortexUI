import { useState } from "react";
import {
  Users, Wifi, Zap, Server, ArrowUpRight,
  Power, RotateCcw, Tag, Shield, Radio, Gauge, TrendingUp,
} from "lucide-react";
import { Link } from "react-router-dom";
import { motion } from "framer-motion";
import {
  useOverview, useSystem, useTrafficSeries,
  useRestartCore, useStopCore, type TrafficRange,
} from "@/api/policy-hooks";
import { useAccountQuota } from "@/api/quota-hooks";
import { useNodes, useVersion } from "@/api/hooks";
import { useAuth } from "@/auth/auth";
import { Card } from "@/components/ui";
import { TrafficSeriesChart } from "@/components/TrafficSeriesChart";
import {
  GlassCard, StatsCard, StatusBadge,
  ProtocolDonutChart, formatDailyBandwidth,
} from "@/components/veltrix";
import { useI18n } from "@/i18n/i18n";
import { useTitle } from "@/lib/useTitle";
import { cn, formatBytes } from "@/lib/utils";

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

/* ════════════════════════════════════════════════════
   OVERVIEW PAGE
   ════════════════════════════════════════════════════ */
export function Overview() {
  useTitle("Overview");
  const { sudo } = useAuth();
  const accountQuota = useAccountQuota();
  const { data, isLoading: overviewLoading } = useOverview();
  const sys = useSystem();
  const nodesQ = useNodes();
  const panelVersion = useVersion().data;
  const { t } = useI18n();
  const [trafficRange, setTrafficRange] = useState<TrafficRange>("24h");
  const trafficSeries = useTrafficSeries(trafficRange);

  const u = data?.users;
  const onlineCount = data?.nodes.online ?? 0;
  const totalNodes = data?.nodes.total ?? 0;
  const byStatus = u?.by_status ?? {};
  const totalUsers = u?.total ?? 0;
  const totalUsed = u?.total_used ?? 0;

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

  const widgets = data?.widgets;
  const trends = widgets?.trends;
  const topUsers = widgets?.top_users ?? [];
  const nodeFleet = widgets?.node_fleet ?? [];
  const protocolSlices = (widgets?.protocols ?? []).map((p, i) => ({
    label: p.label,
    value: p.count,
    color: ["#3B82F6", "#8B5CF6", "#14B8A6", "#64748B", "#F59E0B", "#EC4899"][i % 6],
  }));
  const allHealthy = totalNodes > 0 && onlineCount === totalNodes;
  const standbyNodes = totalNodes - onlineCount;

  const coreLabel = [
    xrayVer !== "—" ? `Xray ${xrayVer}` : null,
    singboxVer !== "—" ? `sing-box ${singboxVer}` : null,
  ].filter(Boolean).join(" + ");

  /* Shield / routing display text */
  const probingBlocked = widgets?.probing?.blocked_scanners ?? 0;
  const probingEnabled = widgets?.probing?.enabled ?? false;
  const probingText = probingEnabled
    ? `${probingBlocked.toLocaleString()} DPI Scanners Blocked`
    : "Probing shield off";

  const activeRules = widgets?.routing?.active_rules ?? 0;
  const routingPacks = widgets?.routing?.routing_packs ?? 0;
  const routingText = activeRules > 0
    ? `${activeRules} Active Rules · ${routingPacks} Packs`
    : "No routing rules active";

  return (
    <div className="space-y-6 animate-page-enter">

      {/* ── HERO ── */}
      <motion.div
        initial={{ opacity: 0, y: 10 }}
        animate={{ opacity: 1, y: 0 }}
        className="relative overflow-hidden rounded-2xl border border-border/70 bg-gradient-to-br from-bg-elevated via-surface to-primary/[0.03] p-5 md:p-6 shadow-xl"
      >
        {/* decorative blobs */}
        <div className="absolute top-0 end-0 w-72 h-72 rounded-full bg-primary/8 blur-3xl pointer-events-none" />
        <div className="absolute bottom-0 start-0 w-56 h-56 rounded-full bg-accent/8 blur-3xl pointer-events-none" />
        <div className="absolute inset-0 bg-grid-pattern opacity-20 pointer-events-none" />

        <div className="relative z-10 flex flex-col xl:flex-row xl:items-start justify-between gap-6">
          {/* Left — title block */}
          <div className="space-y-3 max-w-2xl">
            {/* Badge row */}
            <div className="flex flex-wrap items-center gap-2">
              <StatusBadge
                status={allHealthy ? "optimal" : onlineCount > 0 ? "warning" : "inactive"}
                label={allHealthy ? t("overview.allNodesHealthy") : `${onlineCount}/${totalNodes} ${t("overview.online")}`}
              />
              {coreLabel && (
                <span className="px-2.5 py-0.5 rounded-full bg-primary/12 text-primary border border-primary/25 text-[10px] font-semibold truncate max-w-xs">
                  {coreLabel}
                </span>
              )}
            </div>

            {/* Title */}
            <h1 className="text-xl md:text-2xl font-black text-fg tracking-tight leading-tight whitespace-nowrap">
              {t("overview.commandTower")}
              {panelVersion && (
                <span className="text-primary"> v{panelVersion}</span>
              )}
            </h1>

            {/* Description */}
            <p className="text-[13px] text-fg-muted leading-relaxed max-w-xl">
              {overviewLoading || !s ? (
                <span className="animate-pulse">{t("overview.loadingTelemetry")}</span>
              ) : (
                <>
                  Real-time telemetry and anti-censorship control plane running{" "}
                  {allHealthy ? "optimally" : "in partial mode"} across{" "}
                  {totalNodes > 0 ? `${totalNodes} node${totalNodes !== 1 ? "s" : ""}` : "all nodes"}.
                  {" "}Uptime {fmtUptime(s.uptime_seconds)} · {totalConnections} live connections.
                </>
              )}
            </p>
          </div>

          {/* Right — status cards */}
          <div className="grid grid-cols-2 gap-2.5 w-full xl:max-w-sm flex-shrink-0">
            <div className="flex items-start gap-2.5 rounded-xl border border-border/60 bg-surface/60 backdrop-blur-sm p-3">
              <div className="h-8 w-8 rounded-full bg-amber-500/15 flex items-center justify-center flex-shrink-0">
                <Shield size={15} className="text-amber-400" />
              </div>
              <div className="min-w-0">
                <p className="text-[9px] font-bold uppercase tracking-wider text-fg-subtle leading-tight">
                  {t("overview.activeProbingShield")}
                </p>
                <p className="text-xs font-bold text-fg mt-1 leading-tight truncate">{probingText}</p>
              </div>
            </div>

            <div className="flex items-start gap-2.5 rounded-xl border border-border/60 bg-surface/60 backdrop-blur-sm p-3">
              <div className="h-8 w-8 rounded-full bg-teal-500/15 flex items-center justify-center flex-shrink-0">
                <Radio size={15} className="text-teal-400" />
              </div>
              <div className="min-w-0">
                <p className="text-[9px] font-bold uppercase tracking-wider text-fg-subtle leading-tight">
                  {t("overview.smartRoutingRules")}
                </p>
                <p className="text-xs font-bold text-fg mt-1 leading-tight truncate">{routingText}</p>
              </div>
            </div>
          </div>
        </div>
      </motion.div>

      {/* ── Reseller quota bar ── */}
      {!sudo && accountQuota.data?.usage && (
        <Card className="p-4">
          <div className="mb-3 flex flex-wrap items-center justify-between gap-2">
            <div className="text-sm font-semibold">{t("reseller.overview.pool")}</div>
            <Link to="/reseller-dashboard" className="inline-flex items-center gap-1 text-xs font-medium text-primary hover:underline">
              {t("reseller.overview.viewDashboard")} <ArrowUpRight size={14} />
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

      {/* ── Stat Cards ── */}
      <div className="grid grid-cols-1 sm:grid-cols-2 xl:grid-cols-4 gap-4">
        <StatsCard
          title={t("overview.totalSubscriptions")}
          value={overviewLoading ? "—" : totalUsers.toLocaleString()}
          change={trends?.users_pct}
          icon={<Users size={17} />}
          color="cyan"
          delay={0.05}
          subLabel={`${(byStatus.active ?? 0).toLocaleString()} ${t("overview.activeShort")}`}
        />
        <StatsCard
          title={t("overview.nodeFleetOnline")}
          value={overviewLoading ? "—" : onlineCount}
          suffix={totalNodes > 0 ? `/ ${totalNodes}` : undefined}
          change={0}
          icon={<Server size={17} />}
          color="green"
          delay={0.1}
          subLabel={standbyNodes > 0 ? `${standbyNodes} ${t("overview.standby")}` : undefined}
        />
        <StatsCard
          title={t("overview.dailyBandwidth")}
          value={overviewLoading ? "—" : formatDailyBandwidth(totalUsed)}
          change={trends?.bandwidth_pct}
          icon={<Zap size={17} />}
          color="purple"
          delay={0.15}
          subLabel={
            peakBucket > 0
              ? `Peak ${formatBytes(peakBucket, false)}/min`
              : undefined
          }
        />
        <StatsCard
          title={t("overview.activeSessions")}
          value={overviewLoading ? "—" : totalConnections.toLocaleString()}
          change={trends?.sessions_pct}
          icon={<Wifi size={17} />}
          color="blue"
          delay={0.2}
          subLabel={
            (byStatus.active ?? 0) > 0
              ? `Across ${(byStatus.active ?? 0)} accounts`
              : t("overview.acrossAccounts")
          }
        />
      </div>

      {/* ── Traffic chart + Protocol donut ── */}
      <div className="grid grid-cols-1 xl:grid-cols-3 gap-4">
        {/* Traffic chart — 2/3 width */}
        <GlassCard className="xl:col-span-2 space-y-2.5 !p-3.5">
          <div className="flex flex-wrap items-center justify-between gap-2 border-b border-border/60 pb-2">
            <div>
              <h3 className="text-[13px] font-bold text-fg flex items-center gap-1.5">
                {t("overview.liveTrafficStream")}
                <span className="inline-flex items-center gap-1 rounded-full bg-success/10 px-1.5 py-0.5 text-[8px] font-bold uppercase tracking-wider text-success">
                  <span className="relative flex h-1.5 w-1.5">
                    <span className="absolute inline-flex h-full w-full animate-ping rounded-full bg-success opacity-75" />
                    <span className="relative inline-flex h-1.5 w-1.5 rounded-full bg-success" />
                  </span>
                  Live
                </span>
              </h3>
              <p className="text-[9px] text-fg-subtle mt-0.5">{t("overview.trafficDeltaHint")}</p>
            </div>
            <div className="flex items-center gap-0.5 rounded-lg bg-surface-2/60 p-0.5">
              {(["24h", "7d", "30d"] as TrafficRange[]).map((r) => (
                <button
                  key={r}
                  type="button"
                  onClick={() => setTrafficRange(r)}
                  className={cn(
                    "px-2 py-0.5 rounded-md text-[9px] font-bold uppercase transition-all",
                    trafficRange === r
                      ? "bg-bg-elevated text-primary shadow-sm"
                      : "text-fg-muted hover:text-fg",
                  )}
                >
                  {r}
                </button>
              ))}
            </div>
          </div>
          {trafficSeries.isLoading ? (
            <div className="h-28 animate-pulse rounded-xl bg-surface-2/50" />
          ) : (
            <TrafficSeriesChart points={trafficSeries.data?.points ?? []} />
          )}
        </GlassCard>

        {/* Protocol breakdown — 1/3 width */}
        <GlassCard className="space-y-2.5 !p-3.5">
          <div className="border-b border-border/60 pb-2">
            <h3 className="text-[13px] font-bold text-fg">{t("overview.protocolBreakdown")}</h3>
            <p className="text-[9px] text-fg-subtle mt-0.5">Active connections by transport type</p>
          </div>
          <ProtocolDonutChart
            slices={protocolSlices}
            centerValue={totalConnections || byStatus.active || 0}
            centerLabel={t("overview.sessionsCenter")}
          />
        </GlassCard>
      </div>

      {/* ── Node Fleet + Active Users ── */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        {/* Node Fleet Telemetry */}
        <GlassCard className="space-y-4">
          <div className="flex items-center justify-between border-b border-border/60 pb-3">
            <div>
              <h3 className="text-base font-bold text-fg flex items-center gap-2">
                <Server size={17} className="text-success" />
                {t("overview.nodeFleetTelemetry")}
              </h3>
              <p className="text-[11px] text-fg-subtle mt-0.5">Live health monitoring &amp; load ratios</p>
            </div>
            <Link to="/nodes" className="flex items-center gap-1 text-xs font-semibold text-primary hover:underline">
              View Fleet <ArrowUpRight size={13} />
            </Link>
          </div>

          {overviewLoading ? (
            <div className="h-36 animate-pulse rounded-xl bg-surface-2/50" />
          ) : nodeFleet.length === 0 ? (
            <p className="text-sm text-fg-muted py-8 text-center">{t("overview.noNodesEnrolled")}</p>
          ) : (
            <div className="space-y-3">
              {nodeFleet.map((node) => {
                const load = Math.max(node.cpu_percent ?? 0, node.mem_percent ?? 0);
                const loadColor = load > 75 ? "bg-danger" : load > 50 ? "bg-warning" : "bg-success";
                const loadText = load > 75 ? "text-danger" : load > 50 ? "text-warning" : "text-success";
                return (
                  <div
                    key={node.id}
                    className="p-4 rounded-2xl bg-surface-2/50 border border-border/70 hover:border-primary/30 transition-all"
                  >
                    <div className="flex items-center justify-between gap-3 mb-3">
                      <div className="min-w-0">
                        <p className="text-sm font-bold text-fg truncate">
                          {node.name}{" "}
                          <span className="text-fg-subtle font-normal text-xs">
                            ({node.core === "singbox" ? "sing-box" : "Xray-Core"})
                          </span>
                        </p>
                        <p className="text-[11px] text-fg-subtle mt-0.5 truncate">
                          {node.location}
                          {node.ping_ms > 0 && (
                            <span className="text-fg-muted"> · <span className="text-primary font-semibold">{node.ping_ms}ms</span></span>
                          )}
                        </p>
                      </div>
                      <div className="flex items-center gap-2 flex-shrink-0">
                        {node.users_count > 0 && (
                          <span className="text-[10px] font-semibold bg-primary/10 text-primary px-2 py-0.5 rounded-full">
                            {node.users_count} users
                          </span>
                        )}
                        <StatusBadge status={node.status} label={node.status} pulse={node.status === "active"} />
                      </div>
                    </div>
                    <div className="space-y-1">
                      <div className="flex justify-between text-[11px]">
                        <span className="text-fg-subtle">CPU / RAM</span>
                        <span className={cn("font-bold", loadText)}>{load.toFixed(0)}%</span>
                      </div>
                      <div className="w-full bg-surface-3 h-1.5 rounded-full overflow-hidden">
                        <div
                          className={cn("h-full rounded-full transition-all duration-700", loadColor)}
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

        {/* Active Users Pool */}
        <GlassCard className="space-y-4">
          <div className="flex items-center justify-between border-b border-border/60 pb-3">
            <div>
              <h3 className="text-base font-bold text-fg flex items-center gap-2">
                <Users size={17} className="text-accent" />
                {t("overview.activeUsersPool")}
              </h3>
              <p className="text-[11px] text-fg-subtle mt-0.5">Real-time data accounting &amp; subscriptions</p>
            </div>
            <Link to="/users" className="flex items-center gap-1 text-xs font-semibold text-primary hover:underline">
              All Users <ArrowUpRight size={13} />
            </Link>
          </div>

          {overviewLoading ? (
            <div className="h-36 animate-pulse rounded-xl bg-surface-2/50" />
          ) : topUsers.length === 0 ? (
            <p className="text-sm text-fg-muted py-8 text-center">{t("overview.noUsersYet")}</p>
          ) : (
            <div className="space-y-3">
              {topUsers.map((user) => {
                const usedPct =
                  user.data_limit > 0
                    ? Math.min(100, (user.used_traffic / user.data_limit) * 100)
                    : 0;
                return (
                  <div
                    key={user.id}
                    className="p-4 rounded-2xl bg-surface-2/50 border border-border/70 hover:border-border-strong transition-all flex items-center justify-between gap-3"
                  >
                    <div className="flex items-center gap-3 min-w-0">
                      <div className="h-9 w-9 rounded-xl bg-primary/10 flex items-center justify-center text-primary font-black text-xs flex-shrink-0">
                        {user.username.slice(0, 2).toUpperCase()}
                      </div>
                      <div className="min-w-0">
                        <p className="text-sm font-bold text-fg truncate">{user.username}</p>
                        <p className="text-[11px] text-fg-muted truncate">
                          {user.protocol_label || "—"} · exp: {daysUntil(user.expire_at ?? null)}
                        </p>
                      </div>
                    </div>

                    <div className="flex items-center gap-3 flex-shrink-0">
                      <div className="text-end">
                        <p className="text-sm font-bold text-fg tabular-nums">
                          {formatBytes(user.used_traffic, false)}
                        </p>
                        {user.data_limit > 0 && (
                          <>
                            <p className="text-[10px] text-fg-subtle tabular-nums">
                              / {formatBytes(user.data_limit, false)}
                            </p>
                            <div className="w-20 bg-surface-3 h-1 rounded-full overflow-hidden mt-1 ms-auto">
                              <div
                                className={cn("h-full rounded-full", usedPct > 80 ? "bg-warning" : "bg-primary")}
                                style={{ width: `${usedPct}%` }}
                              />
                            </div>
                          </>
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

      {/* ── Core Engine Controls ── */}
      {(xrayNode || singboxNode) && (
        <div className="grid grid-cols-1 gap-5 md:grid-cols-2">
          {xrayNode && (
            <CoreCard
              name="Xray"
              version={xrayVer}
              running={xrayRunning}
              onRestart={() => restartCore.mutate(xrayNode.id)}
              onStop={() => stopCore.mutate(xrayNode.id)}
            />
          )}
          {singboxNode && (
            <CoreCard
              name="Sing-Box"
              version={singboxVer}
              running={singboxRunning}
              onRestart={() => restartCore.mutate(singboxNode.id)}
              onStop={() => stopCore.mutate(singboxNode.id)}
            />
          )}
        </div>
      )}
    </div>
  );
}
