export interface CryptoCoin {
  id: string;
  symbol: string;
  network: string;
  /** Tailwind color token for chip accent */
  accent: string;
}

/** Supported manual crypto wallets for reseller top-up. */
export const CRYPTO_COINS: CryptoCoin[] = [
  { id: "USDT-TRC20", symbol: "USDT", network: "TRC20", accent: "emerald" },
  { id: "USDT-BEP20", symbol: "USDT", network: "BEP20", accent: "amber" },
  { id: "TRX", symbol: "TRX", network: "Tron", accent: "rose" },
  { id: "TON", symbol: "TON", network: "TON", accent: "sky" },
  { id: "BNB", symbol: "BNB", network: "BEP20", accent: "amber" },
  { id: "LTC", symbol: "LTC", network: "Litecoin", accent: "slate" },
  { id: "BTC", symbol: "BTC", network: "Bitcoin", accent: "orange" },
];

export const CRYPTO_COIN_IDS = CRYPTO_COINS.map((c) => c.id);

/** Package price currencies (fiat + crypto). */
export const WALLET_PACKAGE_CURRENCIES = ["IRR", "USD", ...CRYPTO_COIN_IDS] as const;

export function getCryptoCoin(id: string): CryptoCoin | undefined {
  return CRYPTO_COINS.find((c) => c.id === id);
}

export function configuredCryptoCoins(addresses: Record<string, string> | undefined) {
  return CRYPTO_COINS.filter((c) => (addresses?.[c.id] ?? "").trim() !== "");
}

export function emptyCryptoAddresses(): Record<string, string> {
  return Object.fromEntries(CRYPTO_COINS.map((c) => [c.id, ""]));
}

export function mergeCryptoAddresses(stored?: Record<string, string> | null): Record<string, string> {
  const base = emptyCryptoAddresses();
  if (!stored) return base;
  for (const c of CRYPTO_COINS) {
    if (stored[c.id]) base[c.id] = stored[c.id];
  }
  return base;
}

export function accentClass(accent: string, active: boolean) {
  const map: Record<string, { idle: string; active: string }> = {
    emerald: { idle: "border-emerald-500/30 hover:border-emerald-500/60", active: "border-emerald-500 bg-emerald-500/15 ring-1 ring-emerald-500/40" },
    amber: { idle: "border-amber-500/30 hover:border-amber-500/60", active: "border-amber-500 bg-amber-500/15 ring-1 ring-amber-500/40" },
    rose: { idle: "border-rose-500/30 hover:border-rose-500/60", active: "border-rose-500 bg-rose-500/15 ring-1 ring-rose-500/40" },
    sky: { idle: "border-sky-500/30 hover:border-sky-500/60", active: "border-sky-500 bg-sky-500/15 ring-1 ring-sky-500/40" },
    slate: { idle: "border-slate-400/30 hover:border-slate-400/60", active: "border-slate-400 bg-slate-400/15 ring-1 ring-slate-400/40" },
    orange: { idle: "border-orange-500/30 hover:border-orange-500/60", active: "border-orange-500 bg-orange-500/15 ring-1 ring-orange-500/40" },
  };
  const s = map[accent] ?? map.slate;
  return active ? s.active : s.idle;
}
