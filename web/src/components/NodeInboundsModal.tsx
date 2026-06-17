import { useState } from "react";
import { useCreateInbound, useDeleteInbound, useNodeInbounds, useUpdateInbound, type Inbound } from "@/api/hooks";
import { useReality } from "@/api/policy-hooks";
import type { Node } from "@/api/types";
import { Badge, Button, Input, Select } from "./ui";
import { Modal } from "./Modal";
import { CopyField } from "./CopyField";
import { JsonCodeEditor } from "./JsonCodeEditor";
import { useToast } from "./toast";

// xray supports vless/vmess/trojan/shadowsocks; hysteria2/tuic require a
// sing-box node (switch the local node's core, or add a sing-box node).
const PROTOCOLS = ["vless", "vmess", "trojan", "shadowsocks", "hysteria2", "tuic"];
const NETWORKS = ["tcp", "ws", "grpc", "httpupgrade", "h2", "quic", "udp"];
const SECURITIES = ["none", "tls", "reality"];

// Protocols that run over UDP (QUIC-based) — transport/security are fixed.
const UDP_PROTOCOLS = ["hysteria2", "tuic"];

// randomPort picks a high port (10000–60000) so new inbounds default to a free,
// non-conflicting port. The admin can still type any port.
const randomPort = () => String(10000 + Math.floor(Math.random() * 50000));

const newBlank = () => ({ editId: "", tag: "", protocol: "vless", port: randomPort(), network: "tcp", security: "tls", sni: "", speedLimit: "" });
const blank = newBlank();

const DEFAULT_INBOUND_TEMPLATE = {
  tag: "inbound-443",
  protocol: "vless",
  port: 443,
  settings: { clients: [], decryption: "none" },
  streamSettings: {
    network: "tcp",
    tcpSettings: { header: { type: "none" } },
    security: "none",
  },
};

export function NodeInboundsModal({ node, onClose }: { node: Node | null; onClose: () => void }) {
  const list = useNodeInbounds(node?.id ?? null);
  const create = useCreateInbound();
  const update = useUpdateInbound();
  const del = useDeleteInbound();
  const toast = useToast();
  const [f, setF] = useState({ ...blank });
  const [tab, setTab] = useState<"basics" | "json">("basics");
  const [jsonText, setJsonText] = useState("");
  const [jsonErr, setJsonErr] = useState("");

  if (!node) return null;
  const editing = f.editId !== "";
  const set = (k: keyof typeof f) => (e: React.ChangeEvent<HTMLInputElement | HTMLSelectElement>) =>
    setF((s) => {
      const val = e.target.value;
      // When protocol changes to UDP-based (hysteria2/tuic), lock transport to udp + security to tls
      if (k === "protocol" && UDP_PROTOCOLS.includes(val)) {
        return { ...s, [k]: val, network: "udp", security: "tls" };
      }
      return { ...s, [k]: val };
    });

  function startEdit(ib: Inbound) {
    setF({ editId: ib.id, tag: ib.tag, protocol: ib.protocol, port: String(ib.port), network: ib.network, security: ib.security, sni: "", speedLimit: "" });
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
      setF(newBlank());
    } catch {
      toast.error("Save failed");
    }
  }

  function gotoJson() {
    setTab("json");
    if (!jsonText) setJsonText(JSON.stringify(DEFAULT_INBOUND_TEMPLATE, null, 2));
  }

  async function submitJSON() {
    setJsonErr("");
    if (!node) return;
    let parsed: Record<string, unknown>;
    try {
      parsed = JSON.parse(jsonText);
    } catch {
      setJsonErr("Invalid JSON");
      return;
    }
    const tag = String(parsed.tag ?? "");
    const protocol = String(parsed.protocol ?? "");
    const port = Number(parsed.port ?? 0);
    if (!tag || !protocol || !port) {
      setJsonErr("JSON must include tag, protocol and port");
      return;
    }
    try {
      await create.mutateAsync({ node_id: node.id, tag, protocol, port, raw: parsed, enabled: true });
      toast.success(`Added ${tag}`);
      setJsonText("");
      setTab("basics");
    } catch {
      setJsonErr("Save failed (tag/port taken?)");
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

      <div className="mt-4 flex gap-5 border-t border-border/60 pt-3 text-sm">
        {(["basics", "json"] as const).map((tk) => (
          <button
            key={tk}
            type="button"
            onClick={() => (tk === "json" ? gotoJson() : setTab("basics"))}
            className={`-mb-px border-b-2 pb-2 font-medium transition ${tab === tk ? "border-primary text-primary" : "border-transparent text-fg-muted hover:text-fg"}`}
          >
            {tk === "basics" ? "Basics" : "JSON"}
          </button>
        ))}
      </div>

      {tab === "json" ? (
        <div className="mt-3 space-y-3">
          <p className="text-xs text-fg-muted">Paste a full Xray/sing-box inbound object. tag, protocol and port are read from it; the whole object is stored as a raw override.</p>
          <JsonCodeEditor value={jsonText} onChange={setJsonText} rows={14} />
          {jsonErr && <p className="text-sm text-danger">{jsonErr}</p>}
          <div className="flex justify-end">
            <Button type="button" onClick={submitJSON} disabled={create.isPending}>Add inbound</Button>
          </div>
        </div>
      ) : (
      <form onSubmit={submit} className="mt-3 space-y-3">
        <div className="flex items-center justify-between">
          <p className="text-xs font-medium text-muted-foreground">{editing ? `Edit ${f.tag}` : "Add inbound"}</p>
          {editing && (
            <button type="button" className="text-xs text-muted-foreground underline" onClick={() => setF(newBlank())}>
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
          <Select value={f.network} onChange={set("network")} disabled={UDP_PROTOCOLS.includes(f.protocol)}>
            {NETWORKS.map((n) => <option key={n} value={n}>{n}</option>)}
          </Select>
          <Select value={f.security} onChange={set("security")} disabled={UDP_PROTOCOLS.includes(f.protocol)}>
            {SECURITIES.map((s) => <option key={s} value={s}>{s}</option>)}
          </Select>
        </div>
        <Input placeholder="SNI (comma-separated, optional)" value={f.sni} onChange={set("sni")} />
        {f.security === "reality" && <RealityKeygenSection />}
        <Input placeholder="Speed limit (bytes/sec, 0 = unlimited)" value={f.speedLimit ?? ""} onChange={(e) => setF(s => ({...s, speedLimit: e.target.value}))} inputMode="numeric" />
        <div className="flex justify-end">
          <Button type="submit" disabled={create.isPending || update.isPending}>
            {editing ? "Save changes" : "Add inbound"}
          </Button>
        </div>
      </form>
      )}
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
