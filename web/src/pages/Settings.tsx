import { useState } from "react";
import { QRCodeSVG } from "qrcode.react";
import { Moon, Sun, Monitor, Languages, KeyRound, ShieldCheck } from "lucide-react";
import { useMutation } from "@tanstack/react-query";
import { api } from "@/api/client";
import { useConfirmTOTP, useDisableTOTP, useSetupTOTP } from "@/api/admin-hooks";
import { Button, Card, Input, PageHeader } from "@/components/ui";
import { useToast } from "@/components/toast";
import { useTheme } from "@/theme/theme";
import { useI18n } from "@/i18n/i18n";
import { cn } from "@/lib/utils";

function SegBtn({ active, onClick, children }: { active: boolean; onClick: () => void; children: React.ReactNode }) {
  return (
    <button
      onClick={onClick}
      className={cn(
        "flex flex-1 items-center justify-center gap-1.5 rounded-lg px-3 py-2 text-sm font-medium transition",
        active ? "grad-bg text-white shadow shadow-primary/25" : "text-fg-muted hover:bg-white/[0.06]",
      )}
    >
      {children}
    </button>
  );
}

export function Settings() {
  const { t } = useI18n();
  const { theme, setTheme } = useTheme();
  const { lang, setLang } = useI18n();
  const toast = useToast();

  const setup = useSetupTOTP();
  const confirmTotp = useConfirmTOTP();
  const disableTotp = useDisableTOTP();
  const [url, setUrl] = useState("");
  const [code, setCode] = useState("");
  const [enabled, setEnabled] = useState(false);

  const changePw = useMutation({
    mutationFn: (b: { current: string; new: string }) => api("/api/account/password", { method: "POST", body: b }),
  });
  const [cur, setCur] = useState("");
  const [nw, setNw] = useState("");

  async function submitPw(e: React.FormEvent) {
    e.preventDefault();
    try {
      await changePw.mutateAsync({ current: cur, new: nw });
      toast.success("Password changed");
      setCur("");
      setNw("");
    } catch {
      toast.error("Wrong current password");
    }
  }

  return (
    <div className="max-w-2xl">
      <PageHeader title={t("nav.settings")} />

      <div className="space-y-5">
        <Card>
          <h2 className="mb-4 text-sm font-semibold">Appearance</h2>
          <div className="grid gap-4 sm:grid-cols-2">
            <div>
              <p className="mb-2 text-xs text-fg-muted">{t("theme.label")}</p>
              <div className="flex gap-1 rounded-xl border border-white/[0.06] bg-white/[0.02] p-1">
                <SegBtn active={theme === "light"} onClick={() => setTheme("light")}><Sun size={15} />{t("theme.light")}</SegBtn>
                <SegBtn active={theme === "dark"} onClick={() => setTheme("dark")}><Moon size={15} />{t("theme.dark")}</SegBtn>
                <SegBtn active={theme === "system"} onClick={() => setTheme("system")}><Monitor size={15} />{t("theme.system")}</SegBtn>
              </div>
            </div>
            <div>
              <p className="mb-2 text-xs text-fg-muted">Language</p>
              <div className="flex gap-1 rounded-xl border border-white/[0.06] bg-white/[0.02] p-1">
                <SegBtn active={lang === "en"} onClick={() => setLang("en")}><Languages size={15} />English</SegBtn>
                <SegBtn active={lang === "fa"} onClick={() => setLang("fa")}><Languages size={15} />فارسی</SegBtn>
              </div>
            </div>
          </div>
        </Card>

        <Card>
          <h2 className="mb-1 flex items-center gap-2 text-sm font-semibold"><KeyRound size={15} /> Change password</h2>
          <form onSubmit={submitPw} className="mt-3 grid gap-3 sm:grid-cols-2">
            <Input type="password" placeholder="Current password" value={cur} onChange={(e) => setCur(e.target.value)} required />
            <Input type="password" placeholder="New password" value={nw} onChange={(e) => setNw(e.target.value)} required />
            <div className="sm:col-span-2">
              <Button type="submit" disabled={changePw.isPending}>{t("common.save")}</Button>
            </div>
          </form>
        </Card>

        <Card>
          <h2 className="mb-1 flex items-center gap-2 text-sm font-semibold"><ShieldCheck size={15} /> Two-factor authentication</h2>
          <p className="text-sm text-fg-muted">Protect your account with a time-based code.</p>
          <div className="mt-3">
            {enabled ? (
              <form
                onSubmit={async (e) => {
                  e.preventDefault();
                  try { await disableTotp.mutateAsync(code); setEnabled(false); setCode(""); toast.success("2FA disabled"); }
                  catch { toast.error("Invalid code"); }
                }}
                className="flex gap-2"
              >
                <Input placeholder="Current code" value={code} onChange={(e) => setCode(e.target.value)} inputMode="numeric" />
                <Button type="submit" variant="destructive">Disable</Button>
              </form>
            ) : url ? (
              <form
                onSubmit={async (e) => {
                  e.preventDefault();
                  try { await confirmTotp.mutateAsync(code); setEnabled(true); setUrl(""); setCode(""); toast.success("2FA enabled"); }
                  catch { toast.error("Invalid code"); }
                }}
                className="space-y-3"
              >
                <div className="flex justify-center rounded-xl bg-white p-4"><QRCodeSVG value={url} size={150} /></div>
                <div className="flex gap-2">
                  <Input placeholder="6-digit code" value={code} onChange={(e) => setCode(e.target.value)} inputMode="numeric" autoFocus />
                  <Button type="submit">{t("common.confirm")}</Button>
                </div>
              </form>
            ) : (
              <Button onClick={async () => setUrl((await setup.mutateAsync()).url)} disabled={setup.isPending}>
                Enable 2FA
              </Button>
            )}
          </div>
        </Card>
      </div>
    </div>
  );
}
