import { useState, useEffect, useRef, useCallback, createContext, useContext, type ReactNode } from "react";
import {
  Play, Pause, SkipBack, Settings2,
} from "lucide-react";
import { cn } from "@/lib/utils";

// ════════════════════════════════════════════════════════════════
// Context — animation state shared across all charts
// ════════════════════════════════════════════════════════════════

interface AnimationState {
  paused: boolean;
  speed: number; // 0.25, 0.5, 1, 2, 4
  resetKey: number;
  togglePause: () => void;
  setSpeed: (s: number) => void;
  reset: () => void;
}

const SPEEDS = [0.25, 0.5, 1, 2, 4] as const;

const AnimationContext = createContext<AnimationState>({
  paused: false,
  speed: 1,
  resetKey: 0,
  togglePause: () => {},
  setSpeed: () => {},
  reset: () => {},
});

export function useAnimationControl() {
  return useContext(AnimationContext);
}

// ════════════════════════════════════════════════════════════════
// Provider — wraps the app or dashboard area
// ════════════════════════════════════════════════════════════════

export function AnimationProvider({ children }: { children: ReactNode }) {
  const [paused, setPaused] = useState(false);
  const [speed, setSpeedState] = useState<number>(1);
  const [resetKey, setResetKey] = useState(0);

  const togglePause = useCallback(() => setPaused(p => !p), []);
  const setSpeed = useCallback((s: number) => setSpeedState(s), []);
  const reset = useCallback(() => setResetKey(k => k + 1), []);

  return (
    <AnimationContext.Provider value={{ paused, speed, resetKey, togglePause, setSpeed, reset }}>
      {children}
    </AnimationContext.Provider>
  );
}

// ════════════════════════════════════════════════════════════════
// Control bar — placed above charts
// ════════════════════════════════════════════════════════════════

export function ChartAnimationControls({ className }: { className?: string }) {
  const { paused, speed, togglePause, setSpeed, reset } = useAnimationControl();
  const [open, setOpen] = useState(false);
  const panelRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    function onClick(e: MouseEvent) {
      if (panelRef.current && !panelRef.current.contains(e.target as Node)) {
        setOpen(false);
      }
    }
    if (open) document.addEventListener("mousedown", onClick);
    return () => document.removeEventListener("mousedown", onClick);
  }, [open]);

  return (
    <div ref={panelRef} className={cn("relative flex items-center gap-1", className)}>
      {/* Play/Pause */}
      <button
        type="button"
        onClick={togglePause}
        className="h-7 w-7 flex items-center justify-center rounded-lg border border-border/50 bg-surface-2/40 hover:bg-surface-2/80 transition text-fg-muted hover:text-fg"
        title={paused ? "Resume animations" : "Pause animations"}
      >
        {paused ? <Play size={12} /> : <Pause size={12} />}
      </button>

      {/* Speed indicator */}
      <button
        type="button"
        onClick={() => setOpen(!open)}
        className={cn(
          "h-7 px-2 flex items-center gap-1 rounded-lg border transition font-mono text-[10px] font-bold",
          open
            ? "border-primary/30 bg-primary/10 text-primary"
            : "border-border/50 bg-surface-2/40 text-fg-muted hover:bg-surface-2/80",
        )}
      >
        <Settings2 size={11} />
        {speed}x
      </button>

      {paused && (
        <span className="flex items-center gap-1 text-[9px] text-amber-400 font-semibold animate-pulse">
          <span className="h-1.5 w-1.5 rounded-full bg-amber-400" />
          PAUSED
        </span>
      )}

      {/* Speed dropdown */}
      {open && (
        <div className="absolute top-full mt-1 end-0 z-20 min-w-[140px] rounded-lg border border-border/60 bg-surface shadow-lg py-1">
          <p className="px-3 py-1 text-[9px] font-bold uppercase tracking-wider text-fg-subtle/60">
            Animation Speed
          </p>
          {SPEEDS.map((s) => (
            <button
              key={s}
              type="button"
              onClick={() => { setSpeed(s); setOpen(false); }}
              className={cn(
                "w-full flex items-center justify-between px-3 py-1.5 text-xs transition",
                speed === s
                  ? "text-primary bg-primary/8"
                  : "text-fg-muted hover:bg-surface-2/60",
              )}
            >
              <span>{s}x</span>
              {speed === s && <span className="text-[9px]">✓</span>}
            </button>
          ))}
          <div className="border-t border-border/30 my-1" />
          <button
            type="button"
            onClick={() => { reset(); setOpen(false); }}
            className="w-full flex items-center gap-2 px-3 py-1.5 text-xs text-fg-muted hover:bg-surface-2/60 transition"
          >
            <SkipBack size={12} />
            Reset Animations
          </button>
        </div>
      )}
    </div>
  );
}
