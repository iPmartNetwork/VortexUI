// cn joins class names, dropping falsy values — a tiny dependency-free helper.
export function cn(...classes: Array<string | false | null | undefined>): string {
  return classes.filter(Boolean).join(" ");
}

// formatBytes renders a byte count as a human-readable size.
// 0 with allowUnlimited = true means "unlimited" (∞). Precise 2-decimal output.
export function formatBytes(n: number, allowUnlimited = true): string {
  if (n === 0 && allowUnlimited) return "∞";
  if (n === 0) return "0 B";
  const units = ["B", "KB", "MB", "GB", "TB", "PB"];
  let v = Math.abs(n);
  let i = 0;
  while (v >= 1024 && i < units.length - 1) {
    v /= 1024;
    i++;
  }
  const decimals = i === 0 ? 0 : v < 10 ? 2 : v < 100 ? 1 : 0;
  return `${v.toFixed(decimals)} ${units[i]}`;
}

// formatSpeed renders bytes/sec as a speed string.
export function formatSpeed(bytesPerSec: number): string {
  if (bytesPerSec <= 0) return "0 B/s";
  return formatBytes(bytesPerSec, false) + "/s";
}

// Percentage — safe division.
export function pct(used: number, total: number): number {
  if (total <= 0) return 0;
  return Math.min(100, (used / total) * 100);
}

// Relative time (short).
export function timeAgo(iso: string | null | undefined): string {
  if (!iso) return "—";
  const diff = Date.now() - new Date(iso).getTime();
  const sec = Math.floor(diff / 1000);
  if (sec < 60) return `${sec}s ago`;
  const min = Math.floor(sec / 60);
  if (min < 60) return `${min}m ago`;
  const hrs = Math.floor(min / 60);
  if (hrs < 24) return `${hrs}h ago`;
  const days = Math.floor(hrs / 24);
  return `${days}d ago`;
}
