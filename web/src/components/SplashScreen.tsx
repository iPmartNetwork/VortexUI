import { useEffect, useState, useRef } from "react";
import { motion, AnimatePresence } from "framer-motion";
import { useAuth } from "@/auth/auth";

// ─── Loading stage labels ──────────────────────────────────────

const STAGES = [
  { label: "Initializing", duration: 600 },
  { label: "Loading modules", duration: 800 },
  { label: "Connecting", duration: 700 },
  { label: "Syncing", duration: 500 },
  { label: "Ready", duration: 400 },
];

// ─── Stylized Vortex "V" logo ──────────────────────────────────

function VortexLogo({ size = 80 }: { size?: number }) {
  return (
    <svg
      width={size}
      height={size}
      viewBox="0 0 64 64"
      fill="none"
      xmlns="http://www.w3.org/2000/svg"
      className="drop-shadow-[0_0_20px_hsl(var(--glow-primary)/0.3)]"
    >
      <defs>
        <linearGradient id="vortex-grad" x1="0" y1="0" x2="1" y2="1">
          <stop offset="0%" stopColor="hsl(239, 84%, 74%)" />
          <stop offset="50%" stopColor="hsl(198, 93%, 60%)" />
          <stop offset="100%" stopColor="hsl(187, 100%, 42%)" />
        </linearGradient>
        <linearGradient id="vortex-grad-2" x1="0" y1="0" x2="0" y2="1">
          <stop offset="0%" stopColor="hsl(198, 93%, 60%)" stopOpacity="0.4" />
          <stop offset="100%" stopColor="hsl(198, 93%, 60%)" stopOpacity="0" />
        </linearGradient>
      </defs>

      {/* Outer glow ring */}
      <circle
        cx="32"
        cy="32"
        r="30"
        stroke="url(#vortex-grad)"
        strokeWidth="0.5"
        strokeOpacity="0.3"
        fill="none"
      />

      {/* Stylized V — vortex shape */}
      <motion.path
        d="M18 50 L32 18 L46 50"
        stroke="url(#vortex-grad)"
        strokeWidth="3"
        strokeLinecap="round"
        strokeLinejoin="round"
        fill="none"
        initial={{ pathLength: 0, opacity: 0 }}
        animate={{ pathLength: 1, opacity: 1 }}
        transition={{ duration: 1.2, ease: [0.16, 1, 0.3, 1] }}
      />

      {/* Right swoosh */}
      <motion.path
        d="M32 18 C40 18, 48 24, 50 34"
        stroke="url(#vortex-grad-2)"
        strokeWidth="2"
        strokeLinecap="round"
        fill="none"
        initial={{ pathLength: 0, opacity: 0 }}
        animate={{ pathLength: 1, opacity: 1 }}
        transition={{ duration: 1.5, delay: 0.6, ease: [0.16, 1, 0.3, 1] }}
      />

      {/* Left swoosh */}
      <motion.path
        d="M32 18 C24 18, 16 24, 14 34"
        stroke="url(#vortex-grad-2)"
        strokeWidth="2"
        strokeLinecap="round"
        fill="none"
        initial={{ pathLength: 0, opacity: 0 }}
        animate={{ pathLength: 1, opacity: 1 }}
        transition={{ duration: 1.5, delay: 0.6, ease: [0.16, 1, 0.3, 1] }}
      />

      {/* Center dot */}
      <motion.circle
        cx="32"
        cy="32"
        r="3"
        fill="url(#vortex-grad)"
        initial={{ scale: 0 }}
        animate={{ scale: 1 }}
        transition={{ duration: 0.6, delay: 1.2, ease: [0.16, 1, 0.3, 1] }}
      />
    </svg>
  );
}

// ─── Progress bar with glow ────────────────────────────────────

function ProgressBar({ value }: { value: number }) {
  return (
    <div className="relative w-56">
      {/* Track */}
      <div className="h-[3px] w-full rounded-full bg-white/5 overflow-hidden">
        {/* Fill */}
        <motion.div
          className="h-full rounded-full"
          style={{
            background: "linear-gradient(90deg, hsl(239, 84%, 74%), hsl(198, 93%, 60%), hsl(187, 100%, 42%))",
            boxShadow: "0 0 8px hsl(198, 93%, 60% / 0.5), 0 0 20px hsl(198, 93%, 60% / 0.2)",
          }}
          initial={{ width: "0%" }}
          animate={{ width: `${value}%` }}
          transition={{ duration: 0.4, ease: [0.16, 1, 0.3, 1] }}
        />
      </div>
      {/* Glow underneath */}
      <div
        className="absolute inset-0 h-full rounded-full blur-sm opacity-30"
        style={{
          background: "linear-gradient(90deg, transparent, hsl(198, 93%, 60%), transparent)",
        }}
      />
    </div>
  );
}

// ─── Stage indicator text ──────────────────────────────────────

function StageText({ text, progress }: { text: string; progress: number }) {
  return (
    <div className="flex flex-col items-center gap-2">
      <motion.span
        key={text}
        initial={{ opacity: 0, y: 6 }}
        animate={{ opacity: 1, y: 0 }}
        className="text-xs font-semibold tracking-[0.15em] uppercase text-fg-subtle"
      >
        {text}
      </motion.span>
      <span className="text-[10px] font-mono text-fg-muted tabular-nums tracking-wider">
        {progress}%
      </span>
    </div>
  );
}

// ════════════════════════════════════════════════════════════════
// SPLASH SCREEN
// ════════════════════════════════════════════════════════════════

interface SplashScreenProps {
  onFinish?: () => void;
}

export function SplashScreen({ onFinish }: SplashScreenProps) {
  const { loading: authLoading } = useAuth();
  const [progress, setProgress] = useState(0);
  const [stageIndex, setStageIndex] = useState(0);
  const [exiting, setExiting] = useState(false);
  const [show, setShow] = useState(true);
  const startedAt = useRef(Date.now());

  // ─── Progress simulation ────────────────────────────────────
  useEffect(() => {
    if (exiting) return;

    const totalDuration = STAGES.reduce((s, s2) => s + s2.duration, 0);
    let elapsed = 0;

    const interval = setInterval(() => {
      elapsed += 50;
      const raw = (elapsed / totalDuration) * 100;
      const capped = Math.min(raw, 95); // never reach 100% until auth done
      setProgress(capped);

      // Update stage
      let accum = 0;
      for (let i = 0; i < STAGES.length; i++) {
        accum += STAGES[i].duration;
        if (elapsed <= accum) {
          setStageIndex(i);
          break;
        }
      }

      if (elapsed >= totalDuration) {
        clearInterval(interval);
      }
    }, 50);

    return () => clearInterval(interval);
  }, [exiting]);

  // ─── Auth + minimum time → finish ──────────────────────────
  useEffect(() => {
    if (authLoading || exiting) return;

    const elapsed = Date.now() - startedAt.current;
    const minTime = 2200; // minimum splash duration
    const remaining = Math.max(0, minTime - elapsed);

    const timeout = setTimeout(() => {
      setProgress(100);
      setStageIndex(STAGES.length - 1);

      // Small delay so the user sees "Ready" + 100%
      setTimeout(() => {
        setExiting(true);
        setTimeout(() => {
          setShow(false);
          onFinish?.();
        }, 600);
      }, 400);
    }, remaining);

    return () => clearTimeout(timeout);
  }, [authLoading, exiting, onFinish]);

  return (
    <AnimatePresence>
      {show && (
        <motion.div
          key="splash"
          initial={{ opacity: 1 }}
          exit={{ opacity: 0, scale: 1.02 }}
          transition={{ duration: 0.6, ease: [0.16, 1, 0.3, 1] }}
          className="fixed inset-0 z-[9999] flex flex-col items-center justify-center overflow-hidden"
          style={{
            backgroundColor: "hsl(228, 33%, 4%)",
          }}
        >
          {/* ─── Animated dot pattern ───────────────────────── */}
          <div className="absolute inset-0 bg-dot-pattern-sm animate-pattern-drift opacity-[0.06] pointer-events-none" />

          {/* ─── Ambient glow blobs ──────────────────────────── */}
          <div
            className="absolute -top-40 -start-40 h-80 w-80 rounded-full opacity-[0.08]"
            style={{
              background: "radial-gradient(circle, hsl(239, 84%, 74%), transparent 70%)",
              filter: "blur(60px)",
            }}
          />
          <div
            className="absolute -bottom-40 -end-40 h-80 w-80 rounded-full opacity-[0.06]"
            style={{
              background: "radial-gradient(circle, hsl(187, 100%, 42%), transparent 70%)",
              filter: "blur(60px)",
            }}
          />

          {/* ─── Grid overlay ─────────────────────────────────── */}
          <div className="absolute inset-0 bg-grid-pattern opacity-[0.015] pointer-events-none" />

          {/* ─── Content ─────────────────────────────────────── */}
          <motion.div
            className="relative flex flex-col items-center gap-10"
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.8, ease: [0.16, 1, 0.3, 1] }}
          >
            {/* Vortex logo */}
            <div>
              <VortexLogo size={88} />
            </div>

            {/* Brand text */}
            <div className="flex flex-col items-center gap-1">
              <motion.h1
                className="text-3xl font-black tracking-[0.08em]"
                style={{
                  fontFamily: "'Orbitron', sans-serif",
                  background:
                    "linear-gradient(135deg, hsl(239, 84%, 74%), hsl(198, 93%, 60%), hsl(187, 100%, 42%))",
                  WebkitBackgroundClip: "text",
                  WebkitTextFillColor: "transparent",
                  backgroundClip: "text",
                }}
                initial={{ opacity: 0, y: 10 }}
                animate={{ opacity: 1, y: 0 }}
                transition={{ duration: 0.6, delay: 0.3 }}
              >
                VORTEXUI
              </motion.h1>
              <motion.p
                className="text-[10px] font-medium tracking-[0.25em] uppercase text-fg-muted/60"
                initial={{ opacity: 0 }}
                animate={{ opacity: 1 }}
                transition={{ duration: 0.6, delay: 0.6 }}
              >
                Control Panel
              </motion.p>
            </div>

            {/* Progress section */}
            <motion.div
              className="flex flex-col items-center gap-3"
              initial={{ opacity: 0 }}
              animate={{ opacity: 1 }}
              transition={{ duration: 0.6, delay: 0.9 }}
            >
              <ProgressBar value={progress} />
              <StageText
                text={STAGES[Math.min(stageIndex, STAGES.length - 1)].label}
                progress={progress}
              />
            </motion.div>

            {/* Version */}
            <motion.p
            className="absolute bottom-0 left-1/2 -translate-x-1/2 translate-y-24 text-[10px] font-mono text-fg-subtle/30 tracking-wider"
              initial={{ opacity: 0 }}
              animate={{ opacity: 1 }}
              transition={{ duration: 0.6, delay: 1.2 }}
            >
              v{__PANEL_VERSION__ || "2.0.0"}
            </motion.p>
          </motion.div>
        </motion.div>
      )}
    </AnimatePresence>
  );
}
