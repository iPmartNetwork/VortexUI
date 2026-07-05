import { useState } from "react";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { Info, Plus, Trash2 } from "lucide-react";
import { api } from "@/api/client";
import { Button, Input, Badge, Select } from "@/components/ui";
import { Modal } from "@/components/Modal";
import { GlassCard } from "@/components/veltrix";
import { useToast } from "@/components/toast";
import { useI18n } from "@/i18n/i18n";
import { useTitle } from "@/lib/useTitle";

interface FPPolicy { enabled: boolean; default_action: string; log_unknown: boolean; }
interface FPRule { id: string; name: string; fingerprint: string; ja3_hash: string; action: string; priority: number; enabled: boolean; }
interface FPEvent { id: string; client_ip: string; fingerprint: string; user_agent: string; action: string; created_at: string; }

export function Fingerprint() {
  useTitle("Fingerprint");
  const { t } = useI18n();
  const qc = useQueryClient();
  const toast = useToast();
  const [addOpen, setAddOpen] = useState(false);
  const [form, setForm] = useState<FPPolicy | null>(null);

  const { data: policyData } = useQuery({ queryKey: ["fp-policy"], queryFn: () => api<{ policy: FPPolicy }>("/api/fingerprint/policy") });
  const { data: rulesData } = useQuery({ queryKey: ["fp-rules"], queryFn: () => api<{ rules: FPRule[] }>("/api/fingerprint/rules") });
  const { data: eventsData } = useQuery({ queryKey: ["fp-events"], queryFn: () => api<{ events: FPEvent[] }>("/api/fingerprint/events") });

  const policy = form ?? policyData?.policy;
  const savePol = useMutation({ mutationFn: (p: FPPolicy) => api("/api/fingerprint/policy", { method: "PUT", body: p }), onSuccess: () => { qc.invalidateQueries({ queryKey: ["fp-policy"] }); toast.success("Saved"); } });
  const delRule = useMutation({ mutationFn: (id: string) => api<void>(`/api/fingerprint/rules/${id}`, { method: "DELETE" }), onSuccess: () => qc.invalidateQueries({ queryKey: ["fp-rules"] }) });

  function update<K extends keyof FPPolicy>(field: K, value: FPPolicy[K]) {
    setForm((prev) => ({ ...(prev ?? policyData?.policy ?? ({} as FPPolicy)), [field]: value }));
  }

  return (
    <div className="space-y-5 animate-page-enter">
      <div>
        <h1 className="text-2xl font-bold text-fg tracking-tight">{t("fingerprint.title")}</h1>
        <p className="text-sm text-fg-muted mt-1">{t("fingerprint.subtitle")}</p>
      </div>

      <div className="rounded-xl border border-primary/25 bg-primary/5 px-4 py-3 flex items-start gap-3">
        <div className="h-8 w-8 rounded-full bg-primary/15 flex items-center justify-center text-primary flex-shrink-0">
          <Info size={16} />
        </div>
        <div className="text-xs text-fg-muted leading-relaxed space-y-1.5">
          <p className="font-semibold text-fg text-sm">{t("fingerprint.infoTitle")}</p>
          <p>{t("fingerprint.infoDesc")}</p>
          <ul className="space-y-1 pt-1">
            <li><strong className="text-fg">Allow</strong> — {t("fingerprint.allow")}</li>
            <li><strong className="text-fg">Block</strong> — {t("fingerprint.blockDesc")}</li>
            <li><strong className="text-fg">Log</strong> — {t("fingerprint.logDesc")}</li>
          </ul>
          <p>{t("fingerprint.ja3")}</p>
        </div>
      </div>

      {policy && (
        <GlassCard hover={false} className="!p-5 space-y-3">
          <div className="flex items-center justify-between">
            <h3 className="text-sm font-bold text-fg">{t("fingerprint.policy")}</h3>
            <label className="flex items-center gap-2 text-sm text-fg">
              <input type="checkbox" checked={policy.enabled} onChange={(e) => update("enabled", e.target.checked)} className="rounded" /> Enabled
            </label>
          </div>
          <div className="grid grid-cols-2 gap-3">
            <div>
              <label className="text-xs text-fg-subtle">Default action (unknown FPs)</label>
              <Select value={policy.default_action} onChange={(e) => update("default_action", e.target.value)}>
                <option value="allow">Allow</option>
                <option value="block">Block</option>
                <option value="log">Log</option>
              </Select>
            </div>
            <label className="flex items-end gap-2 text-sm text-fg pb-2">
              <input type="checkbox" checked={policy.log_unknown} onChange={(e) => update("log_unknown", e.target.checked)} className="rounded" /> Log unknown
            </label>
          </div>
          <div className="flex justify-end pt-1 border-t border-border/40">
            <Button onClick={() => policy && savePol.mutate(policy)} disabled={savePol.isPending}>Save</Button>
          </div>
        </GlassCard>
      )}

      <GlassCard hover={false} className="!p-4 space-y-3">
        <div className="flex items-center justify-between">
          <h3 className="text-sm font-bold text-fg">{t("fingerprint.rules")}</h3>
          <Button size="sm" onClick={() => setAddOpen(true)}>
            <Plus size={13} /> {t("fingerprint.addRule")}
          </Button>
        </div>
        <AddRuleModal open={addOpen} onClose={() => setAddOpen(false)} />
        <div className="space-y-2">
          {rulesData?.rules?.map((r) => (
            <div key={r.id} className="flex items-center justify-between rounded-lg bg-surface-2/50 border border-border/40 px-3 py-2.5 text-xs">
              <div className="flex items-center gap-2">
                <strong className="text-fg">{r.name}</strong>
                <Badge color={r.action === "allow" ? "active" : r.action === "block" ? "expired" : "muted"}>{r.action}</Badge>
                <span className="text-fg-muted">{r.fingerprint}</span>
              </div>
              <Button variant="ghost" size="sm" className="text-danger" onClick={() => delRule.mutate(r.id)}>
                <Trash2 size={13} />
              </Button>
            </div>
          ))}
          {(!rulesData?.rules || rulesData.rules.length === 0) && <p className="text-sm text-fg-muted text-center py-6">{t("fingerprint.noRules")}</p>}
        </div>
      </GlassCard>

      <GlassCard hover={false} className="!p-4 space-y-3">
        <h3 className="text-sm font-bold text-fg">{t("fingerprint.recentEvents")}</h3>
        <div className="space-y-2 max-h-[250px] overflow-y-auto">
          {eventsData?.events?.slice(0, 20).map((e) => (
            <div key={e.id} className="flex items-center justify-between rounded-lg bg-surface-2/50 border border-border/40 px-3 py-2.5 text-xs">
              <div><span className="font-mono text-fg">{e.client_ip}</span> <span className="text-fg-muted">— {e.fingerprint}</span></div>
              <Badge color={e.action === "allow" ? "active" : e.action === "block" ? "expired" : "muted"}>{e.action}</Badge>
            </div>
          ))}
          {(!eventsData?.events || eventsData.events.length === 0) && <p className="text-sm text-fg-muted text-center py-6">{t("common.none")}</p>}
        </div>
      </GlassCard>
    </div>
  );
}

function AddRuleModal({ open, onClose }: { open: boolean; onClose: () => void }) {
  const qc = useQueryClient();
  const toast = useToast();
  const [f, setF] = useState({ name: "", fingerprint: "chrome", ja3_hash: "", action: "allow", priority: 0 });
  const create = useMutation({
    mutationFn: (b: typeof f) => api("/api/fingerprint/rules", { method: "POST", body: b }),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["fp-rules"] });
      onClose();
      toast.success("Rule added");
    },
  });
  return (
    <Modal open={open} onClose={onClose} title="Add Fingerprint Rule">
      <form onSubmit={(e) => { e.preventDefault(); create.mutate(f); }} className="space-y-3">
        <Input placeholder="Rule name" value={f.name} onChange={(e) => setF((s) => ({ ...s, name: e.target.value }))} required />
        <Select value={f.fingerprint} onChange={(e) => setF((s) => ({ ...s, fingerprint: e.target.value }))}>
          <option value="chrome">Chrome</option>
          <option value="firefox">Firefox</option>
          <option value="safari">Safari</option>
          <option value="ios">iOS</option>
          <option value="android">Android</option>
          <option value="curl">curl (suspicious)</option>
          <option value="go">Go (suspicious)</option>
          <option value="python">Python (suspicious)</option>
        </Select>
        <Select value={f.action} onChange={(e) => setF((s) => ({ ...s, action: e.target.value }))}>
          <option value="allow">Allow</option>
          <option value="block">Block</option>
          <option value="log">Log</option>
        </Select>
        <Input placeholder="JA3 Hash (optional)" value={f.ja3_hash} onChange={(e) => setF((s) => ({ ...s, ja3_hash: e.target.value }))} />
        <div className="flex justify-end gap-2">
          <Button type="button" variant="ghost" onClick={onClose}>Cancel</Button>
          <Button type="submit" disabled={create.isPending}>Add</Button>
        </div>
      </form>
    </Modal>
  );
}
