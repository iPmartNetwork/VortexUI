import { useEffect, useState } from "react";
import { useResellerPaymentConfig, useSaveResellerPaymentConfig } from "@/api/reseller-payment-hooks";
import { Button, Card, Input, PageHeader } from "@/components/ui";
import { useToast } from "@/components/toast";
import { useI18n } from "@/i18n/i18n";
import { Plus, Trash2 } from "lucide-react";

const METHODS = ["zarinpal", "card_to_card", "crypto"] as const;

export function ResellerPaymentSettings() {
  const { t } = useI18n();
  const toast = useToast();
  const config = useResellerPaymentConfig();
  const save = useSaveResellerPaymentConfig();

  const [cardNumber, setCardNumber] = useState("");
  const [cardHolder, setCardHolder] = useState("");
  const [cardBank, setCardBank] = useState("");
  const [zarinpalMerchantId, setZarinpalMerchantId] = useState("");
  const [manualInstructions, setManualInstructions] = useState("");
  const [enabledMethods, setEnabledMethods] = useState<string[]>([]);
  const [cryptoPairs, setCryptoPairs] = useState<{ coin: string; address: string }[]>([]);

  useEffect(() => {
    if (!config.data?.config) return;
    const c = config.data.config;
    setCardNumber(c.card_number || "");
    setCardHolder(c.card_holder || "");
    setCardBank(c.card_bank || "");
    setZarinpalMerchantId(c.zarinpal_merchant_id || "");
    setManualInstructions(c.manual_instructions || "");
    setEnabledMethods(c.enabled_methods || []);
    const pairs = Object.entries(c.crypto_addresses || {}).map(([coin, address]) => ({ coin, address }));
    setCryptoPairs(pairs.length > 0 ? pairs : [{ coin: "", address: "" }]);
  }, [config.data]);

  function toggleMethod(m: string) {
    setEnabledMethods((prev) =>
      prev.includes(m) ? prev.filter((x) => x !== m) : [...prev, m],
    );
  }

  function addCryptoPair() {
    setCryptoPairs((prev) => [...prev, { coin: "", address: "" }]);
  }

  function removeCryptoPair(idx: number) {
    setCryptoPairs((prev) => prev.filter((_, i) => i !== idx));
  }

  function updateCryptoPair(idx: number, field: "coin" | "address", value: string) {
    setCryptoPairs((prev) => prev.map((p, i) => (i === idx ? { ...p, [field]: value } : p)));
  }

  async function handleSave(e: React.FormEvent) {
    e.preventDefault();
    const crypto_addresses: Record<string, string> = {};
    for (const p of cryptoPairs) {
      if (p.coin.trim() && p.address.trim()) {
        crypto_addresses[p.coin.trim()] = p.address.trim();
      }
    }
    try {
      await save.mutateAsync({
        card_number: cardNumber,
        card_holder: cardHolder,
        card_bank: cardBank,
        crypto_addresses,
        zarinpal_merchant_id: zarinpalMerchantId,
        manual_instructions: manualInstructions,
        enabled_methods: enabledMethods,
      });
      toast.success(t("resellerPayment.saved"));
    } catch {
      toast.error(t("resellerPayment.saveFailed"));
    }
  }

  if (config.isLoading) {
    return <div className="p-8 text-center text-fg-muted">{t("common.loading")}</div>;
  }

  return (
    <div className="space-y-6 animate-page-enter">
      <PageHeader title={t("resellerPayment.title")} subtitle={t("resellerPayment.subtitle")} />

      <Card>
        <p className="mb-6 text-sm text-fg-muted">
          {t("resellerPayment.hint")}
        </p>

        <form onSubmit={handleSave} className="space-y-6">
          {/* Card info */}
          <div className="grid gap-4 sm:grid-cols-3">
            <div>
              <label className="mb-1.5 block text-xs font-medium text-fg-muted">{t("resellerPayment.cardNumber")}</label>
              <Input value={cardNumber} onChange={(e) => setCardNumber(e.target.value)} placeholder="6219-xxxx-xxxx-xxxx" />
            </div>
            <div>
              <label className="mb-1.5 block text-xs font-medium text-fg-muted">{t("resellerPayment.cardHolder")}</label>
              <Input value={cardHolder} onChange={(e) => setCardHolder(e.target.value)} placeholder="Ali Ahmadi" />
            </div>
            <div>
              <label className="mb-1.5 block text-xs font-medium text-fg-muted">{t("resellerPayment.cardBank")}</label>
              <Input value={cardBank} onChange={(e) => setCardBank(e.target.value)} placeholder="Melli" />
            </div>
          </div>

          {/* Crypto addresses */}
          <div>
            <label className="mb-1.5 block text-xs font-medium text-fg-muted">{t("resellerPayment.cryptoAddresses")}</label>
            <div className="space-y-2">
              {cryptoPairs.map((pair, idx) => (
                <div key={idx} className="flex items-center gap-2">
                  <Input
                    className="w-28"
                    value={pair.coin}
                    onChange={(e) => updateCryptoPair(idx, "coin", e.target.value)}
                    placeholder="BTC"
                  />
                  <Input
                    className="flex-1"
                    value={pair.address}
                    onChange={(e) => updateCryptoPair(idx, "address", e.target.value)}
                    placeholder="Wallet address"
                  />
                  <button
                    type="button"
                    onClick={() => removeCryptoPair(idx)}
                    className="grid h-8 w-8 place-items-center rounded-lg text-fg-muted hover:bg-danger/10 hover:text-danger"
                  >
                    <Trash2 size={14} />
                  </button>
                </div>
              ))}
              <Button type="button" variant="ghost" size="sm" onClick={addCryptoPair}>
                <Plus size={14} /> {t("resellerPayment.addCrypto")}
              </Button>
            </div>
          </div>

          {/* ZarinPal */}
          <div>
            <label className="mb-1.5 block text-xs font-medium text-fg-muted">{t("resellerPayment.zarinpalMerchant")}</label>
            <Input value={zarinpalMerchantId} onChange={(e) => setZarinpalMerchantId(e.target.value)} placeholder="xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx" />
          </div>

          {/* Manual instructions */}
          <div>
            <label className="mb-1.5 block text-xs font-medium text-fg-muted">{t("resellerPayment.manualInstructions")}</label>
            <textarea
              className="field min-h-[80px] w-full resize-y"
              value={manualInstructions}
              onChange={(e) => setManualInstructions(e.target.value)}
              placeholder="Instructions for users paying manually..."
            />
          </div>

          {/* Enabled methods */}
          <div>
            <label className="mb-2 block text-xs font-medium text-fg-muted">{t("resellerPayment.enabledMethods")}</label>
            <div className="flex flex-wrap gap-4">
              {METHODS.map((m) => (
                <label key={m} className="flex items-center gap-2 text-sm text-fg">
                  <input
                    type="checkbox"
                    checked={enabledMethods.includes(m)}
                    onChange={() => toggleMethod(m)}
                    className="rounded border-border accent-primary"
                  />
                  {m === "zarinpal" ? "ZarinPal" : m === "card_to_card" ? "Card to Card" : "Crypto"}
                </label>
              ))}
            </div>
          </div>

          <Button type="submit" disabled={save.isPending}>
            {save.isPending ? t("common.loading") : t("common.save")}
          </Button>
        </form>
      </Card>
    </div>
  );
}
