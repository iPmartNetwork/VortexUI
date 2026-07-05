import { useState } from "react";
import { Plus, Trash2, Pencil, Waypoints } from "lucide-react";
import { routingHooks } from "@/api/policy-hooks";
import type { RoutingRule } from "@/api/types";
import { Badge, Button, Input, Select } from "@/components/ui";
import { GlassCard } from "@/components/veltrix";
import { Modal } from "@/components/Modal";
import { NodePicker } from "@/components/NodePicker";
import { useConfirm } from "@/components/confirm";
import { useToast } from "@/components/toast";
import { useI18n } from "@/i18n/i18n";
import { useTitle } from "@/lib/useTitle";

const csv = (s: string) => (s ? s.split(",").map((x) => x.trim()).filter(Boolean) : []);

export function Routing() {
  useTitle("Routing");
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
    <div className="space-y-5 animate-page-enter">
      <div className="flex flex-col lg:flex-row lg:items-center justify-between gap-3">
        <div>
          <h1 className="text-2xl font-bold text-fg tracking-tight">{t("nav.routing")}</h1>
          <p className="text-sm text-fg-muted mt-1">Traffic steering rules</p>
        </div>
        <div className="flex items-center gap-2 flex-shrink-0">
          <NodePicker value={node} onChange={setNode} />
          <Button onClick={() => setOpen(true)}><Plus size={14} /> {t("common.add")}</Button>
        </div>
      </div>

      <GlassCard hover={false} className="!p-0 overflow-hidden">
        <div className="divide-y divide-border/20">
          {sorted.map((r) => (
            <div key={r.id} className="flex items-center justify-between gap-2 px-5 py-3 text-sm hover:bg-surface-2/40">
              <div className="flex items-center gap-3 min-w-0">
                <span className="grid h-6 w-6 flex-shrink-0 place-items-center rounded-md bg-surface-2 text-xs text-fg-muted">{r.priority}</span>
                <span className="font-medium text-fg truncate">{r.name || "rule"}</span>
                <span className="text-fg-subtle flex-shrink-0">→</span>
                <Badge color={r.balancer_tag ? "on_hold" : "muted"}>{r.outbound_tag || r.balancer_tag || "—"}</Badge>
              </div>
              <div className="flex gap-1 flex-shrink-0">
                <Button variant="ghost" size="sm" onClick={() => edit(r)}><Pencil size={15} /></Button>
                <Button variant="ghost" size="sm" className="text-danger" onClick={() => remove(r)}><Trash2 size={15} /></Button>
              </div>
            </div>
          ))}
          {sorted.length === 0 && (
            <div className="flex flex-col items-center gap-2 px-5 py-10 text-center">
              <Waypoints size={22} className="text-fg-subtle" />
              <p className="text-sm text-fg-muted">{t("common.none")}</p>
            </div>
          )}
        </div>
      </GlassCard>

      <Modal open={open} onClose={() => { setOpen(false); setEditing(null); }} title={editing ? `Edit ${editing.name || "rule"}` : t("nav.routing")} className="max-w-lg">
        <form onSubmit={submit} className="space-y-3">
          <div className="grid grid-cols-1 gap-2 sm:grid-cols-3">
            <Input className="sm:col-span-2" placeholder="Name" value={f.name} onChange={set("name")} />
            <Input placeholder="Priority" value={f.priority} onChange={set("priority")} inputMode="numeric" />
          </div>
          <Input placeholder="Inbound tags (comma)" value={f.inbound_tags} onChange={set("inbound_tags")} />
          <div className="grid grid-cols-1 gap-2 sm:grid-cols-2">
            <Input placeholder="Domains (comma)" value={f.domains} onChange={set("domains")} />
            <Input placeholder="IP / CIDR (comma)" value={f.ip} onChange={set("ip")} />
          </div>
          <div className="grid grid-cols-1 gap-2 sm:grid-cols-2">
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
