import { useState } from "react";
import { useCreateAdmin, useRoles } from "@/api/admin-hooks";
import { AdminInboundPicker } from "@/components/AdminInboundPicker";
import { Button, Input, Select } from "./ui";
import { Modal } from "./Modal";

export function CreateAdminModal({ open, onClose }: { open: boolean; onClose: () => void }) {
  const create = useCreateAdmin();
  const roles = useRoles();
  const [username, setUsername] = useState("");
  const [password, setPassword] = useState("");
  const [sudo, setSudo] = useState(false);
  const [roleId, setRoleId] = useState("");
  const [userQuota, setUserQuota] = useState("");
  const [trafficQuota, setTrafficQuota] = useState("");
  const [inboundIds, setInboundIds] = useState<string[]>([]);
  const [error, setError] = useState("");

  function close() {
    setUsername("");
    setPassword("");
    setSudo(false);
    setRoleId("");
    setUserQuota("");
    setTrafficQuota("");
    setInboundIds([]);
    setError("");
    onClose();
  }

  async function submit(e: React.FormEvent) {
    e.preventDefault();
    setError("");
    if (!sudo && !roleId) {
      setError("Select a role for resellers");
      return;
    }
    try {
      await create.mutateAsync({
        username, password, sudo,
        role_id: sudo ? null : roleId,
        user_quota: userQuota ? Number(userQuota) : 0,
        traffic_quota: trafficQuota ? Number(trafficQuota) * 1024 * 1024 * 1024 : 0,
        inbound_ids: sudo ? undefined : inboundIds,
      });
      close();
    } catch {
      setError("Could not create admin (username taken or missing role?)");
    }
  }

  return (
    <Modal open={open} onClose={close} title="New admin">
      <form onSubmit={submit} className="space-y-3">
        <Input placeholder="Username" value={username} onChange={(e) => setUsername(e.target.value)} required autoFocus />
        <Input type="password" placeholder="Password" value={password} onChange={(e) => setPassword(e.target.value)} required />
        <label className="flex items-center gap-2 text-sm">
          <input type="checkbox" checked={sudo} onChange={(e) => setSudo(e.target.checked)} />
          Full access (sudo)
        </label>
        {!sudo && (
          <>
            <label className="block text-xs text-muted-foreground">
              Role
              <Select className="mt-1" value={roleId} onChange={(e) => setRoleId(e.target.value)} required>
                <option value="">Select role…</option>
                {roles.data?.roles.map((r) => (
                  <option key={r.id} value={r.id}>{r.name}</option>
                ))}
              </Select>
            </label>
            <div className="grid grid-cols-2 gap-2">
              <Input placeholder="User quota (0=unlimited)" value={userQuota} onChange={(e) => setUserQuota(e.target.value)} inputMode="numeric" />
              <Input placeholder="Traffic quota (GB, 0=unlimited)" value={trafficQuota} onChange={(e) => setTrafficQuota(e.target.value)} inputMode="numeric" />
            </div>
            <p className="text-[10px] text-fg-subtle">Reseller limits: max users this admin can create, and max traffic per user (GB).</p>
            <AdminInboundPicker selected={inboundIds} onChange={setInboundIds} />
          </>
        )}
        {error && <p className="text-sm text-destructive">{error}</p>}
        <div className="flex justify-end gap-2 pt-1">
          <Button type="button" variant="ghost" onClick={close}>Cancel</Button>
          <Button type="submit" disabled={create.isPending}>
            {create.isPending ? "Creating…" : "Create"}
          </Button>
        </div>
      </form>
    </Modal>
  );
}
