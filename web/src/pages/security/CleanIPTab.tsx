import { useEffect, useState } from "react";
import { Crosshair } from "lucide-react";
import { useCleanIPResults, useScanCleanIP, type CleanIPScan } from "@/api/hooks";
import { Badge, Button, Input } from "@/components/ui";
import { CopyField } from "@/components/CopyField";
import { GlassCard } from "@/components/veltrix";
import { useToast } from "@/components/toast";
import { useI18n } from "@/i18n/i18n";
import { cn } from "@/lib/utils";

const CLOUDFLARE_PRESET = [
  "104.16.0.0", "104.17.0.0", "104.18.0.0", "104.19.0.0", "104.20.0.0",
  "104.21.0.0", "104.22.0.0", "104.24.0.0", "104.25.0.0", "104.26.0.0",
  "104.27.0.0", "172.64.0.0", "172.66.0.0", "172.67.0.0", "188.114.96.0", "162.159.0.0",
];

function parseIPs(raw: string): string[] {
  return raw.split(/[\n,]+/).map((s) => s.trim()).filter(Boolean);
}

function scoreBarClass(score: number): string {
  if (score >= 80) return "bg-success";
  if (score >= 50) return "bg-warning";
  return "bg-danger";
}

export function CleanIPTab() {
  const toast = useToast();
  const { t } = useI18n();
  const [ipsText, setIpsText] = useState("");
  const [port, setPort] = useState("443");
  const [results, setResults] = useState<CleanIPScan[]>([]);
  const [showAdvanced, setShowAdvanced] = useState(false);

  const { data: cached } = useCleanIPResults();
  const scanMut = useScanCleanIP();

  useEffect(() => {
    if (cached?.results && results.length === 0) setResults(cached.results);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [cached]);

  function runScan() {
    const ips = parseIPs(ipsText);
    if (ips.length === 0) {
      toast.error(t("cleanip.noCandidates"));
      return;
    }
    scanMut.mutate(
      { ips, port: Number(port) || 443 },
      {
        onSuccess: (data) => {
          setResults(data.results);
          toast.success(`${t("cleanip.title")}: ${data.results.length}`);
        },
        onError: (e: unknown) => toast.error(e instanceof Error ? e.message : "scan failed"),
      },
    );
  }

  function loadPreset() {
    setIpsText(CLOUDFLARE_PRESET.join("\n"));
    toast.info(t("cleanip.presetLoaded"));
  }

  return (
    <div className="space-y-4">
      <div className="rounded-xl border border-primary/25 bg-primary/5 px-4 py-3 flex flex-col sm:flex-row sm:items-center gap-3">
        <div className="flex items-start gap-3 flex-1 min-w-0">
          <div className="h-8 w-8 rounded-full bg-primary/15 flex items-center justify-center text-primary flex-shrink-0">
            <Crosshair size={16} />
          </div>
          <p className="text-xs text-fg-muted leading-relaxed">{t("security.cleanip.banner")}</p>
        </div>
        <Button size="sm" className="flex-shrink-0" onClick={runScan} disabled={scanMut.isPending}>
          {scanMut.isPending ? t("cleanip.scanning") : t("cleanip.scan")}
        </Button>
      </div>

      <GlassCard hover={false} className="!p-4 space-y-3">
        <div className="grid grid-cols-1 gap-3 md:grid-cols-3">
          <Input placeholder={t("cleanip.port")} value={port} onChange={(e) => setPort(e.target.value)} inputMode="numeric" />
          <Button variant="outline" size="sm" onClick={loadPreset}>{t("cleanip.preset")}</Button>
          <Button variant="outline" size="sm" onClick={() => setShowAdvanced((v) => !v)}>
            {showAdvanced ? t("security.cleanip.hideCandidates") : t("security.cleanip.editCandidates")}
          </Button>
        </div>
        {showAdvanced && (
          <>
            <textarea
              placeholder={t("cleanip.candidatesPlaceholder")}
              value={ipsText}
              onChange={(e) => setIpsText(e.target.value)}
              className="field min-h-[120px] resize-y font-mono text-xs"
              dir="ltr"
            />
            <p className="text-xs text-fg-subtle">{t("cleanip.hint")}</p>
          </>
        )}
      </GlassCard>

      {results.length > 0 ? (
        <GlassCard hover={false} className="!p-0 overflow-hidden">
          <div className="overflow-x-auto">
            <table className="w-full text-sm">
              <thead>
                <tr className="border-b border-border/40 text-[11px] uppercase tracking-wide text-fg-subtle bg-surface-2/30">
                  <th className="py-3 px-4 text-left">{t("cleanip.ip")}</th>
                  <th className="py-3 px-4 text-center">{t("cleanip.latency")}</th>
                  <th className="py-3 px-4 text-center">{t("cleanip.loss")}</th>
                  <th className="py-3 px-4 text-left">{t("cleanip.score")}</th>
                  <th className="py-3 px-4 text-center">{t("cleanip.reachable")}</th>
                  <th className="py-3 px-4 text-right">{t("cleanip.copy")}</th>
                </tr>
              </thead>
              <tbody>
                {results.map((r) => (
                  <tr key={r.id} className="border-b border-border/20 hover:bg-surface-2/40">
                    <td className="py-3 px-4 font-mono text-xs text-fg" dir="ltr">{r.ip}</td>
                    <td className="py-3 px-4 text-center text-xs tabular-nums text-fg-muted">{r.latency_ms} ms</td>
                    <td className="py-3 px-4 text-center text-xs tabular-nums text-fg-muted">{r.loss_pct}%</td>
                    <td className="py-3 px-4">
                      <div className="flex items-center gap-2 min-w-[100px]">
                        <div className="flex-1 h-1.5 rounded-full bg-surface-3 overflow-hidden">
                          <div className={cn("h-full rounded-full", scoreBarClass(r.score))} style={{ width: `${r.score}%` }} />
                        </div>
                        <span className="text-xs font-bold tabular-nums w-8 text-end">{r.score}</span>
                      </div>
                    </td>
                    <td className="py-3 px-4 text-center">
                      <Badge color={r.reachable ? "active" : "down"}>
                        {r.reachable ? t("cleanip.reachable") : t("cleanip.unreachable")}
                      </Badge>
                    </td>
                    <td className="py-3 px-4">
                      <div className="ms-auto w-44">
                        <CopyField value={r.ip} />
                      </div>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </GlassCard>
      ) : (
        <p className="text-sm text-fg-muted text-center py-8">{t("cleanip.empty")}</p>
      )}
    </div>
  );
}
