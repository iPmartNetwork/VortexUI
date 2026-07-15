import { useEffect, useRef, useState } from "react";

interface UseInViewOptions {
  /** Only fire once, then disconnect observer. */
  once?: boolean;
  /** Root margin (same as IntersectionObserver). */
  margin?: string;
  /** Threshold (same as IntersectionObserver). */
  threshold?: number;
}

/**
 * Tracks whether a DOM element is visible in the viewport using IntersectionObserver.
 * Falls back to true when IntersectionObserver is not available (SSR).
 */
export function useInView<T extends HTMLElement = HTMLDivElement>(
  ref: React.RefObject<T | null>,
  options: UseInViewOptions = {},
): boolean {
  const [inView, setInView] = useState(false);
  const onceRef = useRef(options.once ?? false);

  useEffect(() => {
    const el = ref.current;
    if (!el) return;

    if (typeof IntersectionObserver === "undefined") {
      setInView(true);
      return;
    }

    const observer = new IntersectionObserver(
      ([entry]) => {
        if (entry.isIntersecting) {
          setInView(true);
          if (onceRef.current) {
            observer.unobserve(el);
          }
        } else if (!onceRef.current) {
          setInView(false);
        }
      },
      {
        rootMargin: options.margin ?? "0px",
        threshold: options.threshold ?? 0,
      },
    );

    observer.observe(el);
    return () => observer.disconnect();
  }, [ref, options.margin, options.threshold]);

  return inView;
}
