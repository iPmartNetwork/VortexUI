import { useState } from "react";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { Plus } from "lucide-react";
import { api } from "@/api/client";
import { useAuth } from "@/auth/auth";
import { Badge, Button, Input, Select } from "@/components/ui";
import { Modal } from "@/components/Modal";
import { GlassCard } from "@/components/veltrix";
import { useConfirm } from "@/components/confirm";
import { useToast } from "@/components/toast";
import { useI18n } from "@/i18n/i18n";
import { formatBytes } from "@/lib/utils";

interface Plan {
  id: string;
  admin_id?: string;
  name: string;
  description: string;
  data_limit: number;
  duration_days: number;
  device_limit: number;
  reset_strategy: string;
  price_toman: number;
  price_usd: number;
  enabled: boolean;
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

function planPrice(p: Plan) {
  if (p.price_toman > 0) return `${p.price_toman.toLocaleString()} T`;
  if (p.price_usd > 0) return `$${p.price_usd}`;
  return "Free";
}

export function PlansTab() {
  const { t } = useI18n();
  const toast = useToast();
  const confirm = useConfirm();
  const { session } = useAuth();
  const plans = usePlans();
  const del = useDeletePlan();
  const [createOpen, setCreateOpen] = useState(false);

  const owner = session?.admin.username ?? "admin";

  async function remove(p: Plan) {
    if (!(await confirm({ title: `${t("common.delete")} "${p.name}"?`, destructive: true }))) return;
    await del.mutateAsync(p.id);
    toast.success(t("common.delete"));
  }

  return (
    <div className="space-y-4">
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-2">
        <div>
          <h2 className="text-sm font-bold text-fg">{t("reseller.plansTitle")}</h2>
          <p className="text-xs text-fg-muted mt-0.5">{t("reseller.plansDesc")}</p>
        </div>
        <Button size="sm" onClick={() => setCreateOpen(true)}>
          <Plus size={14} /> {t("reseller.createPlan")}
        </Button>
      </div>

      <div className="grid grid-cols-1 gap-4 lg:grid-cols-2">
        {(plans.data?.plans ?? []).map((p) => (
          <GlassCard key={p.id} hover={false} className="!p-5 space-y-3">
            <div className="flex items-start justify-between gap-3">
              <div className="min-w-0">
                <h3 className="text-sm font-bold text-fg">{p.name}</h3>
                <p className="text-[10px] text-fg-subtle font-mono mt-0.5">
                  ID: p-{p.id.slice(0, 6)} · {t("reseller.owner")}: {owner}
                </p>
              </div>
              <Badge color="active">{planPrice(p)}</Badge>
            </div>
            {p.description && <p className="text-xs text-fg-muted leading-relaxed">{p.description}</p>}
            <div className="rounded-lg bg-surface-2/50 border border-border/40 px-3 py-2 grid grid-cols-3 gap-2 text-center text-[11px]">
              <div>
                <span className="text-fg-subtle block">{t("reseller.statTraffic")}</span>
                <strong className="text-fg">{formatBytes(p.data_limit, false)}</strong>
              </div>
              <div>
                <span className="text-fg-subtle block">{t("reseller.statDuration")}</span>
                <strong className="text-primary">{p.duration_days} {t("reseller.days")}</strong>
              </div>
              <div>
                <span className="text-fg-subtle block">{t("reseller.statDevices")}</span>
                <strong className="text-accent">{p.device_limit || "∞"}</strong>
              </div>
            </div>
            <div className="flex items-center justify-between gap-2 pt-1 border-t border-border/40">
              <span className="text-[11px] text-fg-subtle">{t("reseller.portalShop")}</span>
              <div className="flex gap-2">
                <Button variant="ghost" size="sm" className="text-danger text-xs" onClick={() => remove(p)}>
                  {t("common.delete")}
                </Button>
              </div>
            </div>
          </GlassCard>
        ))}
        {(plans.data?.plans?.length ?? 0) === 0 && !plans.isLoading && (
          <p className="col-span-full py-12 text-center text-sm text-fg-muted">{t("reseller.plansEmpty")}</p>
        )}
      </div>

      <CreatePlanModal open={createOpen} onClose={() => setCreateOpen(false)} />
    </div>
  );
}

function CreatePlanModal({ open, onClose }: { open: boolean; onClose: () => void }) {
  const create = useCreatePlan();
  const toast = useToast();
  const { t } = useI18n();
  const [f, setF] = useState({
    name: "",
    description: "",
    data_limit: "50",
    duration_days: "30",
    device_limit: "2",
    reset_strategy: "monthly",
    price_toman: "180000",
    price_usd: "0",
  });

  async function submit(e: React.FormEvent) {
    e.preventDefault();
    try {
      await create.mutateAsync({
        name: f.name,
        description: f.description,
        data_limit: Number(f.data_limit) * 1024 ** 3,
        duration_days: Number(f.duration_days),
        device_limit: Number(f.device_limit),
        reset_strategy: f.reset_strategy,
        price_toman: Number(f.price_toman),
        price_usd: Number(f.price_usd),
      });
      toast.success(t("common.create"));
      onClose();
    } catch {
      toast.error(t("billing.depositFailed"));
    }
  }

  return (
    <Modal open={open} onClose={onClose} title={t("reseller.createPlan")} className="max-w-lg">
      <form onSubmit={submit} className="space-y-3">
        <Input placeholder={t("reseller.planName")} value={f.name} onChange={(e) => setF((s) => ({ ...s, name: e.target.value }))} required />
        <Input placeholder={t("billing.pkgDesc")} value={f.description} onChange={(e) => setF((s) => ({ ...s, description: e.target.value }))} />
        <div className="grid grid-cols-3 gap-2">
          <Input placeholder="GB" value={f.data_limit} onChange={(e) => setF((s) => ({ ...s, data_limit: e.target.value }))} inputMode="numeric" />
          <Input placeholder={t("reseller.days")} value={f.duration_days} onChange={(e) => setF((s) => ({ ...s, duration_days: e.target.value }))} inputMode="numeric" />
          <Input placeholder={t("reseller.statDevices")} value={f.device_limit} onChange={(e) => setF((s) => ({ ...s, device_limit: e.target.value }))} inputMode="numeric" />
        </div>
        <div className="grid grid-cols-2 gap-2">
          <Input placeholder="Toman" value={f.price_toman} onChange={(e) => setF((s) => ({ ...s, price_toman: e.target.value }))} inputMode="numeric" />
          <Input placeholder="USD" value={f.price_usd} onChange={(e) => setF((s) => ({ ...s, price_usd: e.target.value }))} inputMode="decimal" />
        </div>
        <Select value={f.reset_strategy} onChange={(e) => setF((s) => ({ ...s, reset_strategy: e.target.value }))}>
          <option value="monthly">Monthly</option>
          <option value="weekly">Weekly</option>
          <option value="daily">Daily</option>
          <option value="no_reset">No reset</option>
        </Select>
        <div className="flex justify-end gap-2 pt-1">
          <Button type="button" variant="ghost" onClick={onClose}>{t("common.cancel")}</Button>
          <Button type="submit" disabled={create.isPending}>{t("common.create")}</Button>
        </div>
      </form>
    </Modal>
  );
}
