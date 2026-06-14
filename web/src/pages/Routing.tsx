import { useState } from "react";
import { Trash2, Pencil } from "lucide-react";
import { routingHooks } from "@/api/policy-hooks";
import type { RoutingRule } from "@/api/types";
import { Badge, Button, Card, Input, PageHeader, Select } from "@/components/ui";
import { Modal } from "@/components/Modal";
import { NodePicker } from "@/components/NodePicker";
import { useConfirm } from "@/components/confirm";
import { useToast } from "@/components/toast";
import { useI18n } from "@/i18n/i18n";

const csv = (s: string) => (s ? s.split(",").map((x) => x.trim()).filter(Boolean) : []);

export function Routing() {
  const { t } = useI18n();
  const [node, setNode] = useState("");
  const [open, setOpen] = useState(false);
  const [editing, setEditing] = useState<RoutingRule | null>(null);
  const list = routingHooks.useList(node || null);
  const create = routingHooks.useCreate();
  const update = routingHooks.useUpdate();
  const del = routingHooks.useDelete();
  const confirm = useConfirm();
  const toast = useToast();

  const [f, setF] = useState({ name: "", priority: "1", inbound_tags: "", domains: "", ip: "", port: "", network: "", target: "" });
  const set = (k: string) => (e: React.ChangeEvent<HTMLInputElement | HTMLSelectElement>) => setF((s) => ({ ...s, [k]: e.target.value }));

  async function submit(e: React.FormEvent) {
    e.preventDefault();
    const [kind, tag] = f.target.split(":");
    const body = {
      node_id: node, name: f.name, priority: Number(f.priority) || 1,
      inbound_tags: csv(f.inbound_tags), domains: csv(f.domains), ip: csv(f.ip),
      port: f.port, network: f.network,
      outbound_tag: kind === "out" ? tag : "", balancer_tag: kind === "bal" ? tag : "",
      enabled: true,
    };
    if (editing) {
      await update.mutateAsync({ id: editing.id, body });
      toast.success("Updated");
      setEditing(null);
    } else {
      await create.mutateAsync(body);
      toast.success("✓");
    }
    setOpen(false);
    setF({ name: "", priority: "1", inbound_tags: "", domains: "", ip: "", port: "", network: "", target: "" });
  }

  function edit(r: RoutingRule) {
    const target = r.balancer_tag ? `bal:${r.balancer_tag}` : `out:${r.outbound_tag}`;
    setF({ name: r.name, priority: String(r.priority), inbound_tags: r.inbound_tags.join(", "), domains: r.domains.join(", "), ip: r.ip.join(", "), port: r.port, network: r.network, target });
    setEditing(r);
    setOpen(true);
  }

  async function remove(r: RoutingRule) {
    if (await confirm({ title: `${t("common.delete")} ${r.name || r.id}?`, confirmLabel: t("common.delete"), destructive: true })) {
      await del.mutateAsync(r.id);
      toast.success(t("common.delete"));
    }
  }

  const sorted = [...(list.data?.routing ?? [])].sort((a, b) => a.priority - b.priority);

  return (
    <div>
      <PageHeader title={t("nav.routing")} subtitle="Traffic steering rules">
        <NodePicker value={node} onChange={setNode} />
        <Button onClick={() => setOpen(true)}>{t("common.add")}</Button>
      </PageHeader>

      <Card className="p-0">
        <div className="divide-y divide-white/[0.05]">
          {sorted.map((r) => (
            <div key={r.id} className="flex items-center justify-between px-5 py-3 text-sm">
              <div className="flex items-center gap-3">
                <span className="grid h-6 w-6 place-items-center rounded-md bg-white/[0.06] text-xs text-fg-muted">{r.priority}</span>
                <span className="font-medium">{r.name || "rule"}</span>
                <span className="text-fg-subtle">→</span>
                <Badge color={r.balancer_tag ? "on_hold" : "muted"}>{r.outbound_tag || r.balancer_tag || "—"}</Badge>
              </div>
              <div className="flex gap-1">
                <Button variant="ghost" size="sm" onClick={() => edit(r)}><Pencil size={15} /></Button>
                <Button variant="ghost" size="sm" className="text-danger" onClick={() => remove(r)}><Trash2 size={15} /></Button>
              </div>
            </div>
          ))}
          {sorted.length === 0 && <p className="px-5 py-8 text-center text-sm text-fg-muted">{t("common.none")}</p>}
        </div>
      </Card>

      <Modal open={open} onClose={() => { setOpen(false); setEditing(null); }} title={editing ? `Edit ${editing.name || "rule"}` : t("nav.routing")} className="max-w-lg">
        <form onSubmit={submit} className="space-y-3">
          <div className="grid grid-cols-3 gap-2">
            <Input className="col-span-2" placeholder="Name" value={f.name} onChange={set("name")} />
            <Input placeholder="Priority" value={f.priority} onChange={set("priority")} inputMode="numeric" />
          </div>
          <Input placeholder="Inbound tags (comma)" value={f.inbound_tags} onChange={set("inbound_tags")} />
          <div className="grid grid-cols-2 gap-2">
            <Input placeholder="Domains (comma)" value={f.domains} onChange={set("domains")} />
            <Input placeholder="IP / CIDR (comma)" value={f.ip} onChange={set("ip")} />
          </div>
          <div className="grid grid-cols-2 gap-2">
            <Input placeholder='Port ("443" / "1-1000")' value={f.port} onChange={set("port")} />
            <Select value={f.network} onChange={set("network")}>
              <option value="">any network</option><option value="tcp">tcp</option><option value="udp">udp</option>
            </Select>
          </div>
          <Input placeholder="Target — out:<tag> or bal:<tag>" value={f.target} onChange={set("target")} required />
          <div className="flex justify-end gap-2 pt-1">
            <Button type="button" variant="ghost" onClick={() => { setOpen(false); setEditing(null); }}>{t("common.cancel")}</Button>
            <Button type="submit" disabled={create.isPending || update.isPending}>{editing ? t("common.save") : t("common.create")}</Button>
          </div>
        </form>
      </Modal>
    </div>
  );
}
