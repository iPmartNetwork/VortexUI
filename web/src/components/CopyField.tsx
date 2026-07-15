import { useState, useCallback } from "react";
import { motion, AnimatePresence } from "framer-motion";
import { Check, Copy } from "lucide-react";
import { cn } from "@/lib/utils";

interface CopyFieldProps {
  value: string;
  mono?: boolean;
  className?: string;
  label?: string;
  /** Custom display value vs copied value (e.g. masked token) */
  displayValue?: string;
}

/**
 * CopyField shows a value with a one-click copy button and a brief animated
 * confirmation morph — icon flips from Copy → Check with a scale bounce.
 */
export function CopyField({ value, mono = true, className, label, displayValue }: CopyFieldProps) {
  const [state, setState] = useState<"idle" | "copied" | "error">("idle");

  const copy = useCallback(async () => {
    try {
      await navigator.clipboard?.writeText(value);
      setState("copied");
    } catch {
      setState("error");
    }
    setTimeout(() => setState("idle"), 1400);
  }, [value]);

  return (
    <div className={cn("flex items-center gap-2 group", className)}>
      {label && (
        <span className="text-xs font-medium text-fg-subtle whitespace-nowrap">{label}</span>
      )}
      <div
        className={cn(
          "min-w-0 flex-1 truncate rounded-lg border border-border/60 bg-surface-2/30 px-3 py-2 text-xs transition-all duration-200 group-hover:border-border",
          mono && "font-mono",
          state === "error" && "border-danger/40",
        )}
        title={value}
        dir="ltr"
      >
        {displayValue ?? value}
      </div>
      <button
        onClick={copy}
        aria-label={state === "copied" ? "Copied" : "Copy"}
        className={cn(
          "relative grid h-8 w-9 shrink-0 place-items-center rounded-lg border transition-all duration-200 overflow-hidden",
          state === "copied"
            ? "border-success/40 bg-success/10 text-success"
            : state === "error"
              ? "border-danger/40 bg-danger/10 text-danger"
              : "border-border/60 text-fg-muted hover:border-border hover:bg-surface-2/60 hover:text-fg",
        )}
      >
        <AnimatePresence mode="wait">
          {state === "copied" ? (
            <motion.span
              key="check"
              initial={{ scale: 0, rotate: -90 }}
              animate={{ scale: 1, rotate: 0 }}
              exit={{ scale: 0, rotate: 90 }}
              transition={{ type: "spring", stiffness: 500, damping: 24 }}
            >
              <Check size={15} />
            </motion.span>
          ) : state === "error" ? (
            <motion.span
              key="error"
              initial={{ scale: 0 }}
              animate={{ scale: 1 }}
              exit={{ scale: 0 }}
              transition={{ type: "spring", stiffness: 500, damping: 24 }}
            >
              <Check size={15} />
            </motion.span>
          ) : (
            <motion.span
              key="copy"
              initial={{ scale: 0, rotate: 90 }}
              animate={{ scale: 1, rotate: 0 }}
              exit={{ scale: 0, rotate: -90 }}
              transition={{ type: "spring", stiffness: 500, damping: 24 }}
            >
              <Copy size={15} />
            </motion.span>
          )}
        </AnimatePresence>
        {/* Ripple effect on click */}
        <span
          key={state}
          className={cn(
            "absolute inset-0 rounded-lg pointer-events-none",
            state === "copied" && "animate-ping bg-success/15",
            state === "error" && "animate-ping bg-danger/15",
          )}
        />
      </button>
    </div>
  );
}
