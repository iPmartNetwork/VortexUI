import { useState } from "react";
import { Info, ShieldAlert } from "lucide-react";
import { Button, Input, Badge, Select } from "@/components/ui";
import { GlassCard } from "@/components/veltrix";
import { useToast } from "@/components/toast";
import { useI18n } from "@/i18n/i18n";
import { useTitle } from "@/lib/useTitle";
import {
  useIPLimitPolicy,
  useUpdateIPLimitPolicy,
  useIPLimitEvents,
  type IPLimitPolicy,
} from "@/api/hooks";

export function IPLimit() {
  useTitle("IP Limit");
  const { t } = useI18n();
  const toast = useToast();
  const [form, setForm] = useState<IPLimitPolicy | null>(null);

  const { data: policyData } = useIPLimitPolicy();
  const { data: eventsData } = useIPLimitEvents();
  const save = useUpdateIPLimitPolicy();

  const policy = form ?? policyData?.policy;

  function update<K extends keyof IPLimitPolicy>(field: K, value: IPLimitPolicy[K]) {
    setForm((prev) => ({ ...(prev ?? policyData?.policy ?? ({} as IPLimitPolicy)), [field]: value }));
  }

  function onSave() {
    if (!policy) return;
    save.mutate(policy, {
      onSuccess: () => toast.success(t("ipLimit.saved")),
    });
  }

  return (
    <div className="space-y-5 animate-page-enter">
      <div>
        <h1 className="text-2xl font-bold text-fg tracking-tight">{t("ipLimit.title")}</h1>
        <p className="text-sm text-fg-muted mt-1">{t("ipLimit.subtitle")}</p>
      </div>

      <div className="rounded-xl border border-primary/25 bg-primary/5 px-4 py-3 flex items-start gap-3">
        <div className="h-8 w-8 rounded-full bg-primary/15 flex items-center justify-center text-primary flex-shrink-0">
          <Info size={16} />
        </div>
        <div className="text-xs text-fg-muted leading-relaxed space-y-1.5">
          <p className="font-semibold text-fg text-sm">{t("ipLimit.infoTitle")}</p>
          <p>{t("ipLimit.infoDesc")}</p>
          <ul className="space-y-1 pt-1">
            <li><strong className="text-fg">{t("ipLimit.actionWarn")}</strong> — {t("ipLimit.warnDesc")}</li>
            <li><strong className="text-fg">{t("ipLimit.actionDisable")}</strong> — {t("ipLimit.disableDesc")}</li>
            <li><strong className="text-fg">{t("ipLimit.actionKill")}</strong> — {t("ipLimit.killDesc")}</li>
          </ul>
        </div>
      </div>

      {policy && (
        <GlassCard hover={false} className="!p-5 space-y-4">
          <div className="flex items-center justify-between">
            <h3 className="text-sm font-bold text-fg">{t("ipLimit.policy")}</h3>
            <label className="flex items-center gap-2 text-sm text-fg">
              <input
                type="checkbox"
                checked={policy.enabled}
                onChange={(e) => update("enabled", e.target.checked)}
                className="rounded"
              />
              {t("ipLimit.enabled")}
            </label>
          </div>

          <div className="grid grid-cols-1 gap-3 md:grid-cols-3">
            <div>
              <label className="text-xs text-fg-subtle">{t("ipLimit.action")}</label>
              <Select value={policy.action} onChange={(e) => update("action", e.target.value as IPLimitPolicy["action"])}>
                <option value="warn">{t("ipLimit.actionWarn")}</option>
                <option value="disable_temporarily">{t("ipLimit.actionDisable")}</option>
                <option value="kill_connections">{t("ipLimit.actionKill")}</option>
              </Select>
            </div>
            <div>
              <label className="text-xs text-fg-subtle">{t("ipLimit.alertCooldown")}</label>
              <Input
                value={policy.alert_cooldown}
                onChange={(e) => update("alert_cooldown", Number(e.target.value))}
                inputMode="numeric"
              />
            </div>
            <div>
              <label className="text-xs text-fg-subtle">{t("ipLimit.restoreAfter")}</label>
              <Input
                value={policy.restore_after}
                onChange={(e) => update("restore_after", Number(e.target.value))}
                inputMode="numeric"
              />
            </div>
          </div>

          {policy.action === "kill_connections" && (
            <div className="rounded-lg border border-warning/40 bg-warning/10 px-3 py-2 text-xs text-warning flex items-start gap-2">
              <ShieldAlert size={14} className="flex-shrink-0 mt-0.5" />
              {t("ipLimit.killWarning")}
            </div>
          )}

          <div className="flex justify-end pt-1 border-t border-border/40">
            <Button onClick={onSave} disabled={save.isPending}>{t("ipLimit.save")}</Button>
          </div>
        </GlassCard>
      )}

      <GlassCard hover={false} className="!p-0 overflow-hidden">
        <div className="px-4 pt-4 pb-1">
          <h3 className="text-sm font-bold text-fg">{t("ipLimit.events")}</h3>
        </div>
        <div className="overflow-x-auto mt-2">
          <table className="w-full text-sm">
            <thead>
              <tr className="border-b border-border/40 text-[11px] uppercase tracking-wide text-fg-subtle bg-surface-2/30">
                <th className="py-3 px-4 text-left">{t("ipLimit.colTime")}</th>
                <th className="py-3 px-4 text-left">{t("ipLimit.colUser")}</th>
                <th className="py-3 px-4 text-left">{t("ipLimit.colOnlineIPs")}</th>
                <th className="py-3 px-4 text-left">{t("ipLimit.colLimit")}</th>
                <th className="py-3 px-4 text-left">{t("ipLimit.colAction")}</th>
              </tr>
            </thead>
            <tbody>
              {eventsData?.events?.map((e) => (
                <tr key={e.id} className="border-b border-border/20 hover:bg-surface-2/40">
                  <td className="py-3 px-4 text-fg-muted whitespace-nowrap">{new Date(e.created_at).toLocaleString()}</td>
                  <td className="py-3 px-4 font-medium text-fg">{e.username}</td>
                  <td className="py-3 px-4 text-fg-muted">{e.online_ips}</td>
                  <td className="py-3 px-4 text-fg-muted">{e.limit}</td>
                  <td className="py-3 px-4">
                    <Badge color={e.action === "kill_connections" ? "expired" : e.action === "disable_temporarily" ? "limited" : "muted"}>
                      {e.action}
                    </Badge>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
          {(!eventsData?.events || eventsData.events.length === 0) && (
            <p className="text-sm text-fg-muted text-center py-8">{t("ipLimit.noEvents")}</p>
          )}
        </div>
      </GlassCard>
    </div>
  );
}
