import { Users, Wifi, HardDrive, Activity } from "lucide-react";
import { useOverview } from "@/api/policy-hooks";
import { Badge, Card, PageHeader, StatCard } from "@/components/ui";
import { useI18n } from "@/i18n/i18n";
import { formatBytes } from "@/lib/utils";

export function Overview() {
  const { data, isLoading } = useOverview();
  const { t } = useI18n();
  const u = data?.users;
  const n = data?.nodes;

  return (
    <div>
      <PageHeader title={t("nav.overview")} subtitle={t("app.tagline")} />

      {isLoading && <p className="text-sm text-fg-muted">{t("common.loading")}</p>}

      <div className="grid grid-cols-2 gap-4 lg:grid-cols-4">
        <StatCard label={t("nav.users")} value={u?.total ?? 0} accent="grad" icon={<Users size={18} />} />
        <StatCard
          label="Active"
          value={u?.by_status?.active ?? 0}
          accent="success"
          icon={<Activity size={18} />}
        />
        <StatCard
          label={t("nav.nodes")}
          value={`${n?.online ?? 0} / ${n?.total ?? 0}`}
          accent="accent"
          icon={<Wifi size={18} />}
        />
        <StatCard
          label="Traffic"
          value={formatBytes(u?.total_used ?? 0)}
          accent="plain"
          icon={<HardDrive size={18} />}
        />
      </div>

      <div className="mt-6 grid grid-cols-1 gap-4 lg:grid-cols-3">
        <Card className="lg:col-span-2">
          <h2 className="mb-3 text-sm font-semibold text-fg">{t("nav.nodes")}</h2>
          <div className="space-y-2">
            {n?.items.map((node) => (
              <div
                key={node.id}
                className="flex items-center justify-between rounded-xl border border-white/[0.06] bg-white/[0.02] px-4 py-3"
              >
                <div className="flex items-center gap-3">
                  <span
                    className={`h-2 w-2 rounded-full ${node.online ? "bg-success shadow-[0_0_8px] shadow-success" : "bg-fg-subtle"}`}
                  />
                  <span className="text-sm font-medium">{node.name}</span>
                  <span className="text-xs uppercase text-fg-subtle">{node.core}</span>
                </div>
                <div className="flex items-center gap-4 text-xs text-fg-muted">
                  <span>CPU {node.health.cpu_percent.toFixed(0)}%</span>
                  <span>{node.health.connections} conns</span>
                  <Badge color={node.online ? "running" : "down"}>
                    {node.online ? t("nodes.running") : t("nodes.down")}
                  </Badge>
                </div>
              </div>
            ))}
            {n?.items.length === 0 && <p className="text-sm text-fg-muted">{t("nodes.none")}</p>}
          </div>
        </Card>

        <Card>
          <h2 className="mb-3 text-sm font-semibold text-fg">{t("common.status")}</h2>
          <div className="space-y-2.5">
            {["active", "limited", "expired", "disabled", "on_hold"].map((s) => (
              <div key={s} className="flex items-center justify-between">
                <Badge color={s}>{s}</Badge>
                <span className="text-sm font-medium text-fg">{u?.by_status?.[s] ?? 0}</span>
              </div>
            ))}
          </div>
        </Card>
      </div>
    </div>
  );
}
