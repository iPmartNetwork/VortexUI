import { useState } from "react";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { api } from "@/api/client";
import { Button, Card, Input, PageHeader, Select, Badge } from "@/components/ui";
import { Modal } from "@/components/Modal";
import { useToast } from "@/components/toast";
import { useConfirm } from "@/components/confirm";

interface DecoySite {
  id: string;
  node_id: string | null;
  mode: string;
  target_url: string;
  static_html: string;
  enabled: boolean;
  created_at: string;
}

interface Node { id: string; name: string; }

export function DecoyWebsite() {
  const qc = useQueryClient();
  const toast = useToast();
  const confirm = useConfirm();
  const [createOpen, setCreateOpen] = useState(false);

  const { data } = useQuery({ queryKey: ["decoys"], queryFn: () => api<{ decoys: DecoySite[] }>("/api/decoys") });
  const { data: nodesData } = useQuery({ queryKey: ["nodes"], queryFn: () => api<{ nodes: Node[] }>("/api/nodes") });

  const delMut = useMutation({
    mutationFn: (id: string) => api<void>(`/api/decoys/${id}`, { method: "DELETE" }),
    onSuccess: () => qc.invalidateQueries({ queryKey: ["decoys"] }),
  });

  async function remove(d: DecoySite) {
    const ok = await confirm({ title: "Delete this decoy config?", confirmLabel: "Delete", destructive: true });
    if (!ok) return;
    await delMut.mutateAsync(d.id);
    toast.success("Deleted");
  }

  const nodeMap = Object.fromEntries(nodesData?.nodes?.map(n => [n.id, n.name]) ?? []);

  return (
    <div className="space-y-6 animate-fade-in">
      <div className="flex items-center justify-between">
        <PageHeader title="Decoy Website" subtitle="Serve a fake site to probers and scanners" />
        <Button onClick={() => setCreateOpen(true)}>New Decoy</Button>
      </div>
      <CreateDecoyModal open={createOpen} onClose={() => setCreateOpen(false)} nodes={nodesData?.nodes ?? []} />

      <div className="grid grid-cols-1 gap-4 lg:grid-cols-2">
        {data?.decoys?.map((d) => (
          <Card key={d.id} className="space-y-3">
            <div className="flex items-center justify-between">
              <Badge color={d.mode === "proxy" ? "active" : "limited"}>{d.mode}</Badge>
              <span className={`h-2 w-2 rounded-full ${d.enabled ? "bg-success" : "bg-fg-subtle"}`} />
            </div>
            <div className="text-xs text-fg-muted">
              {d.node_id ? `Node: ${nodeMap[d.node_id] || d.node_id.slice(0, 8)}` : "Global (all nodes)"}
            </div>
            {d.mode === "proxy" && <div className="text-xs font-mono text-fg break-all">{d.target_url}</div>}
            {d.mode === "static" && <div className="text-xs text-fg-muted">{d.static_html.slice(0, 100)}...</div>}
            <div className="flex justify-end pt-2">
              <Button variant="ghost" size="sm" className="text-destructive text-xs" onClick={() => remove(d)}>Delete</Button>
            </div>
          </Card>
        ))}
        {(!data?.decoys || data.decoys.length === 0) && (
          <p className="col-span-full text-center text-sm text-fg-muted py-8">No decoy sites configured. Probers will see a connection refused.</p>
        )}
      </div>
    </div>
  );
}

function CreateDecoyModal({ open, onClose, nodes }: { open: boolean; onClose: () => void; nodes: Node[] }) {
  const qc = useQueryClient();
  const toast = useToast();
  const [mode, setMode] = useState("proxy");
  const [nodeId, setNodeId] = useState("");
  const [targetUrl, setTargetUrl] = useState("https://www.google.com");
  const [staticHtml, setStaticHtml] = useState("");

  const create = useMutation({
    mutationFn: (input: Record<string, unknown>) => api("/api/decoys", { method: "POST", body: input }),
    onSuccess: () => { qc.invalidateQueries({ queryKey: ["decoys"] }); onClose(); toast.success("Decoy created"); },
    onError: (e: any) => toast.error(e.message),
  });

  function submit(e: React.FormEvent) {
    e.preventDefault();
    create.mutate({
      node_id: nodeId || undefined,
      mode,
      target_url: mode === "proxy" ? targetUrl : "",
      static_html: mode === "static" ? staticHtml : "",
    });
  }

  return (
    <Modal open={open} onClose={onClose} title="New Decoy Site">
      <form onSubmit={submit} className="space-y-3">
        <Select value={nodeId} onChange={(e) => setNodeId(e.target.value)}>
          <option value="">Global (all nodes)</option>
          {nodes.map(n => <option key={n.id} value={n.id}>{n.name}</option>)}
        </Select>
        <Select value={mode} onChange={(e) => setMode(e.target.value)}>
          <option value="proxy">Reverse Proxy</option>
          <option value="static">Static HTML</option>
        </Select>
        {mode === "proxy" && (
          <Input placeholder="Target URL (e.g. https://www.google.com)" value={targetUrl} onChange={(e) => setTargetUrl(e.target.value)} required />
        )}
        {mode === "static" && (
          <textarea
            placeholder="Paste HTML content..."
            value={staticHtml}
            onChange={(e) => setStaticHtml(e.target.value)}
            className="field min-h-[120px] resize-y font-mono text-xs"
            required
          />
        )}
        <div className="flex justify-end gap-2 pt-2">
          <Button type="button" variant="ghost" onClick={onClose}>Cancel</Button>
          <Button type="submit" disabled={create.isPending}>Create</Button>
        </div>
      </form>
    </Modal>
  );
}
