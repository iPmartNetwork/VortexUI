import { cn } from "@/lib/utils";

interface GlassCardProps {
  children: React.ReactNode;
  className?: string;
  hover?: boolean;
  onClick?: () => void;
  compact?: boolean;
  glow?: boolean | string;
}

export function GlassCard({
  children,
  className,
  hover = false,
  onClick,
  compact = false,
  glow = false,
}: GlassCardProps) {
  return (
    <div
      onClick={onClick}
      className={cn(
        "rounded-2xl bg-bg-elevated border border-border transition-all duration-200",
        compact ? "p-4" : "p-5",
        hover &&
          "hover:border-primary/20 hover:shadow-md hover:-translate-y-px cursor-pointer",
        glow &&
          "hover:border-primary/30 hover:shadow-glow-sm hover:-translate-y-0.5",
        onClick && "cursor-pointer",
        className,
      )}
    >
      {children}
    </div>
  );
}
