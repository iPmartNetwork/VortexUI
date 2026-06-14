import { cn } from "@/lib/utils";

export function Button({
  className,
  variant = "primary",
  size = "md",
  ...props
}: React.ButtonHTMLAttributes<HTMLButtonElement> & {
  variant?: "primary" | "ghost" | "outline" | "destructive";
  size?: "sm" | "md";
}) {
  const variants = {
    primary: "grad-bg text-white shadow-lg shadow-primary/25 hover:shadow-primary/40 hover:brightness-110",
    outline: "border border-border-strong bg-white/[0.03] hover:bg-white/[0.07] text-fg",
    ghost: "bg-transparent hover:bg-white/[0.06] text-fg-muted hover:text-fg",
    destructive: "bg-danger/90 text-white hover:bg-danger shadow-lg shadow-danger/20",
  };
  const sizes = { sm: "h-8 px-3 text-xs", md: "h-9 px-4 text-sm" };
  return (
    <button
      className={cn(
        "inline-flex items-center justify-center gap-2 rounded-xl font-medium transition-all active:scale-[0.98] disabled:pointer-events-none disabled:opacity-50 focus-visible:ring-4 focus-visible:ring-primary/25 outline-none",
        variants[variant],
        sizes[size],
        className,
      )}
      {...props}
    />
  );
}

export function Input({ className, ...props }: React.InputHTMLAttributes<HTMLInputElement>) {
  return <input className={cn("field", className)} {...props} />;
}

export function Select({ className, ...props }: React.SelectHTMLAttributes<HTMLSelectElement>) {
  return <select className={cn("field cursor-pointer pe-8", className)} {...props} />;
}

export function Card({ className, ...props }: React.HTMLAttributes<HTMLDivElement>) {
  return <div className={cn("card p-5", className)} {...props} />;
}

const badgeColors: Record<string, string> = {
  active: "bg-success/15 text-success ring-success/25",
  running: "bg-success/15 text-success ring-success/25",
  limited: "bg-warning/15 text-warning ring-warning/25",
  expired: "bg-danger/15 text-danger ring-danger/25",
  disabled: "bg-fg-subtle/15 text-fg-muted ring-fg-subtle/20",
  down: "bg-danger/15 text-danger ring-danger/25",
  on_hold: "bg-accent/15 text-accent ring-accent/25",
  muted: "bg-white/[0.06] text-fg-muted ring-white/10",
};

export function Badge({ children, color = "muted" }: { children: React.ReactNode; color?: string }) {
  return (
    <span
      className={cn(
        "inline-flex items-center rounded-full px-2.5 py-0.5 text-xs font-medium ring-1 ring-inset",
        badgeColors[color] ?? badgeColors.muted,
      )}
    >
      {children}
    </span>
  );
}

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
    <div className="mb-7 flex flex-wrap items-end justify-between gap-4">
      <div>
        <h1 className="text-2xl font-semibold tracking-tight text-fg">{title}</h1>
        {subtitle && <p className="mt-1 text-sm text-fg-muted">{subtitle}</p>}
      </div>
      {children && <div className="flex items-center gap-2">{children}</div>}
    </div>
  );
}

// StatCard — a glass tile with a colored gradient number, for dashboards.
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
  return (
    <div className="card relative overflow-hidden p-4">
      <div className="absolute inset-x-0 top-0 h-px grad-bg opacity-60" />
      <div className="flex items-start justify-between">
        <div>
          <div className="text-xs text-fg-muted">{label}</div>
          <div className={cn("mt-1 text-2xl font-bold tracking-tight", valueClass)}>{value}</div>
        </div>
        {icon && <div className="text-fg-subtle">{icon}</div>}
      </div>
    </div>
  );
}
