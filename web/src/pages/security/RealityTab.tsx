import { useState } from "react";
import { useMutation, useQuery } from "@tanstack/react-query";
import { Play, Sparkles } from "lucide-react";
import { api } from "@/api/client";
import { Badge, Button, Input, Select } from "@/components/ui";
import { GlassCard } from "@/components/veltrix";
import { useToast } from "@/components/toast";
import { useI18n } from "@/i18n/i18n";
import { cn } from "@/lib/utils";

interface ScanResult {
  sni: string;
  resolved_ip?: string;
  tls_version?: string;
  alpn?: string;
  latency_ms: number;
  score: number;
  valid: boolean;
  error?: string;
}

interface Node {
  id: string;
  name: string;
}

const DEFAULT_SNIS = [
  "www.microsoft.com",
  "dl.google.com",
  "www.apple.com",
  "aws.amazon.com",
  "www.cloudflare.com",
  "support.lenovo.com",
  "www.github.com",
  "www.mozilla.org",
  "www.spotify.com",
  "www.discord.com",
];

function scoreBarClass(score: number): string {
  if (score >= 80) return "bg-success";
  if (score >= 50) return "bg-warning";
  return "bg-danger";
}

function tlsBadge(version?: string, alpn?: string) {
  if (!version) return "—";
  const alpnLabel = alpn?.toLowerCase().includes("h2") ? "H2" : alpn || "";
  return alpnLabel ? `${version} · ${alpnLabel}` : version;
}

export function RealityTab() {
  const { t } = useI18n();
  const toast = useToast();
  const [nodeId, setNodeId] = useState("");
  const [snis, setSnis] = useState(DEFAULT_SNIS.join("\n"));
  const [port, setPort] = useState("443");
  const [results, setResults] = useState<ScanResult[]>([]);
  const [showAdvanced, setShowAdvanced] = useState(false);

  const { data: nodesData } = useQuery({
    queryKey: ["nodes"],
    queryFn: () => api<{ nodes: Node[] }>("/api/nodes"),
  });

  const scanMut = useMutation({
    mutationFn: () =>
      api<{ results: ScanResult[] }>("/api/reality/scan", {
        method: "POST",
        body: {
          node_id: nodeId,
          snis: snis.split("\n").map((s) => s.trim()).filter(Boolean),
          port: Number(port),
        },
      }),
    onSuccess: (data) => {
      setResults(data.results);
      toast.success(`${t("security.reality.scanDone")}: ${data.results.length}`);
    },
    onError: (e: Error) => toast.error(e.message),
  });

  function applySni(sni: string) {
    navigator.clipboard.writeText(sni);
    toast.success(t("security.reality.sniCopied"));
  }

  return (
    <div className="space-y-4">
      <div className="rounded-xl border border-primary/25 bg-primary/5 px-4 py-3 flex flex-col sm:flex-row sm:items-center gap-3">
        <div className="flex items-start gap-3 flex-1 min-w-0">
          <div className="h-8 w-8 rounded-full bg-primary/15 flex items-center justify-center text-primary flex-shrink-0">
            <Sparkles size={16} />
          </div>
          <p className="text-xs text-fg-muted leading-relaxed">{t("security.reality.banner")}</p>
        </div>
        <Button
          size="sm"
          className="flex-shrink-0"
          onClick={() => scanMut.mutate()}
          disabled={!nodeId || scanMut.isPending}
        >
          <Play size={14} />
          {scanMut.isPending ? t("security.reality.scanning") : t("security.reality.runScan")}
        </Button>
      </div>

      <GlassCard hover={false} className="!p-4 space-y-3">
        <div className="grid grid-cols-1 gap-3 md:grid-cols-3">
          <Select value={nodeId} onChange={(e) => setNodeId(e.target.value)}>
            <option value="">{t("security.reality.selectNode")}</option>
            {nodesData?.nodes?.map((n) => (
              <option key={n.id} value={n.id}>{n.name}</option>
            ))}
          </Select>
          <Input placeholder={t("cleanip.port")} value={port} onChange={(e) => setPort(e.target.value)} inputMode="numeric" />
          <Button variant="outline" size="sm" onClick={() => setShowAdvanced((v) => !v)}>
            {showAdvanced ? t("security.reality.hideSnis") : t("security.reality.editSnis")}
          </Button>
        </div>
        {showAdvanced && (
          <textarea
            placeholder="SNI (one per line)…"
            value={snis}
            onChange={(e) => setSnis(e.target.value)}
            className="field min-h-[120px] resize-y font-mono text-xs"
            dir="ltr"
          />
        )}
      </GlassCard>

      {results.length > 0 && (
        <GlassCard hover={false} className="!p-0 overflow-hidden">
          <div className="overflow-x-auto">
            <table className="w-full text-sm">
              <thead>
                <tr className="border-b border-border/40 text-[11px] uppercase tracking-wide text-fg-subtle bg-surface-2/30">
                  <th className="py-3 px-4 text-left">{t("security.reality.colDomain")}</th>
                  <th className="py-3 px-4 text-left">{t("security.reality.colIP")}</th>
                  <th className="py-3 px-4 text-center">{t("security.reality.colTLS")}</th>
                  <th className="py-3 px-4 text-center">{t("cleanip.latency")}</th>
                  <th className="py-3 px-4 text-left">{t("security.reality.colScore")}</th>
                  <th className="py-3 px-4 text-right">{t("common.actions")}</th>
                </tr>
              </thead>
              <tbody>
                {results.map((r) => (
                  <tr key={r.sni} className="border-b border-border/20 hover:bg-surface-2/40">
                    <td className="py-3 px-4 font-mono text-xs text-fg">{r.sni}</td>
                    <td className="py-3 px-4 font-mono text-xs text-fg-muted" dir="ltr">{r.resolved_ip || "—"}</td>
                    <td className="py-3 px-4 text-center">
                      {r.valid ? (
                        <Badge color="active">{tlsBadge(r.tls_version, r.alpn)}</Badge>
                      ) : (
                        <Badge color="down">—</Badge>
                      )}
                    </td>
                    <td className="py-3 px-4 text-center text-xs tabular-nums text-fg-muted">{r.latency_ms} ms</td>
                    <td className="py-3 px-4">
                      <div className="flex items-center gap-2 min-w-[120px]">
                        <div className="flex-1 h-1.5 rounded-full bg-surface-3 overflow-hidden">
                          <div className={cn("h-full rounded-full", scoreBarClass(r.score))} style={{ width: `${r.score}%` }} />
                        </div>
                        <span className="text-xs font-bold tabular-nums w-8 text-end">{r.score}</span>
                      </div>
                    </td>
                    <td className="py-3 px-4 text-right">
                      <Button size="sm" variant="outline" onClick={() => applySni(r.sni)}>
                        {t("security.reality.applyInbound")}
                      </Button>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </GlassCard>
      )}

      {results.length === 0 && !scanMut.isPending && (
        <p className="text-sm text-fg-muted text-center py-8">{t("security.reality.empty")}</p>
      )}
    </div>
  );
}
