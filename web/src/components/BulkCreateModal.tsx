import { useState } from "react";
import { useAllInbounds, useBulkCreateUsers } from "@/api/hooks";
import { Button, Input } from "./ui";
import { Modal } from "./Modal";

const GB = 1024 * 1024 * 1024;

type Result = { created_count: number; failures: { username: string; error: string }[] };

// BulkCreateModal creates many users at once from a shared plan/template,
// 3x-ui style: a username prefix plus a sequential counter.
export function BulkCreateModal({ open, onClose }: { open: boolean; onClose: () => void }) {
  const bulk = useBulkCreateUsers();
  const inbounds = useAllInbounds();

  const [prefix, setPrefix] = useState("user-");
  const [count, setCount] = useState("10");
  const [start, setStart] = useState("1");
  const [pad, setPad] = useState("3");
  const [limitGB, setLimitGB] = useState("");
  const [days, setDays] = useState("");
  const [deviceLimit, setDeviceLimit] = useState("");
  const [selected, setSelected] = useState<string[]>([]);
  const [result, setResult] = useState<Result | null>(null);
  const [error, setError] = useState("");

  function reset() {
    setPrefix("user-"); setCount("10"); setStart("1"); setPad("3");
    setLimitGB(""); setDays(""); setDeviceLimit(""); setSelected([]);
    setResult(null); setError("");
  }
  function close() { reset(); onClose(); }

  const n = Number(count) || 0;
  const s = Number(start) || 1;
  const p = Number(pad) || 0;
  const fmtName = (num: number) => `${prefix}${p > 0 ? String(num).padStart(p, "0") : num}`;
  const preview = n > 0 ? [fmtName(s), n > 1 ? fmtName(s + 1) : null, n > 2 ? "…" : null, n > 2 ? fmtName(s + n - 1) : null].filter(Boolean).join(", ") : "";

  async function submit(e: React.FormEvent) {
    e.preventDefault();
    setError("");
    const expire_at = days ? new Date(Date.now() + Number(days) * 864e5).toISOString() : null;
    try {
      const res = await bulk.mutateAsync({
        prefix,
        count: n,
        start: s,
        pad: p,
        data_limit: limitGB ? Math.round(Number(limitGB) * GB) : 0,
        expire_at,
        device_limit: deviceLimit ? Number(deviceLimit) : 0,
        inbound_ids: selected,
      });
      setResult({ created_count: res.created_count, failures: res.failures });
    } catch {
      setError("Bulk create failed");
    }
  }

  return (
    <Modal open={open} onClose={close} title={result ? "Bulk create finished" : "Add bulk users"}>
      {result ? (
        <div className="space-y-4">
          <p className="text-sm text-fg">
            Created <span className="font-semibold text-success">{result.created_count}</span> users.
          </p>
          {result.failures.length > 0 && (
            <div>
              <p className="mb-1 text-xs font-medium text-danger">{result.failures.length} failed</p>
              <div className="max-h-32 space-y-1 overflow-auto rounded-md border border-border/60 p-2 text-xs">
                {result.failures.map((f) => (
                  <div key={f.username} className="flex justify-between gap-2">
                    <span className="font-mono">{f.username}</span>
                    <span className="text-fg-muted">{f.error}</span>
                  </div>
                ))}
              </div>
            </div>
          )}
          <Button variant="ghost" className="w-full" onClick={close}>Done</Button>
        </div>
      ) : (
        <form onSubmit={submit} className="space-y-3">
          <div className="grid grid-cols-2 gap-2">
            <div>
              <label className="mb-1 block text-xs font-medium text-fg-muted">Username prefix</label>
              <Input value={prefix} onChange={(e) => setPrefix(e.target.value)} required autoFocus />
            </div>
            <div>
              <label className="mb-1 block text-xs font-medium text-fg-muted">Count (1–500)</label>
              <Input value={count} onChange={(e) => setCount(e.target.value)} inputMode="numeric" required />
            </div>
            <div>
              <label className="mb-1 block text-xs font-medium text-fg-muted">Start #</label>
              <Input value={start} onChange={(e) => setStart(e.target.value)} inputMode="numeric" />
            </div>
            <div>
              <label className="mb-1 block text-xs font-medium text-fg-muted">Zero-pad width</label>
              <Input value={pad} onChange={(e) => setPad(e.target.value)} inputMode="numeric" />
            </div>
          </div>

          {preview && (
            <p className="rounded-md bg-surface-2/40 px-3 py-2 text-xs text-fg-muted">
              Preview: <span className="font-mono text-fg">{preview}</span>
            </p>
          )}

          <div>
            <p className="mb-1 text-xs font-medium text-fg-muted">Shared plan</p>
            <div className="grid grid-cols-3 gap-2">
              <Input placeholder="Limit (GB)" value={limitGB} onChange={(e) => setLimitGB(e.target.value)} inputMode="numeric" />
              <Input placeholder="Expiry (days)" value={days} onChange={(e) => setDays(e.target.value)} inputMode="numeric" />
              <Input placeholder="Devices" value={deviceLimit} onChange={(e) => setDeviceLimit(e.target.value)} inputMode="numeric" />
            </div>
          </div>

          <div>
            <p className="mb-1 text-xs font-medium text-fg-muted">Inbounds</p>
            <div className="max-h-32 space-y-1 overflow-auto rounded-md border border-border/60 p-2">
              {inbounds.data?.map((ib) => (
                <label key={ib.id} className="flex items-center gap-2 text-sm">
                  <input
                    type="checkbox"
                    checked={selected.includes(ib.id)}
                    onChange={(e) => setSelected((x) => (e.target.checked ? [...x, ib.id] : x.filter((v) => v !== ib.id)))}
                  />
                  <span>{ib.tag}</span>
                  <span className="text-xs text-fg-muted">{ib.protocol} · {ib.nodeName}</span>
                </label>
              ))}
              {inbounds.data?.length === 0 && <p className="text-xs text-fg-muted">No inbounds yet — create one under Inbounds.</p>}
            </div>
          </div>

          {error && <p className="text-sm text-danger">{error}</p>}
          <div className="flex justify-end gap-2 pt-1">
            <Button type="button" variant="ghost" onClick={close}>Cancel</Button>
            <Button type="submit" disabled={bulk.isPending}>{bulk.isPending ? "Creating…" : `Create ${n || ""}`}</Button>
          </div>
        </form>
      )}
    </Modal>
  );
}
