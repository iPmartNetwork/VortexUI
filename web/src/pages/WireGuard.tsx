import { useState } from "react";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { Shield, Network, Settings, Wrench } from "lucide-react";
import { api } from "@/api/client";
import { Button, Input, Badge } from "@/components/ui";
import { GlassCard } from "@/components/veltrix";
import { Modal } from "@/components/Modal";
import { useToast } from "@/components/toast";
import { useConfirm } from "@/components/confirm";
import { useTitle } from "@/lib/useTitle";
import { useI18n } from "@/i18n/i18n";

interface WGPeer { inbound_id: string; user_id: string; public_key: string; address: string; mtu: number; dns: string; last_handshake?: string; tx_bytes: number; rx_bytes: number; }
interface WGMesh { id: string; name: string; cidr: string; peers: { id: string; node_id: string; public_key: string; endpoint: string; address: string; }[]; }

function fmtBytes(b: number): string { if (!b) return "0 B"; const u = ["B","KB","MB","GB","TB"]; const i = Math.floor(Math.log(b)/Math.log(1024)); return (b/Math.pow(1024,i)).toFixed(1)+" "+u[i]; }

export function WireGuard() {
  const { t: _t } = useI18n();
  useTitle("WireGuard");
  const qc = useQueryClient();
  const toast = useToast();
  const confirm = useConfirm();
  const [inbound, setInbound] = useState("");
  const [showSettings, setShowSettings] = useState<WGPeer|null>(null);
  const [tab, setTab] = useState<"peers"|"mesh">("peers");

  const { data: peers } = useQuery({ queryKey: ["wg-peers", inbound], queryFn: () => api<WGPeer[]>(`/api/v2/wireguard/${inbound}/peers`), enabled: !!inbound });
  const { data: meshes } = useQuery({ queryKey: ["wg-meshes"], queryFn: () => api<WGMesh[]>("/api/v2/wireguard/mesh") });

  const repairMut = useMutation({
    mutationFn: () => api<{ duplicates: number; out_of_range: number }>(`/api/v2/wireguard/${inbound}/repair`, { method: "POST" }),
    onSuccess: (r) => { qc.invalidateQueries({ queryKey: ["wg-peers", inbound] }); toast.success(`Repaired: ${r.duplicates} dup, ${r.out_of_range} OOR`); },
  });

  const updateMut = useMutation({
    mutationFn: (d: { userID: string; mtu: number; dns: string }) => api(`/api/v2/wireguard/${inbound}/peers/${d.userID}`, { method: "PUT", body: { mtu: d.mtu, dns: d.dns } }),
    onSuccess: () => { qc.invalidateQueries({ queryKey: ["wg-peers", inbound] }); setShowSettings(null); toast.success("Updated"); },
  });

  return (
    <div className="space-y-6 p-6">
      <h1 className="text-2xl font-bold flex items-center gap-2"><Shield className="w-6 h-6" />WireGuard</h1>
      <div className="flex gap-2 border-b border-border pb-2">
        <Button variant={tab==="peers"?"primary":"ghost"} size="sm" onClick={()=>setTab("peers")}>Peers</Button>
        <Button variant={tab==="mesh"?"primary":"ghost"} size="sm" onClick={()=>setTab("mesh")}><Network className="w-4 h-4 mr-1"/>Mesh</Button>
      </div>

      {tab === "peers" && (
        <>
          <GlassCard className="p-4">
            <div className="flex items-center gap-4">
              <div className="flex-1"><Input placeholder="Inbound UUID" value={inbound} onChange={(e) => setInbound(e.target.value)} /></div>
              {inbound && <Button variant="outline" onClick={async () => { if (await confirm({ title: "Repair peers?" })) repairMut.mutate(); }}><Wrench className="w-4 h-4 mr-1" />Repair</Button>}
            </div>
          </GlassCard>
          {inbound && peers && peers.length > 0 && (
            <GlassCard className="p-4">
              <table className="w-full text-sm">
                <thead><tr className="border-b"><th className="text-left py-2 px-3">IP</th><th className="text-left py-2 px-3">Key</th><th className="text-left py-2 px-3">MTU</th><th className="text-left py-2 px-3">TX/RX</th><th className="text-right py-2 px-3">Actions</th></tr></thead>
                <tbody>{peers.map((p) => (
                  <tr key={p.user_id} className="border-b hover:bg-surface-2/50">
                    <td className="py-2 px-3"><Badge>{p.address}</Badge></td>
                    <td className="py-2 px-3 font-mono text-xs truncate max-w-[120px]">{p.public_key}</td>
                    <td className="py-2 px-3">{p.mtu}</td>
                    <td className="py-2 px-3 text-xs">{fmtBytes(p.tx_bytes)} / {fmtBytes(p.rx_bytes)}</td>
                    <td className="py-2 px-3 text-right"><Button variant="ghost" size="sm" onClick={() => setShowSettings(p)}><Settings className="w-3 h-3" /></Button></td>
                  </tr>
                ))}</tbody>
              </table>
            </GlassCard>
          )}
        </>
      )}

      {tab === "mesh" && (
        <GlassCard className="p-4">
          <h2 className="text-lg font-semibold mb-3">Mesh Networks</h2>
          {meshes && meshes.length > 0 ? meshes.map((m) => (
            <div key={m.id} className="border border-border rounded-xl p-3 mb-2">
              <span className="font-medium">{m.name}</span> <Badge>{m.cidr}</Badge>
              <span className="ml-2 text-xs text-fg-muted">{m.peers?.length || 0} nodes</span>
            </div>
          )) : <p className="text-fg-muted text-sm">No mesh networks.</p>}
        </GlassCard>
      )}

      <Modal open={!!showSettings} onClose={() => setShowSettings(null)} title="Peer Settings">
        {showSettings && <PeerForm peer={showSettings} onSave={(mtu, dns) => updateMut.mutate({ userID: showSettings.user_id, mtu, dns })} />}
      </Modal>
    </div>
  );
}

function PeerForm({ peer, onSave }: { peer: WGPeer; onSave: (mtu: number, dns: string) => void }) {
  const [mtu, setMtu] = useState(peer.mtu);
  const [dns, setDns] = useState(peer.dns);
  return (
    <div className="space-y-4">
      <Input type="number" min={1280} max={1500} value={mtu} onChange={(e) => setMtu(Number(e.target.value))} />
      <Input value={dns} onChange={(e) => setDns(e.target.value)} placeholder="1.1.1.1" />
      <Button onClick={() => onSave(mtu, dns)}>Save</Button>
    </div>
  );
}
