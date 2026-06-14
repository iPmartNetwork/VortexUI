import { useState } from "react";
import { useDeleteNode, useNodes } from "@/api/hooks";
import type { Node } from "@/api/types";
import { Badge, Button, Card } from "@/components/ui";
import { CreateNodeModal } from "@/components/CreateNodeModal";
import { EditNodeModal } from "@/components/EditNodeModal";
import { NodeInboundsModal } from "@/components/NodeInboundsModal";
import { useConfirm } from "@/components/confirm";
import { useToast } from "@/components/toast";

export function Nodes() {
  const { data, isLoading, error } = useNodes();
  const del = useDeleteNode();
  const confirm = useConfirm();
  const toast = useToast();
  const [createOpen, setCreateOpen] = useState(false);
  const [editing, setEditing] = useState<Node | null>(null);
  const [managing, setManaging] = useState<Node | null>(null);

  async function remove(n: Node) {
    const ok = await confirm({
      title: `Delete node ${n.name}?`,
      message: "Its inbounds are removed and the agent is deregistered.",
      confirmLabel: "Delete",
      destructive: true,
    });
    if (!ok) return;
    try {
      await del.mutateAsync(n.id);
      toast.success(`Deleted ${n.name}`);
    } catch {
      toast.error("Delete failed");
    }
  }

  return (
    <div className="space-y-6">
      <CreateNodeModal open={createOpen} onClose={() => setCreateOpen(false)} />
      <EditNodeModal node={editing} onClose={() => setEditing(null)} />
      <NodeInboundsModal node={managing} onClose={() => setManaging(null)} />

      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-bold tracking-tight">Nodes</h1>
        <Button onClick={() => setCreateOpen(true)}>New node</Button>
      </div>

      {isLoading && <p className="text-sm text-muted-foreground">Loading…</p>}
      {error && <p className="text-sm text-destructive">Failed to load nodes</p>}

      <div className="grid grid-cols-1 gap-4 md:grid-cols-2 lg:grid-cols-3">
        {data?.nodes.map((n) => (
          <Card key={n.id} className="space-y-3">
            <div className="flex items-center justify-between">
              <span className="font-semibold">{n.name}</span>
              <Badge color={n.health.core_running ? "active" : "disabled"}>
                {n.health.core_running ? "running" : "down"}
              </Badge>
            </div>
            <p className="text-sm text-muted-foreground">{n.address}</p>
            <div className="flex gap-4 text-xs text-muted-foreground">
              <span className="uppercase">{n.core}</span>
              <span>CPU {n.health.cpu_percent.toFixed(0)}%</span>
              <span>{n.health.connections} conns</span>
            </div>
            <div className="flex gap-2 border-t pt-3">
              <Button variant="ghost" className="flex-1" onClick={() => setManaging(n)}>
                Inbounds
              </Button>
              <Button variant="ghost" onClick={() => setEditing(n)}>
                Edit
              </Button>
              <Button variant="ghost" className="text-destructive" onClick={() => remove(n)}>
                Delete
              </Button>
            </div>
          </Card>
        ))}
      </div>
      {data?.nodes.length === 0 && <p className="text-sm text-muted-foreground">No nodes yet — add one to get started.</p>}
    </div>
  );
}
