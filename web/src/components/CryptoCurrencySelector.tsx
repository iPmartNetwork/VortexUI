import { Check, Copy } from "lucide-react";
import { cn } from "@/lib/utils";
import {
  CRYPTO_COINS,
  accentClass,
  configuredCryptoCoins,
  type CryptoCoin,
} from "@/lib/crypto-currencies";
import { useI18n } from "@/i18n/i18n";
import { useState } from "react";
import { Input } from "@/components/ui";
import { useToast } from "@/components/toast";

function CoinChip({
  coin,
  active,
  configured,
  onClick,
}: {
  coin: CryptoCoin;
  active: boolean;
  configured?: boolean;
  onClick: () => void;
}) {
  return (
    <button
      type="button"
      onClick={onClick}
      className={cn(
        "flex min-w-[5.5rem] flex-col items-center rounded-xl border px-3 py-2.5 text-center transition",
        accentClass(coin.accent, active),
      )}
    >
      <span className="text-sm font-bold tracking-tight">{coin.symbol}</span>
      <span className="text-[10px] uppercase text-muted-foreground">{coin.network}</span>
      {configured && (
        <span className="mt-1 flex items-center gap-0.5 text-[10px] text-emerald-500">
          <Check size={10} /> OK
        </span>
      )}
    </button>
  );
}

/** Admin: pick a coin and edit its wallet address. */
export function CryptoAddressEditor({
  addresses,
  onChange,
}: {
  addresses: Record<string, string>;
  onChange: (addresses: Record<string, string>) => void;
}) {
  const { t } = useI18n();
  const [selected, setSelected] = useState(CRYPTO_COINS[0].id);
  const coin = CRYPTO_COINS.find((c) => c.id === selected) ?? CRYPTO_COINS[0];

  return (
    <div className="space-y-3 sm:col-span-2">
      <div className="text-sm font-medium">{t("billing.cryptoWallets")}</div>
      <div className="flex flex-wrap gap-2">
        {CRYPTO_COINS.map((c) => (
          <CoinChip
            key={c.id}
            coin={c}
            active={selected === c.id}
            configured={(addresses[c.id] ?? "").trim() !== ""}
            onClick={() => setSelected(c.id)}
          />
        ))}
      </div>
      <Input
        placeholder={`${coin.symbol} (${coin.network}) ${t("billing.walletAddress")}`}
        value={addresses[coin.id] ?? ""}
        onChange={(e) => onChange({ ...addresses, [coin.id]: e.target.value })}
        className="font-mono text-xs"
      />
    </div>
  );
}

/** Reseller: pick configured coin and copy deposit address. */
export function CryptoPaySelector({
  addresses,
  selected,
  onSelect,
}: {
  addresses: Record<string, string>;
  selected: string;
  onSelect: (coinId: string) => void;
}) {
  const { t } = useI18n();
  const toast = useToast();
  const available = configuredCryptoCoins(addresses);

  if (available.length === 0) {
    return <p className="text-sm text-muted-foreground">{t("billing.noCryptoConfigured")}</p>;
  }

  const coin = getCryptoCoinSafe(selected, available);
  const address = addresses[coin.id] ?? "";

  async function copyAddress() {
    try {
      await navigator.clipboard.writeText(address);
      toast.success(t("billing.addressCopied"));
    } catch {
      toast.error(t("billing.copyFailed"));
    }
  }

  return (
    <div className="space-y-3">
      <div className="text-xs font-medium text-muted-foreground">{t("billing.selectCrypto")}</div>
      <div className="flex flex-wrap gap-2">
        {available.map((c) => (
          <CoinChip
            key={c.id}
            coin={c}
            active={coin.id === c.id}
            onClick={() => onSelect(c.id)}
          />
        ))}
      </div>
      <div className={cn("rounded-xl border p-3", accentClass(coin.accent, true))}>
        <div className="mb-1 flex items-center justify-between gap-2">
          <span className="text-xs font-semibold">{coin.symbol} · {coin.network}</span>
          <button
            type="button"
            onClick={copyAddress}
            className="flex items-center gap-1 rounded-lg px-2 py-1 text-xs text-muted-foreground hover:bg-muted/60"
          >
            <Copy size={12} /> {t("cleanip.copy")}
          </button>
        </div>
        <code className="block break-all text-xs">{address}</code>
      </div>
    </div>
  );
}

function getCryptoCoinSafe(selected: string, available: CryptoCoin[]): CryptoCoin {
  return available.find((c) => c.id === selected) ?? available[0];
}

/** Package pricing currency chips. */
export function PackageCurrencyPicker({
  value,
  onChange,
}: {
  value: string;
  onChange: (currency: string) => void;
}) {
  const fiats = ["IRR", "USD"] as const;
  return (
    <div className="space-y-2">
      <div className="flex flex-wrap gap-1.5">
        {fiats.map((c) => (
          <button
            key={c}
            type="button"
            onClick={() => onChange(c)}
            className={cn(
              "rounded-lg border px-3 py-1.5 text-xs font-medium transition",
              value === c ? "border-primary bg-primary/15 ring-1 ring-primary/40" : "border-border hover:border-primary/40",
            )}
          >
            {c}
          </button>
        ))}
      </div>
      <div className="flex flex-wrap gap-1.5">
        {CRYPTO_COINS.map((c) => (
          <button
            key={c.id}
            type="button"
            onClick={() => onChange(c.id)}
            className={cn(
              "rounded-lg border px-2.5 py-1.5 text-xs transition",
              value === c.id ? accentClass(c.accent, true) : accentClass(c.accent, false),
            )}
          >
            <span className="font-semibold">{c.symbol}</span>
            <span className="ms-1 text-[10px] text-muted-foreground">{c.network}</span>
          </button>
        ))}
      </div>
    </div>
  );
}
