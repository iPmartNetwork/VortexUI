import { useEffect, useState } from "react";
import {
  useAccountDeposits,
  useInitWalletDeposit,
  usePaymentInfo,
  useWalletPackages,
  formatPrice,
  type WalletPackage,
} from "@/api/wallet-billing-hooks";
import { Badge, Button, Card, Input } from "@/components/ui";
import { CryptoPaySelector } from "@/components/CryptoCurrencySelector";
import { CardToCardInfo } from "@/components/CardToCardInfo";
import { configuredCryptoCoins } from "@/lib/crypto-currencies";
import { useToast } from "@/components/toast";
import { useI18n } from "@/i18n/i18n";
import { formatBytes } from "@/lib/utils";

export function WalletRechargeSection() {
  const { t } = useI18n();
  const toast = useToast();
  const packages = useWalletPackages(true);
  const paymentInfo = usePaymentInfo();
  const deposits = useAccountDeposits();
  const initDeposit = useInitWalletDeposit();
  const [selected, setSelected] = useState<WalletPackage | null>(null);
  const [method, setMethod] = useState("");
  const [cryptoCoin, setCryptoCoin] = useState("");
  const [txId, setTxId] = useState("");
  const [note, setNote] = useState("");
  const [proof, setProof] = useState("");

  const settings = paymentInfo.data?.settings;
  const zarin = paymentInfo.data?.zarinpal_enabled;
  const cryptoGw = paymentInfo.data?.crypto_enabled;

  useEffect(() => {
    if (method !== "crypto" || !settings?.crypto_addresses) return;
    const available = configuredCryptoCoins(settings.crypto_addresses);
    if (available.length > 0 && !available.some((c) => c.id === cryptoCoin)) {
      setCryptoCoin(available[0].id);
    }
  }, [method, settings, cryptoCoin]);

  async function submit() {
    if (!selected || !method) return;
    if (method === "crypto" && !cryptoCoin) {
      toast.error(t("billing.selectCrypto"));
      return;
    }
    try {
      const res = await initDeposit.mutateAsync({
        package_id: selected.id,
        method,
        crypto_coin: cryptoCoin || undefined,
        tx_id: txId || undefined,
        proof_image: proof || undefined,
        reseller_note: note || undefined,
      });
      if (res.redirect_url) {
        window.location.href = res.redirect_url;
        return;
      }
      toast.success(t("billing.depositSubmitted"));
      setSelected(null);
      setMethod("");
      setCryptoCoin("");
      setTxId("");
      setNote("");
      setProof("");
    } catch (e) {
      toast.error(e instanceof Error ? e.message : t("billing.depositFailed"));
    }
  }

  const availableMethods = (pkg: WalletPackage) =>
    pkg.methods.filter((m) => {
      if (m === "zarinpal") return zarin;
      if (m === "nowpayments") return cryptoGw;
      if (m === "crypto") return configuredCryptoCoins(settings?.crypto_addresses).length > 0;
      return true;
    });

  return (
    <section className="space-y-4">
      <h3 className="text-sm font-semibold">{t("billing.rechargeTitle")}</h3>
      <div className="grid gap-3 sm:grid-cols-2 lg:grid-cols-3">
        {(packages.data?.packages ?? []).map((p) => (
          <Card
            key={p.id}
            className={`cursor-pointer space-y-2 p-4 transition ${selected?.id === p.id ? "ring-2 ring-primary" : ""}`}
            onClick={() => {
              setSelected(p);
              const m = availableMethods(p)[0] ?? "";
              setMethod(m);
              setCryptoCoin("");
            }}
          >
            <div className="font-semibold">{p.name}</div>
            <div className="text-sm text-muted-foreground">{formatBytes(p.traffic_bytes, false)} · {p.user_credits} {t("reseller.account.users")}</div>
            <div className="font-medium">{formatPrice(p.price_amount, p.currency)}</div>
          </Card>
        ))}
        {(packages.data?.packages ?? []).length === 0 && (
          <p className="text-sm text-muted-foreground sm:col-span-2">{t("billing.noPackages")}</p>
        )}
      </div>

      {selected && (
        <Card className="space-y-4 p-4">
          <div className="text-sm font-medium">{t("billing.selectMethod")}: {selected.name}</div>
          <div className="flex flex-wrap gap-2">
            {availableMethods(selected).map((m) => (
              <Button
                key={m}
                size="sm"
                variant={method === m ? "primary" : "ghost"}
                onClick={() => { setMethod(m); setCryptoCoin(""); }}
              >
                {t(`billing.method.${m}` as never)}
              </Button>
            ))}
          </div>

          {method === "card_to_card" && settings && (
            <CardToCardInfo settings={settings} />
          )}

          {method === "crypto" && settings && (
            <CryptoPaySelector
              addresses={settings.crypto_addresses ?? {}}
              selected={cryptoCoin}
              onSelect={setCryptoCoin}
            />
          )}

          {(method === "card_to_card" || method === "crypto") && (
            <>
              <Input placeholder={t("billing.txId")} value={txId} onChange={(e) => setTxId(e.target.value)} />
              <div>
                <label className="mb-1 block text-xs text-muted-foreground">{t("billing.proofImage")}</label>
                <input
                  type="file"
                  accept="image/png,image/jpeg,image/webp"
                  className="block w-full text-sm"
                  onChange={(e) => {
                    const file = e.target.files?.[0];
                    if (!file || file.size > 512_000) return;
                    const reader = new FileReader();
                    reader.onload = () => setProof(String(reader.result ?? ""));
                    reader.readAsDataURL(file);
                  }}
                />
              </div>
              <Input placeholder={t("billing.resellerNote")} value={note} onChange={(e) => setNote(e.target.value)} />
            </>
          )}

          <Button onClick={submit} disabled={initDeposit.isPending || !method || (method === "crypto" && !cryptoCoin)}>
            {method === "zarinpal" || method === "nowpayments" ? t("billing.payOnline") : t("billing.submitDeposit")}
          </Button>
        </Card>
      )}

      {(deposits.data?.deposits ?? []).length > 0 && (
        <Card className="p-0 overflow-x-auto">
          <table className="w-full text-sm">
            <thead className="border-b text-left text-muted-foreground">
              <tr>
                <th className="px-4 py-2">{t("billing.colWhen")}</th>
                <th className="px-4 py-2">{t("billing.colPackage")}</th>
                <th className="px-4 py-2">{t("billing.colMethod")}</th>
                <th className="px-4 py-2">{t("billing.colCrypto")}</th>
                <th className="px-4 py-2">{t("billing.colStatus")}</th>
              </tr>
            </thead>
            <tbody>
              {deposits.data!.deposits.map((d) => (
                <tr key={d.id} className="border-b last:border-0">
                  <td className="px-4 py-2 text-muted-foreground">{new Date(d.created_at).toLocaleString()}</td>
                  <td className="px-4 py-2">{d.package_name ?? "—"}</td>
                  <td className="px-4 py-2">{d.method}</td>
                  <td className="px-4 py-2">{d.crypto_coin ? <Badge>{d.crypto_coin}</Badge> : "—"}</td>
                  <td className="px-4 py-2"><Badge>{d.status}</Badge></td>
                </tr>
              ))}
            </tbody>
          </table>
        </Card>
      )}
    </section>
  );
}
