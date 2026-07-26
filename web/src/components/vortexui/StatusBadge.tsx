import { cn } from "@/lib/utils";

type StatusType =
  | "active"
  | "inactive"
  | "warning"
  | "error"
  | "info"
  | "optimal"
  | "pending";

interface StatusBadgeProps {
  status: StatusType | string;
  label: string;
  pulse?: boolean;
  size?: "sm" | "md";
}

const statusStyles: Record<string, string> = {
  active:
    "bg-success/12 text-success border-success/25 font-semibold",
  optimal:
    "bg-success/12 text-success border-success/25 font-semibold",
  inactive: "bg-fg-subtle/8 text-fg-subtle border-fg-subtle/20",
  warning:
    "bg-warning/12 text-warning border-warning/25 font-semibold",
  error: "bg-danger/12 text-danger border-danger/25 font-semibold",
  info: "bg-primary/12 text-primary border-primary/25 font-semibold",
  pending:
    "bg-warning/12 text-warning border-warning/25 font-semibold",
};

const dotStyles: Record<string, string> = {
  active: "bg-success shadow-[0_0_6px_1px] shadow-success/40",
  optimal: "bg-success shadow-[0_0_6px_1px] shadow-success/40",
  inactive: "bg-fg-subtle",
  warning: "bg-warning shadow-[0_0_6px_1px] shadow-warning/40",
  error: "bg-danger shadow-[0_0_6px_1px] shadow-danger/40",
  info: "bg-primary shadow-[0_0_6px_1px] shadow-primary/40",
  pending: "bg-warning shadow-[0_0_6px_1px] shadow-warning/40",
};

const sizes: Record<string, string> = {
  sm: "px-2 py-0.5 text-[9px]",
  md: "px-2.5 py-0.5 text-[11px]",
};

export function StatusBadge({
  status,
  label,
  pulse = true,
  size = "md",
}: StatusBadgeProps) {
  const normalizedStatus = status.toLowerCase();
  const badgeStyle = statusStyles[normalizedStatus] ?? statusStyles.info;
  const dotStyle = dotStyles[normalizedStatus] ?? dotStyles.info;
  const shouldPulse =
    pulse &&
    (normalizedStatus === "active" ||
      normalizedStatus === "optimal" ||
      normalizedStatus === "pending");

  return (
    <span
      className={cn(
        "inline-flex items-center gap-1.5 rounded-full border transition-all duration-200",
        badgeStyle,
        sizes[size],
      )}
    >
      <span className="relative flex h-1.5 w-1.5 flex-shrink-0">
        {shouldPulse && (
          <span className="absolute inline-flex h-full w-full animate-ping rounded-full bg-current opacity-75" />
        )}
        <span className={cn("relative inline-flex h-1.5 w-1.5 rounded-full", dotStyle)} />
      </span>
      <span>{label}</span>
    </span>
  );
}
