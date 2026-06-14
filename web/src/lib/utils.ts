// cn joins class names, dropping falsy values — a tiny dependency-free helper.
export function cn(...classes: Array<string | false | null | undefined>): string {
  return classes.filter(Boolean).join(" ");
}

// formatBytes renders a byte count as a human-readable size (0 = unlimited).
export function formatBytes(n: number): string {
  if (n === 0) return "∞";
  const units = ["B", "KB", "MB", "GB", "TB"];
  let v = n;
  let i = 0;
  while (v >= 1024 && i < units.length - 1) {
    v /= 1024;
    i++;
  }
  return `${v.toFixed(v < 10 && i > 0 ? 1 : 0)} ${units[i]}`;
}
