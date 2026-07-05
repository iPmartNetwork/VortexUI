import { useState } from "react";
import { Gauge, Users, HardDrive, TrendingUp, Clock, AlertTriangle, Download } from "lucide-react";
import { Link } from "react-router-dom";
import { exportResellerUsersCsv, useResellerDashboard } from "@/api/quota-hooks";
import { useAuth } from "@/auth/auth";
import { Badge, Button } from "@/components/ui";
import { GlassCard, StatsCard } from "@/components/veltrix";
import { useToast } from "@/components/toast";
import { useI18n } from "@/i18n/i18n";
import { useTitle } from "@/lib/useTitle";
import { formatBytes, pct } from "@/lib/utils";

function QuotaBar({ label, used, limit, format = "number" }: { label: string; used: number; limit: number; format?: "number" | "bytes" }) {
  const unlimited = limit <= 0;
  const displayUsed = format === "bytes" ? formatBytes(used, false) : String(used);
  const displayLimit = unlimited ? "∞" : format === "bytes" ? formatBytes(limit, false) : String(limit);
  const p = unlimited ? 0 : pct(used, limit);
  return (
    <GlassCard hover={false} className="!p-5 space-y-3">
      <div className="flex items-center justify-between text-sm">
        <span className="font-medium text-fg">{label}</span>
        <span className="text-fg-muted tabular-nums">{displayUsed} / {displayLimit}</span>
      </div>
      {!unlimited && (
        <div className="h-1.5 rounded-full bg-surface-3 overflow-hidden">
          <div
            className={`h-full rounded-full transition-all ${p >= 90 ? "bg-danger" : p >= 75 ? "bg-warning" : "bg-primary"}`}
            style={{ width: `${Math.min(p, 100)}%` }}
          />
        </div>
      )}
    </GlassCard>
  );
}

export function ResellerDashboard() {
  useTitle("Reseller Dashboard");
  const { session, sudo } = useAuth();
  const { t } = useI18n();
  const toast = useToast();
  const { data, isLoading } = useResellerDashboard();
  const [exporting, setExporting] = useState(false);
  const dash = data?.dashboard;
  const u = dash?.quota;

  async function exportCsv() {
    setExporting(true);
    try {
      exportResellerUsersCsv();
      toast.success(t("reseller.dashboard.exportDone"));
    } catch {
      toast.error(t("reseller.dashboard.exportFailed"));
    } finally {
      setExporting(false);
    }
  }

  if (sudo) {
    return (
      <div className="space-y-4 animate-page-enter">
        <h1 className="text-2xl font-bold text-fg tracking-tight">{t("reseller.dashboard.title")}</h1>
        <GlassCard hover={false} className="!p-6">
          <p className="text-sm text-fg-muted">{t("reseller.dashboard.sudoHint")}</p>
        </GlassCard>
      </div>
    );
  }

  const trafficMode = u?.traffic_quota_mode === "consumed" ? "consumed" : "allocated";
  const trafficUsed = trafficMode === "consumed" ? (u?.traffic_used ?? 0) : (u?.traffic_allocated ?? 0);
  const trafficLabel = trafficMode === "consumed" ? t("reseller.dashboard.trafficConsumed") : t("reseller.dashboard.trafficAssigned");

  return (
    <div className="space-y-5 animate-page-enter">
      <div className="flex flex-col lg:flex-row lg:items-start justify-between gap-4">
        <div>
          <h1 className="text-2xl font-bold text-fg tracking-tight">{t("reseller.dashboard.title")}</h1>
          <p className="text-sm text-fg-muted mt-1">
            {session?.admin?.username ? `@${session.admin.username}` : t("reseller.dashboard.subtitle")}
          </p>
        </div>
        <Button variant="outline" onClick={exportCsv} disabled={exporting} className="flex-shrink-0">
          <Download size={15} />
          {exporting ? t("reseller.dashboard.exporting") : t("reseller.dashboard.exportCsv")}
        </Button>
      </div>

      {isLoading && <p className="text-sm text-fg-muted text-center py-8">{t("reseller.dashboard.loading")}</p>}

      {dash && u && (
        <>
          <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
            <StatsCard
              title={t("reseller.dashboard.accounts")}
              value={u.user_count}
              suffix={u.user_quota > 0 ? `/ ${u.user_quota}` : undefined}
              icon={<Users size={18} />}
              color="blue"
            />
            <StatsCard
              title={t("reseller.dashboard.newUsers")}
              value={`${dash.new_users_7d} / ${dash.new_users_30d}`}
              icon={<TrendingUp size={18} />}
              color="green"
            />
            <StatsCard
              title={t("reseller.dashboard.expiring")}
              value={dash.expiring_soon}
              icon={<Clock size={18} />}
              color="orange"
            />
            <StatsCard
              title={t("reseller.dashboard.mode")}
              value={trafficMode === "consumed" ? "Consumed" : "Allocated"}
              icon={<Gauge size={18} />}
              color="cyan"
            />
          </div>

          <div className="grid gap-4 md:grid-cols-2">
            <QuotaBar label={t("reseller.dashboard.userAccounts")} used={u.user_count} limit={u.user_quota} />
            <QuotaBar label={trafficLabel} used={trafficUsed} limit={u.traffic_quota} format="bytes" />
          </div>

          <div className="grid gap-4 lg:grid-cols-2">
            <GlassCard hover={false} className="!p-0 overflow-hidden">
              <div className="border-b border-border/40 px-5 py-3 text-sm font-bold text-fg">{t("reseller.dashboard.usersByStatus")}</div>
              <div className="flex flex-wrap gap-2 p-4">
                {Object.entries(dash.users_by_status ?? {}).map(([status, count]) => (
                  <Badge key={status} color="muted">{status}: {count}</Badge>
                ))}
                {Object.keys(dash.users_by_status ?? {}).length === 0 && (
                  <p className="text-sm text-fg-muted">{t("reseller.dashboard.noUsers")}</p>
                )}
              </div>
            </GlassCard>

            <GlassCard hover={false} className="!p-0 overflow-hidden">
              <div className="border-b border-border/40 px-5 py-3 text-sm font-bold text-fg">{t("reseller.dashboard.topConsumers")}</div>
              <table className="w-full text-sm">
                <tbody>
                  {dash.top_users?.map((tu) => (
                    <tr key={tu.id} className="border-b border-border/20 last:border-0 hover:bg-surface-2/40">
                      <td className="px-5 py-2.5">
                        <Link to={`/users/${tu.id}`} className="font-medium text-fg hover:text-primary hover:underline">{tu.username}</Link>
                      </td>
                      <td className="px-5 py-2.5 text-fg-muted tabular-nums">{formatBytes(tu.used_traffic, false)}</td>
                      <td className="px-5 py-2.5 text-xs text-fg-subtle">{tu.status}</td>
                    </tr>
                  ))}
                  {(!dash.top_users || dash.top_users.length === 0) && (
                    <tr><td colSpan={3} className="px-5 py-4 text-fg-muted">{t("reseller.dashboard.noUsage")}</td></tr>
                  )}
                </tbody>
              </table>
            </GlassCard>
          </div>

          <div className="grid gap-4 sm:grid-cols-2">
            <GlassCard hover={false} className="!p-5 flex items-center gap-3">
              <div className="h-10 w-10 rounded-xl bg-surface-2 flex items-center justify-center text-fg-muted flex-shrink-0">
                <HardDrive size={18} />
              </div>
              <div className="text-sm">
                <div className="text-fg-subtle">{t("reseller.dashboard.assigned")}</div>
                <div className="font-semibold text-fg">{formatBytes(u.traffic_allocated, false)}</div>
              </div>
            </GlassCard>
            <GlassCard hover={false} className="!p-5 flex items-center gap-3">
              <div className="h-10 w-10 rounded-xl bg-surface-2 flex items-center justify-center text-fg-muted flex-shrink-0">
                <AlertTriangle size={18} />
              </div>
              <div className="text-sm">
                <div className="text-fg-subtle">{t("reseller.dashboard.consumed")}</div>
                <div className="font-semibold text-fg">{formatBytes(u.traffic_used, false)}</div>
              </div>
            </GlassCard>
          </div>
        </>
      )}
    </div>
  );
}
