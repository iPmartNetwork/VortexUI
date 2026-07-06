import { useQuery } from "@tanstack/react-query";
import {
  Activity,
  AlertTriangle,
  CalendarClock,
  CreditCard,
  Database,
  MessageSquarePlus,
  Smartphone,
  Wifi,
} from "lucide-react";
import { Link } from "react-router-dom";
import { QRCodeSVG } from "qrcode.react";
import { portalApi } from "./portalApi";
import { CopyField } from "@/components/CopyField";
import { UsageChart } from "@/components/UsageChart";
import type { UsagePoint } from "@/api/hooks";
import { PageHeader } from "@/components/ui";
import { GlassCard, StatsCard, StatusBadge } from "@/components/veltrix";
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
    queryKey: ["portal-usage"],
    queryFn: () => portalApi<{ points: UsagePoint[] }>("/api/portal/usage?bucket=1d"),
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
    return <div className="py-12 text-center text-sm text-fg-muted">{t("common.loading")}</div>;
  }
  if (error || !data) {
    return <div className="py-12 text-center text-sm text-danger">{t("portal.loadFailed")}</div>;
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

  return (
    <div className="space-y-6 animate-page-enter">
      <PageHeader
        title={t("portal.dashboardWelcome").replace("{name}", data.username)}
        subtitle={t("portal.dashboardSubtitle")}
      >
        <div className="flex flex-wrap items-center gap-2">
          <StatusBadge status={statusType(data.status)} label={data.status} pulse={data.status === "active"} />
          <Link
            to="/portal/plans"
            className="inline-flex items-center gap-1.5 rounded-xl grad-bg px-3 py-2 text-xs font-semibold text-white transition hover:opacity-90"
          >
            <CreditCard size={14} />
            {t("portal.quickRenew")}
          </Link>
          <Link
            to="/portal/tickets"
            className="inline-flex items-center gap-1.5 rounded-xl border border-border bg-bg-elevated px-3 py-2 text-xs font-semibold text-fg transition hover:bg-surface-2"
          >
            <MessageSquarePlus size={14} />
            {t("portal.quickTicket")}
          </Link>
        </div>
      </PageHeader>

      {alerts.length > 0 && (
        <div className="space-y-2">
          {alerts.map((a, i) => (
            <div
              key={i}
              className={cn(
                "flex items-start gap-2 rounded-xl border px-4 py-3 text-sm",
                a.tone === "error" && "border-danger/30 bg-danger/10 text-danger",
                a.tone === "warning" && "border-warning/30 bg-warning/10 text-warning",
                a.tone === "info" && "border-accent/30 bg-accent/10 text-accent",
              )}
            >
              <AlertTriangle size={16} className="mt-0.5 shrink-0" />
              <span>{a.text}</span>
            </div>
          ))}
        </div>
      )}

      <div className="grid gap-6 xl:grid-cols-[minmax(0,1fr)_380px]">
        <div className="space-y-6 min-w-0">
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
            />
            <StatsCard
              title={t("portal.usage")}
              value={data.data_limit > 0 ? `${usagePercent.toFixed(0)}%` : "—"}
              subLabel={data.reset_strategy.replace("_", " ")}
              icon={<Activity size={18} />}
              color={usagePercent > 90 ? "red" : "green"}
            />
            <StatsCard
              title={t("portal.expires")}
              value={data.expire_at ? new Date(data.expire_at).toLocaleDateString() : t("common.never")}
              subLabel={`${t("portal.memberSince")} ${new Date(data.created_at).toLocaleDateString()}`}
              icon={<CalendarClock size={18} />}
              color="orange"
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
            />
          </div>

          {data.data_limit > 0 && (
            <GlassCard hover={false} className="!p-5">
              <div className="mb-3 flex items-center justify-between text-sm">
                <span className="font-medium text-fg">{t("portal.trafficConsumption")}</span>
                <span className="tabular-nums font-semibold text-fg">{usagePercent.toFixed(1)}%</span>
              </div>
              <div className="h-2.5 overflow-hidden rounded-full bg-surface-3">
                <div
                  className={cn(
                    "h-full rounded-full transition-all duration-500",
                    usagePercent > 90 ? "bg-danger" : usagePercent > 70 ? "bg-warning" : "grad-bg",
                  )}
                  style={{ width: `${usagePercent}%` }}
                />
              </div>
            </GlassCard>
          )}

          <GlassCard hover={false} className="!p-5">
            <div className="mb-4 flex flex-wrap items-center justify-between gap-2">
              <h3 className="text-xs font-semibold uppercase tracking-wider text-fg-subtle">
                {t("portal.dailyUsage")}
              </h3>
              {online.data?.live_tracking && (
                <span className="inline-flex items-center gap-1.5 text-xs text-fg-muted">
                  <Wifi size={13} className={liveConnections > 0 ? "text-success" : "text-fg-subtle"} />
                  {liveConnections} {t("portal.liveConnections").toLowerCase()}
                </span>
              )}
            </div>

            {usage.isLoading ? (
              <div className="h-40 animate-pulse rounded-xl bg-surface-2/50" />
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
              <div className="mt-5 overflow-x-auto rounded-xl border border-border/60">
                <table className="w-full text-sm">
                  <thead>
                    <tr className="border-b border-border bg-surface-2/40 text-fg-subtle">
                      <th className="px-4 py-2.5 text-start text-xs font-semibold uppercase tracking-wide">
                        {t("portal.day")}
                      </th>
                      <th className="px-4 py-2.5 text-end text-xs font-semibold uppercase tracking-wide">
                        {t("portal.chartUp")}
                      </th>
                      <th className="px-4 py-2.5 text-end text-xs font-semibold uppercase tracking-wide">
                        {t("portal.chartDown")}
                      </th>
                      <th className="px-4 py-2.5 text-end text-xs font-semibold uppercase tracking-wide">
                        {t("portal.total")}
                      </th>
                    </tr>
                  </thead>
                  <tbody>
                    {[...usagePoints].reverse().map((p) => (
                      <tr key={p.time} className="border-b border-border/40 last:border-0">
                        <td className="px-4 py-2.5 tabular-nums text-fg">
                          {new Date(p.time).toLocaleDateString()}
                        </td>
                        <td className="px-4 py-2.5 text-end tabular-nums text-fg-muted">
                          {formatBytes(p.up, false)}
                        </td>
                        <td className="px-4 py-2.5 text-end tabular-nums text-fg-muted">
                          {formatBytes(p.down, false)}
                        </td>
                        <td className="px-4 py-2.5 text-end tabular-nums font-semibold text-fg">
                          {formatBytes(p.up + p.down, false)}
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            )}
          </GlassCard>
        </div>

        {sub.data && (
          <GlassCard hover={false} className="!p-5 h-fit xl:sticky xl:top-6 space-y-4">
            <h3 className="text-xs font-semibold uppercase tracking-wider text-fg-subtle">
              {t("portal.subscriptionLink")}
            </h3>

            <div className="flex flex-col items-center gap-4">
              <div className="rounded-xl bg-white p-3 shadow-sm">
                <QRCodeSVG value={sub.data.subscription_url} size={148} />
              </div>

              <div className="w-full space-y-3">
                <div>
                  <p className="mb-1 text-[10px] font-bold uppercase tracking-wider text-fg-subtle">URL</p>
                  <CopyField value={sub.data.subscription_url} />
                </div>

                <div className="grid gap-3 sm:grid-cols-1">
                  {(["clash", "singbox", "base64"] as const).map((k) => (
                    <div key={k}>
                      <p className="mb-1 text-[10px] font-bold uppercase tracking-wider text-fg-subtle">{k}</p>
                      <CopyField value={sub.data!.formats[k]} />
                    </div>
                  ))}
                </div>

                {deepLink.data?.deep_link && (
                  <div>
                    <p className="mb-1 text-[10px] font-bold uppercase tracking-wider text-fg-subtle">Deep link</p>
                    <CopyField value={deepLink.data.deep_link} />
                  </div>
                )}
              </div>
            </div>

            {sub.data.links.length > 0 && (
              <div>
                <p className="mb-2 text-xs font-semibold text-fg-muted">
                  {t("portal.configLinks")} ({sub.data.links.length})
                </p>
                <div className="max-h-48 space-y-2 overflow-auto rounded-xl border border-border/50 bg-surface-2/20 p-2">
                  {sub.data.links.map((l, i) => (
                    <CopyField key={i} value={l} />
                  ))}
                </div>
              </div>
            )}
          </GlassCard>
        )}
      </div>
    </div>
  );
}
