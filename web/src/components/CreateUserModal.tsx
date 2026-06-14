import { useState } from "react";
import { QRCodeSVG } from "qrcode.react";
import { useAllInbounds, useCreateUser } from "@/api/hooks";
import type { User } from "@/api/types";
import { Button, Input } from "./ui";
import { Modal } from "./Modal";

const GB = 1024 * 1024 * 1024;

export function CreateUserModal({ open, onClose }: { open: boolean; onClose: () => void }) {
  const create = useCreateUser();
  const inbounds = useAllInbounds();

  const [username, setUsername] = useState("");
  const [limitGB, setLimitGB] = useState("");
  const [days, setDays] = useState("");
  const [deviceLimit, setDeviceLimit] = useState("");
  const [selected, setSelected] = useState<string[]>([]);
  const [created, setCreated] = useState<User | null>(null);
  const [error, setError] = useState("");

  function reset() {
    setUsername("");
    setLimitGB("");
    setDays("");
    setDeviceLimit("");
    setSelected([]);
    setCreated(null);
    setError("");
  }

  function close() {
    reset();
    onClose();
  }

  async function submit(e: React.FormEvent) {
    e.preventDefault();
    setError("");
    const expire_at = days ? new Date(Date.now() + Number(days) * 864e5).toISOString() : null;
    try {
      const res = await create.mutateAsync({
        username,
        data_limit: limitGB ? Math.round(Number(limitGB) * GB) : 0,
        expire_at,
        device_limit: deviceLimit ? Number(deviceLimit) : 0,
        inbound_ids: selected,
      });
      setCreated(res.user);
    } catch {
      setError("Could not create user (name taken?)");
    }
  }

  const subURL = created ? `${window.location.origin}/sub/${created.sub_token}` : "";

  return (
    <Modal open={open} onClose={close} title={created ? "User created" : "New user"}>
      {created ? (
        <div className="space-y-4">
          <p className="text-sm text-muted-foreground">
            Share this subscription link with <span className="font-medium text-foreground">{created.username}</span>.
          </p>
          <div className="flex justify-center rounded-lg bg-white p-4">
            <QRCodeSVG value={subURL} size={160} />
          </div>
          <div className="flex gap-2">
            <Input readOnly value={subURL} className="text-xs" onFocus={(e) => e.currentTarget.select()} />
            <Button onClick={() => navigator.clipboard?.writeText(subURL)}>Copy</Button>
          </div>
          <Button variant="ghost" className="w-full" onClick={close}>
            Done
          </Button>
        </div>
      ) : (
        <form onSubmit={submit} className="space-y-3">
          <Input placeholder="Username" value={username} onChange={(e) => setUsername(e.target.value)} required autoFocus />
          <div className="grid grid-cols-3 gap-2">
            <Input placeholder="Limit (GB)" value={limitGB} onChange={(e) => setLimitGB(e.target.value)} inputMode="numeric" />
            <Input placeholder="Expiry (days)" value={days} onChange={(e) => setDays(e.target.value)} inputMode="numeric" />
            <Input placeholder="Devices" value={deviceLimit} onChange={(e) => setDeviceLimit(e.target.value)} inputMode="numeric" />
          </div>

          <div>
            <p className="mb-1 text-xs font-medium text-muted-foreground">Inbounds</p>
            <div className="max-h-32 space-y-1 overflow-auto rounded-md border p-2">
              {inbounds.data?.map((ib) => (
                <label key={ib.id} className="flex items-center gap-2 text-sm">
                  <input
                    type="checkbox"
                    checked={selected.includes(ib.id)}
                    onChange={(e) =>
                      setSelected((s) => (e.target.checked ? [...s, ib.id] : s.filter((x) => x !== ib.id)))
                    }
                  />
                  <span>{ib.tag}</span>
                  <span className="text-xs text-muted-foreground">
                    {ib.protocol} · {ib.nodeName}
                  </span>
                </label>
              ))}
              {inbounds.data?.length === 0 && (
                <p className="text-xs text-muted-foreground">No inbounds yet — create one under Nodes.</p>
              )}
            </div>
          </div>

          {error && <p className="text-sm text-destructive">{error}</p>}
          <div className="flex justify-end gap-2 pt-1">
            <Button type="button" variant="ghost" onClick={close}>
              Cancel
            </Button>
            <Button type="submit" disabled={create.isPending}>
              {create.isPending ? "Creating…" : "Create"}
            </Button>
          </div>
        </form>
      )}
    </Modal>
  );
}
