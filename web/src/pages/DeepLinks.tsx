import { useState } from "react";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { api } from "@/api/client";
import { Button, Card, Input, PageHeader } from "@/components/ui";
import { useToast } from "@/components/toast";

interface DLConfig { enabled: boolean; scheme: string; base_url: string; app_store_url: string; play_store_url: string; qr_logo_url: string; }

export function DeepLinks() {
  const qc = useQueryClient(); const toast = useToast();
  const [form, setForm] = useState<DLConfig | null>(null);
  const [testToken, setTestToken] = useState("");
  const [testResult, setTestResult] = useState("");

  const { data } = useQuery({ queryKey: ["deeplink-config"], queryFn: () => api<{ config: DLConfig }>("/api/deeplink/config") });
  const cfg = form ?? data?.config;
  const save = useMutation({ mutationFn: (c: DLConfig) => api("/api/deeplink/config", { method: "PUT", body: c }), onSuccess: () => { qc.invalidateQueries({ queryKey: ["deeplink-config"] }); toast.success("Saved"); } });

  function update(field: keyof DLConfig, value: any) { setForm(prev => ({ ...(prev ?? data?.config ?? {} as any), [field]: value })); }

  async function generate() {
    if (!testToken) return;
    const res = await api<{ deep_link: string }>("/api/deeplink/generate", { query: { token: testToken } });
    setTestResult(res.deep_link);
  }

  return (
    <div className="space-y-6 animate-fade-in">
      <PageHeader title="Deep Links & QR" subtitle="One-tap subscription import via custom URL scheme" />
      {cfg && (
        <Card className="space-y-4">
          <div className="flex items-center justify-between"><h3 className="text-sm font-bold text-fg">Configuration</h3>
            <label className="flex items-center gap-2 text-sm"><input type="checkbox" checked={cfg.enabled} onChange={e => update("enabled", e.target.checked)} className="rounded" /> Enabled</label>
          </div>
          <div className="grid grid-cols-2 gap-3">
            <div><label className="text-xs text-fg-subtle">URL Scheme</label><Input value={cfg.scheme} onChange={e => update("scheme", e.target.value)} placeholder="vortex" /></div>
            <div><label className="text-xs text-fg-subtle">Base URL</label><Input value={cfg.base_url} onChange={e => update("base_url", e.target.value)} placeholder="https://panel.example.com" /></div>
            <div><label className="text-xs text-fg-subtle">App Store URL</label><Input value={cfg.app_store_url} onChange={e => update("app_store_url", e.target.value)} /></div>
            <div><label className="text-xs text-fg-subtle">Play Store URL</label><Input value={cfg.play_store_url} onChange={e => update("play_store_url", e.target.value)} /></div>
          </div>
          <div className="flex justify-end"><Button onClick={() => cfg && save.mutate(cfg)} disabled={save.isPending}>Save</Button></div>
        </Card>
      )}
      <Card className="space-y-3">
        <h3 className="text-sm font-bold text-fg">Test Deep Link</h3>
        <div className="flex gap-2">
          <Input placeholder="Sub token..." value={testToken} onChange={e => setTestToken(e.target.value)} className="flex-1" />
          <Button onClick={generate} disabled={!testToken}>Generate</Button>
        </div>
        {testResult && <div className="rounded-lg bg-surface-2 p-3 font-mono text-xs break-all text-fg">{testResult}</div>}
      </Card>
    </div>
  );
}
