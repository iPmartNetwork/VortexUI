import { useMemo } from "react";
import { Coins, Plus, Users, Wallet } from "lucide-react";
import { useAccountQuota } from "@/api/quota-hooks";
import { useAccountWallet } from "@/api/reseller-hooks";
import { useBillingSettings } from "@/api/wallet-billing-hooks";
import { Button } from "@/components/ui";
import { GlassCard } from "@/components/veltrix";
import { useI18n } from "@/i18n/i18n";
import { formatBytes } from "@/lib/utils";
import { configuredCryptoCoins } from "@/lib/crypto-currencies";

export function SummaryWidgets({ onDeposit }: { onDeposit: () => void }) {
  const { t } = useI18n();
  const wallet = useAccountWallet();
  const quota = useAccountQuota();
  const settings = useBillingSettings();

  const w = wallet.data?.wallet;
  const ledger = wallet.data?.ledger ?? [];

  const monthCredit = useMemo(() => {
    const since = Date.now() - 30 * 24 * 60 * 60 * 1000;
    let users = 0;
    for (const e of ledger) {
      if (new Date(e.created_at).getTime() >= since && e.delta_users > 0) users += e.delta_users;
    }
    return users;
  }, [ledger]);

  const cryptoCoins = configuredCryptoCoins(settings.data?.settings?.crypto_addresses);
  const cryptoLabel = cryptoCoins.length > 0
    ? cryptoCoins.map((c) => c.network || c.id.toUpperCase()).slice(0, 3).join(" · ")
    : t("reseller.widgetCryptoEmpty");

  const userCount = quota.data?.usage?.user_count ?? 0;
  const userMax = quota.data?.usage?.user_quota ?? w?.user_credits ?? 0;
  const poolPct = userMax > 0 ? Math.min(100, Math.round((userCount / userMax) * 100)) : 0;

  return (
    <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 xl:grid-cols-4">
      <GlassCard hover={false} className="!p-4 space-y-2">
        <div className="flex items-center justify-between">
          <span className="text-xs font-semibold text-fg-subtle uppercase tracking-wide">{t("reseller.widgetWallet")}</span>
          <Wallet size={16} className="text-primary" />
        </div>
        <p className="text-lg font-bold text-fg tabular-nums">
          {w ? `${w.user_credits.toLocaleString()} ${t("reseller.widgetCredits")}` : "—"}
        </p>
        <p className="text-[11px] text-fg-muted">
          {w ? formatBytes(w.traffic_bytes, false) : "—"}
          {monthCredit > 0 && (
            <span className="text-success ms-1">+{monthCredit.toLocaleString()} {t("reseller.widgetThisMonth")}</span>
          )}
        </p>
      </GlassCard>

      <GlassCard hover={false} className="!p-4 space-y-2">
        <div className="flex items-center justify-between">
          <span className="text-xs font-semibold text-fg-subtle uppercase tracking-wide">{t("reseller.widgetCrypto")}</span>
          <Coins size={16} className="text-warning" />
        </div>
        <p className="text-lg font-bold text-fg">{cryptoCoins.length > 0 ? t("reseller.widgetCryptoOn") : "—"}</p>
        <p className="text-[11px] text-fg-muted font-mono">{cryptoLabel}</p>
      </GlassCard>

      <GlassCard hover={false} className="!p-4 space-y-2">
        <div className="flex items-center justify-between">
          <span className="text-xs font-semibold text-fg-subtle uppercase tracking-wide">{t("reseller.widgetUsers")}</span>
          <Users size={16} className="text-accent" />
        </div>
        <p className="text-lg font-bold text-fg tabular-nums">
          {userCount.toLocaleString()}
          {userMax > 0 && <span className="text-sm font-normal text-fg-muted"> / {userMax.toLocaleString()} max</span>}
        </p>
        {userMax > 0 && (
          <div className="h-1.5 rounded-full bg-surface-3 overflow-hidden">
            <div className="h-full rounded-full bg-gradient-to-r from-accent to-primary" style={{ width: `${poolPct}%` }} />
          </div>
        )}
      </GlassCard>

      <GlassCard hover={false} className="!p-4 flex flex-col justify-between gap-3">
        <div>
          <p className="text-xs font-semibold text-fg-subtle uppercase tracking-wide">{t("reseller.widgetQuick")}</p>
          <p className="text-sm text-fg-muted mt-1">{t("reseller.widgetQuickDesc")}</p>
        </div>
        <Button size="sm" className="w-full" onClick={onDeposit}>
          <Plus size={14} /> {t("reseller.depositWallet")}
        </Button>
      </GlassCard>
    </div>
  );
}
