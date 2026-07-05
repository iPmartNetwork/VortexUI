import { useState } from "react";
import { useQuery, useMutation } from "@tanstack/react-query";
import { Gift, Copy } from "lucide-react";
import { portalApi } from "./portalApi";
import { GlassCard } from "@/components/veltrix";
import { Button, Input } from "@/components/ui";
import { useToast } from "@/components/toast";
import { useI18n } from "@/i18n/i18n";

export function PortalReferral() {
  const { t } = useI18n();
  const toast = useToast();
  const [code, setCode] = useState("");

  const { data, isLoading } = useQuery({
    queryKey: ["portal-referral-code"],
    queryFn: () => portalApi<{ code: { code: string; uses: number; max_uses: number } }>("/api/portal/referral/code"),
  });

  const apply = useMutation({
    mutationFn: (referralCode: string) =>
      portalApi("/api/portal/referral/apply", { method: "POST", body: { code: referralCode } }),
  });

  return (
    <div className="space-y-6 animate-page-enter">
      <h1 className="text-xl font-bold text-fg flex items-center gap-2">
        <Gift size={20} className="text-primary" />
        {t("portal.referral.title")}
      </h1>

      <GlassCard className="space-y-4">
        <p className="text-sm text-fg-muted">{t("portal.referral.yourCode")}</p>
        {isLoading ? (
          <p className="text-sm text-fg-muted">{t("common.loading")}</p>
        ) : (
          <div className="flex flex-wrap items-center gap-2">
            <code className="rounded-lg bg-surface-2 px-3 py-2 text-sm font-mono">{data?.code?.code ?? "—"}</code>
            {data?.code?.code && (
              <Button
                type="button"
                size="sm"
                variant="outline"
                onClick={() => {
                  void navigator.clipboard.writeText(data.code.code);
                  toast.success(t("portal.referral.copied"));
                }}
              >
                <Copy size={14} />
              </Button>
            )}
            {data?.code && (
              <span className="text-xs text-fg-subtle">
                {data.code.uses}/{data.code.max_uses || "∞"}
              </span>
            )}
          </div>
        )}
      </GlassCard>

      <GlassCard className="space-y-3">
        <p className="text-sm font-medium text-fg">{t("portal.referral.applyLabel")}</p>
        <div className="flex flex-col sm:flex-row gap-2">
          <Input
            value={code}
            onChange={(e) => setCode(e.target.value)}
            placeholder={t("portal.referral.applyPlaceholder")}
            className="flex-1"
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
          >
            {t("portal.referral.apply")}
          </Button>
        </div>
      </GlassCard>
    </div>
  );
}
