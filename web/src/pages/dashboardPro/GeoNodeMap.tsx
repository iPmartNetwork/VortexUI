import { useState, useRef, useEffect, useCallback } from "react";
import { MapPin } from "lucide-react";
import { cn } from "@/lib/utils";

import L from "leaflet";
import { MapContainer, TileLayer, Marker, Popup } from "react-leaflet";
import "leaflet/dist/leaflet.css";

// Fix Leaflet default marker icon (broken in bundlers)
import iconUrl from "leaflet/dist/images/marker-icon.png";
import iconRetinaUrl from "leaflet/dist/images/marker-icon-2x.png";
import shadowUrl from "leaflet/dist/images/marker-shadow.png";

const DefaultIcon = L.icon({
  iconUrl,
  iconRetinaUrl,
  shadowUrl,
  iconSize: [25, 41],
  iconAnchor: [12, 41],
  popupAnchor: [1, -34],
  shadowSize: [41, 41],
});
L.Marker.prototype.options.icon = DefaultIcon;

export interface GeoNode {
  node_id: string;
  name: string;
  lat: number;
  lng: number;
  status: string;
}

export function GeoNodeMap({ nodes }: { nodes: GeoNode[] }) {
  const mapRef = useRef<L.Map | null>(null);
  const heatLayerRef = useRef<L.Layer | null>(null);
  const [heatmapEnabled, setHeatmapEnabled] = useState(false);

  // Create custom marker icons
  const onlineIcon = L.divIcon({
    className: "",
    html: `<div style="width:18px;height:18px;display:flex;align-items:center;justify-content:center">
      <div style="width:14px;height:14px;border-radius:50%;background:hsl(var(--primary));box-shadow:0 0 8px hsl(var(--primary) / 0.6);border:2px solid hsl(var(--bg-elevated));"></div>
      <div style="position:absolute;width:14px;height:14px;border-radius:50%;background:hsl(var(--primary));opacity:0.3;animation:ping 2s cubic-bezier(0,0,0.2,1) infinite;"></div>
    </div>`,
    iconSize: [18, 18],
    iconAnchor: [9, 9],
  });

  const offlineIcon = L.divIcon({
    className: "",
    html: `<div style="width:14px;height:14px;border-radius:50%;background:rgb(239 68 68);border:2px solid hsl(var(--bg-elevated));box-shadow:0 0 4px rgb(239 68 68 / 0.4);"></div>`,
    iconSize: [14, 14],
    iconAnchor: [7, 7],
  });

  // Custom heatmap layer using canvas circles
  function applyHeatmap(map: L.Map, nodeList: GeoNode[]) {
    if (heatLayerRef.current) {
      map.removeLayer(heatLayerRef.current);
      heatLayerRef.current = null;
    }
    if (!heatmapEnabled) return;

    const heatLayer = L.layerGroup().addTo(map);

    nodeList.forEach((node) => {
      const intensity = node.status === "online" ? 0.6 : 0.2;
      L.circleMarker([node.lat, node.lng], {
        radius: 20,
        color: "hsl(var(--primary))",
        fillColor: "hsl(var(--primary))",
        fillOpacity: intensity * 0.3,
        weight: 0,
      }).addTo(heatLayer);
    });

    // Add interpolated heat circles between nodes for visual density
    for (let i = 0; i < nodeList.length; i++) {
      for (let j = i + 1; j < nodeList.length; j++) {
        if (j - i > 3) break;
        const a = nodeList[i];
        const b = nodeList[j];
        if (a.status !== "online" || b.status !== "online") continue;
        const midLat = (a.lat + b.lat) / 2;
        const midLng = (a.lng + b.lng) / 2;
        L.circleMarker([midLat, midLng], {
          radius: 12,
          color: "hsl(var(--accent))",
          fillColor: "hsl(var(--accent))",
          fillOpacity: 0.08,
          weight: 0,
        }).addTo(heatLayer);
      }
    }

    heatLayerRef.current = heatLayer;
  }

  const applyHeatmapCb = useCallback((map: L.Map, nodeList: GeoNode[]) => {
    applyHeatmap(map, nodeList);
  }, [heatmapEnabled]);

  useEffect(() => {
    if (mapRef.current) {
      applyHeatmapCb(mapRef.current, nodes);
    }
  }, [applyHeatmapCb, nodes]);

  if (nodes.length === 0) {
    return (
      <div className="h-56 flex items-center justify-center rounded-xl bg-surface-2/20 border border-border/40 text-xs text-fg-subtle">
        <MapPin size={20} className="me-2 opacity-30" />
        No node location data available
      </div>
    );
  }

  const lats = nodes.map((n) => n.lat);
  const lngs = nodes.map((n) => n.lng);
  const center: [number, number] = [
    (Math.min(...lats) + Math.max(...lats)) / 2,
    (Math.min(...lngs) + Math.max(...lngs)) / 2,
  ];

  const onlineCount = nodes.filter((n) => n.status === "online").length;

  return (
    <div
      className="relative rounded-xl overflow-hidden border border-border/50"
      style={{ height: 300 }}
    >
      <MapContainer
        center={center}
        zoom={4}
        className="w-full h-full z-0"
        scrollWheelZoom={true}
        zoomControl={false}
        ref={mapRef}
      >
        <TileLayer
          attribution='&copy; <a href="https://www.openstreetmap.org/copyright">OpenStreetMap</a>'
          url="https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png"
        />
        {nodes.map((node) => (
          <Marker
            key={node.node_id}
            position={[node.lat, node.lng]}
            icon={node.status === "online" ? onlineIcon : offlineIcon}
          >
            <Popup>
              <div className="text-xs leading-relaxed min-w-[120px]">
                <div className="flex items-center gap-2 mb-1.5">
                  <span
                    style={{
                      display: "inline-block",
                      width: 8,
                      height: 8,
                      borderRadius: "50%",
                      background:
                        node.status === "online"
                          ? "hsl(var(--primary))"
                          : "rgb(239 68 68)",
                      boxShadow:
                        node.status === "online"
                          ? "0 0 4px hsl(var(--primary) / 0.6)"
                          : "none",
                    }}
                  />
                  <span style={{ fontWeight: 700 }}>{node.name}</span>
                </div>
                <div style={{ color: "hsl(var(--fg-muted))" }}>
                  {node.status === "online" ? "🟢 Online" : "🔴 Offline"}
                </div>
                <div
                  style={{
                    fontSize: 9,
                    color: "hsl(var(--fg-subtle))",
                    marginTop: 2,
                  }}
                >
                  {node.lat.toFixed(4)}, {node.lng.toFixed(4)}
                </div>
              </div>
            </Popup>
          </Marker>
        ))}
      </MapContainer>

      {/* Legend */}
      <div className="absolute bottom-2 start-2 z-[1000] flex items-center gap-3 text-[10px] text-fg bg-bg/80 backdrop-blur-sm rounded-lg px-2.5 py-1.5 border border-border/50 shadow-sm">
        <div className="flex items-center gap-1">
          <span className="h-2.5 w-2.5 rounded-full bg-primary shadow-[0_0_4px] shadow-primary/50" />
          Online ({onlineCount})
        </div>
        <div className="flex items-center gap-1">
          <span className="h-2.5 w-2.5 rounded-full bg-destructive" />
          Offline ({nodes.length - onlineCount})
        </div>
        <span className="text-fg-subtle">{nodes.length} total</span>
      </div>

      {/* Top controls: heatmap toggle + zoom */}
      <div className="absolute top-2 end-2 z-[1000] flex flex-col gap-1">
        <button
          onClick={() => setHeatmapEnabled((h) => !h)}
          className={cn(
            "h-7 px-2 flex items-center gap-1 rounded-md backdrop-blur-sm border transition text-[9px] font-bold shadow-sm",
            heatmapEnabled
              ? "bg-primary/20 border-primary/40 text-primary"
              : "bg-bg/80 border-border/50 text-fg-muted hover:text-fg hover:bg-surface-2/80",
          )}
          title={
            heatmapEnabled
              ? "Hide heatmap"
              : "Show traffic heatmap"
          }
        >
          <span
            className="inline-block h-2 w-2 rounded-full"
            style={{
              background: heatmapEnabled
                ? "hsl(var(--primary))"
                : "hsl(var(--fg-subtle))",
              boxShadow: heatmapEnabled
                ? "0 0 4px hsl(var(--primary) / 0.6)"
                : "none",
            }}
          />
          HEAT
        </button>
        <button
          onClick={() => mapRef.current?.zoomIn()}
          className="h-7 w-7 flex items-center justify-center rounded-md bg-bg/80 backdrop-blur-sm border border-border/50 text-fg hover:bg-surface-2/80 transition text-xs font-bold shadow-sm"
        >
          +
        </button>
        <button
          onClick={() => mapRef.current?.zoomOut()}
          className="h-7 w-7 flex items-center justify-center rounded-md bg-bg/80 backdrop-blur-sm border border-border/50 text-fg hover:bg-surface-2/80 transition text-xs font-bold shadow-sm"
        >
          −
        </button>
      </div>
    </div>
  );
}
