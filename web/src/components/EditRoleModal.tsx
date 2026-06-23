import { useEffect, useState } from "react";
import { useUpdateRole } from "@/api/admin-hooks";
import { ALL_PERMISSIONS, type Role } from "@/api/types";
import { RESELLER_PERMISSIONS } from "@/auth/permissions";
import { ensureArray } from "@/lib/utils";
import { Button, Input } from "./ui";
import { Modal } from "./Modal";

export function EditRoleModal({ role, onClose }: { role: Role | null; onClose: () => void }) {
  const update = useUpdateRole();
  const [name, setName] = useState("");
  const [perms, setPerms] = useState<string[]>([]);
  const [error, setError] = useState("");

  useEffect(() => {
    if (!role) return;
    setName(role.name);
    setPerms([...ensureArray(role.permissions)]);
    setError("");
  }, [role]);

  function toggle(p: string, on: boolean) {
    setPerms((s) => (on ? [...s, p] : s.filter((x) => x !== p)));
  }

  async function submit(e: React.FormEvent) {
    e.preventDefault();
    if (!role) return;
    setError("");
    if (perms.length === 0) {
      setError("Select at least one permission");
      return;
    }
    try {
      await update.mutateAsync({ id: role.id, input: { name, permissions: perms } });
      onClose();
    } catch {
      setError("Could not update role");
    }
  }

  return (
    <Modal open={!!role} onClose={onClose} title={`Edit role · ${role?.name ?? ""}`}>
      {role && (
        <form onSubmit={submit} className="space-y-3">
          <Input placeholder="Role name" value={name} onChange={(e) => setName(e.target.value)} required autoFocus />
          <div>
            <div className="mb-1 flex items-center justify-between">
              <p className="text-xs font-medium text-muted-foreground">Permissions</p>
              <Button type="button" variant="ghost" className="h-7 px-2 text-xs" onClick={() => setPerms([...RESELLER_PERMISSIONS])}>
                Reseller preset
              </Button>
            </div>
            <div className="grid grid-cols-2 gap-1 rounded-md border p-2">
              {ALL_PERMISSIONS.map((p) => (
                <label key={p} className="flex items-center gap-2 text-sm">
                  <input type="checkbox" checked={perms.includes(p)} onChange={(e) => toggle(p, e.target.checked)} />
                  <span className="font-mono text-xs">{p}</span>
                </label>
              ))}
            </div>
          </div>
          {error && <p className="text-sm text-destructive">{error}</p>}
          <div className="flex justify-end gap-2 pt-1">
            <Button type="button" variant="ghost" onClick={onClose}>Cancel</Button>
            <Button type="submit" disabled={update.isPending}>
              {update.isPending ? "Saving…" : "Save"}
            </Button>
          </div>
        </form>
      )}
    </Modal>
  );
}
