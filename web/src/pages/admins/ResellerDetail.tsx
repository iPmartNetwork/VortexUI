import { useState } from "react";
import { useParams, useNavigate, Link, Navigate } from "react-router-dom";
import {
  ArrowLeft, Users, HardDrive, Wallet, TrendingUp, Gauge, Clock, Shield,
} from "lucide-react";
import { useAdmins, useRoles, useUnsuspendAdmin } from "@/api/admin-hooks";
import { useAdminQuotaUsage } from "@/api/quota-hooks";
import { useAdminWallet, useImpersonateAdmin } from "@/api/reseller-hooks";
import { setToken } from "@/api/client";
import { useAuth } from "@/auth/auth";
import { mergeResellerSettings, RESELLER_SETTING_KEYS, type ResellerSettingKey } from "@/auth/permissions";
import { Badge, Button } from "@/components/ui";
import { GlassCard } from "@/components/veltrix";
import { EditAdminModal } from "@/components/EditAdminModal";
import { WalletTopUpModal } from "@/components/WalletTopUpModal";
import { useToast } from "@/components/toast";
import { useI18n } from "@/i18n/i18n";
import type { TKey } from "@/i18n/dict";
import { useTitle } from "@/lib/useTitle";
import { formatBytes, pct } from "@/lib/utils";

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

function QuotaBar({ label, used, limit, format = "number" }: { label: string; used: number; limit: number; format?: "number" | "bytes" }) {
  const unlimited = limit <= 0;
  const displayUsed = format === "bytes" ? formatBytes(used, false) : String(used);
  const displayLimit = unlimited ? "∞" : format === "bytes" ? formatBytes(limit, false) : String(limit);
  const p = unlimited ? 0 : pct(used, limit);
  return (
    <GlassCard className="space-y-3">
      <div className="flex items-center justify-between text-sm">
        <span className="font-medium text-fg">{label}</span>
        <span className="text-fg-muted">{displayUsed} / {displayLimit}</span>
      </div>
      {!unlimited && (
        <div className="h-2 rounded-full bg-surface-2">
          <div
            className={`h-full rounded-full transition-all ${p >= 90 ? "bg-danger" : p >= 75 ? "bg-warning" : "bg-primary"}`}
            style={{ width: `${Math.min(p, 100)}%` }}
          />
        </div>
      )}
    </GlassCard>
  );
}

function fill(template: string, vars: Record<string, string>) {
  let out = template;
  for (const [k, v] of Object.entries(vars)) out = out.replace(`{${k}}`, v);
  return out;
}

export function ResellerDetail() {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const { sudo, refreshSession } = useAuth();
  const { t } = useI18n();
  const toast = useToast();
  const admins = useAdmins();
  const roles = useRoles();
  const quota = useAdminQuotaUsage(id ?? null);
  const wallet = useAdminWallet(id ?? null);
  const impersonate = useImpersonateAdmin();
  const unsuspend = useUnsuspendAdmin();
  const [editOpen, setEditOpen] = useState(false);
  const [walletOpen, setWalletOpen] = useState(false);

  const admin = admins.data?.admins.find((a) => a.id === id);
  useTitle(admin?.username ?? t("reseller.detail.title"));

  if (!sudo) {
    return <Navigate to="/settings" replace />;
  }

  if (admins.isLoading) {
    return <p className="py-20 text-center text-fg-muted">{t("common.loading")}</p>;
  }

  if (!admin) {
    return (
      <div className="space-y-4 py-12 text-center">
        <p className="text-fg-muted">{t("reseller.detail.notFound")}</p>
        <Button variant="outline" onClick={() => navigate("/settings?tab=admins")}>
          {t("reseller.detail.back")}
        </Button>
      </div>
    );
  }

  if (admin.sudo) {
    return (
      <div className="space-y-4 animate-page-enter">
        <Header username={admin.username} onBack={() => navigate("/settings?tab=admins")} />
        <GlassCard className="text-sm text-fg-muted">{t("reseller.detail.sudoAdmin")}</GlassCard>
      </div>
    );
  }

  const u = quota.data?.usage;
  const roleName = roles.data?.roles.find((r) => r.id === admin.role_id)?.name ?? "—";
  const trafficMode = u?.traffic_quota_mode === "consumed" ? "consumed" : "allocated";
  const trafficUsed = trafficMode === "consumed" ? (u?.traffic_used ?? 0) : (u?.traffic_allocated ?? 0);
  const trafficLabel = trafficMode === "consumed" ? t("reseller.detail.trafficConsumed") : t("reseller.detail.trafficAllocated");
  const mergedSettings = mergeResellerSettings(admin.reseller_settings);
  const enabledSettings = RESELLER_SETTING_KEYS.filter((k) => mergedSettings[k]);
  const ledger = wallet.data?.ledger ?? [];
  const w = wallet.data?.wallet;

  async function loginAs() {
    try {
      const res = await impersonate.mutateAsync(admin!.id);
      setToken(res.token);
      await refreshSession();
      toast.success(fill(t("reseller.admins.impersonateOk"), { name: admin!.username }));
    } catch {
      toast.error(t("reseller.admins.impersonateFail"));
    }
  }

  return (
    <div className="space-y-6 animate-page-enter">
      {editOpen && <EditAdminModal admin={admin} onClose={() => setEditOpen(false)} />}
      {walletOpen && (
        <WalletTopUpModal open adminId={admin.id} username={admin.username} onClose={() => setWalletOpen(false)} />
      )}

      <div className="flex flex-wrap items-start justify-between gap-4">
        <Header username={admin.username} onBack={() => navigate("/settings?tab=admins")} suspended={admin.suspended} totp={admin.totp_enabled} role={roleName} />
        <div className="flex flex-wrap gap-2">
          <Button variant="outline" size="sm" onClick={() => setEditOpen(true)}>{t("common.edit")}</Button>
          <Button variant="outline" size="sm" onClick={() => setWalletOpen(true)}>{t("reseller.admins.walletTopUp")}</Button>
          <Button variant="outline" size="sm" onClick={loginAs}>{t("reseller.admins.loginAs")}</Button>
          {admin.suspended && (
            <Button
              variant="outline"
              size="sm"
              onClick={async () => {
                try {
                  await unsuspend.mutateAsync(admin.id);
                  toast.success(fill(t("reseller.admins.unsuspendOk"), { name: admin.username }));
                } catch {
                  toast.error(t("reseller.admins.unsuspendFail"));
                }
              }}
            >
              {t("reseller.admins.unsuspend")}
            </Button>
          )}
        </div>
      </div>

      {quota.isLoading && <p className="text-sm text-fg-muted">{t("common.loading")}</p>}

      {u && (
        <>
          <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
            <StatCard icon={Users} label={t("reseller.detail.accounts")} value={`${u.user_count}${u.user_quota > 0 ? ` / ${u.user_quota}` : ""}`} />
            <StatCard
              icon={TrendingUp}
              label={t("reseller.detail.trafficUsed")}
              value={formatBytes(u.traffic_used, false)}
            />
            <StatCard
              icon={HardDrive}
              label={t("reseller.detail.poolRemaining")}
              value={u.traffic_remaining != null ? formatBytes(u.traffic_remaining, false) : "∞"}
            />
            <StatCard icon={Gauge} label={t("reseller.detail.quotaMode")} value={trafficMode} />
          </div>

          <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
            <StatCard icon={Wallet} label={t("reseller.detail.walletTraffic")} value={formatBytes(w?.traffic_bytes ?? u.wallet_traffic_bytes ?? 0, false)} />
            <StatCard icon={Wallet} label={t("reseller.detail.walletCredits")} value={String(w?.user_credits ?? u.wallet_user_credits ?? 0)} />
            <StatCard icon={HardDrive} label={t("reseller.detail.trafficAllocated")} value={formatBytes(u.traffic_allocated, false)} />
            <StatCard
              icon={Clock}
              label={t("reseller.detail.trafficRemaining")}
              value={u.users_remaining != null ? String(u.users_remaining) : "∞"}
              sub={t("reseller.admins.left")}
            />
          </div>

          <div className="grid gap-4 md:grid-cols-2">
            <QuotaBar label={t("reseller.dashboard.userAccounts")} used={u.user_count} limit={u.user_quota} />
            <QuotaBar label={trafficLabel} used={trafficUsed} limit={u.traffic_quota} format="bytes" />
          </div>
        </>
      )}

      <div className="grid gap-4 lg:grid-cols-2">
        <GlassCard className="space-y-4">
          <h3 className="text-sm font-bold text-fg">{t("reseller.detail.info")}</h3>
          <dl className="space-y-2 text-sm">
            <Row label={t("reseller.admins.colRole")} value={roleName} />
            <Row label={t("reseller.detail.created")} value={new Date(admin.created_at).toLocaleString()} />
            <Row
              label={t("reseller.detail.lastLogin")}
              value={admin.last_login ? new Date(admin.last_login).toLocaleString() : "—"}
            />
            <Row label={t("reseller.editAdmin.userQuota")} value={admin.user_quota <= 0 ? "∞" : String(admin.user_quota)} />
            <Row
              label={t("reseller.editAdmin.trafficQuota")}
              value={admin.traffic_quota <= 0 ? "∞" : formatBytes(admin.traffic_quota, false)}
            />
          </dl>
        </GlassCard>

        <GlassCard className="space-y-4">
          <h3 className="text-sm font-bold text-fg flex items-center gap-2">
            <Shield size={16} />
            {t("reseller.detail.policies")}
          </h3>
          <dl className="space-y-2 text-sm">
            <Row
              label={t("reseller.editAdmin.maxDataLimit")}
              value={!admin.policy_max_data_limit ? "∞" : formatBytes(admin.policy_max_data_limit, false)}
            />
            <Row
              label={t("reseller.editAdmin.maxExpire")}
              value={!admin.policy_max_expire_days ? "∞" : `${admin.policy_max_expire_days}d`}
            />
            <Row label={t("reseller.editAdmin.allowBulkCreate")} value={admin.policy_allow_bulk_create ? t("common.enabled") : t("common.disabled")} />
            <Row label={t("reseller.editAdmin.allowBulkDelete")} value={admin.policy_allow_bulk_delete ? t("common.enabled") : t("common.disabled")} />
            <Row label={t("reseller.editAdmin.enableAutoSuspend")} value={admin.auto_suspend_enabled ? t("common.enabled") : t("common.disabled")} />
          </dl>
        </GlassCard>
      </div>

      <GlassCard className="space-y-3">
        <h3 className="text-sm font-bold text-fg">{t("reseller.detail.resellerSettings")}</h3>
        <div className="flex flex-wrap gap-1.5">
          {enabledSettings.length === 0 ? (
            <span className="text-sm text-fg-muted">—</span>
          ) : (
            enabledSettings.map((k) => <Badge key={k}>{t(SETTING_LABEL_KEYS[k])}</Badge>)
          )}
        </div>
        <p className="text-xs text-fg-muted">
          <Link to={`/settings?tab=admins&section=access`} className="text-primary hover:underline">
            {t("settings.adminsSection.access")}
          </Link>
        </p>
      </GlassCard>

      <GlassCard className="!p-0 overflow-hidden">
        <div className="border-b border-border/40 px-5 py-3 text-sm font-bold text-fg">{t("reseller.detail.ledger")}</div>
        {wallet.isLoading && <p className="px-5 py-4 text-sm text-fg-muted">{t("common.loading")}</p>}
        {!wallet.isLoading && ledger.length === 0 && (
          <p className="px-5 py-6 text-sm text-fg-muted">{t("reseller.detail.noLedger")}</p>
        )}
        {ledger.length > 0 && (
          <div className="overflow-x-auto">
            <table className="w-full text-sm">
              <thead>
                <tr className="border-b border-border/40 bg-surface-2/30 text-left text-[11px] uppercase tracking-wide text-fg-subtle">
                  <th className="px-5 py-3 font-medium">{t("reseller.detail.ledgerColDate")}</th>
                  <th className="px-5 py-3 font-medium">{t("reseller.detail.ledgerColTraffic")}</th>
                  <th className="px-5 py-3 font-medium">{t("reseller.detail.ledgerColUsers")}</th>
                  <th className="px-5 py-3 font-medium">{t("reseller.detail.ledgerColReason")}</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-border/20">
                {ledger.map((e) => (
                  <tr key={e.id} className="hover:bg-surface-2/40">
                    <td className="px-5 py-3 text-fg-muted">{new Date(e.created_at).toLocaleString()}</td>
                    <td className="px-5 py-3">{e.delta_traffic !== 0 ? formatBytes(e.delta_traffic, false) : "—"}</td>
                    <td className="px-5 py-3">{e.delta_users !== 0 ? (e.delta_users > 0 ? `+${e.delta_users}` : e.delta_users) : "—"}</td>
                    <td className="px-5 py-3 text-fg-muted">{e.reason || "—"}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </GlassCard>
    </div>
  );
}

function Header({
  username,
  onBack,
  suspended,
  totp,
  role,
}: {
  username: string;
  onBack: () => void;
  suspended?: boolean;
  totp?: boolean;
  role?: string;
}) {
  const { t } = useI18n();
  return (
    <div className="flex items-center gap-4">
      <button
        type="button"
        onClick={onBack}
        className="grid h-9 w-9 place-items-center rounded-xl text-fg-muted transition hover:bg-surface-2/60 hover:text-fg"
      >
        <ArrowLeft size={18} />
      </button>
      <div>
        <h1 className="text-xl font-bold text-fg">{username}</h1>
        <div className="mt-0.5 flex flex-wrap items-center gap-2">
          <Badge>{t("reseller.admins.reseller")}</Badge>
          {role && <span className="text-xs text-fg-muted">{role}</span>}
          {suspended && <Badge color="expired">{t("reseller.admins.suspended")}</Badge>}
          <span className="text-xs text-fg-muted">
            2FA: {totp ? t("reseller.admins.totpOn") : t("reseller.admins.totpOff")}
          </span>
        </div>
      </div>
    </div>
  );
}

function StatCard({
  icon: Icon,
  label,
  value,
  sub,
}: {
  icon: typeof Users;
  label: string;
  value: string;
  sub?: string;
}) {
  return (
    <GlassCard className="flex items-center gap-3">
      <Icon className="text-primary shrink-0" size={22} />
      <div>
        <div className="text-xs text-fg-muted">{label}</div>
        <div className="text-lg font-bold text-fg capitalize">
          {value}
          {sub && <span className="ms-1 text-xs font-normal text-fg-muted">{sub}</span>}
        </div>
      </div>
    </GlassCard>
  );
}

function Row({ label, value }: { label: string; value: string }) {
  return (
    <div className="flex justify-between gap-4">
      <dt className="text-fg-muted">{label}</dt>
      <dd className="font-medium text-fg text-end">{value}</dd>
    </div>
  );
}
