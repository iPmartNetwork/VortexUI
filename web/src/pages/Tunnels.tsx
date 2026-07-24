import { useState } from "react";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { Copy, Info, Plus, Trash2, WifiOff, Cable } from "lucide-react";
import { api } from "@/api/client";
import { Button, Input, Select } from "@/components/ui";
import { Modal } from "@/components/Modal";
import { GlassCard, StatusBadge } from "@/components/veltrix";
import { useToast } from "@/components/toast";
import { useConfirm } from "@/components/confirm";
import { useI18n } from "@/i18n/i18n";
import { useTitle } from "@/lib/useTitle";

interface TunnelConfig {
  id: string;
  backend: "backhaul" | "rathole" | "wstunnel";
  port: number;
  transport: string;
  secret: string;
  node_ip: string;
  iran_ip: string;
  enabled: boolean;
  status: string;
  created_at: string;
  updated_at: string;
}

const BACKENDS = [
  { value: "backhaul", label: "Backhaul" },
  { value: "rathole", label: "Rathole" },
  { value: "wstunnel", label: "Wstunnel" },
];

const TRANSPORTS = [
  { value: "tcp", label: "TCP" },
  { value: "ws", label: "WebSocket" },
  { value: "wss", label: "WebSocket (TLS)" },
  { value: "tcpmux", label: "TCP Mux" },
  { value: "wsmux", label: "WS Mux" },
  { value: "wssmux", label: "WSS Mux" },
];

export function Tunnels() {
  useTitle("Iran Bridge Tunnels");
  const { t } = useI18n();
  const qc = useQueryClient();
  const toast = useToast();
  const confirm = useConfirm();
  const [createOpen, setCreateOpen] = useState(false);
  const [bridgeCmd, setBridgeCmd] = useState<string | null>(null);

  const { data } = useQuery({
    queryKey: ["tunnels"],
    queryFn: () => api<{ tunnels: TunnelConfig[] }>("/api/tunnels"),
  });

  const delMut = useMutation({
    mutationFn: (id: string) => api<void>(`/api/tunnels/${id}`, { method: "DELETE" }),
    onSuccess: () => qc.invalidateQueries({ queryKey: ["tunnels"] }),
  });

  async function remove(tunnel: TunnelConfig) {
    const ok = await confirm({
      title: `${t("common.delete")} tunnel on port ${tunnel.port}?`,
      confirmLabel: t("common.delete"),
      destructive: true,
    });
    if (!ok) return;
    await delMut.mutateAsync(tunnel.id);
    toast.success(t("common.deleted"));
  }

  async function generateBridgeCommand(id: string) {
    try {
      const res = await api<{ command: string }>(`/api/tunnels/bridge-command/${id}`);
      setBridgeCmd(res.command);
    } catch {
      toast.error("Failed to generate bridge command");
    }
  }

  function copyToClipboard(text: string) {
    navigator.clipboard.writeText(text);
    toast.success("Copied to clipboard");
  }

  const tunnels = data?.tunnels ?? [];

  return (
    <div className="space-y-5 animate-page-enter">
      <div className="flex flex-col lg:flex-row lg:items-start justify-between gap-4">
        <div>
          <h1 className="text-2xl font-bold text-fg tracking-tight">{t("tunnels.title")}</h1>
          <p className="text-sm text-fg-muted mt-1 max-w-2xl">{t("tunnels.subtitle")}</p>
        </div>
        <Button onClick={() => setCreateOpen(true)} className="flex-shrink-0">
          <Plus size={14} /> {t("tunnels.addTunnel")}
        </Button>
      </div>

      {/* Info banner */}
      <div className="rounded-xl border border-primary/25 bg-primary/5 px-4 py-3 flex items-start gap-3">
        <div className="h-8 w-8 rounded-full bg-primary/15 flex items-center justify-center text-primary flex-shrink-0">
          <Info size={16} />
        </div>
        <div className="text-xs text-fg-muted leading-relaxed space-y-1.5">
          <p className="font-semibold text-fg text-sm">{t("tunnels.infoTitle")}</p>
          <p>{t("tunnels.infoDesc")}</p>
          <ul className="space-y-1 pt-1">
            <li><strong className="text-fg">Backhaul</strong> — {t("tunnels.backhaulDesc")}</li>
            <li><strong className="text-fg">Rathole</strong> — {t("tunnels.ratholeDesc")}</li>
            <li><strong className="text-fg">Wstunnel</strong> — {t("tunnels.wstunnelDesc")}</li>
          </ul>
        </div>
      </div>

      <CreateTunnelModal open={createOpen} onClose={() => setCreateOpen(false)} />

      {/* Tunnel list */}
      <div className="grid grid-cols-1 gap-4 lg:grid-cols-2">
        {tunnels.map((tunnel) => (
          <GlassCard key={tunnel.id} hover className="!p-4 space-y-3">
            <div className="flex items-center justify-between gap-2">
              <div className="flex items-center gap-2 min-w-0">
                <div className="h-9 w-9 rounded-xl bg-primary/10 flex items-center justify-center text-primary flex-shrink-0">
                  <Cable size={16} />
                </div>
                <div className="min-w-0">
                  <div className="font-semibold text-fg text-sm truncate">
                    {tunnel.backend.toUpperCase()} :{tunnel.port}
                  </div>
                  <div className="text-xs text-fg-muted truncate">
                    {tunnel.transport} &middot; {tunnel.iran_ip || "Iran VPS"} → {tunnel.node_ip || "Node"}
                  </div>
                </div>
              </div>
              <StatusBadge
                status={tunnel.status === "online" ? "success" : tunnel.status === "offline" ? "error" : "warning"}
                label={tunnel.status}
              />
            </div>

            <div className="flex items-center gap-2 pt-1">
              <Button
                size="sm"
                variant="outline"
                onClick={() => generateBridgeCommand(tunnel.id)}
              >
                <Copy size={12} /> {t("tunnels.bridgeCmd")}
              </Button>
              <Button
                size="sm"
                variant="ghost"
                className="text-red-500 hover:text-red-600"
                onClick={() => remove(tunnel)}
              >
                <Trash2 size={12} />
              </Button>
            </div>
          </GlassCard>
        ))}
        {tunnels.length === 0 && (
          <div className="col-span-full text-center py-12 text-fg-muted">
            <WifiOff size={32} className="mx-auto mb-3 opacity-40" />
            <p>{t("tunnels.empty")}</p>
          </div>
        )}
      </div>

      {/* Bridge command modal */}
      {bridgeCmd && (
        <Modal open onClose={() => setBridgeCmd(null)} title={t("tunnels.bridgeCmdTitle")}>
          <div className="space-y-3">
            <p className="text-sm text-fg-muted">{t("tunnels.bridgeCmdHint")}</p>
            <div className="relative">
              <pre className="bg-surface-2 rounded-lg p-3 text-xs overflow-x-auto whitespace-pre-wrap break-all font-mono">
                {bridgeCmd}
              </pre>
              <button
                className="absolute top-2 right-2 p-1.5 rounded-md bg-surface-3 hover:bg-surface-4 text-fg-muted hover:text-fg transition"
                onClick={() => copyToClipboard(bridgeCmd)}
                aria-label="Copy command"
              >
                <Copy size={14} />
              </button>
            </div>
            <Button variant="outline" onClick={() => setBridgeCmd(null)} className="w-full">
              {t("common.close")}
            </Button>
          </div>
        </Modal>
      )}
    </div>
  );
}

// ─── Create Tunnel Modal ─────────────────────────────────────────────────────

function CreateTunnelModal({ open, onClose }: { open: boolean; onClose: () => void }) {
  const { t } = useI18n();
  const qc = useQueryClient();
  const toast = useToast();
  const [backend, setBackend] = useState("backhaul");
  const [port, setPort] = useState("443");
  const [transport, setTransport] = useState("tcp");
  const [secret, setSecret] = useState("");
  const [nodeIP, setNodeIP] = useState("");
  const [iranIP, setIranIP] = useState("");
  const [loading, setLoading] = useState(false);

  async function submit(e: React.FormEvent) {
    e.preventDefault();
    setLoading(true);
    try {
      await api("/api/tunnels", {
        method: "POST",
        body: {
          backend,
          port: parseInt(port, 10),
          transport,
          secret,
          node_ip: nodeIP,
          iran_ip: iranIP,
        },
      });
      qc.invalidateQueries({ queryKey: ["tunnels"] });
      toast.success(t("tunnels.created"));
      onClose();
    } catch {
      toast.error("Failed to create tunnel");
    } finally {
      setLoading(false);
    }
  }

  return (
    <Modal open={open} onClose={onClose} title={t("tunnels.addTunnel")}>
      <form onSubmit={submit} className="space-y-4">
        <div className="space-y-1.5">
          <label className="text-sm font-medium text-fg">{t("tunnels.backend")}</label>
          <Select value={backend} onChange={(e) => setBackend(e.target.value)}>
            {BACKENDS.map((b) => <option key={b.value} value={b.value}>{b.label}</option>)}
          </Select>
        </div>
        <div className="space-y-1.5">
          <label className="text-sm font-medium text-fg">{t("tunnels.port")}</label>
          <Input
            type="number"
            min={1}
            max={65535}
            value={port}
            onChange={(e) => setPort(e.target.value)}
            required
          />
        </div>
        <div className="space-y-1.5">
          <label className="text-sm font-medium text-fg">{t("tunnels.transport")}</label>
          <Select value={transport} onChange={(e) => setTransport(e.target.value)}>
            {TRANSPORTS.map((tr) => <option key={tr.value} value={tr.value}>{tr.label}</option>)}
          </Select>
        </div>
        <div className="space-y-1.5">
          <label className="text-sm font-medium text-fg">{t("tunnels.secret")}</label>
          <Input
            placeholder="shared-secret-key"
            value={secret}
            onChange={(e) => setSecret(e.target.value)}
          />
        </div>
        <div className="space-y-1.5">
          <label className="text-sm font-medium text-fg">{t("tunnels.nodeIP")}</label>
          <Input
            placeholder="Kharej (node) server IP"
            value={nodeIP}
            onChange={(e) => setNodeIP(e.target.value)}
          />
        </div>
        <div className="space-y-1.5">
          <label className="text-sm font-medium text-fg">{t("tunnels.iranIP")}</label>
          <Input
            placeholder="Iran VPS IP"
            value={iranIP}
            onChange={(e) => setIranIP(e.target.value)}
          />
        </div>
        <Button type="submit" className="w-full" disabled={loading}>
          {loading ? "Saving…" : t("tunnels.addTunnel")}
        </Button>
      </form>
    </Modal>
  );
}
