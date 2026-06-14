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
    primary:
      "bg-primary text-primary-fg hover:bg-primary-hover shadow-sm shadow-primary/20",
    outline: "border border-border-strong bg-transparent hover:bg-surface-2 text-fg",
    ghost: "bg-transparent hover:bg-surface-2 text-fg-muted hover:text-fg",
    destructive: "bg-danger/90 text-white hover:bg-danger",
  };
  const sizes = { sm: "h-8 px-3 text-xs", md: "h-9 px-4 text-sm" };
  return (
    <button
      className={cn(
        "inline-flex items-center justify-center gap-2 rounded-lg font-medium transition active:scale-[0.98] disabled:pointer-events-none disabled:opacity-50 focus-visible:ring-4 focus-visible:ring-primary/20 outline-none",
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
  active: "bg-success/15 text-success ring-success/20",
  running: "bg-success/15 text-success ring-success/20",
  limited: "bg-warning/15 text-warning ring-warning/20",
  expired: "bg-danger/15 text-danger ring-danger/20",
  disabled: "bg-fg-subtle/15 text-fg-muted ring-fg-subtle/20",
  down: "bg-danger/15 text-danger ring-danger/20",
  on_hold: "bg-accent/15 text-accent ring-accent/20",
  muted: "bg-surface-2 text-fg-muted ring-border",
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

// PageHeader standardizes the title + subtitle + actions row on every page.
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
    <div className="flex flex-wrap items-end justify-between gap-4">
      <div>
        <h1 className="text-2xl font-semibold tracking-tight text-fg">{title}</h1>
        {subtitle && <p className="mt-0.5 text-sm text-fg-muted">{subtitle}</p>}
      </div>
      {children && <div className="flex items-center gap-2">{children}</div>}
    </div>
  );
}
