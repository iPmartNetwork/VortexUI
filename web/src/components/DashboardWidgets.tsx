import { useState } from "react";
import { cn } from "@/lib/utils";
import { Button } from "@/components/ui";
import {
  GripVertical, X, Plus, RotateCcw,
  Activity, Users, Wifi, Zap, BarChart3,
  Globe, Gauge as GaugeIcon, TrendingUp,
} from "lucide-react";

// Widget registry
export type WidgetType =
  | "traffic_chart"
  | "user_stats"
  | "node_health"
  | "gauges"
  | "top_users"
  | "geo_map"
  | "recent_activity"
  | "bandwidth_sparkline";

export interface WidgetConfig {
  id: string;
  type: WidgetType;
  size: "sm" | "md" | "lg"; // col-span-1, 2, 3
  visible: boolean;
}

const WIDGET_META: Record<WidgetType, { label: string; icon: React.ElementType }> = {
  traffic_chart: { label: "Traffic Chart", icon: TrendingUp },
  user_stats: { label: "User Statistics", icon: Users },
  node_health: { label: "Node Health", icon: Wifi },
  gauges: { label: "System Gauges", icon: GaugeIcon },
  top_users: { label: "Top Users", icon: Activity },
  geo_map: { label: "Geo Map", icon: Globe },
  recent_activity: { label: "Recent Activity", icon: BarChart3 },
  bandwidth_sparkline: { label: "Bandwidth Live", icon: Zap },
};

const DEFAULT_LAYOUT: WidgetConfig[] = [
  { id: "w1", type: "user_stats", size: "lg", visible: true },
  { id: "w2", type: "node_health", size: "md", visible: true },
  { id: "w3", type: "gauges", size: "md", visible: true },
  { id: "w4", type: "traffic_chart", size: "lg", visible: true },
  { id: "w5", type: "bandwidth_sparkline", size: "md", visible: true },
  { id: "w6", type: "top_users", size: "md", visible: true },
  { id: "w7", type: "geo_map", size: "lg", visible: true },
  { id: "w8", type: "recent_activity", size: "md", visible: true },
];

const STORAGE_KEY = "vortex.dashboard.layout";

function loadLayout(): WidgetConfig[] {
  try {
    const saved = localStorage.getItem(STORAGE_KEY);
    if (saved) return JSON.parse(saved);
  } catch {}
  return DEFAULT_LAYOUT;
}

function saveLayout(layout: WidgetConfig[]) {
  localStorage.setItem(STORAGE_KEY, JSON.stringify(layout));
}

interface DashboardWidgetsProps {
  renderWidget: (type: WidgetType) => React.ReactNode;
}

export function DashboardWidgets({ renderWidget }: DashboardWidgetsProps) {
  const [layout, setLayout] = useState<WidgetConfig[]>(loadLayout);
  const [editing, setEditing] = useState(false);
  const [dragIdx, setDragIdx] = useState<number | null>(null);

  const visibleWidgets = layout.filter(w => w.visible);

  function updateLayout(newLayout: WidgetConfig[]) {
    setLayout(newLayout);
    saveLayout(newLayout);
  }

  function toggleVisibility(id: string) {
    updateLayout(layout.map(w => w.id === id ? { ...w, visible: !w.visible } : w));
  }

  function cycleSize(id: string) {
    const sizes: ("sm" | "md" | "lg")[] = ["sm", "md", "lg"];
    updateLayout(layout.map(w => {
      if (w.id !== id) return w;
      const next = sizes[(sizes.indexOf(w.size) + 1) % sizes.length];
      return { ...w, size: next };
    }));
  }

  function resetLayout() {
    updateLayout(DEFAULT_LAYOUT);
  }

  function handleDragStart(idx: number) {
    setDragIdx(idx);
  }

  function handleDrop(targetIdx: number) {
    if (dragIdx === null || dragIdx === targetIdx) return;
    const items = [...layout];
    const [moved] = items.splice(dragIdx, 1);
    items.splice(targetIdx, 0, moved);
    updateLayout(items);
    setDragIdx(null);
  }

  const sizeClass = (size: string) => {
    switch (size) {
      case "sm": return "col-span-1";
      case "lg": return "col-span-1 md:col-span-2 lg:col-span-3";
      default: return "col-span-1 md:col-span-2";
    }
  };

  return (
    <div className="space-y-4">
      {/* Toolbar */}
      <div className="flex items-center justify-end gap-2">
        <Button
          variant={editing ? "primary" : "outline"}
          size="sm"
          onClick={() => setEditing(!editing)}
        >
          {editing ? "Done" : "Customize"}
        </Button>
        {editing && (
          <Button variant="ghost" size="sm" onClick={resetLayout}>
            <RotateCcw size={13} /> Reset
          </Button>
        )}
      </div>

      {/* Widget grid */}
      <div className="grid grid-cols-1 gap-4 md:grid-cols-2 lg:grid-cols-3">
        {visibleWidgets.map((widget, idx) => {
          return (
            <div
              key={widget.id}
              className={cn(
                "relative transition-all duration-200",
                sizeClass(widget.size),
                editing && "ring-1 ring-dashed ring-border/60 rounded-2xl",
                dragIdx === idx && "opacity-50",
              )}
              draggable={editing}
              onDragStart={() => handleDragStart(idx)}
              onDragOver={(e) => e.preventDefault()}
              onDrop={() => handleDrop(idx)}
            >
              {/* Edit overlay */}
              {editing && (
                <div className="absolute -top-2 -end-2 z-10 flex gap-1">
                  <button
                    onClick={() => cycleSize(widget.id)}
                    className="grid h-6 w-6 place-items-center rounded-full bg-accent text-white text-[10px] font-bold shadow-md hover:scale-110 transition"
                    title="Resize"
                  >
                    {widget.size.toUpperCase()[0]}
                  </button>
                  <button
                    onClick={() => toggleVisibility(widget.id)}
                    className="grid h-6 w-6 place-items-center rounded-full bg-danger text-white shadow-md hover:scale-110 transition"
                    title="Hide"
                  >
                    <X size={12} />
                  </button>
                </div>
              )}
              {editing && (
                <div className="absolute top-2 start-2 z-10 cursor-grab text-fg-subtle/50">
                  <GripVertical size={16} />
                </div>
              )}
              {renderWidget(widget.type)}
            </div>
          );
        })}
      </div>

      {/* Hidden widgets list (when editing) */}
      {editing && layout.some(w => !w.visible) && (
        <div className="rounded-xl border border-dashed border-border/60 p-4 space-y-2">
          <span className="text-xs font-semibold text-fg-subtle">Hidden widgets:</span>
          <div className="flex flex-wrap gap-2">
            {layout.filter(w => !w.visible).map(w => {
              const meta = WIDGET_META[w.type];
              const Icon = meta.icon;
              return (
                <button
                  key={w.id}
                  onClick={() => toggleVisibility(w.id)}
                  className="flex items-center gap-1.5 rounded-lg border border-border/50 bg-surface-2/40 px-2.5 py-1.5 text-xs text-fg-muted hover:bg-primary/10 hover:text-primary transition"
                >
                  <Plus size={12} /> <Icon size={12} /> {meta.label}
                </button>
              );
            })}
          </div>
        </div>
      )}
    </div>
  );
}
