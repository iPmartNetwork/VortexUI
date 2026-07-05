import { useState } from "react";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { Info, ShieldCheck } from "lucide-react";
import { api } from "@/api/client";
import { Button, Input, Badge } from "@/components/ui";
import { GlassCard, StatsCard } from "@/components/veltrix";
import { useToast } from "@/components/toast";
import { useI18n } from "@/i18n/i18n";
import { useTitle } from "@/lib/useTitle";

interface DoHConfig {
  enabled: boolean;
  listen_addr: string;
  upstream_dns: string[];
  block_ads: boolean;
  block_malware: boolean;
  custom_blocklist: string[];
  log_queries: boolean;
  cache_ttl: number;
}

interface DoHStats {
  total_queries: number;
  blocked_count: number;
  cache_hits: number;
  avg_latency_ms: number;
}

interface QueryLog {
  domain: string;
  type: string;
  client_ip: string;
  blocked: boolean;
  latency_ms: number;
  timestamp: string;
}

export function DoHSettings() {
  useTitle("DoH Settings");
  const { t } = useI18n();
  const qc = useQueryClient();
  const toast = useToast();
  const [form, setForm] = useState<DoHConfig | null>(null);

  const { data: configData } = useQuery({
    queryKey: ["doh-config"],
    queryFn: () => api<{ config: DoHConfig }>("/api/doh/config"),
  });
  const { data: statsData } = useQuery({
    queryKey: ["doh-stats"],
    queryFn: () => api<{ stats: DoHStats }>("/api/doh/stats"),
  });
  const { data: logsData } = useQuery({
    queryKey: ["doh-logs"],
    queryFn: () => api<{ logs: QueryLog[] }>("/api/doh/logs"),
  });

  const config = form ?? configData?.config;
  const stats = statsData?.stats;

  const save = useMutation({
    mutationFn: (c: DoHConfig) => api("/api/doh/config", { method: "PUT", body: c }),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["doh-config"] });
      toast.success("DoH config saved");
    },
  });

  function update<K extends keyof DoHConfig>(field: K, value: DoHConfig[K]) {
    setForm((prev) => ({ ...(prev ?? configData?.config ?? ({} as DoHConfig)), [field]: value }));
  }

  return (
    <div className="space-y-5 animate-page-enter">
      <div>
        <h1 className="text-2xl font-bold text-fg tracking-tight">{t("doh.title")}</h1>
        <p className="text-sm text-fg-muted mt-1">{t("doh.subtitle")}</p>
      </div>

      <div className="rounded-xl border border-primary/25 bg-primary/5 px-4 py-3 flex items-start gap-3">
        <div className="h-8 w-8 rounded-full bg-primary/15 flex items-center justify-center text-primary flex-shrink-0">
          <Info size={16} />
        </div>
        <div className="text-xs text-fg-muted leading-relaxed space-y-1.5">
          <p className="font-semibold text-fg text-sm">{t("doh.infoTitle")}</p>
          <p>{t("doh.infoDesc")}</p>
          <ul className="space-y-1 pt-1">
            <li><strong className="text-fg">Upstream DNS</strong> — {t("doh.upstream")}</li>
            <li><strong className="text-fg">Block ads/malware</strong> — {t("doh.blockAds")}</li>
            <li><strong className="text-fg">Custom blocklist</strong> — {t("doh.blocklist")}</li>
            <li><strong className="text-fg">Cache TTL</strong> — {t("doh.cacheTTL")}</li>
          </ul>
        </div>
      </div>

      {stats && (
        <div className="grid grid-cols-2 gap-4 md:grid-cols-4">
          <StatsCard title={t("doh.totalQueries")} value={stats.total_queries.toLocaleString()} icon={<ShieldCheck size={18} />} color="blue" />
          <StatsCard title={t("doh.blocked")} value={stats.blocked_count.toLocaleString()} icon={<ShieldCheck size={18} />} color="red" />
          <StatsCard title={t("doh.cacheHits")} value={stats.cache_hits.toLocaleString()} icon={<ShieldCheck size={18} />} color="green" />
          <StatsCard title={t("doh.avgLatency")} value={stats.avg_latency_ms} suffix="ms" icon={<ShieldCheck size={18} />} color="cyan" />
        </div>
      )}

      {config && (
        <GlassCard hover={false} className="!p-5 space-y-4">
          <div className="flex items-center justify-between">
            <h3 className="text-sm font-bold text-fg">{t("doh.config")}</h3>
            <label className="flex items-center gap-2 text-sm text-fg">
              <input type="checkbox" checked={config.enabled} onChange={(e) => update("enabled", e.target.checked)} className="rounded" />
              Enabled
            </label>
          </div>
          <div className="grid grid-cols-1 gap-3 sm:grid-cols-2 md:grid-cols-3">
            <div>
              <label className="text-xs text-fg-subtle">{t("doh.listenAddr")}</label>
              <Input value={config.listen_addr} onChange={(e) => update("listen_addr", e.target.value)} placeholder=":8053" />
            </div>
            <div>
              <label className="text-xs text-fg-subtle">Cache TTL (s)</label>
              <Input value={config.cache_ttl} onChange={(e) => update("cache_ttl", Number(e.target.value))} inputMode="numeric" />
            </div>
            <div className="space-y-2">
              <label className="flex items-center gap-2 text-xs text-fg"><input type="checkbox" checked={config.block_ads} onChange={(e) => update("block_ads", e.target.checked)} className="rounded" /> Block ads</label>
              <label className="flex items-center gap-2 text-xs text-fg"><input type="checkbox" checked={config.block_malware} onChange={(e) => update("block_malware", e.target.checked)} className="rounded" /> Block malware</label>
              <label className="flex items-center gap-2 text-xs text-fg"><input type="checkbox" checked={config.log_queries} onChange={(e) => update("log_queries", e.target.checked)} className="rounded" /> Log queries</label>
            </div>
          </div>
          <div>
            <label className="text-xs text-fg-subtle">Upstream DNS (one per line)</label>
            <textarea className="field min-h-[60px] resize-y font-mono text-xs" value={config.upstream_dns?.join("\n") ?? ""} onChange={(e) => update("upstream_dns", e.target.value.split("\n").filter(Boolean))} />
          </div>
          <div>
            <label className="text-xs text-fg-subtle">Custom blocklist (domains, one per line)</label>
            <textarea className="field min-h-[60px] resize-y font-mono text-xs" value={config.custom_blocklist?.join("\n") ?? ""} onChange={(e) => update("custom_blocklist", e.target.value.split("\n").filter(Boolean))} />
          </div>
          <div className="flex justify-end pt-1 border-t border-border/40">
            <Button onClick={() => config && save.mutate(config)} disabled={save.isPending}>Save</Button>
          </div>
        </GlassCard>
      )}

      {logsData?.logs && logsData.logs.length > 0 && (
        <GlassCard hover={false} className="!p-0 overflow-hidden">
          <div className="px-4 pt-4 pb-1">
            <h3 className="text-sm font-bold text-fg">{t("doh.recentQueries")}</h3>
          </div>
          <div className="overflow-x-auto max-h-[300px] mt-2">
            <table className="w-full text-sm">
              <thead>
                <tr className="border-b border-border/40 text-[11px] uppercase tracking-wide text-fg-subtle bg-surface-2/30">
                  <th className="py-3 px-4 text-left">Domain</th>
                  <th className="py-3 px-4 text-center">Type</th>
                  <th className="py-3 px-4 text-center">Client</th>
                  <th className="py-3 px-4 text-center">Status</th>
                  <th className="py-3 px-4 text-right">Latency</th>
                </tr>
              </thead>
              <tbody>
                {logsData.logs.slice(0, 30).map((l, i) => (
                  <tr key={i} className="border-b border-border/20 hover:bg-surface-2/40">
                    <td className="py-2.5 px-4 font-mono text-xs text-fg">{l.domain}</td>
                    <td className="py-2.5 px-4 text-center text-fg-muted">{l.type}</td>
                    <td className="py-2.5 px-4 text-center font-mono text-xs text-fg-muted" dir="ltr">{l.client_ip}</td>
                    <td className="py-2.5 px-4 text-center"><Badge color={l.blocked ? "expired" : "active"}>{l.blocked ? "Blocked" : "OK"}</Badge></td>
                    <td className="py-2.5 px-4 text-right text-fg-muted tabular-nums">{l.latency_ms}ms</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </GlassCard>
      )}
    </div>
  );
}
