import { cn } from "@/lib/utils";
import type { CoreType } from "@/lib/coreTypes";
import { normalizedEnabledCores } from "@/lib/coreTypes";

interface CoreBadgeProps {
  core: CoreType;
  enabledCores?: CoreType[];
  className?: string;
}

function SingleCoreBadge({ core, className }: { core: CoreType; className?: string }) {
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

export function CoreBadge({ core, enabledCores, className }: CoreBadgeProps) {
  const cores = normalizedEnabledCores({ core, enabled_cores: enabledCores });
  if (cores.length <= 1) {
    return <SingleCoreBadge core={cores[0] ?? core} className={className} />;
  }
  return (
    <span className={cn("inline-flex flex-wrap gap-1", className)}>
      {cores.map((c) => (
        <SingleCoreBadge key={c} core={c} />
      ))}
    </span>
  );
}
