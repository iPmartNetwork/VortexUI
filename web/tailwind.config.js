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

        /* Back-compat aliases for pages not yet migrated to the new tokens. */
        background: c("bg"),
        foreground: c("fg"),
        card: c("surface"),
        muted: c("surface-2"),
        "muted-foreground": c("fg-muted"),
        "primary-foreground": c("primary-fg"),
        destructive: c("danger"),
      },
      fontFamily: {
        mono: ["JetBrains Mono", "ui-monospace", "monospace"],
      },
      keyframes: {
        "fade-in": { from: { opacity: "0" }, to: { opacity: "1" } },
        "scale-in": {
          from: { opacity: "0", transform: "scale(0.97)" },
          to: { opacity: "1", transform: "scale(1)" },
        },
        "slide-up": {
          from: { opacity: "0", transform: "translateY(8px)" },
          to: { opacity: "1", transform: "translateY(0)" },
        },
      },
      animation: {
        "fade-in": "fade-in 0.2s ease-out",
        "scale-in": "scale-in 0.16s ease-out",
        "slide-up": "slide-up 0.25s ease-out",
      },
    },
  },
  plugins: [],
};
