import { useState } from "react";
import { useAdmins, useDeleteAdmin, useRoles } from "@/api/admin-hooks";
import { Badge, Button, Card } from "@/components/ui";
import { CreateAdminModal } from "@/components/CreateAdminModal";
import { CreateRoleModal } from "@/components/CreateRoleModal";
import { EditAdminModal } from "@/components/EditAdminModal";
import { useConfirm } from "@/components/confirm";
import { useToast } from "@/components/toast";
import type { Admin } from "@/api/types";

export function Admins() {
  const admins = useAdmins();
  const roles = useRoles();
  const del = useDeleteAdmin();
  const confirm = useConfirm();
  const toast = useToast();
  const [adminOpen, setAdminOpen] = useState(false);
  const [roleOpen, setRoleOpen] = useState(false);
  const [editAdmin, setEditAdmin] = useState<Admin | null>(null);

  const roleName = (id: string | null) => roles.data?.roles.find((r) => r.id === id)?.name ?? "—";

  async function remove(a: Admin) {
    const ok = await confirm({ title: `Delete admin ${a.username}?`, confirmLabel: "Delete", destructive: true });
    if (!ok) return;
    try {
      await del.mutateAsync(a.id);
      toast.success(`Deleted ${a.username}`);
    } catch {
      toast.error("Could not delete (last sudo admin?)");
    }
  }

  return (
    <div className="space-y-8">
      <CreateAdminModal open={adminOpen} onClose={() => setAdminOpen(false)} />
      <CreateRoleModal open={roleOpen} onClose={() => setRoleOpen(false)} />
      <EditAdminModal admin={editAdmin} onClose={() => setEditAdmin(null)} />

      <div>
        <div className="mb-4 flex items-center justify-between">
          <h1 className="text-2xl font-bold tracking-tight">Admins</h1>
          <Button onClick={() => setAdminOpen(true)}>New admin</Button>
        </div>
        <Card className="p-0">
          <table className="w-full text-sm">
            <thead className="border-b text-left text-muted-foreground">
              <tr>
                <th className="px-5 py-3 font-medium">Username</th>
                <th className="px-5 py-3 font-medium">Access</th>
                <th className="px-5 py-3 font-medium">Role</th>
                <th className="px-5 py-3 font-medium">2FA</th>
                <th className="px-5 py-3"></th>
              </tr>
            </thead>
            <tbody>
              {admins.data?.admins.map((a) => (
                <tr key={a.id} className="border-b last:border-0 hover:bg-muted/40">
                  <td className="px-5 py-3 font-medium">{a.username}</td>
                  <td className="px-5 py-3">
                    {a.sudo ? <Badge color="active">sudo</Badge> : <Badge>reseller</Badge>}
                  </td>
                  <td className="px-5 py-3 text-muted-foreground">{a.sudo ? "—" : roleName(a.role_id)}</td>
                  <td className="px-5 py-3 text-muted-foreground">{a.totp_enabled ? "on" : "off"}</td>
                  <td className="px-5 py-3 text-right">
                    {!a.sudo && (
                      <Button variant="ghost" onClick={() => setEditAdmin(a)}>Edit</Button>
                    )}
                    <Button variant="ghost" className="text-destructive" onClick={() => remove(a)}>
                      Delete
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
          <h2 className="text-lg font-semibold">Roles</h2>
          <Button variant="ghost" onClick={() => setRoleOpen(true)}>New role</Button>
        </div>
        <div className="grid grid-cols-1 gap-3 md:grid-cols-2 lg:grid-cols-3">
          {roles.data?.roles.map((r) => (
            <Card key={r.id} className="space-y-2">
              <div className="font-medium">{r.name}</div>
              <div className="flex flex-wrap gap-1">
                {r.permissions.map((p) => (
                  <span key={p} className="rounded bg-muted px-1.5 py-0.5 font-mono text-[11px] text-muted-foreground">
                    {p}
                  </span>
                ))}
              </div>
            </Card>
          ))}
          {roles.data?.roles.length === 0 && (
            <p className="text-sm text-muted-foreground">No roles yet — create a reseller role first.</p>
          )}
        </div>
      </div>
    </div>
  );
}
