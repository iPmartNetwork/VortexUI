import { formatBytes } from "@/lib/utils";

export interface TrafficSeriesPoint {
  time: string;
  up: number;
  down: number;
}

// TrafficSeriesChart renders a dependency-free stacked area chart of fleet-wide
// throughput over time (download stacked on upload), one band per bucket.
export function TrafficSeriesChart({ points }: { points: TrafficSeriesPoint[] }) {
  if (points.length < 2) {
    return <div className="flex h-32 items-center justify-center text-xs text-fg-subtle">Collecting data…</div>;
  }

  const w = 560;
  const h = 130;
  const padX = 6;
  const padTop = 10;
  const padBottom = 18;
  const max = Math.max(1, ...points.map((p) => p.up + p.down));
  const n = points.length;
  const x = (i: number) => padX + (i / (n - 1)) * (w - padX * 2);
  const y = (v: number) => padTop + (1 - v / max) * (h - padTop - padBottom);

  // Build stacked area paths: download (total) behind, upload in front.
  const area = (vals: number[]) => {
    const top = vals.map((v, i) => `${x(i)},${y(v)}`).join(" L ");
    const base = h - padBottom;
    return `M ${x(0)},${base} L ${top} L ${x(n - 1)},${base} Z`;
  };
  const totals = points.map((p) => p.up + p.down);
  const ups = points.map((p) => p.up);

  const peak = Math.max(...totals);
  const total = totals.reduce((a, b) => a + b, 0);

  return (
    <div className="space-y-2">
      <svg viewBox={`0 0 ${w} ${h}`} className="w-full" role="img" aria-label="Fleet-wide traffic over time">
        <defs>
          <linearGradient id="ts-down" x1="0" y1="0" x2="0" y2="1">
            <stop offset="0%" stopColor="#7F77DD" stopOpacity="0.55" />
            <stop offset="100%" stopColor="#7F77DD" stopOpacity="0.04" />
          </linearGradient>
          <linearGradient id="ts-up" x1="0" y1="0" x2="0" y2="1">
            <stop offset="0%" stopColor="#5DCAA5" stopOpacity="0.6" />
            <stop offset="100%" stopColor="#5DCAA5" stopOpacity="0.05" />
          </linearGradient>
        </defs>
        <path d={area(totals)} fill="url(#ts-down)" stroke="#7F77DD" strokeWidth="1" strokeOpacity="0.7" />
        <path d={area(ups)} fill="url(#ts-up)" stroke="#5DCAA5" strokeWidth="1" strokeOpacity="0.8" />
        <line x1={padX} y1={h - padBottom} x2={w - padX} y2={h - padBottom} stroke="currentColor" strokeOpacity="0.12" />
      </svg>
      <div className="flex items-center justify-between text-xs text-fg-muted">
        <span className="flex items-center gap-3">
          <span className="flex items-center gap-1"><span className="inline-block h-2 w-2 rounded-sm" style={{ background: "#5DCAA5" }} /> Up</span>
          <span className="flex items-center gap-1"><span className="inline-block h-2 w-2 rounded-sm" style={{ background: "#7F77DD" }} /> Down</span>
        </span>
        <span>peak {formatBytes(peak, false)}/min · total {formatBytes(total, false)}</span>
      </div>
    </div>
  );
}
