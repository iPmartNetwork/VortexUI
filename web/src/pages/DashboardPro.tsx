import { useState } from "react";
import { useQuery } from "@tanstack/react-query";
import {
  Activity, AlertTriangle, Globe, TrendingUp, Zap,
  Server, Shield, RefreshCw, MapPin, DollarSign, BarChart3,
} from "lucide-react";
import { api } from "@/api/client";
import { Button, Select } from "@/components/ui";
import { GlassCard, StatsCard } from "@/components/veltrix";
import { ProtocolDonutChart, type ProtocolSlice } from "@/components/veltrix/ProtocolDonutChart";
import { useTitle } from "@/lib/useTitle";
import { cn } from "@/lib/utils";

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

interface GeoNode {
  node_id: string;
  name: string;
  lat: number;
  lng: number;
  status: string;
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
      <div className="flex flex-col lg:flex-row lg:items-center justify-between gap-4">
        <div>
          <h1 className="text-2xl font-bold text-fg tracking-tight">Dashboard Pro</h1>
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
      <GlassCard className="p-4">
        <h3 className="text-sm font-semibold text-fg mb-3 flex items-center gap-2">
          <Zap size={14} className="text-primary" /> Quick Actions
        </h3>
        <div className="flex flex-wrap gap-2">
          <Button variant="outline" size="sm"><RefreshCw size={12} /> Refresh Nodes</Button>
          <Button variant="outline" size="sm"><Server size={12} /> Restart Cores</Button>
          <Button variant="outline" size="sm"><Shield size={12} /> Check Certs</Button>
          <Button variant="outline" size="sm"><Globe size={12} /> Update Geo</Button>
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
        <GlassCard className="p-4">
          <div className="flex items-center justify-between mb-4">
            <h3 className="text-sm font-semibold text-fg flex items-center gap-2">
              <BarChart3 size={14} className="text-primary" /> ISP Quality Heatmap
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

        {/* Geographic Node Map */}
        <GlassCard className="p-4">
          <h3 className="text-sm font-semibold text-fg flex items-center gap-2 mb-4">
            <MapPin size={14} className="text-primary" /> Node Locations
          </h3>
          <GeoNodeMap nodes={nodes} />
        </GlassCard>
      </div>

      {/* Revenue + Subscription Analytics */}
      <div className="grid grid-cols-1 xl:grid-cols-2 gap-5">
        {/* Revenue Chart */}
        <GlassCard className="p-4">
          <h3 className="text-sm font-semibold text-fg flex items-center gap-2 mb-4">
            <DollarSign size={14} className="text-primary" /> Revenue
          </h3>
          {revenue ? <RevenueChart report={revenue} /> : <EmptyChart />}
        </GlassCard>

        {/* Subscription Analytics */}
        <GlassCard className="p-4">
          <h3 className="text-sm font-semibold text-fg flex items-center gap-2 mb-4">
            <TrendingUp size={14} className="text-primary" /> Subscription Analytics
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
    critical: "border-red-500/40 bg-red-500/5",
    warning: "border-yellow-500/40 bg-yellow-500/5",
    info: "border-blue-500/40 bg-blue-500/5",
  };
  const iconColor: Record<string, string> = {
    critical: "text-red-400",
    warning: "text-yellow-400",
    info: "text-blue-400",
  };

  return (
    <div className={cn("rounded-xl border p-4", severityStyles[card.severity] ?? severityStyles.info)}>
      <div className="flex items-start gap-3">
        <AlertTriangle size={16} className={cn("mt-0.5 flex-shrink-0", iconColor[card.severity])} />
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
  // Build a lookup map: `${day}-${hour}` -> score
  const cellMap = new Map<string, number>();
  for (const c of cells) {
    cellMap.set(`${c.day}-${c.hour}`, c.score);
  }
  const maxScore = Math.max(...cells.map(c => c.score), 1);

  return (
    <div className="overflow-x-auto">
      <div className="min-w-[600px]">
        {/* Hour labels */}
        <div className="flex gap-px ml-10 mb-1">
          {HOURS.map((h) => (
            <div key={h} className="flex-1 text-center text-[8px] text-fg-subtle">
              {h % 3 === 0 ? `${h}h` : ""}
            </div>
          ))}
        </div>
        {/* Grid rows */}
        {DAYS.map((day, dayIdx) => (
          <div key={day} className="flex items-center gap-px mb-px">
            <span className="w-9 text-[9px] text-fg-muted text-right pr-1 flex-shrink-0">{day}</span>
            {HOURS.map((hour) => {
              const score = cellMap.get(`${dayIdx}-${hour}`) ?? 0;
              const intensity = maxScore > 0 ? score / maxScore : 0;
              return (
                <div
                  key={hour}
                  className="flex-1 aspect-square rounded-[2px] transition-colors"
                  style={{
                    backgroundColor: `hsl(var(--primary) / ${0.1 + intensity * 0.8})`,
                  }}
                  title={`${day} ${hour}:00 — Score: ${score.toFixed(2)}`}
                />
              );
            })}
          </div>
        ))}
        {/* Legend */}
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

function GeoNodeMap({ nodes }: { nodes: GeoNode[] }) {
  if (nodes.length === 0) {
    return <div className="h-48 flex items-center justify-center text-xs text-fg-subtle">No node data available</div>;
  }

  return (
    <div className="relative rounded-xl bg-surface-2/30 p-4 overflow-hidden h-56">
      {/* Simplified world SVG */}
      <svg viewBox="0 0 1000 500" className="w-full h-auto opacity-20" preserveAspectRatio="xMidYMid">
        <ellipse cx="250" cy="200" rx="120" ry="100" fill="hsl(var(--fg-subtle))" opacity="0.3" />
        <ellipse cx="520" cy="180" rx="130" ry="120" fill="hsl(var(--fg-subtle))" opacity="0.3" />
        <ellipse cx="750" cy="230" rx="100" ry="90" fill="hsl(var(--fg-subtle))" opacity="0.3" />
        <ellipse cx="200" cy="350" rx="60" ry="80" fill="hsl(var(--fg-subtle))" opacity="0.3" />
        <ellipse cx="530" cy="380" rx="50" ry="50" fill="hsl(var(--fg-subtle))" opacity="0.3" />
        <ellipse cx="830" cy="380" rx="70" ry="50" fill="hsl(var(--fg-subtle))" opacity="0.3" />
      </svg>

      {/* Node markers */}
      <svg viewBox="0 0 1000 500" className="absolute inset-0 w-full h-full p-4" preserveAspectRatio="xMidYMid">
        {nodes.map((node, i) => {
          // Distribute nodes evenly if no real coordinates
          const x = node.lng !== 0 ? ((node.lng + 180) / 360) * 1000 : 200 + (i * 120) % 700;
          const y = node.lat !== 0 ? ((90 - node.lat) / 180) * 500 : 150 + (i * 80) % 300;
          const isOnline = node.status === "online";
          return (
            <g key={node.node_id}>
              <circle
                cx={x} cy={y} r={12}
                fill={isOnline ? "hsl(var(--primary))" : "hsl(var(--destructive))"}
                opacity={0.2}
              />
              <circle
                cx={x} cy={y} r={6}
                fill={isOnline ? "hsl(var(--primary))" : "hsl(var(--destructive))"}
                opacity={0.8}
                className={isOnline ? "animate-pulse" : ""}
              />
              <text x={x} y={y + 20} textAnchor="middle" fontSize="10" fill="hsl(var(--fg-muted))">
                {node.name}
              </text>
            </g>
          );
        })}
      </svg>

      {/* Status legend */}
      <div className="absolute bottom-3 start-3 flex items-center gap-3 text-[10px] text-fg-subtle">
        <div className="flex items-center gap-1"><span className="h-2 w-2 rounded-full bg-primary" /> Online</div>
        <div className="flex items-center gap-1"><span className="h-2 w-2 rounded-full bg-destructive" /> Offline</div>
        <span className="ml-2">{nodes.length} nodes</span>
      </div>
    </div>
  );
}

function RevenueChart({ report }: { report: RevenueReport }) {
  const series = report.time_series ?? [];

  if (series.length === 0) {
    return (
      <div className="space-y-3">
        <div className="grid grid-cols-3 gap-3">
          <MiniStat label="Income" value={formatAmount(report.total_income)} className="text-green-400" />
          <MiniStat label="Expense" value={formatAmount(report.total_expense)} className="text-red-400" />
          <MiniStat label="Profit" value={formatAmount(report.profit)} className="text-primary" />
        </div>
        <div className="h-32 flex items-center justify-center text-xs text-fg-subtle">No time series data</div>
      </div>
    );
  }

  const maxVal = Math.max(...series.flatMap(s => [s.income, s.expense]), 1);
  const w = 400;
  const h = 120;
  const step = w / Math.max(series.length - 1, 1);

  const incomePath = series.map((p, i) => `${i === 0 ? "M" : "L"}${i * step},${h - (p.income / maxVal) * h}`).join(" ");
  const expensePath = series.map((p, i) => `${i === 0 ? "M" : "L"}${i * step},${h - (p.expense / maxVal) * h}`).join(" ");

  return (
    <div className="space-y-3">
      <div className="grid grid-cols-3 gap-3">
        <MiniStat label="Income" value={formatAmount(report.total_income)} className="text-green-400" />
        <MiniStat label="Expense" value={formatAmount(report.total_expense)} className="text-red-400" />
        <MiniStat label="Profit" value={formatAmount(report.profit)} className="text-primary" />
      </div>
      <div className="overflow-hidden">
        <svg viewBox={`0 0 ${w} ${h}`} className="w-full h-32" preserveAspectRatio="none">
          <path d={incomePath} fill="none" stroke="hsl(142 71% 45%)" strokeWidth="2" />
          <path d={expensePath} fill="none" stroke="hsl(0 84% 60%)" strokeWidth="2" />
        </svg>
      </div>
      <div className="flex items-center gap-4 text-[9px] text-fg-subtle">
        <div className="flex items-center gap-1"><span className="h-1.5 w-3 rounded-full bg-green-400" /> Income</div>
        <div className="flex items-center gap-1"><span className="h-1.5 w-3 rounded-full bg-red-400" /> Expense</div>
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

function MiniStat({ label, value, className }: { label: string; value: string; className?: string }) {
  return (
    <div className="text-center">
      <p className="text-[9px] text-fg-subtle uppercase tracking-wide">{label}</p>
      <p className={cn("text-sm font-bold tabular-nums", className ?? "text-fg")}>{value}</p>
    </div>
  );
}

function EmptyChart() {
  return <div className="h-48 flex items-center justify-center text-xs text-fg-subtle">Loading...</div>;
}

function formatAmount(amount: number): string {
  if (amount >= 1_000_000) return `${(amount / 1_000_000).toFixed(1)}M`;
  if (amount >= 1_000) return `${(amount / 1_000).toFixed(1)}K`;
  return String(amount);
}
