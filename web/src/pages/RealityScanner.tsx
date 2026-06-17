import { useState } from "react";
import { useMutation, useQuery } from "@tanstack/react-query";
import { api } from "@/api/client";
import { Button, Card, Input, PageHeader, Select } from "@/components/ui";
import { useToast } from "@/components/toast";

interface ScanResult {
  sni: string;
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
  "www.google.com", "www.cloudflare.com", "www.microsoft.com",
  "www.apple.com", "www.amazon.com", "www.github.com",
  "www.mozilla.org", "www.wikipedia.org", "www.reddit.com",
  "www.twitch.tv", "www.spotify.com", "www.discord.com",
  "www.figma.com", "www.notion.so", "www.slack.com",
  "www.zoom.us", "www.netlify.com", "www.vercel.com",
  "www.digitalocean.com", "www.linode.com",
];

export function RealityScanner() {
  const toast = useToast();
  const [nodeId, setNodeId] = useState("");
  const [snis, setSnis] = useState(DEFAULT_SNIS.join("\n"));
  const [port, setPort] = useState("443");
  const [results, setResults] = useState<ScanResult[]>([]);

  const { data: nodesData } = useQuery({
    queryKey: ["nodes"],
    queryFn: () => api<{ nodes: Node[] }>("/api/nodes"),
  });

  const scanMut = useMutation({
    mutationFn: () => api<{ results: ScanResult[] }>("/api/reality/scan", {
      method: "POST",
      body: {
        node_id: nodeId,
        snis: snis.split("\n").map(s => s.trim()).filter(Boolean),
        port: Number(port),
      },
    }),
    onSuccess: (data) => { setResults(data.results); toast.success(`Scanned ${data.results.length} SNIs`); },
    onError: (e: any) => toast.error(e.message),
  });

  return (
    <div className="space-y-6 animate-fade-in">
      <PageHeader title="Reality Scanner" subtitle="Probe SNIs to find the best ones for REALITY" />

      <Card className="space-y-4">
        <div className="grid grid-cols-1 gap-3 md:grid-cols-3">
          <Select value={nodeId} onChange={(e) => setNodeId(e.target.value)}>
            <option value="">Select node...</option>
            {nodesData?.nodes?.map(n => <option key={n.id} value={n.id}>{n.name}</option>)}
          </Select>
          <Input placeholder="Port" value={port} onChange={(e) => setPort(e.target.value)} inputMode="numeric" />
          <Button onClick={() => scanMut.mutate()} disabled={!nodeId || scanMut.isPending}>
            {scanMut.isPending ? "Scanning..." : "Start Scan"}
          </Button>
        </div>
        <textarea
          placeholder="Enter SNIs (one per line)..."
          value={snis}
          onChange={(e) => setSnis(e.target.value)}
          className="field min-h-[150px] resize-y font-mono text-xs"
        />
      </Card>

      {results.length > 0 && (
        <Card>
          <div className="overflow-x-auto">
            <table className="w-full text-sm">
              <thead>
                <tr className="border-b border-border/40 text-xs text-fg-subtle">
                  <th className="py-2 text-left">SNI</th>
                  <th className="py-2 text-center">Score</th>
                  <th className="py-2 text-center">Latency</th>
                  <th className="py-2 text-center">Valid</th>
                  <th className="py-2 text-right">Action</th>
                </tr>
              </thead>
              <tbody>
                {results.map((r, i) => (
                  <tr key={i} className="border-b border-border/20 hover:bg-surface-2/40">
                    <td className="py-2 font-mono text-xs text-fg">{r.sni}</td>
                    <td className="py-2 text-center">
                      <span className={`inline-flex h-6 w-10 items-center justify-center rounded-full text-xs font-bold ${
                        r.score >= 80 ? "bg-success/15 text-success" :
                        r.score >= 50 ? "bg-warning/15 text-warning" :
                        "bg-danger/15 text-danger"
                      }`}>
                        {r.score}
                      </span>
                    </td>
                    <td className="py-2 text-center text-xs text-fg-muted">{r.latency_ms}ms</td>
                    <td className="py-2 text-center">{r.valid ? "✓" : "✗"}</td>
                    <td className="py-2 text-right">
                      <Button size="sm" variant="ghost" onClick={() => { navigator.clipboard.writeText(r.sni); toast.success("Copied"); }}>
                        Copy
                      </Button>
                    </td>
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
