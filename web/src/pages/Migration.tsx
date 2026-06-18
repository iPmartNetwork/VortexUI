import { useState } from "react";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { api } from "@/api/client";
import { Button, Card, Input, PageHeader, Badge } from "@/components/ui";
import { useToast } from "@/components/toast";
import { useI18n } from "@/i18n/i18n";

interface MigrationPolicy {
  enabled: boolean;
  health_check_interval: number;
  unhealthy_threshold: number;
  cpu_threshold: number;
  mem_threshold: number;
  packet_loss_max: number;
  migrate_back: boolean;
}

interface MigrationEvent {
  id: string;
  user_id: string;
  username: string;
  from_node_id: string;
  to_node_id: string;
  reason: string;
  status: string;
  error: string;
  created_at: string;
}

export function Migration() {
  const { t } = useI18n();
  const qc = useQueryClient();
  const toast = useToast();

  const { data: policyData } = useQuery({
    queryKey: ["migration-policy"],
    queryFn: () => api<{ policy: MigrationPolicy }>("/api/migration/policy"),
  });

  const { data: eventsData } = useQuery({
    queryKey: ["migration-events"],
    queryFn: () => api<{ events: MigrationEvent[]; total: number }>("/api/migration/events"),
  });

  const { data: nodesData } = useQuery({ queryKey: ["nodes"], queryFn: () => api<{ nodes: { id: string; name: string }[] }>("/api/nodes") });
  const nodeMap = Object.fromEntries(nodesData?.nodes?.map(n => [n.id, n.name]) ?? []);

  const [form, setForm] = useState<MigrationPolicy | null>(null);

  const policy = form ?? policyData?.policy;

  const save = useMutation({
    mutationFn: (p: MigrationPolicy) => api("/api/migration/policy", { method: "PUT", body: p }),
    onSuccess: () => { qc.invalidateQueries({ queryKey: ["migration-policy"] }); toast.success("Policy saved"); },
    onError: (e: any) => toast.error(e.message),
  });

  function update(field: keyof MigrationPolicy, value: any) {
    setForm(prev => ({ ...(prev ?? policyData?.policy ?? {} as any), [field]: value }));
  }

  return (
    <div className="space-y-6 animate-fade-in">
      <PageHeader title={t("migration.title")} subtitle={t("migration.subtitle")} />

      <div className="rounded-lg border border-border/40 bg-surface-2/20 p-4 text-xs text-fg-muted">
        <p>When a node becomes unhealthy (high CPU, memory, or packet loss beyond thresholds), users are automatically migrated to a healthy node. If "Migrate back" is enabled, users return when the original node recovers.</p>
      </div>

      {policy && (
        <Card className="space-y-4">
          <div className="flex items-center justify-between">
            <h3 className="text-sm font-bold text-fg">{t("migration.policy")}</h3>
            <label className="flex items-center gap-2 text-sm">
              <input type="checkbox" checked={policy.enabled} onChange={(e) => update("enabled", e.target.checked)} className="rounded" />
              Enabled
            </label>
          </div>
          <div className="grid grid-cols-2 gap-3 md:grid-cols-3">
            <div>
              <label className="text-xs text-fg-subtle">{t("migration.healthInterval")}</label>
              <Input value={policy.health_check_interval} onChange={(e) => update("health_check_interval", Number(e.target.value))} inputMode="numeric" />
            </div>
            <div>
              <label className="text-xs text-fg-subtle">{t("migration.threshold")}</label>
              <Input value={policy.unhealthy_threshold} onChange={(e) => update("unhealthy_threshold", Number(e.target.value))} inputMode="numeric" />
            </div>
            <div>
              <label className="text-xs text-fg-subtle">{t("migration.cpuThreshold")}</label>
              <Input value={policy.cpu_threshold} onChange={(e) => update("cpu_threshold", Number(e.target.value))} inputMode="numeric" />
            </div>
            <div>
              <label className="text-xs text-fg-subtle">{t("migration.memThreshold")}</label>
              <Input value={policy.mem_threshold} onChange={(e) => update("mem_threshold", Number(e.target.value))} inputMode="numeric" />
            </div>
            <div>
              <label className="text-xs text-fg-subtle">{t("migration.packetLoss")}</label>
              <Input value={policy.packet_loss_max} onChange={(e) => update("packet_loss_max", Number(e.target.value))} inputMode="numeric" />
            </div>
            <div className="flex items-end">
              <label className="flex items-center gap-2 text-sm pb-2">
                <input type="checkbox" checked={policy.migrate_back} onChange={(e) => update("migrate_back", e.target.checked)} className="rounded" />
                {t("migration.migrateBack")}
              </label>
            </div>
          </div>
          <div className="flex justify-end">
            <Button onClick={() => policy && save.mutate(policy)} disabled={save.isPending}>Save Policy</Button>
          </div>
        </Card>
      )}

      <Card>
        <h3 className="text-sm font-bold text-fg mb-3">{t("migration.events")}</h3>
        <div className="overflow-x-auto">
          <table className="w-full text-sm">
            <thead>
              <tr className="border-b border-border/40 text-xs text-fg-subtle">
                <th className="py-2 text-left">Time</th>
                <th className="py-2 text-left">Reason</th>
                <th className="py-2 text-center">Status</th>
                <th className="py-2 text-left">From → To</th>
              </tr>
            </thead>
            <tbody>
              {eventsData?.events?.map((e) => (
                <tr key={e.id} className="border-b border-border/20">
                  <td className="py-2 text-xs text-fg-muted">{new Date(e.created_at).toLocaleString()}</td>
                  <td className="py-2 text-xs text-fg">{e.reason}</td>
                  <td className="py-2 text-center">
                    <Badge color={e.status === "completed" ? "active" : e.status === "failed" ? "expired" : "limited"}>{e.status}</Badge>
                  </td>
                  <td className="py-2 text-xs font-mono text-fg-muted">{nodeMap[e.from_node_id] || e.from_node_id.slice(0, 8)} → {nodeMap[e.to_node_id] || e.to_node_id.slice(0, 8)}</td>
                </tr>
              ))}
              {(!eventsData?.events || eventsData.events.length === 0) && (
                <tr><td colSpan={4} className="py-6 text-center text-fg-muted text-xs">{t("migration.noEvents")}</td></tr>
              )}
            </tbody>
          </table>
        </div>
      </Card>
    </div>
  );
}
