import { cn } from "@/lib/utils";

interface CoreBadgeProps {
  core: "xray" | "singbox";
  className?: string;
}

export function CoreBadge({ core, className }: CoreBadgeProps) {
  const label = core === "singbox" ? "sing-box" : "XRAY";
  const style =
    core === "singbox"
      ? "bg-violet-500/15 text-violet-600 border-violet-500/30 dark:text-violet-400"
      : "bg-cyan-500/15 text-cyan-600 border-cyan-500/30 dark:text-cyan-400";
  return (
    <span
      className={cn(
        "inline-flex items-center rounded-md border px-2 py-0.5 text-[10px] font-semibold uppercase tracking-wide",
        style,
        className,
      )}
    >
      {label}
    </span>
  );
}
