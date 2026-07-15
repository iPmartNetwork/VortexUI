import { useEffect, useRef, useState } from "react";
import { cn } from "@/lib/utils";
import { useInView } from "@/lib/useInView";

interface AnimatedCounterProps {
  value: number;
  duration?: number;
  suffix?: string;
  prefix?: string;
  decimals?: number;
  className?: string;
  formatter?: (value: number) => string;
}

/**
 * Animates a number from 0 to the target value using requestAnimationFrame.
 * Starts only when the element enters the viewport (lazy).
 */
export function AnimatedCounter({
  value,
  duration = 800,
  suffix,
  prefix,
  decimals = 0,
  className,
  formatter,
}: AnimatedCounterProps) {
  const [display, setDisplay] = useState(0);
  const ref = useRef<HTMLSpanElement>(null);
  const inView = useInView(ref, { once: true });
  const started = useRef(false);

  useEffect(() => {
    if (!inView || started.current) return;
    started.current = true;

    const startTime = performance.now();
    const from = 0;
    const diff = value - from;

    function tick(now: number) {
      const elapsed = now - startTime;
      const progress = Math.min(elapsed / duration, 1);
      // Ease-out cubic
      const ease = 1 - Math.pow(1 - progress, 3);
      setDisplay(from + diff * ease);

      if (progress < 1) {
        requestAnimationFrame(tick);
      }
    }

    requestAnimationFrame(tick);
  }, [inView, value, duration]);

  // Reset if value changes after initial animation
  useEffect(() => {
    started.current = false;
  }, [value]);

  function format(n: number): string {
    if (formatter) return formatter(n);
    if (decimals > 0) return n.toFixed(decimals);
    return Math.round(n).toLocaleString();
  }

  return (
    <span ref={ref} className={cn("tabular-nums", className)}>
      {prefix ?? ""}
      {format(display)}
      {suffix ?? ""}
    </span>
  );
}
