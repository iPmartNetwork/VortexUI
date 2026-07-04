import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { api } from "@/api/client";
import { Button, Card, Input, PageHeader } from "@/components/ui";
import { useToast } from "@/components/toast";
import { useI18n } from "@/i18n/i18n";

interface AdminQuotaNotifyConfig {
  enabled: boolean;
  notify_telegram: boolean;
  webhook_url: string;
  notify_at_percent: number[];
  cooldown_minutes: number;
}

export function ResellerQuotaAlerts() {
  const { t } = useI18n();
  const qc = useQueryClient();
  const toast = useToast();
  const { data } = useQuery({
    queryKey: ["admin-quota-notify-config"],
    queryFn: () => api<{ config: AdminQuotaNotifyConfig }>("/api/admin-quota-notify/config"),
  });
  const events = useQuery({
    queryKey: ["admin-quota-notify-events"],
    queryFn: () => api<{ events: { admin_id: string; threshold: number; metric: string; usage_pct: number; created_at: string }[] }>("/api/admin-quota-notify/events"),
  });

  const save = useMutation({
    mutationFn: (cfg: AdminQuotaNotifyConfig) =>
      api("/api/admin-quota-notify/config", { method: "PUT", body: cfg }),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["admin-quota-notify-config"] });
      toast.success(t("reseller.alerts.saved"));
    },
  });

  const cfg = data?.config;
  if (!cfg) return <p className="text-sm text-muted-foreground">{t("common.loading")}</p>;

  return (
    <div className="space-y-6 animate-page-enter">
      <PageHeader title={t("reseller.alerts.title")} subtitle={t("reseller.alerts.subtitle")} />

      <Card className="space-y-4 p-5">
        <label className="flex items-center gap-2 text-sm">
          <input type="checkbox" checked={cfg.enabled} onChange={(e) => save.mutate({ ...cfg, enabled: e.target.checked })} />
          {t("common.enabled")}
        </label>
        <label className="flex items-center gap-2 text-sm">
          <input type="checkbox" checked={cfg.notify_telegram} onChange={(e) => save.mutate({ ...cfg, notify_telegram: e.target.checked })} />
          {t("reseller.alerts.telegram")}
        </label>
        <Input
          placeholder={t("reseller.alerts.webhookUrl")}
          defaultValue={cfg.webhook_url}
          onBlur={(e) => save.mutate({ ...cfg, webhook_url: e.target.value })}
        />
        <Input
          placeholder={t("reseller.alerts.thresholds")}
          defaultValue={(cfg.notify_at_percent ?? [80, 90, 100]).join(", ")}
          onBlur={(e) => {
            const pcts = e.target.value.split(",").map((s) => Number(s.trim())).filter((n) => n > 0);
            save.mutate({ ...cfg, notify_at_percent: pcts.length ? pcts : [80, 90, 100] });
          }}
        />
        <Input
          placeholder={t("reseller.alerts.cooldown")}
          type="number"
          defaultValue={cfg.cooldown_minutes}
          onBlur={(e) => save.mutate({ ...cfg, cooldown_minutes: Number(e.target.value) || 1440 })}
        />
        <Button onClick={() => save.mutate(cfg)} disabled={save.isPending}>{t("common.save")}</Button>
      </Card>

      <Card className="p-0">
        <div className="border-b px-5 py-3 text-sm font-semibold">{t("reseller.alerts.recent")}</div>
        <table className="w-full text-sm">
          <thead className="border-b text-left text-muted-foreground">
            <tr>
              <th className="px-5 py-2">{t("reseller.alerts.colTime")}</th>
              <th className="px-5 py-2">{t("reseller.alerts.colMetric")}</th>
              <th className="px-5 py-2">{t("reseller.alerts.colThreshold")}</th>
              <th className="px-5 py-2">{t("reseller.alerts.colUsage")}</th>
            </tr>
          </thead>
          <tbody>
            {events.data?.events.map((ev, i) => (
              <tr key={i} className="border-b last:border-0">
                <td className="px-5 py-2 text-muted-foreground">{new Date(ev.created_at).toLocaleString()}</td>
                <td className="px-5 py-2">{ev.metric}</td>
                <td className="px-5 py-2">{ev.threshold}%</td>
                <td className="px-5 py-2">{ev.usage_pct}%</td>
              </tr>
            ))}
          </tbody>
        </table>
      </Card>
    </div>
  );
}
