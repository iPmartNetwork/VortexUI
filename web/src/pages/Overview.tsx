import { Users, Wifi, HardDrive, Activity, Cpu, MemoryStick, Server, Globe } from "lucide-react";
import { useOverview } from "@/api/policy-hooks";
import { useNodes } from "@/api/hooks";
import { Badge, Card, PageHeader, StatCard } from "@/components/ui";
import { useI18n } from "@/i18n/i18n";
import { formatBytes } from "@/lib/utils";

function Gauge({ label, value, icon }: { label: string; value: number; icon: React.ReactNode }) {
  const v = Math.min(100, Math.max(0, value));
  const color = v > 85 ? "bg-danger" : v > 60 ? "bg-warning" : "grad-bg";
  return (
    <div>
      <div className="mb-1 flex items-center justify-between text-xs">
        <span className="flex items-center gap-1.5 text-fg-muted">
          {icon}
          {label}
        </span>
        <span className="font-medium text-fg">{v.toFixed(0)}%</span>
      </div>
      <div className="h-1.5 overflow-hidden rounded-full bg-white/[0.06]">
        <div className={`h-full rounded-full ${color}`} style={{ width: `${v}%` }} />
      </div>
    </div>
  );
}

const STATUS_COLORS: Record<string, string> = {
  active: "bg-success",
  limited: "bg-warning",
  expired: "bg-danger",
  disabled: "bg-fg-subtle",
  on_hold: "bg-accent",
};

export function Overview() {
  const { data } = useOverview();
  const nodes = useNodes();
  const { t } = useI18n();
  const u = data?.users;
  const onlineCount = data?.nodes.online ?? 0;
  const totalNodes = data?.nodes.total ?? 0;
  const byStatus = u?.by_status ?? {};
  const statusTotal = Object.values(byStatus).reduce((a, b) => a + b, 0) || 1;

  return (
    <div>
      <PageHeader title={t("nav.overview")} subtitle={t("app.tagline")} />

      <div className="grid grid-cols-2 gap-4 lg:grid-cols-4">
        <StatCard label={t("nav.users")} value={u?.total ?? 0} accent="grad" icon={<Users size={18} />} />
        <StatCard label="Active" value={byStatus.active ?? 0} accent="success" icon={<Activity size={18} />} />
        <StatCard label={t("nav.nodes")} value={`${onlineCount} / ${totalNodes}`} accent="accent" icon={<Wifi size={18} />} />
        <StatCard label="Total traffic" value={formatBytes(u?.total_used ?? 0)} accent="plain" icon={<HardDrive size={18} />} />
      </div>

      <div className="mt-6 grid grid-cols-1 gap-4 lg:grid-cols-3">
        <div className="space-y-4 lg:col-span-2">
          <h2 className="text-sm font-semibold text-fg">{t("nav.nodes")}</h2>
          <div className="grid grid-cols-1 gap-4 md:grid-cols-2">
            {nodes.data?.nodes.map((n) => (
              <Card key={n.id} className="space-y-3.5">
                <div className="flex items-start justify-between">
                  <div className="flex items-center gap-2.5">
                    <div className="grid h-9 w-9 place-items-center rounded-xl bg-white/[0.05] text-fg-muted">
                      <Server size={17} />
                    </div>
                    <div>
                      <div className="flex items-center gap-2 text-sm font-semibold">
                        {n.name}
                        <span className={`h-2 w-2 rounded-full ${n.health.core_running ? "bg-success shadow-[0_0_8px] shadow-success" : "bg-fg-subtle"}`} />
                      </div>
                      <div className="mt-0.5 flex items-center gap-1 text-xs text-fg-subtle" dir="ltr">
                        <Globe size={11} /> {n.address}
                      </div>
                    </div>
                  </div>
                  <Badge color={n.health.core_running ? "running" : "down"}>
                    {n.health.core_running ? t("nodes.running") : t("nodes.down")}
                  </Badge>
                </div>

                <div className="space-y-2">
                  <Gauge label="CPU" value={n.health.cpu_percent} icon={<Cpu size={12} />} />
                  <Gauge label="RAM" value={n.health.mem_percent} icon={<MemoryStick size={12} />} />
                  <Gauge label="Disk" value={n.health.disk_percent} icon={<HardDrive size={12} />} />
                </div>

                <div className="flex flex-wrap items-center gap-x-4 gap-y-1 border-t border-white/[0.06] pt-3 text-xs text-fg-muted">
                  <span className="uppercase text-accent">{n.core}</span>
                  {n.core_version && <span>{n.core_version}</span>}
                  {n.agent_version && <span className="text-fg-subtle">agent {n.agent_version}</span>}
                  <span className="ms-auto font-medium text-fg">{n.health.connections} conns</span>
                </div>
              </Card>
            ))}
            {nodes.data?.nodes.length === 0 && <p className="text-sm text-fg-muted">{t("nodes.none")}</p>}
          </div>
        </div>

        <div className="space-y-4">
          <h2 className="text-sm font-semibold text-fg">{t("common.status")}</h2>
          <Card>
            <div className="mb-4 flex h-2.5 overflow-hidden rounded-full bg-white/[0.06]">
              {Object.entries(byStatus).map(([s, v]) =>
                v ? <div key={s} className={STATUS_COLORS[s] ?? "bg-fg-subtle"} style={{ width: `${(v / statusTotal) * 100}%` }} /> : null,
              )}
            </div>
            <div className="space-y-2.5">
              {["active", "limited", "expired", "disabled", "on_hold"].map((s) => (
                <div key={s} className="flex items-center justify-between">
                  <Badge color={s}>{s}</Badge>
                  <span className="text-sm font-medium text-fg">{byStatus[s] ?? 0}</span>
                </div>
              ))}
            </div>
          </Card>
        </div>
      </div>
    </div>
  );
}
