import { useState } from "react";
import { useNavigate } from "react-router-dom";
import { getApiErrorMessage } from "@/lib/form-errors";
import { motion } from "framer-motion";
import {
  Lock,
  User,
  Eye,
  EyeOff,
  ArrowRight,
  ShieldCheck,
  Key,
  Moon,
  Sun,
  Globe,
} from "lucide-react";
import { useAuth } from "@/auth/auth";
import { useTheme } from "@/theme/theme";
import { useI18n } from "@/i18n/i18n";
import { LANG_OPTIONS } from "@/i18n/lang-options";
import { cn } from "@/lib/utils";
export function Login() {
  const { login } = useAuth();
  const navigate = useNavigate();
  const { resolved, toggle } = useTheme();
  const { t, lang, setLang } = useI18n();
  const [authMode, setAuthMode] = useState<"admin" | "token">("admin");
  const [showPassword, setShowPassword] = useState(false);
  const [username, setUsername] = useState("");
  const [password, setPassword] = useState("");
  const [totp, setTotp] = useState("");
  const [error, setError] = useState("");
  const [busy, setBusy] = useState(false);
  const [langOpen, setLangOpen] = useState(false);

  async function submit(e: React.FormEvent) {
    e.preventDefault();
    if (authMode === "token") {
      navigate("/portal/login");
      return;
    }
    setError("");
    setBusy(true);
    try {
      await login(username, password, totp);
      navigate("/overview");
    } catch (err) {
      const status = (err as { status?: number })?.status;
      setError(
        status === 429
          ? t("login.tooMany")
          : getApiErrorMessage(err, t("login.invalid"), t),
      );
    } finally {
      setBusy(false);
    }
  }

  return (
    <div className="min-h-screen flex items-center justify-center p-4 relative overflow-hidden bg-bg text-fg">
      <div className="absolute inset-0 pointer-events-none overflow-hidden">
        <div className="absolute top-[-30%] start-[20%] w-[500px] h-[500px] rounded-full bg-primary/[0.07] blur-[120px]" />
        <div className="absolute bottom-[-20%] end-[15%] w-[400px] h-[400px] rounded-full bg-accent/[0.06] blur-[100px]" />
      </div>
      <div className="absolute inset-0 bg-grid-pattern opacity-30 pointer-events-none" />

      <div className="absolute end-5 top-5 flex gap-1.5 z-20">
        <button
          type="button"
          onClick={toggle}
          aria-label="theme"
          className="grid h-9 w-9 place-items-center rounded-xl text-fg-muted transition hover:bg-surface-2/60 hover:text-fg"
        >
          {resolved === "dark" ? <Sun size={17} /> : <Moon size={17} />}
        </button>
        <div className="relative">
          <button
            type="button"
            onClick={() => setLangOpen(!langOpen)}
            className="grid h-9 w-9 place-items-center rounded-xl text-fg-muted transition hover:bg-surface-2/60 hover:text-fg"
            aria-label="Language"
          >
            <Globe size={17} />
          </button>
          {langOpen && (
            <>
              <div className="fixed inset-0 z-40" onClick={() => setLangOpen(false)} />
              <div className="absolute end-0 top-full z-50 mt-2 w-36 rounded-xl border border-border/60 bg-bg-elevated/95 p-1 shadow-xl backdrop-blur-xl animate-scale-in">
                {LANG_OPTIONS.map((l) => (
                  <button
                    key={l.code}
                    type="button"
                    onClick={() => {
                      setLang(l.code);
                      setLangOpen(false);
                    }}
                    className={cn(
                      "flex w-full items-center gap-2 rounded-lg px-3 py-2 text-xs font-medium transition",
                      lang === l.code
                        ? "bg-primary/10 text-primary"
                        : "text-fg-muted hover:bg-surface-2/60 hover:text-fg",
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
      </div>

      <motion.div
        initial={{ opacity: 0, y: 16 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ duration: 0.5, ease: [0.16, 1, 0.3, 1] }}
        className="relative z-10 w-full max-w-[420px]"
      >
        <div className="text-center mb-10">
          <motion.h1
            initial={{ opacity: 0, scale: 0.95 }}
            animate={{ opacity: 1, scale: 1 }}
            transition={{ delay: 0.05, duration: 0.4 }}
            className="text-3xl font-black tracking-tight"
          >
            <span className="text-fg">Vortex</span>
            <span className="grad-text">UI</span>
          </motion.h1>
          <motion.p
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
            transition={{ delay: 0.15 }}
            className="text-[13px] text-fg-muted mt-2 tracking-wide"
          >
            {t("login.panelSubtitle")}
          </motion.p>
        </div>

        <div className="rounded-2xl bg-bg-elevated border border-border p-7 shadow-xl space-y-6">
          <div className="flex rounded-xl p-1 bg-surface-2/80 border border-border/60">
            <button
              type="button"
              onClick={() => setAuthMode("admin")}
              className={cn(
                "flex-1 py-2.5 rounded-lg text-[13px] font-semibold transition-all duration-200 flex items-center justify-center gap-2",
                authMode === "admin"
                  ? "bg-bg-elevated text-fg shadow-sm border border-border/80"
                  : "text-fg-muted hover:text-fg",
              )}
            >
              <Lock size={14} />
              {t("login.adminTab")}
            </button>
            <button
              type="button"
              onClick={() => setAuthMode("token")}
              className={cn(
                "flex-1 py-2.5 rounded-lg text-[13px] font-semibold transition-all duration-200 flex items-center justify-center gap-2",
                authMode === "token"
                  ? "bg-bg-elevated text-fg shadow-sm border border-border/80"
                  : "text-fg-muted hover:text-fg",
              )}
            >
              <Key size={14} />
              {t("login.tokenTab")}
            </button>
          </div>

          <form onSubmit={submit} className="space-y-5">
            {authMode === "admin" ? (
              <>
                <div className="space-y-2">
                  <label className="text-xs font-medium text-fg-muted">{t("login.username")}</label>
                  <div className="relative">
                    <User size={16} className="absolute start-3.5 top-1/2 -translate-y-1/2 text-fg-subtle" />
                    <input
                      type="text"
                      required
                      value={username}
                      onChange={(e) => setUsername(e.target.value)}
                      autoFocus
                      className="w-full h-11 rounded-xl border border-border ps-10 pe-4 text-sm text-fg placeholder:text-fg-subtle/60 focus:outline-none focus:border-primary focus:ring-2 focus:ring-primary/15 transition-all input-surface"
                      placeholder={t("login.username")}
                    />
                  </div>
                </div>

                <div className="space-y-2">
                  <label className="text-xs font-medium text-fg-muted">{t("login.password")}</label>
                  <div className="relative">
                    <Lock size={16} className="absolute start-3.5 top-1/2 -translate-y-1/2 text-fg-subtle" />
                    <input
                      type={showPassword ? "text" : "password"}
                      required
                      value={password}
                      onChange={(e) => setPassword(e.target.value)}
                      className="w-full h-11 rounded-xl border border-border ps-10 pe-11 text-sm text-fg placeholder:text-fg-subtle/60 focus:outline-none focus:border-primary focus:ring-2 focus:ring-primary/15 transition-all input-surface"
                      placeholder={t("login.password")}
                    />
                    <button
                      type="button"
                      onClick={() => setShowPassword(!showPassword)}
                      className="absolute end-3.5 top-1/2 -translate-y-1/2 text-fg-subtle hover:text-fg transition-colors"
                    >
                      {showPassword ? <EyeOff size={16} /> : <Eye size={16} />}
                    </button>
                  </div>
                </div>

                <div className="space-y-2">
                  <label className="text-xs font-medium text-fg-muted">{t("login.totp")}</label>
                  <div className="relative">
                    <ShieldCheck size={16} className="absolute start-3.5 top-1/2 -translate-y-1/2 text-fg-subtle" />
                    <input
                      type="text"
                      value={totp}
                      onChange={(e) => setTotp(e.target.value)}
                      inputMode="numeric"
                      className="w-full h-11 rounded-xl border border-border ps-10 pe-4 text-sm text-fg placeholder:text-fg-subtle/60 focus:outline-none focus:border-primary focus:ring-2 focus:ring-primary/15 transition-all input-surface"
                      placeholder={t("login.totp")}
                    />
                  </div>
                </div>
              </>
            ) : (
              <div className="space-y-2">
                <p className="text-[13px] text-fg-muted leading-relaxed">
                  {t("login.tokenHint")}
                </p>
              </div>
            )}

            {error && (
              <div className="rounded-xl border border-danger/20 bg-danger/10 px-3 py-2">
                <p className="text-sm font-medium text-danger">{error}</p>
              </div>
            )}

            <div className="flex items-center justify-between">
              <span className="text-[11px] text-success font-medium flex items-center gap-1">
                <ShieldCheck size={12} /> {t("login.secureConnection")}
              </span>
            </div>

            <button
              type="submit"
              disabled={busy}
              className="w-full h-11 rounded-xl font-semibold text-sm transition-all duration-200 active:scale-[0.98] disabled:opacity-50 disabled:pointer-events-none flex items-center justify-center gap-2 grad-bg text-primary-fg shadow-md hover:shadow-lg hover:brightness-110"
            >
              {busy ? (
                <>
                  <div className="h-4 w-4 border-2 border-primary-fg/30 border-t-primary-fg rounded-full animate-spin" />
                  <span>{t("login.signingIn")}</span>
                </>
              ) : (
                <>
                  <span>
                    {authMode === "admin" ? t("login.submit") : t("login.goPortal")}
                  </span>
                  <ArrowRight size={16} />
                </>
              )}
            </button>
          </form>
        </div>
      </motion.div>
    </div>
  );
}
