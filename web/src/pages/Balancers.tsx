import { useState } from "react";
import { Trash2, Pencil } from "lucide-react";
import { balancerHooks } from "@/api/policy-hooks";
import type { Balancer } from "@/api/types";
import { Badge, Button, Card, Input, PageHeader, Select } from "@/components/ui";
import { Modal } from "@/components/Modal";
import { NodePicker } from "@/components/NodePicker";
import { useConfirm } from "@/components/confirm";
import { useToast } from "@/components/toast";
import { useI18n } from "@/i18n/i18n";

const STRATEGIES = ["random", "roundRobin", "leastPing", "leastLoad"];
const csv = (s: string) => (s ? s.split(",").map((x) => x.trim()).filter(Boolean) : []);

export function Balancers() {
  const { t } = useI18n();
  const [node, setNode] = useState("");
  const [open, setOpen] = useState(false);
  const [editing, setEditing] = useState<Balancer | null>(null);
  const list = balancerHooks.useList(node || null);
  const create = balancerHooks.useCreate();
  const update = balancerHooks.useUpdate();
  const del = balancerHooks.useDelete();
  const confirm = useConfirm();
  const toast = useToast();

  const [f, setF] = useState({ tag: "", selectors: "", strategy: "leastPing", probe_url: "", probe_interval: "10s" });
  const set = (k: string) => (e: React.ChangeEvent<HTMLInputElement | HTMLSelectElement>) => setF((s) => ({ ...s, [k]: e.target.value }));
  const observe = f.strategy === "leastPing" || f.strategy === "leastLoad";

  async function submit(e: React.FormEvent) {
    e.preventDefault();
    const body = {
      node_id: node, tag: f.tag, selectors: csv(f.selectors), strategy: f.strategy,
      observe, probe_url: f.probe_url, probe_interval: f.probe_interval, enabled: true,
    };
    if (editing) {
      await update.mutateAsync({ id: editing.id, body });
      toast.success(`${editing.tag} updated`);
      setEditing(null);
    } else {
      await create.mutateAsync(body);
      toast.success(`${f.tag} ✓`);
    }
    setOpen(false);
    setF({ tag: "", selectors: "", strategy: "leastPing", probe_url: "", probe_interval: "10s" });
  }

  function edit(b: Balancer) {
    setF({ tag: b.tag, selectors: b.selectors.join(", "), strategy: b.strategy, probe_url: b.probe_url, probe_interval: b.probe_interval || "10s" });
    setEditing(b);
    setOpen(true);
  }

  async function remove(b: Balancer) {
    if (await confirm({ title: `${t("common.delete")} ${b.tag}?`, confirmLabel: t("common.delete"), destructive: true })) {
      await del.mutateAsync(b.id);
      toast.success(t("common.delete"));
    }
  }

  return (
    <div>
      <PageHeader title={t("nav.balancers")} subtitle="Outbound load balancing">
        <NodePicker value={node} onChange={setNode} />
        <Button onClick={() => setOpen(true)}>{t("common.add")}</Button>
      </PageHeader>

      <Card className="p-0">
        <div className="divide-y divide-white/[0.05]">
          {list.data?.balancers?.map((b) => (
            <div key={b.id} className="flex items-center justify-between px-5 py-3 text-sm">
              <div className="flex items-center gap-3">
                <span className="font-medium">{b.tag}</span>
                <Badge color="on_hold">{b.strategy}</Badge>
                <span className="text-xs text-fg-muted">{b.selectors.join(", ")}</span>
              </div>
              <div className="flex gap-1">
                <Button variant="ghost" size="sm" onClick={() => edit(b)}><Pencil size={15} /></Button>
                <Button variant="ghost" size="sm" className="text-danger" onClick={() => remove(b)}><Trash2 size={15} /></Button>
              </div>
            </div>
          ))}
          {list.data?.balancers?.length === 0 && <p className="px-5 py-8 text-center text-sm text-fg-muted">{t("common.none")}</p>}
        </div>
      </Card>

      <Modal open={open} onClose={() => { setOpen(false); setEditing(null); }} title={editing ? `Edit ${editing.tag}` : t("nav.balancers")}>
        <form onSubmit={submit} className="space-y-3">
          <Input placeholder="Tag" value={f.tag} onChange={set("tag")} required />
          <Input placeholder="Selectors — outbound tag prefixes (comma)" value={f.selectors} onChange={set("selectors")} required />
          <Select value={f.strategy} onChange={set("strategy")}>
            {STRATEGIES.map((s) => <option key={s} value={s}>{s}</option>)}
          </Select>
          {observe && (
            <div className="grid grid-cols-2 gap-2">
              <Input placeholder="Probe URL" value={f.probe_url} onChange={set("probe_url")} />
              <Input placeholder="Interval (10s)" value={f.probe_interval} onChange={set("probe_interval")} />
            </div>
          )}
          <div className="flex justify-end gap-2 pt-1">
            <Button type="button" variant="ghost" onClick={() => { setOpen(false); setEditing(null); }}>{t("common.cancel")}</Button>
            <Button type="submit" disabled={create.isPending || update.isPending}>{editing ? t("common.save") : t("common.create")}</Button>
          </div>
        </form>
      </Modal>
    </div>
  );
}
