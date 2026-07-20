import { useQuery } from "@tanstack/react-query";
import { Activity, ArrowRightLeft, Globe, Wifi } from "lucide-react";
import { portalApi } from "./portalApi";
import { GlassCard, StatsCard } from "@/components/veltrix";

interface SwitchSummary {
  total_switches: number;
  by_protocol: Record<string, number>;
  by_isp: Record<string, number>;
}

interface ConnectionStats {
  live_connections?: number;
  live_tracking?: boolean;
  switches_24h?: number;
  by_protocol?: Record<string, number>;
  by_isp?: Record<string, number>;
  current_protocol?: string;
}

export function PortalProtocolStatus() {
  const stats = useQuery({
    queryKey: ["portal-connection-stats"],
    queryFn: () => portalApi<ConnectionStats>("/api/portal/connection-stats"),
    refetchInterval: 15_000,
  });

  const history = useQuery({
    queryKey: ["portal-switch-history"],
    queryFn: () => portalApi<{ summary: SwitchSummary | null }>("/api/portal/switch-history"),
    refetchInterval: 60_000,
  });

  const conn = stats.data;
  const summary = history.data?.summary;

  return (
    <div className="space-y-4">
      {/* Real-time Stats */}
      <div className="grid grid-cols-2 gap-3 lg:grid-cols-4">
        <StatsCard
          title="Live Connections"
          value={conn?.live_connections ?? "—"}
          icon={<Wifi size={18} />}
          color="green"
        />
        <StatsCard
          title="Active Protocol"
          value={conn?.current_protocol?.toUpperCase() ?? "—"}
          icon={<Activity size={18} />}
          color="blue"
        />
        <StatsCard
          title="Switches (24h)"
          value={conn?.switches_24h ?? 0}
          icon={<ArrowRightLeft size={18} />}
          color="orange"
        />
        <StatsCard
          title="Week Total"
          value={summary?.total_switches ?? 0}
          icon={<Globe size={18} />}
          color="cyan"
        />
      </div>

      {/* Protocol Breakdown */}
      {summary && summary.total_switches > 0 && (
        <GlassCard hover={false} className="!p-5">
          <h3 className="text-xs font-semibold uppercase tracking-wider text-fg-subtle mb-3">
            Protocol Switch Breakdown (7 days)
          </h3>
          <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
            {/* By Protocol */}
            <div>
              <p className="text-[10px] text-fg-subtle mb-2 font-medium">
                By Protocol
              </p>
              <div className="space-y-1.5">
                {Object.entries(summary.by_protocol ?? {})
                  .sort(([, a], [, b]) => b - a)
                  .slice(0, 5)
                  .map(([proto, count]) => (
                    <div key={proto} className="flex items-center justify-between">
                      <span className="text-xs text-fg-muted">{proto}</span>
                      <div className="flex items-center gap-2">
                        <div className="w-16 h-1.5 rounded-full bg-surface-3 overflow-hidden">
                          <div
                            className="h-full rounded-full grad-bg"
                            style={{
                              width: `${Math.min(100, (count / summary.total_switches) * 100)}%`,
                            }}
                          />
                        </div>
                        <span className="text-xs font-semibold text-fg tabular-nums w-6 text-end">
                          {count}
                        </span>
                      </div>
                    </div>
                  ))}
              </div>
            </div>

            {/* By ISP */}
            {Object.keys(summary.by_isp ?? {}).length > 0 && (
              <div>
                <p className="text-[10px] text-fg-subtle mb-2 font-medium">
                  By ISP
                </p>
                <div className="space-y-1.5">
                  {Object.entries(summary.by_isp ?? {})
                    .sort(([, a], [, b]) => b - a)
                    .slice(0, 5)
                    .map(([isp, count]) => (
                      <div key={isp} className="flex items-center justify-between">
                        <span className="text-xs text-fg-muted">{isp}</span>
                        <span className="text-xs font-semibold text-fg tabular-nums">{count}</span>
                      </div>
                    ))}
                </div>
              </div>
            )}
          </div>
        </GlassCard>
      )}
    </div>
  );
}
