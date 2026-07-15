import { useEffect, useState } from "react";
import { useCapabilities, useCreateInbound, useDeleteInbound, useNodeInbounds, useUpdateInbound, inboundToUpdateInput, type Inbound } from "@/api/hooks";
import { useReality } from "@/api/policy-hooks";
import type { Node } from "@/api/types";
import { Badge, Button, Input, Select } from "./ui";
import { Modal } from "./Modal";
import { CopyField } from "./CopyField";
import { JsonCodeEditor } from "./JsonCodeEditor";
import { SubHostsModal } from "./SubHostsModal";
import { useToast } from "./toast";
import { useI18n } from "@/i18n/i18n";
import { useAuth } from "@/auth/auth";
import type { CoreType } from "@/lib/coreTypes";
import { isMultiCore, resolveInboundCore } from "@/lib/coreTypes";
import { CoreBadge } from "@/components/veltrix/CoreBadge";

// Static fallbacks used only until the per-core capability matrix
// (GET /api/capabilities) loads, so the form still works before the fetch
// resolves. Once `caps` is available the options are filtered per the node's core.
const PROTOCOLS = ["vless", "vmess", "trojan", "shadowsocks", "hysteria2", "tuic", "wireguard", "socks", "http", "naive", "dokodemo"];
const NETWORKS = ["tcp", "ws", "grpc", "httpupgrade", "http", "h2", "xhttp", "kcp", "quic", "udp"];
const SECURITIES = ["none", "tls", "reality"];

// UDP-native protocol fallback (used until caps load). Authoritative list comes
// from cap.udp_native per core.
const UDP_PROTOCOLS = ["hysteria2", "tuic", "wireguard", "hysteria"];

// No-transport protocol fallback (used until caps load). These protocols carry
// no stream transport, so the network select is hidden. Authoritative list comes
// from cap.no_transport per core.
const NO_TRANSPORT_PROTOCOLS = ["hysteria2", "tuic", "wireguard", "socks", "http", "naive", "dokodemo", "hysteria", "shadowtls", "anytls"];

/** List-row label: WireGuard/Hysteria/TUIC are UDP-native — never tcp/ws + tls. */
function inboundTransportLabel(
  ib: { protocol: string; network: string; security: string },
  udpNative: string[],
  noTransport: string[],
): string {
  if (udpNative.includes(ib.protocol)) return "udp";
  if (noTransport.includes(ib.protocol) && !ib.network) return "—";
  const net = ib.network || "tcp";
  const sec = ib.security || "none";
  return `${net}/${sec}`;
}

// randomPort picks a high port (10000–60000) so new inbounds default to a free,
// non-conflicting port. The admin can still type any port.
const randomPort = () => String(10000 + Math.floor(Math.random() * 50000));

const newBlank = () => ({ editId: "", tag: "", core: "" as CoreType | "", protocol: "vless", port: randomPort(), network: "tcp", security: "tls", sni: "", path: "", host: "", flow: "", geoAllow: "", wgPrivateKey: "", wgSubnet: "", wgMtu: "" });
const blank = newBlank();

const DEFAULT_INBOUND_TEMPLATE = {
  tag: "inbound-443",
  protocol: "vless",
  port: 443,
  settings: { clients: [], decryption: "none" },
  streamSettings: {
    network: "tcp",
    tcpSettings: { header: { type: "none" } },
    security: "none",
  },
};

export function NodeInboundsModal({
  node,
  onClose,
  initialEdit,
}: {
  node: Node | null;
  onClose: () => void;
  initialEdit?: Inbound | null;
}) {
  const { can } = useAuth();
  const canWrite = can("inbound:write");
  const list = useNodeInbounds(node?.id ?? null);
  const caps = useCapabilities().data;
  const create = useCreateInbound();
  const update = useUpdateInbound();
  const del = useDeleteInbound();
  const toast = useToast();
  const { t } = useI18n();
  const [f, setF] = useState({ ...blank });
  const [editSnapshot, setEditSnapshot] = useState<Inbound | null>(null);
  const [hostsFor, setHostsFor] = useState<Inbound | null>(null);
  const [realityKeys, setRealityKeys] = useState<{ private_key: string; public_key: string; short_id: string } | null>(null);
  const [tab, setTab] = useState<"basics" | "json">("basics");
  const [jsonText, setJsonText] = useState("");
  const [jsonErr, setJsonErr] = useState("");

  // resetForm clears the basics form back to a blank inbound, including any
  // generated REALITY keys, so a previous edit/create can't leak into the next.
  const resetForm = () => {
    setF(newBlank());
    setRealityKeys(null);
    setEditSnapshot(null);
  };

  // Per-core capability for the form's effective engine, with static fallbacks.
  const nodeMulti = node ? isMultiCore(node) : false;
  const formCore = node ? resolveInboundCore(node, f.core) : "xray";
  const cap = caps?.[formCore === "singbox" ? "singbox" : "xray"];
  const protocols = cap?.protocols ?? PROTOCOLS;
  const networks = cap?.transports ?? NETWORKS;
  const securities = cap?.securities ?? SECURITIES;
  const noTransport = [...new Set([...(cap?.udp_native ?? UDP_PROTOCOLS), ...(cap?.no_transport ?? NO_TRANSPORT_PROTOCOLS)])];
  const udpNative = cap?.udp_native ?? UDP_PROTOCOLS;
  const isNoTransport = noTransport.includes(f.protocol);
  const securitiesFor = (proto: string) => cap?.protocol_securities?.[proto] ?? securities;
  const editing = f.editId !== "";
  const set = (k: keyof typeof f) => (e: React.ChangeEvent<HTMLInputElement | HTMLSelectElement>) =>
    setF((s) => {
      const val = e.target.value;
      if (k === "protocol") {
        const allowed = securitiesFor(val);
        const security = allowed.includes(s.security) ? s.security : (allowed[0] ?? s.security);
        if (noTransport.includes(val)) {
          return { ...s, protocol: val, network: "", security };
        }
        const network = s.network && networks.includes(s.network) ? s.network : (networks[0] ?? "tcp");
        return { ...s, protocol: val, network, security };
      }
      return { ...s, [k]: val };
    });

  function startEdit(ib: Inbound) {
    setEditSnapshot(ib);
    setRealityKeys(null);
    const wg = (ib.raw?.wireguard ?? {}) as Record<string, unknown>;
    setF({ editId: ib.id, tag: ib.tag, core: (ib.core ?? "") as CoreType | "", protocol: ib.protocol, port: String(ib.port), network: udpNative.includes(ib.protocol) || noTransport.includes(ib.protocol) ? "" : ib.network, security: ib.security, sni: (ib.sni ?? []).join(", "), path: ib.path ?? "", host: (ib.host ?? []).join(", "), flow: ib.flow ?? "", geoAllow: (ib.geo_policy?.allowed_countries ?? []).join(", "), wgPrivateKey: typeof wg.private_key === "string" ? wg.private_key : "", wgSubnet: typeof wg.subnet === "string" ? wg.subnet : "", wgMtu: typeof wg.mtu === "number" ? String(wg.mtu) : "" });
  }

  // When the capability matrix loads or the node's core changes, reconcile the
  // form so it can never submit a protocol/network/security the core rejects.
  useEffect(() => {
    if (!cap) return;
    setF((s) => {
      let next = s;
      if (!cap.protocols.includes(next.protocol)) {
        next = { ...next, protocol: cap.protocols[0] ?? next.protocol };
      }
      const noTransportSet = new Set([...cap.udp_native, ...cap.no_transport]);
      if (noTransportSet.has(next.protocol)) {
        if (next.network !== "") next = { ...next, network: "" };
      } else if (!cap.transports.includes(next.network)) {
        next = { ...next, network: cap.transports[0] ?? next.network };
      }
      const allowedSecurities = cap.protocol_securities?.[next.protocol] ?? cap.securities;
      if (!allowedSecurities.includes(next.security)) {
        next = { ...next, security: allowedSecurities[0] ?? next.security };
      }
      return next === s ? s : next;
    });
  }, [cap, formCore]);

  useEffect(() => {
    if (!node || !initialEdit || initialEdit.node_id !== node.id) return;
    startEdit(initialEdit);
  }, [node?.id, initialEdit?.id]);

  if (!node) return null;

  async function toggleEnable(ib: Inbound) {
    try {
      await update.mutateAsync({
        id: ib.id,
        input: inboundToUpdateInput(ib, { enabled: !ib.enabled }),
      });
      toast.success(`${ib.tag} ${ib.enabled ? "disabled" : "enabled"}`);
    } catch {
      toast.error("Toggle failed");
    }
  }

  async function submit(e: React.FormEvent) {
    e.preventDefault();
    if (!node) return;
    const sni = f.sni ? f.sni.split(",").map((s) => s.trim()) : [];
    const host = f.host ? f.host.split(",").map((s) => s.trim()).filter(Boolean) : [];
    const allowed = f.geoAllow ? f.geoAllow.split(",").map((s) => s.trim()).filter(Boolean) : [];
    const geo_policy = allowed.length > 0 ? { allowed_countries: allowed } : null;
    // When REALITY is selected and keys were generated in the form, send them in
    // raw.reality so the saved inbound uses exactly these keys (the public_key we
    // displayed). Without this the backend would auto-generate a different pair.
    let raw: Record<string, unknown> | undefined;
    if (f.security === "reality" && realityKeys) {
      raw = {
        reality: {
          private_key: realityKeys.private_key,
          public_key: realityKeys.public_key,
          short_ids: [realityKeys.short_id],
          ...(sni.length > 0 ? { server_names: sni, dest: `${sni[0]}:443` } : {}),
        },
      };
    }
    // WireGuard inbound: persist its native config under raw.wireguard using the
    // exact keys buildWireGuardEndpoint consumes (private_key, subnet, mtu). The
    // listen port comes from the inbound Port field, so it's not duplicated here.
    if (f.protocol === "wireguard") {
      const wg: Record<string, unknown> = {};
      if (f.wgPrivateKey.trim()) wg.private_key = f.wgPrivateKey.trim();
      if (f.wgSubnet.trim()) wg.subnet = f.wgSubnet.trim();
      if (f.wgMtu.trim()) wg.mtu = Number(f.wgMtu);
      if (Object.keys(wg).length > 0) raw = { ...(raw ?? {}), wireguard: wg };
    }
    const network = isNoTransport ? "" : f.network;
    const corePayload = f.core || undefined;
    try {
      if (editing && editSnapshot) {
        const mergedRaw = raw ?? editSnapshot.raw;
        await update.mutateAsync({
          id: f.editId,
          input: inboundToUpdateInput(editSnapshot, {
            port: Number(f.port),
            network,
            security: f.security,
            core: f.core,
            sni,
            path: f.path,
            host,
            flow: f.flow,
            geo_policy,
            ...(mergedRaw ? { raw: mergedRaw } : {}),
          }),
        });
        toast.success(`Updated ${f.tag}`);
      } else if (editing) {
        await update.mutateAsync({
          id: f.editId,
          input: {
            port: Number(f.port),
            network,
            security: f.security,
            sni,
            path: f.path,
            host,
            flow: f.flow,
            geo_policy,
            core: f.core,
            enabled: true,
            ...(raw ? { raw } : {}),
          },
        });
        toast.success(`Updated ${f.tag}`);
      } else {
        await create.mutateAsync({ node_id: node.id, tag: f.tag, core: corePayload, protocol: f.protocol, port: Number(f.port), network, security: f.security, sni, path: f.path, host, flow: f.flow, geo_policy, enabled: true, ...(raw ? { raw } : {}) });
        toast.success(`Added ${f.tag}`);
      }
      resetForm();
    } catch {
      toast.error("Save failed");
    }
  }

  function gotoJson() {
    setTab("json");
    if (!jsonText) setJsonText(JSON.stringify(DEFAULT_INBOUND_TEMPLATE, null, 2));
  }

  async function submitJSON() {
    setJsonErr("");
    if (!node) return;
    let parsed: Record<string, unknown>;
    try {
      parsed = JSON.parse(jsonText);
    } catch {
      setJsonErr("Invalid JSON");
      return;
    }
    const tag = String(parsed.tag ?? "");
    const protocol = String(parsed.protocol ?? "");
    const port = Number(parsed.port ?? 0);
    if (!tag || !protocol || !port) {
      setJsonErr("JSON must include tag, protocol and port");
      return;
    }
    try {
      await create.mutateAsync({ node_id: node.id, tag, protocol, port, raw: parsed, enabled: true });
      toast.success(`Added ${tag}`);
      setJsonText("");
      setTab("basics");
    } catch {
      setJsonErr("Save failed (tag/port taken?)");
    }
  }

  return (
    <>
    <Modal open={!!node} onClose={onClose} title={`Inbounds · ${node.name}`} className="max-w-xl">
      <div className="space-y-2">
        {list.data?.inbounds?.map((ib) => (
          <div key={ib.id} className="flex items-center justify-between gap-2 rounded-md border px-3 py-2 text-sm">
            <div className="flex items-center gap-2 min-w-0">
              <span className="font-medium truncate min-w-0">{ib.tag}</span>
              <Badge>{ib.protocol}</Badge>
              {nodeMulti && (
                <CoreBadge core={resolveInboundCore(node, ib.core)} className="scale-90" />
              )}
              <span className="truncate text-xs text-muted-foreground">:{ib.port} · {inboundTransportLabel(ib, udpNative, noTransport)}</span>
              <span className={`h-2 w-2 shrink-0 rounded-full ${ib.enabled ? "bg-success" : "bg-fg-subtle"}`} title={ib.enabled ? "Enabled" : "Disabled"} />
            </div>
            <div className="flex items-center gap-1 shrink-0">
              {canWrite && (
                <>
                  <Button variant="ghost" size="sm" onClick={() => toggleEnable(ib)} title={ib.enabled ? "Disable" : "Enable"}>
                    {ib.enabled ? "🟢" : "⏸"}
                  </Button>
                  <Button variant="ghost" size="sm" onClick={() => startEdit(ib)}>Edit</Button>
                </>
              )}
              <Button variant="ghost" size="sm" onClick={() => setHostsFor(ib)} title={t("hosts.title")}>{t("hosts.button")}</Button>
              {canWrite && (
                <Button variant="ghost" size="sm" className="text-destructive" onClick={() => del.mutate(ib.id)}>Remove</Button>
              )}
            </div>
          </div>
        ))}
        {list.data?.inbounds?.length === 0 && <p className="py-2 text-sm text-muted-foreground">No inbounds on this node yet.</p>}
      </div>

      {canWrite && (
      <div className="mt-4 flex gap-5 border-t border-border/60 pt-3 text-sm">
        {(["basics", "json"] as const).map((tk) => (
          <button
            key={tk}
            type="button"
            onClick={() => (tk === "json" ? gotoJson() : setTab("basics"))}
            className={`-mb-px border-b-2 pb-2 font-medium transition ${tab === tk ? "border-primary text-primary" : "border-transparent text-fg-muted hover:text-fg"}`}
          >
            {tk === "basics" ? "Basics" : "JSON"}
          </button>
        ))}
      </div>
      )}

      {canWrite && (tab === "json" ? (
        <div className="mt-3 space-y-3">
          <p className="text-xs text-fg-muted">Paste a full Xray/sing-box inbound object. tag, protocol and port are read from it; the whole object is stored as a raw override.</p>
          <JsonCodeEditor value={jsonText} onChange={setJsonText} rows={14} />
          {jsonErr && <p className="text-sm text-danger">{jsonErr}</p>}
          <div className="flex justify-end">
            <Button type="button" onClick={submitJSON} disabled={create.isPending}>Add inbound</Button>
          </div>
        </div>
      ) : (
      <form onSubmit={submit} className="mt-3 space-y-3">
        <div className="flex items-center justify-between">
          <p className="text-xs font-medium text-muted-foreground">{editing ? `Edit ${f.tag}` : "Add inbound"}</p>
          {editing && (
            <button type="button" className="text-xs text-muted-foreground underline" onClick={resetForm}>
              cancel edit
            </button>
          )}
        </div>
        <div className="grid grid-cols-2 gap-2">
          <Input placeholder="Tag" value={f.tag} onChange={set("tag")} required disabled={editing} />
          <Input placeholder="Port" value={f.port} onChange={set("port")} inputMode="numeric" required />
        </div>
        {nodeMulti && (
          <label className="block text-xs text-muted-foreground">
            Engine
            <Select className="mt-1" value={f.core || ""} onChange={set("core")}>
              <option value="">Default ({node.core === "singbox" ? "sing-box" : "Xray"})</option>
              {(node.enabled_cores ?? [node.core]).map((c) => (
                <option key={c} value={c}>{c === "singbox" ? "sing-box" : "Xray-core"}</option>
              ))}
            </Select>
          </label>
        )}
        <div className="grid grid-cols-3 gap-2">
          <Select value={f.protocol} onChange={set("protocol")} disabled={editing}>
            {protocols.map((p) => <option key={p} value={p}>{p}</option>)}
          </Select>
          {!isNoTransport && (
            <Select value={f.network} onChange={set("network")}>
              {networks.map((n) => <option key={n} value={n}>{n}</option>)}
            </Select>
          )}
          {!isNoTransport && (
            <Select value={f.security} onChange={set("security")}>
              {securitiesFor(f.protocol).map((s) => <option key={s} value={s}>{s}</option>)}
            </Select>
          )}
          {isNoTransport && udpNative.includes(f.protocol) && (
            <span className="col-span-2 flex items-center rounded-md border border-border/60 px-2 text-xs text-muted-foreground">udp</span>
          )}
        </div>
        <Input placeholder="SNI (comma-separated, optional)" value={f.sni} onChange={set("sni")} />
        {["ws", "httpupgrade", "http", "h2", "xhttp"].includes(f.network) && (
          <Input placeholder="Path (e.g. /ws)" value={f.path} onChange={set("path")} />
        )}
        {f.network === "grpc" && (
          <Input placeholder="gRPC serviceName" value={f.path} onChange={set("path")} />
        )}
        {["ws", "httpupgrade", "http", "h2", "xhttp"].includes(f.network) && (
          <Input placeholder="Host (comma-separated, optional)" value={f.host} onChange={set("host")} />
        )}
        {f.protocol === "vless" && (f.security === "tls" || f.security === "reality") && (
          <Input placeholder="Flow (e.g. xtls-rprx-vision, optional)" value={f.flow} onChange={set("flow")} />
        )}
        {f.security === "reality" && <RealityKeygenSection onKeys={setRealityKeys} />}
        {f.protocol === "wireguard" && (
          <div className="space-y-2 rounded-lg border border-white/[0.06] bg-white/[0.02] p-3">
            <span className="text-xs font-medium text-fg-muted">WireGuard</span>
            <Input placeholder="Private key" value={f.wgPrivateKey} onChange={set("wgPrivateKey")} dir="ltr" className="font-mono text-xs" />
            <div className="grid grid-cols-2 gap-2">
              <Input placeholder="Subnet (e.g. 10.7.0.0/24)" value={f.wgSubnet} onChange={set("wgSubnet")} dir="ltr" className="font-mono text-xs" />
              <Input placeholder="MTU (default 1280)" value={f.wgMtu} onChange={set("wgMtu")} inputMode="numeric" />
            </div>
            <p className="text-[10px] text-fg-subtle">Listen port uses the Port field above. Peers are derived from bound users.</p>
          </div>
        )}
        <div>
          <p className="text-[10px] font-medium text-fg-muted mb-1">Geo-blocking (allowed countries, comma-separated ISO codes)</p>
          <Input placeholder="e.g. IR,TR (empty = all allowed)" value={f.geoAllow ?? ""} onChange={(e) => setF(s => ({...s, geoAllow: e.target.value}))} />
        </div>
        <div className="flex justify-end">
          <Button type="submit" disabled={create.isPending || update.isPending}>
            {editing ? "Save changes" : "Add inbound"}
          </Button>
        </div>
      </form>
      ))}
    </Modal>
    <SubHostsModal inbound={hostsFor} onClose={() => setHostsFor(null)} />
    </>
  );
}

function RealityKeygenSection({ onKeys }: { onKeys: (keys: { private_key: string; public_key: string; short_id: string } | null) => void }) {
  const reality = useReality();
  const [keys, setKeys] = useState<{ private_key: string; public_key: string; short_id: string } | null>(null);

  async function generate() {
    const r = await reality.mutateAsync();
    setKeys(r);
    onKeys(r);
  }

  return (
    <div className="space-y-2 rounded-lg border border-white/[0.06] bg-white/[0.02] p-3">
      <div className="flex items-center justify-between">
        <span className="text-xs font-medium text-fg-muted">REALITY keys</span>
        <Button type="button" variant="ghost" size="sm" onClick={generate} disabled={reality.isPending}>
          Generate
        </Button>
      </div>
      {keys && (
        <div className="space-y-1.5">
          <CopyField value={keys.private_key} />
          <CopyField value={keys.public_key} />
          <CopyField value={keys.short_id} />
          <p className="text-[10px] text-fg-subtle">
            Store private_key + short_id in the inbound's raw.reality; give public_key to clients.
          </p>
        </div>
      )}
    </div>
  );
}
