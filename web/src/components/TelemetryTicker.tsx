import { Activity } from "lucide-react";
import { useOverview } from "@/api/policy-hooks";
import { useI18n } from "@/i18n/i18n";

/** Live fleet strip shown in the header — mirrors the Arena command-tower telemetry bar. */
export function TelemetryTicker() {
  const { data } = useOverview();
  const { t } = useI18n();

  const tel = data?.widgets?.telemetry;
  if (!tel) {
    return (
      <div className="hidden lg:flex flex-1 items-center justify-center px-4 min-w-0">
        <span className="text-[11px] text-fg-subtle truncate">{t("overview.loadingTelemetry")}</span>
      </div>
    );
  }

  return (
    <div className="hidden lg:flex flex-1 items-center justify-center gap-2 px-4 min-w-0">
      <Activity size={13} className={`flex-shrink-0 ${tel.online ? "text-primary animate-pulse" : "text-fg-subtle"}`} />
      <span className="text-[11px] font-mono text-fg-muted truncate">
        [{tel.core}] {tel.location || tel.node_name}
        <span className="text-fg-subtle mx-1.5">·</span>
        {tel.ping_ms != null && tel.ping_ms > 0 ? `${tel.ping_ms}ms` : "—"}
        <span className="text-fg-subtle mx-1.5">·</span>
        {tel.connections} {t("overview.liveConnections")}
        <span className="text-fg-subtle mx-1.5">·</span>
        CPU {tel.cpu_percent.toFixed(0)}%
      </span>
    </div>
  );
}
