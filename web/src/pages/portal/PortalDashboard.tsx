import { useQuery } from "@tanstack/react-query";
import { Activity, Calendar, HardDrive, RefreshCw, Smartphone, User } from "lucide-react";
import { portalApi } from "./portalApi";
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

function statusType(s: string): string {
  if (s === "limited") return "warning";
  if (s === "expired") return "error";
  if (s === "disabled") return "inactive";
  if (s === "on_hold") return "info";
  return "active";
}

export function PortalDashboard() {
  const { t } = useI18n();
  const { data, isLoading, error } = useQuery({
    queryKey: ["portal-dashboard"],
    queryFn: () => portalApi<DashboardData>("/api/portal/dashboard"),
    refetchInterval: 15000,
  });

  if (isLoading) {
    return <div className="py-12 text-center text-sm text-fg-muted">{t("common.loading")}</div>;
  }
  if (error || !data) {
    return <div className="py-12 text-center text-sm text-danger">{t("portal.loadFailed")}</div>;
  }

  const usagePercent =
    data.data_limit > 0 ? Math.min((data.used_traffic / data.data_limit) * 100, 100) : 0;

  return (
    <div className="space-y-6 animate-page-enter">
      <PageHeader
        title={t("portal.dashboardWelcome").replace("{name}", data.username)}
        subtitle={t("portal.dashboardSubtitle")}
      >
        <StatusBadge status={statusType(data.status)} label={data.status} pulse={data.status === "active"} />
      </PageHeader>

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
          label={t("portal.deviceLimit")}
          value={data.device_limit > 0 ? data.device_limit : "∞"}
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
        <StatCard
          label={t("portal.usage")}
          value={data.data_limit > 0 ? `${usagePercent.toFixed(0)}%` : "—"}
          icon={<Activity size={18} />}
          accent={usagePercent > 90 ? "warning" : "success"}
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

      <div className="rounded-2xl bg-bg-elevated border border-border p-5 space-y-2">
        <div className="text-[11px] font-semibold text-fg-subtle uppercase tracking-wider">
          {t("portal.subscriptionToken")}
        </div>
        <code className="block text-xs font-mono text-fg-muted break-all">{data.sub_token}</code>
      </div>
    </div>
  );
}
