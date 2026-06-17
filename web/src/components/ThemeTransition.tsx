import { useTheme } from "@/theme/theme";
import { Moon, Sun } from "lucide-react";
import { cn } from "@/lib/utils";

/**
 * Enhanced theme toggle with smooth icon rotation and morph transition.
 * The actual color transitions are handled via CSS (already set on :root).
 */
export function ThemeToggle({ className }: { className?: string }) {
  const { resolved, toggle } = useTheme();
  const isDark = resolved === "dark";

  return (
    <button
      onClick={toggle}
      aria-label={isDark ? "Switch to light mode" : "Switch to dark mode"}
      className={cn(
        "relative grid h-9 w-9 place-items-center rounded-xl text-fg-muted transition-all duration-300 hover:bg-surface-2/60 hover:text-fg",
        className,
      )}
    >
      <div className="relative h-5 w-5">
        <Sun
          size={18}
          className={cn(
            "absolute inset-0 m-auto transition-all duration-500",
            isDark ? "rotate-90 scale-0 opacity-0" : "rotate-0 scale-100 opacity-100",
          )}
        />
        <Moon
          size={18}
          className={cn(
            "absolute inset-0 m-auto transition-all duration-500",
            isDark ? "rotate-0 scale-100 opacity-100" : "-rotate-90 scale-0 opacity-0",
          )}
        />
      </div>
    </button>
  );
}
