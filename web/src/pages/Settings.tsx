import { useState, useRef, useEffect, useMemo, useCallback } from "react";
import { useSearchParams } from "react-router-dom";
import { motion, AnimatePresence } from "framer-motion";
import {
  Moon, Sun, Monitor, Download, Upload,
  Palette, Key, Copy, Check,
  Settings as SettingsIcon, Shield, Bell, ChevronRight, Save, Database, Users,
} from "lucide-react";
import { QRCodeSVG } from "qrcode.react";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { api } from "@/api/client";
import { useConfirmTOTP, useDisableTOTP, useSetupTOTP } from "@/api/admin-hooks";
import { useExportBackup, useRestoreBackup, useExportUserBackup, useRestoreUserBackup, useBackupManifest, useAPITokens, useCreateAPIToken, useDeleteAPIToken } from "@/api/policy-hooks";
import { Button, Input, Switch } from "@/components/ui";
import { GlassCard } from "@/components/veltrix";
import { useConfirm } from "@/components/confirm";
import { useToast } from "@/components/toast";
import { useTheme } from "@/theme/theme";
import { usePanelSettings, useSavePanelSettings, mergePanelSettings } from "@/api/settings-hooks";
import { applyAccentColor } from "@/theme/branding";
import { useI18n } from "@/i18n/i18n";
import { useAuth } from "@/auth/auth";
import { mergeResellerSettings, type ResellerSettingKey } from "@/auth/permissions";
import type { Lang, TKey } from "@/i18n/dict";
import { useTitle } from "@/lib/useTitle";
import { getApiErrorMessage } from "@/lib/form-errors";
import { cn } from "@/lib/utils";
import { AdminsTab } from "@/pages/Admins";

export type SettingsTab = "general" | "security" | "notifications" | "appearance" | "api" | "backup" | "admins";

const ACCENT_SWATCHES = [
  { id: "blue", color: "#3b82f6" },
  { id: "cyan", color: "#06b6d4" },
  { id: "purple", color: "#6366f1" },
  { id: "orange", color: "#f97316" },
  { id: "red", color: "#ef4444" },
  { id: "green", color: "#22c55e" },
];

const TAB_DEFS: {
  id: SettingsTab;
  icon: typeof SettingsIcon;
  labelKey: TKey;
  flag?: ResellerSettingKey | "notifications" | "user_backup" | "admins";
}[] = [
    { id: "general", icon: SettingsIcon, labelKey: "settings.tabGeneral", flag: "branding" },
    { id: "security", icon: Shield, labelKey: "settings.tabSecurity", flag: "password" },
    { id: "notifications", icon: Bell, labelKey: "settings.tabNotifications", flag: "notifications" },
    { id: "appearance", icon: Palette, labelKey: "settings.tabAppearance", flag: "appearance" },
    { id: "api", icon: Key, labelKey: "settings.tabApiKeys", flag: "api_tokens" },
    { id: "backup", icon: Database, labelKey: "settings.tabBackup", flag: "backup" },
    { id: "admins", icon: Users, labelKey: "settings.tabAdmins" },
  ];

function tabVisible(
  id: SettingsTab,
  show: (key: ResellerSettingKey) => boolean,
  sudo: boolean,
  allowUserBackup?: boolean,
): boolean {
  switch (id) {
    case "admins":
      return sudo;
    case "general":
      return show("config_template") || show("sub_update") || show("branding") || show("update");
    case "security":
      return show("password") || show("totp") || show("ip_guard");
    case "notifications":
      return sudo || show("auto_backup");
    case "appearance":
      return show("appearance") || show("branding");
    case "api":
      return show("api_tokens");
    case "backup":
      return show("backup") || show("auto_backup") || (!sudo && !!allowUserBackup);
    default:
      return false;
  }
}

function parseTab(raw: string | null): SettingsTab {
  if (raw === "security" || raw === "notifications" || raw === "appearance" || raw === "api" || raw === "backup" || raw === "admins") {
    return raw;
  }
  return "general";
}

function ToggleRow({
  label,
  description,
  checked,
  onChange,
}: {
  label: string;
  description?: string;
  checked: boolean;
  onChange: (v: boolean) => void;
}) {
  return (
    <div className="flex items-center justify-between gap-4 py-3 border-b border-border/40 last:border-0">
      <div className="min-w-0">
        <p className="text-sm font-medium text-fg">{label}</p>
        {description && <p className="text-xs text-fg-muted mt-0.5">{description}</p>}
      </div>
      <Switch checked={checked} onCheckedChange={onChange} />
    </div>
  );
}

function PanelBlock({ title, description, children }: {
  title?: string;
  description?: string;
  children: React.ReactNode;
}) {
  return (
    <div className="space-y-3 pt-5 border-t border-border/40 first:border-0 first:pt-0">
      {(title || description) && (
        <div>
          {title && <h3 className="text-sm font-semibold text-fg">{title}</h3>}
          {description && <p className="text-xs text-fg-muted mt-0.5">{description}</p>}
        </div>
      )}
      {children}
    </div>
  );
}

function ThemeChip({ active, onClick, icon, label }: { active: boolean; onClick: () => void; icon: React.ReactNode; label: string }) {
  return (
    <button type="button" onClick={onClick} className={cn(
      "flex flex-1 flex-col items-center gap-1.5 rounded-xl border py-3 text-xs font-medium transition-all",
      active ? "border-primary/50 bg-primary/10 text-primary" : "border-border/50 bg-surface-2/30 text-fg-muted hover:bg-surface-2/60 hover:text-fg",
    )}>
      {icon}
      {label}
    </button>
  );
}

const LANGS: { code: Lang; label: string; flag: string }[] = [
  { code: "en", label: "English", flag: "🇺🇸" },
  { code: "fa", label: "فارسی", flag: "🇮🇷" },
  { code: "tr", label: "Türkçe", flag: "🇹🇷" },
  { code: "ar", label: "العربية", flag: "🇸🇦" },
  { code: "ru", label: "Русский", flag: "🇷🇺" },
  { code: "zh", label: "中文", flag: "🇨🇳" },
  { code: "ja", label: "日本語", flag: "🇯🇵" },
  { code: "es", label: "Español", flag: "🇪🇸" },
];

function LangChip({ active, onClick, flag, label }: { active: boolean; onClick: () => void; flag: string; label: string }) {
  return (
    <button type="button" onClick={onClick} className={cn(
      "flex items-center gap-2 rounded-xl border px-3 py-2 text-xs font-medium transition-all",
      active ? "border-primary/50 bg-primary/10 text-primary" : "border-border/50 bg-surface-2/30 text-fg-muted hover:bg-surface-2/60 hover:text-fg",
    )}>
      <span className="text-base leading-none">{flag}</span>
      {label}
    </button>
  );
}

export function Settings() {
  useTitle("Settings");
  const { t } = useI18n();
  const { sudo, session } = useAuth();
  const flags = mergeResellerSettings(session?.admin.reseller_settings);
  const show = useCallback((key: ResellerSettingKey) => sudo || flags[key], [sudo, flags]);
  const toast = useToast();
  const [searchParams, setSearchParams] = useSearchParams();

  const visibleTabs = useMemo(
    () => TAB_DEFS.filter((tab) => tabVisible(tab.id, show, sudo, session?.admin.allow_user_backup)),
    [show, sudo, session?.admin.allow_user_backup],
  );

  const parsed = parseTab(searchParams.get("tab"));
  const tab = visibleTabs.some((t) => t.id === parsed) ? parsed : (visibleTabs[0]?.id ?? "general");

  useEffect(() => {
    if (parsed !== tab) {
      if (tab === "general") setSearchParams({}, { replace: true });
      else setSearchParams({ tab }, { replace: true });
    }
  }, [parsed, tab, setSearchParams]);

  const saveHandlers = useRef<Partial<Record<SettingsTab, () => void | Promise<void>>>>({});

  function registerSave(id: SettingsTab, fn: () => void | Promise<void>) {
    saveHandlers.current[id] = fn;
  }

  async function handleSaveChanges() {
    const fn = saveHandlers.current[tab];
    if (fn) {
      await fn();
    } else {
      toast.success(t("common.save"));
    }
  }

  function setTab(next: SettingsTab) {
    if (next === "general") setSearchParams({}, { replace: true });
    else setSearchParams({ tab: next }, { replace: true });
  }

  const tabTitleKey: Record<SettingsTab, TKey> = {
    general: "settings.generalTitle",
    security: "settings.securityTitle",
    notifications: "settings.notificationsTitle",
    appearance: "settings.appearanceTitle",
    api: "settings.apiKeysTitle",
    backup: "settings.backupTitle",
    admins: "settings.tabAdmins",
  };

  if (visibleTabs.length === 0) {
    return (
      <div className="animate-page-enter">
        <h1 className="text-2xl font-bold text-fg">{t("nav.settings")}</h1>
        <p className="text-sm text-fg-muted mt-2">No settings available for your account.</p>
      </div>
    );
  }

  return (
    <div className="space-y-5 animate-page-enter">
      <div className="flex flex-col sm:flex-row sm:items-start justify-between gap-4">
        <div>
          <h1 className="text-2xl font-bold text-fg tracking-tight">{t("nav.settings")}</h1>
          <p className="text-sm text-fg-muted mt-1">{t("settings.pageSubtitle")}</p>
        </div>
        <Button onClick={handleSaveChanges} className="shrink-0">
          <Save size={16} />
          {t("settings.saveChanges")}
        </Button>
      </div>

      <div className="grid gap-5 lg:grid-cols-[240px_1fr]">
        <GlassCard className="!p-2 h-fit">
          <nav className="flex flex-col gap-0.5">
            {visibleTabs.map(({ id, icon: Icon, labelKey }) => (
              <button
                key={id}
                type="button"
                onClick={() => setTab(id)}
                className={cn(
                  "flex items-center gap-3 rounded-xl px-3 py-2.5 text-sm font-medium transition-colors text-start",
                  tab === id
                    ? "bg-primary/10 text-primary border-s-2 border-primary"
                    : "text-fg-muted hover:bg-surface-2/50 hover:text-fg border-s-2 border-transparent",
                )}
              >
                <Icon size={16} className="shrink-0" />
                <span className="flex-1">{t(labelKey)}</span>
                <ChevronRight size={14} className="opacity-40 shrink-0 rtl:rotate-180" />
              </button>
            ))}
          </nav>
        </GlassCard>

        <GlassCard>
          <div className="flex items-center justify-between gap-3">
            <h2 className="text-lg font-bold text-fg">{t(tabTitleKey[tab])}</h2>
            <motion.div
              key={tab}
              initial={{ opacity: 0, scale: 0.9 }}
              animate={{ opacity: 1, scale: 1 }}
              className="h-1.5 w-1.5 rounded-full bg-primary/60"
            />
          </div>
          <div className="mt-5 relative">
            <AnimatePresence mode="wait">
              <motion.div
                key={tab}
                initial={{ opacity: 0, y: 8, filter: "blur(2px)" }}
                animate={{ opacity: 1, y: 0, filter: "blur(0px)" }}
                exit={{ opacity: 0, y: -8, filter: "blur(2px)" }}
                transition={{ duration: 0.2, ease: [0.16, 1, 0.3, 1] }}
              >
                {tab === "general" && <GeneralTab show={show} registerSave={registerSave} />}
                {tab === "security" && <SecurityTab show={show} sudo={sudo} registerSave={registerSave} />}
                {tab === "notifications" && <NotificationsTab registerSave={registerSave} />}
                {tab === "appearance" && <AppearanceTab show={show} registerSave={registerSave} />}
                {tab === "api" && <APITokenPanel />}
                {tab === "backup" && (
                  <BackupTab
                    show={show}
                    sudo={sudo}
                    allowUserBackup={!!session?.admin.allow_user_backup}
                    registerSave={registerSave}
                  />
                )}
                {tab === "admins" && sudo && <AdminsTab embedded />}
              </motion.div>
            </AnimatePresence>
          </div>
        </GlassCard>
      </div>
    </div>
  );
}

function GeneralTab({
  show,
  registerSave,
}: {
  show: (key: ResellerSettingKey) => boolean;
  registerSave: (id: SettingsTab, fn: () => void | Promise<void>) => void;
}) {
  const { t } = useI18n();
  const toast = useToast();
  const qc = useQueryClient();
  const { data: panelSettings } = usePanelSettings();
  const saveSettings = useSavePanelSettings();
  const [panelName, setPanelName] = useState("VortexUI");
  const [panelDomain, setPanelDomain] = useState("panel.example.com");
  const [subUrlTemplate, setSubUrlTemplate] = useState("https://sub.example.com/{token}");
  const [autoSync, setAutoSync] = useState(true);
  const [debugMode, setDebugMode] = useState(false);
  const [clashRules, setClashRules] = useState("DOMAIN-SUFFIX,ir,DIRECT\nGEOIP,IR,DIRECT");
  const [singboxDNS, setSingboxDNS] = useState('{"servers":[{"address":"https://dns.google/dns-query","tag":"google"}]}');
  const [hours, setHours] = useState("");
  const [checking, setChecking] = useState(false);
  const [formError, setFormError] = useState<string | null>(null);
  const [updateResult, setUpdateResult] = useState<{ available: boolean; current: string; latest: string } | null>(null);

  const cfg = useQuery({
    queryKey: ["sub-settings"],
    queryFn: () => api<{ config: { update_interval: number } }>("/api/sub-settings"),
    enabled: show("sub_update"),
  });

  const fetched = cfg.data?.config.update_interval;
  if (fetched !== undefined && hours === "") {
    setHours(String(fetched));
  }

  useEffect(() => {
    if (!panelSettings) return;
    setPanelName(panelSettings.panel_name);
    setPanelDomain(panelSettings.panel_domain);
    setSubUrlTemplate(panelSettings.sub_url_template);
    setAutoSync(panelSettings.auto_sync_nodes);
    setDebugMode(panelSettings.debug_mode);
    if (panelSettings.clash_rules_extra) setClashRules(panelSettings.clash_rules_extra);
    if (panelSettings.singbox_dns_extra) setSingboxDNS(panelSettings.singbox_dns_extra);
  }, [panelSettings]);

  const saveSub = useMutation({
    mutationFn: (interval: number) =>
      api("/api/sub-settings", { method: "PUT", body: { update_interval: interval } }),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["sub-settings"] });
    },
  });

  const saveAll = useCallback(async () => {
    setFormError(null);
    try {
      await saveSettings.mutateAsync(mergePanelSettings(panelSettings, {
        panel_name: panelName,
        panel_domain: panelDomain,
        sub_url_template: subUrlTemplate,
        auto_sync_nodes: autoSync,
        debug_mode: debugMode,
        clash_rules_extra: clashRules,
        singbox_dns_extra: singboxDNS,
      }));
      if (show("sub_update")) {
        await saveSub.mutateAsync(Number(hours) || 12);
      }
      toast.success(t("common.save"));
    } catch (err) {
      setFormError(getApiErrorMessage(err, t("settings.saveFailed"), t));
    }
  }, [panelSettings, panelName, panelDomain, subUrlTemplate, autoSync, debugMode, clashRules, singboxDNS, hours, show, saveSub, saveSettings, toast, t]);

  useEffect(() => {
    registerSave("general", saveAll);
  }, [registerSave, saveAll]);

  async function checkUpdate() {
    setChecking(true);
    try {
      const res = await api<{ available: boolean; current: string; latest: string }>("/api/update/check");
      setUpdateResult(res);
      toast.success(res.available ? `Update available: ${res.latest}` : "You're on the latest version!");
    } catch {
      toast.error("Check failed");
    }
    setChecking(false);
  }

  return (
    <div className="space-y-1">
      {formError && (
        <div className="rounded-xl border border-danger/20 bg-danger/10 px-3 py-2">
          <p className="text-sm font-medium text-danger">{formError}</p>
        </div>
      )}

      {show("branding") && (
        <>
          <div className="space-y-3 pb-5">
            <div>
              <p className="mb-1 text-xs font-medium text-fg-muted">{t("settings.panelName")}</p>
              <Input value={panelName} onChange={(e) => setPanelName(e.target.value)} />
            </div>
            <div>
              <p className="mb-1 text-xs font-medium text-fg-muted">{t("settings.panelDomain")}</p>
              <Input value={panelDomain} onChange={(e) => setPanelDomain(e.target.value)} />
            </div>
            <div>
              <p className="mb-1 text-xs font-medium text-fg-muted">{t("settings.subUrlTemplate")}</p>
              <Input value={subUrlTemplate} onChange={(e) => setSubUrlTemplate(e.target.value)} />
            </div>
          </div>
          <ToggleRow label={t("settings.autoSyncNodes")} description={t("settings.autoSyncNodesDesc")} checked={autoSync} onChange={setAutoSync} />
          <ToggleRow label={t("settings.debugMode")} description={t("settings.debugModeDesc")} checked={debugMode} onChange={setDebugMode} />
        </>
      )}

      {show("config_template") && (
        <PanelBlock title="Subscription Templates" description="Customize Clash/sing-box configs delivered to users.">
          <div className="space-y-3">
            <div>
              <p className="mb-1 text-xs font-medium text-fg-muted">Clash — extra routing rules (one per line)</p>
              <textarea
                value={clashRules}
                onChange={(e) => setClashRules(e.target.value)}
                rows={4}
                className="w-full rounded-xl border border-border bg-surface-2/30 px-3 py-2 font-mono text-xs text-fg outline-none focus:ring-1 focus:ring-primary/30"
              />
            </div>
            <div>
              <p className="mb-1 text-xs font-medium text-fg-muted">Sing-box — DNS config (JSON)</p>
              <textarea
                value={singboxDNS}
                onChange={(e) => setSingboxDNS(e.target.value)}
                rows={4}
                className="w-full rounded-xl border border-border bg-surface-2/30 px-3 py-2 font-mono text-xs text-fg outline-none focus:ring-1 focus:ring-primary/30"
              />
            </div>
          </div>
        </PanelBlock>
      )}

      {show("sub_update") && (
        <PanelBlock title={t("settings.subUpdate")} description={t("settings.subUpdateDesc")}>
          <div>
            <p className="mb-1 text-xs font-medium text-fg-muted">{t("settings.subUpdateHours")}</p>
            <Input value={hours} onChange={(e) => setHours(e.target.value)} inputMode="numeric" className="w-32" />
          </div>
        </PanelBlock>
      )}

      {show("update") && (
        <PanelBlock title="System Update" description="Check for new VortexUI releases and update panel + core binaries.">
          <div className="flex flex-wrap items-center gap-4">
            <Button type="button" onClick={checkUpdate} disabled={checking}>
              {checking ? "Checking…" : "Check for updates"}
            </Button>
            {updateResult && (
              <span className="text-sm text-fg-muted">
                Current: <strong>{updateResult.current}</strong>
                {updateResult.available && <> → <strong className="text-success">{updateResult.latest}</strong></>}
              </span>
            )}
          </div>
        </PanelBlock>
      )}
    </div>
  );
}

function SecurityTab({
  show,
  sudo,
  registerSave,
}: {
  show: (key: ResellerSettingKey) => boolean;
  sudo: boolean;
  registerSave: (id: SettingsTab, fn: () => void | Promise<void>) => void;
}) {
  const { t } = useI18n();
  const toast = useToast();
  const { data: panelSettings } = usePanelSettings();
  const saveSettings = useSavePanelSettings();
  const [require2FA, setRequire2FA] = useState(false);
  const [apiAccess, setApiAccess] = useState(true);
  const [formError, setFormError] = useState<string | null>(null);
  const [cur, setCur] = useState("");
  const [nw, setNw] = useState("");
  const [whitelist, setWhitelist] = useState("");
  const [blacklist, setBlacklist] = useState("");

  useEffect(() => {
    if (!panelSettings) return;
    setRequire2FA(panelSettings.require_2fa);
    setApiAccess(panelSettings.api_access_enabled);
    setWhitelist(panelSettings.ip_whitelist);
    setBlacklist(panelSettings.ip_blacklist);
  }, [panelSettings]);

  const changePw = useMutation({
    mutationFn: (b: { current: string; new: string }) => api("/api/account/password", { method: "POST", body: b }),
  });

  const saveAll = useCallback(async () => {
    setFormError(null);
    try {
      await saveSettings.mutateAsync(mergePanelSettings(panelSettings, {
        require_2fa: require2FA,
        api_access_enabled: apiAccess,
        ip_whitelist: whitelist,
        ip_blacklist: blacklist,
      }));
      if (show("password") && cur && nw) {
        await changePw.mutateAsync({ current: cur, new: nw });
        setCur("");
        setNw("");
      }
      toast.success(t("common.save"));
    } catch (err) {
      setFormError(getApiErrorMessage(err, t("settings.saveFailed"), t));
    }
  }, [panelSettings, require2FA, apiAccess, whitelist, blacklist, cur, nw, show, changePw, saveSettings, toast, t]);

  useEffect(() => {
    registerSave("security", saveAll);
  }, [registerSave, saveAll]);

  return (
    <div className="space-y-1">
      {formError && (
        <div className="rounded-xl border border-danger/20 bg-danger/10 px-3 py-2">
          <p className="text-sm font-medium text-danger">{formError}</p>
        </div>
      )}

      {sudo && (
        <>
          <ToggleRow label={t("settings.require2FA")} description={t("settings.require2FADesc")} checked={require2FA} onChange={setRequire2FA} />
          <ToggleRow label={t("settings.apiAccess")} description={t("settings.apiAccessDesc")} checked={apiAccess} onChange={setApiAccess} />
        </>
      )}

      {show("password") && (
        <PanelBlock title={t("settings.adminPassword")}>
          <div className="flex flex-col sm:flex-row gap-2">
            <Input type="password" placeholder="Current password" value={cur} onChange={(e) => setCur(e.target.value)} className="flex-1" />
            <Input type="password" placeholder="New password" value={nw} onChange={(e) => setNw(e.target.value)} className="flex-1" />
            <Button type="button" variant="outline" onClick={() => saveAll()} disabled={!cur || !nw || changePw.isPending}>
              Change
            </Button>
          </div>
        </PanelBlock>
      )}

      {show("totp") && <TwoFAPanel />}

      {show("ip_guard") && (
        <PanelBlock title="IP Access Control" description="Restrict panel access by IP. Comma-separated IPs or CIDRs.">
          <div className="space-y-3">
            <div>
              <p className="mb-1 text-xs font-medium text-fg-muted">Whitelist (only these IPs can access)</p>
              <Input placeholder="e.g. 203.0.113.0/24, 198.51.100.5" value={whitelist} onChange={(e) => setWhitelist(e.target.value)} />
            </div>
            <div>
              <p className="mb-1 text-xs font-medium text-fg-muted">Blacklist (block these IPs)</p>
              <Input placeholder="e.g. 10.0.0.0/8" value={blacklist} onChange={(e) => setBlacklist(e.target.value)} />
            </div>
            <p className="text-[10px] text-fg-subtle">{t("settings.ipGuardHint")}</p>
          </div>
        </PanelBlock>
      )}
    </div>
  );
}

function TwoFAPanel() {
  const { t } = useI18n();
  const toast = useToast();
  const setup = useSetupTOTP();
  const confirmTotp = useConfirmTOTP();
  const disableTotp = useDisableTOTP();
  const [url, setUrl] = useState("");
  const [code, setCode] = useState("");
  const [enabled, setEnabled] = useState(false);

  return (
    <PanelBlock title={t("settings.twoFactor")} description="Protect your account with a time-based one-time password.">
      {enabled ? (
        <form onSubmit={async (e) => {
          e.preventDefault();
          try { await disableTotp.mutateAsync(code); setEnabled(false); setCode(""); toast.success("2FA disabled"); }
          catch { toast.error("Invalid code"); }
        }} className="flex gap-2">
          <Input placeholder="Current 6-digit code" value={code} onChange={(e) => setCode(e.target.value)} inputMode="numeric" />
          <Button type="submit" variant="destructive">Disable 2FA</Button>
        </form>
      ) : url ? (
        <form onSubmit={async (e) => {
          e.preventDefault();
          try { await confirmTotp.mutateAsync(code); setEnabled(true); setUrl(""); setCode(""); toast.success("2FA enabled"); }
          catch { toast.error("Invalid code"); }
        }} className="space-y-4">
          <div className="flex justify-center rounded-2xl bg-white p-5 w-fit mx-auto shadow-lg">
            <QRCodeSVG value={url} size={160} />
          </div>
          <div className="flex gap-2">
            <Input placeholder="6-digit code" value={code} onChange={(e) => setCode(e.target.value)} inputMode="numeric" autoFocus />
            <Button type="submit">{t("common.confirm")}</Button>
          </div>
        </form>
      ) : (
        <div className="flex items-center justify-between gap-4">
          <div className="flex items-center gap-3">
            <div className="h-2 w-2 rounded-full bg-danger/60" />
            <span className="text-sm text-fg-muted">Two-factor authentication is <strong>disabled</strong></span>
          </div>
          <Button type="button" onClick={async () => setUrl((await setup.mutateAsync()).url)} disabled={setup.isPending}>
            Enable 2FA
          </Button>
        </div>
      )}
    </PanelBlock>
  );
}

function NotificationsTab({
  registerSave,
}: {
  registerSave: (id: SettingsTab, fn: () => void | Promise<void>) => void;
}) {
  const { t } = useI18n();
  const toast = useToast();
  const { data: panelSettings } = usePanelSettings();
  const saveSettings = useSavePanelSettings();
  const [push, setPush] = useState(true);
  const [email, setEmail] = useState(false);
  const [formError, setFormError] = useState<string | null>(null);
  const [tgToken, setTgToken] = useState("");

  useEffect(() => {
    if (!panelSettings) return;
    setPush(panelSettings.push_notifications);
    setEmail(panelSettings.email_alerts);
    setTgToken(panelSettings.notify_telegram_token);
  }, [panelSettings]);

  const saveAll = useCallback(async () => {
    setFormError(null);
    try {
      await saveSettings.mutateAsync(mergePanelSettings(panelSettings, {
        push_notifications: push,
        email_alerts: email,
        notify_telegram_token: tgToken,
      }));
      toast.success(t("common.save"));
    } catch (err) {
      setFormError(getApiErrorMessage(err, t("settings.saveFailed"), t));
    }
  }, [panelSettings, push, email, tgToken, saveSettings, toast, t]);

  useEffect(() => {
    registerSave("notifications", saveAll);
  }, [registerSave, saveAll]);

  return (
    <div className="space-y-1">
      {formError && (
        <div className="rounded-xl border border-danger/20 bg-danger/10 px-3 py-2">
          <p className="text-sm font-medium text-danger">{formError}</p>
        </div>
      )}

      <ToggleRow label={t("settings.pushNotifications")} description={t("settings.pushNotificationsDesc")} checked={push} onChange={setPush} />
      <ToggleRow label={t("settings.emailAlerts")} description={t("settings.emailAlertsDesc")} checked={email} onChange={setEmail} />
      <PanelBlock title={t("settings.telegramBotToken")}>
        <Input placeholder="Enter bot token…" value={tgToken} onChange={(e) => setTgToken(e.target.value)} />
      </PanelBlock>
    </div>
  );
}

function AppearanceTab({
  show,
  registerSave,
}: {
  show: (key: ResellerSettingKey) => boolean;
  registerSave: (id: SettingsTab, fn: () => void | Promise<void>) => void;
}) {
  const { t, lang, setLang } = useI18n();
  const { theme, setTheme, resolved } = useTheme();
  const toast = useToast();
  const { data: panelSettings } = usePanelSettings();
  const saveSettings = useSavePanelSettings();
  const [accentColor, setAccentColor] = useState("#6366f1");
  const [logoURL, setLogoURL] = useState("");
  const [formError, setFormError] = useState<string | null>(null);
  const [footerText, setFooterText] = useState("© 2026 iPmart Network. All rights reserved.");

  const isDark = theme === "dark" || (theme === "system" && window.matchMedia("(prefers-color-scheme: dark)").matches);

  useEffect(() => {
    if (!panelSettings) return;
    if (panelSettings.accent_color) setAccentColor(panelSettings.accent_color);
    setLogoURL(panelSettings.logo_url);
    setFooterText(panelSettings.footer_text);
  }, [panelSettings]);

  const selectAccent = useCallback(
    (color: string) => {
      setAccentColor(color);
      applyAccentColor(color, resolved);
    },
    [resolved],
  );

  const saveAll = useCallback(async () => {
    setFormError(null);
    try {
      await saveSettings.mutateAsync(mergePanelSettings(panelSettings, {
        accent_color: accentColor,
        logo_url: logoURL,
        footer_text: footerText,
      }));
      applyAccentColor(accentColor, resolved);
      toast.success(t("common.save"));
    } catch (err) {
      setFormError(getApiErrorMessage(err, t("settings.saveFailed"), t));
    }
  }, [accentColor, logoURL, footerText, panelSettings, resolved, saveSettings, toast, t]);

  useEffect(() => {
    registerSave("appearance", saveAll);
  }, [registerSave, saveAll]);

  function toggleDarkMode(on: boolean) {
    setTheme(on ? "dark" : "light");
  }

  return (
    <div className="space-y-1">
      {formError && (
        <div className="rounded-xl border border-danger/20 bg-danger/10 px-3 py-2">
          <p className="text-sm font-medium text-danger">{formError}</p>
        </div>
      )}

      {show("appearance") && (
        <>
          <ToggleRow label={t("settings.darkMode")} description={t("settings.darkModeDesc")} checked={isDark} onChange={toggleDarkMode} />
          <PanelBlock title={t("theme.label")}>
            <div className="flex gap-2">
              <ThemeChip active={theme === "light"} onClick={() => setTheme("light")} icon={<Sun size={18} />} label={t("theme.light")} />
              <ThemeChip active={theme === "dark"} onClick={() => setTheme("dark")} icon={<Moon size={18} />} label={t("theme.dark")} />
              <ThemeChip active={theme === "system"} onClick={() => setTheme("system")} icon={<Monitor size={18} />} label={t("theme.system")} />
            </div>
          </PanelBlock>
          <PanelBlock title="Language">
            <div className="flex flex-wrap gap-2">
              {LANGS.map((l) => (
                <LangChip key={l.code} active={lang === l.code} onClick={() => setLang(l.code)} flag={l.flag} label={l.label} />
              ))}
            </div>
          </PanelBlock>
        </>
      )}

      {show("branding") && (
        <>
          <PanelBlock title={t("settings.accentColor")}>
            <div className="flex flex-wrap gap-3">
              {ACCENT_SWATCHES.map(({ id, color }) => (
                <button
                  key={id}
                  type="button"
                  onClick={() => selectAccent(color)}
                  className={cn(
                    "h-10 w-10 rounded-xl transition-all",
                    accentColor === color ? "ring-2 ring-primary ring-offset-2 ring-offset-bg-elevated scale-105" : "hover:scale-105",
                  )}
                  style={{ backgroundColor: color }}
                  aria-label={id}
                />
              ))}
            </div>
          </PanelBlock>
          <PanelBlock title="Logo & Footer">
            <div className="space-y-3">
              <Input placeholder="https://example.com/logo.svg" value={logoURL} onChange={(e) => setLogoURL(e.target.value)} />
              <Input value={footerText} onChange={(e) => setFooterText(e.target.value)} />
            </div>
          </PanelBlock>
        </>
      )}
    </div>
  );
}

function APITokenPanel() {
  const { t } = useI18n();
  const toast = useToast();
  const tokens = useAPITokens();
  const create = useCreateAPIToken();
  const del = useDeleteAPIToken();
  const [name, setName] = useState("");
  const [raw, setRaw] = useState("");
  const [copiedId, setCopiedId] = useState<string | null>(null);

  async function handleCreate(e: React.FormEvent) {
    e.preventDefault();
    if (!name.trim()) return;
    try {
      const res = await create.mutateAsync(name.trim());
      setRaw(res.raw);
      setName("");
    } catch {
      toast.error("Create failed");
    }
  }

  function copyText(text: string, id: string) {
    navigator.clipboard.writeText(text);
    setCopiedId(id);
    setTimeout(() => setCopiedId(null), 2000);
  }

  return (
    <div className="space-y-4">
      {raw && (
        <div className="rounded-xl border border-success/30 bg-success/5 p-4 space-y-2">
          <p className="text-xs font-semibold text-success">Copy now — won&apos;t be shown again</p>
          <div className="flex items-center gap-2 rounded-lg bg-bg/60 px-3 py-2 font-mono text-xs">
            <span className="flex-1 break-all text-fg">{raw}</span>
            <button type="button" onClick={() => copyText(raw, "new")} className="shrink-0 rounded-md p-1 hover:bg-surface-2/60">
              {copiedId === "new" ? <Check size={14} className="text-success" /> : <Copy size={14} />}
            </button>
          </div>
        </div>
      )}

      {tokens.data?.tokens && tokens.data.tokens.length > 0 ? (
        <div className="divide-y divide-border/40 rounded-xl border border-border/40">
          {tokens.data.tokens.map((tok) => (
            <div key={tok.id} className="flex items-center justify-between gap-3 px-4 py-3">
              <div className="min-w-0">
                <p className="text-sm font-medium text-fg">{tok.name}</p>
                <p className="text-xs font-mono text-fg-subtle truncate">vx_tk_{"•".repeat(16)}</p>
              </div>
              <div className="flex items-center gap-2 shrink-0">
                <button type="button" onClick={() => copyText(tok.id, tok.id)} className="text-xs font-medium text-primary hover:underline">
                  Copy
                </button>
                <button
                  type="button"
                  onClick={() => del.mutateAsync(tok.id).then(() => toast.success("Token revoked"))}
                  className="text-xs font-medium text-danger hover:underline"
                >
                  Revoke
                </button>
              </div>
            </div>
          ))}
        </div>
      ) : !tokens.isLoading ? (
        <p className="text-sm text-fg-subtle py-2">No tokens yet</p>
      ) : null}

      <form onSubmit={handleCreate} className="flex gap-2 pt-2">
        <Input placeholder="Token name (e.g. CI Deploy)" value={name} onChange={(e) => setName(e.target.value)} className="flex-1" />
        <Button type="submit" variant="outline" disabled={create.isPending}>
          <Key size={14} />
          {t("settings.generateKey")}
        </Button>
      </form>
    </div>
  );
}

function formatBytes(n: number) {
  if (!n || n <= 0) return "0 B";
  const units = ["B", "KB", "MB", "GB", "TB"];
  let v = n;
  let i = 0;
  while (v >= 1024 && i < units.length - 1) {
    v /= 1024;
    i++;
  }
  return `${v.toFixed(i === 0 ? 0 : 1)} ${units[i]}`;
}

function BackupTab({
  show,
  sudo,
  allowUserBackup,
  registerSave,
}: {
  show: (key: ResellerSettingKey) => boolean;
  sudo: boolean;
  allowUserBackup: boolean;
  registerSave: (id: SettingsTab, fn: () => void | Promise<void>) => void;
}) {
  const { t } = useI18n();
  const toast = useToast();
  const confirm = useConfirm();
  const { data: panelSettings } = usePanelSettings();
  const saveSettings = useSavePanelSettings();
  const exportBackup = useExportBackup();
  const exportUsers = useExportUserBackup();
  const restoreBackup = useRestoreBackup();
  const restoreUsers = useRestoreUserBackup();
  const { data: manifest } = useBackupManifest(sudo && show("backup"));
  const fileRef = useRef<HTMLInputElement>(null);
  const userFileRef = useRef<HTMLInputElement>(null);
  const [autoBackup, setAutoBackup] = useState(true);
  const [s3Endpoint, setS3Endpoint] = useState("");
  const [s3Bucket, setS3Bucket] = useState("");
  const [interval, setInterval] = useState("24");
  const [tgChat, setTgChat] = useState("");
  const [passphrase, setPassphrase] = useState("");
  const [restoreMode, setRestoreMode] = useState<"config" | "full">("config");
  const [includeTraffic, setIncludeTraffic] = useState(false);
  const [formError, setFormError] = useState<string | null>(null);

  const usersOnly = !sudo && allowUserBackup;
  const canRestore = sudo && show("backup");
  const canRestoreUsers = usersOnly || (sudo && allowUserBackup);

  useEffect(() => {
    if (!panelSettings) return;
    setAutoBackup(panelSettings.auto_backup_enabled);
    setS3Endpoint(panelSettings.auto_backup_s3_endpoint);
    setS3Bucket(panelSettings.auto_backup_s3_bucket);
    setInterval(String(panelSettings.auto_backup_interval_hours || 24));
    setTgChat(panelSettings.auto_backup_telegram_chat_id);
  }, [panelSettings]);

  const saveAuto = useCallback(async () => {
    setFormError(null);
    try {
      await saveSettings.mutateAsync(mergePanelSettings(panelSettings, {
        auto_backup_enabled: autoBackup,
        auto_backup_s3_endpoint: s3Endpoint,
        auto_backup_s3_bucket: s3Bucket,
        auto_backup_interval_hours: Number(interval) || 24,
        auto_backup_telegram_chat_id: tgChat,
      }));
      toast.success(t("common.save"));
    } catch (err) {
      setFormError(getApiErrorMessage(err, t("settings.saveFailed"), t));
    }
  }, [autoBackup, s3Endpoint, s3Bucket, interval, tgChat, panelSettings, saveSettings, toast, t]);

  useEffect(() => {
    registerSave("backup", saveAuto);
  }, [registerSave, saveAuto]);

  async function handleRestore() {
    const file = fileRef.current?.files?.[0];
    if (!file) return;
    const ok = await confirm({
      title: restoreMode === "full" ? "Restore full database?" : "Restore backup?",
      message:
        restoreMode === "full"
          ? "This REPLACES the entire PostgreSQL database. All current data will be lost."
          : "This REPLACES configuration, users, billing tables included in the backup. Current data in those areas will be lost.",
      confirmLabel: "Restore",
      destructive: true,
    });
    if (!ok) return;
    try {
      const res = await restoreBackup.mutateAsync({ file, mode: restoreMode, passphrase: passphrase || undefined });
      const r = res.restored;
      toast.success(`Restored (${res.mode}): ${r.nodes ?? 0} nodes, ${r.users ?? 0} users, ${r.orders ?? 0} orders`);
      if (res.warnings?.length) toast.info(res.warnings.join(" · "));
    } catch (e: unknown) {
      toast.error(e instanceof Error ? e.message : "Restore failed");
    }
    if (fileRef.current) fileRef.current.value = "";
  }

  async function handleUserRestore() {
    const file = userFileRef.current?.files?.[0];
    if (!file) return;
    const ok = await confirm({
      title: "Restore my users backup?",
      message: "This updates your users, bindings, wallet ledger, and orders from the backup file.",
      confirmLabel: "Restore",
      destructive: true,
    });
    if (!ok) return;
    try {
      const res = await restoreUsers.mutateAsync(file);
      toast.success(`Restored ${res.restored.users ?? 0} users, ${res.restored.orders ?? 0} orders`);
    } catch (e: unknown) {
      toast.error(e instanceof Error ? e.message : "Restore failed");
    }
    if (userFileRef.current) userFileRef.current.value = "";
  }

  return (
    <div className="space-y-1">
      {formError && (
        <div className="rounded-xl border border-danger/20 bg-danger/10 px-3 py-2">
          <p className="text-sm font-medium text-danger">{formError}</p>
        </div>
      )}

      {show("auto_backup") && sudo && (
        <>
          <ToggleRow label={t("settings.autoBackup")} description={t("settings.autoBackupDesc")} checked={autoBackup} onChange={setAutoBackup} />
          <p className="text-xs text-fg-muted py-2">{t("settings.lastBackup").replace("{time}", "—").replace("{size}", "—")}</p>
        </>
      )}

      {(show("backup") || usersOnly) && (
        <PanelBlock title={usersOnly ? "My users backup" : t("settings.backup")}>
          {manifest && !usersOnly && (
            <div className="mb-4 rounded-xl border border-border-strong/60 bg-surface/40 p-3 text-xs text-fg-muted space-y-2">
              <p className="font-medium text-fg">Backup preview (v{manifest.version})</p>
              <div className="grid grid-cols-2 gap-x-4 gap-y-1 sm:grid-cols-3">
                <span>{manifest.counts.nodes} nodes</span>
                <span>{manifest.counts.users} users</span>
                <span>{manifest.counts.admins} admins</span>
                <span>{manifest.counts.orders} orders</span>
                <span>{manifest.counts.wallet_ledger} ledger rows</span>
                <span>{manifest.counts.wallet_deposits} deposits</span>
              </div>
              <p>
                Traffic: {formatBytes(manifest.usage.total_used_traffic)} used /{" "}
                {formatBytes(manifest.usage.total_remaining_traffic)} remaining
                {manifest.usage.users_over_limit > 0 ? ` · ${manifest.usage.users_over_limit} over limit` : ""}
              </p>
              {manifest.warnings?.length ? (
                <p className="text-warning">{manifest.warnings.join(" · ")}</p>
              ) : null}
            </div>
          )}
          {!usersOnly && (
            <div className="mb-3 flex flex-wrap gap-3 items-end">
              <div>
                <p className="mb-1 text-xs font-medium text-fg-muted">Encryption passphrase (optional)</p>
                <Input value={passphrase} onChange={(e) => setPassphrase(e.target.value)} type="password" placeholder="AES-256 if set on export" className="max-w-xs" />
              </div>
              <label className="flex items-center gap-2 text-xs text-fg-muted pb-2">
                <input type="checkbox" checked={includeTraffic} onChange={(e) => setIncludeTraffic(e.target.checked)} />
                Include traffic time-series
              </label>
            </div>
          )}
          <div className="flex flex-wrap gap-3">
            <Button
              type="button"
              variant="outline"
              onClick={() =>
                usersOnly
                  ? exportUsers.mutate(undefined, { onError: (e) => toast.error(e instanceof Error ? e.message : "Export failed") })
                  : exportBackup.mutate(
                    { format: "json", passphrase: passphrase || undefined, includeTraffic },
                    { onError: (e) => toast.error(e instanceof Error ? e.message : "Export failed") },
                  )
              }
              disabled={usersOnly ? exportUsers.isPending : exportBackup.isPending}
            >
              <Download size={15} />
              {usersOnly ? "Export my users" : t("settings.backupNow")}
            </Button>
            {!usersOnly && (
              <Button
                type="button"
                variant="outline"
                onClick={() =>
                  exportBackup.mutate(
                    { format: "full" },
                    { onError: (e) => toast.error(e instanceof Error ? e.message : "Full DB backup failed") },
                  )
                }
                disabled={exportBackup.isPending}
              >
                <Download size={15} />
                Full DB backup
              </Button>
            )}
            {canRestore && (
              <>
                <select
                  value={restoreMode}
                  onChange={(e) => setRestoreMode(e.target.value as "config" | "full")}
                  className="rounded-xl border border-border-strong/80 bg-surface/60 px-3 py-2 text-sm"
                >
                  <option value="config">Restore JSON config</option>
                  <option value="full">Restore full DB (.tar.gz)</option>
                </select>
                <label className="cursor-pointer">
                  <input
                    ref={fileRef}
                    type="file"
                    accept={restoreMode === "full" ? ".tar.gz,.gz" : ".json,.bin"}
                    className="hidden"
                    onChange={handleRestore}
                  />
                  <span className="inline-flex items-center gap-2 rounded-xl border border-border-strong/80 bg-surface/60 px-4 py-2 text-sm font-medium text-fg transition hover:bg-surface-2/80">
                    <Upload size={15} />
                    {t("settings.restore")}
                  </span>
                </label>
              </>
            )}
            {canRestoreUsers && usersOnly && (
              <label className="cursor-pointer">
                <input ref={userFileRef} type="file" accept=".json" className="hidden" onChange={handleUserRestore} />
                <span className="inline-flex items-center gap-2 rounded-xl border border-border-strong/80 bg-surface/60 px-4 py-2 text-sm font-medium text-fg transition hover:bg-surface-2/80">
                  <Upload size={15} />
                  Restore my users
                </span>
              </label>
            )}
          </div>
        </PanelBlock>
      )}

      {show("auto_backup") && sudo && (
        <PanelBlock title="Automatic Backup" description="Scheduled backup to Telegram chat or S3-compatible storage.">
          <div className="space-y-3">
            <Input placeholder="Chat ID" value={tgChat} onChange={(e) => setTgChat(e.target.value)} />
            <div className="grid grid-cols-2 gap-2">
              <Input placeholder="S3 endpoint URL" value={s3Endpoint} onChange={(e) => setS3Endpoint(e.target.value)} />
              <Input placeholder="Bucket name" value={s3Bucket} onChange={(e) => setS3Bucket(e.target.value)} />
            </div>
            <div>
              <p className="mb-1 text-xs font-medium text-fg-muted">Interval (hours)</p>
              <Input value={interval} onChange={(e) => setInterval(e.target.value)} inputMode="numeric" className="w-24" />
            </div>
            <p className="text-[10px] text-fg-subtle">Configure via env: VORTEX_BACKUP_TG_TOKEN, VORTEX_BACKUP_TG_CHAT, VORTEX_BACKUP_S3_*</p>
          </div>
        </PanelBlock>
      )}
    </div>
  );
}
