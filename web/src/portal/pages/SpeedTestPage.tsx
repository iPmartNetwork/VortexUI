import { useState } from "react";
import { useMutation } from "@tanstack/react-query";
import { motion, AnimatePresence } from "framer-motion";
import { Zap, Gauge, Globe, Clock, ArrowDownToLine, Signal } from "lucide-react";
import { api } from "@/api/client";
import { Button } from "@/components/ui";
import { GlassCard, StatusBadge } from "@/components/vortexui";
import { cn } from "@/lib/utils";

interface SpeedTestResult {
  node_id: string;
  node_name: string;
  latency_ms: number;
  download_mbps: number;
  upload_mbps?: number;
}

function SpeedGauge({ value, label, color = "primary", size = "md" }: {
  value: number;
  label: string;
  color?: "primary" | "success" | "warning" | "danger";
  size?: "sm" | "md";
}) {
  const pct = Math.min(100, (value / 200) * 100);
  const colors = {
    primary: "from-primary to-accent",
    success: "from-success to-emerald-400",
    warning: "from-warning to-orange-400",
    danger: "from-danger to-red-400",
  };

  return (
    <div className="flex flex-col items-center gap-1.5">
      <span className={cn(
        "font-black tracking-tight tabular-nums",
        size === "md" ? "text-2xl" : "text-lg",
        color === "primary" ? "text-primary" : color === "success" ? "text-success" : color === "warning" ? "text-warning" : "text-danger",
      )}>
        {value.toFixed(1)}
      </span>
      <div className="w-full h-1.5 rounded-full bg-surface-3/60 overflow-hidden">
        <div
          className={cn("h-full rounded-full bg-gradient-to-r transition-all duration-700", colors[color])}
          style={{ width: `${pct}%` }}
        />
      </div>
      <span className="text-[10px] font-semibold text-fg-subtle uppercase tracking-wider flex items-center gap-1">
        {label === "Download" ? <ArrowDownToLine size={11} /> : <Signal size={11} />}
        {label}
      </span>
    </div>
  );
}

export function SpeedTestPage() {
  const [results, setResults] = useState<SpeedTestResult[]>([]);

  const testMut = useMutation({
    mutationFn: (data: { node_id: string; node_endpoint: string }) =>
      api<SpeedTestResult>("/api/v2/portal/speed-test", { method: "POST", body: data }),
    onSuccess: (res) => setResults((p) => [res, ...p.slice(0, 9)]),
  });

  return (
    <div className="space-y-6 animate-page-enter">
      {/* ── Header ── */}
      <div className="relative overflow-hidden rounded-2xl border border-border/60 bg-gradient-to-br from-bg-elevated via-surface to-primary/[0.03] p-5 md:p-6">
        <div className="absolute top-0 end-0 w-56 h-56 rounded-full bg-primary/8 blur-3xl pointer-events-none" />
        <div className="absolute bottom-0 start-0 w-40 h-40 rounded-full bg-accent/8 blur-3xl pointer-events-none" />
        
        <div className="relative z-10 flex items-center gap-3">
          <div className="h-10 w-10 rounded-xl bg-primary/15 flex items-center justify-center">
            <Gauge size={20} className="text-primary" />
          </div>
          <div>
            <h1 className="text-xl md:text-2xl font-black text-fg tracking-tight">
              Speed Test
            </h1>
            <p className="text-[13px] text-fg-muted">
              Test connection speed to available nodes
            </p>
          </div>
        </div>
      </div>

      {/* ── Test Controls ── */}
      <motion.div
        initial={{ opacity: 0, y: 10 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ delay: 0.05 }}
      >
        <GlassCard glow className="!p-5">
          <div className="flex flex-col sm:flex-row items-start sm:items-center justify-between gap-4">
            <div className="space-y-1">
              <h3 className="text-sm font-bold text-fg flex items-center gap-2">
                <Zap size={16} className="text-primary" />
                Run Speed Test
              </h3>
              <p className="text-xs text-fg-muted">
                Measures download speed and latency to the nearest node
              </p>
            </div>
            <Button
              onClick={() => testMut.mutate({ node_id: "", node_endpoint: "" })}
              disabled={testMut.isPending}
              className="gap-2 min-w-[140px]"
            >
              {testMut.isPending ? (
                <span className="flex items-center gap-2">
                  <div className="h-4 w-4 animate-spin rounded-full border-2 border-white border-t-transparent" />
                  Testing...
                </span>
              ) : (
                <span className="flex items-center gap-2">
                  <Gauge size={16} />
                  Run Speed Test
                </span>
              )}
            </Button>
          </div>
        </GlassCard>
      </motion.div>

      {/* ── Latest Result ── */}
      <AnimatePresence>
        {results.length > 0 && (
          <motion.div
            key={results[0].node_id + results[0].latency_ms}
            initial={{ opacity: 0, scale: 0.95 }}
            animate={{ opacity: 1, scale: 1 }}
            exit={{ opacity: 0, scale: 0.95 }}
          >
            <GlassCard glow className="!p-5 relative overflow-hidden">
              <div className="absolute top-0 end-0 w-40 h-40 rounded-full bg-success/5 blur-3xl pointer-events-none" />
              
              <div className="relative z-10">
                <div className="flex items-center justify-between mb-5">
                  <div className="flex items-center gap-2">
                    <div className="h-8 w-8 rounded-lg bg-success/15 flex items-center justify-center">
                      <Signal size={16} className="text-success" />
                    </div>
                    <div>
                      <h3 className="text-sm font-bold text-fg">{results[0].node_name || "Node"}</h3>
                      <span className="text-[10px] text-fg-muted">Latest result</span>
                    </div>
                  </div>
                  <StatusBadge
                    status={results[0].latency_ms < 100 ? "active" : results[0].latency_ms < 200 ? "warning" : "inactive"}
                    label={results[0].latency_ms < 100 ? "Fast" : results[0].latency_ms < 200 ? "Moderate" : "Slow"}
                  />
                </div>

                <div className="grid grid-cols-2 gap-6">
                  <SpeedGauge
                    value={results[0].download_mbps}
                    label="Download"
                    color={results[0].download_mbps > 50 ? "success" : results[0].download_mbps > 20 ? "primary" : "warning"}
                    size="md"
                  />
                  <SpeedGauge
                    value={results[0].upload_mbps ?? results[0].download_mbps * 0.3}
                    label="Upload"
                    color="primary"
                    size="md"
                  />
                </div>

                <div className="mt-4 pt-4 border-t border-border/30 flex items-center justify-center gap-4 text-xs text-fg-muted">
                  <span className="flex items-center gap-1.5">
                    <Clock size={12} className="text-fg-subtle" />
                    {results[0].latency_ms.toFixed(0)}ms latency
                  </span>
                  <span className="flex items-center gap-1.5">
                    <Globe size={12} className="text-fg-subtle" />
                    Node ID: {results[0].node_id.slice(0, 8)}...
                  </span>
                </div>
              </div>
            </GlassCard>
          </motion.div>
        )}
      </AnimatePresence>

      {/* ── History ── */}
      {results.length > 1 && (
        <motion.div
          initial={{ opacity: 0 }}
          animate={{ opacity: 1 }}
          transition={{ delay: 0.2 }}
        >
          <GlassCard glow className="!p-5">
            <h3 className="text-xs font-bold uppercase tracking-wider text-fg-subtle mb-3 flex items-center gap-1.5">
              <Clock size={13} className="text-primary" />
              History ({results.length})
            </h3>
            <div className="space-y-2">
              {results.map((r, i) => {
                const speedColor = r.download_mbps > 50 ? "text-success" : r.download_mbps > 20 ? "text-primary" : "text-warning";
                const latencyColor = r.latency_ms < 100 ? "text-success" : r.latency_ms < 200 ? "text-warning" : "text-danger";
                return (
                  <div
                    key={i}
                    className="flex items-center justify-between rounded-xl border border-border/40 bg-surface-2/30 px-4 py-3 transition-all hover:bg-surface-2/60"
                  >
                    <div className="flex items-center gap-3 min-w-0">
                      <span className="text-[10px] font-bold text-fg-subtle tabular-nums w-5">
                        #{i + 1}
                      </span>
                      <div className="min-w-0">
                        <p className="text-xs font-semibold text-fg truncate">
                          {r.node_name || `Node ${r.node_id.slice(0, 6)}`}
                        </p>
                        <span className={cn("text-[10px] font-mono", latencyColor)}>
                          {r.latency_ms.toFixed(0)}ms
                        </span>
                      </div>
                    </div>
                    <span className={cn("text-sm font-black tabular-nums", speedColor)}>
                      {r.download_mbps.toFixed(1)} <span className="text-[9px] font-semibold text-fg-subtle">Mbps</span>
                    </span>
                  </div>
                );
              })}
            </div>
          </GlassCard>
        </motion.div>
      )}

      {/* ── Empty State ── */}
      {results.length === 0 && !testMut.isPending && (
        <div className="flex flex-col items-center justify-center py-16 gap-3">
          <div className="h-14 w-14 rounded-full bg-surface-2/50 flex items-center justify-center">
            <Gauge size={26} className="text-fg-subtle" />
          </div>
          <p className="text-sm text-fg-muted">No speed tests yet.</p>
          <p className="text-xs text-fg-subtle">Click <strong>Run Speed Test</strong> to check your connection.</p>
        </div>
      )}
    </div>
  );
}
