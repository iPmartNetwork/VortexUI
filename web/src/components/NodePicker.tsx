import { useEffect } from "react";
import { useNodes } from "@/api/hooks";
import { Select } from "./ui";

// NodePicker drives the per-node policy pages. It auto-selects the first node so
// the page shows data immediately.
export function NodePicker({ value, onChange }: { value: string; onChange: (id: string) => void }) {
  const { data } = useNodes();
  const nodes = data?.nodes ?? [];

  useEffect(() => {
    if (!value && nodes.length) onChange(nodes[0].id);
  }, [value, nodes, onChange]);

  return (
    <Select className="w-52" value={value} onChange={(e) => onChange(e.target.value)}>
      {nodes.length === 0 && <option value="">No nodes</option>}
      {nodes.map((n) => (
        <option key={n.id} value={n.id}>
          {n.name} · {n.core}
        </option>
      ))}
    </Select>
  );
}
