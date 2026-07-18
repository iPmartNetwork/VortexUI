import { motion } from "framer-motion";
import { TrendingDown, TrendingUp } from "lucide-react";
import { cn } from "@/lib/utils";

interface StatsCardProps {
  title: string;
  value: React.ReactNode;
  change?: number;
  icon: React.ReactNode;
  color?: "blue" | "green" | "orange" | "red" | "purple" | "cyan";
  delay?: number;
  suffix?: string;
  subLabel?: string;
}

const colorMap: Record<string, { iconBg: string; iconText: string; glowBorder: string; changeBg: string; changeText: string; gradient: string }> = {
  blue:   { iconBg: "bg-blue-500/10",    iconText: "text-blue-400",    glowBorder: "hover:border-blue-500/30",    changeBg: "bg-blue-500/12",   changeText: "text-blue-400",    gradient: "from-blue-500/5 to-transparent"   },
  green:  { iconBg: "bg-emerald-500/10", iconText: "text-emerald-400", glowBorder: "hover:border-emerald-500/30", changeBg: "bg-emerald-500/12", changeText: "text-emerald-400", gradient: "from-green-500/5 to-transparent"   },
  orange: { iconBg: "bg-orange-500/10",  iconText: "text-orange-400",  glowBorder: "hover:border-orange-500/30",  changeBg: "bg-orange-500/12",  changeText: "text-orange-400",  gradient: "from-orange-500/5 to-transparent"  },
  red:    { iconBg: "bg-red-500/10",     iconText: "text-red-400",     glowBorder: "hover:border-red-500/30",     changeBg: "bg-red-500/12",     changeText: "text-red-400",     gradient: "from-red-500/5 to-transparent"     },
  purple: { iconBg: "bg-purple-500/10",  iconText: "text-purple-400",  glowBorder: "hover:border-purple-500/30",  changeBg: "bg-purple-500/12",  changeText: "text-purple-400",  gradient: "from-purple-500/5 to-transparent"  },
  cyan:   { iconBg: "bg-cyan-500/10",    iconText: "text-cyan-400",    glowBorder: "hover:border-cyan-500/30",    changeBg: "bg-cyan-500/12",    changeText: "text-cyan-400",    gradient: "from-cyan-500/5 to-transparent"    },
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
      initial={{ opacity: 0, y: 12 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{ duration: 0.4, delay, ease: [0.16, 1, 0.3, 1] }}
      className={cn(
        "relative rounded-2xl bg-bg-elevated border border-border p-4 transition-all duration-200 group hover:shadow-lg hover:-translate-y-0.5 overflow-hidden",
        `bg-gradient-to-br ${c.gradient}`,
        c.glowBorder,
      )}
    >
      {/* subtle glow background */}
      <div className={cn("absolute -top-6 -end-6 h-20 w-20 rounded-full opacity-20 blur-2xl pointer-events-none", c.iconBg)} />

      <div className="relative flex items-start justify-between gap-3">
        <div className="space-y-1.5 min-w-0">
          <p className="text-[10px] font-bold text-fg-subtle uppercase tracking-wider">{title}</p>

          <div className="flex items-baseline gap-1.5 flex-wrap">
            <h3 className="text-[26px] font-black text-fg tracking-tight tabular-nums leading-none">{value}</h3>
            {suffix && (
              <span className={cn("text-base font-bold", c.iconText)}>{suffix}</span>
            )}
          </div>

          <div className="flex items-center gap-1.5 flex-wrap pt-0.5">
            {typeof change === "number" && (
              <span
                className={cn(
                  "inline-flex items-center gap-0.5 rounded-md px-1.5 py-0.5 text-[10px] font-bold",
                  isPositive
                    ? "bg-success/12 text-success"
                    : "bg-danger/12 text-danger",
                )}
              >
                {isPositive ? <TrendingUp size={10} /> : <TrendingDown size={10} />}
                {isPositive ? "+" : ""}
                {change}%
              </span>
            )}
            {subLabel && (
              <span className="text-[10px] text-fg-subtle truncate">{subLabel}</span>
            )}
          </div>
        </div>

        <div
          className={cn(
            "h-9 w-9 rounded-lg flex items-center justify-center flex-shrink-0 transition-all duration-300 group-hover:scale-110 group-hover:shadow-lg",
            c.iconBg,
            c.iconText,
            `group-hover:shadow-${color}-500/30`,
          )}
        >
          {icon}
        </div>
      </div>
    </motion.div>
  );
}
