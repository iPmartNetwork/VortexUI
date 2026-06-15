import { useEffect, useState } from "react";
import { useQuery } from "@tanstack/react-query";
import { useUpdateUser, useAllInbounds } from "@/api/hooks";
import { api } from "@/api/client";
import type { User } from "@/api/types";
import { Button, Input, Select } from "./ui";
import { Modal } from "./Modal";
import { useToast } from "./toast";

const GB = 1024 * 1024 * 1024;

// toDateInput renders an ISO timestamp as yyyy-mm-dd for <input type=date>.
function toDateInput(iso: string | null): string {
  if (!iso) return "";
  return new Date(iso).toISOString().slice(0, 10);
}

export function EditUserModal({ user, onClose }: { user: User | null; onClose: () => void }) {
  const update = useUpdateUser();
  const toast = useToast();
  const [limitGB, setLimitGB] = useState("");
  const [expire, setExpire] = useState("");
  const [deviceLimit, setDeviceLimit] = useState("");
  const [status, setStatus] = useState("active");
  const [reset, setReset] = useState("no_reset");
  const [note, setNote] = useState("");
  const [error, setError] = useState("");
  const [inbounds, setInbounds] = useState<string[]>([]);

  const allInbounds = useAllInbounds();
  // Load the user's current inbound bindings to pre-select them.
  const detail = useQuery({
    queryKey: ["user", user?.id],
    enabled: !!user,
    queryFn: () => api<{ user: User; inbound_ids: string[] }>(`/api/users/${user!.id}`),
  });

  // Re-seed the form whenever a different user is opened.
  useEffect(() => {
    if (!user) return;
    setLimitGB(user.data_limit ? String(Math.round((user.data_limit / GB) * 100) / 100) : "");
    setExpire(toDateInput(user.expire_at));
    setDeviceLimit(user.device_limit ? String(user.device_limit) : "");
    setStatus(user.status === "disabled" ? "disabled" : "active");
    setReset(user.reset_strategy);
    setNote(user.note ?? "");
    setError("");
  }, [user]);

  useEffect(() => {
    if (detail.data) setInbounds(detail.data.inbound_ids ?? []);
  }, [detail.data]);

  if (!user) return null;

  async function submit(e: React.FormEvent) {
    e.preventDefault();
    if (!user) return;
    setError("");
    try {
      await update.mutateAsync({
        id: user.id,
        input: {
          note,
          status,
          reset_strategy: reset,
          data_limit: limitGB ? Math.round(Number(limitGB) * GB) : 0,
          device_limit: deviceLimit ? Number(deviceLimit) : 0,
          expire_at: expire ? new Date(expire).toISOString() : null,
          inbound_ids: inbounds,
        },
      });
      toast.success(`Saved ${user.username}`);
      onClose();
    } catch {
      setError("Update failed");
    }
  }

  return (
    <Modal open={!!user} onClose={onClose} title={`Edit · ${user.username}`}>
      <form onSubmit={submit} className="space-y-3">
        <div className="grid grid-cols-2 gap-2">
          <label className="text-xs text-muted-foreground">
            Data limit (GB)
            <Input className="mt-1" value={limitGB} onChange={(e) => setLimitGB(e.target.value)} inputMode="numeric" placeholder="0 = ∞" />
          </label>
          <label className="text-xs text-muted-foreground">
            Devices
            <Input className="mt-1" value={deviceLimit} onChange={(e) => setDeviceLimit(e.target.value)} inputMode="numeric" placeholder="0 = ∞" />
          </label>
        </div>
        <label className="block text-xs text-muted-foreground">
          Expires
          <Input className="mt-1" type="date" value={expire} onChange={(e) => setExpire(e.target.value)} />
        </label>
        <div className="grid grid-cols-2 gap-2">
          <label className="text-xs text-muted-foreground">
            Status
            <Select className="mt-1" value={status} onChange={(e) => setStatus(e.target.value)}>
              <option value="active">active</option>
              <option value="disabled">disabled</option>
            </Select>
          </label>
          <label className="text-xs text-muted-foreground">
            Reset
            <Select className="mt-1" value={reset} onChange={(e) => setReset(e.target.value)}>
              <option value="no_reset">no reset</option>
              <option value="daily">daily</option>
              <option value="weekly">weekly</option>
              <option value="monthly">monthly</option>
            </Select>
          </label>
        </div>
        <label className="block text-xs text-muted-foreground">
          Note
          <Input className="mt-1" value={note} onChange={(e) => setNote(e.target.value)} placeholder="optional" />
        </label>
        <div>
          <p className="mb-1 text-xs font-medium text-muted-foreground">Inbounds</p>
          <div className="max-h-32 space-y-1 overflow-auto rounded-md border border-border/60 p-2">
            {allInbounds.data?.map((ib) => (
              <label key={ib.id} className="flex items-center gap-2 text-sm">
                <input
                  type="checkbox"
                  checked={inbounds.includes(ib.id)}
                  onChange={(e) => setInbounds((s) => (e.target.checked ? [...s, ib.id] : s.filter((x) => x !== ib.id)))}
                />
                <span>{ib.tag}</span>
                <span className="text-xs text-muted-foreground">{ib.protocol} · {ib.nodeName}</span>
              </label>
            ))}
            {allInbounds.data?.length === 0 && <p className="text-xs text-muted-foreground">No inbounds yet.</p>}
          </div>
        </div>
        {error && <p className="text-sm text-destructive">{error}</p>}
        <div className="flex justify-end gap-2 pt-1">
          <Button type="button" variant="ghost" onClick={onClose}>
            Cancel
          </Button>
          <Button type="submit" disabled={update.isPending}>
            {update.isPending ? "Saving…" : "Save"}
          </Button>
        </div>
      </form>
    </Modal>
  );
}
