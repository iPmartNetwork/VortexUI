import { useState } from "react";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { Gift } from "lucide-react";
import { api } from "@/api/client";
import { Button, Input, Badge, Select } from "@/components/ui";
import { GlassCard } from "@/components/veltrix";
import { useToast } from "@/components/toast";
import { formatBytes } from "@/lib/utils";
import { useTitle } from "@/lib/useTitle";
import { useI18n } from "@/i18n/i18n";

interface ReferralConfig {
  enabled: boolean;
  reward_type: string;
  reward_amount: number;
  max_referrals: number;
  require_paid: boolean;
}

interface ReferralCode {
  id: string;
  user_id: string;
  username: string;
  code: string;
  uses: number;
  max_uses: number;
  created_at: string;
}

interface ReferralEvent {
  id: string;
  referrer_name: string;
  referred_name: string;
  code_used: string;
  reward_type: string;
  reward_amount: number;
  reward_applied: boolean;
  created_at: string;
}

export function Referrals() {
  const { t } = useI18n();
  useTitle(t("referral.title"));
  const qc = useQueryClient();
  const toast = useToast();
  const [form, setForm] = useState<ReferralConfig | null>(null);

  const { data: configData } = useQuery({
    queryKey: ["referral-config"],
    queryFn: () => api<{ config: ReferralConfig }>("/api/referrals/config"),
  });
  const { data: codesData } = useQuery({
    queryKey: ["referral-codes"],
    queryFn: () => api<{ codes: ReferralCode[] }>("/api/referrals/codes"),
  });
  const { data: eventsData } = useQuery({
    queryKey: ["referral-events"],
    queryFn: () => api<{ events: ReferralEvent[] }>("/api/referrals/events"),
  });

  const config = form ?? configData?.config;

  const save = useMutation({
    mutationFn: (c: ReferralConfig) => api("/api/referrals/config", { method: "PUT", body: c }),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["referral-config"] });
      toast.success(t("referral.configSaved"));
    },
  });

  function update<K extends keyof ReferralConfig>(field: K, value: ReferralConfig[K]) {
    setForm((prev) => ({ ...(prev ?? configData?.config ?? ({} as ReferralConfig)), [field]: value }));
  }

  const amountLabel =
    config?.reward_type === "data"
      ? t("referral.amountBytes")
      : config?.reward_type === "days"
        ? t("referral.amountDays")
        : t("referral.amountPercent");

  return (
    <div className="space-y-5 animate-page-enter">
      <div>
        <h1 className="text-2xl font-bold text-fg tracking-tight">{t("referral.title")}</h1>
        <p className="text-sm text-fg-muted mt-1">{t("referral.subtitle")}</p>
      </div>

      {config && (
        <GlassCard hover={false} className="!p-5 space-y-4">
          <div className="flex items-center justify-between">
            <h3 className="text-sm font-bold text-fg">{t("referral.settings")}</h3>
            <label className="flex items-center gap-2 text-sm text-fg">
              <input type="checkbox" checked={config.enabled} onChange={(e) => update("enabled", e.target.checked)} className="rounded" />
              {t("common.enabled")}
            </label>
          </div>
          <div className="grid grid-cols-1 gap-3 sm:grid-cols-2 md:grid-cols-4">
            <div>
              <label className="text-xs text-fg-subtle">{t("referral.rewardType")}</label>
              <Select value={config.reward_type} onChange={(e) => update("reward_type", e.target.value)}>
                <option value="data">{t("referral.rewardData")}</option>
                <option value="days">{t("referral.rewardDays")}</option>
                <option value="discount">{t("referral.rewardDiscount")}</option>
              </Select>
            </div>
            <div>
              <label className="text-xs text-fg-subtle">{amountLabel}</label>
              <Input value={config.reward_amount} onChange={(e) => update("reward_amount", Number(e.target.value))} inputMode="numeric" />
            </div>
            <div>
              <label className="text-xs text-fg-subtle">{t("referral.maxReferrals")}</label>
              <Input value={config.max_referrals} onChange={(e) => update("max_referrals", Number(e.target.value))} inputMode="numeric" />
            </div>
            <div className="flex items-end">
              <label className="flex items-center gap-2 text-xs pb-2 text-fg">
                <input type="checkbox" checked={config.require_paid} onChange={(e) => update("require_paid", e.target.checked)} className="rounded" />
                {t("referral.requirePaid")}
              </label>
            </div>
          </div>
          <div className="flex justify-end pt-1 border-t border-border/40">
            <Button onClick={() => config && save.mutate(config)} disabled={save.isPending}>{t("common.save")}</Button>
          </div>
        </GlassCard>
      )}

      <div className="grid grid-cols-1 gap-4 lg:grid-cols-2">
        <GlassCard hover={false} className="!p-4 space-y-3">
          <h3 className="text-sm font-bold text-fg">{t("referral.codesTitle")} ({codesData?.codes?.length ?? 0})</h3>
          <div className="space-y-2 max-h-[300px] overflow-y-auto">
            {codesData?.codes?.map((c) => (
              <div key={c.id} className="flex items-center justify-between rounded-lg bg-surface-2/50 border border-border/40 px-3 py-2 text-xs">
                <div>
                  <span className="font-mono text-fg font-bold">{c.code}</span>
                  <span className="ml-2 text-fg-muted">{c.username}</span>
                </div>
                <Badge color="muted">{c.uses} {t("referral.uses")}</Badge>
              </div>
            ))}
            {(!codesData?.codes || codesData.codes.length === 0) && (
              <p className="text-xs text-fg-muted text-center py-4">{t("referral.noCodes")}</p>
            )}
          </div>
        </GlassCard>

        <GlassCard hover={false} className="!p-4 space-y-3">
          <h3 className="text-sm font-bold text-fg flex items-center gap-1.5"><Gift size={14} className="text-primary" /> {t("referral.recent")}</h3>
          <div className="space-y-2 max-h-[300px] overflow-y-auto">
            {eventsData?.events?.map((e) => (
              <div key={e.id} className="rounded-lg bg-surface-2/50 border border-border/40 px-3 py-2 text-xs">
                <div className="flex items-center justify-between">
                  <span className="text-fg">{e.referrer_name} → {e.referred_name}</span>
                  <Badge color={e.reward_applied ? "active" : "limited"}>{e.reward_applied ? t("referral.rewarded") : t("referral.pending")}</Badge>
                </div>
                <div className="text-fg-muted mt-0.5">
                  {e.reward_type}: {e.reward_type === "data" ? formatBytes(e.reward_amount, false) : e.reward_amount} | {new Date(e.created_at).toLocaleDateString()}
                </div>
              </div>
            ))}
            {(!eventsData?.events || eventsData.events.length === 0) && (
              <p className="text-xs text-fg-muted text-center py-4">{t("referral.noEvents")}</p>
            )}
          </div>
        </GlassCard>
      </div>
    </div>
  );
}
