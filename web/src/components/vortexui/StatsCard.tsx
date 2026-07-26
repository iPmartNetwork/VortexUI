import { motion } from "framer-motion";
import { TrendingDown, TrendingUp } from "lucide-react";
import { cn } from "@/lib/utils";
import { useId, useEffect, useState } from "react";

// ── Mini sparkline (inline SVG, lightweight) ──────────────────
// The polyline draws itself using stroke-dasharray offset with CSS transition.
// Path length is estimated mathematically (no DOM dependency).
export function MiniSparkline({ data, color = "hsl(var(--primary))", height = 28, delay = 0 }: {
  data: number[];
  color?: string;
  height?: number;
  /** Stagger delay in seconds before the line draws */
  delay?: number;
}) {
  const uid = useId();
  const gradId = `spk-${uid.replace(/[^a-zA-Z0-9_-]/g, "")}`;
  const [ready, setReady] = useState(delay <= 0);

  // Delayed reveal
  useEffect(() => {
    if (delay <= 0) return;
    const t = setTimeout(() => setReady(true), delay * 1000);
    return () => clearTimeout(t);
  }, [delay]);

  if (!data.length || data.every(v => v === 0)) return <div style={{ height }} />;

  const max = Math.max(...data, 1);
  const w = data.length * 6;
  const pts = data.map((v, i) => `${i * 6 + 3},${height - (v / max) * (height - 2)}`).join(" ");

  // Estimate path length ≈ diagonal across the bounding box × number of segments
  const dx = 6; // horizontal distance per point
  const dy = height;
  const estLen = data.length * Math.sqrt(dx * dx + dy * dy) * 0.5 + w;

  return (
    <div
      style={{
        opacity: ready ? 1 : 0,
        transition: `opacity 0.5s ease-out ${delay}s`,
      }}
    >
      <svg width="100%" viewBox={`0 0 ${w} ${height}`} preserveAspectRatio="none" style={{ height }}>
        <defs>
          <linearGradient id={gradId} x1="0" y1="0" x2="0" y2="1">
            <stop offset="0%" stopColor={color} stopOpacity={0.45} />
            <stop offset="100%" stopColor={color} stopOpacity={0.04} />
          </linearGradient>
        </defs>
        {/* Fill area — fades in with the container */}
        <polyline
          fill={`url(#${gradId})`}
          stroke="none"
          points={`0,${height} ${pts} ${w},${height}`}
        />
        {/* Line — draws left-to-right using estimated stroke-dashoffset */}
        <polyline
          fill="none"
          stroke={color}
          strokeWidth={1.5}
          strokeLinecap="round"
          strokeLinejoin="round"
          points={pts}
          strokeDasharray={ready ? estLen : 0}
          strokeDashoffset={ready ? 0 : estLen}
          style={{
            transition: ready
              ? `stroke-dashoffset 0.8s cubic-bezier(0.16, 1, 0.3, 1) 0.15s, opacity 0.4s ease-out`
              : "none",
            opacity: ready ? 1 : 0,
          }}
        />
      </svg>
    </div>
  );
}

// ── Live pulse dot (animated ping) ────────────────────────────
function LivePulse() {
  return (
    <span className="relative inline-flex h-1.5 w-1.5 ml-1">
      <span className="absolute inline-flex h-full w-full animate-ping rounded-full bg-green-400 opacity-75" />
      <span className="relative inline-flex h-1.5 w-1.5 rounded-full bg-green-500" />
    </span>
  );
}

interface StatsCardProps {
  title: string;
  value: React.ReactNode;
  change?: number;
  icon: React.ReactNode;
  color?: "blue" | "green" | "orange" | "red" | "purple" | "cyan";
  delay?: number;
  suffix?: string;
  subLabel?: string;
  onClick?: () => void;
  /** Sparkline data points (e.g. last 30 traffic samples) */
  sparkline?: number[];
  /** Sparkline color override */
  sparkColor?: string;
  /** Show a live progress bar at the bottom (0-100) */
  progress?: number;
  /** Progress bar color variant */
  progressColor?: "success" | "warning" | "danger" | "primary";
  /** Show animated live pulse dot */
  live?: boolean;
  /** Loading skeleton state */
  loading?: boolean;
  /** Gate animation — only enters when inView is true (scroll-triggered) */
  inView?: boolean;
}

const SPARK_COLORS: Record<string, string> = {
  blue: "#60A5FA",
  green: "#34D399",
  orange: "#FB923C",
  red: "#F87171",
  purple: "#A78BFA",
  cyan: "#22D3EE",
};

const colorMap: Record<
  string,
  {
    iconBg: string;
    iconText: string;
    glowBorder: string;
    gradient: string;
    sparkColor: string;
  }
> = {
  blue: {
    iconBg: "bg-blue-500/10",
    iconText: "text-blue-400",
    glowBorder: "hover:border-blue-500/30",
    gradient: "from-blue-500/5 to-transparent",
    sparkColor: SPARK_COLORS.blue,
  },
  green: {
    iconBg: "bg-emerald-500/10",
    iconText: "text-emerald-400",
    glowBorder: "hover:border-emerald-500/30",
    gradient: "from-green-500/5 to-transparent",
    sparkColor: SPARK_COLORS.green,
  },
  orange: {
    iconBg: "bg-orange-500/10",
    iconText: "text-orange-400",
    glowBorder: "hover:border-orange-500/30",
    gradient: "from-orange-500/5 to-transparent",
    sparkColor: SPARK_COLORS.orange,
  },
  red: {
    iconBg: "bg-red-500/10",
    iconText: "text-red-400",
    glowBorder: "hover:border-red-500/30",
    gradient: "from-red-500/5 to-transparent",
    sparkColor: SPARK_COLORS.red,
  },
  purple: {
    iconBg: "bg-purple-500/10",
    iconText: "text-purple-400",
    glowBorder: "hover:border-purple-500/30",
    gradient: "from-purple-500/5 to-transparent",
    sparkColor: SPARK_COLORS.purple,
  },
  cyan: {
    iconBg: "bg-cyan-500/10",
    iconText: "text-cyan-400",
    glowBorder: "hover:border-cyan-500/30",
    gradient: "from-cyan-500/5 to-transparent",
    sparkColor: SPARK_COLORS.cyan,
  },
};

export function StatsCard({
  title,
  value,
  change,
  icon,
  color = "cyan",
  delay = 0,
  suffix,
  subLabel,
  onClick,
  sparkline,
  sparkColor,
  progress,
  progressColor = "primary",
  live,
  loading,
  inView = true,
}: StatsCardProps) {
  const c = colorMap[color] ?? colorMap.cyan;
  const isPositive = change !== undefined && change >= 0;

  const progressColors = {
    success: "bg-success",
    warning: "bg-warning",
    danger: "bg-danger",
    primary: "bg-primary",
  };

  return (
    <motion.div
      initial={inView ? { opacity: 0, y: 12 } : false}
      animate={inView ? { opacity: 1, y: 0 } : { opacity: 1 }}
      transition={{
        duration: 0.4,
        delay: inView ? delay : 0,
        ease: [0.16, 1, 0.3, 1],
      }}
      onClick={onClick}
      className={cn(
        "relative rounded-2xl bg-bg-elevated border border-border p-4 md:p-5 transition-all duration-200 group overflow-hidden",
        `bg-gradient-to-br ${c.gradient}`,
        c.glowBorder,
        onClick && "cursor-pointer",
      )}
    >
      {/* Glow background orbs */}
      <div
        className={cn(
          "absolute -top-8 -end-8 h-24 w-24 rounded-full opacity-20 blur-3xl pointer-events-none transition-all duration-500 group-hover:opacity-30 group-hover:scale-150",
          c.iconBg,
        )}
      />
      <div
        className={cn(
          "absolute -bottom-8 -start-8 h-16 w-16 rounded-full opacity-10 blur-2xl pointer-events-none transition-all duration-500 group-hover:opacity-20",
          c.iconBg,
        )}
      />

      <div className="relative flex items-start justify-between gap-3">
        <div className="space-y-1.5 min-w-0 flex-1">
          <p className="text-xs font-bold text-fg-subtle uppercase tracking-wider flex items-center gap-1">
            {title}
            {live && <LivePulse />}
          </p>

          {loading ? (
            <div className="space-y-2">
              <div className="h-8 w-24 animate-pulse rounded-lg bg-surface-2/60" />
              <div className="h-3 w-32 animate-pulse rounded bg-surface-2/40" />
            </div>
          ) : (
            <>
              <div className="flex items-baseline gap-1.5 flex-wrap">
                <h3 className="text-[26px] md:text-[30px] font-black text-fg tracking-tight tabular-nums leading-none">
                  {value}
                </h3>
                {suffix && (
                  <span className={cn("text-base font-bold", c.iconText)}>
                    {suffix}
                  </span>
                )}
              </div>

              <div className="flex items-center gap-1.5 flex-wrap pt-0.5">
                {typeof change === "number" && (
                  <span
                    className={cn(
                      "inline-flex items-center gap-0.5 rounded-md px-1.5 py-0.5 text-[10px] font-bold tabular-nums",
                      isPositive
                        ? "bg-success/12 text-success"
                        : "bg-danger/12 text-danger",
                    )}
                  >
                    {isPositive ? (
                      <TrendingUp size={10} />
                    ) : (
                      <TrendingDown size={10} />
                    )}
                    {isPositive ? "+" : ""}
                    {change}%
                  </span>
                )}
                {subLabel && (
                  <span className="text-xs text-fg-subtle truncate">
                    {subLabel}
                  </span>
                )}
              </div>
            </>
          )}
        </div>

        <div
          className={cn(
            "h-9 w-9 rounded-xl flex items-center justify-center flex-shrink-0 transition-all duration-300 group-hover:scale-110 group-hover:shadow-lg",
            c.iconBg,
            c.iconText,
          )}
        >
          {icon}
        </div>
      </div>

      {/* Sparkline row — staggered after card entrance */}
      {sparkline !== undefined && !loading && inView && (
        <div className="mt-2.5 pt-2 border-t border-border/30">
          <MiniSparkline
            data={sparkline}
            color={sparkColor ?? c.sparkColor}
            delay={delay + 0.2}
          />
        </div>
      )}

      {/* Live progress bar */}
      {progress !== undefined && !loading && inView && (
        <div className="mt-2.5 pt-2 border-t border-border/30">
          <div className="flex items-center justify-between text-[9px] text-fg-muted/70 mb-1">
            <span>Bandwidth</span>
            <span className="font-bold tabular-nums">{Math.min(100, Math.max(0, progress))}%</span>
          </div>
          <div className="h-1.5 w-full rounded-full bg-surface-3/50 overflow-hidden">
            <div
              className={cn(
                "h-full rounded-full transition-all duration-700 ease-out",
                progressColors[progressColor],
              )}
              style={{ width: `${Math.min(100, Math.max(0, progress))}%` }}
            />
          </div>
        </div>
      )}
    </motion.div>
  );
}
