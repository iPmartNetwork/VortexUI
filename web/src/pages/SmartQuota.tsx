import { useState } from "react";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { Gauge, Plus, Trash2 } from "lucide-react";
import { api } from "@/api/client";
import { Button, Input } from "@/components/ui";
import { GlassCard } from "@/components/veltrix";
import { Modal } from "@/components/Modal";
import { useToast } from "@/components/toast";
import { useConfirm } from "@/components/confirm";
import { useTitle } from "@/lib/useTitle";

interface QuotaTier {
  usage_percent: number;
  speed_limit: number;
  action: string;
}

interface QuotaPolicy {
  id: string;
  name: string;
  tiers: QuotaTier[];
  enabled: boolean;
  created_at: string;
}

export function SmartQuota() {
  useTitle("Smart Quota");
  const qc = useQueryClient();
  const toast = useToast();
  const confirm = useConfirm();
  const [createOpen, setCreateOpen] = useState(false);

  const { data } = useQuery({
    queryKey: ["quota-policies"],
    queryFn: () => api<{ policies: QuotaPolicy[] }>("/api/quota"),
  });

  const delMut = useMutation({
    mutationFn: (id: string) => api<void>(`/api/quota/${id}`, { method: "DELETE" }),
    onSuccess: () => qc.invalidateQueries({ queryKey: ["quota-policies"] }),
  });

  async function remove(p: QuotaPolicy) {
    const ok = await confirm({ title: `Delete "${p.name}"?`, confirmLabel: "Delete", destructive: true });
    if (!ok) return;
    await delMut.mutateAsync(p.id);
    toast.success("Deleted");
  }

  return (
    <div className="space-y-5 animate-page-enter">
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-3">
        <div>
          <h1 className="text-2xl font-bold text-fg tracking-tight">Smart Quota</h1>
          <p className="text-sm text-fg-muted mt-1">Fair-use policies with progressive speed reduction</p>
        </div>
        <Button onClick={() => setCreateOpen(true)}><Plus size={14} /> New Policy</Button>
      </div>
      <CreatePolicyModal open={createOpen} onClose={() => setCreateOpen(false)} />
      <div className="grid grid-cols-1 gap-4 lg:grid-cols-2">
        {data?.policies?.map((p) => (
          <GlassCard key={p.id} hover className="space-y-3">
            <div className="flex items-center justify-between gap-2">
              <div className="flex items-center gap-2 min-w-0">
                <div className="h-8 w-8 rounded-lg bg-primary/15 flex items-center justify-center text-primary flex-shrink-0">
                  <Gauge size={15} />
                </div>
                <h3 className="text-sm font-bold text-fg truncate">{p.name}</h3>
              </div>
              <span className={`h-2 w-2 rounded-full flex-shrink-0 ${p.enabled ? "bg-success" : "bg-fg-subtle"}`} />
            </div>
            <div className="space-y-1.5 rounded-lg bg-surface-2/50 border border-border/40 px-3 py-2">
              {p.tiers.map((t, i) => (
                <div key={i} className="flex items-center gap-2 text-xs text-fg-muted">
                  <span className="font-mono font-medium text-fg">{t.usage_percent}%</span>
                  <span>→</span>
                  <span>{t.action === "block" ? "Block" : formatSpeed(t.speed_limit)}</span>
                </div>
              ))}
            </div>
            <div className="flex justify-end pt-1 border-t border-border/40">
              <Button variant="ghost" size="sm" className="text-danger" onClick={() => remove(p)}><Trash2 size={13} /> Delete</Button>
            </div>
          </GlassCard>
        ))}
        {(!data?.policies || data.policies.length === 0) && (
          <p className="col-span-full text-center text-sm text-fg-muted py-8">No policies defined. Users will be hard-cut at 100% usage.</p>
        )}
      </div>
    </div>
  );
}

function formatSpeed(bps: number): string {
  if (bps === 0) return "Full speed";
  if (bps >= 1024 * 1024) return `${(bps / (1024 * 1024)).toFixed(1)} MB/s`;
  if (bps >= 1024) return `${(bps / 1024).toFixed(0)} KB/s`;
  return `${bps} B/s`;
}

function CreatePolicyModal({ open, onClose }: { open: boolean; onClose: () => void }) {
  const qc = useQueryClient();
  const toast = useToast();
  const [name, setName] = useState("");
  const [tiers, setTiers] = useState<QuotaTier[]>([
    { usage_percent: 80, speed_limit: 1024 * 1024, action: "reduce" },
    { usage_percent: 100, speed_limit: 512 * 1024, action: "reduce" },
  ]);

  const create = useMutation({
    mutationFn: (input: { name: string; tiers: QuotaTier[] }) => api("/api/quota", { method: "POST", body: input }),
    onSuccess: () => { qc.invalidateQueries({ queryKey: ["quota-policies"] }); onClose(); toast.success("Policy created"); },
  });

  function addTier() {
    setTiers([...tiers, { usage_percent: 100, speed_limit: 0, action: "block" }]);
  }
  function removeTier(idx: number) {
    setTiers(tiers.filter((_, i) => i !== idx));
  }
  function updateTier(idx: number, field: keyof QuotaTier, value: string | number) {
    const copy = [...tiers];
    (copy[idx] as any)[field] = value;
    setTiers(copy);
  }

  return (
    <Modal open={open} onClose={onClose} title="New Quota Policy" className="max-w-lg">
      <form onSubmit={(e) => { e.preventDefault(); create.mutate({ name, tiers }); }} className="space-y-4">
        <Input placeholder="Policy name" value={name} onChange={(e) => setName(e.target.value)} required />
        <div className="space-y-2">
          <div className="flex items-center justify-between">
            <span className="text-xs text-fg-subtle font-medium">Tiers</span>
            <Button type="button" variant="ghost" size="sm" onClick={addTier}>+ Add tier</Button>
          </div>
          {tiers.map((t, i) => (
            <div key={i} className="grid grid-cols-2 gap-2 items-center sm:grid-cols-4">
              <Input placeholder="%" value={t.usage_percent} onChange={(e) => updateTier(i, "usage_percent", Number(e.target.value))} inputMode="numeric" />
              <Input placeholder="Speed (B/s)" value={t.speed_limit} onChange={(e) => updateTier(i, "speed_limit", Number(e.target.value))} inputMode="numeric" />
              <select className="field text-xs" value={t.action} onChange={(e) => updateTier(i, "action", e.target.value)}>
                <option value="reduce">Reduce</option>
                <option value="block">Block</option>
              </select>
              <Button type="button" variant="ghost" size="sm" className="text-danger" onClick={() => removeTier(i)}>×</Button>
            </div>
          ))}
        </div>
        <div className="flex justify-end gap-2 pt-2">
          <Button type="button" variant="ghost" onClick={onClose}>Cancel</Button>
          <Button type="submit" disabled={create.isPending || !name}>Create</Button>
        </div>
      </form>
    </Modal>
  );
}
