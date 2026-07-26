import { cn } from "@/lib/utils";

interface ProtocolBadgeProps {
  label: string;
  className?: string;
}

function badgeStyle(label: string): string {
  const u = label.toUpperCase();
  if (u.startsWith("VLESS")) return "bg-cyan-500/15 text-cyan-600 border-cyan-500/30 dark:text-cyan-400";
  if (u.startsWith("VMESS")) return "bg-blue-500/15 text-blue-600 border-blue-500/30 dark:text-blue-400";
  if (u.startsWith("TROJAN")) return "bg-fuchsia-500/15 text-fuchsia-600 border-fuchsia-500/30 dark:text-fuchsia-400";
  if (u.startsWith("SHADOW") || u.includes("SS")) {
    return "bg-amber-500/15 text-amber-600 border-amber-500/30 dark:text-amber-400";
  }
  if (u.startsWith("HYSTERIA")) return "bg-violet-500/15 text-violet-600 border-violet-500/30 dark:text-violet-400";
  return "bg-surface-2 text-fg-muted border-border/60";
}

export function ProtocolBadge({ label, className }: ProtocolBadgeProps) {
  if (!label || label === "—") {
    return <span className="text-xs text-fg-subtle">—</span>;
  }
  const short = label.split("+")[0]?.trim() || label;
  return (
    <span
      className={cn(
        "inline-flex items-center rounded-md border px-2 py-0.5 text-[10px] font-semibold uppercase tracking-wide",
        badgeStyle(label),
        className,
      )}
    >
      {short}
    </span>
  );
}
