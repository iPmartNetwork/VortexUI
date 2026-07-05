import { useState } from "react";
import { Terminal } from "lucide-react";
import { useLogs } from "@/api/policy-hooks";
import { Select } from "@/components/ui";
import { GlassCard } from "@/components/veltrix";
import { useI18n } from "@/i18n/i18n";
import { useTitle } from "@/lib/useTitle";

const LEVELS = ["debug", "info", "warn", "error"];

// slog numeric levels: debug=-4, info=0, warn=4, error=8.
function levelLabel(n: number): { text: string; cls: string } {
  if (n >= 8) return { text: "ERROR", cls: "text-danger" };
  if (n >= 4) return { text: "WARN", cls: "text-warning" };
  if (n >= 0) return { text: "INFO", cls: "text-accent" };
  return { text: "DEBUG", cls: "text-fg-subtle" };
}

export function Logs() {
  useTitle("Logs");
  const { t } = useI18n();
  const [level, setLevel] = useState("info");
  const { data, isLoading } = useLogs(level);

  return (
    <div className="space-y-5 animate-page-enter">
      <div className="flex flex-col lg:flex-row lg:items-start justify-between gap-4">
        <div>
          <h1 className="text-2xl font-bold text-fg tracking-tight">{t("nav.logs")}</h1>
          <p className="text-sm text-fg-muted mt-1">Panel activity</p>
        </div>
        <Select className="w-32 flex-shrink-0" value={level} onChange={(e) => setLevel(e.target.value)}>
          {LEVELS.map((l) => <option key={l} value={l}>{l}</option>)}
        </Select>
      </div>

      <GlassCard hover={false} className="!p-0 overflow-hidden">
        <div className="flex items-center gap-2 border-b border-border/40 bg-surface-2/30 px-4 py-2.5">
          <Terminal size={13} className="text-fg-subtle" />
          <span className="text-[11px] font-semibold text-fg-subtle uppercase tracking-wide">Panel Log Stream</span>
          <span className="relative flex h-1.5 w-1.5 ms-auto">
            <span className="absolute inline-flex h-full w-full animate-ping rounded-full bg-success/60" />
            <span className="inline-flex h-1.5 w-1.5 rounded-full bg-success" />
          </span>
        </div>
        <div className="max-h-[70vh] overflow-auto divide-y divide-white/[0.04] font-mono text-xs">
          {isLoading && <p className="p-5 text-fg-muted">{t("common.loading")}</p>}
          {data?.entries?.slice().reverse().map((e, i) => {
            const lv = levelLabel(e.level);
            return (
              <div key={i} className="flex gap-3 px-4 py-2 hover:bg-white/[0.02]">
                <span className="shrink-0 text-fg-subtle">{new Date(e.time).toLocaleTimeString()}</span>
                <span className={`w-12 shrink-0 font-semibold ${lv.cls}`}>{lv.text}</span>
                <span className="text-fg">{e.message}</span>
                {e.attrs && Object.keys(e.attrs).length > 0 && (
                  <span className="text-fg-subtle">
                    {Object.entries(e.attrs).map(([k, v]) => `${k}=${v}`).join(" ")}
                  </span>
                )}
              </div>
            );
          })}
          {data?.entries?.length === 0 && <p className="p-5 text-fg-muted">{t("common.none")}</p>}
        </div>
      </GlassCard>
    </div>
  );
}
