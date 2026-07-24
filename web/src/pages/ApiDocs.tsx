import { useState } from "react";
import { useQuery } from "@tanstack/react-query";
import { BookOpen, BarChart3, ExternalLink } from "lucide-react";
import { api } from "@/api/client";
import { Button, Badge } from "@/components/ui";
import { GlassCard } from "@/components/veltrix";
import { useTitle } from "@/lib/useTitle";
import { useI18n } from "@/i18n/i18n";

interface RateLimitInfo { endpoint: string; limit: number; remaining: number; window_sec: number; }

export function ApiDocs() {
  const { t: _t } = useI18n();
  useTitle("API Docs");
  const [tab, setTab] = useState<"docs"|"rate-limits">("docs");

  const { data: rateLimits } = useQuery({
    queryKey: ["rate-limits"],
    queryFn: () => api<RateLimitInfo[]>("/api/v2/rate-limits"),
    enabled: tab === "rate-limits",
    refetchInterval: 5000,
  });

  return (
    <div className="space-y-6 p-6">
      <h1 className="text-2xl font-bold flex items-center gap-2"><BookOpen className="w-6 h-6" />API Documentation</h1>
      <div className="flex gap-2 border-b border-border pb-2">
        <Button variant={tab==="docs"?"primary":"ghost"} size="sm" onClick={()=>setTab("docs")}><BookOpen className="w-4 h-4 mr-1"/>Docs</Button>
        <Button variant={tab==="rate-limits"?"primary":"ghost"} size="sm" onClick={()=>setTab("rate-limits")}><BarChart3 className="w-4 h-4 mr-1"/>Rate Limits</Button>
      </div>

      {tab === "docs" && (
        <GlassCard className="p-6 space-y-4">
          <h2 className="text-lg font-semibold">Swagger UI</h2>
          <p className="text-fg-muted">Interactive API docs with Try-it-out.</p>
          <div className="flex gap-3">
            <a href="/api/v2/docs/" target="_blank" rel="noopener noreferrer"><Button><ExternalLink className="w-4 h-4 mr-1"/>Open Swagger</Button></a>
            <a href="/api/v2/docs/openapi.yaml" target="_blank" rel="noopener noreferrer"><Button variant="outline">Download Spec</Button></a>
          </div>
          <hr className="border-border" />
          <h2 className="text-lg font-semibold">API Versioning</h2>
          <p className="text-fg-muted">Current: v2. Legacy v1 endpoints remain functional.</p>
          <div className="flex gap-2"><Badge color="active">v2</Badge><Badge>v1 (legacy)</Badge></div>
        </GlassCard>
      )}

      {tab === "rate-limits" && (
        <GlassCard className="p-4 space-y-4">
          <h2 className="text-lg font-semibold flex items-center gap-2"><BarChart3 className="w-5 h-5"/>Rate Limit Dashboard</h2>
          {rateLimits && rateLimits.length > 0 ? (
            <table className="w-full text-sm">
              <thead><tr className="border-b"><th className="text-left py-2 px-3">Endpoint</th><th className="text-left py-2 px-3">Limit</th><th className="text-left py-2 px-3">Remaining</th><th className="text-left py-2 px-3">Window</th><th className="text-left py-2 px-3">Usage</th></tr></thead>
              <tbody>{rateLimits.map((rl) => {
                const usage = ((rl.limit - rl.remaining) / rl.limit) * 100;
                return (
                  <tr key={rl.endpoint} className="border-b hover:bg-surface-2/50">
                    <td className="py-2 px-3 font-mono text-xs">{rl.endpoint}</td>
                    <td className="py-2 px-3">{rl.limit}</td>
                    <td className="py-2 px-3"><Badge color={rl.remaining < rl.limit*0.2 ? "expired" : "active"}>{rl.remaining}</Badge></td>
                    <td className="py-2 px-3">{rl.window_sec}s</td>
                    <td className="py-2 px-3"><div className="w-24 h-2 bg-surface-2 rounded-full overflow-hidden"><div className={`h-full rounded-full ${usage>80?"bg-danger":usage>50?"bg-warning":"bg-success"}`} style={{width:`${Math.min(usage,100)}%`}}/></div></td>
                  </tr>
                );
              })}</tbody>
            </table>
          ) : <p className="text-fg-muted text-sm">No rate limit data.</p>}
        </GlassCard>
      )}
    </div>
  );
}
