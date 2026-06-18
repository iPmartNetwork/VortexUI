import { useState } from "react";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { api } from "@/api/client";
import { Button, Card, Input, PageHeader, Badge } from "@/components/ui";
import { useToast } from "@/components/toast";
import { useI18n } from "@/i18n/i18n";

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
    onSuccess: () => { qc.invalidateQueries({ queryKey: ["doh-config"] }); toast.success("DoH config saved"); },
  });

  function update(field: keyof DoHConfig, value: any) {
    setForm(prev => ({ ...(prev ?? configData?.config ?? {} as any), [field]: value }));
  }

  return (
    <div className="space-y-6 animate-fade-in">
      <PageHeader title={t("doh.title")} subtitle={t("doh.subtitle")} />

      <div className="rounded-lg border border-border/40 bg-surface-2/20 p-4 text-xs text-fg-muted space-y-2">
        <p className="font-medium text-fg text-sm">Built-in DNS Privacy Server</p>
        <p>Provides a DNS-over-HTTPS endpoint for users, preventing DNS leaks and enabling ad/malware blocking at the DNS level.</p>
        <ul className="list-disc pl-4 space-y-1">
          <li><strong>Upstream DNS</strong> — Servers to forward queries to (e.g. 1.1.1.1, 8.8.8.8). One IP per line.</li>
          <li><strong>Block ads/malware</strong> — Uses curated blocklists to filter known ad/malware domains.</li>
          <li><strong>Custom blocklist</strong> — Your own domains to block. One domain per line.</li>
          <li><strong>Cache TTL</strong> — How long resolved queries stay cached (seconds).</li>
        </ul>
      </div>

      {stats && (
        <div className="grid grid-cols-2 gap-4 md:grid-cols-4">
          <Card className="space-y-1">
            <div className="text-xs text-fg-subtle">{t("doh.totalQueries")}</div>
            <div className="text-lg font-bold text-fg">{stats.total_queries.toLocaleString()}</div>
          </Card>
          <Card className="space-y-1">
            <div className="text-xs text-fg-subtle">{t("doh.blocked")}</div>
            <div className="text-lg font-bold text-danger">{stats.blocked_count.toLocaleString()}</div>
          </Card>
          <Card className="space-y-1">
            <div className="text-xs text-fg-subtle">{t("doh.cacheHits")}</div>
            <div className="text-lg font-bold text-success">{stats.cache_hits.toLocaleString()}</div>
          </Card>
          <Card className="space-y-1">
            <div className="text-xs text-fg-subtle">{t("doh.avgLatency")}</div>
            <div className="text-lg font-bold text-fg">{stats.avg_latency_ms}ms</div>
          </Card>
        </div>
      )}

      {config && (
        <Card className="space-y-4">
          <div className="flex items-center justify-between">
            <h3 className="text-sm font-bold text-fg">{t("doh.config")}</h3>
            <label className="flex items-center gap-2 text-sm">
              <input type="checkbox" checked={config.enabled} onChange={(e) => update("enabled", e.target.checked)} className="rounded" />
              Enabled
            </label>
          </div>
          <div className="grid grid-cols-2 gap-3 md:grid-cols-3">
            <div>
              <label className="text-xs text-fg-subtle">{t("doh.listenAddr")}</label>
              <Input value={config.listen_addr} onChange={(e) => update("listen_addr", e.target.value)} placeholder=":8053" />
            </div>
            <div>
              <label className="text-xs text-fg-subtle">Cache TTL (s)</label>
              <Input value={config.cache_ttl} onChange={(e) => update("cache_ttl", Number(e.target.value))} inputMode="numeric" />
            </div>
            <div className="space-y-2">
              <label className="flex items-center gap-2 text-xs"><input type="checkbox" checked={config.block_ads} onChange={(e) => update("block_ads", e.target.checked)} className="rounded" /> Block ads</label>
              <label className="flex items-center gap-2 text-xs"><input type="checkbox" checked={config.block_malware} onChange={(e) => update("block_malware", e.target.checked)} className="rounded" /> Block malware</label>
              <label className="flex items-center gap-2 text-xs"><input type="checkbox" checked={config.log_queries} onChange={(e) => update("log_queries", e.target.checked)} className="rounded" /> Log queries</label>
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
          <div className="flex justify-end">
            <Button onClick={() => config && save.mutate(config)} disabled={save.isPending}>Save</Button>
          </div>
        </Card>
      )}

      {logsData?.logs && logsData.logs.length > 0 && (
        <Card>
          <h3 className="text-sm font-bold text-fg mb-3">{t("doh.recentQueries")}</h3>
          <div className="overflow-x-auto max-h-[300px]">
            <table className="w-full text-xs">
              <thead>
                <tr className="border-b border-border/40 text-fg-subtle">
                  <th className="py-1 text-left">Domain</th>
                  <th className="py-1">Type</th>
                  <th className="py-1">Client</th>
                  <th className="py-1">Status</th>
                  <th className="py-1">Latency</th>
                </tr>
              </thead>
              <tbody>
                {logsData.logs.slice(0, 30).map((l, i) => (
                  <tr key={i} className="border-b border-border/20">
                    <td className="py-1 font-mono text-fg">{l.domain}</td>
                    <td className="py-1 text-center text-fg-muted">{l.type}</td>
                    <td className="py-1 text-center font-mono text-fg-muted">{l.client_ip}</td>
                    <td className="py-1 text-center"><Badge color={l.blocked ? "expired" : "active"}>{l.blocked ? "Blocked" : "OK"}</Badge></td>
                    <td className="py-1 text-center text-fg-muted">{l.latency_ms}ms</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </Card>
      )}
    </div>
  );
}
