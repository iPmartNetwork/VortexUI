import { useState } from "react";
import { useQuery } from "@tanstack/react-query";
import { Download, Globe2, TrendingUp, Users } from "lucide-react";
import { api } from "@/api/client";
import { Button, Select } from "@/components/ui";
import { GlassCard, StatsCard } from "@/components/veltrix";
import { formatBytes } from "@/lib/utils";
import { useI18n } from "@/i18n/i18n";
import { useTitle } from "@/lib/useTitle";

interface GeoPoint {
  country: string;
  connections: number;
  bytes_up: number;
  bytes_down: number;
}

interface UserRank {
  user_id: string;
  username: string;
  used_traffic: number;
}

interface PeakHour {
  hour: number;
  connections: number;
  bytes_total: number;
}

interface AnalyticsData {
  geo_breakdown: GeoPoint[];
  top_users: UserRank[];
  peak_hours: PeakHour[];
  total_up: number;
  total_down: number;
}

export function Analytics() {
  useTitle("Analytics");
  const { t } = useI18n();
  const [range, setRange] = useState("7d");
  const rangeMap: Record<string, number> = { "1d": 1, "7d": 7, "30d": 30 };

  const from = Math.floor((Date.now() - (rangeMap[range] ?? 7) * 86400000) / 1000);
  const to = Math.floor(Date.now() / 1000);

  const { data, isLoading, isError } = useQuery({
    queryKey: ["analytics", range],
    queryFn: () => api<AnalyticsData>("/api/analytics", { query: { from: String(from), to: String(to) } }),
  });

  const hasData = !!data && !!(data.geo_breakdown?.length || data.top_users?.length || data.total_up || data.total_down);

  return (
    <div className="space-y-5 animate-page-enter">
      <div className="flex flex-col lg:flex-row lg:items-start justify-between gap-4">
        <div>
          <h1 className="text-2xl font-bold text-fg tracking-tight">{t("analytics.title")}</h1>
          <p className="text-sm text-fg-muted mt-1 max-w-2xl">{t("analytics.subtitle")}</p>
        </div>
        <div className="flex gap-2 flex-shrink-0">
          <Select value={range} onChange={(e) => setRange(e.target.value)}>
            <option value="1d">{t("analytics.last24h")}</option>
            <option value="7d">{t("analytics.last7d")}</option>
            <option value="30d">{t("analytics.last30d")}</option>
          </Select>
          <Button
            variant="outline"
            onClick={() => {
              const token = localStorage.getItem("vortex.token") || "";
              window.open(`/api/analytics/export?from=${from}&to=${to}&access_token=${encodeURIComponent(token)}`, "_blank");
            }}
          >
            <Download size={14} /> {t("analytics.export")}
          </Button>
        </div>
      </div>

      {isLoading && <p className="text-sm text-fg-muted text-center py-8">{t("common.loading")}</p>}
      {isError && <p className="text-sm text-fg-muted text-center py-8">{t("analytics.error")}</p>}

      {hasData && data && (
        <>
          <div className="grid grid-cols-1 gap-4 md:grid-cols-3">
            <StatsCard title={t("analytics.totalUp")} value={formatBytes(data.total_up, false)} icon={<TrendingUp size={18} />} color="green" />
            <StatsCard title={t("analytics.totalDown")} value={formatBytes(data.total_down, false)} icon={<Download size={18} />} color="cyan" />
            <StatsCard title={t("analytics.countries")} value={data.geo_breakdown?.length ?? 0} icon={<Globe2 size={18} />} color="purple" />
          </div>

          <GlassCard hover={false} className="!p-0 overflow-hidden">
            <div className="px-4 pt-4">
              <h3 className="text-sm font-bold text-fg">{t("analytics.byCountry")}</h3>
            </div>
            <div className="overflow-x-auto mt-2">
              <table className="w-full text-sm">
                <thead>
                  <tr className="border-b border-border/40 text-[11px] uppercase tracking-wide text-fg-subtle bg-surface-2/30">
                    <th className="py-3 px-4 text-left">{t("analytics.colCountry")}</th>
                    <th className="py-3 px-4 text-center">{t("analytics.colConnections")}</th>
                    <th className="py-3 px-4 text-right">{t("analytics.colUpload")}</th>
                    <th className="py-3 px-4 text-right">{t("analytics.colDownload")}</th>
                  </tr>
                </thead>
                <tbody>
                  {data.geo_breakdown?.slice(0, 20).map((g, i) => (
                    <tr key={i} className="border-b border-border/20 hover:bg-surface-2/40">
                      <td className="py-3 px-4 font-medium text-fg">{g.country || t("analytics.unknown")}</td>
                      <td className="py-3 px-4 text-center tabular-nums text-fg-muted">{g.connections.toLocaleString()}</td>
                      <td className="py-3 px-4 text-right tabular-nums text-fg-muted">{formatBytes(g.bytes_up, false)}</td>
                      <td className="py-3 px-4 text-right tabular-nums text-fg-muted">{formatBytes(g.bytes_down, false)}</td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          </GlassCard>

          <div className="grid grid-cols-1 gap-4 lg:grid-cols-2">
            <GlassCard hover={false} className="!p-0 overflow-hidden">
              <div className="px-4 pt-4 flex items-center gap-2">
                <Users size={14} className="text-primary" />
                <h3 className="text-sm font-bold text-fg">{t("analytics.topUsers")}</h3>
              </div>
              <div className="overflow-x-auto mt-2">
                <table className="w-full text-sm">
                  <thead>
                    <tr className="border-b border-border/40 text-[11px] uppercase tracking-wide text-fg-subtle bg-surface-2/30">
                      <th className="py-3 px-4 text-left">{t("analytics.colRank")}</th>
                      <th className="py-3 px-4 text-left">{t("analytics.colUsername")}</th>
                      <th className="py-3 px-4 text-right">{t("analytics.colUsedTraffic")}</th>
                    </tr>
                  </thead>
                  <tbody>
                    {data.top_users?.map((u, i) => (
                      <tr key={u.user_id} className="border-b border-border/20 hover:bg-surface-2/40">
                        <td className="py-3 px-4 text-fg-subtle tabular-nums">{i + 1}</td>
                        <td className="py-3 px-4 font-medium text-fg">{u.username}</td>
                        <td className="py-3 px-4 text-right tabular-nums text-fg-muted">{formatBytes(u.used_traffic, false)}</td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            </GlassCard>

            <GlassCard hover={false} className="!p-4 space-y-3">
              <h3 className="text-sm font-bold text-fg">{t("analytics.peakHours")}</h3>
              <div className="flex items-end gap-1 h-40">
                {data.peak_hours?.map((p) => {
                  const maxBytes = Math.max(...(data.peak_hours?.map((h) => h.bytes_total) ?? [1]));
                  const height = maxBytes > 0 ? (p.bytes_total / maxBytes) * 100 : 0;
                  return (
                    <div key={p.hour} className="flex-1 flex flex-col items-center gap-1.5 group">
                      <div
                        className="w-full rounded-t-md bg-gradient-to-t from-primary/70 to-primary transition-all group-hover:from-primary group-hover:to-primary/80"
                        style={{ height: `${height}%`, minHeight: "2px" }}
                        title={`${p.hour}:00 — ${formatBytes(p.bytes_total, false)}`}
                      />
                      <span className="text-[9px] text-fg-subtle tabular-nums">{p.hour}</span>
                    </div>
                  );
                })}
              </div>
            </GlassCard>
          </div>
        </>
      )}

      {!isLoading && !isError && data && !hasData && (
        <p className="text-sm text-fg-muted text-center py-8">{t("analytics.noData")}</p>
      )}
    </div>
  );
}
