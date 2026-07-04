import { useRef, useState } from "react";
import { formatBytes } from "@/lib/utils";

export interface TrafficSeriesPoint {
  time: string;
  up: number;
  down: number;
}

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
    return <div className="flex h-52 items-center justify-center text-xs text-fg-subtle">Collecting data…</div>;
  }

  const w = 620;
  const h = 190;
  const padL = 30;
  const padR = 4;
  const padTop = 12;
  const padBottom = 24;
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

  const tickCount = Math.min(6, n);
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

  const yTicks = [max, max / 2, 0];
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

  // Keep tooltip box inside chart bounds.
  const tooltipW = 108;
  const tooltipX = Math.min(Math.max(hx - tooltipW / 2, padL), w - padR - tooltipW);

  return (
    <div className="space-y-3">
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
            <stop offset="0%" stopColor="#8B7CF6" stopOpacity="0.38" />
            <stop offset="55%" stopColor="#8B7CF6" stopOpacity="0.12" />
            <stop offset="100%" stopColor="#8B7CF6" stopOpacity="0.02" />
          </linearGradient>
          <linearGradient id="ts-up-g" x1="0" y1="0" x2="0" y2="1">
            <stop offset="0%" stopColor="#2DD4BF" stopOpacity="0.4" />
            <stop offset="60%" stopColor="#2DD4BF" stopOpacity="0.1" />
            <stop offset="100%" stopColor="#2DD4BF" stopOpacity="0.02" />
          </linearGradient>
          <filter id="ts-glow" x="-20%" y="-20%" width="140%" height="140%">
            <feGaussianBlur stdDeviation="1.6" result="blur" />
            <feMerge>
              <feMergeNode in="blur" />
              <feMergeNode in="SourceGraphic" />
            </feMerge>
          </filter>
        </defs>

        {/* Y-axis labels */}
        {yTicks.map((v, i) => (
          <text key={i} x={padL - 6} y={py(v) + 3} textAnchor="end" className="fill-current opacity-35" style={{ fontSize: 9 }}>
            {v === 0 ? "0" : formatBytes(v, false).replace(/\.\d+/, "")}
          </text>
        ))}

        {/* Download area (total) */}
        <path d={area(totals)} fill="url(#ts-down-g)" />
        <path
          d={smoothPath(totals)}
          fill="none"
          stroke="#8B7CF6"
          strokeWidth="2"
          strokeLinecap="round"
          strokeLinejoin="round"
          filter="url(#ts-glow)"
        />

        {/* Upload area */}
        <path d={area(ups)} fill="url(#ts-up-g)" />
        <path
          d={smoothPath(ups)}
          fill="none"
          stroke="#2DD4BF"
          strokeWidth="1.75"
          strokeLinecap="round"
          strokeLinejoin="round"
          strokeOpacity="0.95"
        />

        {/* Baseline */}
        <line x1={padL} y1={h - padBottom} x2={w - padR} y2={h - padBottom} stroke="currentColor" strokeOpacity="0.08" />

        {/* X-axis ticks */}
        {xTicks.map(({ x, label }, i) => (
          <text
            key={i}
            x={x}
            y={h - 6}
            textAnchor={i === 0 ? "start" : i === xTicks.length - 1 ? "end" : "middle"}
            className="fill-current opacity-35"
            style={{ fontSize: 9 }}
          >
            {label}
          </text>
        ))}

        {/* Hover crosshair + tooltip */}
        {hp && (
          <g pointerEvents="none">
            <line
              x1={hx}
              y1={padTop}
              x2={hx}
              y2={h - padBottom}
              stroke="currentColor"
              strokeOpacity="0.25"
              strokeDasharray="3 3"
            />
            <circle cx={hx} cy={py(hp.up + hp.down)} r="3.5" fill="#8B7CF6" stroke="white" strokeWidth="1.5" />
            <circle cx={hx} cy={py(hp.up)} r="3.5" fill="#2DD4BF" stroke="white" strokeWidth="1.5" />

            <foreignObject x={tooltipX} y={padTop} width={tooltipW} height={62}>
              <div className="rounded-lg border border-border/70 bg-bg-elevated/95 backdrop-blur-sm px-2.5 py-2 shadow-lg text-[10px] leading-tight">
                <p className="text-fg-subtle font-semibold mb-1">{fmtTooltipTime(hp.time, spanMs)}</p>
                <p className="flex items-center justify-between gap-2 text-fg">
                  <span className="flex items-center gap-1"><span className="h-1.5 w-1.5 rounded-full inline-block" style={{ background: "#8B7CF6" }} />DL</span>
                  <span className="font-bold tabular-nums">{formatBytes(hp.down, false)}</span>
                </p>
                <p className="flex items-center justify-between gap-2 text-fg">
                  <span className="flex items-center gap-1"><span className="h-1.5 w-1.5 rounded-full inline-block" style={{ background: "#2DD4BF" }} />UL</span>
                  <span className="font-bold tabular-nums">{formatBytes(hp.up, false)}</span>
                </p>
              </div>
            </foreignObject>
          </g>
        )}
      </svg>

      <div className="flex items-center justify-between text-[11px] text-fg-muted">
        <span className="flex items-center gap-4">
          <span className="flex items-center gap-1.5">
            <span className="inline-block h-2.5 w-2.5 rounded-full" style={{ background: "#2DD4BF" }} />
            <span>Upload</span>
          </span>
          <span className="flex items-center gap-1.5">
            <span className="inline-block h-2.5 w-2.5 rounded-full" style={{ background: "#8B7CF6" }} />
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
