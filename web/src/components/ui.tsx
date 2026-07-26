import { cn } from "@/lib/utils";
import { motion } from "framer-motion";
import { Loader2 } from "lucide-react";

// ════════════════════════════════════════════════════════════════
// SWITCH — Toggle with LTR/RTL support
// ════════════════════════════════════════════════════════════════
export function Switch({
  checked,
  onCheckedChange,
  className,
  disabled,
  label,
  ...props
}: Omit<React.ButtonHTMLAttributes<HTMLButtonElement>, "onChange"> & {
  checked: boolean;
  onCheckedChange?: (checked: boolean) => void;
  label?: string;
}) {
  return (
    <label className="inline-flex items-center gap-2.5 cursor-pointer select-none group">
      {label && (
        <span className="text-xs font-medium text-fg-muted group-hover:text-fg transition-colors">
          {label}
        </span>
      )}
      <button
        type="button"
        role="switch"
        aria-checked={checked}
        disabled={disabled}
        onClick={() => onCheckedChange?.(!checked)}
        className={cn(
          "relative h-6 w-11 rounded-full transition-all duration-200 flex-shrink-0",
          checked ? "bg-primary" : "bg-surface-3",
          disabled && "pointer-events-none opacity-50",
          "focus-visible:ring-[3px] focus-visible:ring-primary/20",
          className,
        )}
        {...props}
      >
        <span
          aria-hidden
          className={cn(
            "absolute top-0.5 start-0.5 h-5 w-5 rounded-full bg-white shadow-md transition-all duration-200",
            checked && "translate-x-5 rtl:-translate-x-5",
            checked && "shadow-[0_0_8px_-2px_hsl(var(--primary))]",
          )}
        />
      </button>
    </label>
  );
}

// ════════════════════════════════════════════════════════════════
// BUTTON — Variants with cyber/glass styling
// ════════════════════════════════════════════════════════════════
export function Button({
  className,
  variant = "primary",
  size = "md",
  loading,
  children,
  ...props
}: React.ButtonHTMLAttributes<HTMLButtonElement> & {
  variant?: "primary" | "ghost" | "outline" | "destructive" | "success" | "glass";
  size?: "sm" | "md" | "lg";
  loading?: boolean;
}) {
  const base =
    "inline-flex items-center justify-center gap-2 rounded-xl font-medium transition-all duration-200 active:scale-[0.97] disabled:pointer-events-none disabled:opacity-50 outline-none focus-visible:ring-[3px] focus-visible:ring-primary/20";

  const variants: Record<string, string> = {
    primary:
      "grad-bg text-primary-fg shadow-md hover:shadow-lg hover:brightness-110",
    success:
      "grad-bg-2 text-white shadow-md hover:shadow-lg hover:brightness-110",
    outline:
      "border border-border/80 bg-surface/60 hover:bg-surface-2/80 text-fg backdrop-blur-sm hover:border-border-strong",
    ghost:
      "bg-transparent hover:bg-surface-2/60 text-fg-muted hover:text-fg",
    destructive:
      "bg-danger text-white hover:bg-danger/90 shadow-lg shadow-danger/20 hover:shadow-danger/35",
    glass:
      "glass text-fg hover:border-primary/30 hover:shadow-glow-sm",
  };

  const sizes: Record<string, string> = {
    sm: "h-8 px-3 text-xs",
    md: "h-9 px-4 text-sm",
    lg: "h-10 px-5 text-base",
  };

  return (
    <button
      className={cn(base, variants[variant], sizes[size], className)}
      disabled={props.disabled || loading}
      {...props}
    >
      {loading && <Loader2 size={size === "sm" ? 12 : 14} className="animate-spin" />}
      {children}
    </button>
  );
}

// ════════════════════════════════════════════════════════════════
// INPUT — Recessed field with cyber glow focus
// ════════════════════════════════════════════════════════════════
export function Input({
  className,
  icon,
  ...props
}: React.InputHTMLAttributes<HTMLInputElement> & {
  icon?: React.ReactNode;
}) {
  return (
    <div className="relative">
      {icon && (
        <div className="absolute start-3.5 top-1/2 -translate-y-1/2 text-fg-subtle pointer-events-none">
          {icon}
        </div>
      )}
      <input
        className={cn(
          "field input-surface",
          icon ? "ps-10" : undefined,
          className,
        )}
        {...props}
      />
    </div>
  );
}

// ════════════════════════════════════════════════════════════════
// TEXTAREA
// ════════════════════════════════════════════════════════════════
export function Textarea({
  className,
  ...props
}: React.TextareaHTMLAttributes<HTMLTextAreaElement>) {
  return (
    <textarea
      className={cn(
        "field input-surface min-h-[80px] resize-y",
        className,
      )}
      {...props}
    />
  );
}

// ════════════════════════════════════════════════════════════════
// SELECT
// ════════════════════════════════════════════════════════════════
export function Select({
  className,
  children,
  ...props
}: React.SelectHTMLAttributes<HTMLSelectElement>) {
  return (
    <div className="relative">
      <select
        className={cn(
          "field input-surface cursor-pointer pe-8 appearance-none",
          className,
        )}
        {...props}
      >
        {children}
      </select>
      <div className="absolute end-3 top-1/2 -translate-y-1/2 pointer-events-none text-fg-subtle">
        <svg width="10" height="6" viewBox="0 0 10 6" fill="none">
          <path d="M1 1L5 5L9 1" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round" />
        </svg>
      </div>
    </div>
  );
}

// ════════════════════════════════════════════════════════════════
// CARD — Elevated surface
// ════════════════════════════════════════════════════════════════
export function Card({
  className,
  hover = false,
  glow = false,
  ...props
}: React.HTMLAttributes<HTMLDivElement> & {
  hover?: boolean;
  glow?: boolean | string;
}) {
  return (
    <div
      className={cn(
        "rounded-2xl bg-bg-elevated border border-border p-5 transition-all duration-200",
        hover && "hover:shadow-md hover:-translate-y-0.5 cursor-pointer",
        glow && "hover:border-primary/30 hover:shadow-glow-sm",
        className,
      )}
      {...props}
    />
  );
}

// ════════════════════════════════════════════════════════════════
// BADGE — Status pill
// ════════════════════════════════════════════════════════════════
const badgeColors: Record<string, string> = {
  active: "bg-success/15 text-success border-success/30",
  running: "bg-success/15 text-success border-success/30",
  optimal: "bg-success/15 text-success border-success/30",
  limited: "bg-warning/15 text-warning border-warning/30",
  expired: "bg-danger/15 text-danger border-danger/30",
  disabled: "bg-fg-subtle/10 text-fg-subtle border-fg-subtle/20",
  down: "bg-danger/15 text-danger border-danger/30",
  on_hold: "bg-primary/15 text-primary border-primary/30",
  pending: "bg-warning/15 text-warning border-warning/30",
  info: "bg-primary/15 text-primary border-primary/30",
  warning: "bg-warning/15 text-warning border-warning/30",
  success: "bg-success/15 text-success border-success/30",
  muted: "bg-surface-2/80 text-fg-muted border-border/40",
};

export function Badge({
  children,
  color = "muted",
  pulse = false,
}: {
  children: React.ReactNode;
  color?: string;
  pulse?: boolean;
}) {
  return (
    <span
      className={cn(
        "inline-flex items-center rounded-full border px-2.5 py-0.5 text-xs font-semibold uppercase tracking-wide transition-all",
        badgeColors[color] ?? badgeColors.muted,
        pulse && "animate-pulse-glow",
      )}
    >
      {pulse && (
        <span className="relative flex h-1.5 w-1.5 me-1.5">
          <span className="absolute inline-flex h-full w-full animate-ping rounded-full bg-current opacity-75" />
          <span className="relative inline-flex h-1.5 w-1.5 rounded-full bg-current" />
        </span>
      )}
      {children}
    </span>
  );
}

// ════════════════════════════════════════════════════════════════
// PAGE HEADER — Title + subtitle + actions
// ════════════════════════════════════════════════════════════════
export function PageHeader({
  title,
  subtitle,
  badge,
  children,
}: {
  title: string;
  subtitle?: string;
  badge?: string;
  children?: React.ReactNode;
}) {
  return (
    <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
      <div className="min-w-0">
        <div className="flex items-center gap-3 flex-wrap">
          <h1 className="text-xl md:text-2xl font-bold text-fg tracking-tight">
            {title}
          </h1>
          {badge && (
            <span className="rounded-md bg-primary/10 px-2 py-0.5 text-[10px] font-bold text-primary border border-primary/20">
              {badge}
            </span>
          )}
        </div>
        {subtitle && (
          <p className="text-sm text-fg-muted mt-1 max-w-2xl">{subtitle}</p>
        )}
      </div>
      {children && (
        <div className="flex flex-wrap items-center gap-2 flex-shrink-0">
          {children}
        </div>
      )}
    </div>
  );
}

// ════════════════════════════════════════════════════════════════
// STAT CARD — Metric tile (legacy compatible, enhanced)
// ════════════════════════════════════════════════════════════════
export function StatCard({
  label,
  value,
  accent = "grad",
  icon,
  sub,
  trend,
  onClick,
}: {
  label: string;
  value: React.ReactNode;
  accent?: "grad" | "success" | "accent" | "plain" | "warning";
  icon?: React.ReactNode;
  sub?: string;
  trend?: { value: number; positive: boolean };
  onClick?: () => void;
}) {
  const valueStyles: Record<string, string> = {
    grad: "grad-text",
    success: "text-success",
    accent: "text-accent",
    plain: "text-fg",
    warning: "text-warning",
  };

  const iconBoxes: Record<string, string> = {
    grad: "bg-primary/10 text-primary",
    success: "bg-success/10 text-success",
    accent: "bg-accent/10 text-accent",
    plain: "bg-surface-2 text-fg-muted",
    warning: "bg-warning/10 text-warning",
  };

  return (
    <motion.div
      initial={{ opacity: 0, y: 12 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{ duration: 0.4, ease: [0.16, 1, 0.3, 1] }}
      onClick={onClick}
      className={cn(
        "rounded-2xl bg-bg-elevated border border-border p-5 transition-all duration-200",
        "hover:shadow-md hover:-translate-y-0.5",
        onClick && "cursor-pointer",
      )}
    >
      <div className="flex items-start justify-between gap-3">
        <div className="min-w-0 flex-1">
          <div className="text-xs font-semibold text-fg-subtle uppercase tracking-wider">
            {label}
          </div>
          <div
            className={cn(
              "mt-1.5 text-2xl font-bold tracking-tight tabular-nums",
              valueStyles[accent],
            )}
          >
            {value}
          </div>
          {sub && (
            <div className="mt-1 text-[11px] text-fg-muted truncate">{sub}</div>
          )}
          {trend && (
            <div
              className={cn(
                "mt-1.5 inline-flex items-center gap-1 text-[10px] font-semibold",
                trend.positive ? "text-success" : "text-danger",
              )}
            >
              <svg
                width="8"
                height="8"
                viewBox="0 0 8 8"
                fill="none"
                className={trend.positive ? "" : "rotate-180"}
              >
                <path
                  d="M4 1L7 6H1L4 1Z"
                  fill="currentColor"
                />
              </svg>
              {trend.positive ? "+" : ""}
              {trend.value}%
            </div>
          )}
        </div>
        {icon && (
          <div
            className={cn(
              "h-10 w-10 rounded-xl flex items-center justify-center flex-shrink-0 transition-all duration-300",
              "group-hover:scale-110 group-hover:shadow-lg",
              iconBoxes[accent],
            )}
          >
            {icon}
          </div>
        )}
      </div>
    </motion.div>
  );
}

// ════════════════════════════════════════════════════════════════
// KBD — Keyboard shortcut badge
// ════════════════════════════════════════════════════════════════
export function Kbd({
  children,
  className,
}: {
  children: React.ReactNode;
  className?: string;
}) {
  return (      <kbd
      className={cn(
        "inline-flex items-center gap-0.5 rounded-md border border-border/60 bg-surface-2/50 px-1.5 py-0.5 text-xs text-fg-subtle font-mono shadow-sm",
        className,
      )}
    >
      {children}
    </kbd>
  );
}

// ════════════════════════════════════════════════════════════════
// DIVIDER
// ════════════════════════════════════════════════════════════════
export function Divider({
  className,
  label,
}: {
  className?: string;
  label?: string;
}) {
  if (label) {
    return (
      <div className={cn("flex items-center gap-3", className)}>
        <div className="flex-1 h-px bg-gradient-to-r from-border/60 to-transparent" />
        <span className="text-xs font-semibold text-fg-subtle uppercase tracking-wider">
          {label}
        </span>
        <div className="flex-1 h-px bg-gradient-to-l from-border/60 to-transparent" />
      </div>
    );
  }
  return (
    <div
      className={cn(
        "h-px bg-gradient-to-r from-transparent via-border/60 to-transparent",
        className,
      )}
    />
  );
}

// ════════════════════════════════════════════════════════════════
// PROGRESS BAR
// ════════════════════════════════════════════════════════════════
export function ProgressBar({
  value,
  max = 100,
  color = "primary",
  size = "sm",
  showLabel = false,
  className,
}: {
  value: number;
  max?: number;
  color?: "primary" | "success" | "warning" | "danger" | "accent";
  size?: "sm" | "md";
  showLabel?: boolean;
  className?: string;
}) {
  const pct = Math.min(100, Math.max(0, (value / max) * 100));
  const colors: Record<string, string> = {
    primary: "bg-primary",
    success: "bg-success",
    warning: "bg-warning",
    danger: "bg-danger",
    accent: "bg-accent",
  };
  const heights: Record<string, string> = {
    sm: "h-1",
    md: "h-2",
  };

  return (
    <div className={cn("w-full", className)}>
      <div
        className={cn(
          "w-full rounded-full bg-surface-3 overflow-hidden",
          heights[size],
        )}
      >
        <div
          className={cn(
            heights[size],
            "rounded-full transition-all duration-500 ease-[cubic-bezier(0.16,1,0.3,1)]",
            colors[color],
          )}
          style={{ width: `${pct}%` }}
        />
      </div>
      {showLabel && (
        <span className="text-[10px] text-fg-muted mt-0.5 block text-end tabular-nums">
          {pct.toFixed(0)}%
        </span>
      )}
    </div>
  );
}
