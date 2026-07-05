import { useEffect, useState } from "react";
import { Copy, Crosshair, Gauge } from "lucide-react";
import { useCleanIPResults, useScanCleanIP, useMeasureThroughput, type CleanIPScan } from "@/api/hooks";
import { Badge, Button, Input } from "@/components/ui";
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
  const throughputMut = useMeasureThroughput();

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

  function copyIp(ip: string) {
    navigator.clipboard?.writeText(ip);
    toast.success(t("cleanip.copied"));
  }

  function measureSpeed(r: CleanIPScan) {
    throughputMut.mutate(
      { id: r.id, ip: r.ip, port: Number(port) || 443 },
      {
        onSuccess: (data) => {
          setResults((prev) => prev.map((x) => (x.id === r.id ? { ...x, throughput_mbps: data.throughput_mbps } : x)));
        },
        onError: (e: unknown) => toast.error(e instanceof Error ? e.message : t("cleanip.speedError")),
      },
    );
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
        <div className="grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-4 xl:grid-cols-5 gap-4">
          {results.map((r) => (
            <GlassCard key={r.id} hover className="!p-4 space-y-3">
              <div className="flex items-center justify-between gap-2">
                <span
                  className={cn(
                    "h-2 w-2 rounded-full flex-shrink-0",
                    r.reachable ? (r.score >= 80 ? "bg-success" : "bg-warning") : "bg-danger",
                  )}
                />
                <Badge color={r.reachable ? "active" : "down"}>
                  {r.reachable ? t("cleanip.reachable") : t("cleanip.unreachable")}
                </Badge>
              </div>

              <p className="font-mono text-sm font-bold text-fg text-center truncate" dir="ltr" title={r.ip}>
                {r.ip}
              </p>

              <div className="flex items-center gap-2">
                <div className="flex-1 h-1.5 rounded-full bg-surface-3 overflow-hidden">
                  <div className={cn("h-full rounded-full", scoreBarClass(r.score))} style={{ width: `${r.score}%` }} />
                </div>
                <span className="text-xs font-bold tabular-nums">{r.score}</span>
              </div>

              <div className="flex items-center justify-between text-[11px] text-fg-muted">
                <span>
                  {t("cleanip.latency")}: <span className="font-medium text-fg">{r.latency_ms} ms</span>
                </span>
                <span>
                  {t("cleanip.loss")}: <span className="font-medium text-fg">{r.loss_pct}%</span>
                </span>
              </div>

              <div className="flex items-center justify-between text-[11px] text-fg-muted">
                <span>{t("cleanip.speed")}</span>
                <span className="font-medium text-fg">
                  {r.throughput_mbps > 0 ? `${r.throughput_mbps.toFixed(1)} Mbps` : t("cleanip.notMeasured")}
                </span>
              </div>

              <div className="flex items-center gap-2">
                <Button
                  variant="outline"
                  size="sm"
                  className="flex-1"
                  disabled={!r.reachable || (throughputMut.isPending && throughputMut.variables?.id === r.id)}
                  onClick={() => measureSpeed(r)}
                >
                  <Gauge size={13} />
                  {throughputMut.isPending && throughputMut.variables?.id === r.id ? t("cleanip.measuring") : t("cleanip.measureSpeed")}
                </Button>
                <Button variant="outline" size="sm" className="flex-1" onClick={() => copyIp(r.ip)}>
                  <Copy size={13} /> {t("cleanip.copy")}
                </Button>
              </div>
            </GlassCard>
          ))}
        </div>
      ) : (
        <p className="text-sm text-fg-muted text-center py-8">{t("cleanip.empty")}</p>
      )}
    </div>
  );
}
