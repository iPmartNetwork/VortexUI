import { cn } from "@/lib/utils";
import { Button } from "@/components/ui";
import type { LucideIcon } from "lucide-react";

interface EmptyStateProps {
  icon: LucideIcon;
  title: string;
  description?: string;
  action?: {
    label: string;
    onClick: () => void;
  };
  className?: string;
  compact?: boolean;
}

/**
 * Empty state placeholder with icon, message, and optional CTA.
 * Used across all pages when data arrays are empty.
 */
export function EmptyState({
  icon: Icon,
  title,
  description,
  action,
  className,
  compact = false,
}: EmptyStateProps) {
  return (
    <div
      className={cn(
        "flex flex-col items-center justify-center text-center animate-fade-in",
        compact ? "py-8" : "py-16",
        className,
      )}
    >
      <div className="mb-4 grid h-14 w-14 place-items-center rounded-2xl bg-surface-2/80 border border-border/50">
        <Icon size={26} className="text-fg-subtle/60" />
      </div>
      <h3 className="text-sm font-semibold text-fg">{title}</h3>
      {description && (
        <p className="mt-1 max-w-sm text-xs text-fg-muted leading-relaxed">{description}</p>
      )}
      {action && (
        <Button
          onClick={action.onClick}
          className="mt-4"
          variant="outline"
        >
          {action.label}
        </Button>
      )}
    </div>
  );
}
