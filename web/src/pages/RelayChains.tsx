import { useState } from "react";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { api } from "@/api/client";
import { Button, Card, Input, PageHeader, Select } from "@/components/ui";
import { Modal } from "@/components/Modal";
import { useToast } from "@/components/toast";
import { useConfirm } from "@/components/confirm";
import { useI18n } from "@/i18n/i18n";

interface RelayHop {
  type: string;
  address: string;
  port: number;
  protocol: string;
  sni: string;
  path: string;
  host: string;
  note: string;
}

interface RelayChain {
  id: string;
  name: string;
  node_id: string;
  hops: RelayHop[];
  enabled: boolean;
  created_at: string;
}

interface Node { id: string; name: string; }

export function RelayChains() {
  const { t } = useI18n();
  const qc = useQueryClient();
  const toast = useToast();
  const confirm = useConfirm();
  const [createOpen, setCreateOpen] = useState(false);

  const { data } = useQuery({ queryKey: ["relay-chains"], queryFn: () => api<{ chains: RelayChain[] }>("/api/relays") });
  const { data: nodesData } = useQuery({ queryKey: ["nodes"], queryFn: () => api<{ nodes: Node[] }>("/api/nodes") });

  const delMut = useMutation({
    mutationFn: (id: string) => api<void>(`/api/relays/${id}`, { method: "DELETE" }),
    onSuccess: () => qc.invalidateQueries({ queryKey: ["relay-chains"] }),
  });

  async function remove(c: RelayChain) {
    const ok = await confirm({ title: `Delete chain "${c.name}"?`, confirmLabel: "Delete", destructive: true });
    if (!ok) return;
    await delMut.mutateAsync(c.id);
    toast.success("Deleted");
  }

  const nodeMap = Object.fromEntries(nodesData?.nodes?.map(n => [n.id, n.name]) ?? []);

  return (
    <div className="space-y-6 animate-page-enter">
      <div className="flex items-center justify-between">
        <PageHeader title={t("relay.title")} subtitle={t("relay.subtitle")} />
        <Button onClick={() => setCreateOpen(true)}>{t("relay.newChain")}</Button>
      </div>
      <div className="rounded-lg border border-border/40 bg-surface-2/20 p-4 text-xs text-fg-muted space-y-2">
        <p className="font-medium text-fg text-sm">{t("relay.infoTitle")}</p>
        <p>{t("relay.infoDesc")}</p>
        <ul className="list-disc pl-4 space-y-1">
          <li><strong>CDN</strong> — {t("relay.cdn")}</li>
          <li><strong>Relay</strong> — {t("relay.relay")}</li>
          <li><strong>Worker</strong> — {t("relay.worker")}</li>
        </ul>
      </div>
      <CreateChainModal open={createOpen} onClose={() => setCreateOpen(false)} nodes={nodesData?.nodes ?? []} />

      <div className="grid grid-cols-1 gap-4 lg:grid-cols-2">
        {data?.chains?.map((c) => (
          <Card key={c.id} className="space-y-3">
            <div className="flex items-center justify-between">
              <h3 className="text-sm font-bold text-fg">{c.name}</h3>
              <span className={`h-2 w-2 rounded-full ${c.enabled ? "bg-success" : "bg-fg-subtle"}`} />
            </div>
            <div className="text-xs text-fg-muted">Node: {nodeMap[c.node_id] || c.node_id.slice(0, 8)}</div>
            <div className="flex items-center gap-1 flex-wrap">
              {c.hops.map((h, i) => (
                <span key={i} className="inline-flex items-center gap-1">
                  <span className="rounded bg-surface-2 px-2 py-0.5 text-xs font-mono">{h.type}: {h.address}:{h.port}</span>
                  {i < c.hops.length - 1 && <span className="text-fg-subtle">→</span>}
                </span>
              ))}
              <span className="text-fg-subtle">→</span>
              <span className="rounded bg-primary/10 px-2 py-0.5 text-xs font-medium text-primary">Node</span>
            </div>
            <div className="flex justify-end pt-2">
              <Button variant="ghost" size="sm" className="text-destructive text-xs" onClick={() => remove(c)}>Delete</Button>
            </div>
          </Card>
        ))}
        {(!data?.chains || data.chains.length === 0) && (
          <p className="col-span-full text-center text-sm text-fg-muted py-8">{t("relay.noChains")}</p>
        )}
      </div>
    </div>
  );
}

function CreateChainModal({ open, onClose, nodes }: { open: boolean; onClose: () => void; nodes: Node[] }) {
  const qc = useQueryClient();
  const toast = useToast();
  const [name, setName] = useState("");
  const [nodeId, setNodeId] = useState("");
  const [hops, setHops] = useState<RelayHop[]>([
    { type: "cdn", address: "", port: 443, protocol: "ws", sni: "", path: "/", host: "", note: "" },
  ]);

  const create = useMutation({
    mutationFn: (input: { name: string; node_id: string; hops: RelayHop[] }) =>
      api("/api/relays", { method: "POST", body: input }),
    onSuccess: () => { qc.invalidateQueries({ queryKey: ["relay-chains"] }); onClose(); toast.success("Chain created"); },
    onError: (e: any) => toast.error(e.message),
  });

  function addHop() {
    setHops([...hops, { type: "relay", address: "", port: 443, protocol: "tcp", sni: "", path: "", host: "", note: "" }]);
  }
  function removeHop(idx: number) { setHops(hops.filter((_, i) => i !== idx)); }
  function updateHop(idx: number, field: keyof RelayHop, value: string | number) {
    const copy = [...hops];
    (copy[idx] as any)[field] = value;
    setHops(copy);
  }

  return (
    <Modal open={open} onClose={onClose} title="New Relay Chain" className="max-w-xl">
      <form onSubmit={(e) => { e.preventDefault(); create.mutate({ name, node_id: nodeId, hops }); }} className="space-y-4">
        <div className="grid grid-cols-2 gap-2">
          <Input placeholder="Chain name" value={name} onChange={(e) => setName(e.target.value)} required />
          <Select value={nodeId} onChange={(e) => setNodeId(e.target.value)}>
            <option value="">Target node...</option>
            {nodes.map(n => <option key={n.id} value={n.id}>{n.name}</option>)}
          </Select>
        </div>
        <div className="space-y-3">
          <div className="flex items-center justify-between">
            <span className="text-xs text-fg-subtle font-medium">Hops (in order)</span>
            <Button type="button" variant="ghost" size="sm" onClick={addHop}>+ Add hop</Button>
          </div>
          {hops.map((h, i) => (
            <div key={i} className="rounded-lg border border-border/40 p-3 space-y-2">
              <div className="flex items-center justify-between">
                <span className="text-xs font-medium text-fg-muted">Hop {i + 1}</span>
                <Button type="button" variant="ghost" size="sm" className="text-destructive h-6" onClick={() => removeHop(i)}>×</Button>
              </div>
              <div className="grid grid-cols-3 gap-2">
                <select className="field text-xs" value={h.type} onChange={(e) => updateHop(i, "type", e.target.value)}>
                  <option value="cdn">CDN</option>
                  <option value="relay">Relay</option>
                  <option value="worker">Worker</option>
                </select>
                <Input placeholder="Address" value={h.address} onChange={(e) => updateHop(i, "address", e.target.value)} />
                <Input placeholder="Port" value={h.port} onChange={(e) => updateHop(i, "port", Number(e.target.value))} inputMode="numeric" />
              </div>
              <div className="grid grid-cols-3 gap-2">
                <select className="field text-xs" value={h.protocol} onChange={(e) => updateHop(i, "protocol", e.target.value)}>
                  <option value="ws">WebSocket</option>
                  <option value="grpc">gRPC</option>
                  <option value="tcp">TCP</option>
                </select>
                <Input placeholder="SNI" value={h.sni} onChange={(e) => updateHop(i, "sni", e.target.value)} />
                <Input placeholder="Path" value={h.path} onChange={(e) => updateHop(i, "path", e.target.value)} />
              </div>
            </div>
          ))}
        </div>
        <div className="flex justify-end gap-2 pt-2">
          <Button type="button" variant="ghost" onClick={onClose}>Cancel</Button>
          <Button type="submit" disabled={create.isPending || !name || !nodeId}>Create</Button>
        </div>
      </form>
    </Modal>
  );
}
