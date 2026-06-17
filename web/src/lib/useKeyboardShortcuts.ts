import { useEffect } from "react";
import { useNavigate } from "react-router-dom";

/**
 * Global keyboard shortcuts:
 * - Ctrl+K / ⌘K: Command palette (handled by CommandPalette)
 * - n: Navigate to new user (when not in input)
 * - s: Focus search (when not in input)
 * - ?: Show shortcuts help
 */
export function useKeyboardShortcuts() {
  const navigate = useNavigate();

  useEffect(() => {
    function handler(e: KeyboardEvent) {
      // Skip if user is typing in an input/textarea
      const tag = (e.target as HTMLElement).tagName;
      if (tag === "INPUT" || tag === "TEXTAREA" || tag === "SELECT") return;
      if (e.metaKey || e.ctrlKey || e.altKey) return;

      switch (e.key) {
        case "n":
          e.preventDefault();
          navigate("/users");
          break;
        case "s":
          e.preventDefault();
          // Trigger command palette
          window.dispatchEvent(new KeyboardEvent("keydown", { key: "k", metaKey: true }));
          break;
        case "?":
          e.preventDefault();
          window.dispatchEvent(new CustomEvent("vortex:show-shortcuts"));
          break;
      }
    }
    window.addEventListener("keydown", handler);
    return () => window.removeEventListener("keydown", handler);
  }, [navigate]);
}
