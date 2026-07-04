import { useState } from "react";
import { useLogs } from "@/api/policy-hooks";
import { Card, PageHeader, Select } from "@/components/ui";
import { useI18n } from "@/i18n/i18n";

const LEVELS = ["debug", "info", "warn", "error"];

// slog numeric levels: debug=-4, info=0, warn=4, error=8.
function levelLabel(n: number): { text: string; cls: string } {
  if (n >= 8) return { text: "ERROR", cls: "text-danger" };
  if (n >= 4) return { text: "WARN", cls: "text-warning" };
  if (n >= 0) return { text: "INFO", cls: "text-accent" };
  return { text: "DEBUG", cls: "text-fg-subtle" };
}

export function Logs() {
  const { t } = useI18n();
  const [level, setLevel] = useState("info");
  const { data, isLoading } = useLogs(level);

  return (
    <div className="space-y-6 animate-page-enter">
      <PageHeader title={t("nav.logs")} subtitle="Panel activity">
        <Select className="w-32" value={level} onChange={(e) => setLevel(e.target.value)}>
          {LEVELS.map((l) => <option key={l} value={l}>{l}</option>)}
        </Select>
      </PageHeader>

      <Card className="p-0 font-mono text-xs">
        <div className="max-h-[70vh] overflow-auto divide-y divide-white/[0.04]">
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
      </Card>
    </div>
  );
}
