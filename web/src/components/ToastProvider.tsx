import { createContext, useCallback, useContext, useState } from "react";
import { motion, AnimatePresence } from "framer-motion";
import { cn } from "@/lib/utils";
import { CheckCircle2, XCircle, Info, X, AlertTriangle } from "lucide-react";

type ToastType = "success" | "error" | "info" | "warning";

interface Toast {
  id: number;
  type: ToastType;
  message: string;
  duration: number;
  undoFn?: () => void;
}

interface ToastAPI {
  success: (message: string, opts?: { undo?: () => void }) => void;
  error: (message: string) => void;
  info: (message: string) => void;
  warning: (message: string) => void;
}

const ToastContext = createContext<ToastAPI>({
  success: () => {},
  error: () => {},
  info: () => {},
  warning: () => {},
});

export function useToastV2() {
  return useContext(ToastContext);
}

let _id = 0;

export function ToastProviderV2({ children }: { children: React.ReactNode }) {
  const [toasts, setToasts] = useState<Toast[]>([]);

  const push = useCallback((type: ToastType, message: string, opts?: { undo?: () => void; duration?: number }) => {
    const id = ++_id;
    const duration = opts?.duration ?? (type === "error" ? 6000 : 4000);
    setToasts(prev => [...prev, { id, type, message, duration, undoFn: opts?.undo }]);
    setTimeout(() => setToasts(prev => prev.filter(t => t.id !== id)), duration);
  }, []);

  const api: ToastAPI = {
    success: (msg, opts) => push("success", msg, opts),
    error: (msg) => push("error", msg),
    info: (msg) => push("info", msg),
    warning: (msg) => push("warning", msg),
  };

  function dismiss(id: number) {
    setToasts(prev => prev.filter(t => t.id !== id));
  }

  return (
    <ToastContext.Provider value={api}>
      {children}
      {/* Toast container */}
      <div className="fixed bottom-4 end-4 z-[200] flex flex-col-reverse gap-2 w-80 pointer-events-none">
        <AnimatePresence>
          {toasts.map(toast => (
            <ToastItem key={toast.id} toast={toast} onDismiss={() => dismiss(toast.id)} />
          ))}
        </AnimatePresence>
      </div>
    </ToastContext.Provider>
  );
}

const ICONS: Record<ToastType, React.ReactNode> = {
  success: <CheckCircle2 size={18} />,
  error: <XCircle size={18} />,
  info: <Info size={18} />,
  warning: <AlertTriangle size={18} />,
};

const COLORS: Record<ToastType, string> = {
  success: "border-success/30 text-success bg-success/5",
  error: "border-danger/30 text-danger bg-danger/5",
  info: "border-accent/30 text-accent bg-accent/5",
  warning: "border-warning/30 text-warning bg-warning/5",
};

function ToastItem({ toast, onDismiss }: { toast: Toast; onDismiss: () => void }) {
  return (
    <motion.div
      initial={{ opacity: 0, x: 60, scale: 0.92 }}
      animate={{ opacity: 1, x: 0, scale: 1 }}
      exit={{ opacity: 0, x: 60, scale: 0.92, transition: { duration: 0.15 } }}
      transition={{ type: "spring", stiffness: 400, damping: 28 }}
      layout
      className={cn(
        "pointer-events-auto flex items-start gap-3 rounded-xl border px-4 py-3 shadow-lg backdrop-blur-xl",
        COLORS[toast.type],
      )}
    >
      <div className="mt-0.5">{ICONS[toast.type]}</div>
      <div className="flex-1 min-w-0">
        <p className="text-sm font-medium text-fg">{toast.message}</p>
        {toast.undoFn && (
          <button onClick={() => { toast.undoFn?.(); onDismiss(); }} className="mt-1 text-xs font-semibold underline text-primary hover:text-primary-hover">
            Undo
          </button>
        )}
      </div>
      <button onClick={onDismiss} className="grid h-5 w-5 place-items-center rounded-md text-fg-subtle hover:text-fg transition">
        <X size={12} />
      </button>
      {/* Progress bar */}
      <div className="absolute bottom-0 left-0 right-0 h-0.5 overflow-hidden rounded-b-xl">
        <div className="h-full bg-current opacity-30 animate-progress" style={{ animationDuration: `${toast.duration}ms` }} />
      </div>
    </motion.div>
  );
}
