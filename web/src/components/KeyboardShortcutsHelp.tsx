import { useState, useEffect } from "react";
import { X } from "lucide-react";

const SHORTCUTS = [
  { keys: ["Ctrl", "K"], desc: "Open command palette" },
  { keys: ["N"], desc: "Go to Users (new user)" },
  { keys: ["S"], desc: "Quick search" },
  { keys: ["?"], desc: "Show this help" },
  { keys: ["Esc"], desc: "Close modals / palettes" },
];

export function KeyboardShortcutsHelp() {
  const [open, setOpen] = useState(false);

  useEffect(() => {
    function handler() { setOpen(true); }
    window.addEventListener("vortex:show-shortcuts", handler);
    return () => window.removeEventListener("vortex:show-shortcuts", handler);
  }, []);

  if (!open) return null;

  return (
    <div className="fixed inset-0 z-[200] flex items-center justify-center bg-black/50 backdrop-blur-sm animate-fade-in" onClick={() => setOpen(false)}>
      <div className="card w-full max-w-sm p-6 animate-scale-in" onClick={e => e.stopPropagation()}>
        <div className="flex items-center justify-between mb-5">
          <h2 className="text-lg font-semibold text-fg">Keyboard Shortcuts</h2>
          <button onClick={() => setOpen(false)} className="grid h-8 w-8 place-items-center rounded-lg text-fg-subtle hover:bg-surface-2 hover:text-fg">
            <X size={16} />
          </button>
        </div>
        <div className="space-y-3">
          {SHORTCUTS.map(s => (
            <div key={s.desc} className="flex items-center justify-between">
              <span className="text-sm text-fg-muted">{s.desc}</span>
              <div className="flex gap-1">
                {s.keys.map(k => (
                  <kbd key={k} className="inline-flex h-6 min-w-6 items-center justify-center rounded-md border border-border/60 bg-surface-2/50 px-1.5 text-[11px] font-semibold text-fg-subtle">
                    {k}
                  </kbd>
                ))}
              </div>
            </div>
          ))}
        </div>
      </div>
    </div>
  );
}
