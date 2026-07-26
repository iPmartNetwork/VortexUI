import { cn } from "@/lib/utils";

interface PageShellProps {
  children: React.ReactNode;
  className?: string;
  title?: string;
  subtitle?: string;
}

export function PageShell({
  children,
  className,
  title,
  subtitle,
}: PageShellProps) {
  return (
    <div className={cn("space-y-6 animate-page-enter", className)}>
      {title && (
        <div>
          <h1 className="text-xl md:text-2xl font-bold text-fg tracking-tight">
            {title}
          </h1>
          {subtitle && (
            <p className="text-sm text-fg-muted mt-1">{subtitle}</p>
          )}
        </div>
      )}
      {children}
    </div>
  );
}
