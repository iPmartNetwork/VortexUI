import { useState, useRef } from "react";
import {
  Moon, Sun, Monitor, ShieldCheck, Download, Upload,
  Palette, Lock, Cpu, Key, Copy, Trash2, Check, RefreshCw,
} from "lucide-react";
import { QRCodeSVG } from "qrcode.react";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { api } from "@/api/client";
import { useConfirmTOTP, useDisableTOTP, useSetupTOTP } from "@/api/admin-hooks";
import { useExportBackup, useRestoreBackup, useExportUserBackup, useAPITokens, useCreateAPIToken, useDeleteAPIToken } from "@/api/policy-hooks";
import { Button, Card, Input, PageHeader } from "@/components/ui";
import { useConfirm } from "@/components/confirm";
import { useToast } from "@/components/toast";
import { useTheme } from "@/theme/theme";
import { useI18n } from "@/i18n/i18n";
import { useAuth } from "@/auth/auth";
import { mergeResellerSettings, type ResellerSettingKey } from "@/auth/permissions";
import type { Lang } from "@/i18n/dict";
import { cn } from "@/lib/utils";

// ─── Section wrapper with icon + title ───────────────────────────────────────
function Section({ icon, title, description, children }: {
  icon: React.ReactNode; title: string; description?: string; children: React.ReactNode;
}) {
  return (
    <Card className="p-0 overflow-hidden">
      <div className="flex items-start gap-4 border-b border-border/40 px-5 py-4">
        <div className="mt-0.5 grid h-9 w-9 shrink-0 place-items-center rounded-xl bg-primary/10 text-primary">
          {icon}
        </div>
        <div>
          <h2 className="text-sm font-semibold text-fg">{title}</h2>
          {description && <p className="mt-0.5 text-xs text-fg-muted">{description}</p>}
        </div>
      </div>
      <div className="p-5">{children}</div>
    </Card>
  );
}

// ─── Theme chip ───────────────────────────────────────────────────────────────
function ThemeChip({ active, onClick, icon, label }: { active: boolean; onClick: () => void; icon: React.ReactNode; label: string }) {
  return (
    <button onClick={onClick} className={cn(
      "flex flex-1 flex-col items-center gap-1.5 rounded-xl border py-3 text-xs font-medium transition-all",
      active ? "border-primary/50 bg-primary/10 text-primary" : "border-border/50 bg-surface-2/30 text-fg-muted hover:bg-surface-2/60 hover:text-fg",
    )}>
      {icon}
      {label}
    </button>
  );
}

// ─── Lang chip ────────────────────────────────────────────────────────────────
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
    <button onClick={onClick} className={cn(
      "flex items-center gap-2 rounded-xl border px-3 py-2 text-xs font-medium transition-all",
      active ? "border-primary/50 bg-primary/10 text-primary" : "border-border/50 bg-surface-2/30 text-fg-muted hover:bg-surface-2/60 hover:text-fg",
    )}>
      <span className="text-base leading-none">{flag}</span>
      {label}
    </button>
  );
}

// ─── Main Settings Page ───────────────────────────────────────────────────────
export function Settings() {
  const { t } = useI18n();
  const { theme, setTheme } = useTheme();
  const { lang, setLang } = useI18n();
  const { sudo, session } = useAuth();
  const flags = mergeResellerSettings(session?.admin.reseller_settings);
  const show = (key: ResellerSettingKey) => sudo || flags[key];

  return (
    <div className="mx-auto max-w-4xl space-y-5 animate-page-enter">
      <PageHeader title={t("nav.settings")} />

      {show("appearance") && (
      <Section icon={<Palette size={16} />} title={t("settings.appearance")}>
        <div className="space-y-5">
          <div>
            <p className="mb-2 text-xs font-medium text-fg-muted">{t("theme.label")}</p>
            <div className="flex gap-2">
              <ThemeChip active={theme === "light"} onClick={() => setTheme("light")} icon={<Sun size={18} />} label={t("theme.light")} />
              <ThemeChip active={theme === "dark"} onClick={() => setTheme("dark")} icon={<Moon size={18} />} label={t("theme.dark")} />
              <ThemeChip active={theme === "system"} onClick={() => setTheme("system")} icon={<Monitor size={18} />} label={t("theme.system")} />
            </div>
          </div>
          <div>
            <p className="mb-2 text-xs font-medium text-fg-muted">Language</p>
            <div className="flex flex-wrap gap-2">
              {LANGS.map((l) => (
                <LangChip key={l.code} active={lang === l.code} onClick={() => setLang(l.code)} flag={l.flag} label={l.label} />
              ))}
            </div>
          </div>
        </div>
      </Section>
      )}

      {show("password") && <PasswordSection />}
      {show("totp") && <TwoFASection />}
      {show("api_tokens") && <APITokenSection />}
      {show("backup") || (!sudo && !!session?.admin.allow_user_backup) ? (
        <BackupSection usersOnly={!sudo} allowRestore={sudo} />
      ) : null}
      {show("config_template") && <ConfigTemplateSection />}
      {show("sub_update") && <SubUpdateSection />}
      {show("ip_guard") && <IPGuardSection />}
      {show("branding") && <BrandingSection />}
      {show("auto_backup") && <AutoBackupSection />}
      {show("update") && <UpdateSection />}
    </div>
  );
}

// ─── Password Section ─────────────────────────────────────────────────────────
function PasswordSection() {
  const { t } = useI18n();
  const toast = useToast();
  const [cur, setCur] = useState("");
  const [nw, setNw] = useState("");
  const changePw = useMutation({
    mutationFn: (b: { current: string; new: string }) => api("/api/account/password", { method: "POST", body: b }),
  });

  async function submit(e: React.FormEvent) {
    e.preventDefault();
    try {
      await changePw.mutateAsync({ current: cur, new: nw });
      toast.success("Password changed");
      setCur(""); setNw("");
    } catch { toast.error("Wrong current password"); }
  }

  return (
    <Section icon={<Lock size={16} />} title={t("settings.password")}>
      <form onSubmit={submit} className="space-y-3">
        <Input type="password" placeholder="Current password" value={cur} onChange={(e) => setCur(e.target.value)} required />
        <Input type="password" placeholder="New password" value={nw} onChange={(e) => setNw(e.target.value)} required />
        <Button type="submit" disabled={changePw.isPending}>{t("common.save")}</Button>
      </form>
    </Section>
  );
}

// ─── 2FA Section ──────────────────────────────────────────────────────────────
function TwoFASection() {
  const { t } = useI18n();
  const toast = useToast();
  const setup = useSetupTOTP();
  const confirmTotp = useConfirmTOTP();
  const disableTotp = useDisableTOTP();
  const [url, setUrl] = useState("");
  const [code, setCode] = useState("");
  const [enabled, setEnabled] = useState(false);

  return (
    <Section icon={<ShieldCheck size={16} />} title={t("settings.twoFactor")} description="Protect your account with a time-based one-time password.">
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
          <p className="text-center text-sm text-fg-muted">Scan with your authenticator app, then enter the code below.</p>
          <div className="flex gap-2">
            <Input placeholder="6-digit code" value={code} onChange={(e) => setCode(e.target.value)} inputMode="numeric" autoFocus />
            <Button type="submit">{t("common.confirm")}</Button>
          </div>
        </form>
      ) : (
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-3">
            <div className="h-2 w-2 rounded-full bg-fg-subtle/50" />
            <span className="text-sm text-fg-muted">Two-factor authentication is <strong>disabled</strong></span>
          </div>
          <Button onClick={async () => setUrl((await setup.mutateAsync()).url)} disabled={setup.isPending}>
            Enable 2FA
          </Button>
        </div>
      )}
    </Section>
  );
}

// ─── API Token Section ────────────────────────────────────────────────────────
function APITokenSection() {
  const toast = useToast();
  const tokens = useAPITokens();
  const create = useCreateAPIToken();
  const del = useDeleteAPIToken();
  const [name, setName] = useState("");
  const [raw, setRaw] = useState("");
  const [copied, setCopied] = useState(false);

  async function handleCreate(e: React.FormEvent) {
    e.preventDefault();
    if (!name.trim()) return;
    try {
      const res = await create.mutateAsync(name.trim());
      setRaw(res.raw);
      setName("");
    } catch { toast.error("Create failed"); }
  }

  function copyRaw() {
    navigator.clipboard.writeText(raw);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  }

  return (
    <Section icon={<Key size={16} />} title="API Tokens" description="Long-lived tokens for automation. The secret is shown once and never stored.">
      <div className="space-y-4">
        {/* Create form */}
        <form onSubmit={handleCreate} className="flex gap-2">
          <Input placeholder="Token name (e.g. CI Deploy, Telegram Bot)" value={name} onChange={(e) => setName(e.target.value)} className="flex-1" />
          <Button type="submit" disabled={create.isPending}>Create</Button>
        </form>

        {/* One-time secret display */}
        {raw && (
          <div className="rounded-xl border border-success/30 bg-success/5 p-4 space-y-3">
            <div className="flex items-center gap-2">
              <div className="h-2 w-2 rounded-full bg-success animate-pulse" />
              <span className="text-xs font-semibold text-success">Copy now — won't be shown again</span>
            </div>
            <div className="flex items-center gap-2 rounded-lg bg-bg/60 px-3 py-2 font-mono text-xs">
              <span className="flex-1 break-all text-fg">{raw}</span>
              <button onClick={copyRaw} className="shrink-0 rounded-md p-1 transition hover:bg-surface-2/60">
                {copied ? <Check size={14} className="text-success" /> : <Copy size={14} className="text-fg-muted" />}
              </button>
            </div>
          </div>
        )}

        {/* Token list */}
        {tokens.data?.tokens && tokens.data.tokens.length > 0 ? (
          <div className="rounded-xl border border-border/40 divide-y divide-border/30">
            {tokens.data.tokens.map((t) => (
              <div key={t.id} className="flex items-center justify-between px-4 py-3">
                <div className="flex items-center gap-3">
                  <div className="grid h-8 w-8 place-items-center rounded-lg bg-surface-2/60 text-fg-subtle">
                    <Key size={14} />
                  </div>
                  <div>
                    <div className="text-sm font-medium text-fg">{t.name}</div>
                    <div className="text-xs text-fg-subtle">
                      {t.last_used_at
                        ? `Last used ${new Date(t.last_used_at).toLocaleDateString()}`
                        : `Created ${new Date(t.created_at).toLocaleDateString()}`}
                    </div>
                  </div>
                </div>
                <button onClick={() => del.mutateAsync(t.id).then(() => toast.success("Token deleted"))} className="grid h-8 w-8 place-items-center rounded-lg text-fg-subtle transition hover:bg-danger/10 hover:text-danger">
                  <Trash2 size={14} />
                </button>
              </div>
            ))}
          </div>
        ) : (
          tokens.isLoading ? null :
          <p className="text-center text-sm text-fg-subtle py-4">No tokens yet</p>
        )}
      </div>
    </Section>
  );
}

// ─── Backup Section ───────────────────────────────────────────────────────────
function BackupSection({ usersOnly, allowRestore }: { usersOnly?: boolean; allowRestore?: boolean }) {
  const toast = useToast();
  const confirm = useConfirm();
  const exportBackup = useExportBackup();
  const exportUsers = useExportUserBackup();
  const restoreBackup = useRestoreBackup();
  const fileRef = useRef<HTMLInputElement>(null);

  async function handleRestore() {
    const file = fileRef.current?.files?.[0];
    if (!file) return;
    const ok = await confirm({
      title: "Restore backup?",
      message: "This REPLACES the entire configuration. The current config will be lost.",
      confirmLabel: "Restore",
      destructive: true,
    });
    if (!ok) return;
    try {
      const res = await restoreBackup.mutateAsync(file);
      const r = res.restored;
      toast.success(`Restored: ${r.nodes ?? 0} nodes, ${r.users ?? 0} users`);
    } catch (e: any) { toast.error(e.message || "Restore failed"); }
    if (fileRef.current) fileRef.current.value = "";
  }

  return (
    <Section
      icon={<Cpu size={16} />}
      title={usersOnly ? "My users backup" : "Backup / Restore"}
      description={usersOnly
        ? "Download a JSON snapshot of users you created (your accounts only)."
        : "Export or replace the entire proxy configuration — nodes, inbounds, outbounds, routing, users and bindings."}
    >
      <div className="flex flex-wrap gap-3">
        <Button
          variant="outline"
          onClick={() => (usersOnly ? exportUsers : exportBackup).mutate()}
          disabled={usersOnly ? exportUsers.isPending : exportBackup.isPending}
        >
          <Download size={15} /> {usersOnly ? "Export my users" : "Export backup"}
        </Button>
        {allowRestore && (
        <label className="cursor-pointer">
          <input ref={fileRef} type="file" accept=".json" className="hidden" onChange={handleRestore} />
          <span className="inline-flex items-center gap-2 rounded-xl border border-border-strong/80 bg-surface/60 px-4 py-2 text-sm font-medium text-fg transition hover:bg-surface-2/80">
            <Upload size={15} className="text-fg-muted" /> Import & restore
          </span>
        </label>
        )}
      </div>
    </Section>
  );
}

// ─── Config Template Section ──────────────────────────────────────────────────
function ConfigTemplateSection() {
  const toast = useToast();
  const [clashRules, setClashRules] = useState("DOMAIN-SUFFIX,ir,DIRECT\nGEOIP,IR,DIRECT");
  const [singboxDNS, setSingboxDNS] = useState('{"servers":[{"address":"https://dns.google/dns-query","tag":"google"}]}');

  function saveTemplates() {
    // TODO: persist to backend settings API
    localStorage.setItem("vortex_clash_rules", clashRules);
    localStorage.setItem("vortex_singbox_dns", singboxDNS);
    toast.success("Templates saved (local)");
  }

  return (
    <Section icon={<Cpu size={16} />} title="Subscription Templates" description="Customize Clash/sing-box configs delivered to users. Add routing rules, DNS, or proxy groups.">
      <div className="space-y-4">
        <div>
          <p className="mb-1.5 text-xs font-medium text-fg-muted">Clash — extra routing rules (one per line)</p>
          <textarea
            value={clashRules}
            onChange={(e) => setClashRules(e.target.value)}
            rows={4}
            className="w-full rounded-xl border border-border bg-surface-2/30 px-3 py-2 font-mono text-xs text-fg placeholder:text-fg-subtle/50 outline-none focus:ring-1 focus:ring-primary/30"
            placeholder="DOMAIN-SUFFIX,ir,DIRECT&#10;GEOIP,IR,DIRECT"
          />
        </div>
        <div>
          <p className="mb-1.5 text-xs font-medium text-fg-muted">Sing-box — DNS config (JSON)</p>
          <textarea
            value={singboxDNS}
            onChange={(e) => setSingboxDNS(e.target.value)}
            rows={4}
            className="w-full rounded-xl border border-border bg-surface-2/30 px-3 py-2 font-mono text-xs text-fg placeholder:text-fg-subtle/50 outline-none focus:ring-1 focus:ring-primary/30"
            placeholder='{"servers":[{"address":"https://dns.google/dns-query","tag":"dns"}]}'
          />
        </div>
        <Button onClick={saveTemplates}>Save templates</Button>
      </div>
    </Section>
  );
}

// ─── Subscription Auto-Update Section ─────────────────────────────────────────
function SubUpdateSection() {
  const { t } = useI18n();
  const toast = useToast();
  const qc = useQueryClient();
  const [hours, setHours] = useState("");

  const cfg = useQuery({
    queryKey: ["sub-settings"],
    queryFn: () => api<{ config: { update_interval: number } }>("/api/sub-settings"),
  });

  // Sync local state with the fetched value once it arrives.
  const fetched = cfg.data?.config.update_interval;
  if (fetched !== undefined && hours === "") {
    setHours(String(fetched));
  }

  const save = useMutation({
    mutationFn: (interval: number) =>
      api("/api/sub-settings", { method: "PUT", body: { update_interval: interval } }),
    onSuccess: () => {
      toast.success(t("common.save"));
      qc.invalidateQueries({ queryKey: ["sub-settings"] });
    },
    onError: () => toast.error("Save failed"),
  });

  return (
    <Section icon={<RefreshCw size={16} />} title={t("settings.subUpdate")} description={t("settings.subUpdateDesc")}>
      <div className="space-y-3">
        <div>
          <p className="mb-1 text-xs font-medium text-fg-muted">{t("settings.subUpdateHours")}</p>
          <Input
            value={hours}
            onChange={(e) => setHours(e.target.value)}
            inputMode="numeric"
            className="w-32"
            placeholder="12"
          />
        </div>
        <Button
          onClick={() => save.mutate(Number(hours) || 12)}
          disabled={save.isPending || cfg.isLoading}
        >
          {t("common.save")}
        </Button>
      </div>
    </Section>
  );
}

// ─── IP Guard Section ─────────────────────────────────────────────────────────
function IPGuardSection() {
  const toast = useToast();
  const [whitelist, setWhitelist] = useState("");
  const [blacklist, setBlacklist] = useState("");

  function save() {
    localStorage.setItem("vortex_ip_whitelist", whitelist);
    localStorage.setItem("vortex_ip_blacklist", blacklist);
    toast.success("IP rules saved (restart panel to apply)");
  }

  return (
    <Section icon={<ShieldCheck size={16} />} title="IP Access Control" description="Restrict panel access by IP. Comma-separated IPs or CIDRs.">
      <div className="space-y-3">
        <div>
          <p className="mb-1 text-xs font-medium text-fg-muted">Whitelist (only these IPs can access)</p>
          <Input placeholder="e.g. 203.0.113.0/24, 198.51.100.5" value={whitelist} onChange={(e) => setWhitelist(e.target.value)} />
        </div>
        <div>
          <p className="mb-1 text-xs font-medium text-fg-muted">Blacklist (block these IPs)</p>
          <Input placeholder="e.g. 10.0.0.0/8" value={blacklist} onChange={(e) => setBlacklist(e.target.value)} />
        </div>
        <p className="text-[10px] text-fg-subtle">Set via env: VORTEX_IP_WHITELIST / VORTEX_IP_BLACKLIST (comma-separated CIDRs)</p>
        <Button onClick={save}>Save</Button>
      </div>
    </Section>
  );
}

// ─── Custom Branding Section ──────────────────────────────────────────────────
function BrandingSection() {
  const toast = useToast();
  const [panelName, setPanelName] = useState("VortexUI");
  const [accentColor, setAccentColor] = useState("#6366f1");
  const [footerText, setFooterText] = useState("© 2026 iPmart Network. All rights reserved.");
  const [logoURL, setLogoURL] = useState("");

  function save() {
    localStorage.setItem("vortex_branding", JSON.stringify({ panelName, accentColor, footerText, logoURL }));
    toast.success("Branding saved (refresh to apply)");
  }

  return (
    <Section icon={<Palette size={16} />} title="Custom Branding" description="Personalize the panel with your brand name, colors, and logo.">
      <div className="space-y-3">
        <div className="grid grid-cols-2 gap-3">
          <div>
            <p className="mb-1 text-xs font-medium text-fg-muted">Panel Name</p>
            <Input value={panelName} onChange={(e) => setPanelName(e.target.value)} />
          </div>
          <div>
            <p className="mb-1 text-xs font-medium text-fg-muted">Accent Color</p>
            <div className="flex gap-2">
              <Input value={accentColor} onChange={(e) => setAccentColor(e.target.value)} className="flex-1" />
              <input type="color" value={accentColor} onChange={(e) => setAccentColor(e.target.value)} className="h-9 w-9 rounded-lg border border-border cursor-pointer" />
            </div>
          </div>
        </div>
        <div>
          <p className="mb-1 text-xs font-medium text-fg-muted">Logo URL (leave empty for default)</p>
          <Input placeholder="https://example.com/logo.svg" value={logoURL} onChange={(e) => setLogoURL(e.target.value)} />
        </div>
        <div>
          <p className="mb-1 text-xs font-medium text-fg-muted">Footer Text</p>
          <Input value={footerText} onChange={(e) => setFooterText(e.target.value)} />
        </div>
        <Button onClick={save}>Save Branding</Button>
      </div>
    </Section>
  );
}

// ─── Auto Backup Section ──────────────────────────────────────────────────────
function AutoBackupSection() {
  const toast = useToast();
  const [tgToken, setTgToken] = useState("");
  const [tgChat, setTgChat] = useState("");
  const [s3Endpoint, setS3Endpoint] = useState("");
  const [s3Bucket, setS3Bucket] = useState("");
  const [interval, setInterval] = useState("24");

  function save() {
    localStorage.setItem("vortex_autobackup", JSON.stringify({ tgToken, tgChat, s3Endpoint, s3Bucket, interval }));
    toast.success("Auto-backup settings saved");
  }

  return (
    <Section icon={<Upload size={16} />} title="Automatic Backup" description="Scheduled daily backup to Telegram chat or S3-compatible storage.">
      <div className="space-y-3">
        <p className="text-xs font-medium text-fg-muted">Telegram destination</p>
        <div className="grid grid-cols-2 gap-2">
          <Input placeholder="Bot token" value={tgToken} onChange={(e) => setTgToken(e.target.value)} />
          <Input placeholder="Chat ID" value={tgChat} onChange={(e) => setTgChat(e.target.value)} />
        </div>
        <p className="text-xs font-medium text-fg-muted">S3 destination (optional)</p>
        <div className="grid grid-cols-2 gap-2">
          <Input placeholder="S3 endpoint URL" value={s3Endpoint} onChange={(e) => setS3Endpoint(e.target.value)} />
          <Input placeholder="Bucket name" value={s3Bucket} onChange={(e) => setS3Bucket(e.target.value)} />
        </div>
        <div>
          <p className="mb-1 text-xs font-medium text-fg-muted">Interval (hours)</p>
          <Input value={interval} onChange={(e) => setInterval(e.target.value)} inputMode="numeric" className="w-24" />
        </div>
        <p className="text-[10px] text-fg-subtle">Configure via env: VORTEX_BACKUP_TG_TOKEN, VORTEX_BACKUP_TG_CHAT, VORTEX_BACKUP_S3_*</p>
        <Button onClick={save}>Save</Button>
      </div>
    </Section>
  );
}

// ─── Update Checker Section ───────────────────────────────────────────────────
function UpdateSection() {
  const toast = useToast();
  const [checking, setChecking] = useState(false);
  const [result, setResult] = useState<{ available: boolean; current: string; latest: string } | null>(null);

  async function check() {
    setChecking(true);
    try {
      const res = await api<{ available: boolean; current: string; latest: string }>("/api/update/check");
      setResult(res);
      if (res.available) {
        toast.success(`Update available: ${res.latest}`);
      } else {
        toast.success("You're on the latest version!");
      }
    } catch { toast.error("Check failed"); }
    setChecking(false);
  }

  return (
    <Section icon={<Download size={16} />} title="System Update" description="Check for new VortexUI releases and update panel + core binaries.">
      <div className="space-y-3">
        <div className="flex items-center gap-4">
          <Button onClick={check} disabled={checking}>{checking ? "Checking…" : "Check for updates"}</Button>
          {result && (
            <span className="text-sm text-fg-muted">
              Current: <strong>{result.current}</strong>
              {result.available && <> → <strong className="text-success">{result.latest}</strong></>}
            </span>
          )}
        </div>
        {result?.available && (
          <div className="rounded-xl border border-success/30 bg-success/5 p-4">
            <p className="text-sm font-medium text-success">🎉 New version {result.latest} is available!</p>
            <p className="mt-1 text-xs text-fg-muted">Run on your server: <code className="rounded bg-surface-2/60 px-1.5 py-0.5">vortexui update</code></p>
          </div>
        )}
        <p className="text-[10px] text-fg-subtle">Updates include panel binary + xray-core + sing-box. The update command builds from source or pulls pre-built binaries.</p>
      </div>
    </Section>
  );
}
