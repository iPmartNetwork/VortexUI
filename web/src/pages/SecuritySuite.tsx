import { useSearchParams } from "react-router-dom";
import { Crosshair, Scan, SlidersHorizontal, Target } from "lucide-react";
import { Badge } from "@/components/ui";
import { useI18n } from "@/i18n/i18n";
import type { TKey } from "@/i18n/dict";
import { useTitle } from "@/lib/useTitle";
import { cn } from "@/lib/utils";
import { RealityTab } from "@/pages/security/RealityTab";
import { CleanIPTab } from "@/pages/security/CleanIPTab";
import { TLSTricksTab } from "@/pages/security/TLSTricksTab";
import { DecoyProbingTab } from "@/pages/security/DecoyProbingTab";

export type SecurityTab = "reality" | "cleanip" | "tls" | "decoy";

const TABS: { id: SecurityTab; icon: typeof Scan; labelKey: TKey }[] = [
  { id: "reality", icon: Scan, labelKey: "security.tabReality" },
  { id: "cleanip", icon: Crosshair, labelKey: "security.tabCleanIP" },
  { id: "tls", icon: SlidersHorizontal, labelKey: "security.tabTLS" },
  { id: "decoy", icon: Target, labelKey: "security.tabDecoy" },
];

function parseTab(raw: string | null): SecurityTab {
  if (raw === "cleanip" || raw === "tls" || raw === "decoy") return raw;
  return "reality";
}

export function SecuritySuite() {
  useTitle("Security");
  const { t } = useI18n();
  const [searchParams, setSearchParams] = useSearchParams();
  const tab = parseTab(searchParams.get("tab"));

  function setTab(next: SecurityTab) {
    if (next === "reality") setSearchParams({}, { replace: true });
    else setSearchParams({ tab: next }, { replace: true });
  }

  return (
    <div className="space-y-5 animate-page-enter">
      <div className="flex flex-col lg:flex-row lg:items-start justify-between gap-4">
        <div>
          <div className="flex flex-wrap items-center gap-2">
            <h1 className="text-2xl font-bold text-fg tracking-tight">{t("security.pageTitle")}</h1>
            <Badge color="muted">{t("security.badge")}</Badge>
          </div>
          <p className="text-sm text-fg-muted mt-1 max-w-2xl">{t("security.pageSubtitle")}</p>
          <p className="text-xs text-amber-600 dark:text-amber-400 mt-2 max-w-2xl">
            {t("security.runtimeNote")}
          </p>
        </div>
        <div className="flex flex-wrap rounded-xl border border-border/70 bg-surface-2/50 p-0.5 gap-0.5">
          {TABS.map(({ id, icon: Icon, labelKey }) => (
            <button
              key={id}
              type="button"
              onClick={() => setTab(id)}
              className={cn(
                "flex items-center gap-1.5 px-3 py-2 rounded-lg text-xs font-semibold transition-colors whitespace-nowrap",
                tab === id ? "bg-primary text-primary-fg shadow-sm" : "text-fg-muted hover:text-fg",
              )}
            >
              <Icon size={14} />
              {t(labelKey)}
            </button>
          ))}
        </div>
      </div>

      {tab === "reality" && <RealityTab />}
      {tab === "cleanip" && <CleanIPTab />}
      {tab === "tls" && <TLSTricksTab />}
      {tab === "decoy" && <DecoyProbingTab />}
    </div>
  );
}