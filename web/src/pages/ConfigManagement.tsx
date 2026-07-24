import { useState } from "react";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { RefreshCw, Download, RotateCcw, CheckCircle, GitCompare } from "lucide-react";
import { api } from "@/api/client";
import { Button, Input, Badge } from "@/components/ui";
import { GlassCard } from "@/components/veltrix";
import { Modal } from "@/components/Modal";
import { useToast } from "@/components/toast";
import { useConfirm } from "@/components/confirm";
import { useTitle } from "@/lib/useTitle";
import { useI18n } from "@/i18n/i18n";

interface ConfigVersion { id: string; inbound_id: string; version: number; config_data: Record<string, unknown>; comment: string; created_at: string; }
interface ConfigValidationError { field: string; message: string; }
interface ConfigChange { path: string; old_value: unknown; new_value: unknown; type: "added" | "removed" | "modified"; }
interface ConfigDiff { inbound_id: string; old_version: number; new_version: number; changes: ConfigChange[]; }

const protocols = ["vmess","vless","trojan","shadowsocks","hysteria2","tuic","wireguard","shadowtls","naive"];
const networks = ["tcp","ws","grpc","httpupgrade","xhttp"];
const securities = ["none","tls","reality"];

export function ConfigManagement() {
  const { t: _t } = useI18n();
  useTitle("Config Management");
  const qc = useQueryClient();
  const toast = useToast();
  const confirm = useConfirm();

  const [selectedInbound, setSelectedInbound] = useState("");
  const [showDiff, setShowDiff] = useState(false);
  const [diffData, setDiffData] = useState<ConfigDiff | null>(null);
  const [vForm, setVForm] = useState({ protocol: "vless", network: "tcp", security: "reality", config: "{}" });

  const { data: versions } = useQuery({
    queryKey: ["config-versions", selectedInbound],
    queryFn: () => api<ConfigVersion[]>(`/api/v2/inbounds/${selectedInbound}/versions`),
    enabled: !!selectedInbound,
  });

  const validateMut = useMutation({
    mutationFn: (data: { protocol: string; network: string; security: string; config: Record<string, unknown> }) =>
      api<{ valid: boolean; errors?: ConfigValidationError[] }>(`/api/v2/inbounds/${selectedInbound}/validate`, { method: "POST", body: data }),
    onSuccess: (res) => { if (res.valid) toast.success("Config is valid"); },
  });

  const defaultsMut = useMutation({
    mutationFn: () => api<{ config: Record<string, unknown> }>(`/api/v2/inbounds/${selectedInbound}/defaults`, { query: { protocol: vForm.protocol, network: vForm.network, security: vForm.security } }),
    onSuccess: (res) => { setVForm((f) => ({ ...f, config: JSON.stringify(res.config, null, 2) })); toast.success("Defaults loaded"); },
  });

  const rollbackMut = useMutation({
    mutationFn: (version: number) => api(`/api/v2/inbounds/${selectedInbound}/rollback`, { method: "POST", body: { version } }),
    onSuccess: () => { qc.invalidateQueries({ queryKey: ["config-versions", selectedInbound] }); toast.success("Rollback done"); },
  });

  const handleValidate = () => {
    try {
      const config = JSON.parse(vForm.config);
      validateMut.mutate({ protocol: vForm.protocol, network: vForm.network, security: vForm.security, config });
    } catch { toast.error("Invalid JSON"); }
  };

  const handleDiff = async (oldV: number, newV: number) => {
    try {
      const d = await api<ConfigDiff>(`/api/v2/inbounds/${selectedInbound}/diff`, { query: { old_version: String(oldV), new_version: String(newV) } });
      setDiffData(d); setShowDiff(true);
    } catch { toast.error("Diff failed"); }
  };

  return (
    <div className="space-y-6 p-6">
      <h1 className="text-2xl font-bold">Config Management</h1>

      <GlassCard className="p-4">
        <label className="block text-sm font-medium mb-2">Inbound ID</label>
        <Input placeholder="UUID" value={selectedInbound} onChange={(e) => setSelectedInbound(e.target.value)} />
      </GlassCard>

      {selectedInbound && (
        <>
          <GlassCard className="p-4 space-y-4">
            <h2 className="text-lg font-semibold flex items-center gap-2"><CheckCircle className="w-5 h-5" />Validate</h2>
            <div className="grid grid-cols-3 gap-4">
              <select className="field input-surface" value={vForm.protocol} onChange={(e) => setVForm((f) => ({ ...f, protocol: e.target.value }))}>{protocols.map((p) => <option key={p} value={p}>{p}</option>)}</select>
              <select className="field input-surface" value={vForm.network} onChange={(e) => setVForm((f) => ({ ...f, network: e.target.value }))}>{networks.map((n) => <option key={n} value={n}>{n}</option>)}</select>
              <select className="field input-surface" value={vForm.security} onChange={(e) => setVForm((f) => ({ ...f, security: e.target.value }))}>{securities.map((s) => <option key={s} value={s}>{s}</option>)}</select>
            </div>
            <textarea className="w-full h-40 field input-surface font-mono text-sm" value={vForm.config} onChange={(e) => setVForm((f) => ({ ...f, config: e.target.value }))} />
            <div className="flex gap-2">
              <Button onClick={handleValidate}><CheckCircle className="w-4 h-4 mr-1" />Validate</Button>
              <Button variant="outline" onClick={() => defaultsMut.mutate()}><RefreshCw className="w-4 h-4 mr-1" />Load Defaults</Button>
            </div>
            {validateMut.data && !validateMut.data.valid && (
              <div className="mt-2 p-3 bg-danger/10 rounded">
                <ul className="list-disc list-inside text-sm">{validateMut.data.errors?.map((e, i) => <li key={i}><code>{e.field}</code>: {e.message}</li>)}</ul>
              </div>
            )}
          </GlassCard>

          <GlassCard className="p-4 space-y-4">
            <div className="flex items-center justify-between">
              <h2 className="text-lg font-semibold">Versions</h2>
              <Button variant="outline" size="sm" onClick={() => window.open(`/api/v2/inbounds/${selectedInbound}/export`)}><Download className="w-4 h-4 mr-1" />Export</Button>
            </div>
            {versions && versions.length > 0 ? (
              <table className="w-full text-sm">
                <thead><tr className="border-b"><th className="text-left py-2 px-3">Ver</th><th className="text-left py-2 px-3">Comment</th><th className="text-left py-2 px-3">Created</th><th className="text-right py-2 px-3">Actions</th></tr></thead>
                <tbody>{versions.map((v, i) => (
                  <tr key={v.id} className="border-b hover:bg-surface-2/50">
                    <td className="py-2 px-3"><Badge>v{v.version}</Badge></td>
                    <td className="py-2 px-3">{v.comment || "—"}</td>
                    <td className="py-2 px-3">{new Date(v.created_at).toLocaleString()}</td>
                    <td className="py-2 px-3 text-right space-x-1">
                      {i < versions.length - 1 && <Button variant="ghost" size="sm" onClick={() => handleDiff(versions[i+1].version, v.version)}><GitCompare className="w-3 h-3" /></Button>}
                      <Button variant="ghost" size="sm" onClick={async () => { if (await confirm({ title: `Rollback to v${v.version}?` })) rollbackMut.mutate(v.version); }}><RotateCcw className="w-3 h-3" /></Button>
                    </td>
                  </tr>
                ))}</tbody>
              </table>
            ) : <p className="text-fg-muted text-sm">No versions yet.</p>}
          </GlassCard>
        </>
      )}

      <Modal open={showDiff} onClose={() => setShowDiff(false)} title="Config Diff">
        {diffData && (
          <div className="space-y-2 max-h-96 overflow-auto">
            <p className="text-sm text-fg-muted">v{diffData.old_version} → v{diffData.new_version}</p>
            {diffData.changes.length === 0 ? <p className="text-sm">No changes</p> : diffData.changes.map((c, i) => (
              <div key={i} className={`p-2 rounded text-sm font-mono ${c.type === "added" ? "bg-success/10 text-success" : c.type === "removed" ? "bg-danger/10 text-danger" : "bg-warning/10 text-warning"}`}>
                <strong>{c.path}</strong> {c.type === "modified" && <><div>- {JSON.stringify(c.old_value)}</div><div>+ {JSON.stringify(c.new_value)}</div></>}
                {c.type === "added" && <div>+ {JSON.stringify(c.new_value)}</div>}
                {c.type === "removed" && <div>- {JSON.stringify(c.old_value)}</div>}
              </div>
            ))}
          </div>
        )}
      </Modal>
    </div>
  );
}
