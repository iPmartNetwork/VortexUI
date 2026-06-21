import { useEffect, useState } from "react";
import { Trash2, Pencil, CheckCircle2, Server } from "lucide-react";
import {
  useApplyRoutingPack,
  useCreateRoutingPack,
  useDeleteRoutingPack,
  useNodes,
  useRoutingPacks,
  useSetDefaultRoutingPack,
  useUpdateRoutingPack,
  type PackRoutingRule,
  type RoutingPack,
  type RoutingPackBody,
} from "@/api/hooks";
import { Badge, Button, Card, Input, PageHeader, Select } from "@/components/ui";
import { Modal } from "@/components/Modal";
import { useConfirm } from "@/components/confirm";
import { useToast } from "@/components/toast";
import { useI18n } from "@/i18n/i18n";

const csv = (s: string) => (s ? s.split(",").map((x) => x.trim()).filter(Boolean) : []);

// A row in the simple rules editor. Targets follow the same "out:<tag>" /
// "bal:<tag>" shorthand as the Routing page so the editor stays familiar.
interface RuleForm {
  name: string;
  priority: string;
  inbound_tags: string;
  domains: string;
  ip: string;
  port: string;
  network: string;
  target: string;
}

const emptyRule: RuleForm = { name: "", priority: "1", inbound_tags: "", domains: "", ip: "", port: "", network: "", target: "" };

function ruleToForm(r: PackRoutingRule): RuleForm {
  const target = r.balancer_tag ? `bal:${r.balancer_tag}` : r.outbound_tag ? `out:${r.outbound_tag}` : "";
  return {
    name: r.name ?? "",
    priority: String(r.priority ?? 1),
    inbound_tags: (r.inbound_tags ?? []).join(", "),
    domains: (r.domains ?? []).join(", "),
    ip: (r.ip ?? []).join(", "),
    port: r.port ?? "",
    network: r.network ?? "",
    target,
  };
}

function formToRule(f: RuleForm): PackRoutingRule {
  const [kind, tag] = f.target.split(":");
  return {
    priority: Number(f.priority) || 1,
    name: f.name,
    inbound_tags: csv(f.inbound_tags),
    domains: csv(f.domains),
    ip: csv(f.ip),
    port: f.port,
    network: f.network,
    outbound_tag: kind === "out" ? tag : "",
    balancer_tag: kind === "bal" ? tag : "",
  };
}

export function RoutingPacks() {
  const { t } = useI18n();
  const toast = useToast();
  const confirm = useConfirm();

  const packs = useRoutingPacks();
  const create = useCreateRoutingPack();
  const update = useUpdateRoutingPack();
  const del = useDeleteRoutingPack();
  const apply = useApplyRoutingPack();
  const setDefault = useSetDefaultRoutingPack();

  const [editorOpen, setEditorOpen] = useState(false);
  const [editing, setEditing] = useState<RoutingPack | null>(null);
  const [applyFor, setApplyFor] = useState<RoutingPack | null>(null);

  const list = packs.data?.packs ?? [];

  function openCreate() {
    setEditing(null);
    setEditorOpen(true);
  }
  function openEdit(p: RoutingPack) {
    setEditing(p);
    setEditorOpen(true);
  }

  async function remove(p: RoutingPack) {
    if (await confirm({ title: `${t("common.delete")} "${p.name}"?`, confirmLabel: t("common.delete"), destructive: true })) {
      await del.mutateAsync(p.id);
      toast.success(t("common.delete"));
    }
  }

  async function makeDefault(p: RoutingPack) {
    await setDefault.mutateAsync(p.id);
    toast.success(`Default: ${p.name}`);
  }

  return (
    <div className="space-y-6 animate-fade-in">
      <PageHeader title={t("nav.routingPacks")} subtitle="Reusable smart-routing rule sets">
        <Button onClick={openCreate}>{t("common.add")}</Button>
      </PageHeader>

      <div className="grid grid-cols-1 gap-4 lg:grid-cols-2">
        {list.map((p) => (
          <Card key={p.id} className="space-y-3">
            <div className="flex items-start justify-between gap-2">
              <div className="min-w-0">
                <div className="flex items-center gap-2">
                  <h3 className="truncate text-sm font-bold text-fg">{p.name}</h3>
                  <Badge color={p.builtin ? "muted" : "on_hold"}>{p.builtin ? "Built-in" : "Custom"}</Badge>
                </div>
                {p.description && <p className="mt-0.5 truncate text-xs text-fg-muted">{p.description}</p>}
              </div>
            </div>
            <div className="flex flex-wrap items-center gap-2 text-xs text-fg-muted">
              {p.category && <span className="rounded-md bg-surface-2/60 px-2 py-0.5">{p.category}</span>}
              <span>{p.rules.length} {p.rules.length === 1 ? "rule" : "rules"}</span>
            </div>
            <div className="flex flex-wrap justify-end gap-1.5 pt-1">
              <Button variant="ghost" size="sm" onClick={() => setApplyFor(p)}><Server size={14} /> Apply</Button>
              <Button variant="ghost" size="sm" onClick={() => makeDefault(p)}><CheckCircle2 size={14} /> Default</Button>
              {!p.builtin && (
                <>
                  <Button variant="ghost" size="sm" onClick={() => openEdit(p)}><Pencil size={14} /></Button>
                  <Button variant="ghost" size="sm" className="text-danger" onClick={() => remove(p)}><Trash2 size={14} /></Button>
                </>
              )}
            </div>
          </Card>
        ))}
        {list.length === 0 && (
          <p className="col-span-full py-8 text-center text-sm text-fg-muted">{t("common.none")}</p>
        )}
      </div>

      <PackEditor
        open={editorOpen}
        pack={editing}
        onClose={() => setEditorOpen(false)}
        onSubmit={async (body) => {
          if (editing) {
            await update.mutateAsync({ id: editing.id, body });
            toast.success(t("common.save"));
          } else {
            await create.mutateAsync(body);
            toast.success(t("common.create"));
          }
          setEditorOpen(false);
        }}
        pending={create.isPending || update.isPending}
      />

      <ApplyModal
        pack={applyFor}
        onClose={() => setApplyFor(null)}
        onApply={async (nodeId) => {
          const res = await apply.mutateAsync({ node_id: nodeId, pack_id: applyFor!.id });
          if (res.warning) toast.error(res.warning);
          else toast.success("Applied");
          setApplyFor(null);
        }}
        pending={apply.isPending}
      />
    </div>
  );
}

function PackEditor({
  open,
  pack,
  onClose,
  onSubmit,
  pending,
}: {
  open: boolean;
  pack: RoutingPack | null;
  onClose: () => void;
  onSubmit: (body: RoutingPackBody) => void | Promise<void>;
  pending: boolean;
}) {
  const { t } = useI18n();
  const [name, setName] = useState("");
  const [description, setDescription] = useState("");
  const [category, setCategory] = useState("");
  const [rules, setRules] = useState<RuleForm[]>([emptyRule]);

  // Re-seed the form whenever the modal opens for a new/different pack.
  useEffect(() => {
    if (!open) return;
    setName(pack?.name ?? "");
    setDescription(pack?.description ?? "");
    setCategory(pack?.category ?? "");
    setRules(pack && pack.rules.length ? pack.rules.map(ruleToForm) : [{ ...emptyRule }]);
  }, [open, pack]);

  function updateRule(idx: number, field: keyof RuleForm, value: string) {
    setRules((rs) => rs.map((r, i) => (i === idx ? { ...r, [field]: value } : r)));
  }
  function addRule() {
    setRules((rs) => [...rs, { ...emptyRule }]);
  }
  function removeRule(idx: number) {
    setRules((rs) => rs.filter((_, i) => i !== idx));
  }

  function submit(e: React.FormEvent) {
    e.preventDefault();
    const body: RoutingPackBody = {
      name,
      description,
      category,
      rules: rules.filter((r) => r.target.trim()).map(formToRule),
    };
    onSubmit(body);
  }

  return (
    <Modal open={open} onClose={onClose} title={pack ? `Edit ${pack.name}` : "New Routing Pack"} className="max-w-2xl">
      <form onSubmit={submit} className="space-y-3">
        <div className="grid grid-cols-2 gap-2">
          <Input placeholder="Name" value={name} onChange={(e) => setName(e.target.value)} required />
          <Input placeholder="Category" value={category} onChange={(e) => setCategory(e.target.value)} />
        </div>
        <Input placeholder="Description" value={description} onChange={(e) => setDescription(e.target.value)} />

        <div className="space-y-2">
          <div className="flex items-center justify-between">
            <span className="text-xs font-medium text-fg-subtle">Rules</span>
            <Button type="button" variant="ghost" size="sm" onClick={addRule}>+ Add rule</Button>
          </div>
          {rules.map((r, i) => (
            <div key={i} className="space-y-2 rounded-xl border border-border/40 p-3">
              <div className="grid grid-cols-3 gap-2">
                <Input className="col-span-2" placeholder="Name" value={r.name} onChange={(e) => updateRule(i, "name", e.target.value)} />
                <Input placeholder="Priority" value={r.priority} onChange={(e) => updateRule(i, "priority", e.target.value)} inputMode="numeric" />
              </div>
              <Input placeholder="Inbound tags (comma)" value={r.inbound_tags} onChange={(e) => updateRule(i, "inbound_tags", e.target.value)} />
              <div className="grid grid-cols-2 gap-2">
                <Input placeholder="Domains (comma)" value={r.domains} onChange={(e) => updateRule(i, "domains", e.target.value)} />
                <Input placeholder="IP / CIDR (comma)" value={r.ip} onChange={(e) => updateRule(i, "ip", e.target.value)} />
              </div>
              <div className="grid grid-cols-2 gap-2">
                <Input placeholder='Port ("443" / "1-1000")' value={r.port} onChange={(e) => updateRule(i, "port", e.target.value)} />
                <Select value={r.network} onChange={(e) => updateRule(i, "network", e.target.value)}>
                  <option value="">any network</option>
                  <option value="tcp">tcp</option>
                  <option value="udp">udp</option>
                </Select>
              </div>
              <div className="flex items-center gap-2">
                <Input className="flex-1" placeholder="Target — out:<tag> or bal:<tag>" value={r.target} onChange={(e) => updateRule(i, "target", e.target.value)} />
                <Button type="button" variant="ghost" size="sm" className="text-danger" onClick={() => removeRule(i)} disabled={rules.length === 1}><Trash2 size={14} /></Button>
              </div>
            </div>
          ))}
        </div>

        <div className="flex justify-end gap-2 pt-1">
          <Button type="button" variant="ghost" onClick={onClose}>{t("common.cancel")}</Button>
          <Button type="submit" disabled={pending || !name}>{pack ? t("common.save") : t("common.create")}</Button>
        </div>
      </form>
    </Modal>
  );
}

function ApplyModal({
  pack,
  onClose,
  onApply,
  pending,
}: {
  pack: RoutingPack | null;
  onClose: () => void;
  onApply: (nodeId: string) => void | Promise<void>;
  pending: boolean;
}) {
  const { t } = useI18n();
  const { data } = useNodes();
  const nodes = data?.nodes ?? [];
  const [nodeId, setNodeId] = useState("");

  useEffect(() => {
    if (pack && !nodeId && nodes.length) setNodeId(nodes[0].id);
  }, [pack, nodeId, nodes]);

  return (
    <Modal open={!!pack} onClose={onClose} title={pack ? `Apply "${pack.name}" to node` : ""} className="max-w-md">
      <div className="space-y-4">
        <p className="text-sm text-fg-muted">
          Applies the pack's rules to the selected node and resyncs its core.
        </p>
        <Select value={nodeId} onChange={(e) => setNodeId(e.target.value)}>
          {nodes.length === 0 && <option value="">No nodes</option>}
          {nodes.map((n) => (
            <option key={n.id} value={n.id}>{n.name} · {n.core}</option>
          ))}
        </Select>
        <div className="flex justify-end gap-2">
          <Button variant="ghost" onClick={onClose}>{t("common.cancel")}</Button>
          <Button onClick={() => onApply(nodeId)} disabled={!nodeId || pending}>Apply</Button>
        </div>
      </div>
    </Modal>
  );
}
