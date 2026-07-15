import type { CoreType } from "@/lib/coreTypes";
import { cn } from "@/lib/utils";

const OPTIONS: { id: CoreType; label: string; hint: string }[] = [
  { id: "xray", label: "Xray-core", hint: "VMess, VLESS, Trojan, REALITY, …" },
  { id: "singbox", label: "sing-box", hint: "Hysteria2, TUIC, WireGuard, Naive, …" },
];

export function EnabledCoresPicker({
  value,
  onChange,
  defaultCore,
  onDefaultCoreChange,
  disabled,
}: {
  value: CoreType[];
  onChange: (cores: CoreType[]) => void;
  defaultCore: CoreType;
  onDefaultCoreChange: (core: CoreType) => void;
  disabled?: boolean;
}) {
  const toggle = (id: CoreType) => {
    const set = new Set(value);
    if (set.has(id)) {
      if (set.size <= 1) return;
      set.delete(id);
    } else {
      set.add(id);
    }
    const next = OPTIONS.map((o) => o.id).filter((c) => set.has(c));
    onChange(next);
    if (!next.includes(defaultCore)) onDefaultCoreChange(next[0] ?? "xray");
  };

  const multi = value.length > 1;

  return (
    <div className="space-y-2">
      <p className="text-xs text-muted-foreground">Enabled engines on this node</p>
      <div className="grid gap-2 sm:grid-cols-2">
        {OPTIONS.map((opt) => {
          const checked = value.includes(opt.id);
          return (
            <button
              key={opt.id}
              type="button"
              disabled={disabled}
              onClick={() => toggle(opt.id)}
              className={cn(
                "rounded-lg border px-3 py-2 text-left transition",
                checked
                  ? "border-primary/50 bg-primary/10"
                  : "border-border/60 bg-surface-2/30 hover:border-border",
                disabled && "opacity-60",
              )}
            >
              <div className="flex items-center gap-2">
                <span
                  className={cn(
                    "flex h-4 w-4 shrink-0 items-center justify-center rounded border text-[10px]",
                    checked ? "border-primary bg-primary text-primary-foreground" : "border-border",
                  )}
                >
                  {checked ? "✓" : ""}
                </span>
                <span className="text-sm font-medium text-fg">{opt.label}</span>
              </div>
              <p className="mt-1 ps-6 text-[10px] text-fg-subtle">{opt.hint}</p>
            </button>
          );
        })}
      </div>
      {multi && (
        <label className="block text-xs text-muted-foreground">
          Default core for new inbounds
          <select
            className="mt-1 w-full rounded-md border border-border bg-surface-1 px-2 py-1.5 text-sm"
            value={defaultCore}
            disabled={disabled}
            onChange={(e) => onDefaultCoreChange(e.target.value as CoreType)}
          >
            {value.map((c) => (
              <option key={c} value={c}>
                {c === "singbox" ? "sing-box" : "Xray-core"}
              </option>
            ))}
          </select>
          <span className="mt-1 block text-[10px] text-fg-subtle">
            Dual-core nodes need <span className="font-mono">VORTEX_ENABLED_CORES=xray,singbox</span> on the agent.
            Each inbound can override its engine in the inbounds form.
          </span>
        </label>
      )}
    </div>
  );
}
