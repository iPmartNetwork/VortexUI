import { useState } from "react";
import { useQuery } from "@tanstack/react-query";
import { api } from "@/api/client";
import { Button, Card, PageHeader, Select } from "@/components/ui";
import { formatBytes } from "@/lib/utils";
import { useI18n } from "@/i18n/i18n";

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
  const { t } = useI18n();
  const [range, setRange] = useState("7d");
  const rangeMap: Record<string, number> = { "1d": 1, "7d": 7, "30d": 30 };

  const from = Math.floor((Date.now() - (rangeMap[range] ?? 7) * 86400000) / 1000);
  const to = Math.floor(Date.now() / 1000);

  const { data, isLoading, isError } = useQuery({
    queryKey: ["analytics", range],
    queryFn: () => api<AnalyticsData>("/api/analytics", { query: { from: String(from), to: String(to) } }),
  });

  return (
    <div className="space-y-6 animate-fade-in">
      <div className="flex items-center justify-between">
        <PageHeader title={t("analytics.title")} subtitle={t("analytics.subtitle")} />
        <div className="flex gap-2">
          <Select value={range} onChange={(e) => setRange(e.target.value)}>
            <option value="1d">Last 24h</option>
            <option value="7d">Last 7 days</option>
            <option value="30d">Last 30 days</option>
          </Select>
          <Button onClick={() => {
            const token = localStorage.getItem("vortex.token") || "";
            window.open(`/api/analytics/export?from=${from}&to=${to}&access_token=${encodeURIComponent(token)}`, "_blank");
          }}>
            {t("analytics.export")}
          </Button>
        </div>
      </div>

      {isLoading && <div className="text-center text-fg-muted py-8">{t("common.loading")}</div>}
      {isError && <div className="text-center text-fg-muted py-8">{t("analytics.error")}</div>}

      {data && (data.geo_breakdown?.length || data.top_users?.length || data.total_up || data.total_down) && (
        <>
          {/* Summary cards */}
          <div className="grid grid-cols-1 gap-4 md:grid-cols-3">
            <Card className="space-y-2">
              <div className="text-xs text-fg-subtle uppercase">{t("analytics.totalUp")}</div>
              <div className="text-lg font-bold text-fg">{formatBytes(data.total_up, false)}</div>
            </Card>
            <Card className="space-y-2">
              <div className="text-xs text-fg-subtle uppercase">{t("analytics.totalDown")}</div>
              <div className="text-lg font-bold text-fg">{formatBytes(data.total_down, false)}</div>
            </Card>
            <Card className="space-y-2">
              <div className="text-xs text-fg-subtle uppercase">{t("analytics.countries")}</div>
              <div className="text-lg font-bold text-fg">{data.geo_breakdown?.length ?? 0}</div>
            </Card>
          </div>

          {/* Geo breakdown table */}
          <Card>
            <h3 className="text-sm font-bold text-fg mb-3">{t("analytics.byCountry")}</h3>
            <div className="overflow-x-auto">
              <table className="w-full text-sm">
                <thead>
                  <tr className="border-b border-border/40 text-xs text-fg-subtle">
                    <th className="py-2 text-left">Country</th>
                    <th className="py-2 text-center">Connections</th>
                    <th className="py-2 text-right">Upload</th>
                    <th className="py-2 text-right">Download</th>
                  </tr>
                </thead>
                <tbody>
                  {data.geo_breakdown?.slice(0, 20).map((g, i) => (
                    <tr key={i} className="border-b border-border/20">
                      <td className="py-2 text-fg font-medium">{g.country || "Unknown"}</td>
                      <td className="py-2 text-center text-fg-muted">{g.connections.toLocaleString()}</td>
                      <td className="py-2 text-right text-fg-muted">{formatBytes(g.bytes_up, false)}</td>
                      <td className="py-2 text-right text-fg-muted">{formatBytes(g.bytes_down, false)}</td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          </Card>

          {/* Top users */}
          <Card>
            <h3 className="text-sm font-bold text-fg mb-3">{t("analytics.topUsers")}</h3>
            <div className="overflow-x-auto">
              <table className="w-full text-sm">
                <thead>
                  <tr className="border-b border-border/40 text-xs text-fg-subtle">
                    <th className="py-2 text-left">#</th>
                    <th className="py-2 text-left">Username</th>
                    <th className="py-2 text-right">Used Traffic</th>
                  </tr>
                </thead>
                <tbody>
                  {data.top_users?.map((u, i) => (
                    <tr key={u.user_id} className="border-b border-border/20">
                      <td className="py-2 text-fg-muted">{i + 1}</td>
                      <td className="py-2 text-fg font-medium">{u.username}</td>
                      <td className="py-2 text-right text-fg-muted">{formatBytes(u.used_traffic, false)}</td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          </Card>

          {/* Peak hours */}
          <Card>
            <h3 className="text-sm font-bold text-fg mb-3">{t("analytics.peakHours")}</h3>
            <div className="flex items-end gap-1 h-32">
              {data.peak_hours?.map((p) => {
                const maxBytes = Math.max(...(data.peak_hours?.map(h => h.bytes_total) ?? [1]));
                const height = maxBytes > 0 ? (p.bytes_total / maxBytes) * 100 : 0;
                return (
                  <div key={p.hour} className="flex-1 flex flex-col items-center gap-1">
                    <div className="w-full grad-bg rounded-t-sm" style={{ height: `${height}%`, minHeight: "2px" }} title={`${p.hour}:00 — ${formatBytes(p.bytes_total, false)}`} />
                    <span className="text-[9px] text-fg-subtle">{p.hour}</span>
                  </div>
                );
              })}
            </div>
          </Card>
        </>
      )}
      {!isLoading && !isError && data && !data.geo_breakdown?.length && !data.top_users?.length && !data.total_up && !data.total_down && (
        <div className="text-center text-fg-muted py-8">{t("analytics.noData")}</div>
      )}
    </div>
  );
}
