import { useState } from "react";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { api } from "@/api/client";
import { Button, Card, Input, PageHeader, Select } from "@/components/ui";
import { Modal } from "@/components/Modal";
import { useConfirm } from "@/components/confirm";
import { useToast } from "@/components/toast";
import { formatBytes } from "@/lib/utils";

interface Plan {
  id: string;
  name: string;
  description: string;
  data_limit: number;
  duration_days: number;
  device_limit: number;
  reset_strategy: string;
  price_toman: number;
  price_usd: number;
  max_users: number;
  enabled: boolean;
  created_at: string;
}

function usePlans() {
  return useQuery({ queryKey: ["plans"], queryFn: () => api<{ plans: Plan[] }>("/api/plans") });
}

function useCreatePlan() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (input: Record<string, unknown>) => api<{ plan: Plan }>("/api/plans", { method: "POST", body: input }),
    onSuccess: () => qc.invalidateQueries({ queryKey: ["plans"] }),
  });
}

function useDeletePlan() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => api<void>(`/api/plans/${id}`, { method: "DELETE" }),
    onSuccess: () => qc.invalidateQueries({ queryKey: ["plans"] }),
  });
}

export function Plans() {
  const plans = usePlans();
  const del = useDeletePlan();
  const confirm = useConfirm();
  const toast = useToast();
  const [createOpen, setCreateOpen] = useState(false);

  async function remove(p: Plan) {
    const ok = await confirm({ title: `Delete plan "${p.name}"?`, confirmLabel: "Delete", destructive: true });
    if (!ok) return;
    await del.mutateAsync(p.id);
    toast.success(`Deleted ${p.name}`);
  }

  return (
    <div className="space-y-6 animate-fade-in">
      <div className="flex items-center justify-between">
        <PageHeader title="Plans" />
        <Button onClick={() => setCreateOpen(true)}>New plan</Button>
      </div>

      <CreatePlanModal open={createOpen} onClose={() => setCreateOpen(false)} />

      <div className="grid grid-cols-1 gap-4 md:grid-cols-2 lg:grid-cols-3">
        {plans.data?.plans?.map((p) => (
          <Card key={p.id} className="space-y-3">
            <div className="flex items-center justify-between">
              <h3 className="text-sm font-bold text-fg">{p.name}</h3>
              <span className={`h-2 w-2 rounded-full ${p.enabled ? "bg-success" : "bg-fg-subtle"}`} />
            </div>
            {p.description && <p className="text-xs text-fg-muted">{p.description}</p>}
            <div className="grid grid-cols-2 gap-2 text-xs">
              <div><span className="text-fg-subtle">Data:</span> <strong>{formatBytes(p.data_limit, false)}</strong></div>
              <div><span className="text-fg-subtle">Duration:</span> <strong>{p.duration_days}d</strong></div>
              <div><span className="text-fg-subtle">Devices:</span> <strong>{p.device_limit || "∞"}</strong></div>
              <div><span className="text-fg-subtle">Reset:</span> <strong>{p.reset_strategy}</strong></div>
            </div>
            <div className="flex items-center justify-between border-t border-border/40 pt-2">
              <div className="text-sm font-bold text-primary">
                {p.price_toman > 0 ? `${p.price_toman.toLocaleString()} تومان` : p.price_usd > 0 ? `$${p.price_usd}` : "Free"}
              </div>
              <Button variant="ghost" className="text-destructive text-xs" onClick={() => remove(p)}>Delete</Button>
            </div>
          </Card>
        ))}
        {plans.data?.plans?.length === 0 && (
          <p className="col-span-full text-center text-sm text-fg-muted py-8">No plans yet — create one to start selling.</p>
        )}
      </div>
    </div>
  );
}

function CreatePlanModal({ open, onClose }: { open: boolean; onClose: () => void }) {
  const create = useCreatePlan();
  const toast = useToast();
  const [f, setF] = useState({ name: "", description: "", data_limit: "50", duration_days: "30", device_limit: "3", reset_strategy: "monthly", price_toman: "0", price_usd: "0" });

  async function submit(e: React.FormEvent) {
    e.preventDefault();
    try {
      await create.mutateAsync({
        name: f.name,
        description: f.description,
        data_limit: Number(f.data_limit) * 1024 * 1024 * 1024,
        duration_days: Number(f.duration_days),
        device_limit: Number(f.device_limit),
        reset_strategy: f.reset_strategy,
        price_toman: Number(f.price_toman),
        price_usd: Number(f.price_usd),
      });
      toast.success(`Plan "${f.name}" created`);
      onClose();
    } catch { toast.error("Failed to create plan"); }
  }

  return (
    <Modal open={open} onClose={onClose} title="New Plan">
      <form onSubmit={submit} className="space-y-3">
        <Input placeholder="Plan name" value={f.name} onChange={(e) => setF(s => ({...s, name: e.target.value}))} required />
        <Input placeholder="Description (optional)" value={f.description} onChange={(e) => setF(s => ({...s, description: e.target.value}))} />
        <div className="grid grid-cols-2 gap-2">
          <Input placeholder="Data (GB)" value={f.data_limit} onChange={(e) => setF(s => ({...s, data_limit: e.target.value}))} inputMode="numeric" />
          <Input placeholder="Duration (days)" value={f.duration_days} onChange={(e) => setF(s => ({...s, duration_days: e.target.value}))} inputMode="numeric" />
        </div>
        <div className="grid grid-cols-2 gap-2">
          <Input placeholder="Device limit" value={f.device_limit} onChange={(e) => setF(s => ({...s, device_limit: e.target.value}))} inputMode="numeric" />
          <Select value={f.reset_strategy} onChange={(e) => setF(s => ({...s, reset_strategy: e.target.value}))}>
            <option value="no_reset">No reset</option>
            <option value="daily">Daily</option>
            <option value="weekly">Weekly</option>
            <option value="monthly">Monthly</option>
          </Select>
        </div>
        <div className="grid grid-cols-2 gap-2">
          <Input placeholder="Price (Toman)" value={f.price_toman} onChange={(e) => setF(s => ({...s, price_toman: e.target.value}))} inputMode="numeric" />
          <Input placeholder="Price (USD)" value={f.price_usd} onChange={(e) => setF(s => ({...s, price_usd: e.target.value}))} inputMode="decimal" />
        </div>
        <div className="flex justify-end gap-2 pt-2">
          <Button type="button" variant="ghost" onClick={onClose}>Cancel</Button>
          <Button type="submit" disabled={create.isPending}>Create</Button>
        </div>
      </form>
    </Modal>
  );
}
