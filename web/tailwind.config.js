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
        "surface-3": c("surface-3"),
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
        display: ["Orbitron", "sans-serif"],
      },
      fontSize: {
        /* Cyber scale — fine-tuned for dense SaaS dashboards */
        xs: ["0.6875rem", { lineHeight: "1.25", letterSpacing: "0.04em" }],      /* 11px — labels, badges, table headers */
        sm: ["0.8125rem", { lineHeight: "1.4", letterSpacing: "0.01em" }],       /* 13px — table data, small text */
        base: ["0.875rem", { lineHeight: "1.5", letterSpacing: "-0.005em" }],    /* 14px — body text */
        lg: ["0.9375rem", { lineHeight: "1.5", letterSpacing: "-0.005em" }],    /* 15px — large body */
        xl: ["1.125rem", { lineHeight: "1.25", letterSpacing: "-0.011em" }],    /* 18px — subheadings */
        "2xl": ["1.375rem", { lineHeight: "1.2", letterSpacing: "-0.015em" }],  /* 22px — page titles */
        "3xl": ["1.75rem", { lineHeight: "1.15", letterSpacing: "-0.02em" }],   /* 28px — hero headings */
        "4xl": ["2.25rem", { lineHeight: "1.1", letterSpacing: "-0.025em" }],   /* 36px — display */
      },
      fontWeight: {
        normal: "400",
        medium: "500",
        semibold: "600",
        bold: "700",
        black: "800", /* map Inter's newly-loaded 800 weight */
      },
      letterSpacing: {
        tighter: "-0.025em",
        tight: "-0.015em",
        normal: "-0.011em",
        wide: "0.02em",
        wider: "0.04em",
        widest: "0.08em",
      },
      borderRadius: {
        "2xl": "1rem",
        "3xl": "1.25rem",
        "4xl": "1.5rem",
      },
      boxShadow: {
        glow: "0 0 20px -4px hsl(var(--glow-primary) / 0.35)",
        "glow-sm": "0 0 10px -2px hsl(var(--glow-primary) / 0.25)",
        "glow-lg": "0 0 40px -8px hsl(var(--glow-primary) / 0.2), 0 0 80px -16px hsl(var(--glow-primary) / 0.1)",
        "glow-accent": "0 0 20px -4px hsl(var(--glow-accent) / 0.35)",
        "inner-glow": "inset 0 0 20px -8px hsl(var(--glow-primary) / 0.15)",
        "cyber": "0 0 0 1px hsl(var(--primary) / 0.3), 0 0 20px -6px hsl(var(--primary) / 0.25)",
        "cyber-lg": "0 0 0 1px hsl(var(--primary) / 0.4), 0 0 30px -8px hsl(var(--primary) / 0.3), inset 0 0 30px -12px hsl(var(--primary) / 0.1)",
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
        "slide-down": {
          from: { opacity: "0", transform: "translateY(-10px)" },
          to: { opacity: "1", transform: "translateY(0)" },
        },
        shimmer: {
          "0%": { backgroundPosition: "-200% 0" },
          "100%": { backgroundPosition: "200% 0" },
        },
        "pulse-glow": {
          "0%, 100%": { boxShadow: "0 0 20px -4px hsl(var(--glow-primary) / 0.2)" },
          "50%": { boxShadow: "0 0 30px -4px hsl(var(--glow-primary) / 0.5)" },
        },
        "float": {
          "0%, 100%": { transform: "translateY(0)" },
          "50%": { transform: "translateY(-6px)" },
        },
        "breath": {
          "0%, 100%": { opacity: "0.4" },
          "50%": { opacity: "0.8" },
        },
        "rotate-glow": {
          "0%": { transform: "rotate(0deg)" },
          "100%": { transform: "rotate(360deg)" },
        },
        "slide-in-right": {
          from: { opacity: "0", transform: "translateX(100%)" },
          to: { opacity: "1", transform: "translateX(0)" },
        },
        "slide-in-left": {
          from: { opacity: "0", transform: "translateX(-100%)" },
          to: { opacity: "1", transform: "translateX(0)" },
        },
        "progress": {
          from: { width: "100%" },
          to: { width: "0%" },
        },
      },
      animation: {
        "fade-in": "fade-in 0.25s ease-out both",
        "scale-in": "scale-in 0.2s cubic-bezier(0.16, 1, 0.3, 1) both",
        "slide-up": "slide-up 0.3s cubic-bezier(0.16, 1, 0.3, 1) both",
        "slide-down": "slide-down 0.3s cubic-bezier(0.16, 1, 0.3, 1) both",
        shimmer: "shimmer 2s linear infinite",
        "pulse-glow": "pulse-glow 2s ease-in-out infinite",
        float: "float 3s ease-in-out infinite",
        breath: "breath 3s ease-in-out infinite",
        "rotate-glow": "rotate-glow 4s linear infinite",
        "slide-in-right": "slide-in-right 0.3s cubic-bezier(0.16, 1, 0.3, 1) both",
        "slide-in-left": "slide-in-left 0.3s cubic-bezier(0.16, 1, 0.3, 1) both",
        progress: "progress linear forwards",
      },
      backgroundImage: {
        "gradient-radial": "radial-gradient(var(--tw-gradient-stops))",
        "cyber-grid": "linear-gradient(var(--grid-color) 1px, transparent 1px), linear-gradient(90deg, var(--grid-color) 1px, transparent 1px)",
      },
      transitionTimingFunction: {
        "spring": "cubic-bezier(0.16, 1, 0.3, 1)",
        "spring-out": "cubic-bezier(0.32, 0.72, 0, 1)",
      },
    },
  },
  plugins: [
    function ({ addVariant }: any) {
      addVariant("rtl", '[dir="rtl"] &');
    },
  ],
};
