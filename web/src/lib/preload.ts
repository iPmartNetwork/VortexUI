/**
 * Route preloader — proactively fetches lazy-loaded page chunks so navigation
 * feels instant. Use `onMouseEnter` on sidebar/tile elements to trigger
 * preloading before the user clicks.
 *
 * Usage:
 *   import { preloadRoute } from "@/lib/preload";
 *   // In a nav link's onMouseEnter:
 *   onMouseEnter={() => preloadRoute(() => import("@/pages/Users"))}
 */

const preloaded = new Set<string>();

/**
 * Preload a lazy chunk by calling its import function ahead of time.
 * Safe to call repeatedly — the chunk is only fetched once.
 */
export function preloadRoute(importFn: () => Promise<unknown>): void {
  const key = importFn.toString();
  if (preloaded.has(key)) return;
  preloaded.add(key);
  // Trigger the dynamic import silently. React.lazy already caches the
  // resulting module, so the next time the route renders it resolves
  // synchronously from the prefetched chunk.
  importFn().catch(() => { /* ignore preload failures */ });
}

/**
 * Preload a batch of routes, e.g. the most-visited pages on app mount.
 */
export function preloadRoutes(
  fns: (() => Promise<unknown>)[],
): void {
  // Yield to the main thread between each preload so the initial render
  // is not blocked by a burst of network requests.
  let i = 0;
  const next = () => {
    if (i >= fns.length) return;
    preloadRoute(fns[i]);
    i++;
    requestIdleCallback ? requestIdleCallback(next, { timeout: 500 }) : setTimeout(next, 200);
  };
  next();
}
