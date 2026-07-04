import { useState } from "react";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { api } from "@/api/client";
import { Button, Card, Input, PageHeader, Badge } from "@/components/ui";
import { useToast } from "@/components/toast";

interface QNConfig { enabled: boolean; notify_telegram: boolean; notify_webhook: boolean; webhook_url: string; notify_at_percent: number[]; cooldown_minutes: number; }
interface QNEvent { id: string; username: string; percent: number; channel: string; delivered: boolean; sent_at: string; }

export function QuotaNotifications() {
  const qc = useQueryClient(); const toast = useToast();
  const [form, setForm] = useState<QNConfig | null>(null);

  const { data: cfgData } = useQuery({ queryKey: ["qn-config"], queryFn: () => api<{ config: QNConfig }>("/api/quota-notify/config") });
  const { data: eventsData } = useQuery({ queryKey: ["qn-events"], queryFn: () => api<{ events: QNEvent[] }>("/api/quota-notify/events") });

  const cfg = form ?? cfgData?.config;
  const save = useMutation({ mutationFn: (c: QNConfig) => api("/api/quota-notify/config", { method: "PUT", body: c }), onSuccess: () => { qc.invalidateQueries({ queryKey: ["qn-config"] }); toast.success("Saved"); } });

  function update(field: keyof QNConfig, value: any) { setForm(prev => ({ ...(prev ?? cfgData?.config ?? {} as any), [field]: value })); }

  return (
    <div className="space-y-6 animate-page-enter">
      <PageHeader title="Quota Notifications" subtitle="Notify users at each smart quota tier via Telegram or webhook" />
      {cfg && (
        <Card className="space-y-4">
          <div className="flex items-center justify-between"><h3 className="text-sm font-bold text-fg">Settings</h3>
            <label className="flex items-center gap-2 text-sm"><input type="checkbox" checked={cfg.enabled} onChange={e => update("enabled", e.target.checked)} className="rounded" /> Enabled</label>
          </div>
          <div className="grid grid-cols-2 gap-3">
            <div><label className="text-xs text-fg-subtle">Notify at (%)</label><Input value={cfg.notify_at_percent?.join(", ") ?? ""} onChange={e => update("notify_at_percent", e.target.value.split(",").map(s => Number(s.trim())).filter(Boolean))} placeholder="50, 80, 90, 100" /></div>
            <div><label className="text-xs text-fg-subtle">Cooldown (min)</label><Input value={cfg.cooldown_minutes} onChange={e => update("cooldown_minutes", Number(e.target.value))} inputMode="numeric" /></div>
          </div>
          <div className="flex flex-wrap gap-4">
            <label className="flex items-center gap-2 text-sm"><input type="checkbox" checked={cfg.notify_telegram} onChange={e => update("notify_telegram", e.target.checked)} className="rounded" /> Telegram</label>
            <label className="flex items-center gap-2 text-sm"><input type="checkbox" checked={cfg.notify_webhook} onChange={e => update("notify_webhook", e.target.checked)} className="rounded" /> Webhook</label>
          </div>
          {cfg.notify_webhook && <Input placeholder="Webhook URL" value={cfg.webhook_url} onChange={e => update("webhook_url", e.target.value)} />}
          <div className="flex justify-end"><Button onClick={() => cfg && save.mutate(cfg)} disabled={save.isPending}>Save</Button></div>
        </Card>
      )}
      <Card>
        <h3 className="text-sm font-bold text-fg mb-3">Recent Notifications</h3>
        <div className="space-y-2 max-h-[300px] overflow-y-auto">
          {eventsData?.events?.map(e => (
            <div key={e.id} className="flex items-center justify-between rounded-lg bg-surface-2/40 px-3 py-2 text-xs">
              <div><strong className="text-fg">{e.username}</strong> <span className="text-fg-muted">at {e.percent}% via {e.channel}</span></div>
              <Badge color={e.delivered ? "active" : "expired"}>{e.delivered ? "Sent" : "Failed"}</Badge>
            </div>
          ))}
          {(!eventsData?.events || eventsData.events.length === 0) && <p className="text-xs text-fg-muted text-center py-4">No notifications sent yet.</p>}
        </div>
      </Card>
    </div>
  );
}
