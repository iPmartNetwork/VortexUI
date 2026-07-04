import { Gauge, Users, HardDrive } from "lucide-react";
import { useAccountQuota } from "@/api/quota-hooks";
import { useAuth } from "@/auth/auth";
import { Card, PageHeader } from "@/components/ui";
import { useI18n } from "@/i18n/i18n";
import { formatBytes, pct } from "@/lib/utils";

function QuotaBar({ label, used, limit, remainingLabel, format = "number" }: {
  label: string; used: number; limit: number; remainingLabel: string; format?: "number" | "bytes";
}) {
  const unlimited = limit <= 0;
  const displayUsed = format === "bytes" ? formatBytes(used, false) : String(used);
  const displayLimit = unlimited ? "∞" : format === "bytes" ? formatBytes(limit, false) : String(limit);
  const displayRem = unlimited ? "∞" : format === "bytes" ? formatBytes(Math.max(0, limit - used), false) : String(Math.max(0, limit - used));
  const p = unlimited ? 0 : pct(used, limit);
  return (
    <Card className="space-y-3 p-5">
      <div className="flex items-center justify-between text-sm">
        <span className="font-medium">{label}</span>
        <span className="text-muted-foreground">{displayUsed} / {displayLimit}</span>
      </div>
      {!unlimited && (
        <div className="h-2 rounded-full bg-muted">
          <div className="h-full rounded-full bg-primary transition-all" style={{ width: `${p}%` }} />
        </div>
      )}
      <p className="text-xs text-muted-foreground">{remainingLabel}: {displayRem}</p>
    </Card>
  );
}

export function MyQuota() {
  const { session, sudo } = useAuth();
  const { t } = useI18n();
  const { data, isLoading } = useAccountQuota();
  const admin = session?.admin;

  if (sudo) {
    return (
      <div className="space-y-4">
        <PageHeader title={t("reseller.quota.title")} subtitle={t("reseller.quota.sudoHint")} />
        <Card className="p-6 text-sm text-muted-foreground">{t("reseller.quota.sudoHint")}</Card>
      </div>
    );
  }

  const u = data?.usage;

  return (
    <div className="space-y-6 animate-page-enter">
      <PageHeader
        title={t("reseller.quota.title")}
        subtitle={admin?.username ? `${t("reseller.quota.subtitle")} · ${admin.username}` : t("reseller.quota.subtitle")}
      />

      {isLoading && <p className="text-sm text-muted-foreground">{t("reseller.dashboard.loading")}</p>}

      {u && (
        <>
          <div className="grid gap-4 md:grid-cols-3">
            <Card className="flex items-center gap-3 p-5">
              <Users className="text-primary" size={22} />
              <div>
                <div className="text-xs text-muted-foreground">{t("reseller.dashboard.accounts")}</div>
                <div className="text-xl font-bold">{u.user_count}{u.user_quota > 0 ? ` / ${u.user_quota}` : ""}</div>
              </div>
            </Card>
            <Card className="flex items-center gap-3 p-5">
              <HardDrive className="text-accent" size={22} />
              <div>
                <div className="text-xs text-muted-foreground">{t("reseller.dashboard.trafficAssigned")}</div>
                <div className="text-xl font-bold">{formatBytes(u.traffic_allocated, false)}</div>
              </div>
            </Card>
            <Card className="flex items-center gap-3 p-5">
              <Gauge className="text-success" size={22} />
              <div>
                <div className="text-xs text-muted-foreground">{t("reseller.dashboard.trafficConsumed")}</div>
                <div className="text-xl font-bold">{formatBytes(u.traffic_used, false)}</div>
              </div>
            </Card>
          </div>

          <div className="grid gap-4 md:grid-cols-2">
            <QuotaBar label={t("reseller.dashboard.userAccounts")} used={u.user_count} limit={u.user_quota} remainingLabel={t("reseller.quota.remaining")} />
            <QuotaBar label={t("reseller.quota.trafficPool")} used={u.traffic_allocated} limit={u.traffic_quota} remainingLabel={t("reseller.quota.remaining")} format="bytes" />
          </div>

          <Card className="p-4 text-xs text-muted-foreground">{t("reseller.quota.hint")}</Card>
        </>
      )}
    </div>
  );
}
