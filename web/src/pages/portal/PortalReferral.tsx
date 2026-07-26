import { useState } from "react";
import { useQuery, useMutation } from "@tanstack/react-query";
import { motion } from "framer-motion";
import { Gift, Copy, Users, Check, Sparkles } from "lucide-react";
import { portalApi } from "./portalApi";
import { GlassCard } from "@/components/vortexui";
import { Button, Input } from "@/components/ui";
import { useToast } from "@/components/toast";
import { useI18n } from "@/i18n/i18n";

export function PortalReferral() {
  const { t } = useI18n();
  const toast = useToast();
  const [code, setCode] = useState("");
  const [copied, setCopied] = useState(false);

  const { data, isLoading } = useQuery({
    queryKey: ["portal-referral-code"],
    queryFn: () => portalApi<{ code: { code: string; uses: number; max_uses: number } }>("/api/portal/referral/code"),
  });

  const apply = useMutation({
    mutationFn: (referralCode: string) =>
      portalApi("/api/portal/referral/apply", { method: "POST", body: { code: referralCode } }),
  });

  const referralCode = data?.code?.code;
  const uses = data?.code?.uses ?? 0;
  const maxUses = data?.code?.max_uses ?? 0;

  return (
    <div className="space-y-6 animate-page-enter">
      {/* ── Header ── */}
      <div className="relative overflow-hidden rounded-2xl border border-border/60 bg-gradient-to-br from-bg-elevated via-surface to-primary/[0.03] p-5 md:p-6">
        <div className="absolute top-0 end-0 w-56 h-56 rounded-full bg-primary/8 blur-3xl pointer-events-none" />
        <div className="absolute bottom-0 start-0 w-40 h-40 rounded-full bg-accent/8 blur-3xl pointer-events-none" />
        
        <div className="relative z-10">
          <h1 className="text-xl md:text-2xl font-black text-fg tracking-tight flex items-center gap-2">
            <Gift size={22} className="text-primary" />
            {t("portal.referral.title")}
          </h1>
          <p className="text-[13px] text-fg-muted mt-1.5 max-w-lg">
            Share your referral code and earn rewards
          </p>
        </div>
      </div>

      {/* ── Your Referral Code ── */}
      <motion.div
        initial={{ opacity: 0, y: 10 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ delay: 0.05 }}
      >
        <GlassCard glow className="!p-5 relative overflow-hidden">
          <div className="absolute top-0 end-0 w-32 h-32 rounded-full bg-primary/5 blur-3xl pointer-events-none" />
          
          <div className="relative z-10 space-y-4">
            <div className="flex items-center gap-2">
              <div className="h-8 w-8 rounded-lg bg-primary/15 flex items-center justify-center">
                <Users size={16} className="text-primary" />
              </div>
              <div>
                <p className="text-xs font-bold text-fg">{t("portal.referral.yourCode")}</p>
                <p className="text-[10px] text-fg-muted">Share with friends to earn bonuses</p>
              </div>
            </div>

            {isLoading ? (
              <div className="h-12 animate-pulse rounded-xl bg-surface-2/60" />
            ) : referralCode ? (
              <div className="flex flex-col sm:flex-row items-stretch sm:items-center gap-3">
                <div className="flex-1 relative">
                  <div className="h-12 rounded-xl bg-surface-2/60 border border-border/50 flex items-center px-4 font-mono text-sm font-bold text-fg tracking-wider select-all">
                    <Sparkles size={14} className="text-primary mr-2 shrink-0" />
                    {referralCode}
                  </div>
                </div>
                <Button
                  type="button"
                  variant={copied ? "success" : "glass"}
                  onClick={() => {
                    void navigator.clipboard.writeText(referralCode);
                    setCopied(true);
                    toast.success(t("portal.referral.copied"));
                    setTimeout(() => setCopied(false), 2000);
                  }}
                  className="gap-1.5"
                >
                  {copied ? (
                    <><Check size={15} /> Copied</>
                  ) : (
                    <><Copy size={15} /> Copy</>
                  )}
                </Button>
              </div>
            ) : (
              <p className="text-sm text-fg-muted">No referral code available</p>
            )}

            {/* Stats */}
            {data?.code && (
              <div className="flex items-center gap-4 pt-1">
                <div className="flex items-center gap-1.5 text-xs">
                  <Users size={13} className="text-fg-subtle" />
                  <span className="text-fg-muted">Used by</span>
                  <strong className="text-fg tabular-nums">{uses}</strong>
                  {maxUses > 0 && (
                    <span className="text-fg-subtle">/ {maxUses}</span>
                  )}
                </div>
                {maxUses > 0 && (
                  <div className="flex-1 max-w-[120px]">
                    <div className="h-1.5 rounded-full bg-surface-3/60 overflow-hidden">
                      <div
                        className="h-full rounded-full bg-gradient-to-r from-primary to-accent transition-all duration-700"
                        style={{ width: `${Math.min(100, (uses / maxUses) * 100)}%` }}
                      />
                    </div>
                  </div>
                )}
              </div>
            )}
          </div>
        </GlassCard>
      </motion.div>

      {/* ── Apply Referral Code ── */}
      <motion.div
        initial={{ opacity: 0, y: 10 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ delay: 0.1 }}
      >
        <GlassCard glow className="!p-5 relative overflow-hidden">
          <div className="absolute top-0 start-0 w-32 h-32 rounded-full bg-accent/5 blur-3xl pointer-events-none" />
          
          <div className="relative z-10 space-y-4">
            <div className="flex items-center gap-2">
              <div className="h-8 w-8 rounded-lg bg-accent/15 flex items-center justify-center">
                <Gift size={16} className="text-accent" />
              </div>
              <div>
                <p className="text-xs font-bold text-fg">{t("portal.referral.applyLabel")}</p>
                <p className="text-[10px] text-fg-muted">Have a referral code? Apply it here</p>
              </div>
            </div>

            <div className="flex flex-col sm:flex-row gap-2">
              <Input
                value={code}
                onChange={(e) => setCode(e.target.value)}
                placeholder={t("portal.referral.applyPlaceholder")}
                className="flex-1 font-mono text-sm"
              />
              <Button
                type="button"
                disabled={!code.trim() || apply.isPending}
                onClick={async () => {
                  try {
                    await apply.mutateAsync(code.trim());
                    toast.success(t("common.save"));
                    setCode("");
                  } catch (e: unknown) {
                    toast.error(e instanceof Error ? e.message : "Failed");
                  }
                }}
                className="gap-1.5"
              >
                {apply.isPending ? (
                  <><div className="h-3 w-3 animate-spin rounded-full border-2 border-white border-t-transparent" /> Applying</>
                ) : (
                  <><Gift size={15} /> {t("portal.referral.apply")}</>
                )}
              </Button>
            </div>
          </div>
        </GlassCard>
      </motion.div>
    </div>
  );
}
