import { useEffect } from "react";
import { useQueryClient } from "@tanstack/react-query";
import { getToken } from "./client";
import { useToast } from "@/components/toast";

interface LiveEvent {
  type: string;
  username?: string;
  node_name?: string;
  data?: Record<string, unknown>;
}

const EVENT_TYPES = [
  "user.created", "user.deleted", "user.limited", "user.expired",
  "user.reset", "user.ip_limit", "node.down", "node.up",
];

// useLiveEvents opens an SSE connection to the panel's event stream and reacts
// to domain events as they happen: it invalidates the affected React Query
// caches (so views refresh instantly instead of waiting for the poll) and raises
// a toast for the noteworthy ones. Mounted once, inside the authenticated shell.
export function useLiveEvents() {
  const qc = useQueryClient();
  const toast = useToast();

  useEffect(() => {
    const token = getToken();
    if (!token) return;

    const es = new EventSource(`/api/events/stream?access_token=${encodeURIComponent(token)}`);

    const handle = (ev: MessageEvent) => {
      let e: LiveEvent;
      try { e = JSON.parse(ev.data); } catch { return; }

      switch (e.type) {
        case "node.down":
          qc.invalidateQueries({ queryKey: ["nodes"] });
          qc.invalidateQueries({ queryKey: ["overview"] });
          toast.error(`Node down: ${e.node_name ?? "?"}`);
          break;
        case "node.up":
          qc.invalidateQueries({ queryKey: ["nodes"] });
          qc.invalidateQueries({ queryKey: ["overview"] });
          toast.success(`Node up: ${e.node_name ?? "?"}`);
          break;
        case "user.ip_limit":
          qc.invalidateQueries({ queryKey: ["users"] });
          toast.info(`Account sharing: ${e.username ?? "?"} (${e.data?.online_ips ?? "?"} IPs)`);
          break;
        case "user.limited":
          qc.invalidateQueries({ queryKey: ["users"] });
          qc.invalidateQueries({ queryKey: ["overview"] });
          toast.info(`User reached data limit: ${e.username ?? "?"}`);
          break;
        case "user.expired":
          qc.invalidateQueries({ queryKey: ["users"] });
          qc.invalidateQueries({ queryKey: ["overview"] });
          toast.info(`User expired: ${e.username ?? "?"}`);
          break;
        default:
          // created / deleted / reset: refresh lists silently.
          qc.invalidateQueries({ queryKey: ["users"] });
          qc.invalidateQueries({ queryKey: ["overview"] });
      }
    };

    EVENT_TYPES.forEach((t) => es.addEventListener(t, handle as EventListener));

    return () => {
      EVENT_TYPES.forEach((t) => es.removeEventListener(t, handle as EventListener));
      es.close();
    };
  }, [qc, toast]);
}
