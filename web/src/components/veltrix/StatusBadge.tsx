import { cn } from "@/lib/utils";

type StatusType = "active" | "inactive" | "warning" | "error" | "info" | "optimal";

interface StatusBadgeProps {
  status: StatusType | string;
  label: string;
  pulse?: boolean;
}

const statusStyles: Record<string, string> = {
  active: "bg-success/15 text-success border-success/30 font-semibold",
  optimal: "bg-success/15 text-success border-success/30 font-semibold",
  inactive: "bg-fg-subtle/10 text-fg-subtle border-fg-subtle/20",
  warning: "bg-warning/15 text-warning border-warning/30 font-semibold",
  error: "bg-danger/15 text-danger border-danger/30 font-semibold",
  info: "bg-primary/15 text-primary border-primary/30 font-semibold",
};

const dotStyles: Record<string, string> = {
  active: "bg-success",
  optimal: "bg-success",
  inactive: "bg-fg-subtle",
  warning: "bg-warning",
  error: "bg-danger",
  info: "bg-primary",
};

export function StatusBadge({ status, label, pulse = true }: StatusBadgeProps) {
  const normalizedStatus = status.toLowerCase();
  const badgeStyle = statusStyles[normalizedStatus] ?? statusStyles.info;
  const dotStyle = dotStyles[normalizedStatus] ?? dotStyles.info;
  const shouldPulse = pulse && (normalizedStatus === "active" || normalizedStatus === "optimal");

  return (
    <span
      className={cn(
        "inline-flex items-center gap-1.5 rounded-full border px-2.5 py-0.5 text-[11px] font-semibold tracking-wide uppercase transition-transform duration-200 hover:scale-105",
        badgeStyle,
      )}
    >
      <span className="relative flex h-1.5 w-1.5 flex-shrink-0">
        {shouldPulse && (
          <span className={cn("absolute inline-flex h-full w-full animate-ping rounded-full opacity-75", dotStyle)} />
        )}
        <span className={cn("relative inline-flex h-1.5 w-1.5 rounded-full", dotStyle)} />
      </span>
      <span>{label}</span>
    </span>
  );
}
