import { useState, useEffect } from "react";
import { Bell } from "lucide-react";
import { cn } from "@/lib/utils";

export interface Notification {
  id: string;
  title: string;
  message: string;
  type: "info" | "warning" | "error" | "success";
  time: string;
  read: boolean;
}

export function NotificationCenter() {
  const [open, setOpen] = useState(false);
  const [notifications, setNotifications] = useState<Notification[]>([]);
  const unread = notifications.filter(n => !n.read).length;

  // Listen for SSE notifications
  useEffect(() => {
    function handler(e: Event) {
      const detail = (e as CustomEvent).detail;
      if (detail) {
        setNotifications(prev => [
          { id: crypto.randomUUID(), ...detail, time: new Date().toISOString(), read: false },
          ...prev.slice(0, 49),
        ]);
      }
    }
    window.addEventListener("vortex:notification", handler);
    return () => window.removeEventListener("vortex:notification", handler);
  }, []);

  function markAllRead() {
    setNotifications(prev => prev.map(n => ({ ...n, read: true })));
  }

  return (
    <div className="relative">
      <button
        onClick={() => setOpen(!open)}
        className="relative grid h-9 w-9 place-items-center rounded-xl text-fg-muted transition hover:bg-surface-2/60 hover:text-fg"
        aria-label="Notifications"
      >
        <Bell size={18} />
        {unread > 0 && (
          <span className="absolute -end-0.5 -top-0.5 grid h-4 min-w-4 place-items-center rounded-full bg-danger px-1 text-[9px] font-bold text-white">
            {unread > 9 ? "9+" : unread}
          </span>
        )}
      </button>

      {open && (
        <>
          <div className="fixed inset-0 z-40" onClick={() => setOpen(false)} />
          <div className="absolute end-0 top-full z-50 mt-2 w-80 rounded-xl border border-border/60 bg-bg-elevated/95 shadow-2xl backdrop-blur-xl animate-scale-in overflow-hidden">
            <div className="flex items-center justify-between border-b border-border/40 px-4 py-2.5">
              <span className="text-xs font-semibold text-fg">Notifications</span>
              {unread > 0 && (
                <button onClick={markAllRead} className="text-[10px] text-primary hover:underline">Mark all read</button>
              )}
            </div>
            <div className="max-h-[360px] overflow-y-auto">
              {notifications.length === 0 && (
                <div className="py-8 text-center text-xs text-fg-muted">No notifications</div>
              )}
              {notifications.map(n => (
                <div key={n.id} className={cn(
                  "flex gap-3 border-b border-border/20 px-4 py-3 transition",
                  !n.read && "bg-primary/[0.03]",
                )}>
                  <div className={cn(
                    "mt-0.5 h-2 w-2 shrink-0 rounded-full",
                    n.type === "error" ? "bg-danger" : n.type === "warning" ? "bg-warning" : n.type === "success" ? "bg-success" : "bg-accent",
                  )} />
                  <div className="min-w-0 flex-1">
                    <div className="text-xs font-medium text-fg">{n.title}</div>
                    <div className="text-[11px] text-fg-muted truncate">{n.message}</div>
                    <div className="mt-0.5 text-[10px] text-fg-subtle">{new Date(n.time).toLocaleTimeString()}</div>
                  </div>
                </div>
              ))}
            </div>
          </div>
        </>
      )}
    </div>
  );
}
