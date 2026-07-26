import type { UsagePoint } from "@/api/hooks";
import { UsageBarChart } from "@/components/Charts";

/**
 * Recharts-powered usage bar chart.
 * Replaces the previous CSS-grid implementation with interactive tooltips.
 */
export function UsageChart({
  points,
  labels,
}: {
  points: UsagePoint[];
  labels?: { empty?: string; up?: string; down?: string; peak?: string };
}) {
  const emptyText = labels?.empty ?? "No traffic recorded yet.";
  const compact = points.length > 14;

  if (!points.length) {
    return <p className="py-8 text-center text-sm text-fg-muted">{emptyText}</p>;
  }

  return (
    <div className="space-y-3">
      <UsageBarChart points={points} compact={compact} />
      <div className="flex flex-wrap items-center justify-between gap-2 text-xs text-fg-muted">
        <span className="flex items-center gap-3">
          <span className="flex items-center gap-1.5">
            <span className="inline-block h-2 w-2 rounded-sm bg-[#7C3AED]" /> {labels?.down ?? "Down"}
          </span>
          <span className="flex items-center gap-1.5">
            <span className="inline-block h-2 w-2 rounded-sm bg-[#22D3EE]" /> {labels?.up ?? "Up"}
          </span>
        </span>
      </div>
    </div>
  );
}
