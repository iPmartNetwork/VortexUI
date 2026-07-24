import { useState } from "react";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import {
  Shield,
  RefreshCw,
  QrCode,
  Network,
  Settings,
  Wrench,
  Plus,
} from "lucide-react";
import { api } from "@/api/client";
import { Button, Input, Badge } from "@/components/ui";
import { GlassCard } from "@/components/veltrix";
import { Modal } from "@/components/Modal";
import { useToast } from "@/components/toast";
import { useConfirm } from "@/components/confirm";
import { useTitle } from "@/lib/useTitle";
import { useI18n } from "@/i18n/i18n";

// Types
interface WireGuardPeer {
  inbound_id: string;
  user_id: string;
  public_key: string;
  address: string;
  mtu: number;
  dns: string;
  last_handshake?: string;
  tx_bytes: number;
  rx_bytes: number;
}

interface RepairReport {
  inbound_id: string;
  duplicates: number;
  out_of_range: number;
  reassigned: { user_id: string; old_address: string; new_address: string }[];
}

interface WireGuardMesh {
  id: string;
  name: string;
  cidr: string;
  peers: MeshPeer[];
  created_at: string;
}

interface MeshPeer {
  id: string;
  mesh_id: string;
  node_id: string;
  public_key: string;
  endpoint: string;
  address: string;
  keepalive: number;
}

function formatBytes(bytes: number): string {
  if (bytes === 0) return "0 B";
  const units = ["B", "KB", "MB", "GB", "TB"];
  const i = Math.floor(Math.log(bytes) / Math.log(1024));
  return (bytes / Math.pow(1024, i)).toFixed(1) + " " + units[i];
}

export function WireGuard() {
  const { t } = useI18n();
  useTitle(t("wireguard.title", "WireGuard Management"));
  const queryClient = useQueryClient();
  const toast = useToast();
  const confirm = useConfirm();

  const [selectedInbound, setSelectedInbound] = useState("");
  const [showSettings, setShowSettings] = useState<WireGuardPeer | null>(null);
  const [showQR, setShowQR] = useState<string | null>(null);
  const [showMeshForm, setShowMeshForm] = useState(false);
  const [activeTab, setActiveTab] = useState<"peers" | "mesh">("peers");

  // Fetch peers
  const { data: peers, isLoading: peersLoading } = useQuery({
    queryKey: ["wg-peers", selectedInbound],
    queryFn: () =>
      api.get(`/api/v2/wireguard/${selectedInbound}/peers`).then((r) => r.data as WireGuardPeer[]),
    enabled: !!selectedInbound,
  });

  // Fetch meshes
  const { data: meshes } = useQuery({
    queryKey: ["wg-meshes"],
    queryFn: () => api.get("/api/v2/wireguard/mesh").then((r) => r.data as WireGuardMesh[]),
  });

  // Repair mutation
  const repairMutation = useMutation({
    mutationFn: () => api.post(`/api/v2/wireguard/${selectedInbound}/repair`),
    onSuccess: (res) => {
      const report = res.data as RepairReport;
      queryClient.invalidateQueries({ queryKey: ["wg-peers", selectedInbound] });
      toast.success(
        t("wireguard.repairDone", "Repair complete: {{d}} duplicates, {{o}} out-of-range fixed")
          .replace("{{d}}", String(report.duplicates))
          .replace("{{o}}", String(report.out_of_range))
      );
    },
    onError: () => toast.error(t("wireguard.repairFailed", "Repair failed")),
  });

  // Update peer settings
  const updateSettingsMutation = useMutation({
    mutationFn: (data: { userID: string; mtu: number; dns: string }) =>
      api.put(`/api/v2/wireguard/${selectedInbound}/peers/${data.userID}`, {
        mtu: data.mtu,
        dns: data.dns,
      }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["wg-peers", selectedInbound] });
      setShowSettings(null);
      toast.success(t("wireguard.settingsUpdated", "Peer settings updated"));
    },
  });

  // QR code fetch
  const fetchQR = async (userID: string) => {
    try {
      const endpoint = prompt(t("wireguard.enterEndpoint", "Enter server endpoint (host:port):"));
      if (!endpoint) return;
      const res = await api.get(
        `/api/v2/wireguard/${selectedInbound}/peers/${userID}/qr?endpoint=${encodeURIComponent(endpoint)}`,
        { responseType: "blob" }
      );
      const url = URL.createObjectURL(res.data);
      setShowQR(url);
    } catch {
      toast.error(t("wireguard.qrFailed", "Failed to generate QR code"));
    }
  };

  return (
    <div className="space-y-6 p-6">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-bold flex items-center gap-2">
          <Shield className="w-6 h-6" />
          {t("wireguard.title", "WireGuard Management")}
        </h1>
      </div>

      {/* Tabs */}
      <div className="flex gap-2 border-b pb-2">
        <Button
          variant={activeTab === "peers" ? "default" : "ghost"}
          size="sm"
          onClick={() => setActiveTab("peers")}
        >
          {t("wireguard.peers", "Peers")}
        </Button>
        <Button
          variant={activeTab === "mesh" ? "default" : "ghost"}
          size="sm"
          onClick={() => setActiveTab("mesh")}
        >
          <Network className="w-4 h-4 mr-1" />
          {t("wireguard.mesh", "Mesh")}
        </Button>
      </div>

      {activeTab === "peers" && (
        <>
          {/* Inbound selector */}
          <GlassCard className="p-4">
            <div className="flex items-center gap-4">
              <div className="flex-1">
                <label className="block text-sm font-medium mb-1">
                  {t("wireguard.selectInbound", "WireGuard Inbound ID")}
                </label>
                <Input
                  placeholder="UUID"
                  value={selectedInbound}
                  onChange={(e) => setSelectedInbound(e.target.value)}
                />
              </div>
              {selectedInbound && (
                <Button
                  variant="outline"
                  onClick={async () => {
                    const ok = await confirm(
                      t("wireguard.confirmRepair", "Repair will fix duplicate/out-of-range IPs. Continue?")
                    );
                    if (ok) repairMutation.mutate();
                  }}
                  disabled={repairMutation.isPending}
                >
                  <Wrench className="w-4 h-4 mr-1" />
                  {t("wireguard.repair", "Repair")}
                </Button>
              )}
            </div>
          </GlassCard>

          {/* Peer list */}
          {selectedInbound && (
            <GlassCard className="p-4">
              <h2 className="text-lg font-semibold mb-4">
                {t("wireguard.peerList", "Peer List")}
              </h2>
              {peersLoading ? (
                <div className="flex items-center gap-2 text-muted-foreground">
                  <RefreshCw className="w-4 h-4 animate-spin" />
                  {t("common.loading", "Loading...")}
                </div>
              ) : peers && peers.length > 0 ? (
                <div className="overflow-x-auto">
                  <table className="w-full text-sm">
                    <thead>
                      <tr className="border-b">
                        <th className="text-left py-2 px-3">IP</th>
                        <th className="text-left py-2 px-3">Public Key</th>
                        <th className="text-left py-2 px-3">MTU</th>
                        <th className="text-left py-2 px-3">DNS</th>
                        <th className="text-left py-2 px-3">Last Handshake</th>
                        <th className="text-left py-2 px-3">TX / RX</th>
                        <th className="text-right py-2 px-3">Actions</th>
                      </tr>
                    </thead>
                    <tbody>
                      {peers.map((peer) => (
                        <tr key={peer.user_id} className="border-b hover:bg-muted/50">
                          <td className="py-2 px-3 font-mono">
                            <Badge variant="secondary">{peer.address}</Badge>
                          </td>
                          <td className="py-2 px-3 font-mono text-xs truncate max-w-[160px]">
                            {peer.public_key}
                          </td>
                          <td className="py-2 px-3">{peer.mtu}</td>
                          <td className="py-2 px-3">{peer.dns}</td>
                          <td className="py-2 px-3">
                            {peer.last_handshake
                              ? new Date(peer.last_handshake).toLocaleString()
                              : "—"}
                          </td>
                          <td className="py-2 px-3 text-xs">
                            <span className="text-green-600">{formatBytes(peer.tx_bytes)}</span>
                            {" / "}
                            <span className="text-blue-600">{formatBytes(peer.rx_bytes)}</span>
                          </td>
                          <td className="py-2 px-3 text-right space-x-1">
                            <Button
                              variant="ghost"
                              size="sm"
                              onClick={() => setShowSettings(peer)}
                              title="Settings"
                            >
                              <Settings className="w-3 h-3" />
                            </Button>
                            <Button
                              variant="ghost"
                              size="sm"
                              onClick={() => fetchQR(peer.user_id)}
                              title="QR Code"
                            >
                              <QrCode className="w-3 h-3" />
                            </Button>
                          </td>
                        </tr>
                      ))}
                    </tbody>
                  </table>
                </div>
              ) : (
                <p className="text-muted-foreground text-sm">
                  {t("wireguard.noPeers", "No peers allocated yet.")}
                </p>
              )}
            </GlassCard>
          )}
        </>
      )}

      {activeTab === "mesh" && (
        <GlassCard className="p-4 space-y-4">
          <div className="flex items-center justify-between">
            <h2 className="text-lg font-semibold">
              {t("wireguard.meshNetworks", "Mesh Networks")}
            </h2>
            <Button size="sm" onClick={() => setShowMeshForm(true)}>
              <Plus className="w-4 h-4 mr-1" />
              {t("wireguard.createMesh", "Create Mesh")}
            </Button>
          </div>

          {meshes && meshes.length > 0 ? (
            <div className="space-y-3">
              {meshes.map((mesh) => (
                <div key={mesh.id} className="border rounded p-3">
                  <div className="flex items-center justify-between">
                    <div>
                      <span className="font-medium">{mesh.name}</span>
                      <Badge variant="outline" className="ml-2">{mesh.cidr}</Badge>
                    </div>
                    <span className="text-xs text-muted-foreground">
                      {mesh.peers?.length || 0} nodes
                    </span>
                  </div>
                  {mesh.peers && mesh.peers.length > 0 && (
                    <div className="mt-2 text-xs space-y-1">
                      {mesh.peers.map((p) => (
                        <div key={p.id} className="flex items-center gap-3 text-muted-foreground">
                          <span className="font-mono">{p.address}</span>
                          <span>{p.endpoint || "—"}</span>
                          <span className="truncate max-w-[120px]">{p.public_key}</span>
                        </div>
                      ))}
                    </div>
                  )}
                </div>
              ))}
            </div>
          ) : (
            <p className="text-muted-foreground text-sm">
              {t("wireguard.noMeshes", "No mesh networks configured.")}
            </p>
          )}
        </GlassCard>
      )}

      {/* Peer Settings Modal */}
      <Modal
        open={!!showSettings}
        onClose={() => setShowSettings(null)}
        title={t("wireguard.peerSettings", "Peer Settings")}
      >
        {showSettings && (
          <PeerSettingsForm
            peer={showSettings}
            onSave={(mtu, dns) =>
              updateSettingsMutation.mutate({ userID: showSettings.user_id, mtu, dns })
            }
            isPending={updateSettingsMutation.isPending}
          />
        )}
      </Modal>

      {/* QR Code Modal */}
      <Modal open={!!showQR} onClose={() => { if (showQR) URL.revokeObjectURL(showQR); setShowQR(null); }} title="WireGuard QR Code">
        {showQR && (
          <div className="flex flex-col items-center gap-4">
            <img src={showQR} alt="WireGuard QR" className="w-64 h-64" />
            <a href={showQR} download="wireguard-qr.png">
              <Button variant="outline" size="sm">
                {t("common.download", "Download")}
              </Button>
            </a>
          </div>
        )}
      </Modal>

      {/* Create Mesh Modal */}
      <Modal
        open={showMeshForm}
        onClose={() => setShowMeshForm(false)}
        title={t("wireguard.createMesh", "Create Mesh Network")}
      >
        <CreateMeshForm
          onSuccess={() => {
            setShowMeshForm(false);
            queryClient.invalidateQueries({ queryKey: ["wg-meshes"] });
            toast.success(t("wireguard.meshCreated", "Mesh network created"));
          }}
          onError={(msg) => toast.error(msg)}
        />
      </Modal>
    </div>
  );
}

// --- Sub-components ---

function PeerSettingsForm({
  peer,
  onSave,
  isPending,
}: {
  peer: WireGuardPeer;
  onSave: (mtu: number, dns: string) => void;
  isPending: boolean;
}) {
  const { t } = useI18n();
  const [mtu, setMtu] = useState(peer.mtu);
  const [dns, setDns] = useState(peer.dns);

  return (
    <div className="space-y-4">
      <div>
        <label className="block text-sm font-medium mb-1">MTU (1280-1500)</label>
        <Input
          type="number"
          min={1280}
          max={1500}
          value={mtu}
          onChange={(e) => setMtu(Number(e.target.value))}
        />
      </div>
      <div>
        <label className="block text-sm font-medium mb-1">DNS</label>
        <Input value={dns} onChange={(e) => setDns(e.target.value)} placeholder="1.1.1.1" />
      </div>
      <Button onClick={() => onSave(mtu, dns)} disabled={isPending}>
        {t("common.save", "Save")}
      </Button>
    </div>
  );
}

function CreateMeshForm({
  onSuccess,
  onError,
}: {
  onSuccess: () => void;
  onError: (msg: string) => void;
}) {
  const { t } = useI18n();
  const [name, setName] = useState("");
  const [cidr, setCidr] = useState("10.10.0.0/16");
  const [nodeIds, setNodeIds] = useState("");
  const [endpoints, setEndpoints] = useState("");

  const createMutation = useMutation({
    mutationFn: () => {
      const ids = nodeIds.split("\n").map((s) => s.trim()).filter(Boolean);
      const eps: Record<string, string> = {};
      endpoints.split("\n").forEach((line) => {
        const [id, ep] = line.split("=").map((s) => s.trim());
        if (id && ep) eps[id] = ep;
      });
      return api.post("/api/v2/wireguard/mesh", {
        name,
        cidr,
        node_ids: ids,
        endpoints: eps,
      });
    },
    onSuccess: () => onSuccess(),
    onError: (err: Error) => onError(err.message),
  });

  return (
    <div className="space-y-4">
      <div>
        <label className="block text-sm font-medium mb-1">
          {t("wireguard.meshName", "Mesh Name")}
        </label>
        <Input value={name} onChange={(e) => setName(e.target.value)} placeholder="site-to-site" />
      </div>
      <div>
        <label className="block text-sm font-medium mb-1">CIDR</label>
        <Input value={cidr} onChange={(e) => setCidr(e.target.value)} placeholder="10.10.0.0/16" />
      </div>
      <div>
        <label className="block text-sm font-medium mb-1">
          {t("wireguard.nodeIds", "Node IDs (one per line)")}
        </label>
        <textarea
          className="w-full h-24 rounded border p-2 font-mono text-sm bg-background"
          value={nodeIds}
          onChange={(e) => setNodeIds(e.target.value)}
          placeholder={"uuid-1\nuuid-2\nuuid-3"}
        />
      </div>
      <div>
        <label className="block text-sm font-medium mb-1">
          {t("wireguard.endpoints", "Endpoints (node_id=host:port, one per line)")}
        </label>
        <textarea
          className="w-full h-24 rounded border p-2 font-mono text-sm bg-background"
          value={endpoints}
          onChange={(e) => setEndpoints(e.target.value)}
          placeholder={"uuid-1=1.2.3.4:51820\nuuid-2=5.6.7.8:51820"}
        />
      </div>
      <Button onClick={() => createMutation.mutate()} disabled={!name || createMutation.isPending}>
        <Plus className="w-4 h-4 mr-1" />
        {t("wireguard.createMesh", "Create Mesh")}
      </Button>
    </div>
  );
}
