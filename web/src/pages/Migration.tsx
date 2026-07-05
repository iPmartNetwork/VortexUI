import { useState } from "react";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { Info } from "lucide-react";
import { api } from "@/api/client";
import { Button, Input, Badge } from "@/components/ui";
import { GlassCard } from "@/components/veltrix";
import { useToast } from "@/components/toast";
import { useI18n } from "@/i18n/i18n";
import { useTitle } from "@/lib/useTitle";

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
  useTitle("Migration");
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
  const nodeMap = Object.fromEntries(nodesData?.nodes?.map((n) => [n.id, n.name]) ?? []);

  const [form, setForm] = useState<MigrationPolicy | null>(null);

  const policy = form ?? policyData?.policy;

  const save = useMutation({
    mutationFn: (p: MigrationPolicy) => api("/api/migration/policy", { method: "PUT", body: p }),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["migration-policy"] });
      toast.success("Policy saved");
    },
    onError: (e: unknown) => toast.error(e instanceof Error ? e.message : "failed"),
  });

  function update<K extends keyof MigrationPolicy>(field: K, value: MigrationPolicy[K]) {
    setForm((prev) => ({ ...(prev ?? policyData?.policy ?? ({} as MigrationPolicy)), [field]: value }));
  }

  return (
    <div className="space-y-5 animate-page-enter">
      <div>
        <h1 className="text-2xl font-bold text-fg tracking-tight">{t("migration.title")}</h1>
        <p className="text-sm text-fg-muted mt-1">{t("migration.subtitle")}</p>
      </div>

      <div className="rounded-xl border border-primary/25 bg-primary/5 px-4 py-3 flex items-start gap-3">
        <div className="h-8 w-8 rounded-full bg-primary/15 flex items-center justify-center text-primary flex-shrink-0">
          <Info size={16} />
        </div>
        <p className="text-xs text-fg-muted leading-relaxed">{t("migration.info")}</p>
      </div>

      {policy && (
        <GlassCard hover={false} className="!p-5 space-y-4">
          <div className="flex items-center justify-between">
            <h3 className="text-sm font-bold text-fg">{t("migration.policy")}</h3>
            <label className="flex items-center gap-2 text-sm text-fg">
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
              <label className="flex items-center gap-2 text-sm text-fg pb-2">
                <input type="checkbox" checked={policy.migrate_back} onChange={(e) => update("migrate_back", e.target.checked)} className="rounded" />
                {t("migration.migrateBack")}
              </label>
            </div>
          </div>
          <div className="flex justify-end pt-1 border-t border-border/40">
            <Button onClick={() => policy && save.mutate(policy)} disabled={save.isPending}>Save Policy</Button>
          </div>
        </GlassCard>
      )}

      <GlassCard hover={false} className="!p-0 overflow-hidden">
        <div className="px-4 pt-4 pb-1">
          <h3 className="text-sm font-bold text-fg">{t("migration.events")}</h3>
        </div>
        <div className="overflow-x-auto mt-2">
          <table className="w-full text-sm">
            <thead>
              <tr className="border-b border-border/40 text-[11px] uppercase tracking-wide text-fg-subtle bg-surface-2/30">
                <th className="py-3 px-4 text-left">Time</th>
                <th className="py-3 px-4 text-left">Reason</th>
                <th className="py-3 px-4 text-center">Status</th>
                <th className="py-3 px-4 text-left">From → To</th>
              </tr>
            </thead>
            <tbody>
              {eventsData?.events?.map((e) => (
                <tr key={e.id} className="border-b border-border/20 hover:bg-surface-2/40">
                  <td className="py-2.5 px-4 text-xs text-fg-muted">{new Date(e.created_at).toLocaleString()}</td>
                  <td className="py-2.5 px-4 text-xs text-fg">{e.reason}</td>
                  <td className="py-2.5 px-4 text-center">
                    <Badge color={e.status === "completed" ? "active" : e.status === "failed" ? "expired" : "limited"}>{e.status}</Badge>
                  </td>
                  <td className="py-2.5 px-4 text-xs font-mono text-fg-muted">{nodeMap[e.from_node_id] || e.from_node_id.slice(0, 8)} → {nodeMap[e.to_node_id] || e.to_node_id.slice(0, 8)}</td>
                </tr>
              ))}
              {(!eventsData?.events || eventsData.events.length === 0) && (
                <tr><td colSpan={4} className="py-6 text-center text-fg-muted text-xs">{t("migration.noEvents")}</td></tr>
              )}
            </tbody>
          </table>
        </div>
      </GlassCard>
    </div>
  );
}
