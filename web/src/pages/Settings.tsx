import { useState, useRef } from "react";
import {
  Moon, Sun, Monitor, ShieldCheck, Download, Upload,
  Palette, Lock, Cpu, Key, Copy, Trash2, Check,
} from "lucide-react";
import { QRCodeSVG } from "qrcode.react";
import { useMutation } from "@tanstack/react-query";
import { api } from "@/api/client";
import { useConfirmTOTP, useDisableTOTP, useSetupTOTP } from "@/api/admin-hooks";
import { useExportBackup, useRestoreBackup, useAPITokens, useCreateAPIToken, useDeleteAPIToken } from "@/api/policy-hooks";
import { Button, Card, Input, PageHeader } from "@/components/ui";
import { useConfirm } from "@/components/confirm";
import { useToast } from "@/components/toast";
import { useTheme } from "@/theme/theme";
import { useI18n } from "@/i18n/i18n";
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

  return (
    <div className="mx-auto max-w-4xl space-y-5 animate-fade-in">
      <PageHeader title={t("nav.settings")} />

      {/* Appearance */}
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

      {/* Change Password */}
      <PasswordSection />

      {/* 2FA */}
      <TwoFASection />

      {/* API Tokens */}
      <APITokenSection />

      {/* Backup */}
      <BackupSection />
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
function BackupSection() {
  const toast = useToast();
  const confirm = useConfirm();
  const exportBackup = useExportBackup();
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
    <Section icon={<Cpu size={16} />} title={`Backup / Restore`} description="Export or replace the entire proxy configuration — nodes, inbounds, outbounds, routing, users and bindings.">
      <div className="flex flex-wrap gap-3">
        <Button variant="outline" onClick={() => exportBackup.mutate()} disabled={exportBackup.isPending}>
          <Download size={15} /> Export backup
        </Button>
        <label className="cursor-pointer">
          <input ref={fileRef} type="file" accept=".json" className="hidden" onChange={handleRestore} />
          <span className="inline-flex items-center gap-2 rounded-xl border border-border-strong/80 bg-surface/60 px-4 py-2 text-sm font-medium text-fg transition hover:bg-surface-2/80">
            <Upload size={15} className="text-fg-muted" /> Import & restore
          </span>
        </label>
      </div>
    </Section>
  );
}
