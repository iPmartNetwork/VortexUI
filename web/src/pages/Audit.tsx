import { useAudit } from "@/api/policy-hooks";
import { Card, PageHeader } from "@/components/ui";
import { useI18n } from "@/i18n/i18n";

function methodColor(m: string): string {
  return { POST: "text-success", PUT: "text-warning", DELETE: "text-danger", PATCH: "text-accent" }[m] ?? "text-fg-muted";
}
function statusColor(s: number): string {
  if (s >= 500) return "text-danger";
  if (s >= 400) return "text-warning";
  return "text-success";
}

export function Audit() {
  const { data, isLoading } = useAudit();
  const { t } = useI18n();

  return (
    <div>
      <PageHeader title={t("nav.audit")} subtitle={t("audit.subtitle")} />

      <Card className="p-0 text-sm">
        <div className="max-h-[72vh] overflow-auto">
          <table className="w-full">
            <thead className="sticky top-0 border-b bg-surface/80 text-start text-fg-muted backdrop-blur">
              <tr>
                <th className="px-5 py-3 text-start font-medium">{t("audit.time")}</th>
                <th className="px-5 py-3 text-start font-medium">{t("audit.admin")}</th>
                <th className="px-5 py-3 text-start font-medium">{t("audit.action")}</th>
                <th className="px-5 py-3 text-start font-medium">{t("common.status")}</th>
                <th className="px-5 py-3 text-start font-medium">IP</th>
              </tr>
            </thead>
            <tbody>
              {data?.entries?.map((e) => (
                <tr key={e.id} className="border-b border-white/[0.04] hover:bg-white/[0.02]">
                  <td className="whitespace-nowrap px-5 py-2.5 text-fg-subtle">{new Date(e.time).toLocaleString()}</td>
                  <td className="px-5 py-2.5 font-medium">{e.username || "—"}</td>
                  <td className="px-5 py-2.5 font-mono text-xs" dir="ltr">
                    <span className={methodColor(e.method)}>{e.method}</span> <span className="text-fg-muted">{e.path}</span>
                  </td>
                  <td className={`px-5 py-2.5 font-mono ${statusColor(e.status)}`}>{e.status}</td>
                  <td className="px-5 py-2.5 font-mono text-xs text-fg-subtle" dir="ltr">{e.ip}</td>
                </tr>
              ))}
              {!isLoading && data?.entries?.length === 0 && (
                <tr><td colSpan={5} className="px-5 py-10 text-center text-fg-muted">{t("common.none")}</td></tr>
              )}
            </tbody>
          </table>
          {isLoading && <p className="p-5 text-fg-muted">{t("common.loading")}</p>}
        </div>
      </Card>
    </div>
  );
}
