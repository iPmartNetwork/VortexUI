import { createContext, useCallback, useContext, useState } from "react";
import { cn } from "@/lib/utils";
import { CheckCircle2, XCircle, Info } from "lucide-react";

type ToastKind = "success" | "error" | "info";
interface Toast {
  id: number;
  kind: ToastKind;
  message: string;
  leaving?: boolean;
}

const ToastContext = createContext<((kind: ToastKind, message: string) => void) | null>(null);

let nextId = 1;

export function ToastProvider({ children }: { children: React.ReactNode }) {
  const [toasts, setToasts] = useState<Toast[]>([]);

  const push = useCallback((kind: ToastKind, message: string) => {
    const id = nextId++;
    setToasts((t) => [...t, { id, kind, message }]);
    // Start exit animation then remove
    setTimeout(() => setToasts((t) => t.map((x) => x.id === id ? { ...x, leaving: true } : x)), 3000);
    setTimeout(() => setToasts((t) => t.filter((x) => x.id !== id)), 3400);
  }, []);

  return (
    <ToastContext.Provider value={push}>
      {children}
      <div className="fixed end-4 bottom-4 z-[100] flex w-80 flex-col-reverse gap-2">
        {toasts.map((t) => (
          <ToastItem key={t.id} toast={t} onDismiss={() => setToasts((ts) => ts.filter((x) => x.id !== t.id))} />
        ))}
      </div>
    </ToastContext.Provider>
  );
}

function ToastItem({ toast: t, onDismiss }: { toast: Toast; onDismiss: () => void }) {
  const icons = { success: <CheckCircle2 size={16} />, error: <XCircle size={16} />, info: <Info size={16} /> };
  return (
    <div
      onClick={onDismiss}
      className={cn(
        "card flex cursor-pointer items-center gap-3 px-4 py-3 text-sm shadow-xl ring-1 transition-all duration-300",
        t.leaving ? "translate-x-full opacity-0" : "translate-x-0 opacity-100 animate-slide-in-right",
        t.kind === "success" && "ring-success/30 text-success",
        t.kind === "error" && "ring-danger/30 text-danger",
        t.kind === "info" && "ring-border/50 text-fg-muted",
      )}
    >
      {icons[t.kind]}
      <span className="flex-1 text-fg">{t.message}</span>
    </div>
  );
}

export function useToast() {
  const push = useContext(ToastContext);
  if (!push) throw new Error("useToast must be used within ToastProvider");
  return {
    success: (m: string) => push("success", m),
    error: (m: string) => push("error", m),
    info: (m: string) => push("info", m),
  };
}
