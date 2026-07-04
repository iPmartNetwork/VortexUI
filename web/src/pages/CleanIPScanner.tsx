import { useEffect, useState } from "react";
import { useCleanIPResults, useScanCleanIP, type CleanIPScan } from "@/api/hooks";
import { Badge, Button, Card, Input, PageHeader } from "@/components/ui";
import { CopyField } from "@/components/CopyField";
import { useToast } from "@/components/toast";
import { useI18n } from "@/i18n/i18n";

// A small set of well-known Cloudflare anycast IPs admins can probe out of the
// box. These are public CDN front IPs, not internal ranges.
const CLOUDFLARE_PRESET = [
  "104.16.0.0",
  "104.17.0.0",
  "104.18.0.0",
  "104.19.0.0",
  "104.20.0.0",
  "104.21.0.0",
  "104.22.0.0",
  "104.24.0.0",
  "104.25.0.0",
  "104.26.0.0",
  "104.27.0.0",
  "172.64.0.0",
  "172.66.0.0",
  "172.67.0.0",
  "188.114.96.0",
  "162.159.0.0",
];

// parseIPs splits a free-form textarea into trimmed, non-empty candidate IPs,
// accepting both newline- and comma-separated input.
function parseIPs(raw: string): string[] {
  return raw
    .split(/[\n,]+/)
    .map((s) => s.trim())
    .filter(Boolean);
}

export function CleanIPScanner() {
  const toast = useToast();
  const { t } = useI18n();
  const [ipsText, setIpsText] = useState("");
  const [port, setPort] = useState("443");
  const [results, setResults] = useState<CleanIPScan[]>([]);

  const { data: cached } = useCleanIPResults();
  const scanMut = useScanCleanIP();

  // Hydrate from the last cached scan on mount (and whenever it refreshes),
  // unless the admin has already produced fresher results this session.
  useEffect(() => {
    if (cached?.results && results.length === 0) {
      setResults(cached.results);
    }
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
    <div className="space-y-6 animate-page-enter">
      <PageHeader title={t("cleanip.title")} subtitle={t("cleanip.subtitle")} />

      <Card className="space-y-4">
        <div className="grid grid-cols-1 gap-3 md:grid-cols-3">
          <Input
            placeholder={t("cleanip.port")}
            value={port}
            onChange={(e) => setPort(e.target.value)}
            inputMode="numeric"
          />
          <Button variant="outline" onClick={loadPreset}>
            {t("cleanip.preset")}
          </Button>
          <Button onClick={runScan} disabled={scanMut.isPending}>
            {scanMut.isPending ? t("cleanip.scanning") : t("cleanip.scan")}
          </Button>
        </div>
        <textarea
          placeholder={t("cleanip.candidatesPlaceholder")}
          value={ipsText}
          onChange={(e) => setIpsText(e.target.value)}
          className="field min-h-[150px] resize-y font-mono text-xs"
          dir="ltr"
        />
        <p className="text-xs text-fg-subtle">{t("cleanip.hint")}</p>
      </Card>

      {results.length > 0 ? (
        <Card>
          <div className="overflow-x-auto">
            <table className="w-full text-sm">
              <thead>
                <tr className="border-b border-border/40 text-xs text-fg-subtle">
                  <th className="py-2 text-left">{t("cleanip.ip")}</th>
                  <th className="py-2 text-center">{t("cleanip.score")}</th>
                  <th className="py-2 text-center">{t("cleanip.latency")}</th>
                  <th className="py-2 text-center">{t("cleanip.loss")}</th>
                  <th className="py-2 text-center">{t("cleanip.reachable")}</th>
                  <th className="py-2 text-right">{t("cleanip.copy")}</th>
                </tr>
              </thead>
              <tbody>
                {results.map((r) => (
                  <tr key={r.id} className="border-b border-border/20 hover:bg-surface-2/40">
                    <td className="py-2 font-mono text-xs text-fg" dir="ltr">{r.ip}</td>
                    <td className="py-2 text-center">
                      <span
                        className={`inline-flex h-6 w-10 items-center justify-center rounded-full text-xs font-bold ${
                          r.score >= 80
                            ? "bg-success/15 text-success"
                            : r.score >= 50
                              ? "bg-warning/15 text-warning"
                              : "bg-danger/15 text-danger"
                        }`}
                      >
                        {r.score}
                      </span>
                    </td>
                    <td className="py-2 text-center text-xs text-fg-muted">{r.latency_ms}ms</td>
                    <td className="py-2 text-center text-xs text-fg-muted">{r.loss_pct}%</td>
                    <td className="py-2 text-center">
                      <Badge color={r.reachable ? "active" : "down"}>
                        {r.reachable ? t("cleanip.reachable") : t("cleanip.unreachable")}
                      </Badge>
                    </td>
                    <td className="py-2">
                      <div className="ms-auto w-44">
                        <CopyField value={r.ip} />
                      </div>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </Card>
      ) : (
        <Card>
          <p className="py-8 text-center text-sm text-fg-muted">{t("cleanip.empty")}</p>
        </Card>
      )}
    </div>
  );
}
