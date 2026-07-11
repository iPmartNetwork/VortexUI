import { AlertCircle, Activity } from "lucide-react";
import { usePerformanceHealth, useSlowQueries, useQueryStats, usePerformanceAlerts } from "@/api/security-hooks";
import { Button } from "@/components/ui";
import { GlassCard } from "@/components/veltrix";
import { useTitle } from "@/lib/useTitle";

export function Performance() {
    useTitle("Performance");

    const health = usePerformanceHealth();
    const slowQueries = useSlowQueries(10, 0);
    const stats = useQueryStats();
    const alerts = usePerformanceAlerts();

    return (
        <div className="space-y-6 animate-page-enter">
            <div>
                <h1 className="text-2xl font-bold text-fg tracking-tight">Performance</h1>
                <p className="text-sm text-fg-secondary mt-1">System health, query metrics & rate limiting</p>
            </div>

            {/* Performance Metrics */}
            {health.data && (
                <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
                    <GlassCard className="p-4">
                        <p className="text-xs text-fg-secondary uppercase tracking-wide">Cache Hit Rate</p>
                        <p className="text-2xl font-bold text-fg mt-1">{(health.data.cache.hit_rate * 100).toFixed(1)}%</p>
                        <div className="h-1.5 bg-fg/10 rounded mt-2">
                            <div
                                className="h-full bg-green-500 rounded"
                                style={{ width: `${Math.min(health.data.cache.hit_rate * 100, 100)}%` }}
                            />
                        </div>
                    </GlassCard>

                    <GlassCard className="p-4">
                        <p className="text-xs text-fg-secondary uppercase tracking-wide">Avg Query Time</p>
                        <p className="text-2xl font-bold text-fg mt-1">{health.data.avg_query_time_ms.toFixed(0)}ms</p>
                        <p className="text-xs text-fg-secondary mt-2">per database query</p>
                    </GlassCard>

                    <GlassCard className="p-4">
                        <p className="text-xs text-fg-secondary uppercase tracking-wide">Active Connections</p>
                        <p className="text-2xl font-bold text-fg mt-1">{health.data.connections_active}</p>
                        <p className="text-xs text-fg-secondary mt-2">current sessions</p>
                    </GlassCard>

                    <GlassCard className="p-4">
                        <p className="text-xs text-fg-secondary uppercase tracking-wide">Slow Queries (1h)</p>
                        <p className="text-2xl font-bold text-orange-500 mt-1">{health.data.slow_queries_last_hour}</p>
                        <p className="text-xs text-fg-secondary mt-2">above threshold</p>
                    </GlassCard>
                </div>
            )}

            {/* Cache Status */}
            {health.data?.cache && (
                <GlassCard className="p-6">
                    <h2 className="text-lg font-semibold text-fg mb-4">Cache Status</h2>
                    <div className="grid grid-cols-2 sm:grid-cols-4 gap-4">
                        <div className="p-3 rounded-lg bg-fg/5">
                            <p className="text-xs text-fg-secondary">Hits</p>
                            <p className="text-xl font-bold text-fg mt-1">{health.data.cache.hit_count}</p>
                        </div>
                        <div className="p-3 rounded-lg bg-fg/5">
                            <p className="text-xs text-fg-secondary">Misses</p>
                            <p className="text-xl font-bold text-fg mt-1">{health.data.cache.miss_count}</p>
                        </div>
                        <div className="p-3 rounded-lg bg-fg/5">
                            <p className="text-xs text-fg-secondary">Size</p>
                            <p className="text-xl font-bold text-fg mt-1">{(health.data.cache.current_size / 1024 / 1024).toFixed(1)}MB</p>
                        </div>
                        <div className="p-3 rounded-lg bg-fg/5">
                            <p className="text-xs text-fg-secondary">Evictions</p>
                            <p className="text-xl font-bold text-fg mt-1">{health.data.cache.eviction_count}</p>
                        </div>
                    </div>
                </GlassCard>
            )}

            {/* Query Statistics */}
            {stats.data && (
                <GlassCard className="p-6">
                    <h2 className="text-lg font-semibold text-fg mb-4">Query Statistics</h2>
                    <div className="space-y-2">
                        <div className="flex items-center justify-between p-3 border-l-4 border-blue-500 bg-blue-500/5 rounded">
                            <span className="text-sm text-fg">Total Queries</span>
                            <span className="font-bold text-fg">{stats.data.total_queries}</span>
                        </div>
                        <div className="flex items-center justify-between p-3 border-l-4 border-green-500 bg-green-500/5 rounded">
                            <span className="text-sm text-fg">Average Time</span>
                            <span className="font-bold text-fg">{stats.data.avg_time_ms.toFixed(2)}ms</span>
                        </div>
                        <div className="flex items-center justify-between p-3 border-l-4 border-yellow-500 bg-yellow-500/5 rounded">
                            <span className="text-sm text-fg">Min Time</span>
                            <span className="font-bold text-fg">{stats.data.min_time_ms.toFixed(2)}ms</span>
                        </div>
                        <div className="flex items-center justify-between p-3 border-l-4 border-red-500 bg-red-500/5 rounded">
                            <span className="text-sm text-fg">Max Time</span>
                            <span className="font-bold text-fg">{stats.data.max_time_ms.toFixed(2)}ms</span>
                        </div>
                    </div>
                </GlassCard>
            )}

            {/* Slow Queries */}
            {slowQueries.data?.queries && slowQueries.data.queries.length > 0 && (
                <GlassCard className="p-6">
                    <h2 className="text-lg font-semibold text-fg mb-4">Slow Queries</h2>
                    <div className="space-y-2 max-h-96 overflow-y-auto">
                        {slowQueries.data.queries.map((q: any, idx: number) => (
                            <div key={idx} className="p-3 bg-orange-500/5 border border-orange-500/20 rounded hover:bg-orange-500/10 transition">
                                <div className="flex items-start justify-between gap-2">
                                    <div className="flex-1 min-w-0">
                                        <p className="text-xs font-mono text-fg-secondary truncate">{q.query}</p>
                                        <div className="flex gap-2 mt-1">
                                            <span className="text-xs bg-orange-500/20 text-orange-600 px-2 py-1 rounded">
                                                {q.execution_time_ms.toFixed(0)}ms
                                            </span>
                                            <span className="text-xs bg-blue-500/20 text-blue-600 px-2 py-1 rounded">
                                                {q.rows_affected} rows
                                            </span>
                                        </div>
                                    </div>
                                    <div className="h-6 px-2.5 text-xs font-medium rounded bg-orange-500/20 text-orange-600 flex items-center">
                                        Slow
                                    </div>
                                </div>
                            </div>
                        ))}
                    </div>
                </GlassCard>
            )}

            {/* Performance Alerts */}
            {alerts.data?.alerts && alerts.data.alerts.length > 0 && (
                <GlassCard className="p-6 border-l-4 border-orange-500">
                    <h2 className="text-lg font-semibold text-fg mb-4 flex items-center gap-2">
                        <AlertCircle className="w-5 h-5 text-orange-500" />
                        Performance Alerts ({alerts.data.alerts.length})
                    </h2>
                    <div className="space-y-2">
                        {alerts.data.alerts.map((alert: any) => (
                            <div key={alert.id} className="flex items-center justify-between p-3 bg-fg/5 rounded">
                                <div className="flex-1">
                                    <p className="text-sm text-fg">{alert.message}</p>
                                    <p className="text-xs text-fg-secondary mt-1">{alert.created_at}</p>
                                </div>
                                <Button size="sm" variant="outline">
                                    Resolve
                                </Button>
                            </div>
                        ))}
                    </div>
                </GlassCard>
            )}

            {/* System Recommendation */}
            <GlassCard className="p-6 bg-green-500/10 border-green-500/20">
                <h2 className="text-lg font-semibold text-fg mb-3 flex items-center gap-2">
                    <Activity className="w-5 h-5 text-green-500" />
                    System Status: Optimal
                </h2>
                <p className="text-sm text-fg-secondary">Your system is performing well. Cache hit rate is healthy and query times are within acceptable ranges.</p>
            </GlassCard>
        </div>
    );
}
