// Minimal shadcn-style UI primitives (hand-rolled to avoid the generator/runtime).
import { cn } from "@/lib/utils";

export function Button({
  className,
  variant = "primary",
  ...props
}: React.ButtonHTMLAttributes<HTMLButtonElement> & { variant?: "primary" | "ghost" | "destructive" }) {
  const variants = {
    primary: "bg-primary text-primary-foreground hover:opacity-90",
    ghost: "bg-transparent hover:bg-muted",
    destructive: "bg-destructive text-white hover:opacity-90",
  };
  return (
    <button
      className={cn(
        "inline-flex items-center justify-center rounded-md px-4 py-2 text-sm font-medium transition disabled:opacity-50",
        variants[variant],
        className,
      )}
      {...props}
    />
  );
}

export function Input({ className, ...props }: React.InputHTMLAttributes<HTMLInputElement>) {
  return (
    <input
      className={cn(
        "w-full rounded-md border bg-background px-3 py-2 text-sm outline-none focus:ring-2 focus:ring-primary/40",
        className,
      )}
      {...props}
    />
  );
}

export function Card({ className, ...props }: React.HTMLAttributes<HTMLDivElement>) {
  return <div className={cn("rounded-xl border bg-card p-5 shadow-sm", className)} {...props} />;
}

export function Select({ className, ...props }: React.SelectHTMLAttributes<HTMLSelectElement>) {
  return (
    <select
      className={cn(
        "w-full rounded-md border bg-background px-3 py-2 text-sm outline-none focus:ring-2 focus:ring-primary/40",
        className,
      )}
      {...props}
    />
  );
}

export function Badge({ children, color = "muted" }: { children: React.ReactNode; color?: string }) {
  const colors: Record<string, string> = {
    active: "bg-green-500/15 text-green-400",
    limited: "bg-amber-500/15 text-amber-400",
    expired: "bg-red-500/15 text-red-400",
    disabled: "bg-zinc-500/15 text-zinc-400",
    on_hold: "bg-blue-500/15 text-blue-400",
    muted: "bg-muted text-muted-foreground",
  };
  return (
    <span className={cn("rounded-full px-2 py-0.5 text-xs font-medium", colors[color] ?? colors.muted)}>
      {children}
    </span>
  );
}
