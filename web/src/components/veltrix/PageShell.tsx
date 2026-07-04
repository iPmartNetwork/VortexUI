import { cn } from "@/lib/utils";

interface PageShellProps {
  children: React.ReactNode;
  className?: string;
}

/** Standard page wrapper — Veltrix spacing + enter animation. */
export function PageShell({ children, className }: PageShellProps) {
  return <div className={cn("space-y-6 animate-page-enter", className)}>{children}</div>;
}
