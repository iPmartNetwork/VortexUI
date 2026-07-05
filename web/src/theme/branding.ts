export const BRANDING_STORAGE_KEY = "vortex_branding";

export interface BrandingState {
  accentColor?: string;
  logoURL?: string;
  footerText?: string;
  panelName?: string;
}

const BRANDING_VARS = [
  "--primary",
  "--primary-hover",
  "--accent",
  "--accent-2",
  "--ring",
  "--glow-primary",
  "--glow-accent",
] as const;

export function readBranding(): BrandingState {
  try {
    const raw = localStorage.getItem(BRANDING_STORAGE_KEY);
    return raw ? (JSON.parse(raw) as BrandingState) : {};
  } catch {
    return {};
  }
}

export function writeBranding(patch: BrandingState): BrandingState {
  const next = { ...readBranding(), ...patch };
  localStorage.setItem(BRANDING_STORAGE_KEY, JSON.stringify(next));
  return next;
}

function parseHex(hex: string): { r: number; g: number; b: number } | null {
  const m = /^#?([0-9a-f]{6})$/i.exec(hex.trim());
  if (!m) return null;
  const n = parseInt(m[1], 16);
  return { r: (n >> 16) & 255, g: (n >> 8) & 255, b: n & 255 };
}

/** Returns HSL components as "H S% L%" for CSS variables. */
export function hexToHslComponents(hex: string): string | null {
  const rgb = parseHex(hex);
  if (!rgb) return null;

  const r = rgb.r / 255;
  const g = rgb.g / 255;
  const b = rgb.b / 255;
  const max = Math.max(r, g, b);
  const min = Math.min(r, g, b);
  const l = (max + min) / 2;
  let h = 0;
  let s = 0;

  if (max !== min) {
    const d = max - min;
    s = l > 0.5 ? d / (2 - max - min) : d / (max + min);
    switch (max) {
      case r:
        h = ((g - b) / d + (g < b ? 6 : 0)) / 6;
        break;
      case g:
        h = ((b - r) / d + 2) / 6;
        break;
      default:
        h = ((r - g) / d + 4) / 6;
    }
  }

  return `${Math.round(h * 360)} ${Math.round(s * 100)}% ${Math.round(l * 100)}%`;
}

function adjustLightness(hsl: string, delta: number): string {
  const parts = hsl.match(/^(\d+)\s+(\d+)%\s+(\d+)%$/);
  if (!parts) return hsl;
  const h = Number(parts[1]);
  const s = Number(parts[2]);
  const l = Math.min(100, Math.max(0, Number(parts[3]) + delta));
  return `${h} ${s}% ${l}%`;
}

export function clearAccentOverride(): void {
  const root = document.documentElement;
  for (const v of BRANDING_VARS) {
    root.style.removeProperty(v);
  }
}

export function applyAccentColor(hex: string, mode: "dark" | "light"): void {
  const hsl = hexToHslComponents(hex);
  if (!hsl) return;

  const root = document.documentElement;
  const primaryHover = adjustLightness(hsl, mode === "dark" ? 8 : -6);
  const accent2 = adjustLightness(hsl, mode === "dark" ? -4 : 4);

  root.style.setProperty("--primary", hsl);
  root.style.setProperty("--primary-hover", primaryHover);
  root.style.setProperty("--accent", hsl);
  root.style.setProperty("--accent-2", accent2);
  root.style.setProperty("--ring", hsl);
  root.style.setProperty("--glow-primary", hsl);
  root.style.setProperty("--glow-accent", hsl);
}

export function applyBrandingFromStorage(mode?: "dark" | "light"): void {
  const { accentColor } = readBranding();
  if (!accentColor) {
    clearAccentOverride();
    return;
  }

  const resolved =
    mode ??
    (document.documentElement.classList.contains("light") ? "light" : "dark");
  applyAccentColor(accentColor, resolved);
}

export function saveAndApplyBranding(
  patch: BrandingState,
  mode: "dark" | "light",
): void {
  const next = writeBranding(patch);
  if (next.accentColor) {
    applyAccentColor(next.accentColor, mode);
  } else {
    clearAccentOverride();
  }
}
