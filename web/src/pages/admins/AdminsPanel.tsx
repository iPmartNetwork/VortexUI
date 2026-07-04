import { useSearchParams } from "react-router-dom";
import { Shield, Users, UserCog } from "lucide-react";
import { useI18n } from "@/i18n/i18n";
import type { TKey } from "@/i18n/dict";
import { cn } from "@/lib/utils";
import { AdminsListTab } from "@/pages/admins/AdminsListTab";
import { RolesTab } from "@/pages/admins/RolesTab";
import { AccessSettingsTab } from "@/pages/admins/AccessSettingsTab";

export type AdminsSection = "list" | "roles" | "access";

const SECTIONS: { id: AdminsSection; icon: typeof Users; labelKey: TKey }[] = [
  { id: "list", icon: Users, labelKey: "settings.adminsSection.list" },
  { id: "roles", icon: Shield, labelKey: "settings.adminsSection.roles" },
  { id: "access", icon: UserCog, labelKey: "settings.adminsSection.access" },
];

function parseSection(raw: string | null): AdminsSection {
  if (raw === "roles" || raw === "access") return raw;
  return "list";
}

export function AdminsPanel({ embedded = false }: { embedded?: boolean }) {
  const { t } = useI18n();
  const [searchParams, setSearchParams] = useSearchParams();
  const section = parseSection(searchParams.get("section"));

  function setSection(next: AdminsSection) {
    const params = new URLSearchParams(searchParams);
    params.set("tab", "admins");
    if (next === "list") params.delete("section");
    else params.set("section", next);
    setSearchParams(params, { replace: true });
  }

  return (
    <div className={embedded ? "space-y-5" : "space-y-6 animate-page-enter"}>
      {!embedded && (
        <h1 className="text-2xl font-bold tracking-tight">{t("reseller.admins.pageTitle")}</h1>
      )}

      <div className="flex flex-wrap rounded-xl border border-border/70 bg-surface-2/50 p-0.5 gap-0.5">
        {SECTIONS.map(({ id, icon: Icon, labelKey }) => (
          <button
            key={id}
            type="button"
            onClick={() => setSection(id)}
            className={cn(
              "flex items-center gap-1.5 px-3 py-2 rounded-lg text-xs font-semibold transition-colors whitespace-nowrap",
              section === id ? "bg-primary text-primary-fg shadow-sm" : "text-fg-muted hover:text-fg",
            )}
          >
            <Icon size={14} />
            {t(labelKey)}
          </button>
        ))}
      </div>

      {section === "list" && <AdminsListTab embedded={embedded} />}
      {section === "roles" && <RolesTab />}
      {section === "access" && <AccessSettingsTab />}
    </div>
  );
}
