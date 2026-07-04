import { useState } from "react";
import { useNavigate } from "react-router-dom";
import { motion } from "framer-motion";
import { Key, ArrowRight, ShieldCheck } from "lucide-react";
import { Button, Input } from "@/components/ui";
import { useToast } from "@/components/toast";
import { useI18n } from "@/i18n/i18n";

const PORTAL_TOKEN_KEY = "vortex.portal.token";

export function setPortalToken(t: string) {
  localStorage.setItem(PORTAL_TOKEN_KEY, t);
}
export function getPortalToken() {
  return localStorage.getItem(PORTAL_TOKEN_KEY);
}
export function clearPortalToken() {
  localStorage.removeItem(PORTAL_TOKEN_KEY);
}

export function PortalLogin() {
  const [token, setToken] = useState("");
  const [loading, setLoading] = useState(false);
  const navigate = useNavigate();
  const toast = useToast();
  const { t } = useI18n();

  async function submit(e: React.FormEvent) {
    e.preventDefault();
    setLoading(true);
    try {
      const res = await fetch("/api/portal/login", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ token }),
      });
      if (!res.ok) {
        const err = await res.json().catch(() => ({}));
        throw new Error(err.message || t("login.invalid"));
      }
      const data = await res.json();
      setPortalToken(data.token);
      navigate("/portal/dashboard");
    } catch (err: unknown) {
      const message = err instanceof Error ? err.message : t("login.invalid");
      toast.error(message);
    } finally {
      setLoading(false);
    }
  }

  return (
    <div className="min-h-screen flex items-center justify-center p-4 relative overflow-hidden bg-bg text-fg">
      <div className="absolute inset-0 pointer-events-none overflow-hidden">
        <div className="absolute top-[-30%] start-[20%] w-[500px] h-[500px] rounded-full bg-primary/[0.07] blur-[120px]" />
        <div className="absolute bottom-[-20%] end-[15%] w-[400px] h-[400px] rounded-full bg-accent/[0.06] blur-[100px]" />
      </div>
      <div className="absolute inset-0 bg-grid-pattern opacity-30 pointer-events-none" />

      <motion.div
        initial={{ opacity: 0, y: 16 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ duration: 0.5, ease: [0.16, 1, 0.3, 1] }}
        className="relative z-10 w-full max-w-[420px]"
      >
        <div className="text-center mb-10">
          <h1 className="text-3xl font-black tracking-tight">
            <span className="grad-text">{t("portal.brand")}</span>
          </h1>
          <p className="text-[13px] text-fg-muted mt-2">{t("portal.loginSubtitle")}</p>
        </div>

        <div className="rounded-2xl bg-bg-elevated border border-border p-7 shadow-xl">
          <form onSubmit={submit} className="space-y-5">
            <div className="space-y-2">
              <label className="text-xs font-medium text-fg-muted">{t("portal.tokenLabel")}</label>
              <div className="relative">
                <Key size={16} className="absolute start-3.5 top-1/2 -translate-y-1/2 text-fg-subtle" />
                <Input
                  placeholder={t("portal.tokenPlaceholder")}
                  value={token}
                  onChange={(e) => setToken(e.target.value)}
                  required
                  className="ps-10 font-mono text-sm"
                />
              </div>
            </div>
            <span className="text-[11px] text-success font-medium flex items-center gap-1">
              <ShieldCheck size={12} /> {t("portal.secureConnection")}
            </span>
            <Button type="submit" className="w-full h-11" disabled={loading || !token}>
              {loading ? (
                t("portal.loggingIn")
              ) : (
                <span className="flex items-center justify-center gap-2">
                  {t("login.submit")} <ArrowRight size={16} />
                </span>
              )}
            </Button>
          </form>
        </div>
      </motion.div>
    </div>
  );
}
