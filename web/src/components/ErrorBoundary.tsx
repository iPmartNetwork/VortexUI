import { Component } from "react";
import { motion } from "framer-motion";
import {
  AlertTriangle,
  RotateCcw,
  WifiOff,
  RefreshCw,
  FileWarning,
} from "lucide-react";
import { cn } from "@/lib/utils";

interface Props {
  children: React.ReactNode;
  fallback?: React.ReactNode;
  /** Called when the boundary resets (retry) */
  onReset?: () => void;
}

interface State {
  hasError: boolean;
  error: Error | null;
  /** Number of retry attempts */
  retryCount: number;
}

// ─── Error type detection ──────────────────────────────────────

function isChunkLoadError(error: Error): boolean {
  const msg = error.message ?? "";
  const name = error.name ?? "";
  // Common chunk-loading error patterns across bundlers
  return (
    msg.includes("Loading chunk") ||
    msg.includes("ChunkLoadError") ||
    msg.includes("loading CSS chunk") ||
    msg.includes("dynamically imported module") ||
    msg.includes("Failed to fetch dynamically imported module") ||
    msg.includes("Importing a module script failed") ||
    msg.includes("error loading dynamically") ||
    msg.includes("ScriptExternalLoadError") ||
    name === "ChunkLoadError"
  );
}

function getErrorIcon(error: Error | null, isChunkError = false): {
  icon: typeof AlertTriangle;
  label: string;
} {
  if (!error) return { icon: AlertTriangle, label: "Error" };
  if (isChunkError) return { icon: WifiOff, label: "Connection Lost" };
  if (error.message.includes("404"))
    return { icon: FileWarning, label: "Not Found" };
  return { icon: AlertTriangle, label: "Something went wrong" };
}

function getErrorMessage(error: Error | null): string {
  if (!error) return "An unexpected error occurred.";

  if (isChunkLoadError(error)) {
    return (
      "Failed to load a page module. This usually happens when " +
      "your internet connection was interrupted, or you're running " +
      "an older version of the app. Try refreshing the page."
    );
  }

  if (error.message.includes("404")) {
    return "The requested page module was not found. This may be due to a recent deployment update. Please try reloading.";
  }

  // Strip technical details for non-chunk errors
  return error.message || "An unexpected error occurred.";
}

// ─── Glass-styled divider ──────────────────────────────────────

function ErrorDivider() {
  return (
    <div className="h-px bg-gradient-to-r from-transparent via-border/40 to-transparent" />
  );
}

// ─── Retry count indicator ─────────────────────────────────────

function RetryBadge({ count }: { count: number }) {
  if (count === 0) return null;
  return (
    <span className="inline-flex items-center gap-1 rounded-full bg-warning/10 px-2.5 py-0.5 text-[10px] font-semibold text-warning border border-warning/20">
      <RefreshCw size={10} />
      Retry #{count}
    </span>
  );
}

// ════════════════════════════════════════════════════════════════
// ERROR BOUNDARY — Class component (required by React)
// ════════════════════════════════════════════════════════════════
export class ErrorBoundary extends Component<Props, State> {
  state: State = { hasError: false, error: null, retryCount: 0 };

  static getDerivedStateFromError(error: Error): Partial<State> {
    return { hasError: true, error };
  }

  componentDidCatch(error: Error, info: React.ErrorInfo) {
    // Log to console in development
    if (import.meta.env.DEV) {
      console.group("🔴 ErrorBoundary caught an error");
      console.error("Error:", error);
      console.error("Component stack:", info.componentStack);
      console.groupEnd();
    }
  }

  private handleRetry = () => {
    this.props.onReset?.();
    this.setState((prev) => ({
      hasError: false,
      error: null,
      retryCount: prev.retryCount + 1,
    }));
  };

  private handleHardReload = () => {
    window.location.reload();
  };

  render() {
    if (this.state.hasError) {
      // Custom fallback override
      if (this.props.fallback) return this.props.fallback;

      const { error, retryCount } = this.state;
      const isChunk = isChunkLoadError(error!);
      const { icon: ErrorIcon, label } = getErrorIcon(error, isChunk);
      const message = getErrorMessage(error);

      return (
        <motion.div
          initial={{ opacity: 0, scale: 0.96 }}
          animate={{ opacity: 1, scale: 1 }}
          transition={{ duration: 0.3, ease: [0.16, 1, 0.3, 1] }}
          className="flex min-h-[400px] items-center justify-center p-6"
        >
          <div
            className={cn(
              "relative w-full max-w-md overflow-hidden",
              "rounded-2xl bg-bg-elevated border border-border p-8",
              "shadow-lg",
            )}
          >
            {/* Decorative gradient blob */}
            <div
              className={cn(
                "pointer-events-none absolute -top-20 -end-20 h-40 w-40 rounded-full opacity-[0.08]",
                isChunk ? "bg-warning" : "bg-danger",
              )}
              style={{ filter: "blur(40px)" }}
            />

            {/* Error icon */}
            <div
              className={cn(
                "mx-auto grid h-16 w-16 place-items-center rounded-2xl mb-5",
                isChunk
                  ? "bg-warning/10 text-warning"
                  : "bg-danger/10 text-danger",
              )}
            >
              <motion.div
                initial={{ rotate: -12, scale: 0.8 }}
                animate={{ rotate: 0, scale: 1 }}
                transition={{ duration: 0.4, ease: [0.16, 1, 0.3, 1] }}
              >
                <ErrorIcon size={32} strokeWidth={1.5} />
              </motion.div>
            </div>

            {/* Title */}
            <h2
              className={cn(
                "text-lg font-bold text-center mb-1",
                isChunk ? "text-warning" : "text-danger",
              )}
            >
              {label}
            </h2>

            {/* Error type badge */}
            <div className="flex justify-center mb-3">
              <RetryBadge count={retryCount} />
            </div>

            <ErrorDivider />

            {/* Message */}
            <p className="text-sm text-fg-muted text-center leading-relaxed mt-4 mb-6">
              {message}
            </p>

            {/* Action buttons */}
            <div className="flex flex-col gap-2">
              <button
                onClick={this.handleRetry}
                className={cn(
                  "inline-flex items-center justify-center gap-2 rounded-xl px-5 py-2.5",
                  "text-sm font-medium transition-all duration-200 active:scale-[0.97]",
                  "grad-bg text-primary-fg shadow-md hover:shadow-lg hover:brightness-110",
                )}
              >
                <RotateCcw size={14} />
                Try Again
              </button>

              {isChunk && (
                <button
                  onClick={this.handleHardReload}
                  className={cn(
                    "inline-flex items-center justify-center gap-2 rounded-xl px-5 py-2.5",
                    "text-sm font-medium transition-all duration-200 active:scale-[0.97]",
                    "border border-border/80 bg-surface/60 hover:bg-surface-2/80 text-fg backdrop-blur-sm",
                  )}
                >
                  <RefreshCw size={14} />
                  Reload Page
                </button>
              )}

              {/* Technical details (dev only) */}
              {import.meta.env.DEV && error && (
                <details className="mt-3">
                  <summary className="cursor-pointer text-[10px] text-fg-subtle hover:text-fg-muted transition-colors text-center">
                    Technical details
                  </summary>
                  <pre className="mt-2 rounded-xl bg-surface-2/50 p-3 text-[10px] text-fg-muted overflow-auto max-h-28 text-start leading-relaxed font-mono">
                    {error.name}: {error.message}
                    {"\n"}
                    {error.stack?.split("\n").slice(0, 3).join("\n")}
                  </pre>
                </details>
              )}
            </div>
          </div>
        </motion.div>
      );
    }

    return this.props.children;
  }
}
