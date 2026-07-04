import { useRef, useState } from "react";
import { formatBytes } from "@/lib/utils";

export interface TrafficSeriesPoint {
  time: string;
  up: number;
  down: number;
}

const DOWN_COLOR = "#6366F1"; // indigo-500
const UP_COLOR = "#10B981"; // emerald-500

/** Format a bucket timestamp based on the total span of the series. */
function fmtTick(iso: string, spanMs: number): string {
  const d = new Date(iso);
  if (Number.isNaN(d.getTime())) return "";
  if (spanMs <= 32 * 3600 * 1000) {
    return d.toLocaleTimeString([], { hour: "2-digit", minute: "2-digit" });
  }
  if (spanMs <= 9 * 86400 * 1000) {
    return d.toLocaleDateString([], { weekday: "short" });
  }
  return d.toLocaleDateString([], { month: "short", day: "numeric" });
}

function fmtTooltipTime(iso: string, spanMs: number): string {
  const d = new Date(iso);
  if (Number.isNaN(d.getTime())) return iso;
  if (spanMs <= 32 * 3600 * 1000) {
    return d.toLocaleTimeString([], { hour: "2-digit", minute: "2-digit" });
  }
  return d.toLocaleDateString([], { month: "short", day: "numeric", hour: "2-digit", minute: "2-digit" });
}

export function TrafficSeriesChart({ points }: { points: TrafficSeriesPoint[] }) {
  const svgRef = useRef<SVGSVGElement>(null);
  const [hoverIdx, setHoverIdx] = useState<number | null>(null);

  if (points.length < 2) {
    return <div className="flex h-28 items-center justify-center text-xs text-fg-subtle">Collecting data…</div>;
  }

  const w = 620;
  const h = 116;
  const padL = 26;
  const padR = 4;
  const padTop = 8;
  const padBottom = 16;
  const innerH = h - padTop - padBottom;
  const max = Math.max(1, ...points.map((p) => p.up + p.down));
  const n = points.length;

  const px = (i: number) => padL + (i / (n - 1)) * (w - padL - padR);
  const py = (v: number) => padTop + (1 - v / max) * innerH;

  const smoothPath = (vals: number[]) => {
    const pts = vals.map((v, i) => [px(i), py(v)] as [number, number]);
    if (pts.length < 2) return "";
    let d = `M ${pts[0][0]},${pts[0][1]}`;
    for (let i = 1; i < pts.length; i++) {
      const [x0, y0] = pts[i - 1];
      const [x1, y1] = pts[i];
      const cx = (x0 + x1) / 2;
      d += ` C ${cx},${y0} ${cx},${y1} ${x1},${y1}`;
    }
    return d;
  };

  const area = (vals: number[]) => {
    const line = smoothPath(vals);
    if (!line) return "";
    const base = h - padBottom;
    return `${line} L ${px(n - 1)},${base} L ${px(0)},${base} Z`;
  };

  const totals = points.map((p) => p.up + p.down);
  const ups = points.map((p) => p.up);

  const spanMs = (() => {
    const t0 = new Date(points[0].time).getTime();
    const t1 = new Date(points[n - 1].time).getTime();
    return Number.isFinite(t0) && Number.isFinite(t1) ? Math.abs(t1 - t0) : 0;
  })();

  const tickCount = Math.min(5, n);
  const xTicks: { x: number; label: string }[] = [];
  let lastLabel = "";
  for (let k = 0; k < tickCount; k++) {
    const i = Math.round((k / (tickCount - 1)) * (n - 1));
    const label = fmtTick(points[i].time, spanMs);
    const isEdge = k === 0 || k === tickCount - 1;
    if (label && (isEdge || label !== lastLabel)) {
      xTicks.push({ x: px(i), label });
      lastLabel = label;
    }
  }

  const peak = Math.max(...totals);
  const total = totals.reduce((a, b) => a + b, 0);

  function handleMove(e: React.MouseEvent<SVGSVGElement>) {
    const svg = svgRef.current;
    if (!svg) return;
    const rect = svg.getBoundingClientRect();
    const relX = ((e.clientX - rect.left) / rect.width) * w;
    const ratio = (relX - padL) / (w - padL - padR);
    const idx = Math.round(ratio * (n - 1));
    setHoverIdx(Math.max(0, Math.min(n - 1, idx)));
  }

  const hp = hoverIdx !== null ? points[hoverIdx] : null;
  const hx = hoverIdx !== null ? px(hoverIdx) : 0;

  const tooltipW = 96;
  const tooltipX = Math.min(Math.max(hx - tooltipW / 2, padL), w - padR - tooltipW);

  return (
    <div className="space-y-1.5">
      <svg
        ref={svgRef}
        viewBox={`0 0 ${w} ${h}`}
        className="w-full cursor-crosshair select-none"
        role="img"
        aria-label="Fleet-wide traffic over time"
        onMouseMove={handleMove}
        onMouseLeave={() => setHoverIdx(null)}
      >
        <defs>
          <linearGradient id="ts-down-g" x1="0" y1="0" x2="0" y2="1">
            <stop offset="0%" stopColor={DOWN_COLOR} stopOpacity="0.3" />
            <stop offset="70%" stopColor={DOWN_COLOR} stopOpacity="0.05" />
            <stop offset="100%" stopColor={DOWN_COLOR} stopOpacity="0" />
          </linearGradient>
          <linearGradient id="ts-up-g" x1="0" y1="0" x2="0" y2="1">
            <stop offset="0%" stopColor={UP_COLOR} stopOpacity="0.3" />
            <stop offset="70%" stopColor={UP_COLOR} stopOpacity="0.05" />
            <stop offset="100%" stopColor={UP_COLOR} stopOpacity="0" />
          </linearGradient>
        </defs>

        {/* Baseline only — clean, minimal, no grid clutter */}
        <line x1={padL} y1={h - padBottom} x2={w - padR} y2={h - padBottom} stroke="currentColor" strokeOpacity="0.08" />

        {/* Download area (total) */}
        <path d={area(totals)} fill="url(#ts-down-g)" />
        <path d={smoothPath(totals)} fill="none" stroke={DOWN_COLOR} strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round" />

        {/* Upload area */}
        <path d={area(ups)} fill="url(#ts-up-g)" />
        <path d={smoothPath(ups)} fill="none" stroke={UP_COLOR} strokeWidth="1.3" strokeLinecap="round" strokeLinejoin="round" strokeOpacity="0.9" />

        {/* X-axis ticks */}
        {xTicks.map(({ x, label }, i) => (
          <text
            key={i}
            x={x}
            y={h - 3}
            textAnchor={i === 0 ? "start" : i === xTicks.length - 1 ? "end" : "middle"}
            className="fill-current opacity-30"
            style={{ fontSize: 8 }}
          >
            {label}
          </text>
        ))}

        {/* Hover crosshair + tooltip */}
        {hp && (
          <g pointerEvents="none">
            <line x1={hx} y1={padTop} x2={hx} y2={h - padBottom} stroke="currentColor" strokeOpacity="0.22" strokeDasharray="2 3" />
            <circle cx={hx} cy={py(hp.up + hp.down)} r="2.75" fill={DOWN_COLOR} stroke="white" strokeWidth="1.1" />
            <circle cx={hx} cy={py(hp.up)} r="2.75" fill={UP_COLOR} stroke="white" strokeWidth="1.1" />

            <foreignObject x={tooltipX} y={2} width={tooltipW} height={50}>
              <div className="rounded-lg border border-border/70 bg-bg-elevated/95 backdrop-blur-sm px-2 py-1.5 shadow-lg text-[9px] leading-tight">
                <p className="text-fg-subtle font-semibold mb-1">{fmtTooltipTime(hp.time, spanMs)}</p>
                <p className="flex items-center justify-between gap-2 text-fg">
                  <span className="flex items-center gap-1"><span className="h-1.5 w-1.5 rounded-full inline-block" style={{ background: DOWN_COLOR }} />DL</span>
                  <span className="font-bold tabular-nums">{formatBytes(hp.down, false)}</span>
                </p>
                <p className="flex items-center justify-between gap-2 text-fg">
                  <span className="flex items-center gap-1"><span className="h-1.5 w-1.5 rounded-full inline-block" style={{ background: UP_COLOR }} />UL</span>
                  <span className="font-bold tabular-nums">{formatBytes(hp.up, false)}</span>
                </p>
              </div>
            </foreignObject>
          </g>
        )}
      </svg>

      <div className="flex items-center justify-between text-[10px] text-fg-muted">
        <span className="flex items-center gap-3">
          <span className="flex items-center gap-1.5">
            <span className="inline-block h-2 w-2 rounded-full" style={{ background: UP_COLOR }} />
            <span>Upload</span>
          </span>
          <span className="flex items-center gap-1.5">
            <span className="inline-block h-2 w-2 rounded-full" style={{ background: DOWN_COLOR }} />
            <span>Download</span>
          </span>
        </span>
        <span className="text-fg-subtle">
          peak {formatBytes(peak, false)}/min · total {formatBytes(total, false)}
        </span>
      </div>
    </div>
  );
}
