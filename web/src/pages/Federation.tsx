import { useState } from "react";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { api } from "@/api/client";
import { Button, Card, Input, PageHeader, Badge } from "@/components/ui";
import { Modal } from "@/components/Modal";
import { useToast } from "@/components/toast";
import { useI18n } from "@/i18n/i18n";

interface FedConfig { enabled: boolean; cluster_name: string; sso_enabled: boolean; sync_interval: number; }
interface Peer { id: string; name: string; endpoint: string; status: string; sync_users: boolean; sync_nodes: boolean; last_sync: string | null; }
interface SyncEvent { id: string; peer_name: string; direction: string; entity_type: string; count: number; status: string; created_at: string; }

export function Federation() {
  const { t } = useI18n();
  const qc = useQueryClient(); const toast = useToast();
  const [addOpen, setAddOpen] = useState(false);
  const [form, setForm] = useState<FedConfig | null>(null);

  const { data: cfgData } = useQuery({ queryKey: ["fed-config"], queryFn: () => api<{ config: FedConfig }>("/api/federation/config") });
  const { data: peersData } = useQuery({ queryKey: ["fed-peers"], queryFn: () => api<{ peers: Peer[] }>("/api/federation/peers") });
  const { data: eventsData } = useQuery({ queryKey: ["fed-events"], queryFn: () => api<{ events: SyncEvent[] }>("/api/federation/events") });

  const cfg = form ?? cfgData?.config;
  const saveCfg = useMutation({ mutationFn: (c: FedConfig) => api("/api/federation/config", { method: "PUT", body: c }), onSuccess: () => { qc.invalidateQueries({ queryKey: ["fed-config"] }); toast.success("Saved"); } });
  const delPeer = useMutation({ mutationFn: (id: string) => api<void>(`/api/federation/peers/${id}`, { method: "DELETE" }), onSuccess: () => qc.invalidateQueries({ queryKey: ["fed-peers"] }) });

  function update(field: keyof FedConfig, value: any) { setForm(prev => ({ ...(prev ?? cfgData?.config ?? {} as any), [field]: value })); }

  return (
    <div className="space-y-6 animate-fade-in">
      <PageHeader title={t("federation.title")} subtitle={t("federation.subtitle")} />

      <div className="rounded-lg border border-border/40 bg-surface-2/20 p-4 text-xs text-fg-muted space-y-2">
        <p className="font-medium text-fg text-sm">{t("federation.infoTitle")}</p>
        <p>{t("federation.infoDesc")}</p>
        <ul className="list-disc pl-4 space-y-1">
          <li><strong>Sync users</strong> — {t("federation.syncUsers")}</li>
          <li><strong>Sync nodes</strong> — {t("federation.syncNodes")}</li>
          <li><strong>SSO</strong> — {t("federation.sso")}</li>
        </ul>
      </div>
      {cfg && (
        <Card className="space-y-3">
          <div className="flex items-center justify-between"><h3 className="text-sm font-bold text-fg">Config</h3><label className="flex items-center gap-2 text-sm"><input type="checkbox" checked={cfg.enabled} onChange={e => update("enabled", e.target.checked)} className="rounded" /> Enabled</label></div>
          <div className="grid grid-cols-3 gap-3">
            <div><label className="text-xs text-fg-subtle">Cluster name</label><Input value={cfg.cluster_name} onChange={e => update("cluster_name", e.target.value)} /></div>
            <div><label className="text-xs text-fg-subtle">Sync interval (s)</label><Input value={cfg.sync_interval} onChange={e => update("sync_interval", Number(e.target.value))} inputMode="numeric" /></div>
            <label className="flex items-end gap-2 text-sm pb-2"><input type="checkbox" checked={cfg.sso_enabled} onChange={e => update("sso_enabled", e.target.checked)} className="rounded" /> SSO</label>
          </div>
          <div className="flex justify-end"><Button onClick={() => cfg && saveCfg.mutate(cfg)} disabled={saveCfg.isPending}>Save</Button></div>
        </Card>
      )}

      <Card>
        <div className="flex items-center justify-between mb-3"><h3 className="text-sm font-bold text-fg">{t("federation.peers")}</h3><Button size="sm" onClick={() => setAddOpen(true)}>{t("federation.addPeer")}</Button></div>
        <AddPeerModal open={addOpen} onClose={() => setAddOpen(false)} />
        <div className="space-y-2">
          {peersData?.peers?.map(p => (
            <div key={p.id} className="flex items-center justify-between rounded-lg bg-surface-2/40 px-3 py-2">
              <div className="flex items-center gap-3">
                <span className="text-sm font-medium text-fg">{p.name}</span>
                <Badge color={p.status === "connected" ? "active" : p.status === "syncing" ? "limited" : "expired"}>{p.status}</Badge>
                <span className="text-xs text-fg-muted font-mono">{p.endpoint}</span>
              </div>
              <Button variant="ghost" size="sm" className="text-destructive" onClick={() => delPeer.mutate(p.id)}>Remove</Button>
            </div>
          ))}
          {(!peersData?.peers || peersData.peers.length === 0) && <p className="text-xs text-fg-muted text-center py-4">{t("federation.noPeers")}</p>}
        </div>
      </Card>

      <Card>
        <h3 className="text-sm font-bold text-fg mb-3">{t("federation.syncEvents")}</h3>
        <div className="space-y-2 max-h-[200px] overflow-y-auto">
          {eventsData?.events?.map(e => (
            <div key={e.id} className="flex items-center justify-between text-xs rounded-lg bg-surface-2/40 px-3 py-2">
              <span className="text-fg">{e.peer_name} — {e.direction} {e.entity_type} ({e.count})</span>
              <Badge color={e.status === "success" ? "active" : "expired"}>{e.status}</Badge>
            </div>
          ))}
        </div>
      </Card>
    </div>
  );
}

function AddPeerModal({ open, onClose }: { open: boolean; onClose: () => void }) {
  const qc = useQueryClient(); const toast = useToast();
  const [f, setF] = useState({ name: "", endpoint: "", api_key: "", sync_users: true, sync_nodes: true });
  const create = useMutation({ mutationFn: (b: any) => api("/api/federation/peers", { method: "POST", body: b }), onSuccess: () => { qc.invalidateQueries({ queryKey: ["fed-peers"] }); onClose(); toast.success("Peer added"); } });
  return (
    <Modal open={open} onClose={onClose} title="Add Peer Panel">
      <form onSubmit={e => { e.preventDefault(); create.mutate(f); }} className="space-y-3">
        <Input placeholder="Peer name" value={f.name} onChange={e => setF(s => ({...s, name: e.target.value}))} required />
        <Input placeholder="https://panel2.example.com" value={f.endpoint} onChange={e => setF(s => ({...s, endpoint: e.target.value}))} required />
        <Input placeholder="API Key" value={f.api_key} onChange={e => setF(s => ({...s, api_key: e.target.value}))} type="password" />
        <div className="flex gap-4">
          <label className="flex items-center gap-2 text-sm"><input type="checkbox" checked={f.sync_users} onChange={e => setF(s => ({...s, sync_users: e.target.checked}))} /> Sync users</label>
          <label className="flex items-center gap-2 text-sm"><input type="checkbox" checked={f.sync_nodes} onChange={e => setF(s => ({...s, sync_nodes: e.target.checked}))} /> Sync nodes</label>
        </div>
        <div className="flex justify-end gap-2"><Button type="button" variant="ghost" onClick={onClose}>Cancel</Button><Button type="submit" disabled={create.isPending}>Add</Button></div>
      </form>
    </Modal>
  );
}
