import { useState, useEffect, useRef, useId, useCallback, type ReactNode } from "react";
import {
  AreaChart,
  Area,
  BarChart,
  Bar,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  ResponsiveContainer,
  PieChart,
  Pie,
  Cell,
  Sector,
  Brush,
  ReferenceLine,
} from "recharts";
import { Download, Expand, Minimize2, FileDown, Image, Table2, Loader2 } from "lucide-react";
import { cn, formatBytes } from "@/lib/utils";
import { useAnimationControl } from "@/components/ChartAnimationController";
import type { LiveTrafficPoint } from "@/hooks/useLiveTraffic";

// ════════════════════════════════════════════════════════════════
// Chart Export Utilities (PNG + CSV) — no external deps needed
// ════════════════════════════════════════════════════════════════

/** Serialize a Recharts SVG element to a downloadable PNG blob */
export async function exportChartToPNG(
  containerRef: React.RefObject<HTMLDivElement | null>,
  filename = "chart",
): Promise<void> {
  const container = containerRef.current;
  if (!container) return;

  const svgEl = container.querySelector("svg.recharts-surface");
  if (!svgEl) return;

  // Clone the SVG so we don't mutate the live DOM
  const clone = svgEl.cloneNode(true) as SVGElement;

  // Remove foreignObjects (tooltips) — they break Image rendering
  clone.querySelectorAll("foreignObject").forEach((fo) => fo.remove());

  // Remove any transparent/empty groups that might cause rendering artifacts
  clone.querySelectorAll("[opacity='0'], .recharts-tooltip-wrapper").forEach((el) => el.remove());

  // Extract inline styles from computed styles for better fidelity
  const styleEl = document.createElement("style");
  const computed = getComputedStyle(svgEl);
  styleEl.textContent = `
    .recharts-surface { background: transparent; }
    .recharts-cartesian-grid line { stroke: rgba(255,255,255,0.08); }
    .recharts-text { font-family: ${computed.fontFamily || "Inter, sans-serif"}; }
    text { fill: rgba(255,255,255,0.7); font-size: 10px; }
  `;
  clone.prepend(styleEl);

  // Serialize to SVG data URI
  const serializer = new XMLSerializer();
  const svgStr = serializer.serializeToString(clone);
  const svgDataUri = `data:image/svg+xml;base64,${btoa(unescape(encodeURIComponent(svgStr)))}`;

  // Render to canvas at 2x for retina quality
  const img = new window.Image();
  const canvas = document.createElement("canvas");
  const ctx = canvas.getContext("2d");

  return new Promise<void>((resolve, reject) => {
    img.onload = () => {
      const W = img.naturalWidth;
      const H = img.naturalHeight;
      canvas.width = W * 2;
      canvas.height = H * 2;

      // Dark background for the export
      if (ctx) {
        ctx.fillStyle = "#0A0E1A";
        ctx.fillRect(0, 0, canvas.width, canvas.height);
        ctx.scale(2, 2);
        ctx.drawImage(img, 0, 0, W, H);
      }

      canvas.toBlob((blob) => {
        if (!blob) {
          reject(new Error("Canvas toBlob failed"));
          return;
        }
        const url = URL.createObjectURL(blob);
        const a = document.createElement("a");
        a.href = url;
        a.download = `${filename}.png`;
        document.body.appendChild(a);
        a.click();
        document.body.removeChild(a);
        URL.revokeObjectURL(url);
        resolve();
      }, "image/png");
    };
    img.onerror = () => reject(new Error("Failed to load SVG for PNG export"));
    img.src = svgDataUri;
  });
}

/** Export chart data as CSV and trigger download */
export function exportChartToCSV<T extends Record<string, unknown>>(
  data: T[],
  filename = "chart-data",
  columns?: (keyof T)[],
): void {
  if (!data.length) return;

  const keys = (columns ?? Object.keys(data[0])) as string[];
  const header = keys.join(",");
  const rows = data.map((row) =>
    keys
      .map((k) => {
        const v = row[k];
        if (v == null) return "";
        const str = String(v);
        // Escape commas and quotes for CSV
        return str.includes(",") || str.includes('"') || str.includes("\n")
          ? `"${str.replace(/"/g, '""')}"`
          : str;
      })
      .join(","),
  );

  const csv = [header, ...rows].join("\n");
  const blob = new Blob([csv], { type: "text/csv;charset=utf-8;" });
  const url = URL.createObjectURL(blob);
  const a = document.createElement("a");
  a.href = url;
  a.download = `${filename}.csv`;
  document.body.appendChild(a);
  a.click();
  document.body.removeChild(a);
  URL.revokeObjectURL(url);
}

// ════════════════════════════════════════════════════════════════
// ChartToolbar — shared header with export + fullscreen + animation
// ════════════════════════════════════════════════════════════════

export interface ChartToolbarProps {
  title: string;
  subtitle?: string;
  /** Data for CSV export — pass the raw data array */
  csvData?: Record<string, unknown>[];
  csvFilename?: string;
  /** Ref that wraps the chart content — used for PNG export */
  chartRef?: React.RefObject<HTMLDivElement | null>;
  /** Called when user clicks Export PNG — takes precedence over chartRef */
  onExportPNG?: () => void;
  /** Extra controls rendered on the right side */
  children?: ReactNode;
  /** Render function for fullscreen mode */
  renderFullscreenContent?: () => ReactNode;
}

function useExportPNGHandler(chartRef?: React.RefObject<HTMLDivElement | null>, onExportPNG?: () => void) {
  const busyRef = useRef(false);
  const handleExport = useCallback(async () => {
    if (busyRef.current) return;
    busyRef.current = true;
    try {
      if (onExportPNG) {
        onExportPNG();
      } else if (chartRef) {
        await exportChartToPNG(chartRef);
      }
    } catch { /* handled by exportChartToPNG */ }
    busyRef.current = false;
  }, [chartRef, onExportPNG]);
  return handleExport;
}

/**
 * A glass-styled toolbar that sits above every chart.
 * Provides: title/subtitle, PNG export, CSV export, fullscreen toggle.
 * Also shows active animation status from ChartAnimationController.
 */
export function ChartToolbar({
  title,
  subtitle,
  csvData,
  csvFilename,
  chartRef,
  onExportPNG,
  renderFullscreenContent,
  children,
}: ChartToolbarProps) {
  const [exportOpen, setExportOpen] = useState(false);
  const [fullscreen, setFullscreen] = useState(false);
  const [exporting, setExporting] = useState<"png" | "csv" | null>(null);
  const exportRef = useRef<HTMLDivElement>(null);
  const { paused, speed } = useAnimationControl();

  // Close export dropdown on outside click
  useEffect(() => {
    if (!exportOpen) return;
    const handler = (e: MouseEvent) => {
      if (exportRef.current && !exportRef.current.contains(e.target as Node)) {
        setExportOpen(false);
      }
    };
    document.addEventListener("mousedown", handler);
    return () => document.removeEventListener("mousedown", handler);
  }, [exportOpen]);

  // Escape key exits fullscreen
  useEffect(() => {
    if (!fullscreen) return;
    const handler = (e: KeyboardEvent) => {
      if (e.key === "Escape") setFullscreen(false);
    };
    window.addEventListener("keydown", handler);
    return () => window.removeEventListener("keydown", handler);
  }, [fullscreen]);

  const handleExportPNG = useExportPNGHandler(chartRef, onExportPNG);

  return (
    <div
      className={cn(
        "flex items-center justify-between gap-3 px-1 pb-1.5 select-none",
        fullscreen && "px-4 pt-3",
      )}
    >
      {/* Title area */}
      <div className="min-w-0 flex-1">
        <h3 className="text-xs font-semibold text-fg truncate flex items-center gap-2">
          {title}
          {paused && (
            <span className="inline-flex items-center gap-1 px-1.5 py-0.5 rounded bg-amber-400/10 text-[8px] font-bold text-amber-400 uppercase tracking-wider">
              <span className="h-1 w-1 rounded-full bg-amber-400" />
              Paused
            </span>
          )}
          {speed !== 1 && !paused && (
            <span className="text-[8px] font-mono text-fg-subtle/50">{speed}x</span>
          )}
        </h3>
        {subtitle && (
          <p className="text-[9px] text-fg-subtle/60 truncate mt-0.5">{subtitle}</p>
        )}
      </div>

      {/* Controls */}
      <div className="flex items-center gap-1">
        {children}

        {/* Export dropdown */}
        <div ref={exportRef} className="relative">
          <button
            type="button"
            onClick={() => setExportOpen(!exportOpen)}
            disabled={exporting !== null}
            className="h-6 w-6 flex items-center justify-center rounded-md border border-border/50 bg-surface-2/30 hover:bg-surface-2/70 text-fg-muted hover:text-fg transition disabled:opacity-30"
            title="Export chart"
          >
            {exporting ? <Loader2 size={11} className="animate-spin" /> : <Download size={11} />}
          </button>

          {exportOpen && (
            <div className="absolute top-full mt-1 end-0 z-30 min-w-[150px] rounded-lg border border-border/60 bg-surface shadow-lg py-1 animate-scale-in">
              <p className="px-3 py-1 text-[8px] font-bold uppercase tracking-wider text-fg-subtle/50">
                Export
              </p>
              <button
                type="button"                  onClick={() => {
                    setExporting("png");
                    setExportOpen(false);
                    handleExportPNG();
                    setTimeout(() => setExporting(null), 800);
                  }}
                className="w-full flex items-center gap-2 px-3 py-1.5 text-xs text-fg-muted hover:text-fg hover:bg-surface-2/60 transition"
              >
                <Image size={12} />
                Export PNG
              </button>
              {csvData && (
                <button
                  type="button"
                  onClick={() => {
                    setExporting("csv");
                    exportChartToCSV(csvData, csvFilename || title.toLowerCase().replace(/\s+/g, "-"));
                    setExportOpen(false);
                    setTimeout(() => setExporting(null), 500);
                  }}
                  className="w-full flex items-center gap-2 px-3 py-1.5 text-xs text-fg-muted hover:text-fg hover:bg-surface-2/60 transition"
                >
                  <Table2 size={12} />
                  Export CSV
                </button>
              )}
            </div>
          )}
        </div>

        {/* Fullscreen toggle */}
        <button
          type="button"
          onClick={() => setFullscreen(!fullscreen)}
          className="h-6 w-6 flex items-center justify-center rounded-md border border-border/50 bg-surface-2/30 hover:bg-surface-2/70 text-fg-muted hover:text-fg transition"
          title={fullscreen ? "Exit fullscreen" : "Fullscreen"}
        >
          {fullscreen ? <Minimize2 size={11} /> : <Expand size={11} />}
        </button>
      </div>

      {/* Fullscreen overlay */}
      {fullscreen && (
        <div
          className="fixed inset-0 z-50 flex items-center justify-center bg-black/70 backdrop-blur-xl animate-fade-in"
          onClick={() => setFullscreen(false)}
        >
          <div
            className="relative w-[90vw] max-w-6xl max-h-[85vh] glass rounded-2xl p-6 flex flex-col overflow-hidden"
            onClick={(e) => e.stopPropagation()}
          >
            <div className="flex items-center justify-between mb-4 flex-shrink-0">
              <h2 className="text-sm font-bold text-fg">{title}</h2>
              <div className="flex items-center gap-2">
                <button
                  type="button"
                  onClick={async () => {
                    try {
                      if (onExportPNG) onExportPNG();
                      else if (chartRef) await exportChartToPNG(chartRef);
                    } catch { /* silent */ }
                  }}
                  className="h-7 px-2.5 flex items-center gap-1.5 rounded-lg border border-border/60 text-[10px] font-medium text-fg-muted hover:text-fg hover:bg-surface-2/60 transition"
                >
                  <Image size={12} />
                  PNG
                </button>
                {csvData && (
                  <button
                    type="button"
                    onClick={() => {
                      exportChartToCSV(csvData, csvFilename || title.toLowerCase().replace(/\s+/g, "-"));
                    }}
                    className="h-7 px-2.5 flex items-center gap-1.5 rounded-lg border border-border/60 text-[10px] font-medium text-fg-muted hover:text-fg hover:bg-surface-2/60 transition"
                  >
                    <Table2 size={12} />
                    CSV
                  </button>
                )}
                <button
                  type="button"
                  onClick={() => setFullscreen(false)}
                  className="h-7 w-7 flex items-center justify-center rounded-lg border border-border/60 text-fg-muted hover:text-fg hover:bg-surface-2/60 transition"
                >
                  <Minimize2 size={13} />
                </button>
              </div>
            </div>
            <div className="flex-1 w-full min-h-0">
              {renderFullscreenContent ? renderFullscreenContent() : null}
            </div>
          </div>
        </div>
      )}
    </div>
  );
}

// ════════════════════════════════════════════════════════════════
// ChartExportButton — standalone export button that triggers PNG via ref
// ════════════════════════════════════════════════════════════════
interface ChartExportButtonProps {
  chartRef: React.RefObject<HTMLDivElement | null>;
  filename?: string;
  onExportStart?: () => void;
  onExportEnd?: () => void;
}

export function ChartExportButton({ chartRef, filename = "chart", onExportStart, onExportEnd }: ChartExportButtonProps) {
  const [busy, setBusy] = useState(false);

  const handleExport = useCallback(async () => {
    setBusy(true);
    onExportStart?.();
    try {
      await exportChartToPNG(chartRef, filename);
    } catch (err) {
      console.warn("Chart PNG export failed:", err);
    } finally {
      setBusy(false);
      onExportEnd?.();
    }
  }, [chartRef, filename, onExportStart, onExportEnd]);

  return (
    <button
      type="button"
      onClick={handleExport}
      disabled={busy}
      className="h-6 w-6 flex items-center justify-center rounded-md border border-border/50 bg-surface-2/30 hover:bg-surface-2/70 text-fg-muted hover:text-fg transition disabled:opacity-30"
      title="Download PNG"
    >
      {busy ? <Loader2 size={11} className="animate-spin" /> : <FileDown size={11} />}
    </button>
  );
}

// ════════════════════════════════════════════════════════════════
// useAnimationProps — connects Recharts animation to AnimationController
// ════════════════════════════════════════════════════════════════
interface AnimationProps {
  isAnimationActive?: boolean;
  animationDuration?: number;
  animationEasing?: "ease" | "ease-in" | "ease-out" | "ease-in-out" | "linear" | undefined;
  /** Reset key to force re-animation */
  key?: string | number;
}

export function useChartAnimation(baseDuration = 600): AnimationProps {
  const { paused, speed, resetKey } = useAnimationControl();

  return {
    isAnimationActive: !paused,
    animationDuration: Math.round(baseDuration / speed),
    animationEasing: "ease-out",
    key: resetKey,
  };
}

// ════════════════════════════════════════════════════════════════
// Chart Loading Skeleton
// ════════════════════════════════════════════════════════════════
export function ChartSkeleton({ className, type = "area" }: { className?: string; type?: "area" | "bar" | "donut" }) {
  return (
    <div className={cn("rounded-2xl bg-bg-elevated border border-border p-5 space-y-3", className)} aria-hidden="true">
      {/* Toolbar skeleton */}
      <div className="flex items-center justify-between">
        <div className="h-3 w-28 rounded-md shimmer" />
        <div className="h-5 w-12 rounded-md shimmer" />
      </div>
      {/* Chart body skeleton */}
      {type === "donut" ? (
        <div className="flex flex-col items-center gap-4 py-4">
          <div className="relative h-32 w-32">
            <div className="absolute inset-0 rounded-full shimmer" />
            <div className="absolute inset-4 rounded-full bg-bg-elevated" />
          </div>
          <div className="flex flex-wrap justify-center gap-2">
            {Array.from({ length: 3 }).map((_, i) => (
              <div key={i} className="flex items-center gap-1.5">
                <div className="h-2 w-2 rounded-full shimmer" />
                <div className="h-2.5 w-14 rounded shimmer" />
              </div>
            ))}
          </div>
        </div>
      ) : (
        <div className="space-y-2 py-3">
          {/* Grid lines */}
          {Array.from({ length: 4 }).map((_, i) => (
            <div key={i} className="h-px w-full bg-border/10" />
          ))}
          {/* Area fill skeleton */}
          <div className="relative h-28 w-full overflow-hidden rounded-xl">
            <div
              className="absolute inset-0 shimmer"
              style={{
                maskImage: "linear-gradient(to bottom, rgba(0,0,0,0.4) 0%, transparent 90%)",
                WebkitMaskImage: "linear-gradient(to bottom, rgba(0,0,0,0.4) 0%, transparent 90%)",
              }}
            />
          </div>
        </div>
      )}
    </div>
  );
}

// ════════════════════════════════════════════════════════════════
// Types
// ════════════════════════════════════════════════════════════════
export interface TrafficPoint {
  time: string;
  up: number;
  down: number;
}

export interface UsagePoint {
  time: string;
  up: number;
  down: number;
}

export interface RevenuePoint {
  date: string;
  income: number;
  expense: number;
}

export interface PieSlice {
  label: string;
  value: number;
  color: string;
}

// ════════════════════════════════════════════════════════════════
// Helpers
// ════════════════════════════════════════════════════════════════

/** Format bytes for Recharts Y-axis — removes decimal for cleaner labels */
function axisBytes(v: number) {
  return formatBytes(v, false).replace(/\.\d+/, "");
}

/** Unique gradient/brush IDs per chart instance */
function useSafeId(prefix: string) {
  const uid = useId();
  return `${prefix}-${uid.replace(/[^a-zA-Z0-9_-]/g, "")}`;
}

/** Create a unique drop-shadow filter ID per chart instance */
function useShadowId() {
  return useSafeId("shadow");
}

// ════════════════════════════════════════════════════════════════
// Shared Chart components
// ════════════════════════════════════════════════════════════════

/** Default margin for area/bar charts */
const CHART_MARGIN = { top: 6, right: 4, left: -16, bottom: 0 };

/** Common grid style */
function ChartGrid() {
  return <CartesianGrid strokeDasharray="3 3" stroke="hsl(var(--border) / 0.22)" vertical={false} />;
}

/** Gradient defs generator — creates `downGradId` and `upGradId` */
function TrafficGradients({ downId, upId }: { downId: string; upId: string }) {
  return (
    <defs>
      <linearGradient id={downId} x1="0" y1="0" x2="0" y2="1">
        <stop offset="0%" stopColor="#0EA5E9" stopOpacity={0.55} />
        <stop offset="100%" stopColor="#0EA5E9" stopOpacity={0.03} />
      </linearGradient>
      <linearGradient id={upId} x1="0" y1="0" x2="0" y2="1">
        <stop offset="0%" stopColor="#6366F1" stopOpacity={0.55} />
        <stop offset="100%" stopColor="#6366F1" stopOpacity={0.03} />
      </linearGradient>
    </defs>
  );
}

/** Shared drop-shadow SVG filter — use via useShadowId() for a unique ID per instance */
export function ChartEffects({ id }: { id: string }) {
  return (
    <defs>
      <filter id={id} x="-20%" y="-20%" width="140%" height="140%">
        <feDropShadow dx={0} dy={2} stdDeviation={4} floodColor="rgba(0,0,0,0.25)" floodOpacity={0.3} />
      </filter>
    </defs>
  );
}

// ════════════════════════════════════════════════════════════════
// Interactive Legend (clickable to toggle series)
// ════════════════════════════════════════════════════════════════
interface LegendEntry {
  dataKey: string;
  name: string;
  color: string;
  icon?: "line" | "rect" | "circle" | "square";
}

function InteractiveLegend({
  entries,
  hidden,
  onToggle,
}: {
  entries: LegendEntry[];
  hidden: Set<string>;
  onToggle: (key: string) => void;
}) {
  return (
    <div className="flex flex-wrap items-center gap-3 px-1 py-1.5">
      {entries.map((e) => {
        const isHidden = hidden.has(e.dataKey);
        return (
          <button
            key={e.dataKey}
            type="button"
            onClick={() => onToggle(e.dataKey)}
            className={cn(
              "flex items-center gap-1.5 text-[10px] font-medium transition-all duration-200 hover:opacity-100",
              isHidden ? "opacity-30 line-through" : "opacity-80",
            )}
          >
            {e.icon === "rect" || e.icon === "square" ? (
              <span
                className="inline-block h-2.5 w-2.5 rounded-sm flex-shrink-0"
                style={{ background: e.color }}
              />
            ) : (
              <span
                className="inline-block h-2 w-2 rounded-full flex-shrink-0"
                style={{ background: e.color, boxShadow: `0 0 4px ${e.color}60` }}
              />
            )}
            {e.name}
          </button>
        );
      })}
    </div>
  );
}

// ════════════════════════════════════════════════════════════════
// Custom Tooltip (shared)
// ════════════════════════════════════════════════════════════════
function ChartTooltip({ active, payload, label, formatter }: any) {
  if (!active || !payload?.length) return null;
  const fmt = formatter ?? ((v: number) => formatBytes(v, false));
  return (
    <div className="glass rounded-xl px-3.5 py-2.5 text-xs shadow-xl border border-border/60 min-w-[150px] backdrop-blur-lg">
      <p className="font-semibold text-fg mb-1.5 pb-1.5 border-b border-border/30">{label}</p>
      {payload.map((entry: any, i: number) => (
        <div key={i} className="flex items-center justify-between gap-4 py-0.5">
          <span className="flex items-center gap-1.5">
            {entry.payload.fill || entry.color ? (
              <span
                className="inline-block h-2 w-2 rounded-full flex-shrink-0"
                style={{ background: entry.payload.fill || entry.color }}
              />
            ) : null}
            <span className="text-fg-muted">{entry.name}</span>
          </span>
          <span className="font-bold text-fg tabular-nums">{fmt(entry.value)}</span>
        </div>
      ))}
    </div>
  );
}

// ════════════════════════════════════════════════════════════════
// Custom Brush component (styled) — use INSIDE the chart component
// ════════════════════════════════════════════════════════════════
function ChartBrush({ data, dataKey, ...props }: any) {
  if (!data || data.length < 10) return null;
  return (
    <Brush
      data={data}
      dataKey={dataKey}
      height={28}
      stroke="hsl(var(--border) / 0.4)"
      fill="hsl(var(--surface) / 0.3)"
      travellerWidth={8}
      gap={1}
      padding={{ top: 2, bottom: 2 }}
      {...props}
    >
      <AreaChart>
        <defs>
          <linearGradient id="brushGrad" x1="0" y1="1" x2="0" y2="0">
            <stop offset="0%" stopColor="hsl(var(--primary))" stopOpacity={0.5} />
            <stop offset="100%" stopColor="hsl(var(--primary))" stopOpacity={0.02} />
          </linearGradient>
        </defs>
        <Area
          type="monotone"
          dataKey={dataKey ?? "down"}
          stroke="hsl(var(--primary))"
          strokeWidth={1}
          fill="url(#brushGrad)"
          dot={false}
        />
      </AreaChart>
    </Brush>
  );
}

// ════════════════════════════════════════════════════════════════
// Speed badge (existing, unchanged)
// ════════════════════════════════════════════════════════════════
function SpeedBadge({ label, bps, color }: { label: string; bps: number; color: string }) {
  const [animVal, setAnimVal] = useState(0);
  const rafRef = useRef<number>(0);
  const lastTimeRef = useRef<number>(0);

  useEffect(() => {
    let running = true;
    const SPEED = 4;
    lastTimeRef.current = performance.now();

    const tick = (now: number) => {
      if (!running) return;
      const dt = Math.min((now - lastTimeRef.current) / 1000, 0.1);
      lastTimeRef.current = now;

      setAnimVal((prev) => {
        const diff = bps - prev;
        const step = diff * SPEED * dt;
        if (Math.abs(diff) < 1) return bps;
        return prev + step;
      });
      rafRef.current = requestAnimationFrame(tick);
    };
    rafRef.current = requestAnimationFrame(tick);
    return () => {
      running = false;
      cancelAnimationFrame(rafRef.current);
    };
  }, [bps]);

  const display = animVal < 0.5 ? "—"
    : animVal >= 1_000_000 ? `${(animVal / 1_000_000).toFixed(2)} MB/s`
    : animVal >= 1_000 ? `${(animVal / 1_000).toFixed(1)} KB/s`
    : `${animVal.toFixed(1)} B/s`;

  const isActive = animVal >= 0.5;

  return (
    <span className="flex items-center gap-1.5">
      <span
        className="inline-block h-2 w-2 rounded-full flex-shrink-0 transition-opacity duration-300"
        style={{
          background: color,
          boxShadow: isActive ? `0 0 6px 1px ${color}80` : "none",
          opacity: isActive ? 1 : 0.3,
        }}
      />
      <span className="text-fg-muted text-[10px]">{label}</span>
      <span
        className="font-bold tabular-nums text-[11px] transition-colors duration-300"
        style={{ color: isActive ? color : "hsl(var(--fg-muted))" }}
      >
        {display}
      </span>
    </span>
  );
}

// ════════════════════════════════════════════════════════════════
// UseLegendToggle — shared hook for interactive legend
// ════════════════════════════════════════════════════════════════
function useLegendToggle(_keys?: string[]) {
  const [hidden, setHidden] = useState<Set<string>>(new Set());
  const toggle = useCallback((key: string) => {
    setHidden((prev) => {
      const next = new Set(prev);
      if (next.has(key)) next.delete(key);
      else next.add(key);
      return next;
    });
  }, []);
  return { hidden, toggle };
}

// ════════════════════════════════════════════════════════════════
// Live traffic area chart — real-time with Brush + Legend
// ════════════════════════════════════════════════════════════════
export function LiveTrafficAreaChart({
  livePoints,
  downSpeed,
  upSpeed,
}: {
  livePoints: LiveTrafficPoint[];
  downSpeed: number;
  upSpeed: number;
}) {
  const downGradId = useSafeId("lD");
  const upGradId = useSafeId("lU");
  const shadowId = useShadowId();
  const { hidden, toggle } = useLegendToggle(["down", "up"]);

  const data = livePoints.map((p) => ({
    down: p.down,
    up: p.up,
    label: new Date(p.time).toLocaleTimeString([], {
      hour: "2-digit",
      minute: "2-digit",
      second: livePoints.length < 30 ? "2-digit" : undefined,
    }),
  }));

  if (!data.length) {
    return <div className="flex h-44 items-center justify-center text-xs text-fg-subtle">No traffic data yet</div>;
  }

  const totalBw = data.reduce((s, p) => s + p.up + p.down, 0);
  const avgDown = data.reduce((s, p) => s + p.down, 0) / data.length;
  const avgUp = data.reduce((s, p) => s + p.up, 0) / data.length;

  const legendEntries: LegendEntry[] = [
    { dataKey: "down", name: "Download", color: "#0EA5E9" },
    { dataKey: "up", name: "Upload", color: "#6366F1" },
  ];

  return (
    <div className="space-y-2">
      <div className="flex items-center gap-4 px-1">
        <SpeedBadge label="Down" bps={downSpeed} color="#0EA5E9" />
        <SpeedBadge label="Up" bps={upSpeed} color="#6366F1" />
      </div>

      <InteractiveLegend entries={legendEntries} hidden={hidden} onToggle={toggle} />

      <div className="h-44 w-full">
        <ResponsiveContainer width="100%" height="100%">
          <AreaChart data={data} margin={CHART_MARGIN}>
            <ChartEffects id={shadowId} />
            <TrafficGradients downId={downGradId} upId={upGradId} />
            <ChartGrid />
            <XAxis
              dataKey="label"
              tick={{ fontSize: 8, fill: "hsl(var(--fg-subtle) / 0.45)" }}
              tickLine={false}
              axisLine={false}
              interval="preserveStartEnd"
              minTickGap={40}
            />
            <YAxis
              tick={{ fontSize: 8, fill: "hsl(var(--fg-subtle) / 0.45)" }}
              tickLine={false}
              axisLine={false}
              tickFormatter={axisBytes}
              width={36}
            />
            <Tooltip content={<ChartTooltip />} cursor={{ stroke: "hsl(var(--fg) / 0.12)", strokeDasharray: "3 3" }} />

            {!hidden.has("down") && (
              <ReferenceLine
                y={avgDown}
                stroke="#0EA5E9"
                strokeDasharray="4 4"
                strokeOpacity={0.35}
                label={{
                  value: `avg ${formatBytes(avgDown, false)}`,
                  position: "insideTopRight",
                  fontSize: 8,
                  fill: "#0EA5E9",
                  opacity: 0.6,
                }}
              />
            )}
            {!hidden.has("up") && (
              <ReferenceLine
                y={avgUp}
                stroke="#6366F1"
                strokeDasharray="4 4"
                strokeOpacity={0.35}
                label={{
                  value: `avg ${formatBytes(avgUp, false)}`,
                  position: "insideBottomRight",
                  fontSize: 8,
                  fill: "#6366F1",
                  opacity: 0.6,
                }}
              />
            )}

            {!hidden.has("down") && (
              <Area
                type="monotone"
                dataKey="down"
                name="Download"
                stroke="#0EA5E9"
                strokeWidth={2}
                fill={`url(#${downGradId})`}
                dot={false}
                isAnimationActive={true}
                animationDuration={600}
                animationEasing="ease-out"
                activeDot={{ r: 4, fill: "#0EA5E9", stroke: "hsl(var(--bg-elevated))", strokeWidth: 2, filter: `url(#${shadowId})` }}
              />
            )}
            {!hidden.has("up") && (
              <Area
                type="monotone"
                dataKey="up"
                name="Upload"
                stroke="#6366F1"
                strokeWidth={2}
                fill={`url(#${upGradId})`}
                dot={false}
                isAnimationActive={true}
                animationDuration={600}
                animationEasing="ease-out"
                activeDot={{ r: 4, fill: "#6366F1", stroke: "hsl(var(--bg-elevated))", strokeWidth: 2, filter: `url(#${shadowId})` }}
              />
            )}

            <ChartBrush data={data} dataKey="down" />
          </AreaChart>
        </ResponsiveContainer>
      </div>

      <div className="flex items-center justify-between text-[10px] text-fg-muted">
        <span className="flex items-center gap-3">
          <span className="flex items-center gap-1.5">
            <span className="inline-block h-2 w-2 rounded-full bg-[#0EA5E9]" />
            <span>Download</span>
          </span>
          <span className="flex items-center gap-1.5">
            <span className="inline-block h-2 w-2 rounded-full bg-[#6366F1]" />
            <span>Upload</span>
          </span>
        </span>
        <span className="text-fg-subtle tabular-nums">total {formatBytes(totalBw, false)}</span>
      </div>
    </div>
  );
}

// ════════════════════════════════════════════════════════════════
// Traffic Area Chart (static/historical) — with Brush + Legend + ReferenceLine
// ════════════════════════════════════════════════════════════════
export function TrafficAreaChart({ points }: { points: TrafficPoint[] }) {
  const downGradId = useSafeId("tD");
  const upGradId = useSafeId("tU");
  const shadowId = useShadowId();
  const { hidden, toggle } = useLegendToggle(["down", "up"]);

  if (!points.length) {
    return <div className="flex h-44 items-center justify-center text-xs text-fg-subtle">No traffic data yet</div>;
  }

  const data = points.map((p) => ({
    ...p,
    label: new Date(p.time).toLocaleTimeString([], { hour: "2-digit", minute: "2-digit" }),
  }));

  const avgDown = data.reduce((s, p) => s + p.down, 0) / data.length;
  const avgUp = data.reduce((s, p) => s + p.up, 0) / data.length;

  const legendEntries: LegendEntry[] = [
    { dataKey: "down", name: "Download", color: "#0EA5E9" },
    { dataKey: "up", name: "Upload", color: "#6366F1" },
  ];

  return (
    <div className="space-y-2">
      <InteractiveLegend entries={legendEntries} hidden={hidden} onToggle={toggle} />

      <div className="h-52 w-full">
        <ResponsiveContainer width="100%" height="100%">
          <AreaChart data={data} margin={CHART_MARGIN}>
            <ChartEffects id={shadowId} />
            <TrafficGradients downId={downGradId} upId={upGradId} />
            <ChartGrid />
            <XAxis
              dataKey="label"
              tick={{ fontSize: 9, fill: "hsl(var(--fg-subtle) / 0.55)" }}
              tickLine={false}
              axisLine={false}
              interval="preserveStartEnd"
            />
            <YAxis
              tick={{ fontSize: 9, fill: "hsl(var(--fg-subtle) / 0.55)" }}
              tickLine={false}
              axisLine={false}
              tickFormatter={axisBytes}
              width={40}
            />
            <Tooltip content={<ChartTooltip />} cursor={{ stroke: "hsl(var(--fg) / 0.15)", strokeDasharray: "3 3" }} />

            {!hidden.has("down") && (
              <ReferenceLine
                y={avgDown}
                stroke="#0EA5E9"
                strokeDasharray="4 4"
                strokeOpacity={0.3}
                label={{ value: `avg ${formatBytes(avgDown, false)}`, position: "insideTopLeft", fontSize: 8, fill: "#0EA5E9", opacity: 0.5 }}
              />
            )}
            {!hidden.has("up") && (
              <ReferenceLine
                y={avgUp}
                stroke="#6366F1"
                strokeDasharray="4 4"
                strokeOpacity={0.3}
                label={{ value: `avg ${formatBytes(avgUp, false)}`, position: "insideBottomLeft", fontSize: 8, fill: "#6366F1", opacity: 0.5 }}
              />
            )}

            {!hidden.has("down") && (
              <Area
                type="monotone"
                dataKey="down"
                name="Download"
                stroke="#0EA5E9"
                strokeWidth={2}
                fill={`url(#${downGradId})`}
                dot={false}
                activeDot={{ r: 4, fill: "#0EA5E9", stroke: "hsl(var(--bg-elevated))", strokeWidth: 2, filter: `url(#${shadowId})` }}
              />
            )}
            {!hidden.has("up") && (
              <Area
                type="monotone"
                dataKey="up"
                name="Upload"
                stroke="#6366F1"
                strokeWidth={2}
                fill={`url(#${upGradId})`}
                dot={false}
                activeDot={{ r: 4, fill: "#6366F1", stroke: "hsl(var(--bg-elevated))", strokeWidth: 2, filter: `url(#${shadowId})` }}
              />
            )}

            <ChartBrush data={data} dataKey="down" />
          </AreaChart>
        </ResponsiveContainer>
      </div>

      <div className="flex items-center justify-between text-[10px] text-fg-muted">
        <span className="flex items-center gap-3">
          <span className="flex items-center gap-1.5">
            <span className="inline-block h-2 w-2 rounded-full bg-[#0EA5E9]" />
            <span>Download</span>
          </span>
          <span className="flex items-center gap-1.5">
            <span className="inline-block h-2 w-2 rounded-full bg-[#6366F1]" />
            <span>Upload</span>
          </span>
        </span>
        <span className="text-fg-subtle tabular-nums">
          total {formatBytes(data.reduce((s, p) => s + p.up + p.down, 0), false)}
        </span>
      </div>
    </div>
  );
}

// ════════════════════════════════════════════════════════════════
// Revenue Area Chart — with Brush, Legend, ReferenceLine
// ════════════════════════════════════════════════════════════════
export function RevenueAreaChart({ data: series }: { data: RevenuePoint[] }) {
  const inGradId = useSafeId("rI");
  const exGradId = useSafeId("rE");
  const shadowId = useShadowId();
  const { hidden, toggle } = useLegendToggle(["income", "expense"]);

  if (!series.length) {
    return <div className="h-32 flex items-center justify-center text-xs text-fg-subtle">No revenue data</div>;
  }

  const fmtCurrency = (v: number) =>
    v >= 1_000_000 ? `${(v / 1_000_000).toFixed(1)}M` : v >= 1_000 ? `${(v / 1_000).toFixed(1)}K` : String(v);

  const avgIncome = series.reduce((s, p) => s + p.income, 0) / series.length;
  const avgExpense = series.reduce((s, p) => s + p.expense, 0) / series.length;

  const legendEntries: LegendEntry[] = [
    { dataKey: "income", name: "Income", color: "#22c55e", icon: "rect" },
    { dataKey: "expense", name: "Expense", color: "#ef4444", icon: "rect" },
  ];

  return (
    <div className="space-y-1">
      <InteractiveLegend entries={legendEntries} hidden={hidden} onToggle={toggle} />

      <div className="h-36 w-full">
        <ResponsiveContainer width="100%" height="100%">
          <AreaChart data={series} margin={{ top: 4, right: 4, left: 0, bottom: 0 }}>
            <ChartEffects id={shadowId} />
            <defs>
              <linearGradient id={inGradId} x1="0" y1="0" x2="0" y2="1">
                <stop offset="0%" stopColor="#22c55e" stopOpacity={0.5} />
                <stop offset="100%" stopColor="#22c55e" stopOpacity={0.03} />
              </linearGradient>
              <linearGradient id={exGradId} x1="0" y1="0" x2="0" y2="1">
                <stop offset="0%" stopColor="#ef4444" stopOpacity={0.35} />
                <stop offset="100%" stopColor="#ef4444" stopOpacity={0.03} />
              </linearGradient>
            </defs>
            <CartesianGrid strokeDasharray="3 3" stroke="hsl(var(--border) / 0.18)" vertical={false} />
            <XAxis
              dataKey="date"
              tick={{ fontSize: 8, fill: "hsl(var(--fg-subtle) / 0.45)" }}
              tickLine={false}
              axisLine={false}
            />
            <YAxis
              tick={{ fontSize: 8, fill: "hsl(var(--fg-subtle) / 0.45)" }}
              tickLine={false}
              axisLine={false}
              tickFormatter={fmtCurrency}
              width={30}
            />
            <Tooltip
              content={<ChartTooltip formatter={fmtCurrency} />}
              cursor={{ stroke: "hsl(var(--fg) / 0.12)", strokeDasharray: "3 3" }}
            />

            {!hidden.has("income") && (
              <ReferenceLine
                y={avgIncome}
                stroke="#22c55e"
                strokeDasharray="4 4"
                strokeOpacity={0.3}
                label={{ value: `avg ${fmtCurrency(avgIncome)}`, position: "insideTopLeft", fontSize: 8, fill: "#22c55e", opacity: 0.5 }}
              />
            )}
            {!hidden.has("expense") && (
              <ReferenceLine
                y={avgExpense}
                stroke="#ef4444"
                strokeDasharray="4 4"
                strokeOpacity={0.3}
                label={{ value: `avg ${fmtCurrency(avgExpense)}`, position: "insideBottomLeft", fontSize: 8, fill: "#ef4444", opacity: 0.5 }}
              />
            )}

            {!hidden.has("income") && (
              <Area
                type="monotone"
                dataKey="income"
                name="Income"
                stroke="#22c55e"
                strokeWidth={2}
                fill={`url(#${inGradId})`}
                dot={false}
                activeDot={{ r: 3, fill: "#22c55e", stroke: "hsl(var(--bg-elevated))", strokeWidth: 2, filter: `url(#${shadowId})` }}
              />
            )}
            {!hidden.has("expense") && (
              <Area
                type="monotone"
                dataKey="expense"
                name="Expense"
                stroke="#ef4444"
                strokeWidth={2}
                fill={`url(#${exGradId})`}
                dot={false}
                activeDot={{ r: 3, fill: "#ef4444", stroke: "hsl(var(--bg-elevated))", strokeWidth: 2, filter: `url(#${shadowId})` }}
              />
            )}

            <ChartBrush data={series} dataKey="income" />
          </AreaChart>
        </ResponsiveContainer>
      </div>
    </div>
  );
}

// ════════════════════════════════════════════════════════════════
// Usage Bar Chart — with Brush + ReferenceLine
// ════════════════════════════════════════════════════════════════
export function UsageBarChart({ points, compact }: { points: UsagePoint[]; compact?: boolean }) {
  const shadowId = useShadowId();
  const { hidden, toggle } = useLegendToggle(["down", "up"]);

  if (!points.length) {
    return <div className="h-32 flex items-center justify-center text-xs text-fg-subtle">No usage data</div>;
  }

  const data = points.map((p) => ({
    ...p,
    label: new Date(p.time).toLocaleDateString([], { weekday: "short", day: "numeric" }),
  }));

  const avgTotal = data.reduce((s, p) => s + p.down + p.up, 0) / data.length;

  const legendEntries: LegendEntry[] = [
    { dataKey: "down", name: "Download", color: "#7C3AED", icon: "square" },
    { dataKey: "up", name: "Upload", color: "#22D3EE", icon: "square" },
  ];

  return (
    <div className="space-y-1">
      <InteractiveLegend entries={legendEntries} hidden={hidden} onToggle={toggle} />

      <div className={compact ? "h-28 w-full" : "h-36 w-full"}>
        <ResponsiveContainer width="100%" height="100%">
          <BarChart data={data} margin={{ top: 4, right: 4, left: -12, bottom: 0 }}>
            <ChartEffects id={shadowId} />
            <CartesianGrid strokeDasharray="3 3" stroke="hsl(var(--border) / 0.18)" vertical={false} />
            <XAxis
              dataKey="label"
              tick={{ fontSize: compact ? 8 : 9, fill: "hsl(var(--fg-subtle) / 0.55)" }}
              tickLine={false}
              axisLine={false}
              interval="preserveStartEnd"
            />
            <YAxis
              tick={{ fontSize: 8, fill: "hsl(var(--fg-subtle) / 0.45)" }}
              tickLine={false}
              axisLine={false}
              tickFormatter={axisBytes}
              width={36}
            />
            <Tooltip content={<ChartTooltip />} cursor={{ fill: "hsl(var(--fg) / 0.05)" }} />

            <ReferenceLine
              y={avgTotal}
              stroke="hsl(var(--fg-subtle))"
              strokeDasharray="4 4"
              strokeOpacity={0.25}
              label={{ value: `avg ${formatBytes(avgTotal, false)}`, position: "insideTopLeft", fontSize: 8, fill: "hsl(var(--fg-subtle))", opacity: 0.5 }}
            />

            {!hidden.has("down") && (
              <Bar
                dataKey="down"
                name="Download"
                fill="#7C3AED"
                radius={[3, 3, 0, 0]}
                stackId="a"
                isAnimationActive={true}
                animationDuration={400}
              />
            )}
            {!hidden.has("up") && (
              <Bar
                dataKey="up"
                name="Upload"
                fill="#22D3EE"
                radius={[3, 3, 0, 0]}
                stackId="a"
                isAnimationActive={true}
                animationDuration={400}
              />
            )}

            <ChartBrush data={data} dataKey="down" />
          </BarChart>
        </ResponsiveContainer>
      </div>
    </div>
  );
}

// ════════════════════════════════════════════════════════════════
// Peak Hours Bar Chart — with Brush + ReferenceLine
// ════════════════════════════════════════════════════════════════
export function PeakHoursChart({ data }: { data: { hour: number; bytes_total: number }[] }) {
  const shadowId = useShadowId();
  const series = data.map((p) => ({
    hour: `${p.hour}:00`,
    traffic: p.bytes_total,
  }));

  if (!series.length) {
    return <div className="h-40 flex items-center justify-center text-xs text-fg-subtle">No peak hour data</div>;
  }

  const avgTraffic = series.reduce((s, p) => s + p.traffic, 0) / series.length;
  const peakTraffic = Math.max(...series.map((p) => p.traffic));

  return (
    <div className="space-y-1">
      <div className="h-44 w-full">
        <ResponsiveContainer width="100%" height="100%">
          <BarChart data={series} margin={{ top: 4, right: 4, left: -12, bottom: 0 }}>
            <ChartEffects id={shadowId} />
            <defs>
              <linearGradient id="peakBarGrad" x1="0" y1="0" x2="0" y2="1">
                <stop offset="0%" stopColor="hsl(var(--primary))" stopOpacity={0.8} />
                <stop offset="100%" stopColor="hsl(var(--primary))" stopOpacity={0.3} />
              </linearGradient>
            </defs>
            <CartesianGrid strokeDasharray="3 3" stroke="hsl(var(--border) / 0.18)" vertical={false} />
            <XAxis
              dataKey="hour"
              tick={{ fontSize: 8, fill: "hsl(var(--fg-subtle) / 0.5)" }}
              tickLine={false}
              axisLine={false}
              interval={2}
            />
            <YAxis
              tick={{ fontSize: 8, fill: "hsl(var(--fg-subtle) / 0.5)" }}
              tickLine={false}
              axisLine={false}
              tickFormatter={axisBytes}
              width={36}
            />
            <Tooltip
              content={<ChartTooltip formatter={(v: number) => formatBytes(v, false)} />}
              cursor={{ fill: "hsl(var(--fg) / 0.05)" }}
            />

            <ReferenceLine
              y={avgTraffic}
              stroke="hsl(var(--fg-subtle))"
              strokeDasharray="4 4"
              strokeOpacity={0.25}
              label={{ value: `avg ${formatBytes(avgTraffic, false)}`, position: "insideTopLeft", fontSize: 8, fill: "hsl(var(--fg-subtle))", opacity: 0.5 }}
            />

            <Bar
              dataKey="traffic"
              name="Traffic"
              fill="url(#peakBarGrad)"
              radius={[3, 3, 0, 0]}
              isAnimationActive={true}
              animationDuration={500}
              shape={(props: any) => {
                const { x, y, width, height, payload } = props;
                const isPeak = payload.traffic >= peakTraffic * 0.85;
                const color = isPeak ? "#F59E0B" : "url(#peakBarGrad)";
                return (
                  <rect
                    x={x}
                    y={y}
                    width={width}
                    height={height}
                    fill={color}
                    rx={3}
                    ry={3}
                    filter={isPeak ? `url(#${shadowId})` : undefined}
                  />
                );
              }}
            />

            <ChartBrush data={series} dataKey="traffic" />
          </BarChart>
        </ResponsiveContainer>
      </div>
    </div>
  );
}

// ════════════════════════════════════════════════════════════════
// Donut Chart — v2 with gradient fills, click-to-filter, smoother transitions
// ════════════════════════════════════════════════════════════════
export function DonutChart({
  slices,
  centerLabel,
  centerValue,
  loading,
  loadingLabel = "Loading…",
  onSliceClick,
}: {
  slices: PieSlice[];
  centerLabel: string;
  centerValue: string | number;
  loading?: boolean;
  loadingLabel?: string;
  /** Called when a legend item is clicked — returns the slice label */
  onSliceClick?: (label: string) => void;
}) {
  const [activeIndex, setActiveIndex] = useState<number | null>(null);
  const [hiddenSlices, setHiddenSlices] = useState<Set<string>>(new Set());
  const gradId = useSafeId("pie-glow");
  const anim = useChartAnimation(800);

  const visibleSlices = slices.filter((s) => !hiddenSlices.has(s.label));
  const total = visibleSlices.reduce((s, x) => s + x.value, 0);

  if (loading) {
    return (
      <div className="flex flex-col items-center gap-2">
        <div className="relative h-40 w-40 flex-shrink-0">
          <div className="absolute inset-0 rounded-full shimmer" />
          <div className="absolute inset-4 rounded-full bg-bg-elevated" />
        </div>
        <span className="text-[10px] text-fg-subtle">{loadingLabel}</span>
      </div>
    );
  }

  if (!total) {
    return <div className="h-40 flex items-center justify-center text-xs text-fg-subtle">No data</div>;
  }

  const renderActiveShape = (props: any) => {
    const { cx, cy, innerRadius, outerRadius, startAngle, endAngle, fill, payload } = props;
    const isActive = activeIndex !== null && visibleSlices[activeIndex]?.label === payload?.label;

    return (
      <g>
        {/* Outer glow ring */}
        {isActive && (
          <Sector
            cx={cx}
            cy={cy}
            innerRadius={outerRadius + 4}
            outerRadius={outerRadius + 10}
            startAngle={startAngle}
            endAngle={endAngle}
            fill={fill}
            opacity={0.2}
            filter={`url(#${gradId})`}
          />
        )}
        {/* Main segment */}
        <Sector
          cx={cx}
          cy={cy}
          innerRadius={innerRadius}
          outerRadius={isActive ? outerRadius + 6 : outerRadius}
          startAngle={startAngle}
          endAngle={endAngle}
          fill={fill}
          opacity={isActive ? 1 : 0.85}
          filter={isActive ? `url(#${gradId})` : undefined}
          stroke={isActive ? "hsl(var(--bg-elevated))" : "none"}
          strokeWidth={isActive ? 2 : 0}
        />
        {/* Inner highlight for 3D effect */}
        {isActive && (
          <Sector
            cx={cx}
            cy={cy}
            innerRadius={innerRadius}
            outerRadius={innerRadius + 4}
            startAngle={startAngle}
            endAngle={endAngle}
            fill="white"
            opacity={0.12}
          />
        )}
      </g>
    );
  };

  const hovered = activeIndex !== null && activeIndex < visibleSlices.length ? visibleSlices[activeIndex] : null;
  const hoveredPct = hovered ? ((hovered.value / total) * 100).toFixed(0) : null;

  return (
    <div className="flex flex-col items-center gap-2">
      <div className="relative h-40 w-40 flex-shrink-0">
        {/* SVG filter definitions */}
        <svg style={{ position: "absolute", width: 0, height: 0 }}>
          <defs>
            <filter id={gradId} x="-50%" y="-50%" width="200%" height="200%">
              <feGaussianBlur in="SourceAlpha" stdDeviation="3" result="blur" />
              <feOffset dx="0" dy="2" result="offsetBlur" />
              <feMerge>
                <feMergeNode in="offsetBlur" />
                <feMergeNode in="SourceGraphic" />
              </feMerge>
            </filter>
            <linearGradient id={`${gradId}-grad`} x1="0" y1="0" x2="1" y2="1">
              <stop offset="0%" stopColor="hsl(var(--glow-primary))" stopOpacity="0.8" />
              <stop offset="100%" stopColor="hsl(var(--glow-accent))" stopOpacity="0.6" />
            </linearGradient>
          </defs>
        </svg>

        <ResponsiveContainer width="100%" height="100%">
          <PieChart>
            <Pie
              data={visibleSlices}
              cx="50%"
              cy="50%"
              innerRadius={38}
              outerRadius={56}
              dataKey="value"
              startAngle={90}
              endAngle={-270}
              stroke="none"
              activeShape={activeIndex !== null ? renderActiveShape : undefined}
              onMouseEnter={(_, i) => setActiveIndex(i)}
              onMouseLeave={() => setActiveIndex(null)}
              {...anim}
            >
              {visibleSlices.map((slice) => {
                const isActive = activeIndex !== null && visibleSlices[activeIndex]?.label === slice.label;
                return (
                  <Cell
                    key={slice.label}
                    fill={slice.color}
                    opacity={activeIndex !== null ? (isActive ? 1 : 0.25) : 0.9}
                    stroke={isActive ? "hsl(var(--bg-elevated))" : "none"}
                    strokeWidth={isActive ? 2 : 0}
                  />
                );
              })}
            </Pie>
          </PieChart>
        </ResponsiveContainer>

        {/* Center text — animated between states */}
        <div className="absolute inset-0 flex flex-col items-center justify-center text-center px-2 pointer-events-none">
          <div className="transition-all duration-200 ease-out">
            {hovered ? (
              <div key="hovered" className="animate-fade-in">
                <span className="text-lg font-black text-fg tabular-nums leading-none">{hoveredPct}%</span>
                <span className="text-[8px] font-bold uppercase tracking-wide text-fg-subtle mt-0.5 block max-w-[70px] truncate mx-auto">
                  {hovered.label}
                </span>
                <span className="text-[9px] text-fg-muted/70 tabular-nums mt-0.5 block font-mono">
                  {hovered.value.toLocaleString()}
                </span>
              </div>
            ) : (
              <div key="default" className="animate-fade-in">
                <span className="text-2xl font-black text-fg tabular-nums leading-none">{centerValue}</span>
                <span className="text-[7.5px] font-bold uppercase tracking-widest text-fg-subtle mt-0.5 block">
                  {centerLabel}
                </span>
              </div>
            )}
          </div>
        </div>
      </div>

      {/* Legend — interactive hover + click-to-toggle */}
      <div className="w-full flex flex-wrap justify-center gap-1.5">
        {slices.map((slice) => {
          const isHidden = hiddenSlices.has(slice.label);
          return (
            <div
              key={slice.label}
              onMouseEnter={() => {
                const idx = visibleSlices.findIndex((s) => s.label === slice.label);
                if (idx >= 0) setActiveIndex(idx);
              }}
              onMouseLeave={() => setActiveIndex(null)}
              onClick={() => {
                if (onSliceClick) {
                  onSliceClick(slice.label);
                } else if (visibleSlices.length > 1) {
                  setHiddenSlices((prev) => {
                    const next = new Set(prev);
                    if (next.has(slice.label)) next.delete(slice.label);
                    else next.add(slice.label);
                    return next;
                  });
                }
              }}
              className={cn(
                "flex items-center gap-1.5 px-2.5 py-1 rounded-lg text-[10px] cursor-pointer transition-all duration-200 select-none",
                activeIndex !== null && visibleSlices[activeIndex]?.label === slice.label
                  ? "bg-surface-2/80 shadow-sm border border-border/40 scale-105"
                  : isHidden
                    ? "opacity-30 line-through hover:opacity-60 hover:bg-surface-2/20 border border-transparent"
                    : "hover:bg-surface-2/40 border border-transparent",
              )}
            >
              <span
                className="h-2.5 w-2.5 rounded-full flex-shrink-0"
                style={{
                  background: slice.color,
                  boxShadow: activeIndex !== null && visibleSlices[activeIndex]?.label === slice.label
                    ? `0 0 6px ${slice.color}80`
                    : "none",
                }}
              />
              <span className="text-fg-muted">{slice.label}</span>
              <span className="font-bold text-fg tabular-nums">
                {((slice.value / slices.reduce((s, x) => s + x.value, 0)) * 100).toFixed(0)}%
              </span>
            </div>
          );
        })}
      </div>
    </div>
  );
}

// ════════════════════════════════════════════════════════════════
// Mini Sparkline (Recharts-based)
// ════════════════════════════════════════════════════════════════
export function Sparkline({ data, color = "hsl(var(--primary))" }: { data: number[]; color?: string }) {
  const gradId = useSafeId("spk");
  const anim = useChartAnimation(200);

  if (!data.length) return null;

  const chartData = data.map((v, i) => ({ i, v }));

  return (
    <div className="h-8 w-full">
      <ResponsiveContainer width="100%" height="100%">
        <AreaChart data={chartData} margin={{ top: 0, right: 0, left: 0, bottom: 0 }}>
          <defs>
            <linearGradient id={gradId} x1="0" y1="0" x2="0" y2="1">
              <stop offset="0%" stopColor={color} stopOpacity={0.45} />
              <stop offset="100%" stopColor={color} stopOpacity={0.02} />
            </linearGradient>
          </defs>
          <Area
            type="monotone"
            dataKey="v"
            stroke={color}
            strokeWidth={1.5}
            fill={`url(#${gradId})`}
            dot={false}
            {...anim}
          />
        </AreaChart>
      </ResponsiveContainer>
    </div>
  );
}

// ════════════════════════════════════════════════════════════════
// Mini Progress Bar (unchanged)
// ════════════════════════════════════════════════════════════════
export function MiniProgressBar({ value, color = "primary" }: { value: number; color?: "primary" | "success" | "warning" | "danger" | "accent" }) {
  const pct = Math.min(100, Math.max(0, value || 0));
  const colors = {
    primary: "bg-primary",
    success: "bg-success",
    warning: "bg-warning",
    danger: "bg-danger",
    accent: "bg-accent",
  };

  return (
    <div className="h-1.5 w-full rounded-full bg-surface-3/60 overflow-hidden">
      <div
        className={cn("h-full rounded-full transition-all duration-700", colors[color])}
        style={{ width: `${pct}%` }}
      />
    </div>
  );
}
