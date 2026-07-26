import { useSearchParams } from "react-router-dom";
import { motion } from "framer-motion";
import { Shield, Users, UserCog, Sparkles } from "lucide-react";
import { useI18n } from "@/i18n/i18n";
import type { TKey } from "@/i18n/dict";
import { cn } from "@/lib/utils";
import { AdminsListTab } from "@/pages/admins/AdminsListTab";
import { RolesTab } from "@/pages/admins/RolesTab";
import { AccessSettingsTab } from "@/pages/admins/AccessSettingsTab";

export type AdminsSection = "list" | "roles" | "access";

const SECTIONS: { id: AdminsSection; icon: typeof Users; labelKey: TKey; descKey?: TKey }[] = [
  {
    id: "list",
    icon: Users,
    labelKey: "settings.adminsSection.list",
    descKey: "settings.adminsSection.list",
  },
  {
    id: "roles",
    icon: Shield,
    labelKey: "settings.adminsSection.roles",
    descKey: "settings.adminsSection.rolesDesc",
  },
  {
    id: "access",
    icon: UserCog,
    labelKey: "settings.adminsSection.access",
    descKey: "settings.adminsSection.accessDesc",
  },
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

  const currentSection = SECTIONS.find((s) => s.id === section)!;

  return (
    <div className={embedded ? "space-y-5" : "space-y-6 animate-page-enter"}>
      {!embedded && (
        <div className="flex items-center gap-3">
          <h1 className="text-2xl font-bold tracking-tight text-fg">
            {t("reseller.admins.pageTitle")}
          </h1>
          <span className="inline-flex items-center gap-1 rounded-md bg-primary/10 px-2 py-0.5 text-[10px] font-bold text-primary border border-primary/20">
            <Sparkles size={10} /> MANAGEMENT
          </span>
        </div>
      )}

      {/* Cyber Section Switcher */}
      <div className="glass rounded-2xl p-1.5">
        <div className="flex flex-wrap gap-1">
          {SECTIONS.map(({ id, icon: Icon, labelKey }) => (
            <button
              key={id}
              type="button"
              onClick={() => setSection(id)}
              className={cn(
                "relative flex items-center gap-2 px-4 py-2.5 rounded-xl text-xs font-semibold transition-all duration-200",
                section === id
                  ? "bg-primary text-primary-fg shadow-cyber"
                  : "text-fg-muted hover:text-fg hover:bg-surface-2/40",
              )}
            >
              {section === id && (
                <motion.div
                  layoutId="active-section"
                  className="absolute inset-0 rounded-xl bg-primary"
                  transition={{ type: "spring", stiffness: 400, damping: 28 }}
                />
              )}
              <Icon size={14} className="relative z-10" />
              <span className="relative z-10">{t(labelKey)}</span>
            </button>
          ))}
        </div>
      </div>

      {/* Section Description */}
      <p className="text-xs text-fg-muted -mt-3">
        {t(currentSection.descKey ?? currentSection.labelKey)}
      </p>

      {/* Content with fade transition */}
      <motion.div
        key={section}
        initial={{ opacity: 0, y: 8 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ duration: 0.25, ease: [0.16, 1, 0.3, 1] }}
      >
        {section === "list" && <AdminsListTab embedded={embedded} />}
        {section === "roles" && <RolesTab />}
        {section === "access" && <AccessSettingsTab />}
      </motion.div>
    </div>
  );
}
