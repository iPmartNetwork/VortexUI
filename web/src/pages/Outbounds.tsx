import { useState } from "react";
import { Trash2, Pencil } from "lucide-react";
import { outboundHooks } from "@/api/policy-hooks";
import type { Outbound } from "@/api/types";
import { Badge, Button, Card, Input, PageHeader, Select } from "@/components/ui";
import { Modal } from "@/components/Modal";
import { NodePicker } from "@/components/NodePicker";
import { useConfirm } from "@/components/confirm";
import { useToast } from "@/components/toast";
import { useI18n } from "@/i18n/i18n";
import { JsonCodeEditor } from "@/components/JsonCodeEditor";
import { DEFAULT_OUTBOUND_TEMPLATE, parseShareLink } from "@/lib/outbound-uri";

const PROTOCOLS = ["freedom", "blackhole", "dns", "vless", "vmess", "trojan", "shadowsocks", "socks", "http"];

export function Outbounds() {
  const { t } = useI18n();
  const [node, setNode] = useState("");
  const [open, setOpen] = useState(false);
  const [editing, setEditing] = useState<Outbound | null>(null);
  const list = outboundHooks.useList(node || null);
  const create = outboundHooks.useCreate();
  const update = outboundHooks.useUpdate();
  const del = outboundHooks.useDelete();
  const confirm = useConfirm();
  const toast = useToast();

  const [f, setF] = useState({ tag: "", protocol: "freedom", address: "", port: "", uuid: "", password: "", method: "aes-128-gcm", security: "none", sni: "" });
  const set = (k: string) => (e: React.ChangeEvent<HTMLInputElement | HTMLSelectElement>) => setF((s) => ({ ...s, [k]: e.target.value }));
  const proxy = !["freedom", "blackhole", "dns"].includes(f.protocol);

  const [tab, setTab] = useState<"basics" | "json">("basics");
  const [jsonText, setJsonText] = useState("");
  const [jsonErr, setJsonErr] = useState("");
  const [importUri, setImportUri] = useState("");

  const defaultTemplate = () => JSON.stringify(DEFAULT_OUTBOUND_TEMPLATE, null, 2);

  function gotoJson() {
    setTab("json");
    if (!jsonText) setJsonText(defaultTemplate());
  }

  // importLink converts a pasted share link into outbound JSON in the editor.
  function importLink() {
    setJsonErr("");
    if (!importUri.trim()) return;
    try {
      const obj = parseShareLink(importUri);
      setJsonText(JSON.stringify(obj, null, 2));
      setImportUri("");
    } catch {
      setJsonErr(t("outbounds.importFailed"));
    }
  }

  function shortId() {
    return Math.random().toString(16).slice(2, 6);
  }

  async function submitJSON() {
    setJsonErr("");
    let parsed: Record<string, unknown>;
    try {
      parsed = JSON.parse(jsonText);
    } catch {
      setJsonErr(t("outbounds.invalidJson"));
      return;
    }
    const protocol = String(parsed.protocol ?? "");
    if (!protocol) {
      setJsonErr(t("outbounds.jsonNeedsProtocol"));
      return;
    }
    // 3x-ui templates omit the tag; generate one when absent.
    const tag = String(parsed.tag ?? "") || `out-${protocol}-${shortId()}`;
    parsed.tag = tag;
    try {
      if (editing) {
        await update.mutateAsync({ id: editing.id, body: { tag, protocol, raw: parsed, enabled: true } });
        toast.success(`${editing.tag} updated`);
        setEditing(null);
      } else {
        await create.mutateAsync({ node_id: node, tag, protocol, raw: parsed, enabled: true });
        toast.success(`${tag} ✓`);
      }
      closeModal();
    } catch {
      setJsonErr(t("outbounds.saveFailed"));
    }
  }

  function closeModal() {
    setOpen(false);
    setEditing(null);
    setTab("basics");
    setJsonText("");
    setJsonErr("");
    setImportUri("");
    setF({ tag: "", protocol: "freedom", address: "", port: "", uuid: "", password: "", method: "aes-128-gcm", security: "none", sni: "" });
  }

  async function submit(e: React.FormEvent) {
    e.preventDefault();
    if (editing) {
      await update.mutateAsync({
        id: editing.id,
        body: {
          protocol: f.protocol, address: f.address, port: Number(f.port) || 0,
          uuid: f.uuid, password: f.password, method: f.method, security: f.security,
          sni: f.sni, enabled: true,
        },
      });
      toast.success(`${editing.tag} updated`);
      setEditing(null);
    } else {
      await create.mutateAsync({
        node_id: node, tag: f.tag, protocol: f.protocol, address: f.address, port: Number(f.port) || 0,
        uuid: f.uuid, password: f.password, method: f.method, security: f.security,
        sni: f.sni, enabled: true,
      });
      toast.success(`${f.tag} ✓`);
    }
    closeModal();
  }

  function edit(o: Outbound) {
    setF({ tag: o.tag, protocol: o.protocol, address: o.address, port: String(o.port || ""), uuid: o.uuid, password: o.password, method: o.method || "aes-128-gcm", security: o.security || "none", sni: o.sni });
    setJsonText(o.raw ? JSON.stringify(o.raw, null, 2) : "");
    setJsonErr("");
    setTab("basics");
    setEditing(o);
    setOpen(true);
  }

  async function remove(o: Outbound) {
    if (await confirm({ title: `${t("common.delete")} ${o.tag}?`, confirmLabel: t("common.delete"), destructive: true })) {
      await del.mutateAsync(o.id);
      toast.success(t("common.delete"));
    }
  }

  return (
    <div>
      <PageHeader title={t("nav.outbounds")} subtitle="Egress handlers per node">
        <NodePicker value={node} onChange={setNode} />
        <Button onClick={() => setOpen(true)}>{t("common.add")}</Button>
      </PageHeader>

      <Card className="p-0">
        <div className="divide-y divide-white/[0.05]">
          {list.data?.outbounds?.map((o) => (
            <div key={o.id} className="flex items-center justify-between px-5 py-3">
              <div className="flex items-center gap-3 text-sm">
                <span className="font-medium">{o.tag}</span>
                <Badge>{o.protocol}</Badge>
                {o.address && <span className="text-xs text-fg-muted">{o.address}:{o.port}</span>}
              </div>
              <Button variant="ghost" size="sm" onClick={() => edit(o)} title="Edit">
                <Pencil size={15} />
              </Button>
              <Button variant="ghost" size="sm" className="text-danger" onClick={() => remove(o)}>
                <Trash2 size={15} />
              </Button>
            </div>
          ))}
          {list.data?.outbounds?.length === 0 && <p className="px-5 py-8 text-center text-sm text-fg-muted">{t("common.none")}</p>}
        </div>
      </Card>

      <Modal open={open} onClose={closeModal} title={editing ? `Edit ${editing.tag}` : t("nav.outbounds")}>
        {/* Tabs */}
        <div className="mb-4 flex gap-5 border-b border-border/60 text-sm">
          {(["basics", "json"] as const).map((tk) => (
            <button
              key={tk}
              type="button"
              onClick={() => (tk === "json" ? gotoJson() : setTab("basics"))}
              className={`-mb-px border-b-2 pb-2 font-medium transition ${tab === tk ? "border-primary text-primary" : "border-transparent text-fg-muted hover:text-fg"}`}
            >
              {tk === "basics" ? t("outbounds.basics") : "JSON"}
            </button>
          ))}
        </div>

        {tab === "basics" ? (
          <form onSubmit={submit} className="space-y-3">
            <div className="grid grid-cols-2 gap-2">
              <Input placeholder="Tag" value={f.tag} onChange={set("tag")} required />
              <Select value={f.protocol} onChange={set("protocol")}>
                {PROTOCOLS.map((p) => <option key={p} value={p}>{p}</option>)}
              </Select>
            </div>
            {proxy && (
              <>
                <div className="grid grid-cols-2 gap-2">
                  <Input placeholder="Address" value={f.address} onChange={set("address")} />
                  <Input placeholder="Port" value={f.port} onChange={set("port")} inputMode="numeric" />
                </div>
                {(f.protocol === "vless" || f.protocol === "vmess") && <Input placeholder="UUID" value={f.uuid} onChange={set("uuid")} />}
                {["trojan", "shadowsocks", "socks", "http"].includes(f.protocol) && <Input placeholder="Password" value={f.password} onChange={set("password")} />}
                {f.protocol === "shadowsocks" && <Input placeholder="Method" value={f.method} onChange={set("method")} />}
                <div className="grid grid-cols-2 gap-2">
                  <Select value={f.security} onChange={set("security")}>
                    <option value="none">none</option><option value="tls">tls</option><option value="reality">reality</option>
                  </Select>
                  <Input placeholder="SNI" value={f.sni} onChange={set("sni")} />
                </div>
              </>
            )}
            <div className="flex justify-end gap-2 pt-1">
              <Button type="button" variant="ghost" onClick={closeModal}>{t("common.cancel")}</Button>
              <Button type="submit" disabled={create.isPending || update.isPending}>{editing ? t("common.save") : t("common.create")}</Button>
            </div>
          </form>
        ) : (
          <div className="space-y-3">
            {/* Share-link import bar */}
            <div className="flex gap-2">
              <Input
                value={importUri}
                onChange={(e) => setImportUri(e.target.value)}
                onKeyDown={(e) => { if (e.key === "Enter") { e.preventDefault(); importLink(); } }}
                placeholder="vmess:// vless:// trojan:// ss:// hysteria2:// wireguard://"
                className="flex-1 font-mono text-xs"
                dir="ltr"
              />
              <Button type="button" onClick={importLink}>{t("outbounds.import")}</Button>
            </div>

            <JsonCodeEditor value={jsonText} onChange={setJsonText} rows={16} />

            {jsonErr && <p className="text-sm text-danger">{jsonErr}</p>}
            <div className="flex justify-end gap-2 pt-1">
              <Button type="button" variant="ghost" onClick={closeModal}>{t("common.cancel")}</Button>
              <Button type="button" onClick={submitJSON} disabled={create.isPending || update.isPending}>{editing ? t("common.save") : t("common.create")}</Button>
            </div>
          </div>
        )}
      </Modal>
    </div>
  );
}
