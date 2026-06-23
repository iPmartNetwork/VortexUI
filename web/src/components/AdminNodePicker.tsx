import { useNodes } from "@/api/hooks";

export function AdminNodePicker({
  selected,
  onChange,
}: {
  selected: string[];
  onChange: (ids: string[]) => void;
}) {
  const nodes = useNodes();

  function toggle(id: string, on: boolean) {
    onChange(on ? [...selected, id] : selected.filter((x) => x !== id));
  }

  return (
    <div>
      <p className="mb-1 text-xs font-medium text-muted-foreground">Allowed nodes</p>
      <p className="mb-2 text-[10px] text-fg-subtle">Reseller only sees and uses these nodes.</p>
      <div className="max-h-32 space-y-1 overflow-y-auto rounded-md border p-2">
        {nodes.data?.nodes?.map((n) => (
          <label key={n.id} className="flex items-center gap-2 text-sm">
            <input type="checkbox" checked={selected.includes(n.id)} onChange={(e) => toggle(n.id, e.target.checked)} />
            <span className="font-medium">{n.name}</span>
            <span className="text-xs text-muted-foreground">({n.core})</span>
          </label>
        ))}
        {(nodes.data?.nodes?.length ?? 0) === 0 && (
          <p className="text-xs text-muted-foreground">No nodes yet.</p>
        )}
      </div>
    </div>
  );
}
