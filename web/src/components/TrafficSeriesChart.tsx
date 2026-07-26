import { TrafficAreaChart, type TrafficPoint } from "@/components/Charts";

export type { TrafficPoint as TrafficSeriesPoint };

/**
 * Recharts-powered traffic area chart.
 * Replaces the previous custom-SVG implementation with interactive hover
 * tooltips, auto-scaling axes, and responsive container.
 */
export function TrafficSeriesChart({ points }: { points: TrafficPoint[] }) {
  return <TrafficAreaChart points={points} />;
}

