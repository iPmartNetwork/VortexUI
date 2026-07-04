import { useEffect, useState } from "react";
import { useUpdateNode } from "@/api/hooks";
import type { Node } from "@/api/types";
import { findLocationPreset, SERVER_LOCATIONS } from "@/lib/serverLocations";
import { Button, Input, Select } from "./ui";
import { Modal } from "./Modal";
import { useToast } from "./toast";

const AUTO_LOCATION = "__auto__";
const CUSTOM_LOCATION = "__custom__";

export function EditNodeModal({ node, onClose }: { node: Node | null; onClose: () => void }) {
  const update = useUpdateNode();
  const toast = useToast();
  const [name, setName] = useState("");
  const [address, setAddress] = useState("");
  const [ratio, setRatio] = useState("");
  const [endpoint, setEndpoint] = useState("");
  const [locationKey, setLocationKey] = useState(AUTO_LOCATION);
  const [region, setRegion] = useState("");
  const [countryCode, setCountryCode] = useState("");
  const [speedLimit, setSpeedLimit] = useState("");
  const [geoBlock, setGeoBlock] = useState("");
  const [error, setError] = useState("");

  useEffect(() => {
    if (!node) return;
    setName(node.name);
    setAddress(node.address);
    setRatio(String(node.usage_ratio));
    setEndpoint(node.endpoint || "");
    setRegion(node.region || "");
    setCountryCode(node.country_code || "");
    if (node.location_auto !== false) {
      setLocationKey(AUTO_LOCATION);
    } else {
      const preset = findLocationPreset(node.region, node.country_code);
      setLocationKey(preset ? `${preset.code}-${preset.city}` : CUSTOM_LOCATION);
    }
    setSpeedLimit(node.speed_limit ? String(node.speed_limit) : "");
    setGeoBlock(node.geo_block?.join(",") ?? "");
    setError("");
  }, [node]);

  if (!node) return null;

  const isCustomLocation = locationKey === CUSTOM_LOCATION;
  const isAutoLocation = locationKey === AUTO_LOCATION;

  async function submit(e: React.FormEvent) {
    e.preventDefault();
    if (!node) return;
    setError("");
    const preset = SERVER_LOCATIONS.find((p) => `${p.code}-${p.city}` === locationKey);
    try {
      await update.mutateAsync({
        id: node.id,
        input: {
          name, address, usage_ratio: ratio ? Number(ratio) : undefined, endpoint,
          location_auto: isAutoLocation,
          region: isAutoLocation ? "" : preset ? preset.label : region,
          country_code: isAutoLocation ? "" : preset ? preset.code : countryCode,
          speed_limit: speedLimit ? Number(speedLimit) : 0,
          geo_block: geoBlock ? geoBlock.split(",").map(s => s.trim()).filter(Boolean) : [],
        },
      });
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
          Server location
          <Select className="mt-1" value={locationKey} onChange={(e) => setLocationKey(e.target.value)}>
            <option value={AUTO_LOCATION}>Auto-detect from IP (GeoIP)</option>
            {SERVER_LOCATIONS.map((loc) => (
              <option key={`${loc.code}-${loc.city}`} value={`${loc.code}-${loc.city}`}>{loc.label}</option>
            ))}
            <option value={CUSTOM_LOCATION}>Custom…</option>
          </Select>
        </label>
        {isCustomLocation && (
          <div className="grid grid-cols-2 gap-2">
            <label className="block text-xs text-muted-foreground">
              Region label (display)
              <Input className="mt-1" placeholder="e.g. Frankfurt, DE" value={region} onChange={(e) => setRegion(e.target.value)} />
            </label>
            <label className="block text-xs text-muted-foreground">
              Country code (ISO)
              <Input className="mt-1" placeholder="e.g. DE" maxLength={2} value={countryCode} onChange={(e) => setCountryCode(e.target.value.toUpperCase())} />
            </label>
          </div>
        )}
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
