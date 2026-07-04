import { cn } from "@/lib/utils";

interface GlassCardProps {
  children: React.ReactNode;
  className?: string;
  hover?: boolean;
  onClick?: () => void;
  /** Tighter padding for dense dashboard widgets. */
  compact?: boolean;
}

export function GlassCard({ children, className, hover = false, onClick, compact = false }: GlassCardProps) {
  return (
    <div
      onClick={onClick}
      className={cn(
        "rounded-2xl bg-bg-elevated border border-border transition-all duration-200",
        compact ? "p-4" : "p-5",
        hover && "hover:border-primary/20 hover:shadow-md hover:-translate-y-px cursor-pointer",
        onClick && "cursor-pointer",
        className,
      )}
    >
      {children}
    </div>
  );
}
