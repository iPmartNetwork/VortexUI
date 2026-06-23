import { useState } from "react";
import { useAdmins, useAdjustAdminQuota, useDeleteAdmin, useDeleteRole, useRoles, useUnsuspendAdmin } from "@/api/admin-hooks";
import { useImpersonateAdmin } from "@/api/reseller-hooks";
import { useResellerQuotaUsage } from "@/api/quota-hooks";
import { setToken } from "@/api/client";
import { useAuth } from "@/auth/auth";
import { Badge, Button, Card } from "@/components/ui";
import { CreateAdminModal } from "@/components/CreateAdminModal";
import { CreateRoleModal } from "@/components/CreateRoleModal";
import { EditAdminModal } from "@/components/EditAdminModal";
import { EditRoleModal } from "@/components/EditRoleModal";
import { useConfirm } from "@/components/confirm";
import { useToast } from "@/components/toast";
import { useI18n } from "@/i18n/i18n";
import type { Admin, Role } from "@/api/types";
import { formatBytes } from "@/lib/utils";

function fill(template: string, vars: Record<string, string>) {
  let out = template;
  for (const [k, v] of Object.entries(vars)) out = out.replace(`{${k}}`, v);
  return out;
}

export function Admins() {
  const { sudo, refreshSession } = useAuth();
  const { t } = useI18n();
  const admins = useAdmins();
  const impersonate = useImpersonateAdmin();
  const quotaUsage = useResellerQuotaUsage();
  const adjustQuota = useAdjustAdminQuota();
  const roles = useRoles();
  const del = useDeleteAdmin();
  const unsuspend = useUnsuspendAdmin();
  const delRole = useDeleteRole();
  const confirm = useConfirm();
  const toast = useToast();
  const [adminOpen, setAdminOpen] = useState(false);
  const [roleOpen, setRoleOpen] = useState(false);
  const [editAdmin, setEditAdmin] = useState<Admin | null>(null);
  const [editRole, setEditRole] = useState<Role | null>(null);

  const roleName = (id: string | null) => roles.data?.roles.find((r) => r.id === id)?.name ?? "—";
  const usageFor = (adminId: string) => quotaUsage.data?.usage.find((u) => u.admin_id === adminId);

  const loading = admins.isLoading || roles.isLoading;
  const loadError = admins.isError || roles.isError;
  const adminList = admins.data?.admins ?? [];
  const roleList = roles.data?.roles ?? [];
  const usageList = quotaUsage.data?.usage ?? [];

  if (loading) {
    return <p className="text-sm text-muted-foreground">{t("common.loading")}</p>;
  }

  if (loadError) {
    return (
      <p className="text-sm text-destructive">
        {t("reseller.admins.loadFailed")}
      </p>
    );
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

  async function removeRole(r: Role) {
    const ok = await confirm({
      title: fill(t("reseller.admins.deleteRoleTitle"), { name: r.name }),
      message: t("reseller.admins.deleteRoleMsg"),
      confirmLabel: t("common.delete"),
      destructive: true,
    });
    if (!ok) return;
    try {
      await delRole.mutateAsync(r.id);
      toast.success(fill(t("reseller.admins.deleteRoleOk"), { name: r.name }));
    } catch {
      toast.error(t("reseller.admins.deleteRoleFail"));
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
    <div className="space-y-8">
      {adminOpen && <CreateAdminModal open onClose={() => setAdminOpen(false)} />}
      {roleOpen && <CreateRoleModal open onClose={() => setRoleOpen(false)} />}
      {editAdmin && <EditAdminModal admin={editAdmin} onClose={() => setEditAdmin(null)} />}
      {editRole && <EditRoleModal role={editRole} onClose={() => setEditRole(null)} />}

      <div>
        <div className="mb-4 flex items-center justify-between">
          <h1 className="text-2xl font-bold tracking-tight">{t("reseller.admins.pageTitle")}</h1>
          <Button onClick={() => setAdminOpen(true)}>{t("reseller.admins.newAdmin")}</Button>
        </div>

        {usageList.length > 0 && (
          <Card className="mb-6 p-0">
            <div className="border-b px-5 py-3 text-sm font-semibold">{t("reseller.admins.quotaUsage")}</div>
            <table className="w-full text-sm">
              <thead className="border-b text-left text-muted-foreground">
                <tr>
                  <th className="px-5 py-3 font-medium">{t("reseller.admins.colReseller")}</th>
                  <th className="px-5 py-3 font-medium">{t("reseller.admins.colAccounts")}</th>
                  <th className="px-5 py-3 font-medium">{t("reseller.admins.colAssigned")}</th>
                  <th className="px-5 py-3 font-medium">{t("reseller.admins.colConsumed")}</th>
                  <th className="px-5 py-3 font-medium">{t("reseller.admins.colPoolLeft")}</th>
                  <th className="px-5 py-3 font-medium"></th>
                </tr>
              </thead>
              <tbody>
                {usageList.map((u) => (
                  <tr key={u.admin_id} className="border-b last:border-0">
                    <td className="px-5 py-3 font-medium">{u.username}</td>
                    <td className="px-5 py-3 text-muted-foreground">
                      {u.user_count}{u.user_quota > 0 ? ` / ${u.user_quota}` : ""}
                      {u.users_remaining != null && <span className="ms-1 text-xs">({u.users_remaining} {t("reseller.admins.left")})</span>}
                    </td>
                    <td className="px-5 py-3 text-muted-foreground">{formatBytes(u.traffic_allocated, false)}</td>
                    <td className="px-5 py-3 text-muted-foreground">{formatBytes(u.traffic_used, false)}</td>
                    <td className="px-5 py-3 text-muted-foreground">
                      {u.traffic_remaining != null ? formatBytes(u.traffic_remaining, false) : "∞"}
                    </td>
                    <td className="px-5 py-3 text-right space-x-1">
                      <Button variant="ghost" size="sm" onClick={() => quickAdjust(u.admin_id, 50)}>{t("reseller.admins.addUsers")}</Button>
                      <Button variant="ghost" size="sm" onClick={() => quickAdjust(u.admin_id, undefined, 10)}>{t("reseller.admins.addTraffic10")}</Button>
                      <Button variant="ghost" size="sm" onClick={() => quickAdjust(u.admin_id, undefined, 50)}>{t("reseller.admins.addTraffic50")}</Button>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </Card>
        )}

        <Card className="p-0">
          <table className="w-full text-sm">
            <thead className="border-b text-left text-muted-foreground">
              <tr>
                <th className="px-5 py-3 font-medium">{t("reseller.admins.colUsername")}</th>
                <th className="px-5 py-3 font-medium">{t("reseller.admins.colAccess")}</th>
                <th className="px-5 py-3 font-medium">{t("reseller.admins.colRole")}</th>
                <th className="px-5 py-3 font-medium">{t("reseller.admins.col2fa")}</th>
                <th className="px-5 py-3 font-medium hidden lg:table-cell">{t("reseller.admins.colUsage")}</th>
                <th className="px-5 py-3"></th>
              </tr>
            </thead>
            <tbody>
              {adminList.map((a) => (
                <tr key={a.id} className="border-b last:border-0 hover:bg-muted/40">
                  <td className="px-5 py-3 font-medium">
                    {a.username}
                    {a.suspended && <span className="ms-2"><Badge color="expired">{t("reseller.admins.suspended")}</Badge></span>}
                  </td>
                  <td className="px-5 py-3">
                    {a.sudo ? <Badge color="active">{t("reseller.admins.sudo")}</Badge> : <Badge>{t("reseller.admins.reseller")}</Badge>}
                  </td>
                  <td className="px-5 py-3 text-muted-foreground">{a.sudo ? "—" : roleName(a.role_id)}</td>
                  <td className="px-5 py-3 text-muted-foreground">{a.totp_enabled ? t("reseller.admins.totpOn") : t("reseller.admins.totpOff")}</td>
                  <td className="px-5 py-3 text-muted-foreground hidden lg:table-cell">
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
                    {!a.sudo && (
                      <Button variant="ghost" onClick={() => setEditAdmin(a)}>{t("common.edit")}</Button>
                    )}
                    {sudo && a.suspended && (
                      <Button variant="ghost" onClick={async () => {
                        try {
                          await unsuspend.mutateAsync(a.id);
                          toast.success(fill(t("reseller.admins.unsuspendOk"), { name: a.username }));
                        } catch {
                          toast.error(t("reseller.admins.unsuspendFail"));
                        }
                      }}>{t("reseller.admins.unsuspend")}</Button>
                    )}
                    <Button variant="ghost" className="text-destructive" onClick={() => remove(a)}>
                      {t("common.delete")}
                    </Button>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </Card>
      </div>

      <div>
        <div className="mb-4 flex items-center justify-between">
          <h2 className="text-lg font-semibold">{t("reseller.admins.roles")}</h2>
          <Button variant="ghost" onClick={() => setRoleOpen(true)}>{t("reseller.admins.newRole")}</Button>
        </div>
        <div className="grid grid-cols-1 gap-3 md:grid-cols-2 lg:grid-cols-3">
          {roleList.map((r) => (
            <Card key={r.id} className="space-y-2">
              <div className="flex items-start justify-between gap-2">
                <div className="font-medium">{r.name}</div>
                <div className="flex shrink-0 gap-1">
                  <Button variant="ghost" size="sm" onClick={() => setEditRole(r)}>{t("common.edit")}</Button>
                  <Button variant="ghost" size="sm" className="text-destructive" onClick={() => removeRole(r)}>{t("common.delete")}</Button>
                </div>
              </div>
              <div className="flex flex-wrap gap-1">
                {r.permissions.map((p) => (
                  <span key={p} className="rounded bg-muted px-1.5 py-0.5 font-mono text-[11px] text-muted-foreground">
                    {p}
                  </span>
                ))}
              </div>
            </Card>
          ))}
          {roleList.length === 0 && (
            <p className="text-sm text-muted-foreground">{t("reseller.admins.noRoles")}</p>
          )}
        </div>
      </div>
    </div>
  );
}
