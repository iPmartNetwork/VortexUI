import { useState } from "react";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { Copy, CheckCircle } from "lucide-react";
import { api } from "@/api/client";
import { Button } from "./ui";
import { useToast } from "./toast";
import { cn } from "@/lib/utils";

interface InboundTrafficStats {
  inbound_id: string;
  upload: number;
  download: number;
  total: number;
  daily: { date: string; upload: number; download: number }[];
}

interface CertStatusResponse {
  status: "valid" | "expiring" | "expired" | "reality" | "none";
  expires_at?: string;
  days_remaining?: number;
}

function formatBytes(bytes: number): string {
  if (bytes === 0) return "0 B";
  const units = ["B", "KB", "MB", "GB", "TB"];
  const i = Math.floor(Math.log(bytes) / Math.log(1024));
  return `${(bytes / Math.pow(1024, i)).toFixed(1)} ${units[i]}`;
}

export function InboundExpandedPanel({ inboundId, notes }: { inboundId: string; notes?: string }) {
  const toast = useToast();
  const qc = useQueryClient();
  const [linkCopied, setLinkCopied] = useState(false);

  // Fetch traffic stats
  const stats = useQuery({
    queryKey: ["inbound-stats", inboundId],
    queryFn: () => api<InboundTrafficStats>(`/api/inbounds/${inboundId}/stats`),
    enabled: !!inboundId,
  });

  // Fetch cert status
  const cert = useQuery({
    queryKey: ["cert-status", inboundId],
    queryFn: () => api<CertStatusResponse>(`/api/inbounds/${inboundId}/cert-status`),
    enabled: !!inboundId,
  });

  // Clone mutation
  const cloneMutation = useMutation({
    mutationFn: () => api<{ inbound: unknown }>(`/api/inbounds/${inboundId}/clone`, { method: "POST", body: {} }),
    onSuccess: () => {
      toast.success("Inbound cloned successfully");
      qc.invalidateQueries({ queryKey: ["inbounds-fleet"] });
    },
    onError: () => toast.error("Clone failed"),
  });

  async function copyShareLink() {
    try {
      const res = await api<{ link?: string; error?: string }>(`/api/inbounds/${inboundId}/share-link`);
      if (res.link) {
        await navigator.clipboard.writeText(res.link);
        setLinkCopied(true);
        setTimeout(() => setLinkCopied(false), 2000);
        toast.success("Share link copied!");
      } else {
        toast.error(res.error || "Share link not available");
      }
    } catch {
      toast.error("Failed to generate share link");
    }
  }

  const trafficData = stats.data;
  const certData = cert.data;

  return (
    <div className="mt-3 ms-12 space-y-3 border-s-2 border-primary/20 ps-4">
      {/* Actions row */}
      <div className="flex flex-wrap items-center gap-2">
        <Button variant="ghost" size="sm" onClick={() => cloneMutation.mutate()} disabled={cloneMutation.isPending}>
          {cloneMutation.isPending ? "Cloning..." : "🔁 Clone"}
        </Button>
        <Button variant="ghost" size="sm" onClick={copyShareLink}>
          {linkCopied ? <><CheckCircle size={12} className="text-success" /> Copied</> : <><Copy size={12} /> Share Link</>}
        </Button>
      </div>

      {/* Cert Status */}
      {certData && certData.status !== "none" && (
        <div className="flex items-center gap-2">
          <span className="text-[10px] font-semibold text-fg-subtle uppercase">Certificate:</span>
          <span className={cn(
            "inline-flex items-center gap-1 rounded-md px-2 py-0.5 text-[10px] font-semibold border",
            certData.status === "valid" ? "bg-success/10 border-success/20 text-success" :
            certData.status === "expiring" ? "bg-amber-500/10 border-amber-500/20 text-amber-400" :
            certData.status === "expired" ? "bg-danger/10 border-danger/20 text-danger" :
            "bg-purple-500/10 border-purple-500/20 text-purple-400"
          )}>
            {certData.status === "valid" && `✓ Valid (${certData.days_remaining}d remaining)`}
            {certData.status === "expiring" && `⚠ Expiring in ${certData.days_remaining} days`}
            {certData.status === "expired" && "✗ Expired"}
            {certData.status === "reality" && "REALITY (no cert)"}
          </span>
        </div>
      )}

      {/* Notes */}
      {notes && (
        <div className="rounded-lg bg-surface/40 border border-border/30 px-3 py-2">
          <p className="text-[10px] font-semibold text-fg-subtle uppercase mb-1">Notes</p>
          <p className="text-xs text-fg-muted whitespace-pre-wrap">{notes}</p>
        </div>
      )}

      {/* Traffic Stats */}
      {trafficData && trafficData.total > 0 && (
        <div className="space-y-2">
          <div className="flex items-center gap-4 text-xs">
            <span className="text-fg-subtle">↑ {formatBytes(trafficData.upload)}</span>
            <span className="text-fg-subtle">↓ {formatBytes(trafficData.download)}</span>
            <span className="text-fg font-semibold">Total: {formatBytes(trafficData.total)}</span>
          </div>
          {/* Mini daily chart (last 7 days) */}
          {trafficData.daily && trafficData.daily.length > 0 && (
            <div className="flex items-end gap-0.5 h-10">
              {trafficData.daily.slice(0, 7).reverse().map((day, i) => {
                const max = Math.max(...trafficData.daily.slice(0, 7).map(d => d.upload + d.download), 1);
                const h = ((day.upload + day.download) / max) * 100;
                return (
                  <div
                    key={i}
                    className="flex-1 rounded-t bg-primary/40 hover:bg-primary/60 transition-colors"
                    style={{ height: `${Math.max(h, 4)}%` }}
                    title={`${day.date}: ↑${formatBytes(day.upload)} ↓${formatBytes(day.download)}`}
                  />
                );
              })}
            </div>
          )}
        </div>
      )}
      {trafficData && trafficData.total === 0 && (
        <p className="text-[10px] text-fg-subtle">No traffic recorded yet.</p>
      )}
      {stats.isLoading && (
        <p className="text-[10px] text-fg-subtle animate-pulse">Loading stats...</p>
      )}
    </div>
  );
}
