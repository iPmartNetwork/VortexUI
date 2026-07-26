import { useState } from "react";
import { Plus } from "lucide-react";
import { useDeleteRole, useRoles } from "@/api/admin-hooks";
import { Button } from "@/components/ui";
import { GlassCard } from "@/components/vortexui";
import { CreateRoleModal } from "@/components/CreateRoleModal";
import { EditRoleModal } from "@/components/EditRoleModal";
import { useConfirm } from "@/components/confirm";
import { useToast } from "@/components/toast";
import { useI18n } from "@/i18n/i18n";
import type { Role } from "@/api/types";

function fill(template: string, vars: Record<string, string>) {
  let out = template;
  for (const [k, v] of Object.entries(vars)) out = out.replace(`{${k}}`, v);
  return out;
}

export function RolesTab() {
  const { t } = useI18n();
  const roles = useRoles();
  const delRole = useDeleteRole();
  const confirm = useConfirm();
  const toast = useToast();
  const [roleOpen, setRoleOpen] = useState(false);
  const [editRole, setEditRole] = useState<Role | null>(null);

  const roleList = roles.data?.roles ?? [];

  if (roles.isLoading) {
    return <p className="text-sm text-fg-muted">{t("common.loading")}</p>;
  }

  if (roles.isError) {
    return <p className="text-sm text-danger">{t("reseller.admins.loadFailed")}</p>;
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

  return (
    <div className="space-y-4">
      {roleOpen && <CreateRoleModal open onClose={() => setRoleOpen(false)} />}
      {editRole && <EditRoleModal role={editRole} onClose={() => setEditRole(null)} />}

      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-2">
        <p className="text-sm text-fg-muted">{t("settings.adminsSection.rolesDesc")}</p>
        <Button variant="glass" onClick={() => setRoleOpen(true)} size="sm">
          <Plus size={14} /> {t("reseller.admins.newRole")}
        </Button>
      </div>

      {roleList.length === 0 ? (
        <div className="text-center py-12 text-sm text-fg-muted">
          {t("reseller.admins.noRoles")}
        </div>
      ) : (
        <div className="grid grid-cols-1 gap-3 md:grid-cols-2 lg:grid-cols-3">
          {roleList.map((r) => (
            <GlassCard key={r.id} hover glow className="space-y-3">
              <div className="flex items-start justify-between gap-2">
                <div className="flex items-center gap-2">
                  <span className="h-2 w-2 rounded-full bg-primary shadow-[0_0_6px] shadow-primary/50" />
                  <span className="font-semibold text-fg">{r.name}</span>
                </div>
                <div className="flex shrink-0 gap-1">
                  <Button variant="ghost" size="sm" onClick={() => setEditRole(r)}>{t("common.edit")}</Button>
                  <Button variant="ghost" size="sm" className="text-danger hover:text-danger" onClick={() => removeRole(r)}>{t("common.delete")}</Button>
                </div>
              </div>
              <div className="flex flex-wrap gap-1.5">
                {r.permissions.map((p) => (
                  <span
                    key={p}
                    className="rounded-lg bg-surface-2/60 border border-border/40 px-2 py-0.5 font-mono text-[10px] text-fg-muted"
                  >
                    {p}
                  </span>
                ))}
              </div>
            </GlassCard>
          ))}
        </div>
      )}
    </div>
  );
}
