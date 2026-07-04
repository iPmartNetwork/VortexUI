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
  const total = slices.reduce((s, x) => s + x.value, 0);
  if (total === 0) {
    return (
      <div className={cn("flex h-48 items-center justify-center text-xs text-fg-subtle", className)}>
        No protocol data yet
      </div>
    );
  }

  const r = 44;
  const circ = 2 * Math.PI * r;
  let offset = 0;
  const rings = slices.map((slice, i) => {
    const pct = slice.value / total;
    const dash = pct * circ;
    const ring = (
      <circle
        key={slice.label}
        cx="50"
        cy="50"
        r={r}
        fill="none"
        stroke={slice.color || DEFAULT_COLORS[i % DEFAULT_COLORS.length]}
        strokeWidth="9"
        strokeDasharray={`${dash} ${circ - dash}`}
        strokeDashoffset={-offset}
        strokeLinecap="round"
        className="transition-all duration-500"
      />
    );
    offset += dash;
    return ring;
  });

  // 2×2 grid — pair slices into rows of 2
  const pairs: ProtocolSlice[][] = [];
  for (let i = 0; i < slices.length; i += 2) {
    pairs.push(slices.slice(i, i + 2));
  }

  return (
    <div className={cn("flex flex-col items-center gap-5", className)}>
      {/* Donut */}
      <div className="relative h-44 w-44 flex-shrink-0">
        <svg viewBox="0 0 100 100" className="h-full w-full -rotate-90">
          <circle cx="50" cy="50" r={r} fill="none" stroke="hsl(var(--border))" strokeWidth="9" opacity="0.3" />
          {rings}
        </svg>
        <div className="absolute inset-0 flex flex-col items-center justify-center text-center px-3">
          <span className="text-2xl font-black text-fg tabular-nums leading-none">{centerValue}</span>
          <span className="text-[9px] font-bold uppercase tracking-widest text-fg-subtle mt-1">{centerLabel}</span>
        </div>
      </div>

      {/* 2×2 legend grid */}
      <div className="w-full grid grid-cols-2 gap-x-4 gap-y-2.5">
        {slices.map((slice, i) => {
          const pct = ((slice.value / total) * 100).toFixed(0);
          return (
            <div key={slice.label} className="flex items-center justify-between gap-1.5 text-xs min-w-0">
              <div className="flex items-center gap-1.5 min-w-0">
                <span
                  className="h-2.5 w-2.5 rounded-full flex-shrink-0"
                  style={{ background: slice.color || DEFAULT_COLORS[i % DEFAULT_COLORS.length] }}
                />
                <span className="text-fg-muted truncate text-[11px]">{slice.label}</span>
              </div>
              <span className="font-bold text-fg tabular-nums flex-shrink-0 text-[11px]">{pct}%</span>
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
