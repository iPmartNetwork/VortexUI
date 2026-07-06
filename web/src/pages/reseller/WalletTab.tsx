import { useEffect, useMemo, useState } from "react";
import { useAccountWallet } from "@/api/reseller-hooks";
import {
  useBillingSettings,
  useInitWalletDeposit,
  usePaymentInfo,
  useSaveBillingSettings,
  useWalletPackages,
  type WalletPackage,
} from "@/api/wallet-billing-hooks";
import { Button, Input, Select } from "@/components/ui";
import { GlassCard } from "@/components/veltrix";
import { CardToCardInfo } from "@/components/CardToCardInfo";
import { CryptoAddressEditor } from "@/components/CryptoCurrencySelector";
import { WalletRechargeSection } from "@/components/WalletRechargeSection";
import { useToast } from "@/components/toast";
import { useI18n } from "@/i18n/i18n";
import { formatBytes } from "@/lib/utils";
import { cn } from "@/lib/utils";
import { mergeCryptoAddresses } from "@/lib/crypto-currencies";

const PRESET_AMOUNTS = [500_000, 1_000_000, 5_000_000];

function ledgerAmount(entry: { delta_users: number; delta_traffic: number; reason: string }) {
  const m = entry.reason.match(/([\d,]+)\s*Toman/i);
  if (m) return { text: `+${m[1]} T`, positive: true };
  if (entry.delta_users > 0 || entry.delta_traffic > 0) return { text: `+${entry.delta_users || "quota"}`, positive: true };
  return { text: entry.reason.slice(0, 40), positive: false };
}

export function WalletTab() {
  const { t } = useI18n();
  const toast = useToast();
  const wallet = useAccountWallet();
  const packages = useWalletPackages(true);
  const paymentInfo = usePaymentInfo();
  const initDeposit = useInitWalletDeposit();
  const settings = useBillingSettings();
  const saveSettings = useSaveBillingSettings();

  const [amount, setAmount] = useState(1_000_000);
  const [method, setMethod] = useState("zarinpal");
  const [showSettings, setShowSettings] = useState(false);
  const [form, setForm] = useState({
    card_number: "",
    card_holder: "",
    card_bank: "",
    manual_instructions: "",
    crypto_addresses: mergeCryptoAddresses(),
  });

  useEffect(() => {
    if (!settings.data?.settings) return;
    const s = settings.data.settings;
    setForm({
      card_number: s.card_number,
      card_holder: s.card_holder,
      card_bank: s.card_bank,
      manual_instructions: s.manual_instructions,
      crypto_addresses: mergeCryptoAddresses(s.crypto_addresses),
    });
  }, [settings.data]);

  const ledger = wallet.data?.ledger ?? [];
  const matchedPkg = useMemo((): WalletPackage | undefined => {
    const pkgs = packages.data?.packages ?? [];
    return pkgs.find((p) => p.currency === "IRR" && Math.abs(p.price_amount - amount) < amount * 0.01)
      ?? pkgs.find((p) => p.currency === "IRR" && p.price_amount >= amount)
      ?? pkgs[0];
  }, [packages.data, amount]);

  async function proceedPayment() {
    if (!matchedPkg) {
      toast.error(t("reseller.noPackageForAmount"));
      return;
    }
    try {
      const res = await initDeposit.mutateAsync({ package_id: matchedPkg.id, method });
      if (res.redirect_url) {
        window.location.href = res.redirect_url;
        return;
      }
      toast.success(t("billing.depositSubmitted"));
    } catch (e) {
      toast.error(e instanceof Error ? e.message : t("billing.depositFailed"));
    }
  }

  async function onSaveSettings() {
    try {
      await saveSettings.mutateAsync(form);
      toast.success(t("billing.settingsSaved"));
    } catch {
      toast.error(t("billing.settingsFailed"));
    }
  }

  const zarinEnabled = paymentInfo.data?.zarinpal_enabled;

  return (
    <div className="space-y-4">
      <div className="grid grid-cols-1 gap-4 lg:grid-cols-2">
        <GlassCard hover={false} className="!p-5 space-y-4">
          <div>
            <h2 className="text-sm font-bold text-fg">{t("reseller.topUpTitle")}</h2>
            <p className="text-xs text-fg-muted mt-0.5">{t("reseller.topUpDesc")}</p>
          </div>
          <div>
            <p className="text-xs text-fg-subtle mb-2">{t("reseller.depositAmount")}</p>
            <div className="grid grid-cols-3 gap-2">
              {PRESET_AMOUNTS.map((a) => (
                <button
                  key={a}
                  type="button"
                  onClick={() => setAmount(a)}
                  className={cn(
                    "rounded-xl border px-3 py-2.5 text-xs font-bold transition",
                    amount === a ? "border-primary bg-primary/10 text-primary" : "border-border/60 hover:border-primary/40",
                  )}
                >
                  {a.toLocaleString()} T
                </button>
              ))}
            </div>
          </div>
          {matchedPkg && (
            <p className="text-[11px] text-fg-muted">
              {matchedPkg.name} · {formatBytes(matchedPkg.traffic_bytes, false)} · {matchedPkg.user_credits} users
            </p>
          )}
          <div>
            <p className="text-xs text-fg-subtle mb-1">{t("reseller.paymentMethod")}</p>
            <Select value={method} onChange={(e) => setMethod(e.target.value)}>
              {zarinEnabled && <option value="zarinpal">{t("reseller.zarinpal")}</option>}
              <option value="card_to_card">{t("billing.cardToCard")}</option>
              <option value="crypto">Crypto</option>
            </Select>
          </div>
          {method === "card_to_card" && paymentInfo.data?.settings && (
            <CardToCardInfo settings={paymentInfo.data.settings} />
          )}
          <Button className="w-full" onClick={proceedPayment} disabled={initDeposit.isPending || !matchedPkg}>
            {t("reseller.proceedPayment")}
          </Button>
        </GlassCard>

        <GlassCard hover={false} className="!p-5 space-y-3">
          <h2 className="text-sm font-bold text-fg">{t("reseller.ledgerTitle")}</h2>
          <div className="space-y-2 max-h-[280px] overflow-y-auto">
            {ledger.slice(0, 12).map((e) => {
              const { text, positive } = ledgerAmount(e);
              return (
                <div key={e.id} className="flex items-start justify-between gap-2 text-xs border-b border-border/30 pb-2">
                  <div className="min-w-0">
                    <p className="text-fg truncate">{e.reason}</p>
                    <p className="text-fg-subtle text-[10px]">{new Date(e.created_at).toLocaleString()}</p>
                  </div>
                  <span className={cn("font-bold tabular-nums flex-shrink-0", positive ? "text-success" : "text-danger")}>
                    {text}
                  </span>
                </div>
              );
            })}
            {ledger.length === 0 && <p className="text-sm text-fg-muted py-4 text-center">{t("reseller.ledgerEmpty")}</p>}
          </div>
        </GlassCard>
      </div>

      <GlassCard hover={false} className="!p-5">
        <button
          type="button"
          className="text-sm font-bold text-fg mb-3"
          onClick={() => setShowSettings((v) => !v)}
        >
          {showSettings ? "▼" : "▶"} {t("billing.tab.settings")}
        </button>
        {showSettings && (
          <div className="grid gap-3 sm:grid-cols-2 pt-2">
            <Input placeholder={t("billing.cardNumber")} value={form.card_number} onChange={(e) => setForm({ ...form, card_number: e.target.value })} />
            <Input placeholder={t("billing.cardHolder")} value={form.card_holder} onChange={(e) => setForm({ ...form, card_holder: e.target.value })} />
            <Input placeholder={t("billing.cardBank")} value={form.card_bank} onChange={(e) => setForm({ ...form, card_bank: e.target.value })} />
            <CryptoAddressEditor addresses={form.crypto_addresses} onChange={(crypto_addresses) => setForm({ ...form, crypto_addresses })} />
            <Input className="sm:col-span-2" placeholder={t("billing.manualInstructions")} value={form.manual_instructions} onChange={(e) => setForm({ ...form, manual_instructions: e.target.value })} />
            <Button onClick={onSaveSettings} disabled={saveSettings.isPending}>{t("common.save")}</Button>
          </div>
        )}
      </GlassCard>

      <GlassCard hover={false} className="!p-5">
        <h2 className="text-sm font-bold text-fg mb-3">{t("reseller.packagesSection")}</h2>
        <WalletRechargeSection />
      </GlassCard>
    </div>
  );
}
