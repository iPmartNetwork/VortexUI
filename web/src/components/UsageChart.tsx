import type { UsagePoint } from "@/api/hooks";
import { formatBytes } from "@/lib/utils";
import { cn } from "@/lib/utils";

function fmtDay(iso: string, compact: boolean): string {
  const d = new Date(iso);
  if (Number.isNaN(d.getTime())) return "";
  if (compact) return d.toLocaleDateString([], { day: "numeric" });
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
  const compact = points.length > 14;
  const barMaxW = compact ? "max-w-4 sm:max-w-5" : "max-w-9 sm:max-w-10";
  const gridGap = compact ? "gap-0.5 sm:gap-1" : "gap-1.5 sm:gap-2";

  return (
    <div className="space-y-3">
      <div
        className={cn("grid h-[128px] items-end border-b border-border/50 pb-2", gridGap)}
        style={{ gridTemplateColumns: `repeat(${points.length}, minmax(0, 1fr))` }}
      >
        {points.map((p, i) => {
          const total = p.up + p.down;
          const downPct = total > 0 ? Math.max((p.down / max) * 100, 0.5) : 0;
          const upPct = total > 0 ? Math.max((p.up / max) * 100, 0.5) : 0;
          return (
            <div key={p.time ?? i} className="flex h-full min-w-0 flex-col items-center justify-end">
              <div
                className={cn("flex h-full w-full flex-col justify-end", barMaxW)}
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
        className={cn("grid", gridGap)}
        style={{ gridTemplateColumns: `repeat(${points.length}, minmax(0, 1fr))` }}
      >
        {points.map((p, i) => (
          <div key={p.time ?? i} className="min-w-0 text-center">
            <span
              className={cn(
                "block truncate font-medium text-fg-subtle",
                compact ? "text-[9px] sm:text-[10px]" : "text-[10px] sm:text-[11px]",
              )}
            >
              {fmtDay(p.time, compact)}
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
