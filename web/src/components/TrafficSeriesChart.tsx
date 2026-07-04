import { formatBytes } from "@/lib/utils";

export interface TrafficSeriesPoint {
  time: string;
  up: number;
  down: number;
}

export function TrafficSeriesChart({ points }: { points: TrafficSeriesPoint[] }) {
  if (points.length < 2) {
    return <div className="flex h-44 items-center justify-center text-xs text-fg-subtle">Collecting data…</div>;
  }

  const w = 600;
  const h = 160;
  const padX = 8;
  const padTop = 12;
  const padBottom = 24;
  const innerH = h - padTop - padBottom;
  const max = Math.max(1, ...points.map((p) => p.up + p.down));
  const n = points.length;

  const px = (i: number) => padX + (i / (n - 1)) * (w - padX * 2);
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
    const base = py(0) + innerH / 10;
    return `${line} L ${px(n - 1)},${base} L ${px(0)},${base} Z`;
  };

  const totals = points.map((p) => p.up + p.down);
  const ups = points.map((p) => p.up);

  // Y-axis labels
  const yLabels = [max, max * 0.5, 0].map((v) => ({
    y: py(v),
    label: formatBytes(v, false),
  }));

  // X-axis time ticks (first, mid, last)
  const xTicks = [0, Math.floor(n / 2), n - 1].map((i) => ({
    x: px(i),
    label: points[i]?.time ? new Date(points[i].time).toLocaleTimeString([], { hour: "2-digit", minute: "2-digit" }) : "",
  }));

  const peak = Math.max(...totals);
  const total = totals.reduce((a, b) => a + b, 0);

  return (
    <div className="space-y-3">
      <svg viewBox={`0 0 ${w} ${h}`} className="w-full" role="img" aria-label="Fleet-wide traffic over time">
        <defs>
          <linearGradient id="ts-down-g" x1="0" y1="0" x2="0" y2="1">
            <stop offset="0%" stopColor="#7C6EF0" stopOpacity="0.5" />
            <stop offset="85%" stopColor="#7C6EF0" stopOpacity="0.04" />
          </linearGradient>
          <linearGradient id="ts-up-g" x1="0" y1="0" x2="0" y2="1">
            <stop offset="0%" stopColor="#34D399" stopOpacity="0.55" />
            <stop offset="85%" stopColor="#34D399" stopOpacity="0.04" />
          </linearGradient>
        </defs>

        {/* Grid lines */}
        {yLabels.map(({ y }, i) => (
          <line key={i} x1={padX} y1={y} x2={w - padX} y2={y} stroke="currentColor" strokeOpacity="0.07" strokeDasharray="4 4" />
        ))}

        {/* Download area (total) */}
        <path d={area(totals)} fill="url(#ts-down-g)" />
        <path d={smoothPath(totals)} fill="none" stroke="#7C6EF0" strokeWidth="1.5" strokeOpacity="0.8" />

        {/* Upload area */}
        <path d={area(ups)} fill="url(#ts-up-g)" />
        <path d={smoothPath(ups)} fill="none" stroke="#34D399" strokeWidth="1.5" strokeOpacity="0.9" />

        {/* Baseline */}
        <line x1={padX} y1={h - padBottom} x2={w - padX} y2={h - padBottom} stroke="currentColor" strokeOpacity="0.1" />

        {/* Y-axis labels */}
        {yLabels.slice(0, 2).map(({ y, label }, i) => (
          <text key={i} x={padX} y={y - 3} className="text-[8px] fill-current opacity-30" style={{ fontSize: 8 }}>
            {label}
          </text>
        ))}

        {/* X-axis ticks */}
        {xTicks.map(({ x, label }, i) => (
          <text
            key={i}
            x={x}
            y={h - 5}
            textAnchor={i === 0 ? "start" : i === xTicks.length - 1 ? "end" : "middle"}
            className="fill-current opacity-25"
            style={{ fontSize: 8 }}
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
