import { cn } from "@/lib/utils";
import { X } from "lucide-react";

interface BottomSheetProps {
  open: boolean;
  onClose: () => void;
  title?: string;
  children: React.ReactNode;
  className?: string;
}

/**
 * Mobile-friendly bottom sheet modal that slides up from the bottom.
 * Better for mobile UX than center modals.
 */
export function BottomSheet({ open, onClose, title, children, className }: BottomSheetProps) {
  if (!open) return null;

  return (
    <div className="fixed inset-0 z-[100]" onClick={onClose}>
      {/* Backdrop */}
      <div className="absolute inset-0 bg-black/50 backdrop-blur-sm animate-fade-in" />

      {/* Sheet */}
      <div
        className={cn(
          "absolute bottom-0 inset-x-0 max-h-[85vh] rounded-t-3xl border-t border-border/50 bg-bg-elevated/98 shadow-2xl backdrop-blur-xl overflow-hidden animate-slide-up safe-bottom",
          className,
        )}
        onClick={e => e.stopPropagation()}
      >
        {/* Handle bar */}
        <div className="flex justify-center pt-3 pb-1">
          <div className="h-1 w-10 rounded-full bg-border/60" />
        </div>

        {/* Header */}
        {title && (
          <div className="flex items-center justify-between px-5 py-2 border-b border-border/30">
            <h2 className="text-base font-semibold text-fg">{title}</h2>
            <button onClick={onClose} className="grid h-8 w-8 place-items-center rounded-xl text-fg-subtle hover:bg-surface-2/60 hover:text-fg">
              <X size={16} />
            </button>
          </div>
        )}

        {/* Content */}
        <div className="overflow-y-auto px-5 py-4 max-h-[70vh]">
          {children}
        </div>
      </div>
    </div>
  );
}
