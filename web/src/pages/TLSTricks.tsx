import { useState } from "react";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { api } from "@/api/client";
import { Button, Card, Input, PageHeader, Badge, Select } from "@/components/ui";
import { Modal } from "@/components/Modal";
import { useToast } from "@/components/toast";
import { useConfirm } from "@/components/confirm";

interface TLSProfile {
  id: string; name: string; isp: string; fragment_enabled: boolean; fragment_size: string;
  fragment_interval: string; fragment_packets: string; mux_enabled: boolean; mux_concurrency: number;
  utls_fingerprint: string; padding_enabled: boolean; padding_size: string; ech_enabled: boolean;
  auto_detect: boolean; enabled: boolean;
}

export function TLSTricks() {
  const qc = useQueryClient(); const toast = useToast(); const confirm = useConfirm();
  const [createOpen, setCreateOpen] = useState(false);

  const { data } = useQuery({ queryKey: ["tls-tricks"], queryFn: () => api<{ profiles: TLSProfile[] }>("/api/tls-tricks") });
  const { data: presetsData } = useQuery({ queryKey: ["tls-presets"], queryFn: () => api<{ presets: any[] }>("/api/tls-tricks/presets") });

  const del = useMutation({ mutationFn: (id: string) => api<void>(`/api/tls-tricks/${id}`, { method: "DELETE" }), onSuccess: () => qc.invalidateQueries({ queryKey: ["tls-tricks"] }) });
  const fromPreset = useMutation({ mutationFn: (isp: string) => api("/api/tls-tricks/preset", { method: "POST", body: { isp } }), onSuccess: () => { qc.invalidateQueries({ queryKey: ["tls-tricks"] }); toast.success("Profile created from preset"); } });

  return (
    <div className="space-y-6 animate-fade-in">
      <div className="flex items-center justify-between">
        <PageHeader title="TLS Tricks" subtitle="ISP-specific fragment, mux, padding, and ECH profiles" />
        <Button onClick={() => setCreateOpen(true)}>Custom Profile</Button>
      </div>

      {/* Quick presets */}
      <Card>
        <h3 className="text-sm font-bold text-fg mb-3">Quick Presets (One-Click)</h3>
        <div className="flex flex-wrap gap-2">
          {presetsData?.presets?.map((p: any) => (
            <Button key={p.isp} variant="outline" size="sm" onClick={() => fromPreset.mutate(p.isp)} disabled={fromPreset.isPending}>
              {p.name}
            </Button>
          ))}
        </div>
      </Card>

      <CreateProfileModal open={createOpen} onClose={() => setCreateOpen(false)} />

      <div className="grid grid-cols-1 gap-4 lg:grid-cols-2">
        {data?.profiles?.map(p => (
          <Card key={p.id} className="space-y-3">
            <div className="flex items-center justify-between">
              <h3 className="text-sm font-bold text-fg">{p.name}</h3>
              <span className={`h-2 w-2 rounded-full ${p.enabled ? "bg-success" : "bg-fg-subtle"}`} />
            </div>
            <div className="grid grid-cols-2 gap-1 text-xs text-fg-muted">
              <div>Fragment: <strong className="text-fg">{p.fragment_enabled ? p.fragment_size : "off"}</strong></div>
              <div>Mux: <strong className="text-fg">{p.mux_enabled ? p.mux_concurrency + "x" : "off"}</strong></div>
              <div>uTLS: <strong className="text-fg">{p.utls_fingerprint}</strong></div>
              <div>Padding: <strong className="text-fg">{p.padding_enabled ? p.padding_size : "off"}</strong></div>
              <div>ECH: <strong className="text-fg">{p.ech_enabled ? "on" : "off"}</strong></div>
              <div>ISP: <Badge color="muted">{p.isp}</Badge></div>
            </div>
            <div className="flex justify-end"><Button variant="ghost" size="sm" className="text-destructive text-xs" onClick={async () => { if (await confirm({ title: "Delete?", destructive: true })) del.mutate(p.id); }}>Delete</Button></div>
          </Card>
        ))}
        {(!data?.profiles || data.profiles.length === 0) && <p className="col-span-full text-center text-sm text-fg-muted py-8">No profiles. Use a preset or create a custom one.</p>}
      </div>
    </div>
  );
}

function CreateProfileModal({ open, onClose }: { open: boolean; onClose: () => void }) {
  const qc = useQueryClient(); const toast = useToast();
  const [f, setF] = useState({ name: "", isp: "custom", fragment_enabled: true, fragment_size: "10-50", fragment_interval: "10-20", fragment_packets: "tlshello", mux_enabled: true, mux_concurrency: 8, utls_fingerprint: "chrome", padding_enabled: true, padding_size: "100-200", ech_enabled: false, enabled: true });
  const create = useMutation({ mutationFn: (b: any) => api("/api/tls-tricks", { method: "POST", body: b }), onSuccess: () => { qc.invalidateQueries({ queryKey: ["tls-tricks"] }); onClose(); toast.success("Created"); } });
  return (
    <Modal open={open} onClose={onClose} title="Custom TLS Profile" className="max-w-lg">
      <form onSubmit={e => { e.preventDefault(); create.mutate(f); }} className="space-y-3">
        <Input placeholder="Profile name" value={f.name} onChange={e => setF(s => ({...s, name: e.target.value}))} required />
        <div className="grid grid-cols-2 gap-2">
          <div><label className="text-xs text-fg-subtle">Fragment size</label><Input value={f.fragment_size} onChange={e => setF(s => ({...s, fragment_size: e.target.value}))} /></div>
          <div><label className="text-xs text-fg-subtle">Fragment interval</label><Input value={f.fragment_interval} onChange={e => setF(s => ({...s, fragment_interval: e.target.value}))} /></div>
          <div><label className="text-xs text-fg-subtle">uTLS fingerprint</label><Select value={f.utls_fingerprint} onChange={e => setF(s => ({...s, utls_fingerprint: e.target.value}))}><option value="chrome">Chrome</option><option value="firefox">Firefox</option><option value="safari">Safari</option><option value="random">Random</option></Select></div>
          <div><label className="text-xs text-fg-subtle">Mux concurrency</label><Input value={f.mux_concurrency} onChange={e => setF(s => ({...s, mux_concurrency: Number(e.target.value)}))} inputMode="numeric" /></div>
          <div><label className="text-xs text-fg-subtle">Padding size</label><Input value={f.padding_size} onChange={e => setF(s => ({...s, padding_size: e.target.value}))} /></div>
        </div>
        <div className="flex flex-wrap gap-3">
          <label className="flex items-center gap-1 text-xs"><input type="checkbox" checked={f.fragment_enabled} onChange={e => setF(s => ({...s, fragment_enabled: e.target.checked}))} /> Fragment</label>
          <label className="flex items-center gap-1 text-xs"><input type="checkbox" checked={f.mux_enabled} onChange={e => setF(s => ({...s, mux_enabled: e.target.checked}))} /> Mux</label>
          <label className="flex items-center gap-1 text-xs"><input type="checkbox" checked={f.padding_enabled} onChange={e => setF(s => ({...s, padding_enabled: e.target.checked}))} /> Padding</label>
          <label className="flex items-center gap-1 text-xs"><input type="checkbox" checked={f.ech_enabled} onChange={e => setF(s => ({...s, ech_enabled: e.target.checked}))} /> ECH</label>
        </div>
        <div className="flex justify-end gap-2 pt-2"><Button type="button" variant="ghost" onClick={onClose}>Cancel</Button><Button type="submit" disabled={create.isPending}>Create</Button></div>
      </form>
    </Modal>
  );
}
