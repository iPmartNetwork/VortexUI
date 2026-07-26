import { cn } from "@/lib/utils";
import { Button } from "@/components/ui";
import type { LucideIcon } from "lucide-react";
import { Inbox } from "lucide-react";

interface EmptyStateProps {
  icon?: LucideIcon;
  title: string;
  description?: string;
  action?: {
    label: string;
    onClick: () => void;
  };
  className?: string;
  compact?: boolean;
}

export function EmptyState({
  icon: Icon = Inbox,
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
        compact ? "py-10" : "py-20",
        className,
      )}
    >
      <div className="mb-5 grid h-16 w-16 place-items-center rounded-2xl bg-surface-2/60 border border-border/40">
        <Icon size={28} className="text-fg-subtle/50" />
      </div>
      <h3 className="text-sm font-semibold text-fg">{title}</h3>
      {description && (
        <p className="mt-1 max-w-sm text-xs text-fg-muted leading-relaxed">
          {description}
        </p>
      )}
      {action && (
        <Button onClick={action.onClick} variant="glass" className="mt-5">
          {action.label}
        </Button>
      )}
    </div>
  );
}
