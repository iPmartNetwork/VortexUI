import { useState } from "react";
import { Button, Card, Input, PageHeader, Badge, Select } from "@/components/ui";
import { useToast } from "@/components/toast";
import { useI18n } from "@/i18n/i18n";
import {
  useIPLimitPolicy,
  useUpdateIPLimitPolicy,
  useIPLimitEvents,
  type IPLimitPolicy,
} from "@/api/hooks";

export function IPLimit() {
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
    <div className="space-y-6 animate-page-enter">
      <PageHeader title={t("ipLimit.title")} subtitle={t("ipLimit.subtitle")} />

      <div className="rounded-lg border border-border/40 bg-surface-2/20 p-4 text-xs text-fg-muted space-y-2">
        <p className="font-medium text-fg text-sm">{t("ipLimit.infoTitle")}</p>
        <p>{t("ipLimit.infoDesc")}</p>
        <ul className="list-disc pl-4 space-y-1">
          <li><strong>{t("ipLimit.actionWarn")}</strong> — {t("ipLimit.warnDesc")}</li>
          <li><strong>{t("ipLimit.actionDisable")}</strong> — {t("ipLimit.disableDesc")}</li>
          <li><strong>{t("ipLimit.actionKill")}</strong> — {t("ipLimit.killDesc")}</li>
        </ul>
      </div>

      {policy && (
        <Card className="space-y-4">
          <div className="flex items-center justify-between">
            <h3 className="text-sm font-bold text-fg">{t("ipLimit.policy")}</h3>
            <label className="flex items-center gap-2 text-sm">
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
            <div className="rounded-lg border border-warning/40 bg-warning/10 px-3 py-2 text-xs text-warning">
              {t("ipLimit.killWarning")}
            </div>
          )}

          <div className="flex justify-end">
            <Button onClick={onSave} disabled={save.isPending}>{t("ipLimit.save")}</Button>
          </div>
        </Card>
      )}

      <Card>
        <h3 className="text-sm font-bold text-fg mb-3">{t("ipLimit.events")}</h3>
        <div className="overflow-x-auto">
          <table className="w-full text-xs">
            <thead>
              <tr className="text-left text-fg-subtle border-b border-border/40">
                <th className="py-2 pe-3 font-medium">{t("ipLimit.colTime")}</th>
                <th className="py-2 pe-3 font-medium">{t("ipLimit.colUser")}</th>
                <th className="py-2 pe-3 font-medium">{t("ipLimit.colOnlineIPs")}</th>
                <th className="py-2 pe-3 font-medium">{t("ipLimit.colLimit")}</th>
                <th className="py-2 pe-3 font-medium">{t("ipLimit.colAction")}</th>
              </tr>
            </thead>
            <tbody>
              {eventsData?.events?.map((e) => (
                <tr key={e.id} className="border-b border-border/20">
                  <td className="py-2 pe-3 text-fg-muted whitespace-nowrap">{new Date(e.created_at).toLocaleString()}</td>
                  <td className="py-2 pe-3 font-medium text-fg">{e.username}</td>
                  <td className="py-2 pe-3 text-fg-muted">{e.online_ips}</td>
                  <td className="py-2 pe-3 text-fg-muted">{e.limit}</td>
                  <td className="py-2 pe-3">
                    <Badge color={e.action === "kill_connections" ? "expired" : e.action === "disable_temporarily" ? "limited" : "muted"}>
                      {e.action}
                    </Badge>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
          {(!eventsData?.events || eventsData.events.length === 0) && (
            <p className="text-xs text-fg-muted text-center py-4">{t("ipLimit.noEvents")}</p>
          )}
        </div>
      </Card>
    </div>
  );
}
