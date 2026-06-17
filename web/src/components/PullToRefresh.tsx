import { useState, useRef } from "react";
import { cn } from "@/lib/utils";
import { RefreshCw } from "lucide-react";

interface PullToRefreshProps {
  onRefresh: () => Promise<void>;
  children: React.ReactNode;
  className?: string;
}

/**
 * Pull-to-refresh wrapper for mobile portal pages.
 * Shows a spinner when pulled down past threshold.
 */
export function PullToRefresh({ onRefresh, children, className }: PullToRefreshProps) {
  const [pulling, setPulling] = useState(false);
  const [refreshing, setRefreshing] = useState(false);
  const [pullDistance, setPullDistance] = useState(0);
  const startY = useRef(0);
  const containerRef = useRef<HTMLDivElement>(null);

  const threshold = 60;

  function handleTouchStart(e: React.TouchEvent) {
    if (containerRef.current && containerRef.current.scrollTop === 0) {
      startY.current = e.touches[0].clientY;
      setPulling(true);
    }
  }

  function handleTouchMove(e: React.TouchEvent) {
    if (!pulling) return;
    const diff = e.touches[0].clientY - startY.current;
    if (diff > 0) {
      setPullDistance(Math.min(diff * 0.5, threshold * 1.5));
    }
  }

  async function handleTouchEnd() {
    if (pullDistance >= threshold && !refreshing) {
      setRefreshing(true);
      setPullDistance(threshold);
      try {
        await onRefresh();
      } finally {
        setRefreshing(false);
      }
    }
    setPulling(false);
    setPullDistance(0);
  }

  return (
    <div
      ref={containerRef}
      className={cn("relative overflow-auto", className)}
      onTouchStart={handleTouchStart}
      onTouchMove={handleTouchMove}
      onTouchEnd={handleTouchEnd}
    >
      {/* Pull indicator */}
      <div
        className="flex items-center justify-center overflow-hidden transition-all"
        style={{ height: pullDistance }}
      >
        <RefreshCw
          size={20}
          className={cn(
            "text-primary transition-transform",
            refreshing && "animate-spin",
            pullDistance >= threshold && "scale-110",
          )}
          style={{ transform: `rotate(${pullDistance * 3}deg)` }}
        />
      </div>

      {children}
    </div>
  );
}
