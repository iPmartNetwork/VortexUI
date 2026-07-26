import { useQuery } from "@tanstack/react-query";
import {
  Activity,
  AlertTriangle,
  CalendarClock,
  Database,
  MessageSquarePlus,
  Smartphone,
  Wifi,
  Zap,
} from "lucide-react";
import { Link } from "react-router-dom";
import { motion } from "framer-motion";
import { PortalProtocolStatus } from "./PortalProtocolStatus";
import { QRCodeSVG } from "qrcode.react";
import { portalApi } from "./portalApi";
import { CopyField } from "@/components/CopyField";
import { UsageChart } from "@/components/UsageChart";
import type { UsagePoint } from "@/api/hooks";

import { GlassCard, StatsCard, StatusBadge } from "@/components/vortexui";
import { formatBytes } from "@/lib/utils";
import { cn } from "@/lib/utils";
import { useI18n } from "@/i18n/i18n";

interface DashboardData {
  username: string;
  status: string;
  data_limit: number;
  used_traffic: number;
  expire_at: string | null;
  device_limit: number;
  reset_strategy: string;
  sub_token: string;
  created_at: string;
}

interface SubscriptionData {
  subscription_url: string;
  formats: Record<string, string>;
  links: string[];
}

interface OnlineData {
  live_connections: number;
  active_devices: number;
  device_tracking: boolean;
  live_tracking: boolean;
}

function statusType(s: string): string {
  if (s === "limited") return "warning";
  if (s === "expired") return "error";
  if (s === "disabled") return "inactive";
  if (s === "on_hold") return "info";
  return "active";
}

const USAGE_DAYS = 30;

function usageFromUnix(): number {
  return Math.floor(Date.now() / 1000) - USAGE_DAYS * 86400;
}

function daysUntil(iso: string): number {
  const ms = new Date(iso).getTime() - Date.now();
  return Math.ceil(ms / (24 * 60 * 60 * 1000));
}

export function PortalDashboard() {
  const { t } = useI18n();

  const { data, isLoading, error } = useQuery({
    queryKey: ["portal-dashboard"],
    queryFn: () => portalApi<DashboardData>("/api/portal/dashboard"),
    refetchInterval: 30_000,
  });

  const sub = useQuery({
    queryKey: ["portal-subscription"],
    queryFn: () => portalApi<SubscriptionData>("/api/portal/subscription"),
    enabled: !!data,
  });

  const online = useQuery({
    queryKey: ["portal-online"],
    queryFn: () => portalApi<OnlineData>("/api/portal/online"),
    enabled: !!data,
    refetchInterval: 15_000,
  });

  const usage = useQuery({
    queryKey: ["portal-usage", USAGE_DAYS],
    queryFn: () =>
      portalApi<{ points: UsagePoint[] }>(
        `/api/portal/usage?bucket=1d&from=${usageFromUnix()}`,
      ),
    enabled: !!data,
    refetchInterval: 60_000,
  });

  const deepLink = useQuery({
    queryKey: ["portal-deeplink"],
    queryFn: () => portalApi<{ deep_link: string }>("/api/portal/deeplink"),
    enabled: !!data,
    retry: false,
  });

  if (isLoading) {
    return (
      <div className="flex items-center justify-center py-20">
        <div className="flex flex-col items-center gap-3">
          <div className="h-8 w-8 animate-spin rounded-full border-2 border-primary border-t-transparent" />
          <p className="text-sm text-fg-muted">{t("common.loading")}</p>
        </div>
      </div>
    );
  }
  if (error || !data) {
    return (
      <div className="flex flex-col items-center justify-center py-20 gap-3">
        <AlertTriangle size={32} className="text-danger" />
        <p className="text-sm text-danger">{t("portal.loadFailed")}</p>
      </div>
    );
  }

  const usagePercent =
    data.data_limit > 0 ? Math.min((data.used_traffic / data.data_limit) * 100, 100) : 0;
  const deviceLimitLabel = data.device_limit > 0 ? String(data.device_limit) : "∞";
  const activeDevices = online.data?.active_devices ?? 0;
  const liveConnections = online.data?.live_connections ?? 0;
  const usagePoints = usage.data?.points ?? [];

  const alerts: { tone: "warning" | "error" | "info"; text: string }[] = [];
  if (data.status === "expired") alerts.push({ tone: "error", text: t("portal.alert.expired") });
  else if (data.status === "limited") alerts.push({ tone: "warning", text: t("portal.alert.limited") });
  else if (data.status === "disabled") alerts.push({ tone: "error", text: t("portal.alert.disabled") });
  if (data.data_limit > 0 && usagePercent >= 90) {
    alerts.push({ tone: "warning", text: t("portal.alert.quotaHigh").replace("{pct}", usagePercent.toFixed(0)) });
  }
  if (data.expire_at) {
    const days = daysUntil(data.expire_at);
    if (days <= 0 && data.status !== "expired") alerts.push({ tone: "error", text: t("portal.alert.expired") });
    else if (days > 0 && days <= 7) {
      alerts.push({ tone: "warning", text: t("portal.alert.expiresSoon").replace("{days}", String(days)) });
    }
  }

  // Sparkline from usage points
  const sparkData = usagePoints.slice(-20).map(p => p.down + p.up);

  return (
    <div className="space-y-6 animate-page-enter">
      {/* ── Hero ── */}
      <motion.div
        initial={{ opacity: 0, y: 10 }}
        animate={{ opacity: 1, y: 0 }}
        className="relative overflow-hidden rounded-2xl border border-border/70 bg-gradient-to-br from-bg-elevated via-surface to-primary/[0.03] p-5 md:p-6 shadow-xl"
      >
        <div className="absolute top-0 end-0 w-64 h-64 rounded-full bg-primary/8 blur-3xl pointer-events-none" />
        <div className="absolute bottom-0 start-0 w-48 h-48 rounded-full bg-accent/8 blur-3xl pointer-events-none" />
        <div className="absolute inset-0 bg-grid-pattern opacity-20 pointer-events-none" />

        <div className="relative z-10 flex flex-col sm:flex-row items-start justify-between gap-4">
          <div className="space-y-2">
            <div className="flex items-center gap-2">
              <StatusBadge
                status={statusType(data.status) as any}
                label={data.status}
                pulse={data.status === "active"}
              />
              {online.data?.live_tracking && liveConnections > 0 && (
                <span className="inline-flex items-center gap-1 rounded-full bg-green-500/15 border border-green-500/30 px-2 py-0.5 text-[9px] font-bold uppercase tracking-wider text-green-400">
                  <span className="relative flex h-1.5 w-1.5">
                    <span className="absolute inline-flex h-full w-full animate-ping rounded-full bg-green-400 opacity-75" />
                    <span className="relative inline-flex h-1.5 w-1.5 rounded-full bg-green-500" />
                  </span>
                  {liveConnections} {t("portal.liveConnections")}
                </span>
              )}
            </div>
            <h1 className="text-xl md:text-2xl font-black text-fg tracking-tight">
              {t("portal.dashboardWelcome").replace("{name}", data.username)}
            </h1>
            <p className="text-[13px] text-fg-muted leading-relaxed">
              {t("portal.dashboardSubtitle")}
            </p>
          </div>

          <div className="flex flex-wrap items-center gap-2 flex-shrink-0">
            <Link
              to="/portal/plans"
              className="inline-flex items-center gap-1.5 rounded-xl bg-gradient-to-r from-primary to-primary/80 px-3.5 py-2 text-xs font-bold text-white shadow-lg shadow-primary/30 transition-all duration-200 hover:shadow-primary/50 hover:scale-105"
            >
              <Zap size={14} />
              {t("portal.quickRenew")}
            </Link>
            <Link
              to="/portal/tickets"
              className="inline-flex items-center gap-1.5 rounded-xl border border-border bg-bg-elevated/80 px-3.5 py-2 text-xs font-semibold text-fg transition-all duration-200 hover:bg-surface-2 hover:shadow-sm backdrop-blur-sm"
            >
              <MessageSquarePlus size={14} />
              {t("portal.quickTicket")}
            </Link>
          </div>
        </div>
      </motion.div>

      {/* ── Alerts ── */}
      {alerts.length > 0 && (
        <div className="space-y-2">
          {alerts.map((a, i) => (
            <motion.div
              key={i}
              initial={{ opacity: 0, x: -10 }}
              animate={{ opacity: 1, x: 0 }}
              transition={{ delay: i * 0.1 }}
              className={cn(
                "flex items-start gap-2.5 rounded-xl border px-4 py-3 text-sm shadow-sm backdrop-blur-sm",
                a.tone === "error" && "border-danger/40 bg-danger/10 text-danger shadow-danger/5",
                a.tone === "warning" && "border-warning/40 bg-warning/10 text-warning shadow-warning/5",
                a.tone === "info" && "border-accent/40 bg-accent/10 text-accent shadow-accent/5",
              )}
            >
              <AlertTriangle size={16} className="mt-0.5 shrink-0" />
              <span className="font-medium">{a.text}</span>
            </motion.div>
          ))}
        </div>
      )}

      {/* ── Stats Cards ── */}
      <div className="grid grid-cols-2 gap-3 lg:grid-cols-4 lg:gap-4">
        <StatsCard
          title={t("portal.dataUsed")}
          value={formatBytes(data.used_traffic, false)}
          subLabel={
            data.data_limit > 0
              ? `${t("portal.ofLimit")} ${formatBytes(data.data_limit, false)}`
              : t("portal.unlimitedPlan")
          }
          icon={<Database size={18} />}
          color="blue"
          sparkline={sparkData}
          live
        />
        <StatsCard
          title={t("portal.usage")}
          value={data.data_limit > 0 ? `${usagePercent.toFixed(0)}%` : "—"}
          subLabel={data.reset_strategy.replace("_", " ")}
          icon={<Activity size={18} />}
          color={usagePercent > 90 ? "red" : "green"}
          progress={usagePercent}
          progressColor={usagePercent > 90 ? "danger" : usagePercent > 70 ? "warning" : "success"}
          live
        />
        <StatsCard
          title={t("portal.expires")}
          value={data.expire_at ? new Date(data.expire_at).toLocaleDateString() : t("common.never")}
          subLabel={`${t("portal.memberSince")} ${new Date(data.created_at).toLocaleDateString()}`}
          icon={<CalendarClock size={18} />}
          color="orange"
          sparkline={sparkData.slice(0, 10)}
          sparkColor="#FB923C"
        />
        <StatsCard
          title={t("portal.activeDevices")}
          value={`${online.isLoading ? "…" : activeDevices} / ${deviceLimitLabel}`}
          subLabel={
            online.data?.live_tracking
              ? `${liveConnections} ${t("portal.liveConnections").toLowerCase()}`
              : t("portal.devicesUnavailable")
          }
          icon={<Smartphone size={18} />}
          color={data.device_limit > 0 && activeDevices >= data.device_limit ? "red" : "cyan"}
          live
        />
      </div>

      {/* ── Traffic Usage Bar ── */}
      {data.data_limit > 0 && (
        <GlassCard glow className="!p-5 relative overflow-hidden">
          <div className="absolute top-0 end-0 w-32 h-32 rounded-full bg-primary/5 blur-3xl pointer-events-none" />
          <div className="relative z-10">
            <div className="mb-3 flex items-center justify-between">
              <span className="text-xs font-bold uppercase tracking-wider text-fg-subtle">
                <Database size={13} className="inline me-1.5 text-primary" />
                {t("portal.trafficConsumption")}
              </span>
              <span className="tabular-nums font-bold text-fg text-sm">{usagePercent.toFixed(1)}%</span>
            </div>
            <div className="h-2.5 overflow-hidden rounded-full bg-surface-3/60 border border-border/30">
              <div
                className={cn(
                  "h-full rounded-full transition-all duration-700 ease-out shadow-sm",
                  usagePercent > 90 ? "bg-gradient-to-r from-danger to-red-600 shadow-danger/30" : 
                  usagePercent > 70 ? "bg-gradient-to-r from-warning to-orange-500 shadow-warning/20" : 
                  "bg-gradient-to-r from-primary to-accent shadow-primary/20",
                )}
                style={{ width: `${usagePercent}%` }}
              />
            </div>
            <div className="flex justify-between mt-1.5 text-[10px] text-fg-subtle/70">
              <span>0 GB</span>
              <span>{formatBytes(data.data_limit, false)}</span>
            </div>
          </div>
        </GlassCard>
      )}

      {/* ── Protocol Status ── */}
      <PortalProtocolStatus />

      {/* ── Usage Chart + Subscription ── */}
      <div className="grid gap-6 xl:grid-cols-[minmax(0,1fr)_380px] xl:items-stretch">
        <GlassCard glow className="!p-5 flex h-full min-h-0 flex-col">
          <div className="mb-4 flex flex-wrap items-center justify-between gap-2">
            <h3 className="text-xs font-bold uppercase tracking-wider text-fg-subtle flex items-center gap-1.5">
              <Activity size={14} className="text-primary" />
              {t("portal.dailyUsage")}
            </h3>
            {online.data?.live_tracking && (
              <span className="inline-flex items-center gap-1.5 text-xs text-fg-muted bg-surface-2/50 px-2 py-1 rounded-lg">
                <Wifi size={13} className={liveConnections > 0 ? "text-success" : "text-fg-subtle"} />
                {liveConnections} {t("portal.liveConnections").toLowerCase()}
              </span>
            )}
          </div>

          {usage.isLoading ? (
            <div className="h-40 animate-pulse rounded-xl bg-gradient-to-br from-surface-2/50 to-surface-3/30" />
          ) : (
            <UsageChart
              points={usagePoints}
              labels={{
                empty: t("portal.noTrafficYet"),
                up: t("portal.chartUp"),
                down: t("portal.chartDown"),
                peak: t("portal.chartPeak"),
              }}
            />
          )}

          {usagePoints.length > 0 && (
            <div className="mt-5 min-h-0 flex-1 overflow-x-auto rounded-xl border border-border/50 bg-surface-2/20">
              <table className="w-full text-sm">
                <thead>
                  <tr className="border-b border-border/40 bg-surface-2/40 text-fg-subtle">
                    <th className="px-4 py-2.5 text-start text-[10px] font-bold uppercase tracking-wider">
                      {t("portal.day")}
                    </th>
                    <th className="px-4 py-2.5 text-end text-[10px] font-bold uppercase tracking-wider">
                      {t("portal.chartUp")}
                    </th>
                    <th className="px-4 py-2.5 text-end text-[10px] font-bold uppercase tracking-wider">
                      {t("portal.chartDown")}
                    </th>
                    <th className="px-4 py-2.5 text-end text-[10px] font-bold uppercase tracking-wider">
                      {t("portal.total")}
                    </th>
                  </tr>
                </thead>
                <tbody>
                  {[...usagePoints].reverse().map((p) => (
                    <tr key={p.time} className="border-b border-border/20 last:border-0 hover:bg-surface-2/30 transition-colors">
                      <td className="px-4 py-2.5 tabular-nums text-fg text-xs">
                        {new Date(p.time).toLocaleDateString([], { weekday: "short", day: "numeric" })}
                      </td>
                      <td className="px-4 py-2.5 text-end tabular-nums text-fg-muted text-xs">
                        {formatBytes(p.up, false)}
                      </td>
                      <td className="px-4 py-2.5 text-end tabular-nums text-fg-muted text-xs">
                        {formatBytes(p.down, false)}
                      </td>
                      <td className="px-4 py-2.5 text-end tabular-nums font-semibold text-fg text-xs">
                        {formatBytes(p.up + p.down, false)}
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          )}
        </GlassCard>

        {sub.data && (
          <GlassCard glow className="!p-5 flex h-full min-h-0 flex-col">
            <h3 className="text-xs font-bold uppercase tracking-wider text-fg-subtle flex items-center gap-1.5">
              <Zap size={14} className="text-primary" />
              {t("portal.subscriptionLink")}
            </h3>

            <div className="mt-4 flex flex-col items-center gap-4">
              <div className="rounded-xl bg-white p-3 shadow-lg shadow-primary/10 ring-1 ring-border/20">
                <QRCodeSVG value={sub.data.subscription_url} size={148} />
              </div>

              <div className="w-full space-y-3">
                <div>
                  <p className="mb-1.5 text-[9px] font-bold uppercase tracking-wider text-fg-subtle/70">URL</p>
                  <CopyField value={sub.data.subscription_url} />
                </div>

                {(["clash", "singbox", "base64"] as const).map((k) => (
                  <div key={k}>
                    <p className="mb-1.5 text-[9px] font-bold uppercase tracking-wider text-fg-subtle/70">{k}</p>
                    <CopyField value={sub.data!.formats[k]} />
                  </div>
                ))}

                {deepLink.data?.deep_link && (
                  <div>
                    <p className="mb-1.5 text-[9px] font-bold uppercase tracking-wider text-fg-subtle/70">Deep link</p>
                    <CopyField value={deepLink.data.deep_link} />
                  </div>
                )}
              </div>
            </div>

            {sub.data.links.length > 0 ? (
              <div className="mt-auto flex min-h-0 flex-1 flex-col pt-4">
                <p className="mb-2 text-[10px] font-semibold text-fg-muted flex items-center gap-1">
                  <Wifi size={12} className="text-accent" />
                  {t("portal.configLinks")} ({sub.data.links.length})
                </p>
                <div className="min-h-0 flex-1 space-y-2 overflow-auto rounded-xl border border-border/40 bg-surface-2/30 p-2">
                  {sub.data.links.map((l, i) => (
                    <CopyField key={i} value={l} />
                  ))}
                </div>
              </div>
            ) : (
              <div className="flex-1" aria-hidden />
            )}
          </GlassCard>
        )}
      </div>
    </div>
  );
}
