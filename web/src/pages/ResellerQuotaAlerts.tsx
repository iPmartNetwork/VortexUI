import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { api } from "@/api/client";
import { Button, Input } from "@/components/ui";
import { GlassCard } from "@/components/veltrix";
import { useToast } from "@/components/toast";
import { useI18n } from "@/i18n/i18n";
import { useTitle } from "@/lib/useTitle";

interface AdminQuotaNotifyConfig {
  enabled: boolean;
  notify_telegram: boolean;
  webhook_url: string;
  notify_at_percent: number[];
  cooldown_minutes: number;
}

export function ResellerQuotaAlerts() {
  useTitle("Quota Alerts");
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
  if (!cfg) return <p className="text-sm text-fg-muted text-center py-8">{t("common.loading")}</p>;

  return (
    <div className="space-y-5 animate-page-enter">
      <div>
        <h1 className="text-2xl font-bold text-fg tracking-tight">{t("reseller.alerts.title")}</h1>
        <p className="text-sm text-fg-muted mt-1">{t("reseller.alerts.subtitle")}</p>
      </div>

      <GlassCard hover={false} className="!p-5 space-y-4">
        <label className="flex items-center gap-2 text-sm text-fg">
          <input type="checkbox" checked={cfg.enabled} onChange={(e) => save.mutate({ ...cfg, enabled: e.target.checked })} className="rounded" />
          {t("common.enabled")}
        </label>
        <label className="flex items-center gap-2 text-sm text-fg">
          <input type="checkbox" checked={cfg.notify_telegram} onChange={(e) => save.mutate({ ...cfg, notify_telegram: e.target.checked })} className="rounded" />
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
        <div className="flex justify-end pt-1 border-t border-border/40">
          <Button onClick={() => save.mutate(cfg)} disabled={save.isPending}>{t("common.save")}</Button>
        </div>
      </GlassCard>

      <GlassCard hover={false} className="!p-0 overflow-hidden">
        <div className="px-4 pt-4 pb-1">
          <h3 className="text-sm font-bold text-fg">{t("reseller.alerts.recent")}</h3>
        </div>
        <div className="overflow-x-auto mt-2">
          <table className="w-full text-sm">
            <thead>
              <tr className="border-b border-border/40 text-[11px] uppercase tracking-wide text-fg-subtle bg-surface-2/30">
                <th className="px-4 py-3 text-left">{t("reseller.alerts.colTime")}</th>
                <th className="px-4 py-3 text-left">{t("reseller.alerts.colMetric")}</th>
                <th className="px-4 py-3 text-left">{t("reseller.alerts.colThreshold")}</th>
                <th className="px-4 py-3 text-left">{t("reseller.alerts.colUsage")}</th>
              </tr>
            </thead>
            <tbody>
              {events.data?.events.map((ev, i) => (
                <tr key={i} className="border-b border-border/20 last:border-0 hover:bg-surface-2/40">
                  <td className="px-4 py-2.5 text-fg-muted">{new Date(ev.created_at).toLocaleString()}</td>
                  <td className="px-4 py-2.5 text-fg">{ev.metric}</td>
                  <td className="px-4 py-2.5 text-fg-muted">{ev.threshold}%</td>
                  <td className="px-4 py-2.5 text-fg-muted">{ev.usage_pct}%</td>
                </tr>
              ))}
              {(!events.data?.events || events.data.events.length === 0) && (
                <tr><td colSpan={4} className="px-4 py-6 text-center text-fg-muted">{t("common.none")}</td></tr>
              )}
            </tbody>
          </table>
        </div>
      </GlassCard>
    </div>
  );
}
