import { useState, lazy, Suspense } from "react";
import { useQuery, useQueryClient } from "@tanstack/react-query";
import {
  Activity, AlertTriangle, Globe, TrendingUp, Zap,
  Server, Shield, RefreshCw, MapPin, DollarSign, BarChart3, Wifi,
  Sparkles,
} from "lucide-react";
import { api } from "@/api/client";
import { Button, Select } from "@/components/ui";
import { GlassCard, StatsCard } from "@/components/vortexui";
import { ProtocolDonutChart, type ProtocolSlice } from "@/components/vortexui/ProtocolDonutChart";
import { useToast } from "@/components/toast";
import { RevenueAreaChart } from "@/components/Charts";
import { useTitle } from "@/lib/useTitle";
import { cn } from "@/lib/utils";
import type { GeoNode } from "@/pages/dashboardPro/GeoNodeMap";

const LazyGeoNodeMap = lazy(() => import("@/pages/dashboardPro/GeoNodeMap").then((m) => ({ default: m.GeoNodeMap })));

// ---------- Types ----------

interface DiagnosticCard {
  severity: string;
  title: string;
  description: string;
  actions: string[];
}

interface CertHealthStatus {
  domain: string;
  expires_at: string;
  valid: boolean;
}

interface DailyCheckWidget {
  nodes_online: number;
  nodes_total: number;
  traffic_anomaly: boolean;
  cert_status: CertHealthStatus[];
  diagnostics: DiagnosticCard[];
}

interface HeatmapCell {
  day: number;
  hour: number;
  score: number;
}

interface ISPHeatmap {
  isp: string;
  cells: HeatmapCell[];
}

interface RevenueDataPoint {
  date: string;
  income: number;
  expense: number;
}

interface RevenueReport {
  total_income: number;
  total_expense: number;
  profit: number;
  time_series: RevenueDataPoint[];
}

interface FormatCount {
  format: string;
  count: number;
}

interface ISPCount {
  isp: string;
  count: number;
}

interface SubAnalyticsReport {
  by_format: FormatCount[];
  by_isp: ISPCount[];
}

// ---------- Main Component ----------

export function DashboardPro() {
  useTitle("Dashboard Pro");
  const toast = useToast();
  const qc = useQueryClient();
  const [ispName, setIspName] = useState("MCI");
  const [range, setRange] = useState("30d");

  const rangeMs: Record<string, number> = { "7d": 7, "14d": 14, "30d": 30 };
  const days = rangeMs[range] ?? 30;
  const from = new Date(Date.now() - days * 86400000).toISOString().split("T")[0];
  const to = new Date().toISOString().split("T")[0];

  const { data: dailyCheck } = useQuery({
    queryKey: ["dashboard-pro", "daily-check"],
    queryFn: () => api<{ daily_check: DailyCheckWidget }>("/api/v2/dashboard/daily-check"),
    refetchInterval: 60_000,
  });

  const { data: heatmapData } = useQuery({
    queryKey: ["dashboard-pro", "isp-heatmap", ispName],
    queryFn: () => api<{ heatmap: ISPHeatmap }>("/api/v2/dashboard/isp-heatmap", { query: { isp: ispName, days: "7" } }),
  });

  const { data: geoData } = useQuery({
    queryKey: ["dashboard-pro", "geo-map"],
    queryFn: () => api<{ nodes: GeoNode[] }>("/api/v2/dashboard/geo-map"),
  });

  const { data: revenueData } = useQuery({
    queryKey: ["dashboard-pro", "revenue", from, to],
    queryFn: () => api<{ revenue: RevenueReport }>("/api/v2/dashboard/revenue", { query: { from, to } }),
  });

  const { data: subData } = useQuery({
    queryKey: ["dashboard-pro", "sub-analytics", from, to],
    queryFn: () => api<{ sub_analytics: SubAnalyticsReport }>("/api/v2/dashboard/sub-analytics", { query: { from, to } }),
  });

  const check = dailyCheck?.daily_check;
  const heatmap = heatmapData?.heatmap;
  const nodes = geoData?.nodes ?? [];
  const revenue = revenueData?.revenue;
  const subAnalytics = subData?.sub_analytics;

  return (
    <div className="space-y-6 animate-page-enter">
      {/* Header */}
      <div className="flex flex-col lg:flex-row lg:items-center justify-between gap-4 relative overflow-hidden rounded-2xl border border-border/70 bg-gradient-to-br from-bg-elevated via-surface to-primary/[0.02] p-5 md:p-6 shadow-xl">
        <div className="absolute inset-0 bg-dot-pattern animate-pattern-drift opacity-20 pointer-events-none" />
        <div className="absolute top-0 end-0 w-72 h-72 rounded-full bg-primary/8 blur-3xl pointer-events-none" />
        <div className="absolute bottom-0 start-0 w-56 h-56 rounded-full bg-accent/8 blur-3xl pointer-events-none" />
        <div className="relative z-10 flex flex-col xl:flex-row xl:items-start justify-between gap-6 w-full">
        <div>
          <div className="flex items-center gap-3">
            <h1 className="text-2xl font-bold text-fg tracking-tight">Dashboard Pro</h1>
            <span className="inline-flex items-center gap-1 rounded-md bg-primary/10 px-2 py-0.5 text-[10px] font-bold text-primary border border-primary/20">
              <Sparkles size={10} /> ADVANCED
            </span>
          </div>
          <p className="text-sm text-fg-muted mt-1">Advanced monitoring, analytics, and daily workflow</p>
        </div>
        <div className="flex gap-2 flex-shrink-0">
          <Select value={range} onChange={(e) => setRange(e.target.value)}>
            <option value="7d">Last 7 days</option>
            <option value="14d">Last 14 days</option>
            <option value="30d">Last 30 days</option>
          </Select>
        </div>
      </div>
    </div>

      {/* Daily Check Status Cards */}
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
        <StatsCard
          title="Nodes Online"
          value={check ? `${check.nodes_online}/${check.nodes_total}` : "—"}
          icon={<Server size={18} />}
          color={check && check.nodes_online === check.nodes_total ? "green" : "red"}
        />
        <StatsCard
          title="Traffic Status"
          value={check?.traffic_anomaly ? "Anomaly" : "Normal"}
          icon={<Activity size={18} />}
          color={check?.traffic_anomaly ? "orange" : "green"}
        />
        <StatsCard
          title="Certificates"
          value={check?.cert_status?.length ? `${check.cert_status.filter(c => c.valid).length} valid` : "OK"}
          icon={<Shield size={18} />}
          color="blue"
        />
        <StatsCard
          title="Diagnostics"
          value={check?.diagnostics?.length ? String(check.diagnostics.length) : "0"}
          icon={<AlertTriangle size={18} />}
          color={check?.diagnostics?.length ? "orange" : "green"}
        />
      </div>

      {/* Quick Actions */}
      <GlassCard className="p-4" glow>
        <h3 className="text-sm font-semibold text-fg mb-3 flex items-center gap-2">
          <span className="p-1 rounded-lg bg-primary/15 border border-primary/30">
            <Zap size={14} className="text-primary" />
          </span>
          Quick Actions
        </h3>
        <div className="flex flex-wrap gap-2">
          <Button variant="glass" size="sm" onClick={async () => { await api('/api/v2/dashboard/actions/refresh-nodes', { method: 'POST' }); toast.success("Nodes refreshed"); qc.invalidateQueries({ queryKey: ["dashboard-pro"] }); }}><RefreshCw size={12} /> Refresh Nodes</Button>
          <Button variant="glass" size="sm" onClick={async () => { await api('/api/v2/dashboard/actions/restart-cores', { method: 'POST' }); toast.success("Restart signal sent"); }}><Server size={12} /> Restart Cores</Button>
          <Button variant="glass" size="sm" onClick={async () => { await api('/api/v2/dashboard/actions/check-certs', { method: 'POST' }); toast.success("Certificates valid"); }}><Shield size={12} /> Check Certs</Button>
          <Button variant="glass" size="sm" onClick={async () => { await api('/api/v2/dashboard/actions/update-geo', { method: 'POST' }); toast.success("Geo data updated"); }}><Globe size={12} /> Update Geo</Button>
          <Button variant="glass" size="sm" onClick={async () => {
            const host = prompt("Enter server IP or domain to check:");
            if (!host) return;
            const port = prompt("Port (default 443):", "443");
            try {
              const res = await api<{ verdict: string; ok: number; total: number }>("/api/v2/reachability/check", { method: "POST", body: { host, port: parseInt(port || "443") } });
              toast.success(`Iran reachability: ${res.verdict} (${res.ok}/${res.total} nodes)`);
            } catch { toast.error("Check failed"); }
          }}><Wifi size={12} /> Iran Check</Button>
        </div>
      </GlassCard>

      {/* Diagnostic Cards */}
      {check?.diagnostics && check.diagnostics.length > 0 && (
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          {check.diagnostics.map((diag, i) => (
            <DiagnosticCardComponent key={i} card={diag} />
          ))}
        </div>
      )}

      {/* Two-column layout: Heatmap + GeoMap */}
      <div className="grid grid-cols-1 xl:grid-cols-2 gap-5">
        {/* ISP Heatmap */}
        <GlassCard className="p-4" glow>
          <div className="flex items-center justify-between mb-4">
            <h3 className="text-sm font-semibold text-fg flex items-center gap-2">
              <span className="p-1 rounded-lg bg-primary/15 border border-primary/30">
                <BarChart3 size={14} className="text-primary" />
              </span>
              ISP Quality Heatmap
            </h3>
            <Select value={ispName} onChange={(e) => setIspName(e.target.value)} className="w-28">
              <option value="MCI">MCI</option>
              <option value="MTN">MTN</option>
              <option value="Rightel">Rightel</option>
              <option value="Mokhaberat">Mokhaberat</option>
            </Select>
          </div>
          <ISPHeatmapGrid cells={heatmap?.cells ?? []} />
        </GlassCard>

        {/* Geographic Node Map — lazy-loaded (Leaflet is heavy) */}
        <GlassCard className="p-4" glow>
          <h3 className="text-sm font-semibold text-fg flex items-center gap-2 mb-4">
            <span className="p-1 rounded-lg bg-primary/15 border border-primary/30">
              <MapPin size={14} className="text-primary" />
            </span>
            Node Locations
          </h3>
          <Suspense fallback={<div className="h-[300px] rounded-xl bg-surface-2/20 border border-border/40 flex items-center justify-center text-xs text-fg-subtle"><MapPin size={20} className="me-2 opacity-30" />Loading map…</div>}>
            <LazyGeoNodeMap nodes={nodes} />
          </Suspense>
        </GlassCard>
      </div>

      {/* Revenue + Subscription Analytics */}
      <div className="grid grid-cols-1 xl:grid-cols-2 gap-5">
        {/* Revenue Chart */}
        <GlassCard className="p-4" glow>
          <h3 className="text-sm font-semibold text-fg flex items-center gap-2 mb-4">
            <span className="p-1 rounded-lg bg-success/15 border border-success/30">
              <DollarSign size={14} className="text-success" />
            </span>
            Revenue
          </h3>
          {revenue ? <RevenueChart report={revenue} /> : <EmptyChart />}
        </GlassCard>

        {/* Subscription Analytics */}
        <GlassCard className="p-4" glow>
          <h3 className="text-sm font-semibold text-fg flex items-center gap-2 mb-4">
            <span className="p-1 rounded-lg bg-accent/15 border border-accent/30">
              <TrendingUp size={14} className="text-accent" />
            </span>
            Subscription Analytics
          </h3>
          {subAnalytics ? <SubAnalyticsCharts data={subAnalytics} /> : <EmptyChart />}
        </GlassCard>
      </div>
    </div>
  );
}

// ---------- Sub-Components ----------

function DiagnosticCardComponent({ card }: { card: DiagnosticCard }) {
  const severityStyles: Record<string, string> = {
    critical: "border-red-500/30 bg-red-500/5",
    warning: "border-yellow-500/30 bg-yellow-500/5",
    info: "border-blue-500/30 bg-blue-500/5",
  };
  const iconColor: Record<string, string> = {
    critical: "text-red-400",
    warning: "text-yellow-400",
    info: "text-blue-400",
  };

  return (
    <div className={cn("rounded-xl border p-4 backdrop-blur-sm", severityStyles[card.severity] ?? severityStyles.info)}>
      <div className="flex items-start gap-3">
        <div className={cn("p-1.5 rounded-lg bg-surface-2/60", iconColor[card.severity])}>
          <AlertTriangle size={16} className={cn("", iconColor[card.severity])} />
        </div>
        <div className="flex-1 min-w-0">
          <h4 className="text-sm font-semibold text-fg">{card.title}</h4>
          <p className="text-xs text-fg-muted mt-1">{card.description}</p>
          {card.actions.length > 0 && (
            <div className="flex gap-2 mt-3">
              {card.actions.map((action) => (
                <Button key={action} variant="outline" size="sm" className="text-[10px] h-6 px-2">
                  {action.replace("_", " ")}
                </Button>
              ))}
            </div>
          )}
        </div>
      </div>
    </div>
  );
}

const DAYS = ["Sun", "Mon", "Tue", "Wed", "Thu", "Fri", "Sat"];
const HOURS = Array.from({ length: 24 }, (_, i) => i);

function ISPHeatmapGrid({ cells }: { cells: HeatmapCell[] }) {
  const cellMap = new Map<string, number>();
  for (const c of cells) {
    cellMap.set(`${c.day}-${c.hour}`, c.score);
  }
  const maxScore = Math.max(...cells.map(c => c.score), 1);

  return (
    <div className="overflow-x-auto">
      <div className="min-w-[600px]">
        <div className="flex gap-px ml-10 mb-1">
          {HOURS.map((h) => (
            <div key={h} className="flex-1 text-center text-[8px] text-fg-subtle">
              {h % 3 === 0 ? `${h}h` : ""}
            </div>
          ))}
        </div>
        {DAYS.map((day, dayIdx) => (
          <div key={day} className="flex items-center gap-px mb-px">
            <span className="w-9 text-[9px] text-fg-muted text-right pr-1 flex-shrink-0">{day}</span>
            {HOURS.map((hour) => {
              const score = cellMap.get(`${dayIdx}-${hour}`) ?? 0;
              const intensity = maxScore > 0 ? score / maxScore : 0;
              return (
                <div
                  key={hour}
                  className="flex-1 aspect-square rounded-[2px] transition-all duration-200 hover:scale-125 hover:z-10"
                  style={{
                    backgroundColor: `hsl(var(--primary) / ${0.1 + intensity * 0.8})`,
                  }}
                  title={`${day} ${hour}:00 — Score: ${score.toFixed(2)}`}
                />
              );
            })}
          </div>
        ))}
        <div className="flex items-center gap-2 mt-3 ml-10 text-[9px] text-fg-subtle">
          <span>Low</span>
          <div className="flex gap-px">
            {[0.1, 0.3, 0.5, 0.7, 0.9].map((v) => (
              <div
                key={v}
                className="w-3 h-3 rounded-[2px]"
                style={{ backgroundColor: `hsl(var(--primary) / ${v})` }}
              />
            ))}
          </div>
          <span>High</span>
        </div>
      </div>
    </div>
  );
}

function RevenueChart({ report }: { report: RevenueReport }) {
  const series = report.time_series ?? [];

  return (
    <div className="space-y-3">
      <div className="grid grid-cols-3 gap-3">
        <div className="p-2 rounded-lg bg-success/8 border border-success/20 text-center">
          <p className="text-[9px] text-fg-subtle uppercase tracking-wide">Income</p>
          <p className="text-sm font-bold text-green-400 tabular-nums">{formatAmount(report.total_income)}</p>
        </div>
        <div className="p-2 rounded-lg bg-danger/8 border border-danger/20 text-center">
          <p className="text-[9px] text-fg-subtle uppercase tracking-wide">Expense</p>
          <p className="text-sm font-bold text-red-400 tabular-nums">{formatAmount(report.total_expense)}</p>
        </div>
        <div className="p-2 rounded-lg bg-primary/8 border border-primary/20 text-center">
          <p className="text-[9px] text-fg-subtle uppercase tracking-wide">Profit</p>
          <p className="text-sm font-bold text-primary tabular-nums">{formatAmount(report.profit)}</p>
        </div>
      </div>
      <RevenueAreaChart
        data={series.length > 0 ? series : [{ date: "-", income: 0, expense: 0 }]}
      />
      <div className="flex items-center gap-4 text-[9px] text-fg-subtle">
        <div className="flex items-center gap-1"><span className="h-1.5 w-3 rounded-full bg-green-400 shadow-[0_0_4px] shadow-green-400/50" /> Income</div>
        <div className="flex items-center gap-1"><span className="h-1.5 w-3 rounded-full bg-red-400 shadow-[0_0_4px] shadow-red-400/50" /> Expense</div>
      </div>
    </div>
  );
}

function SubAnalyticsCharts({ data }: { data: SubAnalyticsReport }) {
  const formatSlices: ProtocolSlice[] = (data.by_format ?? []).map((f, i) => ({
    label: f.format || "unknown",
    value: f.count,
    color: DONUT_COLORS[i % DONUT_COLORS.length],
  }));

  const ispSlices: ProtocolSlice[] = (data.by_isp ?? []).map((s, i) => ({
    label: s.isp || "unknown",
    value: s.count,
    color: DONUT_COLORS[(i + 3) % DONUT_COLORS.length],
  }));

  const totalFetches = formatSlices.reduce((acc, s) => acc + s.value, 0);

  return (
    <div className="grid grid-cols-2 gap-4">
      <div>
        <p className="text-[10px] text-fg-subtle uppercase tracking-wide mb-2 text-center">By Format</p>
        {formatSlices.length > 0 ? (
          <ProtocolDonutChart slices={formatSlices} centerLabel="Fetches" centerValue={totalFetches} />
        ) : (
          <div className="h-40 flex items-center justify-center text-xs text-fg-subtle">No data</div>
        )}
      </div>
      <div>
        <p className="text-[10px] text-fg-subtle uppercase tracking-wide mb-2 text-center">By ISP</p>
        {ispSlices.length > 0 ? (
          <ProtocolDonutChart slices={ispSlices} centerLabel="ISPs" centerValue={ispSlices.length} />
        ) : (
          <div className="h-40 flex items-center justify-center text-xs text-fg-subtle">No data</div>
        )}
      </div>
    </div>
  );
}

// ---------- Helpers ----------

const DONUT_COLORS = ["#22D3EE", "#3B82F6", "#10B981", "#8B5CF6", "#F59E0B", "#F43F5E", "#6366F1", "#EC4899"];

function EmptyChart() {
  return <div className="h-48 flex items-center justify-center text-xs text-fg-subtle">Loading...</div>;
}

function formatAmount(amount: number): string {
  if (amount >= 1_000_000) return `${(amount / 1_000_000).toFixed(1)}M`;
  if (amount >= 1_000) return `${(amount / 1_000).toFixed(1)}K`;
  return String(amount);
}
