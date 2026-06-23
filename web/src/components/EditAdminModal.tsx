import { useEffect, useState } from "react";
import { useAdminInbounds, useRoles, useUpdateAdmin } from "@/api/admin-hooks";
import type { Admin } from "@/api/types";
import { AdminInboundPicker } from "@/components/AdminInboundPicker";
import { Button, Input, Select } from "./ui";
import { Modal } from "./Modal";

export function EditAdminModal({ admin, onClose }: { admin: Admin | null; onClose: () => void }) {
  const update = useUpdateAdmin();
  const roles = useRoles();
  const assigned = useAdminInbounds(admin?.id ?? null);
  const [roleId, setRoleId] = useState("");
  const [userQuota, setUserQuota] = useState("");
  const [trafficQuota, setTrafficQuota] = useState("");
  const [inboundIds, setInboundIds] = useState<string[]>([]);
  const [error, setError] = useState("");

  useEffect(() => {
    if (!admin) return;
    setRoleId(admin.role_id ?? "");
    setUserQuota(admin.user_quota ? String(admin.user_quota) : "");
    setTrafficQuota(admin.traffic_quota ? String(Math.round(admin.traffic_quota / (1024 ** 3))) : "");
    setError("");
  }, [admin]);

  useEffect(() => {
    if (assigned.data) setInboundIds(assigned.data.inbound_ids ?? []);
  }, [assigned.data]);

  async function submit(e: React.FormEvent) {
    e.preventDefault();
    if (!admin || admin.sudo) return;
    setError("");
    if (!roleId) {
      setError("Role is required for resellers");
      return;
    }
    try {
      await update.mutateAsync({
        id: admin.id,
        input: {
          sudo: false,
          role_id: roleId,
          user_quota: userQuota ? Number(userQuota) : 0,
          traffic_quota: trafficQuota ? Number(trafficQuota) * 1024 * 1024 * 1024 : 0,
          inbound_ids: inboundIds,
        },
      });
      onClose();
    } catch {
      setError("Could not update admin");
    }
  }

  return (
    <Modal open={!!admin} onClose={onClose} title={`Edit ${admin?.username ?? "admin"}`}>
      {admin && !admin.sudo && (
        <form onSubmit={submit} className="space-y-3">
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
          <AdminInboundPicker selected={inboundIds} onChange={setInboundIds} />
          {error && <p className="text-sm text-destructive">{error}</p>}
          <div className="flex justify-end gap-2 pt-1">
            <Button type="button" variant="ghost" onClick={onClose}>Cancel</Button>
            <Button type="submit" disabled={update.isPending}>
              {update.isPending ? "Saving…" : "Save"}
            </Button>
          </div>
        </form>
      )}
      {admin?.sudo && (
        <p className="text-sm text-muted-foreground">Sudo admins have full access; edit role/quota is not applicable.</p>
      )}
    </Modal>
  );
}
