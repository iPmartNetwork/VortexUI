import { useQuery } from "@tanstack/react-query";
import {
  Activity,
  AlertTriangle,
  Calendar,
  CreditCard,
  HardDrive,
  MessageSquarePlus,
  RefreshCw,
  Smartphone,
  User,
  Wifi,
} from "lucide-react";
import { Link } from "react-router-dom";
import { QRCodeSVG } from "qrcode.react";
import { portalApi } from "./portalApi";
import { CopyField } from "@/components/CopyField";
import { UsageChart } from "@/components/UsageChart";
import type { UsagePoint } from "@/api/hooks";
import { PageHeader, StatCard } from "@/components/ui";
import { StatusBadge } from "@/components/veltrix";
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

  const alerts: { tone: "warning" | "error" | "info"; text: string }[] = [];
  if (data.status === "expired") {
    alerts.push({ tone: "error", text: t("portal.alert.expired") });
  } else if (data.status === "limited") {
    alerts.push({ tone: "warning", text: t("portal.alert.limited") });
  } else if (data.status === "disabled") {
    alerts.push({ tone: "error", text: t("portal.alert.disabled") });
  }
  if (data.data_limit > 0 && usagePercent >= 90) {
    alerts.push({ tone: "warning", text: t("portal.alert.quotaHigh").replace("{pct}", usagePercent.toFixed(0)) });
  }
  if (data.expire_at) {
    const days = daysUntil(data.expire_at);
    if (days <= 0 && data.status !== "expired") {
      alerts.push({ tone: "error", text: t("portal.alert.expired") });
    } else if (days > 0 && days <= 7) {
      alerts.push({ tone: "warning", text: t("portal.alert.expiresSoon").replace("{days}", String(days)) });
    }
  }

  const deviceLimitLabel =
    data.device_limit > 0 ? String(data.device_limit) : "∞";
  const activeDevices = online.data?.active_devices ?? 0;
  const liveConnections = online.data?.live_connections ?? 0;

  const usagePoints = usage.data?.points ?? [];

  return (
    <div className="space-y-6 animate-page-enter">
      <PageHeader
        title={t("portal.dashboardWelcome").replace("{name}", data.username)}
        subtitle={t("portal.dashboardSubtitle")}
      >
        <StatusBadge status={statusType(data.status)} label={data.status} pulse={data.status === "active"} />
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

      <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
        <StatCard
          label={t("portal.liveConnections")}
          value={online.isLoading ? "…" : String(liveConnections)}
          sub={online.data?.live_tracking ? t("portal.liveNow") : t("portal.liveUnavailable")}
          icon={<Wifi size={18} />}
          accent="accent"
        />
        <StatCard
          label={t("portal.activeDevices")}
          value={online.isLoading ? "…" : `${activeDevices} / ${deviceLimitLabel}`}
          sub={online.data?.device_tracking ? t("portal.devicesOnline") : t("portal.devicesUnavailable")}
          icon={<Smartphone size={18} />}
          accent={data.device_limit > 0 && activeDevices >= data.device_limit ? "warning" : "plain"}
        />
      </div>

      <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3">
        <StatCard
          label={t("portal.dataUsed")}
          value={formatBytes(data.used_traffic, false)}
          sub={
            data.data_limit > 0
              ? `${t("portal.ofLimit")} ${formatBytes(data.data_limit, false)}`
              : t("portal.unlimitedPlan")
          }
          icon={<HardDrive size={18} />}
          accent="grad"
        />
        <StatCard
          label={t("portal.expires")}
          value={data.expire_at ? new Date(data.expire_at).toLocaleDateString() : t("common.never")}
          icon={<Calendar size={18} />}
          accent="accent"
        />
        <StatCard
          label={t("portal.usage")}
          value={data.data_limit > 0 ? `${usagePercent.toFixed(0)}%` : "—"}
          icon={<Activity size={18} />}
          accent={usagePercent > 90 ? "warning" : "success"}
        />
        <StatCard
          label={t("portal.deviceLimit")}
          value={deviceLimitLabel}
          icon={<Smartphone size={18} />}
          accent="plain"
        />
        <StatCard
          label={t("portal.resetStrategy")}
          value={data.reset_strategy.replace("_", " ")}
          icon={<RefreshCw size={18} />}
          accent="plain"
        />
        <StatCard
          label={t("portal.memberSince")}
          value={new Date(data.created_at).toLocaleDateString()}
          icon={<User size={18} />}
          accent="plain"
        />
      </div>

      {data.data_limit > 0 && (
        <div className="rounded-2xl bg-bg-elevated border border-border p-5">
          <div className="flex justify-between text-xs text-fg-muted mb-2">
            <span>{t("portal.trafficConsumption")}</span>
            <span className="tabular-nums">{usagePercent.toFixed(1)}%</span>
          </div>
          <div className="h-2 rounded-full bg-surface-3 overflow-hidden">
            <div
              className={cn(
                "h-full rounded-full transition-all duration-500",
                usagePercent > 90 ? "bg-danger" : usagePercent > 70 ? "bg-warning" : "grad-bg",
              )}
              style={{ width: `${usagePercent}%` }}
            />
          </div>
        </div>
      )}

      <div className="rounded-2xl bg-bg-elevated border border-border p-5 space-y-4">
        <h3 className="text-[11px] font-semibold text-fg-subtle uppercase tracking-wider">
          {t("portal.dailyUsage")}
        </h3>
        <UsageChart
          points={usagePoints}
          labels={{
            empty: t("portal.noTrafficYet"),
            up: t("portal.chartUp"),
            down: t("portal.chartDown"),
            peak: t("portal.chartPeak"),
          }}
        />
        {usagePoints.length > 0 && (
          <div className="overflow-x-auto">
            <table className="w-full text-xs">
              <thead>
                <tr className="border-b border-border text-fg-subtle">
                  <th className="py-2 text-start font-medium">{t("portal.day")}</th>
                  <th className="py-2 text-end font-medium">{t("portal.chartUp")}</th>
                  <th className="py-2 text-end font-medium">{t("portal.chartDown")}</th>
                  <th className="py-2 text-end font-medium">{t("portal.total")}</th>
                </tr>
              </thead>
              <tbody>
                {[...usagePoints].reverse().map((p) => (
                  <tr key={p.time} className="border-b border-border/50 text-fg-muted">
                    <td className="py-2 tabular-nums">{new Date(p.time).toLocaleDateString()}</td>
                    <td className="py-2 text-end tabular-nums">{formatBytes(p.up, false)}</td>
                    <td className="py-2 text-end tabular-nums">{formatBytes(p.down, false)}</td>
                    <td className="py-2 text-end tabular-nums font-medium text-fg">
                      {formatBytes(p.up + p.down, false)}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </div>

      {sub.data && (
        <div className="rounded-2xl bg-bg-elevated border border-border p-5 space-y-4">
          <h3 className="text-[11px] font-semibold text-fg-subtle uppercase tracking-wider">
            {t("portal.subscriptionLink")}
          </h3>
          <div className="flex flex-col items-center gap-4 sm:flex-row">
            <div className="rounded-xl bg-white p-3">
              <QRCodeSVG value={sub.data.subscription_url} size={120} />
            </div>
            <div className="flex-1 space-y-2 w-full">
              <CopyField value={sub.data.subscription_url} />
              <div className="grid grid-cols-1 gap-2 sm:grid-cols-2">
                {(["clash", "singbox", "base64"] as const).map((k) => (
                  <CopyField key={k} value={sub.data!.formats[k]} />
                ))}
              </div>
              {deepLink.data?.deep_link && (
                <CopyField value={deepLink.data.deep_link} />
              )}
            </div>
          </div>
          {sub.data.links.length > 0 && (
            <div>
              <p className="mb-1.5 text-xs font-medium text-fg-muted">
                {t("portal.configLinks")} ({sub.data.links.length})
              </p>
              <div className="max-h-32 space-y-1.5 overflow-auto">
                {sub.data.links.map((l, i) => (
                  <CopyField key={i} value={l} />
                ))}
              </div>
            </div>
          )}
        </div>
      )}

      <div className="flex flex-wrap gap-3">
        <Link
          to="/portal/plans"
          className="inline-flex items-center gap-2 rounded-xl grad-bg px-4 py-2.5 text-sm font-medium text-white transition hover:opacity-90"
        >
          <CreditCard size={16} />
          {t("portal.quickRenew")}
        </Link>
        <Link
          to="/portal/tickets"
          className="inline-flex items-center gap-2 rounded-xl border border-border bg-bg-elevated px-4 py-2.5 text-sm font-medium text-fg transition hover:bg-surface-2"
        >
          <MessageSquarePlus size={16} />
          {t("portal.quickTicket")}
        </Link>
      </div>
    </div>
  );
}
