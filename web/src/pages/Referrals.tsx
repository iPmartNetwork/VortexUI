import { useState } from "react";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { api } from "@/api/client";
import { Button, Card, Input, PageHeader, Badge, Select } from "@/components/ui";
import { useToast } from "@/components/toast";
import { formatBytes } from "@/lib/utils";

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
    onSuccess: () => { qc.invalidateQueries({ queryKey: ["referral-config"] }); toast.success("Config saved"); },
  });

  function update(field: keyof ReferralConfig, value: any) {
    setForm(prev => ({ ...(prev ?? configData?.config ?? {} as any), [field]: value }));
  }

  return (
    <div className="space-y-6 animate-fade-in">
      <PageHeader title="Referral System" subtitle="Invite codes, rewards, and tracking" />

      {config && (
        <Card className="space-y-4">
          <div className="flex items-center justify-between">
            <h3 className="text-sm font-bold text-fg">Referral Program Settings</h3>
            <label className="flex items-center gap-2 text-sm">
              <input type="checkbox" checked={config.enabled} onChange={(e) => update("enabled", e.target.checked)} className="rounded" />
              Enabled
            </label>
          </div>
          <div className="grid grid-cols-2 gap-3 md:grid-cols-4">
            <div>
              <label className="text-xs text-fg-subtle">Reward type</label>
              <Select value={config.reward_type} onChange={(e) => update("reward_type", e.target.value)}>
                <option value="data">Extra data</option>
                <option value="days">Extra days</option>
                <option value="discount">Discount</option>
              </Select>
            </div>
            <div>
              <label className="text-xs text-fg-subtle">Amount {config.reward_type === "data" ? "(bytes)" : config.reward_type === "days" ? "(days)" : "(%)"}</label>
              <Input value={config.reward_amount} onChange={(e) => update("reward_amount", Number(e.target.value))} inputMode="numeric" />
            </div>
            <div>
              <label className="text-xs text-fg-subtle">Max referrals (0=∞)</label>
              <Input value={config.max_referrals} onChange={(e) => update("max_referrals", Number(e.target.value))} inputMode="numeric" />
            </div>
            <div className="flex items-end">
              <label className="flex items-center gap-2 text-xs pb-2">
                <input type="checkbox" checked={config.require_paid} onChange={(e) => update("require_paid", e.target.checked)} className="rounded" />
                Require paid
              </label>
            </div>
          </div>
          <div className="flex justify-end">
            <Button onClick={() => config && save.mutate(config)} disabled={save.isPending}>Save</Button>
          </div>
        </Card>
      )}

      <div className="grid grid-cols-1 gap-4 lg:grid-cols-2">
        <Card>
          <h3 className="text-sm font-bold text-fg mb-3">Referral Codes ({codesData?.codes?.length ?? 0})</h3>
          <div className="space-y-2 max-h-[300px] overflow-y-auto">
            {codesData?.codes?.map((c) => (
              <div key={c.id} className="flex items-center justify-between rounded-lg bg-surface-2/40 px-3 py-2 text-xs">
                <div>
                  <span className="font-mono text-fg font-bold">{c.code}</span>
                  <span className="ml-2 text-fg-muted">{c.username}</span>
                </div>
                <Badge color="muted">{c.uses} uses</Badge>
              </div>
            ))}
            {(!codesData?.codes || codesData.codes.length === 0) && (
              <p className="text-xs text-fg-muted text-center py-4">No codes generated yet.</p>
            )}
          </div>
        </Card>

        <Card>
          <h3 className="text-sm font-bold text-fg mb-3">Recent Referrals</h3>
          <div className="space-y-2 max-h-[300px] overflow-y-auto">
            {eventsData?.events?.map((e) => (
              <div key={e.id} className="rounded-lg bg-surface-2/40 px-3 py-2 text-xs">
                <div className="flex items-center justify-between">
                  <span className="text-fg">{e.referrer_name} → {e.referred_name}</span>
                  <Badge color={e.reward_applied ? "active" : "limited"}>{e.reward_applied ? "Rewarded" : "Pending"}</Badge>
                </div>
                <div className="text-fg-muted mt-0.5">{e.reward_type}: {e.reward_type === "data" ? formatBytes(e.reward_amount, false) : e.reward_amount} | {new Date(e.created_at).toLocaleDateString()}</div>
              </div>
            ))}
            {(!eventsData?.events || eventsData.events.length === 0) && (
              <p className="text-xs text-fg-muted text-center py-4">No referrals yet.</p>
            )}
          </div>
        </Card>
      </div>
    </div>
  );
}
