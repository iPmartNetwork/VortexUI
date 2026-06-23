import { useQuery } from "@tanstack/react-query";
import { api } from "@/api/client";

interface Plan {
  id: string;
  name: string;
  enabled: boolean;
}

export function AdminPlanPicker({
  selected,
  onChange,
}: {
  selected: string[];
  onChange: (ids: string[]) => void;
}) {
  const plans = useQuery({
    queryKey: ["plans-all"],
    queryFn: () => api<{ plans: Plan[] }>("/api/plans"),
  });

  function toggle(id: string, on: boolean) {
    onChange(on ? [...selected, id] : selected.filter((x) => x !== id));
  }

  return (
    <div>
      <p className="mb-1 text-xs font-medium text-muted-foreground">Allowed plans</p>
      <p className="mb-2 text-[10px] text-fg-subtle">Reseller can sell only these subscription plans.</p>
      <div className="max-h-32 space-y-1 overflow-y-auto rounded-md border p-2">
        {plans.data?.plans.map((p) => (
          <label key={p.id} className="flex items-center gap-2 text-sm">
            <input type="checkbox" checked={selected.includes(p.id)} onChange={(e) => toggle(p.id, e.target.checked)} />
            <span>{p.name}</span>
            {!p.enabled && <span className="text-xs text-muted-foreground">(disabled)</span>}
          </label>
        ))}
        {plans.data?.plans.length === 0 && (
          <p className="text-xs text-muted-foreground">No plans yet — create them under Plans.</p>
        )}
      </div>
    </div>
  );
}
