import { useState } from "react";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { api } from "@/api/client";
import { Button, Input, Badge } from "@/components/ui";
import { GlassCard } from "@/components/veltrix";
import { useToast } from "@/components/toast";
import { useTitle } from "@/lib/useTitle";
import { useI18n } from "@/i18n/i18n";

interface QNConfig { enabled: boolean; notify_telegram: boolean; notify_webhook: boolean; webhook_url: string; notify_at_percent: number[]; cooldown_minutes: number; }
interface QNEvent { id: string; username: string; percent: number; channel: string; delivered: boolean; sent_at: string; }

export function QuotaNotifications() {
  const { t } = useI18n();
  useTitle(t("nav.quotaNotify"));
  const qc = useQueryClient();
  const toast = useToast();
  const [form, setForm] = useState<QNConfig | null>(null);

  const { data: cfgData } = useQuery({ queryKey: ["qn-config"], queryFn: () => api<{ config: QNConfig }>("/api/quota-notify/config") });
  const { data: eventsData } = useQuery({ queryKey: ["qn-events"], queryFn: () => api<{ events: QNEvent[] }>("/api/quota-notify/events") });

  const cfg = form ?? cfgData?.config;
  const save = useMutation({
    mutationFn: (c: QNConfig) => api("/api/quota-notify/config", { method: "PUT", body: c }),
    onSuccess: () => { qc.invalidateQueries({ queryKey: ["qn-config"] }); toast.success(t("common.saved")); },
  });

  function update<K extends keyof QNConfig>(field: K, value: QNConfig[K]) {
    setForm((prev) => ({ ...(prev ?? cfgData?.config ?? ({} as QNConfig)), [field]: value }));
  }

  return (
    <div className="space-y-5 animate-page-enter">
      <div>
        <h1 className="text-2xl font-bold text-fg tracking-tight">{t("nav.quotaNotify")}</h1>
        <p className="text-sm text-fg-muted mt-1">{t("quotaNotify.subtitle")}</p>
      </div>

      {cfg && (
        <GlassCard hover={false} className="!p-5 space-y-4">
          <div className="flex items-center justify-between">
            <h3 className="text-sm font-bold text-fg">{t("quotaNotify.settings")}</h3>
            <label className="flex items-center gap-2 text-sm text-fg">
              <input type="checkbox" checked={cfg.enabled} onChange={(e) => update("enabled", e.target.checked)} className="rounded" /> {t("common.enabled")}
            </label>
          </div>
          <div className="grid grid-cols-1 gap-3 sm:grid-cols-2">
            <div>
              <label className="text-xs text-fg-subtle">{t("quotaNotify.notifyAt")}</label>
              <Input
                value={cfg.notify_at_percent?.join(", ") ?? ""}
                onChange={(e) => update("notify_at_percent", e.target.value.split(",").map((s) => Number(s.trim())).filter(Boolean))}
                placeholder="50, 80, 90, 100"
              />
            </div>
            <div><label className="text-xs text-fg-subtle">{t("quotaNotify.cooldown")}</label><Input value={cfg.cooldown_minutes} onChange={(e) => update("cooldown_minutes", Number(e.target.value))} inputMode="numeric" /></div>
          </div>
          <div className="flex flex-wrap gap-4">
            <label className="flex items-center gap-2 text-sm text-fg"><input type="checkbox" checked={cfg.notify_telegram} onChange={(e) => update("notify_telegram", e.target.checked)} className="rounded" /> {t("quotaNotify.telegram")}</label>
            <label className="flex items-center gap-2 text-sm text-fg"><input type="checkbox" checked={cfg.notify_webhook} onChange={(e) => update("notify_webhook", e.target.checked)} className="rounded" /> {t("quotaNotify.webhook")}</label>
          </div>
          {cfg.notify_webhook && <Input placeholder={t("quotaNotify.webhookUrl")} value={cfg.webhook_url} onChange={(e) => update("webhook_url", e.target.value)} />}
          <div className="flex justify-end pt-1 border-t border-border/40">
            <Button onClick={() => cfg && save.mutate(cfg)} disabled={save.isPending}>{t("common.save")}</Button>
          </div>
        </GlassCard>
      )}

      <GlassCard hover={false} className="!p-4 space-y-3">
        <h3 className="text-sm font-bold text-fg">{t("quotaNotify.recent")}</h3>
        <div className="space-y-2 max-h-[300px] overflow-y-auto">
          {eventsData?.events?.map((e) => (
            <div key={e.id} className="flex items-center justify-between rounded-lg bg-surface-2/50 border border-border/40 px-3 py-2.5 text-xs">
              <div><strong className="text-fg">{e.username}</strong> <span className="text-fg-muted">{t("quotaNotify.atPrefix")} {e.percent}% {t("quotaNotify.viaPrefix")} {e.channel}</span></div>
              <Badge color={e.delivered ? "active" : "expired"}>{e.delivered ? t("quotaNotify.sent") : t("quotaNotify.failed")}</Badge>
            </div>
          ))}
          {(!eventsData?.events || eventsData.events.length === 0) && <p className="text-sm text-fg-muted text-center py-6">{t("quotaNotify.empty")}</p>}
        </div>
      </GlassCard>
    </div>
  );
}
