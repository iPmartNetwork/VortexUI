import { useState } from "react";
import { useCreateAdmin, useRoles } from "@/api/admin-hooks";
import { AdminInboundPicker } from "@/components/AdminInboundPicker";
import { AdminNodePicker } from "@/components/AdminNodePicker";
import { AdminPlanPicker } from "@/components/AdminPlanPicker";
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
  const [trafficMode, setTrafficMode] = useState("allocated");
  const [inboundIds, setInboundIds] = useState<string[]>([]);
  const [nodeIds, setNodeIds] = useState<string[]>([]);
  const [planIds, setPlanIds] = useState<string[]>([]);
  const [error, setError] = useState("");

  function close() {
    setUsername("");
    setPassword("");
    setSudo(false);
    setRoleId("");
    setUserQuota("");
    setTrafficQuota("");
    setTrafficMode("allocated");
    setInboundIds([]);
    setNodeIds([]);
    setPlanIds([]);
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
        traffic_quota_mode: sudo ? undefined : trafficMode,
        inbound_ids: sudo ? undefined : inboundIds,
        node_ids: sudo ? undefined : nodeIds,
        plan_ids: sudo ? undefined : planIds,
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
              <Input placeholder="User quota (0=∞)" value={userQuota} onChange={(e) => setUserQuota(e.target.value)} inputMode="numeric" />
              <Input placeholder="Traffic quota (GB, 0=∞)" value={trafficQuota} onChange={(e) => setTrafficQuota(e.target.value)} inputMode="numeric" />
            </div>
            <label className="block text-xs text-muted-foreground">
              Traffic quota mode
              <Select className="mt-1" value={trafficMode} onChange={(e) => setTrafficMode(e.target.value)}>
                <option value="allocated">Allocated — sum of user data limits</option>
                <option value="consumed">Consumed — actual traffic used</option>
              </Select>
            </label>
            <AdminInboundPicker selected={inboundIds} onChange={setInboundIds} />
            <AdminNodePicker selected={nodeIds} onChange={setNodeIds} />
            <AdminPlanPicker selected={planIds} onChange={setPlanIds} />
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
