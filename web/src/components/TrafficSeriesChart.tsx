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

export function TrafficSeriesChart({ points }: { points: TrafficSeriesPoint[] }) {
  if (points.length < 2) {
    return <div className="flex h-44 items-center justify-center text-xs text-fg-subtle">Collecting data…</div>;
  }

  const w = 600;
  const h = 150;
  const padL = 34;
  const padR = 6;
  const padTop = 14;
  const padBottom = 22;
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

  // Deduplicated X-axis ticks, evenly spaced.
  const spanMs = (() => {
    const t0 = new Date(points[0].time).getTime();
    const t1 = new Date(points[n - 1].time).getTime();
    return Number.isFinite(t0) && Number.isFinite(t1) ? Math.abs(t1 - t0) : 0;
  })();
  const tickCount = Math.min(6, n);
  const xTicks: { x: number; label: string }[] = [];
  const seen = new Set<string>();
  for (let k = 0; k < tickCount; k++) {
    const i = Math.round((k / (tickCount - 1)) * (n - 1));
    const label = fmtTick(points[i].time, spanMs);
    if (label && !seen.has(label)) {
      seen.add(label);
      xTicks.push({ x: px(i), label });
    }
  }

  const yTicks = [max, max / 2, 0];

  const peak = Math.max(...totals);
  const total = totals.reduce((a, b) => a + b, 0);

  return (
    <div className="space-y-2.5">
      <svg viewBox={`0 0 ${w} ${h}`} className="w-full" role="img" aria-label="Fleet-wide traffic over time">
        <defs>
          <linearGradient id="ts-down-g" x1="0" y1="0" x2="0" y2="1">
            <stop offset="0%" stopColor="#7C6EF0" stopOpacity="0.5" />
            <stop offset="90%" stopColor="#7C6EF0" stopOpacity="0.03" />
          </linearGradient>
          <linearGradient id="ts-up-g" x1="0" y1="0" x2="0" y2="1">
            <stop offset="0%" stopColor="#34D399" stopOpacity="0.55" />
            <stop offset="90%" stopColor="#34D399" stopOpacity="0.03" />
          </linearGradient>
        </defs>

        {/* Y-axis labels (outside plot area, left) */}
        {yTicks.map((v, i) => (
          <text
            key={i}
            x={padL - 6}
            y={py(v) + 3}
            textAnchor="end"
            className="fill-current opacity-30"
            style={{ fontSize: 8.5 }}
          >
            {v === 0 ? "0" : formatBytes(v, false).replace(/\.\d+/, "")}
          </text>
        ))}

        {/* Download area (total) */}
        <path d={area(totals)} fill="url(#ts-down-g)" />
        <path d={smoothPath(totals)} fill="none" stroke="#7C6EF0" strokeWidth="1.75" strokeOpacity="0.85" strokeLinecap="round" />

        {/* Upload area */}
        <path d={area(ups)} fill="url(#ts-up-g)" />
        <path d={smoothPath(ups)} fill="none" stroke="#34D399" strokeWidth="1.75" strokeOpacity="0.9" strokeLinecap="round" />

        {/* Baseline */}
        <line x1={padL} y1={h - padBottom} x2={w - padR} y2={h - padBottom} stroke="currentColor" strokeOpacity="0.1" />

        {/* X-axis ticks */}
        {xTicks.map(({ x, label }, i) => (
          <text
            key={i}
            x={x}
            y={h - 6}
            textAnchor={i === 0 ? "start" : i === xTicks.length - 1 ? "end" : "middle"}
            className="fill-current opacity-30"
            style={{ fontSize: 8.5 }}
          >
            {label}
          </text>
        ))}
      </svg>

      <div className="flex items-center justify-between text-[11px] text-fg-muted">
        <span className="flex items-center gap-4">
          <span className="flex items-center gap-1.5">
            <span className="inline-block h-2.5 w-2.5 rounded-full" style={{ background: "#34D399" }} />
            <span>Upload</span>
          </span>
          <span className="flex items-center gap-1.5">
            <span className="inline-block h-2.5 w-2.5 rounded-full" style={{ background: "#7C6EF0" }} />
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
