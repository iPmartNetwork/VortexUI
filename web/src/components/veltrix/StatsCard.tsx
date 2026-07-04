import { motion } from "framer-motion";
import { TrendingDown, TrendingUp } from "lucide-react";
import { cn } from "@/lib/utils";

interface StatsCardProps {
  title: string;
  value: string | number;
  change?: number;
  icon: React.ReactNode;
  color?: "blue" | "green" | "orange" | "red" | "purple" | "cyan";
  delay?: number;
  suffix?: string;
  subLabel?: string;
}

const colorMap: Record<string, { iconBg: string; iconText: string; glowBorder: string }> = {
  blue: { iconBg: "bg-blue-500/10", iconText: "text-blue-400", glowBorder: "hover:border-blue-500/30" },
  green: { iconBg: "bg-emerald-500/10", iconText: "text-emerald-400", glowBorder: "hover:border-emerald-500/30" },
  orange: { iconBg: "bg-orange-500/10", iconText: "text-orange-400", glowBorder: "hover:border-orange-500/30" },
  red: { iconBg: "bg-red-500/10", iconText: "text-red-400", glowBorder: "hover:border-red-500/30" },
  purple: { iconBg: "bg-purple-500/10", iconText: "text-purple-400", glowBorder: "hover:border-purple-500/30" },
  cyan: { iconBg: "bg-cyan-500/10", iconText: "text-cyan-400", glowBorder: "hover:border-cyan-500/30" },
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
}: StatsCardProps) {
  const c = colorMap[color] ?? colorMap.cyan;
  const isPositive = change !== undefined && change >= 0;

  return (
    <motion.div
      initial={{ opacity: 0, y: 10 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{ duration: 0.35, delay, ease: [0.16, 1, 0.3, 1] }}
      className={cn(
        "relative rounded-2xl bg-bg-elevated border border-border p-5 transition-all duration-200 group hover:shadow-md hover:-translate-y-px",
        c.glowBorder,
      )}
    >
      <div className="flex items-start justify-between">
        <div className="space-y-1.5">
          <p className="text-[11px] font-semibold text-fg-subtle uppercase tracking-wider">{title}</p>
          <div className="flex items-baseline gap-1.5">
            <h3 className="text-2xl font-bold text-fg tracking-tight tabular-nums">{value}</h3>
            {suffix && <span className={cn("text-sm font-semibold", c.iconText)}>{suffix}</span>}
          </div>
          <div className="flex items-center gap-2 pt-0.5">
            {typeof change === "number" && (
              <span
                className={cn(
                  "inline-flex items-center gap-0.5 rounded-md px-1.5 py-0.5 text-[10px] font-semibold",
                  isPositive ? "bg-success/10 text-success" : "bg-danger/10 text-danger",
                )}
              >
                {isPositive ? <TrendingUp size={10} /> : <TrendingDown size={10} />}
                {isPositive ? "+" : ""}
                {change}%
              </span>
            )}
            {subLabel && <span className="text-[11px] text-fg-subtle truncate">{subLabel}</span>}
          </div>
        </div>
        <div
          className={cn(
            "h-10 w-10 rounded-xl flex items-center justify-center flex-shrink-0 transition-transform duration-200 group-hover:scale-105",
            c.iconBg,
            c.iconText,
          )}
        >
          {icon}
        </div>
      </div>
    </motion.div>
  );
}
