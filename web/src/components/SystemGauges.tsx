import { useSystem, useOverview } from "@/api/policy-hooks";
import { Gauge } from "@/components/Gauge";
import { GlassCard } from "@/components/vortexui";
import { motion, type Variants } from "framer-motion";

const container: Variants = {
  hidden: { opacity: 0 },
  show: {
    opacity: 1,
    transition: { staggerChildren: 0.08 },
  },
};

const item: Variants = {
  hidden: { opacity: 0, y: 12 },
  show: { opacity: 1, y: 0, transition: { duration: 0.4, ease: [0.16, 1, 0.3, 1] as const } },
};

export function SystemGauges() {
  const { data: sys } = useSystem();
  const { data: overview } = useOverview();

  const cpu = sys?.cpu_percent ?? 0;
  const mem = sys?.mem_percent ?? 0;
  const disk = sys?.disk_percent ?? 0;

  const fleetItems = overview?.nodes.items ?? [];
  const totalConnections = fleetItems.reduce((sum, n) => sum + (n.health?.connections ?? 0), 0);
  const cpuLimit = 80;
  const memoLimit = 85;
  const diskLimit = 90;

  type GaugeColor = "primary" | "accent" | "success" | "warning" | "danger";

  const gauges: { value: number; label: string; sublabel: string; color: GaugeColor; limit: number }[] = [
    {
      value: cpu,
      label: "CPU",
      sublabel: `${sys?.cpu_percent?.toFixed(1) ?? "—"}% · ${sys?.goroutines ?? "—"} goroutines`,
      color: cpu > cpuLimit ? "danger" : cpu > 60 ? "warning" : "primary",
      limit: cpuLimit,
    },
    {
      value: mem,
      label: "RAM",
      sublabel: `${sys?.mem_percent?.toFixed(1) ?? "—"}% · ${sys?.mem_alloc_bytes ? formatMB(sys.mem_alloc_bytes) : "—"}`,
      color: mem > memoLimit ? "danger" : mem > 70 ? "warning" : "success",
      limit: memoLimit,
    },
    {
      value: disk,
      label: "Disk",
      sublabel: `${sys?.disk_percent?.toFixed(1) ?? "—"}% used`,
      color: disk > diskLimit ? "danger" : disk > 75 ? "warning" : "accent",
      limit: diskLimit,
    },
    {
      value: Math.min(100, (totalConnections / 500) * 100),
      label: "Connections",
      sublabel: `${totalConnections} active · ${sys?.hostname ?? "—"}`,
      color: totalConnections > 400 ? "danger" : totalConnections > 200 ? "warning" : "success",
      limit: 500,
    },
  ];

  return (
    <motion.div
      variants={container}
      initial="hidden"
      animate="show"
      className="grid grid-cols-2 sm:grid-cols-4 gap-3"
    >
      {gauges.map((g) => (
        <motion.div key={g.label} variants={item}>
          <GlassCard className="!p-3 flex flex-col items-center gap-1.5 border-border/50" glow={g.value > g.limit * 0.9}>
            <Gauge
              value={g.value}
              label=""
              size={80}
              color={g.color}
            />
            <div className="text-center mt-1">
              <span className="text-[11px] font-bold text-fg">{g.label}</span>
              {g.value > g.limit && (
                <span className="ms-1.5 text-[9px] font-bold text-danger animate-pulse">
                  HIGH
                </span>
              )}
              <p className="text-[9px] text-fg-subtle/70 truncate max-w-[90px]">
                {g.sublabel}
              </p>
            </div>
            {/* Mini progress line */}
            <div className="w-full h-1 rounded-full bg-surface-3/50 overflow-hidden mt-0.5">
              <motion.div
                className="h-full rounded-full"
                style={{
                  background: g.color === "danger" ? "hsl(var(--destructive))"
                    : g.color === "warning" ? "hsl(var(--warning))"
                    : g.color === "success" ? "hsl(var(--success))"
                    : "hsl(var(--primary))",
                }}
                initial={{ width: 0 }}
                animate={{ width: `${Math.min(100, g.value)}%` }}
                transition={{ duration: 0.8, ease: "easeOut" }}
              />
            </div>
          </GlassCard>
        </motion.div>
      ))}
    </motion.div>
  );
}

function formatMB(bytes: number): string {
  const mb = bytes / (1024 * 1024);
  return mb >= 1024 ? `${(mb / 1024).toFixed(1)} GB` : `${mb.toFixed(0)} MB`;
}
