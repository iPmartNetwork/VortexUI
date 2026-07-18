import { Wifi, Globe, BarChart3 } from "lucide-react";
import { GlassCard } from "@/components/veltrix";
import { useTitle } from "@/lib/useTitle";

export function ConnectionQuality() {
  useTitle("Connection Quality");
  return (
    <div className="space-y-5 animate-page-enter">
      <div>
        <h1 className="text-2xl font-bold text-fg tracking-tight">Connection Quality</h1>
        <p className="text-sm text-fg-muted mt-1">
          Monitor ISP performance, connection paths, and real-time bandwidth
        </p>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
        <GlassCard className="!p-4 space-y-2">
          <div className="flex items-center gap-2">
            <div className="h-8 w-8 rounded-lg bg-primary/10 flex items-center justify-center text-primary">
              <BarChart3 size={16} />
            </div>
            <h3 className="text-sm font-bold text-fg">ISP Heatmap</h3>
          </div>
          <p className="text-xs text-fg-muted">Quality metrics per ISP x hour of day</p>
          <div className="h-24 rounded-lg bg-surface-2/30 border border-border/20 flex items-center justify-center">
            <span className="text-[10px] text-fg-subtle">Coming soon</span>
          </div>
        </GlassCard>

        <GlassCard className="!p-4 space-y-2">
          <div className="flex items-center gap-2">
            <div className="h-8 w-8 rounded-lg bg-green-500/10 flex items-center justify-center text-green-400">
              <Wifi size={16} />
            </div>
            <h3 className="text-sm font-bold text-fg">Bandwidth</h3>
          </div>
          <p className="text-xs text-fg-muted">Real-time per-user bandwidth tracking</p>
          <div className="h-24 rounded-lg bg-surface-2/30 border border-border/20 flex items-center justify-center">
            <span className="text-[10px] text-fg-subtle">Coming soon</span>
          </div>
        </GlassCard>

        <GlassCard className="!p-4 space-y-2">
          <div className="flex items-center gap-2">
            <div className="h-8 w-8 rounded-lg bg-purple-500/10 flex items-center justify-center text-purple-400">
              <Globe size={16} />
            </div>
            <h3 className="text-sm font-bold text-fg">Connection Path</h3>
          </div>
          <p className="text-xs text-fg-muted">Visualize traffic hop-by-hop path</p>
          <div className="h-24 rounded-lg bg-surface-2/30 border border-border/20 flex items-center justify-center">
            <span className="text-[10px] text-fg-subtle">Coming soon</span>
          </div>
        </GlassCard>
      </div>
    </div>
  );
}
