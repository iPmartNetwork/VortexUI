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
      "grad-bg text-white shadow-lg shadow-primary/20 hover:shadow-primary/35 hover:brightness-110 glow-primary/0 hover:glow-primary",
    outline:
      "border border-border-strong/80 bg-surface/60 hover:bg-surface-2/80 text-fg backdrop-blur-sm",
    ghost:
      "bg-transparent hover:bg-surface-2/60 text-fg-muted hover:text-fg",
    destructive:
      "bg-danger text-white hover:bg-danger/90 shadow-lg shadow-danger/20 hover:shadow-danger/35",
  };
  const sizes = { sm: "h-8 px-3 text-xs", md: "h-9 px-4 text-sm" };
  return <button className={cn(base, variants[variant], sizes[size], className)} {...props} />;
}

// --- Input --- (inset dark field with glow focus ring)
export function Input({ className, ...props }: React.InputHTMLAttributes<HTMLInputElement>) {
  return <input className={cn("field", className)} {...props} />;
}

// --- Select ---
export function Select({ className, ...props }: React.SelectHTMLAttributes<HTMLSelectElement>) {
  return <select className={cn("field cursor-pointer pe-8 appearance-none", className)} {...props} />;
}

// --- Card --- (glass surface)
export function Card({ className, ...props }: React.HTMLAttributes<HTMLDivElement>) {
  return <div className={cn("card p-5", className)} {...props} />;
}

// --- Badge --- (pill with status color + ring)
const badgeColors: Record<string, string> = {
  active: "bg-success/12 text-success ring-success/20",
  running: "bg-success/12 text-success ring-success/20",
  limited: "bg-warning/12 text-warning ring-warning/20",
  expired: "bg-danger/12 text-danger ring-danger/20",
  disabled: "bg-fg-subtle/12 text-fg-muted ring-fg-subtle/15",
  down: "bg-danger/12 text-danger ring-danger/20",
  on_hold: "bg-accent/12 text-accent ring-accent/20",
  muted: "bg-surface-2/80 text-fg-muted ring-border/40",
};

export function Badge({ children, color = "muted" }: { children: React.ReactNode; color?: string }) {
  return (
    <span
      className={cn(
        "inline-flex items-center rounded-full px-2.5 py-0.5 text-[11px] font-semibold uppercase tracking-wide ring-1 ring-inset",
        badgeColors[color] ?? badgeColors.muted,
      )}
    >
      {children}
    </span>
  );
}

// --- Page Header ---
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
    <div className="mb-8 flex flex-wrap items-end justify-between gap-4">
      <div>
        <h1 className="text-[1.6rem] font-bold tracking-tight text-fg">{title}</h1>
        {subtitle && <p className="mt-1 text-sm text-fg-muted">{subtitle}</p>}
      </div>
      {children && <div className="flex items-center gap-2.5">{children}</div>}
    </div>
  );
}

// --- StatCard --- (glass tile with accent number + glow border top)
export function StatCard({
  label,
  value,
  accent = "grad",
  icon,
}: {
  label: string;
  value: React.ReactNode;
  accent?: "grad" | "success" | "accent" | "plain";
  icon?: React.ReactNode;
}) {
  const valueClass = {
    grad: "grad-text",
    success: "text-success",
    accent: "text-accent",
    plain: "text-fg",
  }[accent];

  const glowClass = {
    grad: "via-primary/60",
    success: "via-success/60",
    accent: "via-accent/60",
    plain: "via-fg-subtle/30",
  }[accent];

  return (
    <div className="card relative overflow-hidden p-5">
      {/* Top glow line */}
      <div className={cn("absolute inset-x-0 top-0 h-px bg-gradient-to-r from-transparent to-transparent", glowClass)} />
      <div className="flex items-start justify-between">
        <div>
          <div className="text-xs font-medium text-fg-muted">{label}</div>
          <div className={cn("mt-1.5 text-2xl font-bold tracking-tight", valueClass)}>{value}</div>
        </div>
        {icon && <div className="rounded-lg bg-surface-2/60 p-2 text-fg-subtle">{icon}</div>}
      </div>
    </div>
  );
}
