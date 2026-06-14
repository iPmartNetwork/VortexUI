import { useState } from "react";
import { useCreateRole } from "@/api/admin-hooks";
import { ALL_PERMISSIONS } from "@/api/types";
import { Button, Input } from "./ui";
import { Modal } from "./Modal";

export function CreateRoleModal({ open, onClose }: { open: boolean; onClose: () => void }) {
  const create = useCreateRole();
  const [name, setName] = useState("");
  const [perms, setPerms] = useState<string[]>([]);

  function close() {
    setName("");
    setPerms([]);
    onClose();
  }

  function toggle(p: string, on: boolean) {
    setPerms((s) => (on ? [...s, p] : s.filter((x) => x !== p)));
  }

  async function submit(e: React.FormEvent) {
    e.preventDefault();
    await create.mutateAsync({ name, permissions: perms });
    close();
  }

  return (
    <Modal open={open} onClose={close} title="New role">
      <form onSubmit={submit} className="space-y-3">
        <Input placeholder="Role name (e.g. reseller)" value={name} onChange={(e) => setName(e.target.value)} required autoFocus />
        <div>
          <p className="mb-1 text-xs font-medium text-muted-foreground">Permissions</p>
          <div className="grid grid-cols-2 gap-1 rounded-md border p-2">
            {ALL_PERMISSIONS.map((p) => (
              <label key={p} className="flex items-center gap-2 text-sm">
                <input type="checkbox" checked={perms.includes(p)} onChange={(e) => toggle(p, e.target.checked)} />
                <span className="font-mono text-xs">{p}</span>
              </label>
            ))}
          </div>
        </div>
        <div className="flex justify-end gap-2 pt-1">
          <Button type="button" variant="ghost" onClick={close}>Cancel</Button>
          <Button type="submit" disabled={create.isPending}>
            {create.isPending ? "Creating…" : "Create role"}
          </Button>
        </div>
      </form>
    </Modal>
  );
}
