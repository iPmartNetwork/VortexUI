import { useState } from "react";
import { Gauge, Users, HardDrive, TrendingUp, Clock, AlertTriangle, Download } from "lucide-react";
import { Link } from "react-router-dom";
import { exportResellerUsersCsv, useResellerDashboard } from "@/api/quota-hooks";
import { useAuth } from "@/auth/auth";
import { Badge, Button, Card, PageHeader } from "@/components/ui";
import { useToast } from "@/components/toast";
import { useI18n } from "@/i18n/i18n";
import { formatBytes, pct } from "@/lib/utils";

function QuotaBar({ label, used, limit, format = "number" }: { label: string; used: number; limit: number; format?: "number" | "bytes" }) {
  const unlimited = limit <= 0;
  const displayUsed = format === "bytes" ? formatBytes(used, false) : String(used);
  const displayLimit = unlimited ? "∞" : format === "bytes" ? formatBytes(limit, false) : String(limit);
  const p = unlimited ? 0 : pct(used, limit);
  return (
    <Card className="space-y-3 p-5">
      <div className="flex items-center justify-between text-sm">
        <span className="font-medium">{label}</span>
        <span className="text-muted-foreground">{displayUsed} / {displayLimit}</span>
      </div>
      {!unlimited && (
        <div className="h-2 rounded-full bg-muted">
          <div className={`h-full rounded-full transition-all ${p >= 90 ? "bg-destructive" : p >= 75 ? "bg-amber-500" : "bg-primary"}`} style={{ width: `${Math.min(p, 100)}%` }} />
        </div>
      )}
    </Card>
  );
}

export function ResellerDashboard() {
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
      <div className="space-y-4">
        <PageHeader title={t("reseller.dashboard.title")} subtitle={t("reseller.dashboard.sudoHint")} />
        <Card className="p-6 text-sm text-muted-foreground">{t("reseller.dashboard.sudoHint")}</Card>
      </div>
    );
  }

  const trafficMode = u?.traffic_quota_mode === "consumed" ? "consumed" : "allocated";
  const trafficUsed = trafficMode === "consumed" ? (u?.traffic_used ?? 0) : (u?.traffic_allocated ?? 0);
  const trafficLabel = trafficMode === "consumed" ? t("reseller.dashboard.trafficConsumed") : t("reseller.dashboard.trafficAssigned");

  return (
    <div className="space-y-6 animate-page-enter">
      <div className="flex flex-wrap items-end justify-between gap-3">
        <PageHeader
          title={t("reseller.dashboard.title")}
          subtitle={session?.admin?.username ? `@${session.admin.username}` : t("reseller.dashboard.subtitle")}
        />
        <Button variant="outline" onClick={exportCsv} disabled={exporting}>
          <Download size={15} />
          {exporting ? t("reseller.dashboard.exporting") : t("reseller.dashboard.exportCsv")}
        </Button>
      </div>

      {isLoading && <p className="text-sm text-muted-foreground">{t("reseller.dashboard.loading")}</p>}

      {dash && u && (
        <>
          <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
            <Card className="flex items-center gap-3 p-5">
              <Users className="text-primary" size={22} />
              <div>
                <div className="text-xs text-muted-foreground">{t("reseller.dashboard.accounts")}</div>
                <div className="text-xl font-bold">{u.user_count}{u.user_quota > 0 ? ` / ${u.user_quota}` : ""}</div>
              </div>
            </Card>
            <Card className="flex items-center gap-3 p-5">
              <TrendingUp className="text-accent" size={22} />
              <div>
                <div className="text-xs text-muted-foreground">{t("reseller.dashboard.newUsers")}</div>
                <div className="text-xl font-bold">{dash.new_users_7d} / {dash.new_users_30d}</div>
              </div>
            </Card>
            <Card className="flex items-center gap-3 p-5">
              <Clock className="text-amber-500" size={22} />
              <div>
                <div className="text-xs text-muted-foreground">{t("reseller.dashboard.expiring")}</div>
                <div className="text-xl font-bold">{dash.expiring_soon}</div>
              </div>
            </Card>
            <Card className="flex items-center gap-3 p-5">
              <Gauge className="text-success" size={22} />
              <div>
                <div className="text-xs text-muted-foreground">{t("reseller.dashboard.mode")}</div>
                <div className="text-sm font-semibold capitalize">{trafficMode}</div>
              </div>
            </Card>
          </div>

          <div className="grid gap-4 md:grid-cols-2">
            <QuotaBar label={t("reseller.dashboard.userAccounts")} used={u.user_count} limit={u.user_quota} />
            <QuotaBar label={trafficLabel} used={trafficUsed} limit={u.traffic_quota} format="bytes" />
          </div>

          <div className="grid gap-4 lg:grid-cols-2">
            <Card className="p-0">
              <div className="border-b px-5 py-3 text-sm font-semibold">{t("reseller.dashboard.usersByStatus")}</div>
              <div className="flex flex-wrap gap-2 p-4">
                {Object.entries(dash.users_by_status ?? {}).map(([status, count]) => (
                  <Badge key={status}>{status}: {count}</Badge>
                ))}
                {Object.keys(dash.users_by_status ?? {}).length === 0 && (
                  <p className="text-sm text-muted-foreground">{t("reseller.dashboard.noUsers")}</p>
                )}
              </div>
            </Card>

            <Card className="p-0">
              <div className="border-b px-5 py-3 text-sm font-semibold">{t("reseller.dashboard.topConsumers")}</div>
              <table className="w-full text-sm">
                <tbody>
                  {dash.top_users?.map((tu) => (
                    <tr key={tu.id} className="border-b last:border-0">
                      <td className="px-5 py-2">
                        <Link to={`/users/${tu.id}`} className="font-medium hover:underline">{tu.username}</Link>
                      </td>
                      <td className="px-5 py-2 text-muted-foreground">{formatBytes(tu.used_traffic, false)}</td>
                      <td className="px-5 py-2 text-xs text-muted-foreground">{tu.status}</td>
                    </tr>
                  ))}
                  {(!dash.top_users || dash.top_users.length === 0) && (
                    <tr><td colSpan={3} className="px-5 py-4 text-muted-foreground">{t("reseller.dashboard.noUsage")}</td></tr>
                  )}
                </tbody>
              </table>
            </Card>
          </div>

          <div className="grid gap-4 sm:grid-cols-2">
            <Card className="flex items-center gap-3 p-5">
              <HardDrive size={20} className="text-muted-foreground" />
              <div className="text-sm">
                <div className="text-muted-foreground">{t("reseller.dashboard.assigned")}</div>
                <div className="font-semibold">{formatBytes(u.traffic_allocated, false)}</div>
              </div>
            </Card>
            <Card className="flex items-center gap-3 p-5">
              <AlertTriangle size={20} className="text-muted-foreground" />
              <div className="text-sm">
                <div className="text-muted-foreground">{t("reseller.dashboard.consumed")}</div>
                <div className="font-semibold">{formatBytes(u.traffic_used, false)}</div>
              </div>
            </Card>
          </div>
        </>
      )}
    </div>
  );
}
