import { useState, useEffect } from "react";
import { api } from "@/api/client";

export function PortConflictIndicator({ nodeId, port }: { nodeId: string; port: string }) {
  const [status, setStatus] = useState<"idle" | "checking" | "available" | "conflict">("idle");
  const [conflictTag, setConflictTag] = useState("");

  useEffect(() => {
    const portNum = Number(port);
    if (!nodeId || !portNum || portNum <= 0) {
      setStatus("idle");
      return;
    }
    setStatus("checking");
    const timer = setTimeout(async () => {
      try {
        const res = await api<{ available: boolean; conflict_tag?: string }>(
          "/api/inbounds/check-port",
          { query: { node_id: nodeId, port: portNum } }
        );
        if (res.available) {
          setStatus("available");
          setConflictTag("");
        } else {
          setStatus("conflict");
          setConflictTag(res.conflict_tag || "");
        }
      } catch {
        setStatus("idle");
      }
    }, 300); // debounce 300ms
    return () => clearTimeout(timer);
  }, [nodeId, port]);

  if (status === "idle") return null;
  if (status === "checking") {
    return <span className="text-[10px] text-fg-subtle animate-pulse">Checking...</span>;
  }
  if (status === "available") {
    return <span className="text-[10px] text-success font-medium">&check; Available</span>;
  }
  return (
    <span className="text-[10px] text-danger font-medium">
      &warning; Port used by &ldquo;{conflictTag}&rdquo;
    </span>
  );
}
