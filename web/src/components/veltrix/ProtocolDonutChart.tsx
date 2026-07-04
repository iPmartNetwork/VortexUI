import { useState } from "react";
import { cn, formatBytes } from "@/lib/utils";

export interface ProtocolSlice {
  label: string;
  value: number;
  color: string;
}

const DEFAULT_COLORS = ["#3B82F6", "#8B5CF6", "#14B8A6", "#64748B", "#F59E0B", "#EC4899"];

export function ProtocolDonutChart({
  slices,
  centerLabel,
  centerValue,
  className,
}: {
  slices: ProtocolSlice[];
  centerLabel: string;
  centerValue: string | number;
  className?: string;
}) {
  const [hoverIdx, setHoverIdx] = useState<number | null>(null);
  const total = slices.reduce((s, x) => s + x.value, 0);
  if (total === 0) {
    return (
      <div className={cn("flex h-48 items-center justify-center text-xs text-fg-subtle", className)}>
        No protocol data yet
      </div>
    );
  }

  const r = 36;
  const circ = 2 * Math.PI * r;
  let offset = 0;
  const arcs = slices.map((slice, i) => {
    const pct = slice.value / total;
    const dash = pct * circ;
    const arcOffset = offset;
    offset += dash;
    return { slice, i, dash, arcOffset, pct };
  });

  const hovered = hoverIdx !== null ? slices[hoverIdx] : null;
  const hoveredPct = hovered ? ((hovered.value / total) * 100).toFixed(0) : null;

  return (
    <div className={cn("flex flex-col items-center gap-3", className)}>
      {/* Donut */}
      <div className="relative h-32 w-32 flex-shrink-0">
        <svg viewBox="0 0 100 100" className="h-full w-full -rotate-90">
          <circle cx="50" cy="50" r={r} fill="none" stroke="hsl(var(--border))" strokeWidth="8" opacity="0.22" />
          {arcs.map(({ slice, i, dash, arcOffset }) => (
            <circle
              key={slice.label}
              cx="50"
              cy="50"
              r={r}
              fill="none"
              stroke={slice.color || DEFAULT_COLORS[i % DEFAULT_COLORS.length]}
              strokeWidth={hoverIdx === i ? "9.5" : "8"}
              strokeDasharray={`${dash} ${circ - dash}`}
              strokeDashoffset={-arcOffset}
              strokeLinecap="round"
              opacity={hoverIdx !== null && hoverIdx !== i ? 0.3 : 1}
              className="transition-all duration-200 cursor-pointer"
              onMouseEnter={() => setHoverIdx(i)}
              onMouseLeave={() => setHoverIdx(null)}
            />
          ))}
        </svg>
        <div className="absolute inset-0 flex flex-col items-center justify-center text-center px-2 pointer-events-none">
          {hovered ? (
            <>
              <span className="text-sm font-black text-fg tabular-nums leading-none">{hoveredPct}%</span>
              <span className="text-[7px] font-bold uppercase tracking-wide text-fg-subtle mt-1 max-w-[64px] truncate">
                {hovered.label}
              </span>
            </>
          ) : (
            <>
              <span className="text-xl font-black text-fg tabular-nums leading-none">{centerValue}</span>
              <span className="text-[7px] font-bold uppercase tracking-widest text-fg-subtle mt-1">{centerLabel}</span>
            </>
          )}
        </div>
      </div>

      {/* 2-column legend grid — lone trailing item on an odd count is centered
          so rows never look lopsided (2 left / 1 stranded on the right). */}
      <div className="w-full grid grid-cols-2 gap-x-3 gap-y-1.5">
        {slices.map((slice, i) => {
          const pct = ((slice.value / total) * 100).toFixed(0);
          const isLoneTrailing = i === slices.length - 1 && slices.length % 2 !== 0;
          return (
            <div
              key={slice.label}
              onMouseEnter={() => setHoverIdx(i)}
              onMouseLeave={() => setHoverIdx(null)}
              className={cn(
                "flex items-center gap-1.5 text-xs min-w-0 rounded-md px-1.5 -mx-1.5 py-0.5 cursor-pointer transition-colors",
                hoverIdx === i ? "bg-surface-2/70" : "hover:bg-surface-2/40",
                isLoneTrailing && "col-span-2 justify-self-center",
              )}
            >
              <span
                className="rounded-full flex-shrink-0"
                style={{ background: slice.color || DEFAULT_COLORS[i % DEFAULT_COLORS.length], height: 7, width: 7 }}
              />
              <span
                className={cn(
                  "text-[9.5px] whitespace-nowrap overflow-hidden text-ellipsis",
                  hoverIdx === i ? "text-fg font-semibold" : "text-fg-muted",
                )}
              >
                {slice.label}
              </span>
              <span className="font-bold text-fg tabular-nums text-[9.5px] ms-2 flex-shrink-0">{pct}%</span>
            </div>
          );
        })}
      </div>
    </div>
  );
}

/** Group inbound rows into display labels for the protocol donut. */
export function buildProtocolSlices(
  inbounds: { protocol: string; transport?: string; security?: string }[],
): ProtocolSlice[] {
  const counts = new Map<string, number>();
  for (const ib of inbounds) {
    const p = (ib.protocol || "unknown").toLowerCase();
    const t = (ib.transport || "").toLowerCase();
    const s = (ib.security || "").toLowerCase();
    let label = p.toUpperCase();
    if (p === "vless" && s.includes("reality")) label = "VLESS+Reality";
    else if (p === "vmess" && t.includes("ws")) label = "VMess+WS+CDN";
    else if (p === "hysteria2" || p === "hysteria") label = "Hysteria2 UDP";
    else if (p === "trojan" || p === "shadowsocks") label = "Trojan / SS";
    else label = `${p}${t ? `+${t}` : ""}`;
    counts.set(label, (counts.get(label) ?? 0) + 1);
  }
  const colors = DEFAULT_COLORS;
  return [...counts.entries()]
    .sort((a, b) => b[1] - a[1])
    .map(([label, value], i) => ({ label, value, color: colors[i % colors.length] }));
}

export function formatDailyBandwidth(bytes: number): string {
  if (bytes >= 1024 ** 4) return `${(bytes / 1024 ** 4).toFixed(2)} TiB`;
  if (bytes >= 1024 ** 3) return `${(bytes / 1024 ** 3).toFixed(2)} GiB`;
  return formatBytes(bytes, false);
}
