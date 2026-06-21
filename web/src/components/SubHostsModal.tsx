import { useState } from "react";
import {
  useSubHosts,
  useCreateSubHost,
  useUpdateSubHost,
  useDeleteSubHost,
  useReorderSubHosts,
  type Inbound,
  type SubHost,
  type SubHostBody,
  type HostSecurity,
} from "@/api/hooks";
import { useI18n } from "@/i18n/i18n";
import { Badge, Button, Input, Select } from "./ui";
import { Modal } from "./Modal";
import { useToast } from "./toast";

// Common uTLS fingerprints. "" = no explicit fingerprint (client default).
const FINGERPRINTS = ["", "chrome", "firefox", "safari", "ios", "android", "edge", "random", "randomized"];
const SECURITIES: HostSecurity[] = ["inbound_default", "none", "tls", "reality"];

const blank = {
  editId: "",
  remark: "",
  address: "",
  port: "",
  sni: "",
  host: "",
  path: "",
  alpn: "",
  fingerprint: "",
  security: "inbound_default" as HostSecurity,
  allow_insecure: false,
  mux_enable: false,
  fragment: "",
  enabled: true,
};

// securityColor maps a host's security mode onto a Badge color.
function securityColor(s: HostSecurity): string {
  switch (s) {
    case "reality":
      return "on_hold";
    case "tls":
      return "active";
    case "none":
      return "disabled";
    default:
      return "muted";
  }
}

export function SubHostsModal({ inbound, onClose }: { inbound: Inbound | null; onClose: () => void }) {
  const { t } = useI18n();
  const inboundId = inbound?.id ?? null;
  const list = useSubHosts(inboundId);
  const create = useCreateSubHost(inboundId);
  const update = useUpdateSubHost(inboundId);
  const del = useDeleteSubHost(inboundId);
  const reorder = useReorderSubHosts(inboundId);
  const toast = useToast();
  const [f, setF] = useState({ ...blank });

  if (!inbound) return null;
  const editing = f.editId !== "";
  const hosts = list.data?.hosts ?? [];

  const setText = (k: keyof typeof f) => (e: React.ChangeEvent<HTMLInputElement | HTMLSelectElement>) =>
    setF((s) => ({ ...s, [k]: e.target.value }));
  const setBool = (k: keyof typeof f) => (e: React.ChangeEvent<HTMLInputElement>) =>
    setF((s) => ({ ...s, [k]: e.target.checked }));

  function resetForm() {
    setF({ ...blank });
  }

  function startEdit(h: SubHost) {
    setF({
      editId: h.id,
      remark: h.remark,
      address: h.address,
      port: h.port == null ? "" : String(h.port),
      sni: h.sni,
      host: h.host,
      path: h.path,
      alpn: h.alpn,
      fingerprint: h.fingerprint,
      security: h.security,
      allow_insecure: h.allow_insecure,
      mux_enable: h.mux_enable,
      fragment: h.fragment,
      enabled: h.enabled,
    });
  }

  function buildBody(): SubHostBody {
    const trimmedPort = f.port.trim();
    return {
      remark: f.remark.trim(),
      address: f.address.trim(),
      port: trimmedPort === "" ? null : Number(trimmedPort),
      sni: f.sni.trim(),
      host: f.host.trim(),
      path: f.path.trim(),
      alpn: f.alpn.trim(),
      fingerprint: f.fingerprint,
      security: f.security,
      allow_insecure: f.allow_insecure,
      mux_enable: f.mux_enable,
      fragment: f.fragment.trim(),
      enabled: f.enabled,
    };
  }

  async function submit(e: React.FormEvent) {
    e.preventDefault();
    if (f.remark.trim() === "") {
      toast.error(t("hosts.remarkRequired"));
      return;
    }
    const body = buildBody();
    try {
      if (editing) {
        await update.mutateAsync({ id: f.editId, body });
      } else {
        await create.mutateAsync(body);
      }
      toast.success(t("hosts.saved"));
      resetForm();
    } catch {
      toast.error(t("hosts.saveFailed"));
    }
  }

  async function toggleEnable(h: SubHost) {
    try {
      await update.mutateAsync({
        id: h.id,
        body: {
          remark: h.remark,
          address: h.address,
          port: h.port,
          sni: h.sni,
          host: h.host,
          path: h.path,
          alpn: h.alpn,
          fingerprint: h.fingerprint,
          security: h.security,
          allow_insecure: h.allow_insecure,
          mux_enable: h.mux_enable,
          fragment: h.fragment,
          enabled: !h.enabled,
        },
      });
    } catch {
      toast.error(t("hosts.saveFailed"));
    }
  }

  async function remove(h: SubHost) {
    try {
      await del.mutateAsync(h.id);
      toast.success(t("hosts.deleted"));
      if (f.editId === h.id) resetForm();
    } catch {
      toast.error(t("hosts.deleteFailed"));
    }
  }

  // move swaps a host with its neighbour and persists the new order via the
  // reorder endpoint (ids in their new priority order).
  async function move(index: number, dir: -1 | 1) {
    const target = index + dir;
    if (target < 0 || target >= hosts.length) return;
    const ids = hosts.map((h) => h.id);
    [ids[index], ids[target]] = [ids[target], ids[index]];
    try {
      await reorder.mutateAsync(ids);
    } catch {
      toast.error(t("hosts.reorderFailed"));
    }
  }

  return (
    <Modal open={!!inbound} onClose={onClose} title={`${t("hosts.title")} · ${inbound.tag}`} className="max-w-lg">
      <p className="mb-3 text-xs text-fg-muted">{t("hosts.help")}</p>

      <div className="space-y-2">
        {hosts.map((h, i) => (
          <div key={h.id} className="flex items-center justify-between rounded-md border px-3 py-2 text-sm">
            <div className="flex min-w-0 items-center gap-2">
              <span
                className={`h-2 w-2 shrink-0 rounded-full ${h.enabled ? "bg-success" : "bg-fg-subtle"}`}
                title={h.enabled ? t("common.enabled") : t("common.disabled")}
              />
              <span className="truncate font-medium">{h.remark || "—"}</span>
              <span className="truncate text-xs text-muted-foreground" dir="ltr">
                {h.address || inbound.tag}
                {h.port != null ? `:${h.port}` : ""}
              </span>
              <Badge color={securityColor(h.security)}>{h.security === "inbound_default" ? "default" : h.security}</Badge>
            </div>
            <div className="flex shrink-0 gap-1">
              <Button variant="ghost" size="sm" onClick={() => move(i, -1)} disabled={i === 0 || reorder.isPending} title={t("hosts.moveUp")}>
                ↑
              </Button>
              <Button variant="ghost" size="sm" onClick={() => move(i, 1)} disabled={i === hosts.length - 1 || reorder.isPending} title={t("hosts.moveDown")}>
                ↓
              </Button>
              <Button variant="ghost" size="sm" onClick={() => toggleEnable(h)} title={h.enabled ? t("common.disabled") : t("common.enabled")}>
                {h.enabled ? "🟢" : "⏸"}
              </Button>
              <Button variant="ghost" size="sm" onClick={() => startEdit(h)}>{t("common.edit")}</Button>
              <Button variant="ghost" size="sm" className="text-destructive" onClick={() => remove(h)}>{t("common.delete")}</Button>
            </div>
          </div>
        ))}
        {hosts.length === 0 && <p className="py-2 text-sm text-muted-foreground">{t("hosts.empty")}</p>}
      </div>

      <form onSubmit={submit} className="mt-4 space-y-3 border-t border-border/60 pt-3">
        <div className="flex items-center justify-between">
          <p className="text-xs font-medium text-muted-foreground">{editing ? t("hosts.editTitle") : t("hosts.add")}</p>
          {editing && (
            <button type="button" className="text-xs text-muted-foreground underline" onClick={resetForm}>
              {t("hosts.cancelEdit")}
            </button>
          )}
        </div>

        <Input placeholder={t("hosts.remarkPlaceholder")} value={f.remark} onChange={setText("remark")} required />
        <Input placeholder={t("hosts.addressPlaceholder")} value={f.address} onChange={setText("address")} />
        <div className="grid grid-cols-2 gap-2">
          <Input placeholder={t("hosts.portPlaceholder")} value={f.port} onChange={setText("port")} inputMode="numeric" />
          <Input placeholder={t("hosts.sni")} value={f.sni} onChange={setText("sni")} />
        </div>
        <div className="grid grid-cols-2 gap-2">
          <Input placeholder={t("hosts.host")} value={f.host} onChange={setText("host")} />
          <Input placeholder={t("hosts.path")} value={f.path} onChange={setText("path")} />
        </div>
        <div className="grid grid-cols-2 gap-2">
          <Input placeholder={t("hosts.alpn")} value={f.alpn} onChange={setText("alpn")} />
          <div>
            <label className="mb-1 block text-[10px] font-medium text-fg-muted">{t("hosts.fingerprint")}</label>
            <Select value={f.fingerprint} onChange={setText("fingerprint")}>
              {FINGERPRINTS.map((fp) => (
                <option key={fp} value={fp}>{fp === "" ? t("common.none") : fp}</option>
              ))}
            </Select>
          </div>
        </div>
        <div className="grid grid-cols-2 gap-2">
          <div>
            <label className="mb-1 block text-[10px] font-medium text-fg-muted">{t("hosts.security")}</label>
            <Select value={f.security} onChange={setText("security")}>
              {SECURITIES.map((s) => (
                <option key={s} value={s}>{s === "inbound_default" ? t("hosts.securityInboundDefault") : s}</option>
              ))}
            </Select>
          </div>
          <Input placeholder={t("hosts.fragmentPlaceholder")} value={f.fragment} onChange={setText("fragment")} />
        </div>

        <label className="flex items-center gap-2 text-xs text-fg-muted">
          <input type="checkbox" className="h-4 w-4 accent-primary" checked={f.allow_insecure} onChange={setBool("allow_insecure")} />
          {t("hosts.allowInsecure")}
        </label>
        <label className="flex items-center gap-2 text-xs text-fg-muted">
          <input type="checkbox" className="h-4 w-4 accent-primary" checked={f.mux_enable} onChange={setBool("mux_enable")} />
          {t("hosts.muxEnable")}
        </label>

        <div className="flex justify-end">
          <Button type="submit" disabled={create.isPending || update.isPending}>
            {editing ? t("hosts.saveChanges") : t("hosts.add")}
          </Button>
        </div>
      </form>
    </Modal>
  );
}
