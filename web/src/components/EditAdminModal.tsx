import { useEffect, useState } from "react";
import { useAdminInbounds, useAdminNodes, useAdminPlans, useRoles, useUpdateAdmin } from "@/api/admin-hooks";
import type { Admin } from "@/api/types";
import { AdminInboundPicker } from "@/components/AdminInboundPicker";
import { AdminNodePicker } from "@/components/AdminNodePicker";
import { AdminPlanPicker } from "@/components/AdminPlanPicker";
import { DEFAULT_RESELLER_SETTINGS, RESELLER_SETTING_KEYS, mergeResellerSettings, type ResellerSettingKey } from "@/auth/permissions";
import { useI18n } from "@/i18n/i18n";
import { Button, Input, Select } from "./ui";
import { Modal } from "./Modal";

const SETTING_LABELS: Record<ResellerSettingKey, string> = {
  appearance: "Appearance",
  password: "Change password",
  totp: "Two-factor auth",
  api_tokens: "API tokens",
  backup: "Backup & restore",
  config_template: "Subscription template",
  sub_update: "Subscription auto-update",
  ip_guard: "IP guard",
  branding: "Custom branding",
  auto_backup: "Auto backup",
  update: "Update checker",
  billing: "Payment settings",
};

export function EditAdminModal({ admin, onClose }: { admin: Admin | null; onClose: () => void }) {
  const { t } = useI18n();
  const update = useUpdateAdmin();
  const roles = useRoles();
  const assigned = useAdminInbounds(admin?.id ?? null);
  const nodes = useAdminNodes(admin?.id ?? null);
  const plans = useAdminPlans(admin?.id ?? null);
  const [roleId, setRoleId] = useState("");
  const [userQuota, setUserQuota] = useState("");
  const [trafficQuota, setTrafficQuota] = useState("");
  const [trafficMode, setTrafficMode] = useState("allocated");
  const [inboundIds, setInboundIds] = useState<string[]>([]);
  const [nodeIds, setNodeIds] = useState<string[]>([]);
  const [planIds, setPlanIds] = useState<string[]>([]);
  const [policyMaxGB, setPolicyMaxGB] = useState("");
  const [policyMaxExpireDays, setPolicyMaxExpireDays] = useState("");
  const [allowBulkCreate, setAllowBulkCreate] = useState(true);
  const [allowBulkDelete, setAllowBulkDelete] = useState(true);
  const [autoSuspend, setAutoSuspend] = useState(true);
  const [ipViolationThreshold, setIpViolationThreshold] = useState("");
  const [suspendGraceMinutes, setSuspendGraceMinutes] = useState("");
  const [allowSubResellers, setAllowSubResellers] = useState(false);
  const [allowUserBackup, setAllowUserBackup] = useState(false);
  const [resellerSettings, setResellerSettings] = useState(DEFAULT_RESELLER_SETTINGS);
  const [error, setError] = useState("");

  useEffect(() => {
    if (!admin) return;
    setRoleId(admin.role_id ?? "");
    setUserQuota(admin.user_quota ? String(admin.user_quota) : "");
    setTrafficQuota(admin.traffic_quota ? String(Math.round(admin.traffic_quota / (1024 ** 3))) : "");
    setTrafficMode(admin.traffic_quota_mode || "allocated");
    setPolicyMaxGB(admin.policy_max_data_limit ? String(Math.round(admin.policy_max_data_limit / (1024 ** 3))) : "");
    setPolicyMaxExpireDays(admin.policy_max_expire_days ? String(admin.policy_max_expire_days) : "");
    setAllowBulkCreate(admin.policy_allow_bulk_create !== false);
    setAllowBulkDelete(admin.policy_allow_bulk_delete !== false);
    setAutoSuspend(admin.auto_suspend_enabled !== false);
    setIpViolationThreshold(admin.ip_violation_suspend_threshold ? String(admin.ip_violation_suspend_threshold) : "");
    setSuspendGraceMinutes(admin.suspend_grace_minutes ? String(admin.suspend_grace_minutes) : "60");
    setAllowSubResellers(!!admin.allow_sub_resellers);
    setAllowUserBackup(!!admin.allow_user_backup);
    setResellerSettings(mergeResellerSettings(admin.reseller_settings));
    setError("");
  }, [admin]);

  useEffect(() => {
    if (assigned.data) setInboundIds(assigned.data.inbound_ids ?? []);
  }, [assigned.data]);

  useEffect(() => {
    if (nodes.data) setNodeIds(nodes.data.node_ids ?? []);
  }, [nodes.data]);

  useEffect(() => {
    if (plans.data) setPlanIds(plans.data.plan_ids ?? []);
  }, [plans.data]);

  async function submit(e: React.FormEvent) {
    e.preventDefault();
    if (!admin || admin.sudo) return;
    setError("");
    if (!roleId) {
      setError(t("reseller.editAdmin.roleRequired"));
      return;
    }
    try {
      await update.mutateAsync({
        id: admin.id,
        input: {
          sudo: false,
          role_id: roleId,
          user_quota: userQuota ? Number(userQuota) : 0,
          traffic_quota: trafficQuota ? Number(trafficQuota) * 1024 * 1024 * 1024 : 0,
          traffic_quota_mode: trafficMode,
          inbound_ids: inboundIds,
          node_ids: nodeIds,
          plan_ids: planIds,
          policy_max_data_limit: policyMaxGB ? Number(policyMaxGB) * 1024 * 1024 * 1024 : 0,
          policy_max_expire_days: policyMaxExpireDays ? Number(policyMaxExpireDays) : 0,
          policy_allow_bulk_create: allowBulkCreate,
          policy_allow_bulk_delete: allowBulkDelete,
          auto_suspend_enabled: autoSuspend,
          ip_violation_suspend_threshold: ipViolationThreshold ? Number(ipViolationThreshold) : 0,
          suspend_grace_minutes: suspendGraceMinutes ? Number(suspendGraceMinutes) : 60,
          allow_sub_resellers: allowSubResellers,
          allow_user_backup: allowUserBackup,
          reseller_settings: resellerSettings,
        },
      });
      onClose();
    } catch {
      setError(t("reseller.editAdmin.updateFailed"));
    }
  }

  const title = t("reseller.editAdmin.title").replace("{name}", admin?.username ?? "admin");

  return (
    <Modal open={!!admin} onClose={onClose} title={title} className="max-w-2xl flex flex-col max-h-[90vh]">
      {admin && !admin.sudo && (
        <form onSubmit={submit} className="flex flex-col overflow-hidden">
          <div className="overflow-y-auto space-y-3 pr-1">
            <label className="block text-xs text-muted-foreground">
              {t("reseller.editAdmin.role")}
              <Select className="mt-1" value={roleId} onChange={(e) => setRoleId(e.target.value)} required>
                <option value="">{t("reseller.editAdmin.selectRole")}</option>
                {(roles.data?.roles ?? []).map((r) => (
                  <option key={r.id} value={r.id}>{r.name}</option>
                ))}
              </Select>
            </label>
            <div className="grid grid-cols-2 gap-2">
              <Input placeholder={t("reseller.editAdmin.userQuota")} value={userQuota} onChange={(e) => setUserQuota(e.target.value)} inputMode="numeric" />
              <Input placeholder={t("reseller.editAdmin.trafficQuota")} value={trafficQuota} onChange={(e) => setTrafficQuota(e.target.value)} inputMode="numeric" />
            </div>
            <label className="block text-xs text-muted-foreground">
              {t("reseller.editAdmin.trafficMode")}
              <Select className="mt-1" value={trafficMode} onChange={(e) => setTrafficMode(e.target.value)}>
                <option value="allocated">{t("reseller.editAdmin.allocated")}</option>
                <option value="consumed">{t("reseller.editAdmin.consumed")}</option>
              </Select>
            </label>
            <AdminInboundPicker selected={inboundIds} onChange={setInboundIds} />
            <AdminNodePicker selected={nodeIds} onChange={setNodeIds} />
            <AdminPlanPicker selected={planIds} onChange={setPlanIds} />
            <div className="rounded-md border p-3 space-y-2">
              <p className="text-xs font-semibold text-muted-foreground">{t("reseller.editAdmin.policyLimits")}</p>
              <div className="grid grid-cols-2 gap-2">
                <Input placeholder={t("reseller.editAdmin.maxDataLimit")} value={policyMaxGB} onChange={(e) => setPolicyMaxGB(e.target.value)} inputMode="numeric" />
                <Input placeholder={t("reseller.editAdmin.maxExpire")} value={policyMaxExpireDays} onChange={(e) => setPolicyMaxExpireDays(e.target.value)} inputMode="numeric" />
              </div>
              <label className="flex items-center gap-2 text-sm">
                <input type="checkbox" checked={allowBulkCreate} onChange={(e) => setAllowBulkCreate(e.target.checked)} />
                {t("reseller.editAdmin.allowBulkCreate")}
              </label>
              <label className="flex items-center gap-2 text-sm">
                <input type="checkbox" checked={allowBulkDelete} onChange={(e) => setAllowBulkDelete(e.target.checked)} />
                {t("reseller.editAdmin.allowBulkDelete")}
              </label>
            </div>
            <div className="rounded-md border p-3 space-y-2">
              <p className="text-xs font-semibold text-muted-foreground">{t("reseller.editAdmin.autoSuspend")}</p>
              <label className="flex items-center gap-2 text-sm">
                <input type="checkbox" checked={autoSuspend} onChange={(e) => setAutoSuspend(e.target.checked)} />
                {t("reseller.editAdmin.enableAutoSuspend")}
              </label>
              <div className="grid grid-cols-2 gap-2">
                <Input placeholder={t("reseller.editAdmin.ipViolations")} value={ipViolationThreshold} onChange={(e) => setIpViolationThreshold(e.target.value)} inputMode="numeric" />
                <Input placeholder={t("reseller.editAdmin.quotaGrace")} value={suspendGraceMinutes} onChange={(e) => setSuspendGraceMinutes(e.target.value)} inputMode="numeric" />
              </div>
            </div>
            <div className="rounded-md border p-3 space-y-2">
              <p className="text-xs font-semibold text-muted-foreground">Reseller capabilities</p>
              <label className="flex items-center gap-2 text-sm">
                <input type="checkbox" checked={allowSubResellers} onChange={(e) => setAllowSubResellers(e.target.checked)} />
                Allow creating sub-resellers
              </label>
              <label className="flex items-center gap-2 text-sm">
                <input type="checkbox" checked={allowUserBackup} onChange={(e) => setAllowUserBackup(e.target.checked)} />
                Allow backup of own users only
              </label>
              <p className="pt-1 text-[10px] text-muted-foreground">Settings page sections</p>
              <div className="grid gap-1 sm:grid-cols-2">
                {RESELLER_SETTING_KEYS.map((key) => (
                  <label key={key} className="flex items-center gap-2 text-sm">
                    <input
                      type="checkbox"
                      checked={resellerSettings[key]}
                      onChange={(e) => setResellerSettings({ ...resellerSettings, [key]: e.target.checked })}
                    />
                    {SETTING_LABELS[key]}
                  </label>
                ))}
              </div>
            </div>
            {error && <p className="text-sm text-destructive">{error}</p>}
          </div>
          <div className="sticky bottom-0 flex justify-end gap-2 pt-3 border-t border-border/40 bg-bg-elevated mt-3">
            <Button type="button" variant="ghost" onClick={onClose}>{t("common.cancel")}</Button>
            <Button type="submit" disabled={update.isPending}>
              {update.isPending ? t("reseller.editAdmin.saving") : t("common.save")}
            </Button>
          </div>
        </form>
      )}
      {admin?.sudo && (
        <p className="text-sm text-muted-foreground">{t("reseller.editAdmin.sudoHint")}</p>
      )}
    </Modal>
  );
}
