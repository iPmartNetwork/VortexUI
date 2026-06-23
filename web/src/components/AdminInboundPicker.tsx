import { useAllInbounds } from "@/api/hooks";

/** Multi-select checklist of inbounds for reseller allowlist assignment. */
export function AdminInboundPicker({
  selected,
  onChange,
}: {
  selected: string[];
  onChange: (ids: string[]) => void;
}) {
  const inbounds = useAllInbounds();

  function toggle(id: string, on: boolean) {
    onChange(on ? [...selected, id] : selected.filter((x) => x !== id));
  }

  return (
    <div>
      <p className="mb-1 text-xs font-medium text-muted-foreground">Allowed inbounds</p>
      <p className="mb-2 text-[10px] text-fg-subtle">Reseller can only assign users to these inbounds.</p>
      <div className="max-h-40 space-y-1 overflow-y-auto rounded-md border p-2">
        {inbounds.data?.map((ib) => (
          <label key={ib.id} className="flex items-center gap-2 text-sm">
            <input
              type="checkbox"
              checked={selected.includes(ib.id)}
              onChange={(e) => toggle(ib.id, e.target.checked)}
            />
            <span className="font-mono text-xs">{ib.tag}</span>
            <span className="text-muted-foreground">({ib.nodeName} · {ib.protocol})</span>
          </label>
        ))}
        {inbounds.data?.length === 0 && (
          <p className="text-xs text-muted-foreground">No inbounds yet — create them under Nodes first.</p>
        )}
      </div>
    </div>
  );
}
