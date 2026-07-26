import { Link } from "react-router-dom";
import { useAdmins, useRoles } from "@/api/admin-hooks";
import { ALL_PERMISSIONS } from "@/api/types";
import { mergeResellerSettings, RESELLER_SETTING_KEYS, type ResellerSettingKey } from "@/auth/permissions";
import { Badge } from "@/components/ui";
import { GlassCard } from "@/components/vortexui";
import { useI18n } from "@/i18n/i18n";
import type { TKey } from "@/i18n/dict";

const SETTING_LABEL_KEYS: Record<ResellerSettingKey, TKey> = {
  appearance: "settings.adminsSection.settingAppearance",
  password: "settings.adminsSection.settingPassword",
  totp: "settings.adminsSection.settingTotp",
  api_tokens: "settings.adminsSection.settingApi",
  backup: "settings.adminsSection.settingBackup",
  config_template: "settings.adminsSection.settingConfig",
  sub_update: "settings.adminsSection.settingSubUpdate",
  ip_guard: "settings.adminsSection.settingIpGuard",
  branding: "settings.adminsSection.settingBranding",
  auto_backup: "settings.adminsSection.settingAutoBackup",
  update: "settings.adminsSection.settingUpdate",
  billing: "settings.adminsSection.settingBilling",
};

export function AccessSettingsTab() {
  const { t } = useI18n();
  const admins = useAdmins();
  const roles = useRoles();

  const roleName = (id: string | null) => roles.data?.roles.find((r) => r.id === id)?.name ?? "—";
  const resellerList = (admins.data?.admins ?? []).filter((a) => !a.sudo);

  if (admins.isLoading || roles.isLoading) {
    return <p className="text-sm text-fg-muted">{t("common.loading")}</p>;
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center gap-2 rounded-xl border border-primary/20 bg-primary/5 px-4 py-3 text-xs text-fg-muted">
        <span className="h-1.5 w-1.5 rounded-full bg-primary shadow-[0_0_6px] shadow-primary/50 flex-shrink-0" />
        {t("settings.adminsSection.accessDesc")}
      </div>

      <div className="grid gap-4 lg:grid-cols-2">
        <GlassCard className="space-y-3" glow>
          <div className="flex items-center gap-2 border-b border-border/30 pb-2 mb-2">
            <span className="h-2 w-2 rounded-full bg-accent" />
            <h3 className="text-sm font-bold text-fg">{t("settings.adminsSection.permissionsRef")}</h3>
          </div>
          <div className="flex flex-wrap gap-1.5">
            {ALL_PERMISSIONS.map((p) => (
              <span
                key={p}
                className="rounded-lg bg-surface-2/50 border border-border/30 px-2 py-0.5 font-mono text-[10px] text-fg-muted"
              >
                {p}
              </span>
            ))}
          </div>
        </GlassCard>

        <GlassCard className="space-y-3" glow>
          <div className="flex items-center gap-2 border-b border-border/30 pb-2 mb-2">
            <span className="h-2 w-2 rounded-full bg-primary" />
            <h3 className="text-sm font-bold text-fg">{t("settings.adminsSection.resellerSettingsRef")}</h3>
          </div>
          <ul className="space-y-1.5 text-sm text-fg-muted">
            {RESELLER_SETTING_KEYS.map((key) => (
              <li key={key} className="flex items-center gap-2 hover:bg-surface-2/30 rounded-lg px-2 py-1 -mx-2 transition-colors">
                <span className="font-mono text-xs text-fg min-w-[80px]">{key}</span>
                <span className="text-fg-subtle">—</span>
                <span>{t(SETTING_LABEL_KEYS[key])}</span>
              </li>
            ))}
          </ul>
        </GlassCard>
      </div>

      <GlassCard className="!p-0 overflow-hidden" glow>
        <div className="flex items-center justify-between border-b border-border/40 px-5 py-3">
          <span className="text-sm font-bold text-fg">{t("settings.adminsSection.resellerAccessMatrix")}</span>
          <span className="text-[10px] text-fg-subtle tabular-nums">{resellerList.length} resellers</span>
        </div>
        <div className="overflow-x-auto">
          <table className="w-full text-sm">
            <thead>
              <tr className="border-b border-border/40 bg-surface-2/30 text-left text-[11px] uppercase tracking-wide text-fg-subtle">
                <th className="px-5 py-3 font-medium">{t("reseller.admins.colUsername")}</th>
                <th className="px-5 py-3 font-medium">{t("reseller.admins.colRole")}</th>
                <th className="px-5 py-3 font-medium">{t("settings.adminsSection.colEnabledSettings")}</th>
                <th className="px-5 py-3"></th>
              </tr>
            </thead>
            <tbody className="divide-y divide-border/20">
              {resellerList.map((a) => {
                const merged = mergeResellerSettings(a.reseller_settings);
                const enabled = RESELLER_SETTING_KEYS.filter((k) => merged[k]);
                return (
                  <tr key={a.id} className="hover:bg-surface-2/40 transition-colors duration-150">
                    <td className="px-5 py-3 font-medium">
                      <Link to={`/settings/admins/${a.id}`} className="text-primary hover:underline">
                        {a.username}
                      </Link>
                    </td>
                    <td className="px-5 py-3 text-fg-muted">{roleName(a.role_id)}</td>
                    <td className="px-5 py-3">
                      <div className="flex flex-wrap gap-1">
                        {enabled.length === 0 ? (
                          <span className="text-fg-muted">—</span>
                        ) : (
                          enabled.map((k) => (
                            <Badge key={k}>{t(SETTING_LABEL_KEYS[k])}</Badge>
                          ))
                        )}
                      </div>
                    </td>
                    <td className="px-5 py-3 text-right">
                      <Link
                        to={`/settings/admins/${a.id}`}
                        className="text-sm font-medium text-primary hover:underline inline-flex items-center gap-1"
                      >
                        {t("settings.adminsSection.viewDetail")}
                        <span aria-hidden>&rarr;</span>
                      </Link>
                    </td>
                  </tr>
                );
              })}
              {resellerList.length === 0 && (
                <tr>
                  <td colSpan={4} className="px-5 py-6 text-center text-fg-muted">
                    {t("reseller.admins.noRoles")}
                  </td>
                </tr>
              )}
            </tbody>
          </table>
        </div>
      </GlassCard>
    </div>
  );
}
