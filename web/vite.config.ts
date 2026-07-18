import { readFileSync } from "node:fs";
import path from "node:path";
import { defineConfig } from "vitest/config";
import react from "@vitejs/plugin-react";

function readPanelVersion(): string {
  try {
    return readFileSync(path.resolve(__dirname, "../VERSION"), "utf8").trim();
  } catch {
    return "0.0.0";
  }
}

const panelVersion = readPanelVersion();

// Dev server proxies API + subscription calls to the local panel so the SPA and
// backend share an origin during development.
export default defineConfig({
  plugins: [react()],
  define: {
    // Fallback for the sidebar before GET /api/version resolves; keep in sync with ../VERSION.
    __PANEL_VERSION__: JSON.stringify(panelVersion),
  },
  resolve: {
    alias: { "@": path.resolve(__dirname, "src") },
  },
  build: {
    rollupOptions: {
      output: {
        manualChunks: {
          "vendor-react": ["react", "react-dom", "react-router-dom"],
          "vendor-query": ["@tanstack/react-query"],
          "vendor-motion": ["framer-motion"],
          "vendor-charts": ["qrcode.react"],
        },
      },
    },
  },
  test: {
    environment: "node",
    include: ["src/**/*.test.ts"],
  },
  server: {
    port: 5173,
    proxy: {
      "/api": "http://localhost:8080",
      "/sub": "http://localhost:8080",
    },
  },
});
