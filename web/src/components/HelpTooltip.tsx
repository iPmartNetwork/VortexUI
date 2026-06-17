import { useState } from "react";
import { HelpCircle } from "lucide-react";
import { cn } from "@/lib/utils";

interface HelpTooltipProps {
  text: string;
  position?: "top" | "bottom" | "left" | "right";
  className?: string;
}

export function HelpTooltip({ text, position = "top", className }: HelpTooltipProps) {
  const [show, setShow] = useState(false);

  const positionClasses = {
    top: "bottom-full left-1/2 -translate-x-1/2 mb-2",
    bottom: "top-full left-1/2 -translate-x-1/2 mt-2",
    left: "right-full top-1/2 -translate-y-1/2 mr-2",
    right: "left-full top-1/2 -translate-y-1/2 ml-2",
  };

  return (
    <span className={cn("relative inline-flex", className)}>
      <button
        type="button"
        onMouseEnter={() => setShow(true)}
        onMouseLeave={() => setShow(false)}
        onFocus={() => setShow(true)}
        onBlur={() => setShow(false)}
        className="inline-grid h-4 w-4 place-items-center rounded-full text-fg-subtle/60 transition hover:text-fg-muted"
        aria-label="Help"
      >
        <HelpCircle size={13} />
      </button>
      {show && (
        <div className={cn(
          "absolute z-50 w-52 rounded-lg border border-border/50 bg-bg-elevated/95 px-3 py-2 text-[11px] text-fg-muted shadow-xl backdrop-blur-xl animate-fade-in",
          positionClasses[position],
        )}>
          {text}
        </div>
      )}
    </span>
  );
}
