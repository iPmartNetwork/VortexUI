import { useState } from "react";
import { useQuery, useMutation } from "@tanstack/react-query";
import { RefreshCw } from "lucide-react";
import { api } from "@/api/client";
import { Button, Input } from "@/components/ui";
import { GlassCard } from "@/components/veltrix";
import { useToast } from "@/components/toast";
import { useTitle } from "@/lib/useTitle";

export function IPRotation() {
  useTitle("IP Rotation");
  const toast = useToast();
  const [provider, setProvider] = useState("digitalocean");
  const [token, setToken] = useState("");
  const [currentIP, setCurrentIP] = useState("");
  const [newIP, setNewIP] = useState("");

  const { data: detect } = useQuery({ queryKey: ["ip-detect"], queryFn: () => api<{ ip: string; provider: string; org: string }>("/api/v2/ip-rotation/detect") });
  const { data: providers } = useQuery({ queryKey: ["ip-providers"], queryFn: () => api<{ providers: { id: string; name: string }[] }>("/api/v2/ip-rotation/providers") });

  const rotateMut = useMutation({
    mutationFn: () => api<{ new_ip: string }>("/api/v2/ip-rotation/rotate", { method: "POST", body: { provider, token, current_ip: currentIP || detect?.ip } }),
    onSuccess: (res) => { setNewIP(res.new_ip); toast.success("New IP allocated: " + res.new_ip); },
    onError: () => toast.error("Rotation failed"),
  });

  return (
    <div className="space-y-6 p-6">
      <h1 className="text-2xl font-bold flex items-center gap-2"><RefreshCw className="w-6 h-6" />IP Rotation</h1>
      <p className="text-sm text-fg-muted">Rotate your server IP via cloud provider API when current IP gets blocked.</p>

      {detect && (
        <GlassCard className="p-4">
          <h3 className="font-semibold mb-2">Current Server</h3>
          <div className="grid grid-cols-3 gap-4 text-sm">
            <div><span className="text-fg-muted">IP:</span> <strong>{detect.ip}</strong></div>
            <div><span className="text-fg-muted">Provider:</span> <strong>{detect.provider}</strong></div>
            <div><span className="text-fg-muted">Org:</span> <strong>{detect.org}</strong></div>
          </div>
        </GlassCard>
      )}

      <GlassCard className="p-4 space-y-4">
        <h3 className="font-semibold">Rotate IP</h3>
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          <div>
            <label className="text-xs font-medium">Provider</label>
            <select className="field input-surface w-full mt-1" value={provider} onChange={(e) => setProvider(e.target.value)}>
              {(providers?.providers ?? []).map((p) => <option key={p.id} value={p.id}>{p.name}</option>)}
            </select>
          </div>
          <div>
            <label className="text-xs font-medium">API Token</label>
            <Input type="password" className="w-full mt-1" value={token} onChange={(e) => setToken(e.target.value)} placeholder="Provider API token" />
          </div>
        </div>
        <div>
          <label className="text-xs font-medium">Current IP (auto-detected)</label>
          <Input className="w-full mt-1" value={currentIP || detect?.ip || ""} onChange={(e) => setCurrentIP(e.target.value)} />
        </div>
        <Button onClick={() => rotateMut.mutate()} disabled={!token || rotateMut.isPending}>
          <RefreshCw size={14} className={rotateMut.isPending ? "animate-spin" : ""} /> Rotate IP
        </Button>
        {newIP && <p className="text-sm text-green-500 font-medium">New IP: {newIP}</p>}
      </GlassCard>
    </div>
  );
}
