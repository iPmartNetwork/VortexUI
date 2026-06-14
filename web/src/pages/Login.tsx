import { useState } from "react";
import { useNavigate } from "react-router-dom";
import { Moon, Sun, Globe } from "lucide-react";
import { useAuth } from "@/auth/auth";
import { useTheme } from "@/theme/theme";
import { useI18n } from "@/i18n/i18n";
import type { Lang } from "@/i18n/dict";
import { Input } from "@/components/ui";
import { cn } from "@/lib/utils";

export function Login() {
  const { login } = useAuth();
  const navigate = useNavigate();
  const { resolved, toggle } = useTheme();
  const { t, lang, setLang } = useI18n();
  const [username, setUsername] = useState("");
  const [password, setPassword] = useState("");
  const [totp, setTotp] = useState("");
  const [error, setError] = useState("");
  const [busy, setBusy] = useState(false);

  async function submit(e: React.FormEvent) {
    e.preventDefault();
    setError("");
    setBusy(true);
    try {
      await login(username, password, totp);
      navigate("/overview");
    } catch (err) {
      const status = (err as { status?: number })?.status;
      setError(status === 429 ? t("login.tooMany") : t("login.invalid"));
    } finally {
      setBusy(false);
    }
  }

  return (
    <div className="relative flex min-h-screen items-center justify-center overflow-hidden p-4">
      {/* Background glow orbs */}
      <div aria-hidden className="pointer-events-none absolute -top-32 start-1/4 h-[500px] w-[500px] rounded-full bg-primary/12 blur-[160px]" />
      <div aria-hidden className="pointer-events-none absolute -bottom-40 end-1/4 h-[400px] w-[400px] rounded-full bg-accent/8 blur-[140px]" />

      <div className="absolute end-5 top-5 flex gap-1.5">
        <button onClick={toggle} aria-label="theme" className="grid h-9 w-9 place-items-center rounded-xl text-fg-muted transition hover:bg-surface-2/60 hover:text-fg">
          {resolved === "dark" ? <Sun size={17} /> : <Moon size={17} />}
        </button>
        <LoginLangMenu lang={lang as Lang} setLang={setLang} />
      </div>

      <div className="card w-full max-w-[380px] animate-scale-in p-8">
        <div className="mb-8 pt-2 text-center">
          <h1 className="text-[2rem] font-black tracking-wider text-primary" style={{ fontFamily: "'Orbitron', sans-serif" }}>VortexUI</h1>
        </div>

        <form onSubmit={submit} className="space-y-3.5">
          <Input placeholder={t("login.username")} value={username} onChange={(e) => setUsername(e.target.value)} autoFocus />
          <Input type="password" placeholder={t("login.password")} value={password} onChange={(e) => setPassword(e.target.value)} />
          <Input placeholder={t("login.totp")} value={totp} onChange={(e) => setTotp(e.target.value)} inputMode="numeric" />
          {error && <p className="text-sm font-medium text-danger">{error}</p>}
          <button
            type="submit"
            disabled={busy}
            className="w-full h-11 rounded-xl text-sm font-semibold text-white transition-all active:scale-[0.97] disabled:opacity-50 shadow-lg"
            style={{
              background: "linear-gradient(to right, hsl(212 100% 55%), hsl(190 90% 48%))",
              boxShadow: "0 4px 20px -4px hsl(200 90% 50% / 0.4)",
            }}
          >
            {busy ? t("login.signingIn") : t("login.submit")}
          </button>
        </form>
      </div>
    </div>
  );
}

const LANG_OPTIONS: { code: Lang; label: string }[] = [
  { code: "en", label: "English" },
  { code: "fa", label: "فارسی" },
  { code: "tr", label: "Türkçe" },
  { code: "ar", label: "العربية" },
  { code: "ru", label: "Русский" },
  { code: "zh", label: "中文" },
  { code: "ja", label: "日本語" },
  { code: "es", label: "Español" },
];

function LoginLangMenu({ lang, setLang }: { lang: Lang; setLang: (l: Lang) => void }) {
  const [open, setOpen] = useState(false);
  return (
    <div className="relative">
      <button
        onClick={() => setOpen(!open)}
        className="grid h-9 w-9 place-items-center rounded-xl text-fg-muted transition hover:bg-surface-2/60 hover:text-fg"
        aria-label="Language"
      >
        <Globe size={17} />
      </button>
      {open && (
        <>
          <div className="fixed inset-0 z-40" onClick={() => setOpen(false)} />
          <div className="absolute end-0 top-full z-50 mt-2 w-36 rounded-xl border border-border/60 bg-bg-elevated/95 p-1 shadow-xl backdrop-blur-xl animate-scale-in">
            {LANG_OPTIONS.map((l) => (
              <button
                key={l.code}
                onClick={() => { setLang(l.code); setOpen(false); }}
                className={cn(
                  "flex w-full items-center gap-2 rounded-lg px-3 py-2 text-xs font-medium transition",
                  lang === l.code ? "bg-primary/10 text-primary" : "text-fg-muted hover:bg-surface-2/60 hover:text-fg",
                )}
              >
                <span className="w-5 text-[10px] font-bold text-fg-subtle">{l.code.toUpperCase()}</span>
                {l.label}
              </button>
            ))}
          </div>
        </>
      )}
    </div>
  );
}
