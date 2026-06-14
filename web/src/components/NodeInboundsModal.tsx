import { useState } from "react";
import { useCreateInbound, useDeleteInbound, useNodeInbounds, useUpdateInbound, type Inbound } from "@/api/hooks";
import { useReality } from "@/api/policy-hooks";
import type { Node } from "@/api/types";
import { Badge, Button, Input, Select } from "./ui";
import { Modal } from "./Modal";
import { CopyField } from "./CopyField";
import { useToast } from "./toast";

const PROTOCOLS = ["vless", "vmess", "trojan", "shadowsocks"];
const NETWORKS = ["tcp", "ws", "grpc"];
const SECURITIES = ["none", "tls", "reality"];

const blank = { editId: "", tag: "", protocol: "vless", port: "", network: "tcp", security: "tls", sni: "" };

export function NodeInboundsModal({ node, onClose }: { node: Node | null; onClose: () => void }) {
  const list = useNodeInbounds(node?.id ?? null);
  const create = useCreateInbound();
  const update = useUpdateInbound();
  const del = useDeleteInbound();
  const toast = useToast();
  const [f, setF] = useState({ ...blank });

  if (!node) return null;
  const editing = f.editId !== "";
  const set = (k: keyof typeof f) => (e: React.ChangeEvent<HTMLInputElement | HTMLSelectElement>) =>
    setF((s) => ({ ...s, [k]: e.target.value }));

  function startEdit(ib: Inbound) {
    setF({ editId: ib.id, tag: ib.tag, protocol: ib.protocol, port: String(ib.port), network: ib.network, security: ib.security, sni: "" });
  }

  async function toggleEnable(ib: Inbound) {
    try {
      await update.mutateAsync({ id: ib.id, input: { port: ib.port, network: ib.network, security: ib.security, enabled: !ib.enabled } });
      toast.success(`${ib.tag} ${ib.enabled ? "disabled" : "enabled"}`);
    } catch {
      toast.error("Toggle failed");
    }
  }

  async function submit(e: React.FormEvent) {
    e.preventDefault();
    if (!node) return;
    const sni = f.sni ? f.sni.split(",").map((s) => s.trim()) : [];
    try {
      if (editing) {
        await update.mutateAsync({ id: f.editId, input: { port: Number(f.port), network: f.network, security: f.security, sni, enabled: true } });
        toast.success(`Updated ${f.tag}`);
      } else {
        await create.mutateAsync({ node_id: node.id, tag: f.tag, protocol: f.protocol, port: Number(f.port), network: f.network, security: f.security, sni, enabled: true });
        toast.success(`Added ${f.tag}`);
      }
      setF({ ...blank });
    } catch {
      toast.error("Save failed");
    }
  }

  return (
    <Modal open={!!node} onClose={onClose} title={`Inbounds · ${node.name}`} className="max-w-lg">
      <div className="space-y-2">
        {list.data?.inbounds?.map((ib) => (
          <div key={ib.id} className="flex items-center justify-between rounded-md border px-3 py-2 text-sm">
            <div className="flex items-center gap-2">
              <span className="font-medium">{ib.tag}</span>
              <Badge>{ib.protocol}</Badge>
              <span className="text-xs text-muted-foreground">:{ib.port} · {ib.network}/{ib.security}</span>
              <span className={`h-2 w-2 rounded-full ${ib.enabled ? "bg-success" : "bg-fg-subtle"}`} title={ib.enabled ? "Enabled" : "Disabled"} />
            </div>
            <div className="flex gap-1">
              <Button variant="ghost" size="sm" onClick={() => toggleEnable(ib)} title={ib.enabled ? "Disable" : "Enable"}>
                {ib.enabled ? "🟢" : "⏸"}
              </Button>
              <Button variant="ghost" size="sm" onClick={() => startEdit(ib)}>Edit</Button>
              <Button variant="ghost" size="sm" className="text-destructive" onClick={() => del.mutate(ib.id)}>Remove</Button>
            </div>
          </div>
        ))}
        {list.data?.inbounds?.length === 0 && <p className="py-2 text-sm text-muted-foreground">No inbounds on this node yet.</p>}
      </div>

      <form onSubmit={submit} className="mt-4 space-y-3 border-t pt-4">
        <div className="flex items-center justify-between">
          <p className="text-xs font-medium text-muted-foreground">{editing ? `Edit ${f.tag}` : "Add inbound"}</p>
          {editing && (
            <button type="button" className="text-xs text-muted-foreground underline" onClick={() => setF({ ...blank })}>
              cancel edit
            </button>
          )}
        </div>
        <div className="grid grid-cols-2 gap-2">
          <Input placeholder="Tag" value={f.tag} onChange={set("tag")} required disabled={editing} />
          <Input placeholder="Port" value={f.port} onChange={set("port")} inputMode="numeric" required />
        </div>
        <div className="grid grid-cols-3 gap-2">
          <Select value={f.protocol} onChange={set("protocol")} disabled={editing}>
            {PROTOCOLS.map((p) => <option key={p} value={p}>{p}</option>)}
          </Select>
          <Select value={f.network} onChange={set("network")}>
            {NETWORKS.map((n) => <option key={n} value={n}>{n}</option>)}
          </Select>
          <Select value={f.security} onChange={set("security")}>
            {SECURITIES.map((s) => <option key={s} value={s}>{s}</option>)}
          </Select>
        </div>
        <Input placeholder="SNI (comma-separated, optional)" value={f.sni} onChange={set("sni")} />
        {f.security === "reality" && <RealityKeygenSection />}
        <div className="flex justify-end">
          <Button type="submit" disabled={create.isPending || update.isPending}>
            {editing ? "Save changes" : "Add inbound"}
          </Button>
        </div>
      </form>
    </Modal>
  );
}

function RealityKeygenSection() {
  const reality = useReality();
  const [keys, setKeys] = useState<{ private_key: string; public_key: string; short_id: string } | null>(null);

  async function generate() {
    const r = await reality.mutateAsync();
    setKeys(r);
  }

  return (
    <div className="space-y-2 rounded-lg border border-white/[0.06] bg-white/[0.02] p-3">
      <div className="flex items-center justify-between">
        <span className="text-xs font-medium text-fg-muted">REALITY keys</span>
        <Button type="button" variant="ghost" size="sm" onClick={generate} disabled={reality.isPending}>
          Generate
        </Button>
      </div>
      {keys && (
        <div className="space-y-1.5">
          <CopyField value={keys.private_key} />
          <CopyField value={keys.public_key} />
          <CopyField value={keys.short_id} />
          <p className="text-[10px] text-fg-subtle">
            Store private_key + short_id in the inbound's raw.reality; give public_key to clients.
          </p>
        </div>
      )}
    </div>
  );
}
