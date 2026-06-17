import { useEffect, useState } from "react";
import { useUpdateNode } from "@/api/hooks";
import type { Node } from "@/api/types";
import { Button, Input } from "./ui";
import { Modal } from "./Modal";
import { useToast } from "./toast";

export function EditNodeModal({ node, onClose }: { node: Node | null; onClose: () => void }) {
  const update = useUpdateNode();
  const toast = useToast();
  const [name, setName] = useState("");
  const [address, setAddress] = useState("");
  const [ratio, setRatio] = useState("");
  const [endpoint, setEndpoint] = useState("");
  const [speedLimit, setSpeedLimit] = useState("");
  const [geoBlock, setGeoBlock] = useState("");
  const [error, setError] = useState("");

  useEffect(() => {
    if (!node) return;
    setName(node.name);
    setAddress(node.address);
    setRatio(String(node.usage_ratio));
    setEndpoint(node.endpoint || "");
    setSpeedLimit(node.speed_limit ? String(node.speed_limit) : "");
    setGeoBlock(node.geo_block?.join(",") ?? "");
    setError("");
  }, [node]);

  if (!node) return null;

  async function submit(e: React.FormEvent) {
    e.preventDefault();
    if (!node) return;
    setError("");
    try {
      await update.mutateAsync({ id: node.id, input: { name, address, usage_ratio: ratio ? Number(ratio) : undefined, endpoint, speed_limit: speedLimit ? Number(speedLimit) : 0, geo_block: geoBlock ? geoBlock.split(",").map(s => s.trim()).filter(Boolean) : [] } });
      toast.success(`Saved ${name}`);
      onClose();
    } catch {
      setError("Update failed");
    }
  }

  return (
    <Modal open={!!node} onClose={onClose} title={`Edit · ${node.name}`}>
      <form onSubmit={submit} className="space-y-3">
        <label className="block text-xs text-muted-foreground">
          Name
          <Input className="mt-1" value={name} onChange={(e) => setName(e.target.value)} required />
        </label>
        <label className="block text-xs text-muted-foreground">
          Agent address
          <Input className="mt-1" value={address} onChange={(e) => setAddress(e.target.value)} required />
        </label>
        <label className="block text-xs text-muted-foreground">
          Usage ratio
          <Input className="mt-1" value={ratio} onChange={(e) => setRatio(e.target.value)} inputMode="decimal" />
        </label>
        <label className="block text-xs text-muted-foreground">
          Endpoint (tunnel/CDN address)
          <Input className="mt-1" placeholder="Leave empty to use real IP" value={endpoint} onChange={(e) => setEndpoint(e.target.value)} />
        </label>
        <p className="text-[10px] text-fg-subtle">Subscription links will use this address instead of the real server IP. Useful for tunneled or relay setups.</p>
        <label className="block text-xs text-muted-foreground">
          Speed limit (bytes/sec)
          <Input className="mt-1" placeholder="0 = unlimited" value={speedLimit} onChange={(e) => setSpeedLimit(e.target.value)} inputMode="numeric" />
          <span className="text-[10px] text-fg-subtle">Per-user download speed cap on this node. Enter 0 or leave empty for unlimited. Example: 1048576 = 1 MB/s</span>
        </label>
        <label className="block text-xs text-muted-foreground">
          Geo-blocking (allowed countries)
          <Input className="mt-1" placeholder="e.g. IR,TR,AE (empty = all allowed)" value={geoBlock} onChange={(e) => setGeoBlock(e.target.value)} />
          <span className="text-[10px] text-fg-subtle">Comma-separated ISO country codes. Only users from these countries can connect. Leave empty to allow all countries.</span>
        </label>
        {error && <p className="text-sm text-destructive">{error}</p>}
        <div className="flex justify-end gap-2 pt-1">
          <Button type="button" variant="ghost" onClick={onClose}>Cancel</Button>
          <Button type="submit" disabled={update.isPending}>{update.isPending ? "Saving…" : "Save"}</Button>
        </div>
      </form>
    </Modal>
  );
}
