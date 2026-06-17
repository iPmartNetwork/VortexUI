import { useLocation } from "react-router-dom";

/**
 * Wraps page content with a CSS-based fade+slide transition.
 * Uses key={pathname} to trigger re-mount animation on route changes.
 * No external deps needed — pure CSS animation.
 */
export function PageTransition({ children }: { children: React.ReactNode }) {
  const location = useLocation();
  return (
    <div key={location.pathname} className="animate-page-enter">
      {children}
    </div>
  );
}
