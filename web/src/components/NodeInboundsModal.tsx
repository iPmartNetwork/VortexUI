import { useState } from "react";
import { useCreateInbound, useDeleteInbound, useNodeInbounds } from "@/api/hooks";
import type { Node } from "@/api/types";
import { Badge, Button, Input, Select } from "./ui";
import { Modal } from "./Modal";

const PROTOCOLS = ["vless", "vmess", "trojan", "shadowsocks"];
const NETWORKS = ["tcp", "ws", "grpc"];
const SECURITIES = ["none", "tls", "reality"];

export function NodeInboundsModal({ node, onClose }: { node: Node | null; onClose: () => void }) {
  const list = useNodeInbounds(node?.id ?? null);
  const create = useCreateInbound();
  const del = useDeleteInbound();

  const [tag, setTag] = useState("");
  const [protocol, setProtocol] = useState("vless");
  const [port, setPort] = useState("");
  const [network, setNetwork] = useState("tcp");
  const [security, setSecurity] = useState("tls");
  const [sni, setSni] = useState("");

  if (!node) return null;

  async function add(e: React.FormEvent) {
    e.preventDefault();
    if (!node) return;
    await create.mutateAsync({
      node_id: node.id,
      tag,
      protocol,
      port: Number(port),
      network,
      security,
      sni: sni ? sni.split(",").map((s) => s.trim()) : [],
      enabled: true,
    });
    setTag("");
    setPort("");
    setSni("");
  }

  return (
    <Modal open={!!node} onClose={onClose} title={`Inbounds · ${node.name}`} className="max-w-lg">
      <div className="space-y-2">
        {list.data?.inbounds?.map((ib) => (
          <div key={ib.id} className="flex items-center justify-between rounded-md border px-3 py-2 text-sm">
            <div className="flex items-center gap-2">
              <span className="font-medium">{ib.tag}</span>
              <Badge>{ib.protocol}</Badge>
              <span className="text-xs text-muted-foreground">
                :{ib.port} · {ib.network}/{ib.security}
              </span>
            </div>
            <Button variant="ghost" className="text-destructive" onClick={() => del.mutate(ib.id)}>
              Remove
            </Button>
          </div>
        ))}
        {list.data?.inbounds?.length === 0 && (
          <p className="py-2 text-sm text-muted-foreground">No inbounds on this node yet.</p>
        )}
      </div>

      <form onSubmit={add} className="mt-4 space-y-3 border-t pt-4">
        <p className="text-xs font-medium text-muted-foreground">Add inbound</p>
        <div className="grid grid-cols-2 gap-2">
          <Input placeholder="Tag (e.g. vless-ws)" value={tag} onChange={(e) => setTag(e.target.value)} required />
          <Input placeholder="Port" value={port} onChange={(e) => setPort(e.target.value)} inputMode="numeric" required />
        </div>
        <div className="grid grid-cols-3 gap-2">
          <Select value={protocol} onChange={(e) => setProtocol(e.target.value)}>
            {PROTOCOLS.map((p) => (
              <option key={p} value={p}>{p}</option>
            ))}
          </Select>
          <Select value={network} onChange={(e) => setNetwork(e.target.value)}>
            {NETWORKS.map((n) => (
              <option key={n} value={n}>{n}</option>
            ))}
          </Select>
          <Select value={security} onChange={(e) => setSecurity(e.target.value)}>
            {SECURITIES.map((s) => (
              <option key={s} value={s}>{s}</option>
            ))}
          </Select>
        </div>
        <Input placeholder="SNI (comma-separated, optional)" value={sni} onChange={(e) => setSni(e.target.value)} />
        <div className="flex justify-end">
          <Button type="submit" disabled={create.isPending}>
            {create.isPending ? "Adding…" : "Add inbound"}
          </Button>
        </div>
      </form>
    </Modal>
  );
}
