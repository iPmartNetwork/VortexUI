import { useEffect, useRef, useState, useMemo } from "react";
import { getToken } from "@/api/client";
import { useOverview } from "@/api/policy-hooks";
import { useTrafficSeries, type TrafficRange } from "@/api/policy-hooks";

// ══════════════════════════════════════════════════════════════════
// Types
// ══════════════════════════════════════════════════════════════════

export interface LiveTrafficPoint {
  time: number;
  up: number;
  down: number;
}

export type ConnectionStatus = "connecting" | "live" | "polling" | "error";

export interface LiveTrafficReturn {
  /** Smoothed data points for chart rendering */
  chartPoints: LiveTrafficPoint[];
  /** Current download speed (bytes/sec), EMA-smoothed */
  currentDownSpeed: number;
  /** Current upload speed (bytes/sec), EMA-smoothed */
  currentUpSpeed: number;
  /** Peak bandwidth observed (bytes/sec) */
  peakBandwidth: number;
  /** Total download this session (bytes) */
  totalDown: number;
  /** Total upload this session (bytes) */
  totalUp: number;
  /** SSE / polling connection status */
  connectionStatus: ConnectionStatus;
  /** True when SSE is actively connected */
  isLive: boolean;
  /** Latest bandwidth delta (bytes/sec) for sparkline */
  speedDelta: number;
  /** Human-readable speed string */
  speedLabel: string;
}

// ══════════════════════════════════════════════════════════════════
// Constants
// ══════════════════════════════════════════════════════════════════

const MAX_CHART_POINTS = 60;
const SMOOTHING_ALPHA = 0.2; // EMA factor — lower = smoother
const IDLE_THRESHOLD = 0.5; // bytes/sec below which we show "idle"

// ══════════════════════════════════════════════════════════════════
// Helpers
// ══════════════════════════════════════════════════════════════════

function fmtSpeed(bytesPerSec: number): string {
  if (bytesPerSec < IDLE_THRESHOLD) return "Idle";
  if (bytesPerSec >= 1_000_000) return `${(bytesPerSec / 1_000_000).toFixed(2)} MB/s`;
  if (bytesPerSec >= 1_000) return `${(bytesPerSec / 1_000).toFixed(1)} KB/s`;
  return `${bytesPerSec.toFixed(0)} B/s`;
}

// ══════════════════════════════════════════════════════════════════
// Hook
// ══════════════════════════════════════════════════════════════════

export function useLiveTraffic(range: TrafficRange = "24h"): LiveTrafficReturn {
  const { data: overview } = useOverview();
  const trafficSeries = useTrafficSeries(range);

  // ── State ──────────────────────────────────────────────────────
  const [connStatus, setConnStatus] = useState<ConnectionStatus>("connecting");
  const [peakBw, setPeakBw] = useState(0);

  // Smooth animation buffer — updated per rAF frame
  const [displayDown, setDisplayDown] = useState(0);
  const [displayUp, setDisplayUp] = useState(0);
  const [displayPoints, setDisplayPoints] = useState<LiveTrafficPoint[]>([]);

  // Refs for calculation
  const smoothDownRef = useRef(0);
  const smoothUpRef = useRef(0);
  const lastTotalRef = useRef(0);
  const lastTimeRef = useRef(Date.now());
  const pointsRef = useRef<LiveTrafficPoint[]>([]);
  const peakRef = useRef(0);
  const rafRef = useRef<number>(0);
  const targetDownRef = useRef(0);
  const targetUpRef = useRef(0);
  const sseRef = useRef<EventSource | null>(null);

  // ── SSE connection ─────────────────────────────────────────────
  useEffect(() => {
    const token = getToken();
    if (!token) {
      setConnStatus("polling");
      return;
    }

    let mounted = true;

    const tryConnect = () => {
      try {
        const es = new EventSource(
          `/api/events/stream?access_token=${encodeURIComponent(token)}`,
        );

        // Handle any SSE message that might contain traffic data
        es.addEventListener("traffic", (ev: MessageEvent) => {
          if (!mounted) return;
          try {
            const d = JSON.parse(ev.data) as {
              up?: number;
              down?: number;
              total_up?: number;
              total_down?: number;
            };
            const now = Date.now();
            const down = d.down ?? 0;
            const up = d.up ?? 0;
            const totalDown = d.total_down ?? 0;
            const totalUp = d.total_up ?? 0;

            // Update peak
            const total = down + up;
            if (total > peakRef.current) {
              peakRef.current = total;
              setPeakBw(total);
            }

            // Set target for smooth animation
            targetDownRef.current = down;
            targetUpRef.current = up;

            // Add data point
            const pt: LiveTrafficPoint = { time: now, up, down };
            pointsRef.current = [...pointsRef.current.slice(-(MAX_CHART_POINTS - 1)), pt];

            lastTotalRef.current = totalDown + totalUp;
            lastTimeRef.current = now;
          } catch {
            // ignore malformed
          }
        });

        es.addEventListener("open", () => {
          if (!mounted) return;
          setConnStatus("live");
        });

        es.onerror = () => {
          if (!mounted) return;
          es.close();
          setConnStatus("polling");
        };

        sseRef.current = es;
      } catch {
        if (mounted) setConnStatus("polling");
      }
    };

    tryConnect();

    return () => {
      mounted = false;
      sseRef.current?.close();
    };
  }, []);

  // ── Process overview polling data ──────────────────────────────
  useEffect(() => {
    if (connStatus === "live") return; // SSE already handles it
    if (!overview?.users?.total_used && overview?.users?.total_used !== 0) return;

    const now = Date.now();
    const totalBytes = overview.users.total_used;

    if (lastTotalRef.current > 0) {
      const elapsed = (now - lastTimeRef.current) / 1000;
      if (elapsed > 0.1) {
        const delta = Math.max(0, totalBytes - lastTotalRef.current);
        const speed = delta / elapsed; // bytes/sec

        // Estimate up/down split (70/30 typical for proxy traffic)
        const downSpeed = speed * 0.7;
        const upSpeed = speed * 0.3;

        // Update target for smooth animation
        targetDownRef.current = downSpeed;
        targetUpRef.current = upSpeed;

        // Track peak
        if (speed > peakRef.current) {
          peakRef.current = speed;
          setPeakBw(speed);
        }

        // Add data point
        const pt: LiveTrafficPoint = {
          time: now,
          up: upSpeed,
          down: downSpeed,
        };
        pointsRef.current = [...pointsRef.current.slice(-(MAX_CHART_POINTS - 1)), pt];
      }
    }

    lastTotalRef.current = totalBytes;
    lastTimeRef.current = now;
  }, [overview, connStatus]);

  // ── Use traffic series data as base for chart ──────────────────
  useEffect(() => {
    const series = trafficSeries.data?.points ?? [];
    if (series.length === 0) return;

    const converted: LiveTrafficPoint[] = series.map((p) => ({
      time: new Date(p.time).getTime(),
      up: p.up,
      down: p.down,
    }));

    // Merge: keep recent real-time points + historical series
    const realtime = pointsRef.current;
    if (realtime.length > 0) {
      const latestRealtime = realtime[realtime.length - 1].time;
      const cutoff = latestRealtime - 120_000; // 2 min window
      const historical = converted.filter((p) => p.time < cutoff);
      pointsRef.current = [...historical.slice(-20), ...realtime];
    } else {
      pointsRef.current = converted.slice(-MAX_CHART_POINTS);
    }
  }, [trafficSeries.data]);

  // ── Smooth animation loop ──────────────────────────────────────
  useEffect(() => {
    let running = true;

    const tick = () => {
      if (!running) return;

      // EMA towards target
      smoothDownRef.current =
        smoothDownRef.current === 0
          ? targetDownRef.current
          : smoothDownRef.current * (1 - SMOOTHING_ALPHA) +
            targetDownRef.current * SMOOTHING_ALPHA;

      smoothUpRef.current =
        smoothUpRef.current === 0
          ? targetUpRef.current
          : smoothUpRef.current * (1 - SMOOTHING_ALPHA) +
            targetUpRef.current * SMOOTHING_ALPHA;

      // Clamp near-zero to 0 for cleaner display
      const d = Math.abs(smoothDownRef.current) < 0.01 ? 0 : smoothDownRef.current;
      const u = Math.abs(smoothUpRef.current) < 0.01 ? 0 : smoothUpRef.current;

      setDisplayDown(d);
      setDisplayUp(u);
      setDisplayPoints([...pointsRef.current]);

      rafRef.current = requestAnimationFrame(tick);
    };

    rafRef.current = requestAnimationFrame(tick);

    return () => {
      running = false;
      cancelAnimationFrame(rafRef.current);
    };
  }, []);

  // ── Memoised return ────────────────────────────────────────────
  return useMemo(
    () => ({
      chartPoints: displayPoints,
      currentDownSpeed: displayDown,
      currentUpSpeed: displayUp,
      peakBandwidth: peakBw,
      totalDown: lastTotalRef.current * 0.7,
      totalUp: lastTotalRef.current * 0.3,
      connectionStatus: connStatus,
      isLive: connStatus === "live",
      speedDelta: displayDown + displayUp,
      speedLabel: fmtSpeed(displayDown + displayUp),
    }),
    [displayPoints, displayDown, displayUp, peakBw, connStatus],
  );
}
