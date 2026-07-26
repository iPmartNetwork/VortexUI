import { cn } from "@/lib/utils";
import type { ConnectionStatus } from "@/hooks/useLiveTraffic";

interface LiveIndicatorProps {
  status: ConnectionStatus;
  speedLabel?: string;
  className?: string;
}

const STATUS_CONFIG = {
  live: {
    dot: "bg-green-500 shadow-[0_0_8px_2px] shadow-green-500/60",
    ping: "bg-green-400",
    label: "LIVE",
    pulse: true,
  },
  connecting: {
    dot: "bg-amber-400 shadow-[0_0_8px_2px] shadow-amber-400/50",
    ping: "bg-amber-300",
    label: "CONNECTING",
    pulse: true,
  },
  polling: {
    dot: "bg-blue-400 shadow-[0_0_8px_2px] shadow-blue-400/40",
    ping: "bg-blue-300",
    label: "POLLING",
    pulse: false,
  },
  error: {
    dot: "bg-red-500 shadow-[0_0_8px_2px] shadow-red-500/50",
    ping: "bg-red-400",
    label: "OFFLINE",
    pulse: false,
  },
} as const;

export function LiveIndicator({ status, speedLabel, className }: LiveIndicatorProps) {
  const cfg = STATUS_CONFIG[status];

  return (
    <span
      className={cn(
        "inline-flex items-center gap-1.5 rounded-full px-2 py-0.5 text-[7px] font-black uppercase tracking-wider border transition-all duration-500",
        status === "live" && "bg-green-500/15 border-green-500/30 text-green-400",
        status === "connecting" && "bg-amber-500/15 border-amber-500/30 text-amber-300",
        status === "polling" && "bg-blue-500/12 border-blue-500/25 text-blue-400",
        status === "error" && "bg-red-500/15 border-red-500/30 text-red-400",
        className,
      )}
    >
      {/* Animated dot */}
      <span className="relative flex h-1.5 w-1.5">
        {cfg.pulse && (
          <span
            className={cn(
              "absolute inline-flex h-full w-full animate-ping rounded-full opacity-75",
              cfg.ping,
            )}
          />
        )}
        <span className={cn("relative inline-flex h-1.5 w-1.5 rounded-full", cfg.dot)} />
      </span>

      {/* Label */}
      <span>{cfg.label}</span>

      {/* Speed */}
      {speedLabel && status === "live" && (
        <span className="ml-0.5 font-bold text-green-400/80 tracking-normal">
          {speedLabel}
        </span>
      )}
      {speedLabel && status === "polling" && (
        <span className="ml-0.5 font-bold text-blue-400/70 tracking-normal">
          {speedLabel}
        </span>
      )}
    </span>
  );
}
