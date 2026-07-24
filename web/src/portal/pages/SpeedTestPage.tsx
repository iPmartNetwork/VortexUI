import { useState } from "react";
import { useMutation } from "@tanstack/react-query";
import { Zap, Gauge } from "lucide-react";
import { api } from "@/api/client";
import { Button } from "@/components/ui";
import { GlassCard } from "@/components/veltrix";
import { useI18n } from "@/i18n/i18n";

interface SpeedTestResult {
  node_id: string;
  node_name: string;
  latency_ms: number;
  download_mbps: number;
  timestamp: string;
}

export function SpeedTestPage() {
  const { t } = useI18n();
  const [results, setResults] = useState<SpeedTestResult[]>([]);

  const testMutation = useMutation({
    mutationFn: (data: { node_id: string; node_endpoint: string }) =>
      api.post("/api/v2/portal/speed-test", data),
    onSuccess: (res) => {
      setResults((prev) => [res.data as SpeedTestResult, ...prev.slice(0, 9)]);
    },
  });

  return (
    <div className="space-y-6">
      <h2 className="text-xl font-bold flex items-center gap-2">
        <Zap className="w-5 h-5" />
        {t("portal.speedTest", "Speed Test")}
      </h2>

      <GlassCard className="p-4">
        <p className="text-sm text-muted-foreground mb-4">
          {t("portal.speedTestDesc", "Test connection speed to available nodes.")}
        </p>
        <Button
          onClick={() =>
            testMutation.mutate({ node_id: "", node_endpoint: "" })
          }
          disabled={testMutation.isPending}
        >
          <Gauge className="w-4 h-4 mr-1" />
          {testMutation.isPending
            ? t("portal.testing", "Testing...")
            : t("portal.runTest", "Run Speed Test")}
        </Button>
      </GlassCard>

      {results.length > 0 && (
        <GlassCard className="p-4">
          <h3 className="font-medium mb-3">{t("portal.results", "Results")}</h3>
          <div className="space-y-2">
            {results.map((r, i) => (
              <div key={i} className="flex items-center justify-between border rounded p-3">
                <div>
                  <span className="font-medium">{r.node_name || "Node"}</span>
                  <span className="ml-2 text-sm text-muted-foreground">
                    {r.latency_ms}ms
                  </span>
                </div>
                <span className="text-green-600 font-mono">
                  {r.download_mbps.toFixed(1)} Mbps
                </span>
              </div>
            ))}
          </div>
        </GlassCard>
      )}
    </div>
  );
}
