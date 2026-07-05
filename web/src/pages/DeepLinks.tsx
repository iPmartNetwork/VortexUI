import { useState } from "react";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { Link2 } from "lucide-react";
import { api } from "@/api/client";
import { Button, Input } from "@/components/ui";
import { GlassCard } from "@/components/veltrix";
import { useToast } from "@/components/toast";
import { useTitle } from "@/lib/useTitle";
import { useI18n } from "@/i18n/i18n";

interface DLConfig { enabled: boolean; scheme: string; base_url: string; app_store_url: string; play_store_url: string; qr_logo_url: string; }

export function DeepLinks() {
  const { t } = useI18n();
  useTitle(t("deepLink.title"));
  const qc = useQueryClient();
  const toast = useToast();
  const [form, setForm] = useState<DLConfig | null>(null);
  const [testToken, setTestToken] = useState("");
  const [testResult, setTestResult] = useState("");

  const { data } = useQuery({ queryKey: ["deeplink-config"], queryFn: () => api<{ config: DLConfig }>("/api/deeplink/config") });
  const cfg = form ?? data?.config;
  const save = useMutation({
    mutationFn: (c: DLConfig) => api("/api/deeplink/config", { method: "PUT", body: c }),
    onSuccess: () => { qc.invalidateQueries({ queryKey: ["deeplink-config"] }); toast.success(t("common.saved")); },
  });

  function update<K extends keyof DLConfig>(field: K, value: DLConfig[K]) {
    setForm((prev) => ({ ...(prev ?? data?.config ?? ({} as DLConfig)), [field]: value }));
  }

  async function generate() {
    if (!testToken) return;
    const res = await api<{ deep_link: string }>("/api/deeplink/generate", { query: { token: testToken } });
    setTestResult(res.deep_link);
  }

  return (
    <div className="space-y-5 animate-page-enter">
      <div>
        <h1 className="text-2xl font-bold text-fg tracking-tight">{t("deepLink.title")}</h1>
        <p className="text-sm text-fg-muted mt-1">{t("deepLink.subtitle")}</p>
      </div>

      {cfg && (
        <GlassCard hover={false} className="!p-5 space-y-4">
          <div className="flex items-center justify-between">
            <h3 className="text-sm font-bold text-fg">{t("deepLink.config")}</h3>
            <label className="flex items-center gap-2 text-sm text-fg">
              <input type="checkbox" checked={cfg.enabled} onChange={(e) => update("enabled", e.target.checked)} className="rounded" /> {t("common.enabled")}
            </label>
          </div>
          <div className="grid grid-cols-1 gap-3 sm:grid-cols-2">
            <div><label className="text-xs text-fg-subtle">{t("deepLink.urlScheme")}</label><Input value={cfg.scheme} onChange={(e) => update("scheme", e.target.value)} placeholder="vortex" /></div>
            <div><label className="text-xs text-fg-subtle">{t("deepLink.baseUrl")}</label><Input value={cfg.base_url} onChange={(e) => update("base_url", e.target.value)} placeholder="https://panel.example.com" /></div>
            <div><label className="text-xs text-fg-subtle">{t("deepLink.appStore")}</label><Input value={cfg.app_store_url} onChange={(e) => update("app_store_url", e.target.value)} /></div>
            <div><label className="text-xs text-fg-subtle">{t("deepLink.playStore")}</label><Input value={cfg.play_store_url} onChange={(e) => update("play_store_url", e.target.value)} /></div>
          </div>
          <div className="flex justify-end pt-1 border-t border-border/40">
            <Button onClick={() => cfg && save.mutate(cfg)} disabled={save.isPending}>{t("common.save")}</Button>
          </div>
        </GlassCard>
      )}

      <GlassCard hover={false} className="!p-5 space-y-3">
        <h3 className="text-sm font-bold text-fg flex items-center gap-1.5"><Link2 size={14} className="text-primary" /> {t("deepLink.testTitle")}</h3>
        <div className="flex gap-2">
          <Input placeholder={t("deepLink.subToken")} value={testToken} onChange={(e) => setTestToken(e.target.value)} className="flex-1" />
          <Button onClick={generate} disabled={!testToken}>{t("deepLink.generate")}</Button>
        </div>
        {testResult && <div className="rounded-lg bg-surface-2/50 border border-border/40 p-3 font-mono text-xs break-all text-fg">{testResult}</div>}
      </GlassCard>
    </div>
  );
}
