import { useState } from "react";
import { Link } from "react-router-dom";
import { Plus } from "lucide-react";
import { useAdmins, useAdjustAdminQuota, useDeleteAdmin, useRoles, useUnsuspendAdmin } from "@/api/admin-hooks";
import { useImpersonateAdmin } from "@/api/reseller-hooks";
import { useResellerQuotaUsage } from "@/api/quota-hooks";
import { setToken } from "@/api/client";
import { useAuth } from "@/auth/auth";
import { Badge, Button } from "@/components/ui";
import { GlassCard } from "@/components/veltrix";
import { CreateAdminModal } from "@/components/CreateAdminModal";
import { EditAdminModal } from "@/components/EditAdminModal";
import { WalletTopUpModal } from "@/components/WalletTopUpModal";
import { useConfirm } from "@/components/confirm";
import { useToast } from "@/components/toast";
import { useI18n } from "@/i18n/i18n";
import type { Admin } from "@/api/types";
import { formatBytes } from "@/lib/utils";

function fill(template: string, vars: Record<string, string>) {
  let out = template;
  for (const [k, v] of Object.entries(vars)) out = out.replace(`{${k}}`, v);
  return out;
}

export function AdminsListTab(_props: { embedded?: boolean }) {
  const { sudo, refreshSession } = useAuth();
  const { t } = useI18n();
  const admins = useAdmins();
  const impersonate = useImpersonateAdmin();
  const quotaUsage = useResellerQuotaUsage();
  const adjustQuota = useAdjustAdminQuota();
  const roles = useRoles();
  const del = useDeleteAdmin();
  const unsuspend = useUnsuspendAdmin();
  const confirm = useConfirm();
  const toast = useToast();
  const [adminOpen, setAdminOpen] = useState(false);
  const [editAdmin, setEditAdmin] = useState<Admin | null>(null);
  const [walletTopUp, setWalletTopUp] = useState<{ id: string; username: string } | null>(null);

  const roleName = (id: string | null) => roles.data?.roles.find((r) => r.id === id)?.name ?? "—";
  const usageFor = (adminId: string) => quotaUsage.data?.usage.find((u) => u.admin_id === adminId);

  const loading = admins.isLoading || roles.isLoading;
  const loadError = admins.isError || roles.isError;
  const adminList = admins.data?.admins ?? [];
  const usageList = quotaUsage.data?.usage ?? [];

  if (loading) {
    return <p className="text-sm text-fg-muted">{t("common.loading")}</p>;
  }

  if (loadError) {
    return <p className="text-sm text-danger">{t("reseller.admins.loadFailed")}</p>;
  }

  async function loginAs(a: Admin) {
    try {
      const res = await impersonate.mutateAsync(a.id);
      setToken(res.token);
      await refreshSession();
      toast.success(fill(t("reseller.admins.impersonateOk"), { name: a.username }));
    } catch {
      toast.error(t("reseller.admins.impersonateFail"));
    }
  }

  async function remove(a: Admin) {
    const ok = await confirm({
      title: fill(t("reseller.admins.deleteAdminTitle"), { name: a.username }),
      confirmLabel: t("common.delete"),
      destructive: true,
    });
    if (!ok) return;
    try {
      await del.mutateAsync(a.id);
      toast.success(fill(t("reseller.admins.deleteAdminOk"), { name: a.username }));
    } catch {
      toast.error(t("reseller.admins.deleteAdminFail"));
    }
  }

  async function quickAdjust(adminId: string, userDelta?: number, trafficGb?: number) {
    try {
      await adjustQuota.mutateAsync({
        id: adminId,
        user_quota_delta: userDelta,
        traffic_quota_delta: trafficGb ? trafficGb * 1024 * 1024 * 1024 : undefined,
      });
      toast.success(t("reseller.admins.adjustOk"));
    } catch {
      toast.error(t("reseller.admins.adjustFail"));
    }
  }

  return (
    <div className="space-y-6">
      {adminOpen && <CreateAdminModal open onClose={() => setAdminOpen(false)} />}
      {editAdmin && <EditAdminModal admin={editAdmin} onClose={() => setEditAdmin(null)} />}
      {walletTopUp && (
        <WalletTopUpModal
          open
          adminId={walletTopUp.id}
          username={walletTopUp.username}
          onClose={() => setWalletTopUp(null)}
        />
      )}

      <div className="flex justify-end">
        <Button onClick={() => setAdminOpen(true)}><Plus size={14} /> {t("reseller.admins.newAdmin")}</Button>
      </div>

      {usageList.length > 0 && (
        <GlassCard hover={false} className="!p-0 overflow-hidden">
          <div className="border-b border-border/40 px-5 py-3 text-sm font-bold text-fg">{t("reseller.admins.quotaUsage")}</div>
          <div className="overflow-x-auto">
            <table className="w-full text-sm">
              <thead>
                <tr className="border-b border-border/40 bg-surface-2/30 text-left text-[11px] uppercase tracking-wide text-fg-subtle">
                  <th className="px-5 py-3 font-medium">{t("reseller.admins.colReseller")}</th>
                  <th className="px-5 py-3 font-medium">{t("reseller.admins.colAccounts")}</th>
                  <th className="px-5 py-3 font-medium">{t("reseller.admins.colAssigned")}</th>
                  <th className="px-5 py-3 font-medium">{t("reseller.admins.colConsumed")}</th>
                  <th className="px-5 py-3 font-medium">{t("reseller.admins.colPoolLeft")}</th>
                  <th className="px-5 py-3 font-medium">{t("reseller.admins.colWallet")}</th>
                  <th className="px-5 py-3 font-medium"></th>
                </tr>
              </thead>
              <tbody className="divide-y divide-border/20">
                {usageList.map((u) => (
                  <tr key={u.admin_id} className="hover:bg-surface-2/40">
                    <td className="px-5 py-3 font-medium">
                      <Link
                        to={`/settings/admins/${u.admin_id}`}
                        className="text-primary hover:underline"
                      >
                        {u.username}
                      </Link>
                    </td>
                    <td className="px-5 py-3 text-fg-muted">
                      {u.user_count}{u.user_quota > 0 ? ` / ${u.user_quota}` : ""}
                      {u.users_remaining != null && (
                        <span className="ms-1 text-xs">({u.users_remaining} {t("reseller.admins.left")})</span>
                      )}
                    </td>
                    <td className="px-5 py-3 text-fg-muted">{formatBytes(u.traffic_allocated, false)}</td>
                    <td className="px-5 py-3 text-fg-muted">{formatBytes(u.traffic_used, false)}</td>
                    <td className="px-5 py-3 text-fg-muted">
                      {u.traffic_remaining != null ? formatBytes(u.traffic_remaining, false) : "∞"}
                    </td>
                    <td className="px-5 py-3 text-fg-muted">
                      {formatBytes(u.wallet_traffic_bytes ?? 0, false)} · {u.wallet_user_credits ?? 0}
                    </td>
                    <td className="px-5 py-3 text-right space-x-1">
                      <Button variant="ghost" size="sm" onClick={() => setWalletTopUp({ id: u.admin_id, username: u.username })}>
                        {t("reseller.admins.walletTopUp")}
                      </Button>
                      <Button variant="ghost" size="sm" onClick={() => quickAdjust(u.admin_id, 50)}>{t("reseller.admins.addUsers")}</Button>
                      <Button variant="ghost" size="sm" onClick={() => quickAdjust(u.admin_id, undefined, 10)}>{t("reseller.admins.addTraffic10")}</Button>
                      <Button variant="ghost" size="sm" onClick={() => quickAdjust(u.admin_id, undefined, 50)}>{t("reseller.admins.addTraffic50")}</Button>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </GlassCard>
      )}

      <GlassCard hover={false} className="!p-0 overflow-hidden">
        <div className="overflow-x-auto">
          <table className="w-full text-sm">
            <thead>
              <tr className="border-b border-border/40 bg-surface-2/30 text-left text-[11px] uppercase tracking-wide text-fg-subtle">
                <th className="px-5 py-3 font-medium">{t("reseller.admins.colUsername")}</th>
                <th className="px-5 py-3 font-medium">{t("reseller.admins.colAccess")}</th>
                <th className="px-5 py-3 font-medium">{t("reseller.admins.colRole")}</th>
                <th className="px-5 py-3 font-medium">{t("reseller.admins.col2fa")}</th>
                <th className="px-5 py-3 font-medium hidden lg:table-cell">{t("reseller.admins.colUsage")}</th>
                <th className="px-5 py-3"></th>
              </tr>
            </thead>
            <tbody className="divide-y divide-border/20">
              {adminList.map((a) => (
                <tr key={a.id} className="hover:bg-surface-2/40">
                  <td className="px-5 py-3 font-medium">
                    {!a.sudo ? (
                      <Link to={`/settings/admins/${a.id}`} className="text-primary hover:underline">
                        {a.username}
                      </Link>
                    ) : (
                      a.username
                    )}
                    {a.suspended && (
                      <span className="ms-2">
                        <Badge color="expired">{t("reseller.admins.suspended")}</Badge>
                      </span>
                    )}
                  </td>
                  <td className="px-5 py-3">
                    {a.sudo ? <Badge color="active">{t("reseller.admins.sudo")}</Badge> : <Badge>{t("reseller.admins.reseller")}</Badge>}
                  </td>
                  <td className="px-5 py-3 text-fg-muted">{a.sudo ? "—" : roleName(a.role_id)}</td>
                  <td className="px-5 py-3 text-fg-muted">{a.totp_enabled ? t("reseller.admins.totpOn") : t("reseller.admins.totpOff")}</td>
                  <td className="px-5 py-3 text-fg-muted hidden lg:table-cell">
                    {!a.sudo && (() => {
                      const u = usageFor(a.id);
                      if (!u) return "—";
                      return `${u.user_count}${u.user_quota > 0 ? `/${u.user_quota}` : ""} · ${formatBytes(u.traffic_used, false)} ${t("reseller.admins.used")}`;
                    })()}
                  </td>
                  <td className="px-5 py-3 text-right space-x-1">
                    {sudo && !a.sudo && (
                      <Button variant="ghost" onClick={() => loginAs(a)}>{t("reseller.admins.loginAs")}</Button>
                    )}
                    {sudo && !a.sudo && (
                      <Button variant="ghost" onClick={() => setWalletTopUp({ id: a.id, username: a.username })}>
                        {t("reseller.admins.walletTopUp")}
                      </Button>
                    )}
                    {!a.sudo && (
                      <Button variant="ghost" onClick={() => setEditAdmin(a)}>{t("common.edit")}</Button>
                    )}
                    {sudo && a.suspended && (
                      <Button
                        variant="ghost"
                        onClick={async () => {
                          try {
                            await unsuspend.mutateAsync(a.id);
                            toast.success(fill(t("reseller.admins.unsuspendOk"), { name: a.username }));
                          } catch {
                            toast.error(t("reseller.admins.unsuspendFail"));
                          }
                        }}
                      >
                        {t("reseller.admins.unsuspend")}
                      </Button>
                    )}
                    <Button variant="ghost" className="text-danger" onClick={() => remove(a)}>
                      {t("common.delete")}
                    </Button>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </GlassCard>
    </div>
  );
}
