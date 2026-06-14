import { useState } from "react";
import { Check, Copy } from "lucide-react";
import { cn } from "@/lib/utils";

// CopyField shows a value with a one-click copy button and a brief confirmation.
export function CopyField({ value, mono = true, className }: { value: string; mono?: boolean; className?: string }) {
  const [done, setDone] = useState(false);
  function copy() {
    navigator.clipboard?.writeText(value);
    setDone(true);
    setTimeout(() => setDone(false), 1200);
  }
  return (
    <div className={cn("flex items-center gap-2", className)}>
      <div
        className={cn(
          "min-w-0 flex-1 truncate rounded-lg border border-white/[0.08] bg-white/[0.03] px-3 py-2 text-xs text-fg-muted",
          mono && "font-mono",
        )}
        title={value}
        dir="ltr"
      >
        {value}
      </div>
      <button
        onClick={copy}
        aria-label="copy"
        className="grid h-8 w-9 shrink-0 place-items-center rounded-lg border border-white/[0.08] text-fg-muted transition hover:bg-white/[0.06] hover:text-fg"
      >
        {done ? <Check size={15} className="text-success" /> : <Copy size={15} />}
      </button>
    </div>
  );
}
