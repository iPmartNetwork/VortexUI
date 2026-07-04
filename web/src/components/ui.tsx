import { cn } from "@/lib/utils";

// --- Button --- (gradient primary, soft ghost, bordered outline, danger)
export function Button({
  className,
  variant = "primary",
  size = "md",
  ...props
}: React.ButtonHTMLAttributes<HTMLButtonElement> & {
  variant?: "primary" | "ghost" | "outline" | "destructive";
  size?: "sm" | "md";
}) {
  const base =
    "inline-flex items-center justify-center gap-2 rounded-xl font-medium transition-all duration-200 active:scale-[0.97] disabled:pointer-events-none disabled:opacity-50 outline-none focus-visible:ring-[3px] focus-visible:ring-primary/20";
  const variants = {
    primary:
      "grad-bg text-primary-fg shadow-md hover:shadow-lg hover:brightness-110",
    outline:
      "border border-border/80 bg-surface/60 hover:bg-surface-2/80 text-fg backdrop-blur-sm",
    ghost: "bg-transparent hover:bg-surface-2/60 text-fg-muted hover:text-fg",
    destructive:
      "bg-danger text-white hover:bg-danger/90 shadow-lg shadow-danger/20 hover:shadow-danger/35",
  };
  const sizes = { sm: "h-8 px-3 text-xs", md: "h-9 px-4 text-sm" };
  return <button className={cn(base, variants[variant], sizes[size], className)} {...props} />;
}

// --- Input --- (inset field with glow focus ring)
export function Input({ className, ...props }: React.InputHTMLAttributes<HTMLInputElement>) {
  return <input className={cn("field input-surface", className)} {...props} />;
}

// --- Select ---
export function Select({ className, ...props }: React.SelectHTMLAttributes<HTMLSelectElement>) {
  return (
    <select className={cn("field input-surface cursor-pointer pe-8 appearance-none", className)} {...props} />
  );
}

// --- Card --- (Veltrix elevated surface — used across all admin pages)
export function Card({ className, ...props }: React.HTMLAttributes<HTMLDivElement>) {
  return (
    <div
      className={cn(
        "rounded-2xl bg-bg-elevated border border-border p-5 transition-all duration-200",
        className,
      )}
      {...props}
    />
  );
}

// --- Badge --- (pill with status color + ring)
const badgeColors: Record<string, string> = {
  active: "bg-success/15 text-success border-success/30",
  running: "bg-success/15 text-success border-success/30",
  limited: "bg-warning/15 text-warning border-warning/30",
  expired: "bg-danger/15 text-danger border-danger/30",
  disabled: "bg-fg-subtle/10 text-fg-subtle border-fg-subtle/20",
  down: "bg-danger/15 text-danger border-danger/30",
  on_hold: "bg-primary/15 text-primary border-primary/30",
  muted: "bg-surface-2/80 text-fg-muted border-border/40",
};

export function Badge({ children, color = "muted" }: { children: React.ReactNode; color?: string }) {
  return (
    <span
      className={cn(
        "inline-flex items-center rounded-full border px-2.5 py-0.5 text-[11px] font-semibold uppercase tracking-wide",
        badgeColors[color] ?? badgeColors.muted,
      )}
    >
      {children}
    </span>
  );
}

// --- Page Header --- (Veltrix page title bar)
export function PageHeader({
  title,
  subtitle,
  children,
}: {
  title: string;
  subtitle?: string;
  children?: React.ReactNode;
}) {
  return (
    <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
      <div>
        <h1 className="text-xl font-bold text-fg tracking-tight">{title}</h1>
        {subtitle && <p className="text-sm text-fg-muted mt-0.5">{subtitle}</p>}
      </div>
      {children && <div className="flex flex-wrap items-center gap-2">{children}</div>}
    </div>
  );
}

// --- StatCard --- (metric tile — live values only, no mock deltas)
export function StatCard({
  label,
  value,
  accent = "grad",
  icon,
  sub,
}: {
  label: string;
  value: React.ReactNode;
  accent?: "grad" | "success" | "accent" | "plain" | "warning";
  icon?: React.ReactNode;
  sub?: string;
}) {
  const valueClass = {
    grad: "grad-text",
    success: "text-success",
    accent: "text-accent",
    plain: "text-fg",
    warning: "text-warning",
  }[accent];

  const iconBox = {
    grad: "bg-primary/10 text-primary",
    success: "bg-success/10 text-success",
    accent: "bg-accent/10 text-accent",
    plain: "bg-surface-2 text-fg-muted",
    warning: "bg-warning/10 text-warning",
  }[accent];

  return (
    <div className="rounded-2xl bg-bg-elevated border border-border p-5 transition-all duration-200 hover:shadow-md hover:-translate-y-px">
      <div className="flex items-start justify-between gap-3">
        <div className="min-w-0">
          <div className="text-[11px] font-semibold text-fg-subtle uppercase tracking-wider">{label}</div>
          <div className={cn("mt-1.5 text-2xl font-bold tracking-tight tabular-nums", valueClass)}>{value}</div>
          {sub && <div className="mt-1 text-[11px] text-fg-muted truncate">{sub}</div>}
        </div>
        {icon && (
          <div className={cn("h-10 w-10 rounded-xl flex items-center justify-center flex-shrink-0", iconBox)}>
            {icon}
          </div>
        )}
      </div>
    </div>
  );
}
