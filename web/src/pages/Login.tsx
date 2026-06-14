import { useState } from "react";
import { useNavigate } from "react-router-dom";
import { Network, Moon, Sun, Languages } from "lucide-react";
import { useAuth } from "@/auth/auth";
import { useTheme } from "@/theme/theme";
import { useI18n } from "@/i18n/i18n";
import { Button, Input } from "@/components/ui";

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
    } catch {
      setError(t("login.invalid"));
    } finally {
      setBusy(false);
    }
  }

  return (
    <div className="relative flex min-h-screen items-center justify-center overflow-hidden p-4">
      <div
        aria-hidden
        className="pointer-events-none absolute -top-40 start-1/2 h-[480px] w-[480px] -translate-x-1/2 rounded-full bg-primary/20 blur-[120px]"
      />
      <div className="absolute end-4 top-4 flex gap-1">
        <button onClick={toggle} aria-label="theme" className="grid h-9 w-9 place-items-center rounded-lg text-fg-muted hover:bg-surface-2 hover:text-fg">
          {resolved === "dark" ? <Sun size={18} /> : <Moon size={18} />}
        </button>
        <button onClick={() => setLang(lang === "en" ? "fa" : "en")} aria-label="language" className="grid h-9 w-9 place-items-center rounded-lg text-fg-muted hover:bg-surface-2 hover:text-fg">
          <Languages size={18} />
        </button>
      </div>

      <div className="card w-full max-w-sm animate-scale-in p-7">
        <div className="mb-6 flex flex-col items-center text-center">
          <div className="grad-bg mb-3 grid h-12 w-12 place-items-center rounded-2xl text-white shadow-lg shadow-primary/40">
            <Network size={24} strokeWidth={2.5} />
          </div>
          <h1 className="text-xl font-semibold tracking-tight grad-text">{t("login.title")}</h1>
          <p className="mt-1 text-sm text-fg-muted">{t("login.subtitle")}</p>
        </div>

        <form onSubmit={submit} className="space-y-3">
          <Input placeholder={t("login.username")} value={username} onChange={(e) => setUsername(e.target.value)} autoFocus />
          <Input type="password" placeholder={t("login.password")} value={password} onChange={(e) => setPassword(e.target.value)} />
          <Input placeholder={t("login.totp")} value={totp} onChange={(e) => setTotp(e.target.value)} inputMode="numeric" />
          {error && <p className="text-sm text-danger">{error}</p>}
          <Button type="submit" className="w-full" disabled={busy}>
            {busy ? t("login.signingIn") : t("login.submit")}
          </Button>
        </form>
      </div>
    </div>
  );
}
