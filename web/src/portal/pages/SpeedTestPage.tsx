import { useState } from "react";
import { useMutation } from "@tanstack/react-query";
import { Zap, Gauge } from "lucide-react";
import { api } from "@/api/client";
import { Button } from "@/components/ui";
import { GlassCard } from "@/components/veltrix";

interface SpeedTestResult { node_id: string; node_name: string; latency_ms: number; download_mbps: number; }

export function SpeedTestPage() {
  const [results, setResults] = useState<SpeedTestResult[]>([]);

  const testMut = useMutation({
    mutationFn: (data: { node_id: string; node_endpoint: string }) =>
      api<SpeedTestResult>("/api/v2/portal/speed-test", { method: "POST", body: data }),
    onSuccess: (res) => setResults((p) => [res, ...p.slice(0, 9)]),
  });

  return (
    <div className="space-y-6">
      <h2 className="text-xl font-bold flex items-center gap-2"><Zap className="w-5 h-5"/>Speed Test</h2>
      <GlassCard className="p-4">
        <p className="text-sm text-fg-muted mb-4">Test connection speed to available nodes.</p>
        <Button onClick={() => testMut.mutate({ node_id: "", node_endpoint: "" })} disabled={testMut.isPending}>
          <Gauge className="w-4 h-4 mr-1"/>{testMut.isPending ? "Testing..." : "Run Speed Test"}
        </Button>
      </GlassCard>
      {results.length > 0 && (
        <GlassCard className="p-4">
          <h3 className="font-medium mb-3">Results</h3>
          <div className="space-y-2">{results.map((r, i) => (
            <div key={i} className="flex items-center justify-between border border-border rounded-xl p-3">
              <div><span className="font-medium">{r.node_name || "Node"}</span><span className="ml-2 text-sm text-fg-muted">{r.latency_ms}ms</span></div>
              <span className="text-success font-mono">{r.download_mbps.toFixed(1)} Mbps</span>
            </div>
          ))}</div>
        </GlassCard>
      )}
    </div>
  );
}
