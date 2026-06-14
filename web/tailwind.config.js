/** @type {import('tailwindcss').Config} */
const c = (v) => `hsl(var(--${v}) / <alpha-value>)`;

export default {
  content: ["./index.html", "./src/**/*.{ts,tsx}"],
  darkMode: "class",
  theme: {
    extend: {
      colors: {
        bg: c("bg"),
        "bg-elevated": c("bg-elevated"),
        surface: c("surface"),
        "surface-2": c("surface-2"),
        border: c("border"),
        "border-strong": c("border-strong"),
        fg: c("fg"),
        "fg-muted": c("fg-muted"),
        "fg-subtle": c("fg-subtle"),
        primary: c("primary"),
        "primary-hover": c("primary-hover"),
        "primary-fg": c("primary-fg"),
        accent: c("accent"),
        "accent-2": c("accent-2"),
        success: c("success"),
        warning: c("warning"),
        danger: c("danger"),
        ring: c("ring"),

        /* Legacy compat aliases */
        background: c("bg"),
        foreground: c("fg"),
        card: c("surface"),
        muted: c("surface-2"),
        "muted-foreground": c("fg-muted"),
        "primary-foreground": c("primary-fg"),
        destructive: c("danger"),
      },
      fontFamily: {
        sans: ["Inter", "ui-sans-serif", "system-ui", "-apple-system", "sans-serif"],
        mono: ["JetBrains Mono", "ui-monospace", "monospace"],
      },
      borderRadius: {
        "2xl": "1rem",
        "3xl": "1.25rem",
      },
      boxShadow: {
        glow: "0 0 20px -4px hsl(var(--glow-primary) / 0.35)",
        "glow-sm": "0 0 10px -2px hsl(var(--glow-primary) / 0.25)",
      },
      keyframes: {
        "fade-in": { from: { opacity: "0" }, to: { opacity: "1" } },
        "scale-in": {
          from: { opacity: "0", transform: "scale(0.96) translateY(4px)" },
          to: { opacity: "1", transform: "scale(1) translateY(0)" },
        },
        "slide-up": {
          from: { opacity: "0", transform: "translateY(10px)" },
          to: { opacity: "1", transform: "translateY(0)" },
        },
        shimmer: {
          "0%": { backgroundPosition: "-200% 0" },
          "100%": { backgroundPosition: "200% 0" },
        },
      },
      animation: {
        "fade-in": "fade-in 0.25s ease-out both",
        "scale-in": "scale-in 0.2s cubic-bezier(0.16, 1, 0.3, 1) both",
        "slide-up": "slide-up 0.3s cubic-bezier(0.16, 1, 0.3, 1) both",
        shimmer: "shimmer 2s linear infinite",
      },
    },
  },
  plugins: [
    function ({ addVariant }: any) {
      addVariant("rtl", '[dir="rtl"] &');
    },
  ],
};
