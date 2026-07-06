import { useState } from "react";
import { Link, useSearchParams } from "react-router-dom";
import {
  ArrowRight,
  Globe,
  Layers,
  Pencil,
  Plus,
  RefreshCw,
  Scale,
  Server,
  Shield,
  Trash2,
  Copy,
  CheckCircle2,
  ArrowUpRight,
} from "lucide-react";
import {
  useApplyRoutingPack,
  useBalancersFleet,
  useCreateRoutingPack,
  useDefaultRoutingPack,
  useDeleteRoutingPack,
  useNodes,
  useRoutingPacks,
  useSetDefaultRoutingPack,
  useUpdateRoutingPack,
  type RoutingPack,
  type RoutingPackBody,
} from "@/api/hooks";
import { balancerHooks, useUpdateGeo } from "@/api/policy-hooks";
import type { BalancerFleetRow } from "@/api/types";
import { Badge, Button, Input, Select, Switch } from "@/components/ui";
import { Modal } from "@/components/Modal";
import { NodePicker } from "@/components/NodePicker";
import { GlassCard, StatusBadge } from "@/components/veltrix";
import { useConfirm } from "@/components/confirm";
import { useToast } from "@/components/toast";
import { useI18n } from "@/i18n/i18n";
import { useTitle } from "@/lib/useTitle";
import { useAuth } from "@/auth/auth";
import { cn } from "@/lib/utils";
import { ApplyModal, PackEditor } from "@/pages/RoutingPacks";
import { OutboundsTab } from "@/pages/Outbounds";

type Tab = "packs" | "balancers" | "outbounds";

const STRATEGIES = ["random", "roundRobin", "leastPing", "leastLoad"];
const csv = (s: string) => (s ? s.split(",").map((x) => x.trim()).filter(Boolean) : []);

function formatStrategy(s: string): string {
  return s.replace(/([a-z])([A-Z])/g, "$1 $2").toUpperCase().replace(/\s+/g, "");
}

function packMatchers(p: RoutingPack): number {
  let n = 0;
  for (const r of p.rules ?? []) {
    n += (r.domains?.length ?? 0) + (r.ip?.length ?? 0) + 1;
  }
  return n;
}

function packIcon(p: RoutingPack) {
  const name = p.name.toLowerCase();
  if (name.includes("ad") || name.includes("block") || name.includes("malware")) {
    return { icon: Shield, color: "text-danger bg-danger/10" };
  }
  if (name.includes("iran") || name.includes("direct") || p.category?.toLowerCase().includes("geo")) {
    return { icon: Globe, color: "text-success bg-success/10" };
  }
  return { icon: Layers, color: "text-primary bg-primary/10" };
}

export function RoutingBalancers() {
  useTitle("Routing");
  const { t } = useI18n();
  const [searchParams, setSearchParams] = useSearchParams();
  const raw = searchParams.get("tab");
  const tab: Tab = raw === "balancers" || raw === "outbounds" ? raw : "packs";

  function setTab(next: Tab) {
    if (next === "packs") setSearchParams({}, { replace: true });
    else setSearchParams({ tab: next }, { replace: true });
  }

  return (
    <div className="space-y-5 animate-page-enter">
      <div className="flex flex-col lg:flex-row lg:items-start justify-between gap-4">
        <div>
          <div className="flex flex-wrap items-center gap-2">
            <h1 className="text-2xl font-bold text-fg tracking-tight">{t("routing.pageTitle")}</h1>
            <Badge color="muted">Xray + sing-box</Badge>
          </div>
          <p className="text-sm text-fg-muted mt-1 max-w-2xl">{t("routing.pageSubtitle")}</p>
        </div>
        <div className="flex flex-wrap rounded-xl border border-border/70 bg-surface-2/50 p-0.5 gap-0.5">
          <button
            type="button"
            onClick={() => setTab("packs")}
            className={cn(
              "flex items-center gap-1.5 px-3.5 py-2 rounded-lg text-xs font-semibold transition-colors whitespace-nowrap",
              tab === "packs" ? "bg-primary text-primary-fg shadow-sm" : "text-fg-muted hover:text-fg",
            )}
          >
            <Layers size={14} />
            {t("routing.tabPacks")}
          </button>
          <button
            type="button"
            onClick={() => setTab("balancers")}
            className={cn(
              "flex items-center gap-1.5 px-3.5 py-2 rounded-lg text-xs font-semibold transition-colors whitespace-nowrap",
              tab === "balancers" ? "bg-primary text-primary-fg shadow-sm" : "text-fg-muted hover:text-fg",
            )}
          >
            <Scale size={14} />
            {t("routing.tabBalancers")}
          </button>
          <button
            type="button"
            onClick={() => setTab("outbounds")}
            className={cn(
              "flex items-center gap-1.5 px-3.5 py-2 rounded-lg text-xs font-semibold transition-colors whitespace-nowrap",
              tab === "outbounds" ? "bg-primary text-primary-fg shadow-sm" : "text-fg-muted hover:text-fg",
            )}
          >
            <ArrowUpRight size={14} />
            {t("routing.tabOutbounds")}
          </button>
        </div>
      </div>

      {tab === "packs" && <RoutingPacksTab />}
      {tab === "balancers" && <BalancersTab />}
      {tab === "outbounds" && <OutboundsTab />}
    </div>
  );
}

function RoutingPacksTab() {
  const { t } = useI18n();
  const toast = useToast();
  const confirm = useConfirm();
  const { can } = useAuth();
  const canWrite = can("inbound:write");
  const updateGeo = useUpdateGeo();
  const nodes = useNodes();

  const packs = useRoutingPacks();
  const defaultPack = useDefaultRoutingPack();
  const create = useCreateRoutingPack();
  const update = useUpdateRoutingPack();
  const del = useDeleteRoutingPack();
  const apply = useApplyRoutingPack();
  const setDefault = useSetDefaultRoutingPack();

  const [editorOpen, setEditorOpen] = useState(false);
  const [editing, setEditing] = useState<RoutingPack | null>(null);
  const [seed, setSeed] = useState<RoutingPack | null>(null);
  const [applyFor, setApplyFor] = useState<RoutingPack | null>(null);

  const list = packs.data?.packs ?? [];
  const activeDefault = defaultPack.data?.pack_id ?? "";

  function openCreate() {
    setEditing(null);
    setSeed(null);
    setEditorOpen(true);
  }
  function openEdit(p: RoutingPack) {
    setEditing(p);
    setSeed(p);
    setEditorOpen(true);
  }
  function openClone(p: RoutingPack) {
    const clone: RoutingPack = {
      ...p,
      id: "",
      builtin: false,
      name: `${p.name} (custom)`,
      rules: JSON.parse(JSON.stringify(p.rules ?? [])) as RoutingPack["rules"],
      outbounds: p.outbounds ? (JSON.parse(JSON.stringify(p.outbounds)) as unknown[]) : undefined,
    };
    setEditing(null);
    setSeed(clone);
    setEditorOpen(true);
  }

  async function remove(p: RoutingPack) {
    if (await confirm({ title: `${t("common.delete")} "${p.name}"?`, confirmLabel: t("common.delete"), destructive: true })) {
      await del.mutateAsync(p.id);
      toast.success(t("common.delete"));
    }
  }

  async function togglePack(p: RoutingPack, on: boolean) {
    if (!canWrite) return;
    if (on) {
      await setDefault.mutateAsync(p.id);
      toast.success(`${p.name} → default`);
    } else if (activeDefault === p.id) {
      await setDefault.mutateAsync("");
      toast.success("Default cleared");
    }
  }

  async function updateAllGeo() {
    const nodeList = nodes.data?.nodes ?? [];
    if (nodeList.length === 0) return;
    toast.info("Updating geo on all nodes…");
    for (const n of nodeList) {
      try {
        await updateGeo.mutateAsync(n.id);
      } catch {
        /* continue */
      }
    }
    toast.success("Geo update queued on all nodes");
  }

  return (
    <>
      <div className="rounded-xl border border-primary/25 bg-primary/5 px-4 py-3 flex flex-col sm:flex-row sm:items-center gap-3">
        <div className="flex items-start gap-3 flex-1 min-w-0">
          <div className="h-8 w-8 rounded-full bg-primary/15 flex items-center justify-center text-primary flex-shrink-0 mt-0.5">
            <CheckCircle2 size={16} />
          </div>
          <p className="text-xs text-fg-muted leading-relaxed">{t("routing.packsInfo")}</p>
        </div>
        {canWrite && (
          <Button variant="outline" size="sm" className="flex-shrink-0" onClick={updateAllGeo}>
            <RefreshCw size={14} /> {t("routing.updateGeo")}
          </Button>
        )}
      </div>

      <div className="flex justify-end gap-2">
        {canWrite && (
          <Button onClick={openCreate}>
            <Plus size={15} /> {t("common.add")}
          </Button>
        )}
        <Link
          to="/routing/node-rules"
          className="inline-flex items-center gap-1.5 h-9 px-3 rounded-lg border border-border/70 bg-surface-2/50 text-xs font-semibold text-fg hover:bg-surface-2 transition"
        >
          {t("routing.nodeRules")}
        </Link>
      </div>

      <div className="grid grid-cols-1 gap-4 lg:grid-cols-2">
        {list.map((p) => {
          const { icon: Icon, color } = packIcon(p);
          const isDefault = activeDefault === p.id;
          const matchers = packMatchers(p);
          return (
            <GlassCard key={p.id} hover={false} className="!p-5 space-y-4">
              <div className="flex items-start justify-between gap-3">
                <div className="flex items-start gap-3 min-w-0">
                  <div className={cn("h-10 w-10 rounded-xl flex items-center justify-center flex-shrink-0", color)}>
                    <Icon size={18} />
                  </div>
                  <div className="min-w-0">
                    <h3 className="text-sm font-bold text-fg truncate">{p.name}</h3>
                    <p className="text-[11px] text-fg-subtle mt-0.5">
                      {matchers.toLocaleString()} {t("routing.matchers")}
                    </p>
                  </div>
                </div>
                {canWrite && (
                  <Switch
                    checked={isDefault}
                    onCheckedChange={(v) => togglePack(p, v)}
                  />
                )}
              </div>
              {p.description && <p className="text-xs text-fg-muted leading-relaxed">{p.description}</p>}
              <div className="flex flex-wrap items-center justify-between gap-2 pt-1 border-t border-border/40">
                <StatusBadge
                  status={isDefault ? "active" : "inactive"}
                  label={isDefault ? t("routing.activeAllNodes") : t("routing.notDefault")}
                  pulse={false}
                />
                <div className="flex flex-wrap items-center gap-1">
                  <Button variant="ghost" size="sm" onClick={() => setApplyFor(p)}>
                    <Server size={13} /> Apply
                  </Button>
                  {!p.builtin && canWrite && (
                    <>
                      <Button variant="ghost" size="sm" onClick={() => openEdit(p)}><Pencil size={13} /></Button>
                      <Button variant="ghost" size="sm" className="text-danger" onClick={() => remove(p)}><Trash2 size={13} /></Button>
                    </>
                  )}
                  {p.builtin && canWrite && (
                    <Button variant="ghost" size="sm" onClick={() => openClone(p)}><Copy size={13} /></Button>
                  )}
                  <button
                    type="button"
                    onClick={() => (p.builtin ? openClone(p) : openEdit(p))}
                    className="text-xs text-primary hover:underline inline-flex items-center gap-1 ms-1"
                  >
                    {t("routing.configureRules")} <ArrowRight size={12} />
                  </button>
                </div>
              </div>
            </GlassCard>
          );
        })}
        {list.length === 0 && !packs.isLoading && (
          <p className="col-span-full py-12 text-center text-sm text-fg-muted">{t("common.none")}</p>
        )}
      </div>

      <PackEditor
        open={editorOpen}
        seed={seed}
        isEdit={!!editing}
        onClose={() => setEditorOpen(false)}
        onSubmit={async (body: RoutingPackBody) => {
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
    </>
  );
}

function BalancersTab() {
  const { t } = useI18n();
  const toast = useToast();
  const confirm = useConfirm();
  const { can } = useAuth();
  const canWrite = can("inbound:write");

  const fleet = useBalancersFleet();
  const create = balancerHooks.useCreate();
  const update = balancerHooks.useUpdate();
  const del = balancerHooks.useDelete();

  const [open, setOpen] = useState(false);
  const [editing, setEditing] = useState<BalancerFleetRow | null>(null);
  const [node, setNode] = useState("");

  const [f, setF] = useState({ tag: "", selectors: "", strategy: "leastPing", probe_url: "https://www.gstatic.com/generate_204", probe_interval: "10s" });
  const set = (k: string) => (e: React.ChangeEvent<HTMLInputElement | HTMLSelectElement>) => setF((s) => ({ ...s, [k]: e.target.value }));
  const observe = f.strategy === "leastPing" || f.strategy === "leastLoad";

  const balancers = fleet.data?.balancers ?? [];

  async function submit(e: React.FormEvent) {
    e.preventDefault();
    const body = {
      node_id: node,
      tag: f.tag,
      selectors: csv(f.selectors),
      strategy: f.strategy,
      observe,
      probe_url: f.probe_url,
      probe_interval: f.probe_interval,
      enabled: true,
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
    setF({ tag: "", selectors: "", strategy: "leastPing", probe_url: "https://www.gstatic.com/generate_204", probe_interval: "10s" });
  }

  function edit(b: BalancerFleetRow) {
    setNode(b.node_id);
    setF({
      tag: b.tag,
      selectors: b.selectors.join(", "),
      strategy: b.strategy,
      probe_url: b.probe_url || "https://www.gstatic.com/generate_204",
      probe_interval: b.probe_interval || "10s",
    });
    setEditing(b);
    setOpen(true);
  }

  async function remove(b: BalancerFleetRow) {
    if (await confirm({ title: `${t("common.delete")} ${b.tag}?`, confirmLabel: t("common.delete"), destructive: true })) {
      await del.mutateAsync(b.id);
      toast.success(t("common.delete"));
    }
  }

  return (
    <>
      <p className="text-xs text-fg-muted">{t("routing.balancersInfo")}</p>

      {canWrite && (
        <div className="flex justify-end">
          <Button onClick={() => { setEditing(null); setOpen(true); }}>
            <Plus size={15} /> {t("common.add")}
          </Button>
        </div>
      )}

      {fleet.isLoading && <div className="text-sm text-fg-muted text-center py-8">{t("common.loading")}</div>}

      <div className="grid grid-cols-1 gap-4 lg:grid-cols-2">
        {balancers.map((b) => (
          <GlassCard key={b.id} hover={false} className="!p-5 space-y-3">
            <div className="flex items-start justify-between gap-2">
              <div className="min-w-0">
                <p className="font-mono text-sm font-bold text-fg truncate">{b.tag}</p>
                <p className="text-[10px] text-primary font-semibold mt-1 uppercase tracking-wide">
                  {t("routing.strategy")}: {formatStrategy(b.strategy)}
                </p>
              </div>
              <StatusBadge
                status={b.enabled && (b.observe || b.strategy === "leastPing" || b.strategy === "leastLoad") ? "active" : "inactive"}
                label={b.enabled ? "PROBING ACTIVE" : "OFFLINE"}
                pulse={b.enabled}
              />
            </div>
            <p className="text-xs text-fg-muted">{b.node_name}</p>
            <div className="space-y-1.5 text-xs border-t border-border/40 pt-3">
              <div className="flex justify-between gap-2">
                <span className="text-fg-subtle">{t("routing.selectors")}</span>
                <span className="text-fg font-mono text-[11px] text-end truncate max-w-[60%]">{b.selectors.join(", ") || "—"}</span>
              </div>
              <div className="flex justify-between gap-2">
                <span className="text-fg-subtle">{t("routing.probeInterval")}</span>
                <span className="text-fg tabular-nums">{b.probe_interval || "10s"}</span>
              </div>
            </div>
            {canWrite && (
              <div className="flex justify-end gap-1 pt-1">
                <Button variant="ghost" size="sm" onClick={() => edit(b)}><Pencil size={14} /></Button>
                <Button variant="ghost" size="sm" className="text-danger" onClick={() => remove(b)}><Trash2 size={14} /></Button>
              </div>
            )}
          </GlassCard>
        ))}
        {balancers.length === 0 && !fleet.isLoading && (
          <p className="col-span-full py-12 text-center text-sm text-fg-muted">{t("common.none")}</p>
        )}
      </div>

      <Modal open={open} onClose={() => { setOpen(false); setEditing(null); }} title={editing ? `Edit ${editing.tag}` : t("nav.balancers")}>
        <form onSubmit={submit} className="space-y-3">
          {!editing && <NodePicker value={node} onChange={setNode} />}
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
            <Button type="submit" disabled={create.isPending || update.isPending || (!editing && !node)}>
              {editing ? t("common.save") : t("common.create")}
            </Button>
          </div>
        </form>
      </Modal>
    </>
  );
}