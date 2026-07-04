import { useRef, useState } from "react";
import { formatBytes } from "@/lib/utils";

export interface TrafficSeriesPoint {
  time: string;
  up: number;
  down: number;
}

const DOWN_COLOR = "#0EA5E9"; // sky-500
const UP_COLOR = "#6366F1"; // indigo-500

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
    return <div className="flex h-44 items-center justify-center text-xs text-fg-subtle">Collecting data…</div>;
  }

  const w = 720;
  const h = 190;
  const padL = 34;
  const padR = 6;
  const padTop = 14;
  const padBottom = 22;
  const innerH = h - padTop - padBottom;
  const downs = points.map((p) => p.down);
  const ups = points.map((p) => p.up);
  const max = Math.max(1, ...downs, ...ups);
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

  const spanMs = (() => {
    const t0 = new Date(points[0].time).getTime();
    const t1 = new Date(points[n - 1].time).getTime();
    return Number.isFinite(t0) && Number.isFinite(t1) ? Math.abs(t1 - t0) : 0;
  })();

  const tickCount = Math.min(7, n);
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

  const yTicks = [max, (max * 3) / 4, max / 2, max / 4, 0];
  const totals = points.map((p) => p.up + p.down);
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

  const tooltipW = 100;
  const tooltipX = Math.min(Math.max(hx - tooltipW / 2, padL), w - padR - tooltipW);

  return (
    <div className="space-y-2">
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
            <stop offset="0%" stopColor={DOWN_COLOR} stopOpacity="0.5" />
            <stop offset="55%" stopColor={DOWN_COLOR} stopOpacity="0.15" />
            <stop offset="100%" stopColor={DOWN_COLOR} stopOpacity="0.01" />
          </linearGradient>
          <linearGradient id="ts-up-g" x1="0" y1="0" x2="0" y2="1">
            <stop offset="0%" stopColor={UP_COLOR} stopOpacity="0.55" />
            <stop offset="55%" stopColor={UP_COLOR} stopOpacity="0.16" />
            <stop offset="100%" stopColor={UP_COLOR} stopOpacity="0.01" />
          </linearGradient>
        </defs>

        {/* Faint horizontal grid */}
        {yTicks.map((v, i) => (
          <line
            key={`grid-${i}`}
            x1={padL}
            y1={py(v)}
            x2={w - padR}
            y2={py(v)}
            stroke="currentColor"
            strokeOpacity={v === 0 ? 0.12 : 0.06}
            strokeDasharray={v === 0 ? undefined : "3 4"}
          />
        ))}

        {/* Y-axis labels */}
        {yTicks.map((v, i) => (
          <text key={i} x={padL - 8} y={py(v) + 3} textAnchor="end" className="fill-current opacity-40" style={{ fontSize: 9 }}>
            {v === 0 ? "0" : formatBytes(v, false).replace(/\.\d+/, "")}
          </text>
        ))}

        {/* Download */}
        <path d={area(downs)} fill="url(#ts-down-g)" />
        <path d={smoothPath(downs)} fill="none" stroke={DOWN_COLOR} strokeWidth="2.25" strokeLinecap="round" strokeLinejoin="round" />

        {/* Upload */}
        <path d={area(ups)} fill="url(#ts-up-g)" />
        <path d={smoothPath(ups)} fill="none" stroke={UP_COLOR} strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" />

        {/* X-axis ticks */}
        {xTicks.map(({ x, label }, i) => (
          <text
            key={i}
            x={x}
            y={h - 6}
            textAnchor={i === 0 ? "start" : i === xTicks.length - 1 ? "end" : "middle"}
            className="fill-current opacity-40"
            style={{ fontSize: 9 }}
          >
            {label}
          </text>
        ))}

        {/* Hover crosshair + tooltip */}
        {hp && (
          <g pointerEvents="none">
            <line x1={hx} y1={padTop} x2={hx} y2={h - padBottom} stroke="currentColor" strokeOpacity="0.22" strokeDasharray="2 3" />
            <circle cx={hx} cy={py(hp.down)} r="3.25" fill={DOWN_COLOR} stroke="white" strokeWidth="1.25" />
            <circle cx={hx} cy={py(hp.up)} r="3.25" fill={UP_COLOR} stroke="white" strokeWidth="1.25" />

            <foreignObject x={tooltipX} y={padTop} width={tooltipW} height={54}>
              <div className="rounded-lg border border-border/70 bg-bg-elevated/95 backdrop-blur-sm px-2 py-1.5 shadow-lg text-[9.5px] leading-tight">
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
            <span className="inline-block h-2 w-2 rounded-full" style={{ background: DOWN_COLOR }} />
            <span>Download</span>
          </span>
          <span className="flex items-center gap-1.5">
            <span className="inline-block h-2 w-2 rounded-full" style={{ background: UP_COLOR }} />
            <span>Upload</span>
          </span>
        </span>
        <span className="text-fg-subtle">
          peak {formatBytes(peak, false)}/min · total {formatBytes(total, false)}
        </span>
      </div>
    </div>
  );
}
