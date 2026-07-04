import { useState } from "react";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { api } from "@/api/client";
import { Button, Card, Input, PageHeader, Badge, Select } from "@/components/ui";
import { useToast } from "@/components/toast";
import { useI18n } from "@/i18n/i18n";

interface ProbingPolicy {
  enabled: boolean;
  action: string;
  block_duration: number;
  max_probe_per_min: number;
  whitelisted_ips: string[];
  honeypot_html: string;
  notify_telegram: boolean;
}

interface ProbeEvent {
  id: string;
  source_ip: string;
  port: number;
  method: string;
  fingerprint: string;
  action: string;
  created_at: string;
}

interface BlockedIP {
  ip: string;
  reason: string;
  blocked_at: string;
  expires_at: string;
}

export function ProbingProtection() {
  const { t } = useI18n();
  const qc = useQueryClient();
  const toast = useToast();
  const [form, setForm] = useState<ProbingPolicy | null>(null);

  const { data: policyData } = useQuery({
    queryKey: ["probing-policy"],
    queryFn: () => api<{ policy: ProbingPolicy }>("/api/probing/policy"),
  });
  const { data: eventsData } = useQuery({
    queryKey: ["probing-events"],
    queryFn: () => api<{ events: ProbeEvent[] }>("/api/probing/events"),
  });
  const { data: blockedData } = useQuery({
    queryKey: ["probing-blocked"],
    queryFn: () => api<{ blocked_ips: BlockedIP[] }>("/api/probing/blocked"),
  });

  const policy = form ?? policyData?.policy;

  const save = useMutation({
    mutationFn: (p: ProbingPolicy) => api("/api/probing/policy", { method: "PUT", body: p }),
    onSuccess: () => { qc.invalidateQueries({ queryKey: ["probing-policy"] }); toast.success("Policy saved"); },
  });

  const unblock = useMutation({
    mutationFn: (ip: string) => api("/api/probing/unblock", { method: "POST", body: { ip } }),
    onSuccess: () => { qc.invalidateQueries({ queryKey: ["probing-blocked"] }); toast.success("Unblocked"); },
  });

  function update(field: keyof ProbingPolicy, value: any) {
    setForm(prev => ({ ...(prev ?? policyData?.policy ?? {} as any), [field]: value }));
  }

  return (
    <div className="space-y-6 animate-page-enter">
      <PageHeader title={t("probing.title")} subtitle={t("probing.subtitle")} />

      <div className="rounded-lg border border-border/40 bg-surface-2/20 p-4 text-xs text-fg-muted space-y-2">
        <p className="font-medium text-fg text-sm">{t("probing.infoTitle")}</p>
        <p>{t("probing.infoDesc")}</p>
        <ul className="list-disc pl-4 space-y-1">
          <li><strong>Block</strong> — {t("probing.block")}</li>
          <li><strong>Honeypot</strong> — {t("probing.honeypot")}</li>
          <li><strong>Log only</strong> — {t("probing.logOnly")}</li>
        </ul>
      </div>

      {policy && (
        <Card className="space-y-4">
          <div className="flex items-center justify-between">
            <h3 className="text-sm font-bold text-fg">{t("probing.policy")}</h3>
            <label className="flex items-center gap-2 text-sm">
              <input type="checkbox" checked={policy.enabled} onChange={(e) => update("enabled", e.target.checked)} className="rounded" />
              Enabled
            </label>
          </div>
          <div className="grid grid-cols-2 gap-3 md:grid-cols-3">
            <div>
              <label className="text-xs text-fg-subtle">{t("probing.action")}</label>
              <Select value={policy.action} onChange={(e) => update("action", e.target.value)}>
                <option value="block">Block</option>
                <option value="honeypot">Honeypot</option>
                <option value="log">Log only</option>
              </Select>
            </div>
            <div>
              <label className="text-xs text-fg-subtle">{t("probing.blockDuration")}</label>
              <Input value={policy.block_duration} onChange={(e) => update("block_duration", Number(e.target.value))} inputMode="numeric" />
            </div>
            <div>
              <label className="text-xs text-fg-subtle">{t("probing.maxProbes")}</label>
              <Input value={policy.max_probe_per_min} onChange={(e) => update("max_probe_per_min", Number(e.target.value))} inputMode="numeric" />
            </div>
          </div>
          <div>
            <label className="text-xs text-fg-subtle">{t("probing.whitelist")}</label>
            <textarea
              className="field min-h-[60px] resize-y font-mono text-xs"
              value={policy.whitelisted_ips?.join("\n") ?? ""}
              onChange={(e) => update("whitelisted_ips", e.target.value.split("\n").filter(Boolean))}
            />
          </div>
          <label className="flex items-center gap-2 text-sm">
            <input type="checkbox" checked={policy.notify_telegram} onChange={(e) => update("notify_telegram", e.target.checked)} className="rounded" />
            {t("probing.notifyTg")}
          </label>
          <div className="flex justify-end">
            <Button onClick={() => policy && save.mutate(policy)} disabled={save.isPending}>Save</Button>
          </div>
        </Card>
      )}

      <div className="grid grid-cols-1 gap-4 lg:grid-cols-2">
        <Card>
          <h3 className="text-sm font-bold text-fg mb-3">{t("probing.blockedIPs")} ({blockedData?.blocked_ips?.length ?? 0})</h3>
          <div className="space-y-2 max-h-[300px] overflow-y-auto">
            {blockedData?.blocked_ips?.map((b) => (
              <div key={b.ip} className="flex items-center justify-between rounded-lg bg-surface-2/40 px-3 py-2">
                <div>
                  <span className="font-mono text-xs text-fg">{b.ip}</span>
                  <span className="ml-2 text-xs text-fg-muted">{b.reason}</span>
                </div>
                <Button variant="ghost" size="sm" onClick={() => unblock.mutate(b.ip)}>Unblock</Button>
              </div>
            ))}
            {(!blockedData?.blocked_ips || blockedData.blocked_ips.length === 0) && (
              <p className="text-xs text-fg-muted text-center py-4">{t("probing.noBlocked")}</p>
            )}
          </div>
        </Card>

        <Card>
          <h3 className="text-sm font-bold text-fg mb-3">{t("probing.recentProbes")}</h3>
          <div className="space-y-2 max-h-[300px] overflow-y-auto">
            {eventsData?.events?.slice(0, 20).map((e) => (
              <div key={e.id} className="flex items-center justify-between rounded-lg bg-surface-2/40 px-3 py-2 text-xs">
                <div className="space-y-0.5">
                  <div className="font-mono text-fg">{e.source_ip}:{e.port}</div>
                  <div className="text-fg-muted">{e.method} — {new Date(e.created_at).toLocaleString()}</div>
                </div>
                <Badge color={e.action === "block" ? "expired" : e.action === "honeypot" ? "limited" : "muted"}>{e.action}</Badge>
              </div>
            ))}
            {(!eventsData?.events || eventsData.events.length === 0) && (
              <p className="text-xs text-fg-muted text-center py-4">{t("probing.noProbes")}</p>
            )}
          </div>
        </Card>
      </div>
    </div>
  );
}
