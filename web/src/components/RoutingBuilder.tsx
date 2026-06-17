import { useState } from "react";
import { GripVertical, Plus, Trash2, ArrowRight } from "lucide-react";
import { Button, Card, Input, Select } from "./ui";

interface RuleItem {
  id: string;
  name: string;
  type: "domain" | "ip" | "port" | "protocol";
  value: string;
  outbound: string;
  enabled: boolean;
}

const OUTBOUNDS = ["direct", "blocked", "proxy", "warp"];
const RULE_TYPES = [
  { value: "domain", label: "Domain" },
  { value: "ip", label: "IP/GeoIP" },
  { value: "port", label: "Port" },
  { value: "protocol", label: "Protocol" },
];

interface Props {
  rules: RuleItem[];
  onChange: (rules: RuleItem[]) => void;
}

export function RoutingBuilder({ rules, onChange }: Props) {
  const [dragIdx, setDragIdx] = useState<number | null>(null);

  function addRule() {
    onChange([...rules, {
      id: crypto.randomUUID(),
      name: "",
      type: "domain",
      value: "",
      outbound: "direct",
      enabled: true,
    }]);
  }

  function removeRule(id: string) {
    onChange(rules.filter(r => r.id !== id));
  }

  function updateRule(id: string, field: keyof RuleItem, value: any) {
    onChange(rules.map(r => r.id === id ? { ...r, [field]: value } : r));
  }

  function handleDragStart(idx: number) {
    setDragIdx(idx);
  }

  function handleDragOver(e: React.DragEvent, idx: number) {
    e.preventDefault();
    if (dragIdx === null || dragIdx === idx) return;
    const newRules = [...rules];
    const [dragged] = newRules.splice(dragIdx, 1);
    newRules.splice(idx, 0, dragged);
    onChange(newRules);
    setDragIdx(idx);
  }

  function handleDragEnd() {
    setDragIdx(null);
  }

  return (
    <div className="space-y-3">
      <div className="flex items-center justify-between">
        <h3 className="text-sm font-semibold text-fg">Routing Rules</h3>
        <Button variant="ghost" size="sm" onClick={addRule}>
          <Plus size={14} /> Add Rule
        </Button>
      </div>

      <p className="text-[10px] text-fg-subtle">Drag to reorder. Rules are evaluated top-to-bottom; first match wins.</p>

      <div className="space-y-2">
        {rules.map((rule, idx) => (
          <Card
            key={rule.id}
            draggable
            onDragStart={() => handleDragStart(idx)}
            onDragOver={(e) => handleDragOver(e, idx)}
            onDragEnd={handleDragEnd}
            className={`flex items-center gap-2 p-3 cursor-move transition ${dragIdx === idx ? "opacity-50 ring-2 ring-primary" : ""} ${!rule.enabled ? "opacity-40" : ""}`}
          >
            <GripVertical size={14} className="shrink-0 text-fg-subtle" />

            <span className="text-[10px] font-bold text-fg-subtle w-5">{idx + 1}</span>

            <Select value={rule.type} onChange={(e) => updateRule(rule.id, "type", e.target.value)} className="w-24 text-xs">
              {RULE_TYPES.map(t => <option key={t.value} value={t.value}>{t.label}</option>)}
            </Select>

            <Input
              placeholder={rule.type === "domain" ? "geosite:ads, example.com" : rule.type === "ip" ? "geoip:ir, 10.0.0.0/8" : "443, 80"}
              value={rule.value}
              onChange={(e) => updateRule(rule.id, "value", e.target.value)}
              className="flex-1 text-xs"
            />

            <ArrowRight size={12} className="text-fg-subtle shrink-0" />

            <Select value={rule.outbound} onChange={(e) => updateRule(rule.id, "outbound", e.target.value)} className="w-24 text-xs">
              {OUTBOUNDS.map(o => <option key={o} value={o}>{o}</option>)}
            </Select>

            <button
              onClick={() => updateRule(rule.id, "enabled", !rule.enabled)}
              className={`h-4 w-4 rounded-full border ${rule.enabled ? "bg-success border-success" : "bg-transparent border-fg-subtle"}`}
              title={rule.enabled ? "Enabled" : "Disabled"}
            />

            <button onClick={() => removeRule(rule.id)} className="text-fg-subtle hover:text-danger transition">
              <Trash2 size={13} />
            </button>
          </Card>
        ))}
      </div>

      {rules.length === 0 && (
        <p className="text-center text-sm text-fg-muted py-6">No rules yet. Click "Add Rule" or apply a template.</p>
      )}
    </div>
  );
}
