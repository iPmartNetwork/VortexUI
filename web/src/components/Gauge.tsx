import { cn } from "@/lib/utils";

interface GaugeProps {
  value: number; // 0-100
  label: string;
  sublabel?: string;
  size?: number;
  color?: "primary" | "accent" | "success" | "warning" | "danger";
  className?: string;
}

export function Gauge({ value, label, sublabel, size = 120, color = "primary", className }: GaugeProps) {
  const clamped = Math.max(0, Math.min(100, value));
  const radius = (size - 16) / 2;
  const circumference = 2 * Math.PI * radius;
  const dashoffset = circumference * (1 - clamped / 100);

  const colors = {
    primary: "stroke-primary",
    accent: "stroke-accent",
    success: "stroke-success",
    warning: "stroke-warning",
    danger: "stroke-danger",
  };

  const glowColors = {
    primary: "drop-shadow-[0_0_6px_hsl(var(--primary)/0.4)]",
    accent: "drop-shadow-[0_0_6px_hsl(var(--accent)/0.4)]",
    success: "drop-shadow-[0_0_6px_hsl(var(--success)/0.4)]",
    warning: "drop-shadow-[0_0_6px_hsl(var(--warning)/0.4)]",
    danger: "drop-shadow-[0_0_6px_hsl(var(--danger)/0.4)]",
  };

  return (
    <div className={cn("flex flex-col items-center gap-2", className)}>
      <div className="relative" style={{ width: size, height: size }}>
        <svg
          width={size}
          height={size}
          viewBox={`0 0 ${size} ${size}`}
          className={cn("transform -rotate-90", glowColors[color])}
        >
          {/* Background track */}
          <circle
            cx={size / 2}
            cy={size / 2}
            r={radius}
            fill="none"
            stroke="hsl(var(--border))"
            strokeWidth="8"
            opacity="0.3"
          />
          {/* Value arc */}
          <circle
            cx={size / 2}
            cy={size / 2}
            r={radius}
            fill="none"
            className={colors[color]}
            strokeWidth="8"
            strokeLinecap="round"
            strokeDasharray={circumference}
            strokeDashoffset={dashoffset}
            style={{ transition: "stroke-dashoffset 0.8s cubic-bezier(0.4, 0, 0.2, 1)" }}
          />
        </svg>
        {/* Center text */}
        <div className="absolute inset-0 flex flex-col items-center justify-center">
          <span className="text-xl font-bold text-fg">{Math.round(clamped)}%</span>
        </div>
      </div>
      <div className="text-center">
        <div className="text-xs font-medium text-fg-muted">{label}</div>
        {sublabel && <div className="text-[10px] text-fg-subtle">{sublabel}</div>}
      </div>
    </div>
  );
}
