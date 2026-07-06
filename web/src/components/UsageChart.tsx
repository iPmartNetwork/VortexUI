import type { UsagePoint } from "@/api/hooks";
import { formatBytes } from "@/lib/utils";

function fmtDay(iso: string): string {
  const d = new Date(iso);
  if (Number.isNaN(d.getTime())) return "";
  return d.toLocaleDateString([], { weekday: "short", day: "numeric" });
}

// Responsive stacked bar chart — bar width capped so full-width layouts stay proportional.
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
    return <p className="py-8 text-center text-sm text-fg-muted">{emptyText}</p>;
  }

  const max = Math.max(1, ...points.map((p) => p.up + p.down));

  return (
    <div className="space-y-3">
      <div
        className="grid h-[128px] items-end gap-1.5 border-b border-border/50 pb-2 sm:gap-2"
        style={{ gridTemplateColumns: `repeat(${points.length}, minmax(0, 1fr))` }}
      >
        {points.map((p, i) => {
          const total = p.up + p.down;
          const downPct = total > 0 ? Math.max((p.down / max) * 100, 0.5) : 0;
          const upPct = total > 0 ? Math.max((p.up / max) * 100, 0.5) : 0;
          return (
            <div key={p.time ?? i} className="flex h-full min-w-0 flex-col items-center justify-end">
              <div
                className="flex h-full w-full max-w-9 flex-col justify-end sm:max-w-10"
                title={`${formatBytes(total)} (${upLabel} ${formatBytes(p.up)}, ${downLabel} ${formatBytes(p.down)})`}
              >
                {total > 0 ? (
                  <>
                    <div
                      className="w-full bg-[#7F77DD]"
                      style={{
                        height: `${downPct}%`,
                        minHeight: p.down > 0 ? 2 : 0,
                        borderBottomLeftRadius: p.up > 0 ? 0 : 3,
                        borderBottomRightRadius: p.up > 0 ? 0 : 3,
                      }}
                    />
                    <div
                      className="w-full rounded-t-[3px] bg-[#5DCAA5]"
                      style={{ height: `${upPct}%`, minHeight: p.up > 0 ? 2 : 0 }}
                    />
                  </>
                ) : (
                  <div className="h-1 w-full rounded-full bg-surface-3/80" />
                )}
              </div>
            </div>
          );
        })}
      </div>

      <div
        className="grid gap-1.5 sm:gap-2"
        style={{ gridTemplateColumns: `repeat(${points.length}, minmax(0, 1fr))` }}
      >
        {points.map((p, i) => (
          <div key={p.time ?? i} className="min-w-0 text-center">
            <span className="block truncate text-[10px] font-medium text-fg-subtle sm:text-[11px]">
              {fmtDay(p.time)}
            </span>
          </div>
        ))}
      </div>

      <div className="flex flex-wrap items-center justify-between gap-2 text-xs text-fg-muted">
        <span className="flex items-center gap-3">
          <span className="flex items-center gap-1.5">
            <span className="inline-block h-2 w-2 rounded-sm bg-[#5DCAA5]" /> {upLabel}
          </span>
          <span className="flex items-center gap-1.5">
            <span className="inline-block h-2 w-2 rounded-sm bg-[#7F77DD]" /> {downLabel}
          </span>
        </span>
        <span className="tabular-nums">
          {peakLabel} {formatBytes(max)}/day
        </span>
      </div>
    </div>
  );
}
