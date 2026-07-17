import { useEffect, useState } from "react";
import { useCapabilities, useCreateInbound, useDeleteInbound, useNodeInbounds, useUpdateInbound, inboundToUpdateInput, type Inbound } from "@/api/hooks";
import { useReality } from "@/api/policy-hooks";
import type { Node } from "@/api/types";
import { Button, Input, Select } from "./ui";
import { Modal } from "./Modal";
import { JsonCodeEditor } from "./JsonCodeEditor";
import { SubHostsModal } from "./SubHostsModal";
import { useToast } from "./toast";
import { useAuth } from "@/auth/auth";
import type { CoreType } from "@/lib/coreTypes";
import { isMultiCore, resolveInboundCore } from "@/lib/coreTypes";
import { CoreBadge } from "@/components/veltrix/CoreBadge";
import { cn } from "@/lib/utils";
import { PortConflictIndicator } from "./PortConflictIndicator";

const PROTOCOLS = ["vless", "vmess", "trojan", "shadowsocks", "hysteria2", "tuic", "wireguard", "socks", "http", "naive", "dokodemo"];
const NETWORKS = ["tcp", "ws", "grpc", "httpupgrade", "http", "h2", "xhttp", "kcp", "quic", "udp"];
const SECURITIES = ["none", "tls", "reality"];
const UDP_PROTOCOLS = ["hysteria2", "tuic", "wireguard", "hysteria"];
const NO_TRANSPORT_PROTOCOLS = ["hysteria2", "tuic", "wireguard", "socks", "http", "naive", "dokodemo", "hysteria", "shadowtls", "anytls"];

const PROTOCOL_COLORS: Record<string, string> = {
  vless: "bg-cyan-500/20 text-cyan-300 border-cyan-500/30",
  vmess: "bg-blue-500/20 text-blue-300 border-blue-500/30",
  trojan: "bg-green-500/20 text-green-300 border-green-500/30",
  shadowsocks: "bg-yellow-500/20 text-yellow-300 border-yellow-500/30",
  hysteria2: "bg-purple-500/20 text-purple-300 border-purple-500/30",
  tuic: "bg-pink-500/20 text-pink-300 border-pink-500/30",
  wireguard: "bg-emerald-500/20 text-emerald-300 border-emerald-500/30",
  socks: "bg-orange-500/20 text-orange-300 border-orange-500/30",
  http: "bg-red-500/20 text-red-300 border-red-500/30",
  naive: "bg-indigo-500/20 text-indigo-300 border-indigo-500/30",
  dokodemo: "bg-gray-500/20 text-gray-300 border-gray-500/30",
};

const PROTOCOL_DESCRIPTIONS: Record<string, string> = {
  vless: "Lightweight proxy protocol with encryption, ideal for CDN setups",
  vmess: "Encrypted protocol designed for anti-censorship, works well with WS+TLS",
  trojan: "HTTPS-traffic mimicking proxy with TLS, hard to detect",
  shadowsocks: "Simple encrypted tunnel, lightweight and fast",
  hysteria2: "UDP-based protocol optimized for unstable/high-latency networks",
  tuic: "Lightweight QUIC-based protocol, good for mobile networks",
  wireguard: "High-performance VPN protocol with built-in encryption",
  socks: "Simple SOCKS5 proxy, no encryption (use with TLS)",
  http: "Plain HTTP proxy (use with TLS for security)",
  naive: "NaiveProxy — Chrome-standard traffic mimicry",
  dokodemo: "Transparent TCP/UDP redirect (port forwarding)",
};

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

const randomPort = () => String(10000 + Math.floor(Math.random() * 50000));

const newBlank = () => ({ editId: "", tag: "", core: "" as CoreType | "", protocol: "vless", port: randomPort(), portEnd: "", network: "tcp", security: "tls", sni: "", path: "", host: "", flow: "", geoAllow: "", wgPrivateKey: "", wgSubnet: "", wgMtu: "", listen: "", speedLimit: "", notes: "" });
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

function InboundListItem({
  ib,
  canWrite,
  onEdit,
  onToggle,
  onDelete,
  onHosts,
  udpNative,
  noTransport,
  nodeMulti,
  node,
}: {
  ib: Inbound;
  canWrite: boolean;
  onEdit: () => void;
  onToggle: () => void;
  onDelete: () => void;
  onHosts: () => void;
  udpNative: string[];
  noTransport: string[];
  nodeMulti: boolean;
  node: Node;
}) {
  return (
    <div className="group flex items-center justify-between gap-3 rounded-xl border border-border/50 bg-surface/50 px-4 py-3 text-sm hover:bg-surface-2/40 hover:border-border/80 transition-all">
      <div className="flex items-center gap-3 min-w-0 flex-1">
        {/* Protocol dot */}
        <div className={cn(
          "h-8 w-8 rounded-lg flex items-center justify-center flex-shrink-0 border text-[10px] font-bold",
          ib.enabled
            ? PROTOCOL_COLORS[ib.protocol] ?? "bg-primary/10 text-primary border-primary/20"
            : "bg-surface-2/60 text-fg-subtle border-border/40",
        )}>
          {ib.protocol.slice(0, 2).toUpperCase()}
        </div>
        {/* Info */}
        <div className="min-w-0 flex-1">
          <div className="flex items-center gap-1.5 flex-wrap">
            <span className={cn(
              "font-semibold text-sm truncate max-w-[400px]",
              ib.enabled ? "text-fg" : "text-fg-muted",
            )}>
              {ib.tag}
            </span>
            <span className={cn(
              "inline-flex items-center rounded-md border px-1.5 py-0.5 text-[10px] font-semibold",
              ib.enabled ? "border-border/60 bg-surface-2/70 text-fg-muted" : "border-border/30 bg-surface-2/40 text-fg-subtle opacity-50",
            )}>{ib.protocol}</span>
            {nodeMulti && (
              <CoreBadge core={resolveInboundCore(node, ib.core)} className="scale-75" />
            )}
          </div>
          <div className="flex items-center gap-2 mt-0.5">
            <span className="text-xs font-mono text-fg-subtle">:{ib.port}</span>
            <span className="text-[10px] text-fg-subtle px-1.5 py-0.5 rounded bg-surface-2/50 border border-border/30">
              {inboundTransportLabel(ib, udpNative, noTransport)}
            </span>
            <span className={cn(
              "text-[9px] font-bold uppercase px-1.5 py-0.5 rounded",
              ib.enabled ? "text-success bg-success/10" : "text-fg-subtle bg-surface-2/50"
            )}>
              {ib.enabled ? "ON" : "OFF"}
            </span>
          </div>
        </div>
      </div>
      {/* Actions */}
      <div className="flex items-center gap-1.5 flex-shrink-0">
        {canWrite && (
          <>
            <button type="button" onClick={onToggle}
              className="h-7 px-2.5 rounded-lg text-[10px] font-medium border border-border/50 text-fg-subtle hover:text-fg hover:bg-surface-2/60 transition-all"
              title={ib.enabled ? "Disable" : "Enable"}
            >
              {ib.enabled ? "Disable" : "Enable"}
            </button>
            <button type="button" onClick={onEdit}
              className="h-7 px-2.5 rounded-lg text-[10px] font-medium border border-primary/30 text-primary/80 hover:text-primary hover:bg-primary/10 transition-all"
              title="Edit inbound"
            >
              Edit
            </button>
            <button type="button" onClick={onDelete}
              className="h-7 px-2.5 rounded-lg text-[10px] font-medium border border-danger/30 text-danger/70 hover:text-danger hover:bg-danger/10 transition-all"
              title="Delete inbound"
            >
              Delete
            </button>
          </>
        )}
        <button type="button" onClick={onHosts}
          className="h-7 px-2.5 rounded-lg text-[10px] font-medium border border-border/50 text-fg-subtle hover:text-fg hover:bg-surface-2/60 transition-all"
          title="Subscription hosts"
        >
          Hosts
        </button>
      </div>
    </div>
  );
}

function SectionCard({ title, description, children }: { title: string; description?: string; children: React.ReactNode }) {
  return (
    <div className="rounded-xl border border-border/40 bg-surface/20 p-4 space-y-3">
      <div>
        <h4 className="text-xs font-bold text-fg tracking-wide">{title}</h4>
        {description && <p className="text-[10px] text-fg-subtle mt-0.5">{description}</p>}
      </div>
      {children}
    </div>
  );
}

function FieldLabel({ label, hint, required }: { label: string; hint?: string; required?: boolean }) {
  return (
    <div className="flex items-center gap-1.5 mb-1">
      <span className="text-[11px] font-semibold text-fg-muted tracking-wide uppercase">
        {label}
        {required && <span className="text-danger ml-0.5">*</span>}
      </span>
      {hint && (
        <span className="text-[9px] text-fg-subtle" title={hint}>ⓘ</span>
      )}
    </div>
  );
}

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
  const [f, setF] = useState({ ...blank });
  const [editSnapshot, setEditSnapshot] = useState<Inbound | null>(null);
  const [hostsFor, setHostsFor] = useState<Inbound | null>(null);
  const [realityKeys, setRealityKeys] = useState<{ private_key: string; public_key: string; short_id: string } | null>(null);
  const [tab, setTab] = useState<"basics" | "json">("basics");
  const [jsonText, setJsonText] = useState("");
  const [jsonErr, setJsonErr] = useState("");

  const resetForm = () => {
    setF(newBlank());
    setRealityKeys(null);
    setEditSnapshot(null);
    setTab("basics");
  };

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
    setF({
      editId: ib.id, tag: ib.tag, core: (ib.core ?? "") as CoreType | "",
      protocol: ib.protocol, port: String(ib.port),
      network: udpNative.includes(ib.protocol) || noTransport.includes(ib.protocol) ? "" : ib.network,
      security: ib.security,
      sni: (ib.sni ?? []).join(", "), path: ib.path ?? "",
      host: (ib.host ?? []).join(", "), flow: ib.flow ?? "",
      geoAllow: (ib.geo_policy?.allowed_countries ?? []).join(", "),
      wgPrivateKey: typeof wg.private_key === "string" ? wg.private_key : "",
      wgSubnet: typeof wg.subnet === "string" ? wg.subnet : "",
      wgMtu: typeof wg.mtu === "number" ? String(wg.mtu) : "",
      listen: ib.listen || "",
      speedLimit: ib.speed_limit ? String(ib.speed_limit / 125000) : "",
      portEnd: ib.port_end ? String(ib.port_end) : "",
      notes: ib.notes || "",
    });
    setTab("basics");
  }

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
    if (f.protocol === "wireguard") {
      const wg: Record<string, unknown> = {};
      if (f.wgPrivateKey.trim()) wg.private_key = f.wgPrivateKey.trim();
      if (f.wgSubnet.trim()) wg.subnet = f.wgSubnet.trim();
      if (f.wgMtu.trim()) wg.mtu = Number(f.wgMtu);
      if (Object.keys(wg).length > 0) raw = { ...(raw ?? {}), wireguard: wg };
    }
    const network = isNoTransport ? "" : f.network;
    const corePayload = f.core || undefined;
    const speedLimitBytes = f.speedLimit ? Number(f.speedLimit) * 125000 : 0;
    try {
      if (editing && editSnapshot) {
        const mergedRaw = raw ?? editSnapshot.raw;
        await update.mutateAsync({
          id: f.editId,
          input: inboundToUpdateInput(editSnapshot, {
            port: Number(f.port), network, security: f.security,
            core: f.core, sni, path: f.path, host, flow: f.flow, geo_policy,
            ...(mergedRaw ? { raw: mergedRaw } : {}),
            speed_limit: speedLimitBytes,
            port_end: f.portEnd ? Number(f.portEnd) : 0,
            notes: f.notes,
            listen: f.listen || "",
          }),
        });
        toast.success(`Updated ${f.tag}`);
      } else if (editing) {
        await update.mutateAsync({
          id: f.editId,
          input: {
            port: Number(f.port), network, security: f.security, sni, path: f.path,
            host, flow: f.flow, geo_policy, core: f.core, enabled: true,
            ...(raw ? { raw } : {}),
            speed_limit: speedLimitBytes,
            port_end: f.portEnd ? Number(f.portEnd) : 0,
            notes: f.notes,
            listen: f.listen || "",
          },
        });
        toast.success(`Updated ${f.tag}`);
      } else {
        await create.mutateAsync({
          node_id: node.id, tag: f.tag, core: corePayload, protocol: f.protocol,
          port: Number(f.port), network, security: f.security, sni, path: f.path,
          host, flow: f.flow, geo_policy, enabled: true,
          ...(raw ? { raw } : {}),
          speed_limit: speedLimitBytes,
          port_end: f.portEnd ? Number(f.portEnd) : 0,
          notes: f.notes,
          listen: f.listen || "",
        });
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

  const existingInbounds = list.data?.inbounds ?? [];
  const protoColor = PROTOCOL_COLORS[f.protocol] ?? "bg-primary/10 text-primary border-primary/20";

  return (
    <>
    <Modal open={!!node} onClose={onClose} title="" className="!max-w-6xl">
      {/* Header */}
      {node && (
        <div className="flex items-center justify-between border-b border-border/40 pb-4 mb-4">
          <div>
            <h2 className="text-lg font-bold text-fg">Inbound Manager</h2>
            <p className="text-xs text-fg-muted mt-0.5">
              <span className="font-medium text-fg-subtle">{node.name}</span>
              {nodeMulti && (
                <span className="ml-2">
                  <CoreBadge core={resolveInboundCore(node, "")} />
                </span>
              )}
            </p>
          </div>
          {/* Summary badge */}
          <div className="hidden sm:flex items-center gap-1.5 rounded-lg bg-surface-2/60 border border-border/40 px-3 py-1.5">
            <span className="text-[10px] text-fg-subtle font-medium">{existingInbounds.length} inbound{existingInbounds.length !== 1 ? "s" : ""}</span>
          </div>
        </div>
      )}

      {/* Existing inbounds list */}
      {existingInbounds.length > 0 && (
        <div className="space-y-1.5 mb-5">
          {existingInbounds.map((ib) => (
            <InboundListItem
              key={ib.id}
              ib={ib}
              canWrite={canWrite}
              onEdit={() => startEdit(ib)}
              onToggle={() => toggleEnable(ib)}
              onDelete={() => del.mutate(ib.id)}
              onHosts={() => setHostsFor(ib)}
              udpNative={udpNative}
              noTransport={noTransport}
              nodeMulti={nodeMulti}
              node={node}
            />
          ))}
        </div>
      )}

      {existingInbounds.length === 0 && (
        <div className="flex items-center gap-2.5 rounded-xl border border-dashed border-border/60 bg-surface/20 px-4 py-3 mb-5">
          <div className="h-6 w-6 rounded-lg bg-primary/10 flex items-center justify-center text-primary flex-shrink-0">
            <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
              <circle cx="12" cy="12" r="10" /><line x1="12" y1="8" x2="12" y2="16" /><line x1="8" y1="12" x2="16" y2="12" />
            </svg>
          </div>
          <p className="text-xs text-fg-muted">No inbounds on this node yet. Create your first one below.</p>
        </div>
      )}

      {canWrite && (
      <>
        {/* Tab switcher */}
        <div className="flex gap-6 border-b border-border/30 mb-4">
          {(["basics", "json"] as const).map((tk) => (
            <button
              key={tk}
              type="button"
              onClick={() => (tk === "json" ? gotoJson() : setTab("basics"))}
              className={cn(
                "pb-2 text-xs font-bold tracking-wider uppercase transition-all border-b-2 -mb-px",
                tab === tk
                  ? "border-primary text-primary"
                  : "border-transparent text-fg-subtle hover:text-fg-muted",
              )}
            >
              {tk === "basics" ? "⚡ Quick Form" : "⚙️ Raw JSON"}
            </button>
          ))}
          {editing && (
            <button type="button" onClick={resetForm}
              className="ml-auto pb-2 text-[10px] text-fg-subtle hover:text-primary transition-colors border-b-2 border-transparent"
            >
              + New inbound
            </button>
          )}
        </div>

        {tab === "json" ? (
          <div className="space-y-3">
            <div className="rounded-xl bg-amber-500/5 border border-amber-500/20 px-3.5 py-2.5">
              <p className="text-[10px] font-semibold text-amber-400 uppercase tracking-wide">Advanced</p>
              <p className="text-xs text-fg-muted mt-1 leading-relaxed">
                Paste a full Xray or sing-box inbound JSON object. The <code className="text-fg-subtle bg-surface-2/60 px-1 rounded text-[10px]">tag</code>, <code className="text-fg-subtle bg-surface-2/60 px-1 rounded text-[10px]">protocol</code>, and <code className="text-fg-subtle bg-surface-2/60 px-1 rounded text-[10px]">port</code> fields are required. The entire object is stored as a raw config override — only use this if you know the exact format your core expects.
              </p>
            </div>
            <JsonCodeEditor value={jsonText} onChange={setJsonText} rows={14} />
            {jsonErr && (
              <div className="flex items-center gap-2 rounded-lg bg-danger/10 border border-danger/20 px-3 py-2">
                <span className="text-xs text-danger font-medium">{jsonErr}</span>
              </div>
            )}
            <div className="flex justify-end">
              <Button type="button" onClick={submitJSON} disabled={create.isPending}>
                {create.isPending ? "Saving..." : "Add inbound from JSON"}
              </Button>
            </div>
          </div>
        ) : (
          <form onSubmit={submit} className="space-y-4">
            {/* Section: General */}
            <SectionCard title="General" description="Basic identification and protocol selection">
              <div className="grid grid-cols-1 sm:grid-cols-3 gap-3">
                <div>
                  <FieldLabel label="Tag" required hint="A unique name for this inbound (e.g. vless-ws-443)" />
                  <Input placeholder="e.g. inbound-main" value={f.tag} onChange={set("tag")} required disabled={editing} />
                </div>
                <div>
                  <FieldLabel label="Port" required hint="The port this inbound listens on. Random port is pre-filled." />
                  <div className="flex items-center gap-1.5">
                    <Input placeholder="443" value={f.port} onChange={set("port")} inputMode="numeric" required className="flex-1" />
                    <button
                      type="button"
                      onClick={() => setF(s => ({ ...s, port: randomPort() }))}
                      className="h-9 w-9 rounded-lg border border-border/50 bg-surface/40 flex items-center justify-center text-fg-subtle hover:text-primary hover:border-primary/40 hover:bg-primary/5 transition-all flex-shrink-0"
                      title="Generate random port"
                    >
                      🎲
                    </button>
                  </div>
                  <PortConflictIndicator nodeId={node?.id ?? ""} port={f.port} />
                </div>
                <div>
                  <FieldLabel label="Listen" hint="IP address to bind. 0.0.0.0 binds all interfaces." />
                  <Input placeholder="0.0.0.0 (all interfaces)" value={f.listen} onChange={set("listen")} dir="ltr" className="font-mono text-xs" />
                </div>
              </div>
              {["hysteria2", "tuic"].includes(f.protocol) && (
                <div>
                  <FieldLabel label="Port End (Range)" hint="End of port range for multi-port. Leave empty for single port." />
                  <Input placeholder="e.g. 3000 (empty = single port)" value={f.portEnd} onChange={set("portEnd")} inputMode="numeric" />
                </div>
              )}
              <div>
                <FieldLabel label="Protocol" required hint="The proxy protocol to use. Different protocols have different features and transport options." />
                <div className="relative">
                  <Select value={f.protocol} onChange={set("protocol")} disabled={editing} className="w-full">
                    {protocols.map((p) => <option key={p} value={p}>{p}</option>)}
                  </Select>
                </div>
                {/* Protocol description + badge */}
                <div className="flex items-center gap-2 mt-1.5">
                  <span className={cn("inline-flex items-center gap-1 rounded-md border px-2 py-0.5 text-[11px] font-bold", protoColor)}>
                    {f.protocol.toUpperCase()}
                  </span>
                  <span className="text-[10px] text-fg-subtle leading-relaxed">
                    {PROTOCOL_DESCRIPTIONS[f.protocol] ?? ""}
                  </span>
                </div>
              </div>
            </SectionCard>

            {/* Section: Engine + Transport */}
            <SectionCard title="Transport" description="How data flows through this inbound">
              {nodeMulti && (
                <div>
                  <FieldLabel label="Engine" hint="Which proxy core runs this inbound. Defaults to the node's primary core." />
                  <Select value={f.core || ""} onChange={set("core")}>
                    <option value="">Default ({node.core === "singbox" ? "sing-box" : "Xray"})</option>
                    {(node.enabled_cores ?? [node.core]).map((c) => (
                      <option key={c} value={c}>{c === "singbox" ? "sing-box" : "Xray-core"}</option>
                    ))}
                  </Select>
                </div>
              )}
              {!isNoTransport ? (
                <div className="grid grid-cols-2 gap-3">
                  <div>
                    <FieldLabel label="Network" hint="Transport protocol. WS+CDN works best for anti-censorship." />
                    <Select value={f.network} onChange={set("network")}>
                      {networks.map((n) => <option key={n} value={n}>{n}</option>)}
                    </Select>
                  </div>
                  <div>
                    <FieldLabel label="Security" hint="TLS encrypts traffic. REALITY also obfuscates as a real website." />
                    <Select value={f.security} onChange={set("security")}>
                      {securitiesFor(f.protocol).map((s) => <option key={s} value={s}>{s}</option>)}
                    </Select>
                  </div>
                </div>
              ) : (
                <div className="flex items-center gap-2 rounded-lg bg-surface-2/50 border border-border/30 px-3 py-2">
                  {udpNative.includes(f.protocol) ? (
                    <>
                      <span className="h-2 w-2 rounded-full bg-purple-500" />
                      <span className="text-xs font-medium text-fg-subtle">
                        UDP-native protocol — no transport layer needed
                      </span>
                    </>
                  ) : (
                    <>
                      <span className="h-2 w-2 rounded-full bg-amber-500" />
                      <span className="text-xs font-medium text-fg-subtle">
                        This protocol has no transport options
                      </span>
                    </>
                  )}
                </div>
              )}
            </SectionCard>

            {/* Section: Network Options (conditional) */}
            {!isNoTransport && (
              <SectionCard title="Network Options" description="TLS/REALITY settings and stream customization">
                <div>
                  <FieldLabel label="SNI" hint="Server Name Indication — the domain your CDN or TLS expects. Comma-separated for multiple." />
                  <Input placeholder="e.g. example.com, cdn.example.com" value={f.sni} onChange={set("sni")} dir="ltr" />
                </div>
                {["ws", "httpupgrade", "http", "h2", "xhttp"].includes(f.network) && (
                  <>
                    <div>
                      <FieldLabel label="Path" hint="WebSocket/HTTP path that CDN routes on." />
                      <Input placeholder="e.g. /ws, /api/v1" value={f.path} onChange={set("path")} />
                    </div>
                    <div>
                      <FieldLabel label="Host" hint="HTTP Host header. Comma-separated for multiple hosts." />
                      <Input placeholder="e.g. example.com (optional)" value={f.host} onChange={set("host")} dir="ltr" />
                    </div>
                  </>
                )}
                {f.network === "grpc" && (
                  <div>
                    <FieldLabel label="gRPC Service Name" hint="The service name for gRPC transport." />
                    <Input placeholder="e.g. myservice" value={f.path} onChange={set("path")} />
                  </div>
                )}
                {f.protocol === "vless" && (f.security === "tls" || f.security === "reality") && (
                  <div>
                    <FieldLabel label="Flow" hint="Flow control for VLESS (e.g. xtls-rprx-vision). Only for VLESS+TLS/REALITY." />
                    <Input placeholder="e.g. xtls-rprx-vision (optional)" value={f.flow} onChange={set("flow")} />
                  </div>
                )}
              </SectionCard>
            )}

            {/* Section: REALITY */}
            {f.security === "reality" && (
              <SectionCard title="REALITY Keys" description="Generate REALITY keys for TLS 1.3 session resumption handshake.">
                <RealityKeygenSection onKeys={setRealityKeys} />
              </SectionCard>
            )}

            {/* Section: WireGuard */}
            {f.protocol === "wireguard" && (
              <SectionCard title="WireGuard Settings" description="Native WireGuard inbound configuration.">
                <div>
                  <FieldLabel label="Private Key" hint="WireGuard private key for this interface." />
                  <Input placeholder="Enter private key" value={f.wgPrivateKey} onChange={set("wgPrivateKey")} dir="ltr" className="font-mono text-xs" />
                </div>
                <div className="grid grid-cols-2 gap-3">
                  <div>
                    <FieldLabel label="Subnet" hint="Subnet for the WireGuard interface (e.g. 10.7.0.0/24)." />
                    <Input placeholder="e.g. 10.7.0.0/24" value={f.wgSubnet} onChange={set("wgSubnet")} dir="ltr" className="font-mono text-xs" />
                  </div>
                  <div>
                    <FieldLabel label="MTU" hint="Maximum Transmission Unit. Default 1280." />
                    <Input placeholder="1280" value={f.wgMtu} onChange={set("wgMtu")} inputMode="numeric" />
                  </div>
                </div>
                <p className="text-[10px] text-fg-subtle bg-surface-2/40 rounded-lg px-2.5 py-1.5 border border-border/20">
                  Listens on port specified above. Peers are derived from users linked to this node.
                </p>
              </SectionCard>
            )}

            {/* Section: Geo-filtering */}
            <SectionCard title="Geo-blocking" description="Restrict inbound access by country (optional).">
              <div>
                <FieldLabel label="Allowed Countries" hint="Comma-separated ISO 3166-1 alpha-2 country codes. Leave empty to allow all." />
                <Input placeholder="e.g. IR, TR (empty = all allowed)" value={f.geoAllow ?? ""} onChange={(e) => setF(s => ({...s, geoAllow: e.target.value}))} dir="ltr" />
              </div>
              <p className="text-[10px] text-fg-subtle leading-relaxed">
                Only connections from these countries will be accepted. Useful for limiting attack surface.
                Requires geoip databases to be configured on the node.
              </p>
            </SectionCard>

            {/* Section: Speed Limit */}
            <SectionCard title="Speed Limit" description="Per-user download speed limit on this inbound (0 = unlimited).">
              <div>
                <FieldLabel label="Speed (Mbps)" hint="Download speed limit in megabits per second. Leave 0 or empty for unlimited." />
                <Input placeholder="e.g. 10 (0 = unlimited)" value={f.speedLimit} onChange={set("speedLimit")} inputMode="numeric" />
                {f.speedLimit && Number(f.speedLimit) > 0 && (
                  <p className="text-[10px] text-fg-subtle mt-1">
                    &asymp; {(Number(f.speedLimit) * 125000 / 1024 / 1024).toFixed(1)} MB/s
                  </p>
                )}
              </div>
            </SectionCard>

            {/* Section: Notes */}
            <SectionCard title="Notes" description="Operator notes for documentation (not shown to users).">
              <textarea
                className="w-full rounded-lg border border-border/50 bg-surface/30 px-3 py-2 text-sm text-fg placeholder:text-fg-subtle focus:border-primary/50 focus:ring-1 focus:ring-primary/20 outline-none resize-y min-h-[60px]"
                placeholder="e.g. CDN behind Cloudflare, Direct connection for Hysteria2..."
                value={f.notes}
                onChange={(e) => setF(s => ({ ...s, notes: e.target.value }))}
                rows={2}
              />
            </SectionCard>

            {/* Submit */}
            <div className="flex items-center justify-between pt-1">
              {editing && (
                <button type="button" onClick={resetForm}
                  className="text-[11px] text-fg-subtle hover:text-fg transition-colors underline underline-offset-2"
                >
                  Cancel editing
                </button>
              )}
              <div className="flex items-center gap-2 ml-auto">
                <Button type="submit" disabled={create.isPending || update.isPending} className={editing ? "bg-primary" : ""}>
                  {create.isPending || update.isPending ? (
                    <span className="flex items-center gap-1.5">
                      <span className="h-3 w-3 rounded-full border-2 border-current border-t-transparent animate-spin" />
                      Saving...
                    </span>
                  ) : editing ? (
                    "Save Changes"
                  ) : (
                    "Add Inbound"
                  )}
                </Button>
              </div>
            </div>
          </form>
        )}
      </>
      )}
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
    <div className="space-y-3">
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-2">
          <div className="h-5 w-5 rounded bg-purple-500/15 flex items-center justify-center text-purple-400">
            <svg width="10" height="10" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round">
              <rect x="3" y="11" width="18" height="11" rx="2" ry="2" /><path d="M7 11V7a5 5 0 0 1 10 0v4" />
            </svg>
          </div>
          <span className="text-xs font-semibold text-fg">Key Material</span>
        </div>
        <Button type="button" variant="ghost" size="sm" onClick={generate} disabled={reality.isPending}>
          {reality.isPending ? (
            <span className="flex items-center gap-1.5">
              <span className="h-3 w-3 rounded-full border-2 border-current border-t-transparent animate-spin" />
              Generating...
            </span>
          ) : "Generate Keys"}
        </Button>
      </div>
      {keys && (
        <div className="space-y-2">
          <div className="rounded-lg bg-surface-2/40 border border-border/30 p-2.5 space-y-1.5">
            <KeyRow label="Private Key" value={keys.private_key} />
            <KeyRow label="Public Key" value={keys.public_key} />
            <KeyRow label="Short ID" value={keys.short_id} />
          </div>
          <div className="flex items-start gap-1.5 rounded-lg bg-blue-500/5 border border-blue-500/15 px-2.5 py-2">
            <span className="text-[10px] text-blue-400 mt-px">💡</span>
            <p className="text-[10px] text-fg-subtle leading-relaxed">
              Private key goes into the inbound config. Share the public key with your clients.
              This key pair is used for TLS 1.3 session resumption handshake.
            </p>
          </div>
        </div>
      )}
    </div>
  );
}

function KeyRow({ label, value }: { label: string; value: string }) {
  const [copied, setCopied] = useState(false);
  return (
    <div className="flex items-center gap-2 group/key">
      <span className="text-[10px] font-semibold text-fg-subtle w-[68px] flex-shrink-0">{label}</span>
      <code className="flex-1 text-[11px] font-mono text-fg truncate bg-surface/50 px-1.5 py-0.5 rounded border border-border/20">{value}</code>
      <button
        type="button"
        onClick={() => { navigator.clipboard.writeText(value); setCopied(true); setTimeout(() => setCopied(false), 1500); }}
        className="h-5 w-5 rounded flex items-center justify-center text-fg-subtle hover:text-fg hover:bg-surface-2/60 transition-all opacity-0 group-hover/key:opacity-100"
        title={`Copy ${label}`}
      >
        {copied ? (
          <svg width="10" height="10" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round" className="text-success">
            <polyline points="20 6 9 17 4 12" />
          </svg>
        ) : (
          <svg width="10" height="10" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
            <rect x="9" y="9" width="13" height="13" rx="2" ry="2" /><path d="M5 15H4a2 2 0 0 1-2-2V4a2 2 0 0 1 2-2h9a2 2 0 0 1 2 2v1" />
          </svg>
        )}
      </button>
    </div>
  );
}
