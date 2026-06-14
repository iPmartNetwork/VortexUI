import { useState } from "react";
import { useDeleteUser, useUsers } from "@/api/hooks";
import type { User } from "@/api/types";
import { Badge, Button, Card, Input } from "@/components/ui";
import { CreateUserModal } from "@/components/CreateUserModal";
import { EditUserModal } from "@/components/EditUserModal";
import { UserUsageModal } from "@/components/UserUsageModal";
import { useConfirm } from "@/components/confirm";
import { useToast } from "@/components/toast";
import { formatBytes } from "@/lib/utils";

export function Users() {
  const [search, setSearch] = useState("");
  const [modalOpen, setModalOpen] = useState(false);
  const [editing, setEditing] = useState<User | null>(null);
  const [viewing, setViewing] = useState<User | null>(null);
  const { data, isLoading, error } = useUsers({ search, limit: 50 });
  const del = useDeleteUser();
  const confirm = useConfirm();
  const toast = useToast();

  async function remove(u: User) {
    const ok = await confirm({
      title: `Delete ${u.username}?`,
      message: "This removes the user and revokes their access on all nodes.",
      confirmLabel: "Delete",
      destructive: true,
    });
    if (!ok) return;
    try {
      await del.mutateAsync(u.id);
      toast.success(`Deleted ${u.username}`);
    } catch {
      toast.error("Delete failed");
    }
  }

  return (
    <div className="space-y-6">
      <CreateUserModal open={modalOpen} onClose={() => setModalOpen(false)} />
      <EditUserModal user={editing} onClose={() => setEditing(null)} />
      <UserUsageModal user={viewing} onClose={() => setViewing(null)} />
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold tracking-tight">Users</h1>
          <p className="text-sm text-muted-foreground">{data?.total ?? 0} total</p>
        </div>
        <div className="flex gap-2">
          <Input
            className="max-w-xs"
            placeholder="Search…"
            value={search}
            onChange={(e) => setSearch(e.target.value)}
          />
          <Button onClick={() => setModalOpen(true)}>New user</Button>
        </div>
      </div>

      <Card className="p-0">
        {isLoading && <div className="p-6 text-sm text-muted-foreground">Loading…</div>}
        {error && <div className="p-6 text-sm text-destructive">Failed to load users</div>}
        {data && (
          <table className="w-full text-sm">
            <thead className="border-b text-left text-muted-foreground">
              <tr>
                <th className="px-5 py-3 font-medium">Username</th>
                <th className="px-5 py-3 font-medium">Status</th>
                <th className="px-5 py-3 font-medium">Usage</th>
                <th className="px-5 py-3 font-medium">Expires</th>
                <th className="px-5 py-3"></th>
              </tr>
            </thead>
            <tbody>
              {data.users.map((u) => (
                <tr key={u.id} className="border-b last:border-0 hover:bg-muted/40">
                  <td className="px-5 py-3 font-medium">{u.username}</td>
                  <td className="px-5 py-3">
                    <Badge color={u.status}>{u.status}</Badge>
                  </td>
                  <td className="px-5 py-3 text-muted-foreground">
                    {formatBytes(u.used_traffic)} / {formatBytes(u.data_limit)}
                  </td>
                  <td className="px-5 py-3 text-muted-foreground">
                    {u.expire_at ? new Date(u.expire_at).toLocaleDateString() : "Never"}
                  </td>
                  <td className="px-5 py-3 text-right">
                    <Button variant="ghost" onClick={() => setViewing(u)}>
                      Usage
                    </Button>
                    <Button variant="ghost" onClick={() => setEditing(u)}>
                      Edit
                    </Button>
                    <Button variant="ghost" className="text-destructive" onClick={() => remove(u)}>
                      Delete
                    </Button>
                  </td>
                </tr>
              ))}
              {data.users.length === 0 && (
                <tr>
                  <td colSpan={5} className="px-5 py-8 text-center text-muted-foreground">
                    No users found
                  </td>
                </tr>
              )}
            </tbody>
          </table>
        )}
      </Card>
    </div>
  );
}
