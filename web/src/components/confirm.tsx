import { createContext, useCallback, useContext, useRef, useState } from "react";
import { Button } from "./ui";
import { Modal } from "./Modal";

interface ConfirmOptions {
  title: string;
  message?: string;
  confirmLabel?: string;
  destructive?: boolean;
}

const ConfirmContext = createContext<((o: ConfirmOptions) => Promise<boolean>) | null>(null);

// ConfirmProvider exposes a promise-based confirm() — cleaner than window.confirm
// and styled like the rest of the panel.
export function ConfirmProvider({ children }: { children: React.ReactNode }) {
  const [opts, setOpts] = useState<ConfirmOptions | null>(null);
  const resolver = useRef<(v: boolean) => void>();

  const confirm = useCallback((o: ConfirmOptions) => {
    setOpts(o);
    return new Promise<boolean>((resolve) => {
      resolver.current = resolve;
    });
  }, []);

  function settle(v: boolean) {
    resolver.current?.(v);
    setOpts(null);
  }

  return (
    <ConfirmContext.Provider value={confirm}>
      {children}
      {opts && (
        <Modal open onClose={() => settle(false)} title={opts.title} className="max-w-sm">
          {opts.message && <p className="mb-4 text-sm text-muted-foreground">{opts.message}</p>}
          <div className="flex justify-end gap-2">
            <Button variant="ghost" onClick={() => settle(false)}>
              Cancel
            </Button>
            <Button variant={opts.destructive ? "destructive" : "primary"} onClick={() => settle(true)}>
              {opts.confirmLabel ?? "Confirm"}
            </Button>
          </div>
        </Modal>
      )}
    </ConfirmContext.Provider>
  );
}

export function useConfirm() {
  const confirm = useContext(ConfirmContext);
  if (!confirm) throw new Error("useConfirm must be used within ConfirmProvider");
  return confirm;
}
