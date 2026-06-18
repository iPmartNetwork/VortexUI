import { useState } from "react";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { api } from "@/api/client";
import { Button, Card, Input, PageHeader, Badge, Select } from "@/components/ui";
import { Modal } from "@/components/Modal";
import { useToast } from "@/components/toast";
import { useI18n } from "@/i18n/i18n";

interface FPPolicy { enabled: boolean; default_action: string; log_unknown: boolean; }
interface FPRule { id: string; name: string; fingerprint: string; ja3_hash: string; action: string; priority: number; enabled: boolean; }
interface FPEvent { id: string; client_ip: string; fingerprint: string; user_agent: string; action: string; created_at: string; }

export function Fingerprint() {
  const { t } = useI18n();
  const qc = useQueryClient(); const toast = useToast();
  const [addOpen, setAddOpen] = useState(false);
  const [form, setForm] = useState<FPPolicy | null>(null);

  const { data: policyData } = useQuery({ queryKey: ["fp-policy"], queryFn: () => api<{ policy: FPPolicy }>("/api/fingerprint/policy") });
  const { data: rulesData } = useQuery({ queryKey: ["fp-rules"], queryFn: () => api<{ rules: FPRule[] }>("/api/fingerprint/rules") });
  const { data: eventsData } = useQuery({ queryKey: ["fp-events"], queryFn: () => api<{ events: FPEvent[] }>("/api/fingerprint/events") });

  const policy = form ?? policyData?.policy;
  const savePol = useMutation({ mutationFn: (p: FPPolicy) => api("/api/fingerprint/policy", { method: "PUT", body: p }), onSuccess: () => { qc.invalidateQueries({ queryKey: ["fp-policy"] }); toast.success("Saved"); } });
  const delRule = useMutation({ mutationFn: (id: string) => api<void>(`/api/fingerprint/rules/${id}`, { method: "DELETE" }), onSuccess: () => qc.invalidateQueries({ queryKey: ["fp-rules"] }) });

  function update(field: keyof FPPolicy, value: any) { setForm(prev => ({ ...(prev ?? policyData?.policy ?? {} as any), [field]: value })); }

  return (
    <div className="space-y-6 animate-fade-in">
      <PageHeader title={t("fingerprint.title")} subtitle={t("fingerprint.subtitle")} />

      <div className="rounded-lg border border-border/40 bg-surface-2/20 p-4 text-xs text-fg-muted space-y-2">
        <p className="font-medium text-fg text-sm">TLS Client Fingerprinting</p>
        <p>Each TLS client (browser, app, bot) has a unique fingerprint based on its ClientHello packet. This feature validates incoming connections against known fingerprint patterns.</p>
        <ul className="list-disc pl-4 space-y-1">
          <li><strong>Allow</strong> — Let connections with this fingerprint pass through.</li>
          <li><strong>Block</strong> — Reject connections matching this pattern (e.g. known scanner tools).</li>
          <li><strong>Log</strong> — Record the connection without blocking.</li>
        </ul>
        <p>JA3 hash is a standard method to fingerprint TLS clients. Use it to identify specific tools or bots.</p>
      </div>
      {policy && (
        <Card className="space-y-3">
          <div className="flex items-center justify-between">
            <h3 className="text-sm font-bold text-fg">{t("fingerprint.policy")}</h3>
            <label className="flex items-center gap-2 text-sm"><input type="checkbox" checked={policy.enabled} onChange={e => update("enabled", e.target.checked)} className="rounded" /> Enabled</label>
          </div>
          <div className="grid grid-cols-2 gap-3">
            <div><label className="text-xs text-fg-subtle">Default action (unknown FPs)</label><Select value={policy.default_action} onChange={e => update("default_action", e.target.value)}><option value="allow">Allow</option><option value="block">Block</option><option value="log">Log</option></Select></div>
            <label className="flex items-end gap-2 text-sm pb-2"><input type="checkbox" checked={policy.log_unknown} onChange={e => update("log_unknown", e.target.checked)} className="rounded" /> Log unknown</label>
          </div>
          <div className="flex justify-end"><Button onClick={() => policy && savePol.mutate(policy)} disabled={savePol.isPending}>Save</Button></div>
        </Card>
      )}
      <Card>
        <div className="flex items-center justify-between mb-3"><h3 className="text-sm font-bold text-fg">{t("fingerprint.rules")}</h3><Button size="sm" onClick={() => setAddOpen(true)}>{t("fingerprint.addRule")}</Button></div>
        <AddRuleModal open={addOpen} onClose={() => setAddOpen(false)} />
        <div className="space-y-2">
          {rulesData?.rules?.map(r => (
            <div key={r.id} className="flex items-center justify-between rounded-lg bg-surface-2/40 px-3 py-2 text-xs">
              <div className="flex items-center gap-2"><strong className="text-fg">{r.name}</strong><Badge color={r.action === "allow" ? "active" : r.action === "block" ? "expired" : "muted"}>{r.action}</Badge><span className="text-fg-muted">{r.fingerprint}</span></div>
              <Button variant="ghost" size="sm" className="text-destructive" onClick={() => delRule.mutate(r.id)}>Del</Button>
            </div>
          ))}
          {(!rulesData?.rules || rulesData.rules.length === 0) && <p className="text-xs text-fg-muted text-center py-4">{t("fingerprint.noRules")}</p>}
        </div>
      </Card>
      <Card>
        <h3 className="text-sm font-bold text-fg mb-3">{t("fingerprint.recentEvents")}</h3>
        <div className="space-y-2 max-h-[250px] overflow-y-auto">
          {eventsData?.events?.slice(0, 20).map(e => (
            <div key={e.id} className="flex items-center justify-between rounded-lg bg-surface-2/40 px-3 py-2 text-xs">
              <div><span className="font-mono text-fg">{e.client_ip}</span> <span className="text-fg-muted">— {e.fingerprint}</span></div>
              <Badge color={e.action === "allow" ? "active" : e.action === "block" ? "expired" : "muted"}>{e.action}</Badge>
            </div>
          ))}
        </div>
      </Card>
    </div>
  );
}

function AddRuleModal({ open, onClose }: { open: boolean; onClose: () => void }) {
  const qc = useQueryClient(); const toast = useToast();
  const [f, setF] = useState({ name: "", fingerprint: "chrome", ja3_hash: "", action: "allow", priority: 0 });
  const create = useMutation({ mutationFn: (b: any) => api("/api/fingerprint/rules", { method: "POST", body: b }), onSuccess: () => { qc.invalidateQueries({ queryKey: ["fp-rules"] }); onClose(); toast.success("Rule added"); } });
  return (<Modal open={open} onClose={onClose} title="Add Fingerprint Rule"><form onSubmit={e => { e.preventDefault(); create.mutate(f); }} className="space-y-3"><Input placeholder="Rule name" value={f.name} onChange={e => setF(s => ({...s, name: e.target.value}))} required /><Select value={f.fingerprint} onChange={e => setF(s => ({...s, fingerprint: e.target.value}))}><option value="chrome">Chrome</option><option value="firefox">Firefox</option><option value="safari">Safari</option><option value="ios">iOS</option><option value="android">Android</option><option value="curl">curl (suspicious)</option><option value="go">Go (suspicious)</option><option value="python">Python (suspicious)</option></Select><Select value={f.action} onChange={e => setF(s => ({...s, action: e.target.value}))}><option value="allow">Allow</option><option value="block">Block</option><option value="log">Log</option></Select><Input placeholder="JA3 Hash (optional)" value={f.ja3_hash} onChange={e => setF(s => ({...s, ja3_hash: e.target.value}))} /><div className="flex justify-end gap-2"><Button type="button" variant="ghost" onClick={onClose}>Cancel</Button><Button type="submit" disabled={create.isPending}>Add</Button></div></form></Modal>);
}
