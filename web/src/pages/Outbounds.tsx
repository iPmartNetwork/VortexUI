import { useState } from "react";
import { Trash2 } from "lucide-react";
import { outboundHooks } from "@/api/policy-hooks";
import type { Outbound } from "@/api/types";
import { Badge, Button, Card, Input, PageHeader, Select } from "@/components/ui";
import { Modal } from "@/components/Modal";
import { NodePicker } from "@/components/NodePicker";
import { useConfirm } from "@/components/confirm";
import { useToast } from "@/components/toast";
import { useI18n } from "@/i18n/i18n";

const PROTOCOLS = ["freedom", "blackhole", "dns", "vless", "vmess", "trojan", "shadowsocks", "socks", "http"];

export function Outbounds() {
  const { t } = useI18n();
  const [node, setNode] = useState("");
  const [open, setOpen] = useState(false);
  const list = outboundHooks.useList(node || null);
  const create = outboundHooks.useCreate();
  const del = outboundHooks.useDelete();
  const confirm = useConfirm();
  const toast = useToast();

  const [f, setF] = useState({ tag: "", protocol: "freedom", address: "", port: "", uuid: "", password: "", method: "aes-128-gcm", security: "none", sni: "" });
  const set = (k: string) => (e: React.ChangeEvent<HTMLInputElement | HTMLSelectElement>) => setF((s) => ({ ...s, [k]: e.target.value }));
  const proxy = !["freedom", "blackhole", "dns"].includes(f.protocol);

  async function submit(e: React.FormEvent) {
    e.preventDefault();
    await create.mutateAsync({
      node_id: node, tag: f.tag, protocol: f.protocol, address: f.address, port: Number(f.port) || 0,
      uuid: f.uuid, password: f.password, method: f.method, security: f.security,
      sni: f.sni, enabled: true,
    });
    toast.success(`${f.tag} ✓`);
    setOpen(false);
    setF({ tag: "", protocol: "freedom", address: "", port: "", uuid: "", password: "", method: "aes-128-gcm", security: "none", sni: "" });
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
              <Button variant="ghost" size="sm" className="text-danger" onClick={() => remove(o)}>
                <Trash2 size={15} />
              </Button>
            </div>
          ))}
          {list.data?.outbounds?.length === 0 && <p className="px-5 py-8 text-center text-sm text-fg-muted">{t("common.none")}</p>}
        </div>
      </Card>

      <Modal open={open} onClose={() => setOpen(false)} title={t("nav.outbounds")}>
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
            <Button type="button" variant="ghost" onClick={() => setOpen(false)}>{t("common.cancel")}</Button>
            <Button type="submit" disabled={create.isPending}>{t("common.create")}</Button>
          </div>
        </form>
      </Modal>
    </div>
  );
}
