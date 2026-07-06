import type { UsagePoint } from "@/api/hooks";
import { formatBytes } from "@/lib/utils";

// A dependency-free stacked bar chart (download on top of upload) per bucket.
export function UsageChart({
  points,
  labels,
}: {
  points: UsagePoint[];
  labels?: { empty?: string; up?: string; down?: string; peak?: string };
}) {
  const emptyText = labels?.empty ?? "No traffic recorded yet.";
  const upLabel = labels?.up ?? "Up";
  const downLabel = labels?.down ?? "Down";
  const peakLabel = labels?.peak ?? "peak";

  if (points.length === 0) {
    return <p className="py-8 text-center text-sm text-muted-foreground">{emptyText}</p>;
  }

  const w = 440;
  const h = 160;
  const pad = 8;
  const max = Math.max(1, ...points.map((p) => p.up + p.down));
  const bw = (w - pad * 2) / points.length;
  const scale = (v: number) => (v / max) * (h - 20);

  return (
    <div className="space-y-2">
      <svg viewBox={`0 0 ${w} ${h}`} className="w-full" role="img" aria-label="Traffic usage by day">
        {points.map((p, i) => {
          const x = pad + i * bw + bw * 0.15;
          const bar = bw * 0.7;
          const upH = scale(p.up);
          const downH = scale(p.down);
          const base = h - 16;
          return (
            <g key={i}>
              <rect x={x} y={base - downH} width={bar} height={downH} rx={2} fill="#7F77DD" />
              <rect x={x} y={base - downH - upH} width={bar} height={upH} rx={2} fill="#5DCAA5" />
            </g>
          );
        })}
        <line x1={pad} y1={h - 16} x2={w - pad} y2={h - 16} stroke="currentColor" strokeOpacity="0.15" />
      </svg>
      <div className="flex items-center justify-between text-xs text-muted-foreground">
        <span className="flex items-center gap-3">
          <span className="flex items-center gap-1">
            <span className="inline-block h-2 w-2 rounded-sm" style={{ background: "#5DCAA5" }} /> {upLabel}
          </span>
          <span className="flex items-center gap-1">
            <span className="inline-block h-2 w-2 rounded-sm" style={{ background: "#7F77DD" }} /> {downLabel}
          </span>
        </span>
        <span>{peakLabel} {formatBytes(max)}/day</span>
      </div>
    </div>
  );
}
