import { History } from "lucide-react";
import { useAudit } from "@/api/policy-hooks";
import { GlassCard } from "@/components/veltrix";
import { useI18n } from "@/i18n/i18n";
import { useTitle } from "@/lib/useTitle";

export function Audit() {
  useTitle("Audit");
  const { t } = useI18n();
  const { data, isLoading } = useAudit();

  return (
    <div className="space-y-5 animate-page-enter">
      <div>
        <h1 className="text-2xl font-bold text-fg tracking-tight flex items-center gap-2">
          <History size={22} className="text-primary" />
          {t("nav.audit")}
        </h1>
        <p className="text-sm text-fg-muted mt-1">{t("audit.subtitle")}</p>
      </div>

      <GlassCard hover={false} className="!p-0 overflow-hidden">
        <div className="overflow-x-auto">
          <table className="w-full text-sm">
            <thead>
              <tr className="border-b border-border/40 bg-surface-2/30 text-xs text-fg-muted">
                <th className="px-4 py-3 text-start font-medium">{t("audit.time")}</th>
                <th className="px-4 py-3 text-start font-medium">{t("audit.admin")}</th>
                <th className="px-4 py-3 text-start font-medium">{t("audit.action")}</th>
                <th className="px-4 py-3 text-start font-medium">{t("audit.path")}</th>
                <th className="px-4 py-3 text-start font-medium">{t("audit.status")}</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-border/30">
              {isLoading && (
                <tr>
                  <td colSpan={5} className="px-4 py-8 text-center text-fg-muted">{t("common.loading")}</td>
                </tr>
              )}
              {data?.entries?.map((e) => (
                <tr key={e.id} className="hover:bg-surface-2/20">
                  <td className="px-4 py-2.5 text-fg-muted whitespace-nowrap">{new Date(e.time).toLocaleString()}</td>
                  <td className="px-4 py-2.5 font-medium">{e.username || "—"}</td>
                  <td className="px-4 py-2.5">
                    <span className="rounded-md bg-primary/10 px-2 py-0.5 text-xs font-semibold text-primary">{e.method}</span>
                  </td>
                  <td className="px-4 py-2.5 font-mono text-xs text-fg-muted">{e.path}</td>
                  <td className="px-4 py-2.5">{e.status}</td>
                </tr>
              ))}
              {!isLoading && (!data?.entries || data.entries.length === 0) && (
                <tr>
                  <td colSpan={5} className="px-4 py-8 text-center text-fg-muted">{t("common.none")}</td>
                </tr>
              )}
            </tbody>
          </table>
        </div>
      </GlassCard>
    </div>
  );
}
