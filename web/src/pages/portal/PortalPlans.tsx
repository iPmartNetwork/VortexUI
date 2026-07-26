import { useState, useEffect } from "react";
import { useQuery } from "@tanstack/react-query";
import { motion, AnimatePresence } from "framer-motion";
import { Check, CreditCard, Database, Smartphone, Clock, Zap, ArrowRight, X } from "lucide-react";
import { portalApi } from "./portalApi";
import { Button, Select } from "@/components/ui";
import { GlassCard } from "@/components/vortexui";
import { CardToCardInfo } from "@/components/CardToCardInfo";
import { CryptoPaySelector } from "@/components/CryptoCurrencySelector";
import { configuredCryptoCoins } from "@/lib/crypto-currencies";
import type { BillingSettings } from "@/api/wallet-billing-hooks";
import { useToast } from "@/components/toast";
import { useI18n } from "@/i18n/i18n";
import { formatBytes, cn } from "@/lib/utils";

interface Plan {
  id: string;
  name: string;
  description: string;
  data_limit: number;
  duration_days: number;
  device_limit: number;
  price_toman: number;
  price_usd: number;
}

interface DashboardData {
  sub_token: string;
}

function fileToBase64(file: File): Promise<string> {
  return new Promise((resolve, reject) => {
    const reader = new FileReader();
    reader.onload = () => resolve(reader.result as string);
    reader.onerror = reject;
    reader.readAsDataURL(file);
  });
}

function formatPrice(plan: Plan): string {
  if (plan.price_toman > 0) {
    if (plan.price_toman >= 1_000_000) return `${(plan.price_toman / 1_000_000).toFixed(1)}M`;
    if (plan.price_toman >= 1_000) return `${(plan.price_toman / 1_000).toFixed(0)}K`;
    return String(plan.price_toman);
  }
  if (plan.price_usd > 0) {
    if (plan.price_usd >= 1) return `$${plan.price_usd.toFixed(2)}`;
    return `$${plan.price_usd}`;
  }
  return "Free";
}

function formatCurrency(plan: Plan): string {
  if (plan.price_toman > 0) return "Toman";
  if (plan.price_usd > 0) return "USD";
  return "";
}

export function PortalPlans() {
  const { t } = useI18n();
  const toast = useToast();
  const [selectedPlan, setSelectedPlan] = useState<Plan | null>(null);
  const [gateway, setGateway] = useState<string>("");
  const [purchasing, setPurchasing] = useState(false);
  const [txId, setTxId] = useState("");
  const [proofImage, setProofImage] = useState<string>("");
  const [cryptoCoin, setCryptoCoin] = useState("USDT");

  const { data, isLoading } = useQuery({
    queryKey: ["portal-plans"],
    queryFn: () => portalApi<{ plans: Plan[] }>("/api/portal/plans"),
  });

  const { data: dashData } = useQuery({
    queryKey: ["portal-dashboard"],
    queryFn: () => portalApi<DashboardData>("/api/portal/dashboard"),
  });

  const { data: paymentData } = useQuery({
    queryKey: ["portal-payment-info"],
    queryFn: () => portalApi<{ settings: BillingSettings }>("/api/portal/payment-info"),
  });
  const paymentSettings = paymentData?.settings;

  useEffect(() => {
    setTxId("");
    setProofImage("");
    const coins = configuredCryptoCoins(paymentSettings?.crypto_addresses);
    setCryptoCoin(coins[0]?.id ?? "USDT");
  }, [selectedPlan?.id, gateway, paymentSettings?.crypto_addresses]);

  async function handleProofFile(e: React.ChangeEvent<HTMLInputElement>) {
    const file = e.target.files?.[0];
    if (!file) return;
    const base64 = await fileToBase64(file);
    setProofImage(base64);
  }

  async function handlePurchase() {
    if (!selectedPlan || !gateway) return;
    const subToken = dashData?.sub_token;
    if (!subToken) {
      toast.error("Could not resolve subscription token. Please re-login.");
      return;
    }

    if (gateway === "card_to_card") {
      if (!proofImage && !txId) {
        toast.error("Please upload a receipt image or enter a reference number.");
        return;
      }
    } else if (gateway === "crypto") {
      if (!txId) {
        toast.error("Please enter the transaction hash.");
        return;
      }
    }

    setPurchasing(true);
    try {
      const body: Record<string, unknown> = {
        plan_id: selectedPlan.id,
        sub_token: subToken,
        gateway,
      };

      if (gateway === "card_to_card") {
        body.tx_id = txId || "receipt";
        if (proofImage) body.proof_image = proofImage;
      } else if (gateway === "crypto") {
        body.tx_id = txId;
        body.crypto_coin = cryptoCoin;
        if (proofImage) body.proof_image = proofImage;
      }

      const res = await fetch("/api/shop/purchase", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(body),
      });
      if (!res.ok) {
        const err = await res.json().catch(() => ({}));
        throw new Error(err.message || `Purchase failed (${res.status})`);
      }
      const result = await res.json();
      if (result.redirect_url) {
        window.location.href = result.redirect_url;
      } else if (result.status === "pending_review") {
        toast.success("Payment submitted for review.");
        setSelectedPlan(null);
        setTxId("");
        setProofImage("");
      } else {
        toast.success("Purchase initiated successfully.");
        setSelectedPlan(null);
      }
    } catch (err: any) {
      toast.error(err.message || "Purchase failed. Please try again.");
    } finally {
      setPurchasing(false);
    }
  }

  function openGatewaySelector(plan: Plan) {
    setSelectedPlan(plan);
    if (plan.price_toman > 0) setGateway("zarinpal");
    else if (plan.price_usd > 0) setGateway("crypto");
    else setGateway("");
  }

  if (isLoading) {
    return (
      <div className="flex items-center justify-center py-20">
        <div className="h-8 w-8 animate-spin rounded-full border-2 border-primary border-t-transparent" />
      </div>
    );
  }

  const plans = data?.plans ?? [];

  return (
    <div className="space-y-6 animate-page-enter">
      {/* ── Header ── */}
      <div className="relative overflow-hidden rounded-2xl border border-border/60 bg-gradient-to-br from-bg-elevated via-surface to-primary/[0.03] p-5 md:p-6">
        <div className="absolute top-0 end-0 w-56 h-56 rounded-full bg-primary/8 blur-3xl pointer-events-none" />
        <div className="absolute bottom-0 start-0 w-40 h-40 rounded-full bg-accent/8 blur-3xl pointer-events-none" />
        
        <div className="relative z-10">
          <h1 className="text-xl md:text-2xl font-black text-fg tracking-tight flex items-center gap-2">
            <CreditCard size={22} className="text-primary" />
            {t("portal.plans.title")}
          </h1>
          <p className="text-[13px] text-fg-muted mt-1.5 max-w-lg">{t("portal.plans.subtitle")}</p>
        </div>
      </div>

      {/* ── Plans Grid ── */}
      <div className="grid grid-cols-1 gap-4 md:grid-cols-2 lg:grid-cols-3">
        {plans.map((p, i) => {
          const isSelected = selectedPlan?.id === p.id;
          return (
            <motion.div
              key={p.id}
              initial={{ opacity: 0, y: 12 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ delay: i * 0.07 }}
            >
              <GlassCard
                glow
                className={cn(
                  "relative overflow-hidden transition-all duration-300",
                  isSelected && "ring-2 ring-primary shadow-lg shadow-primary/20",
                  "hover:shadow-lg",
                )}
              >
                {/* Gradient line at top */}
                <div className="absolute inset-x-0 top-0 h-0.5 bg-gradient-to-r from-primary/0 via-primary to-accent/0" />

                {/* Header */}
                <div className="space-y-1">
                  <div className="flex items-center justify-between">
                    <h3 className="text-base font-bold text-fg">{p.name}</h3>
                    {i === 1 && plans.length > 2 && (
                      <span className="text-[8px] font-black uppercase tracking-widest bg-gradient-to-r from-primary to-accent text-transparent bg-clip-text border border-primary/30 rounded-full px-2 py-0.5">
                        Popular
                      </span>
                    )}
                  </div>
                  {p.description && (
                    <p className="text-xs text-fg-muted leading-relaxed">{p.description}</p>
                  )}
                </div>

                {/* Price */}
                <div className="mt-4 flex items-baseline gap-0.5">
                  <span className="text-2xl md:text-3xl font-black text-fg tracking-tight">
                    {formatPrice(p)}
                  </span>
                  <span className="text-xs text-fg-subtle ml-1">{formatCurrency(p)}</span>
                </div>

                {/* Features */}
                <div className="mt-4 grid grid-cols-2 gap-y-2 gap-x-4">
                  <div className="flex items-center gap-2 text-xs">
                    <Database size={13} className="text-primary shrink-0" />
                    <span className="text-fg-muted">
                      <strong className="text-fg">{formatBytes(p.data_limit, false)}</strong>
                    </span>
                  </div>
                  <div className="flex items-center gap-2 text-xs">
                    <Clock size={13} className="text-accent shrink-0" />
                    <span className="text-fg-muted">
                      <strong className="text-fg">{p.duration_days}d</strong>
                    </span>
                  </div>
                  <div className="flex items-center gap-2 text-xs">
                    <Smartphone size={13} className="text-success shrink-0" />
                    <span className="text-fg-muted">
                      <strong className="text-fg">{p.device_limit || "∞"}</strong> devices
                    </span>
                  </div>
                </div>

                {/* Actions */}
                <div className="mt-5 flex items-center justify-between border-t border-border/40 pt-4">
                  <Button
                    size="sm"
                    onClick={() => openGatewaySelector(p)}
                    className={cn(
                      "flex-1 transition-all duration-200",
                      isSelected && "ring-2 ring-primary",
                    )}
                  >
                    {isSelected ? (
                      <span className="flex items-center gap-1.5">
                        <Check size={14} /> Selected
                      </span>
                    ) : (
                      <span className="flex items-center gap-1.5">
                        {t("portal.plans.purchase")} <ArrowRight size={14} />
                      </span>
                    )}
                  </Button>
                </div>

                {/* ── Inline Gateway Selector ── */}
                <AnimatePresence>
                  {isSelected && (
                    <motion.div
                      initial={{ height: 0, opacity: 0 }}
                      animate={{ height: "auto", opacity: 1 }}
                      exit={{ height: 0, opacity: 0 }}
                      transition={{ duration: 0.3, ease: [0.16, 1, 0.3, 1] }}
                      className="overflow-hidden"
                    >
                      <div className="space-y-3 border-t border-border/40 pt-4 mt-4">
                        <label className="block text-[10px] font-bold uppercase tracking-wider text-fg-subtle/70">
                          {t("portal.plans.gateway")}
                        </label>
                        <Select value={gateway} onChange={(e) => setGateway(e.target.value)} className="text-xs">
                          <option value="" disabled>Select gateway...</option>
                          {p.price_toman > 0 && <option value="zarinpal">ZarinPal (Toman)</option>}
                          {p.price_toman > 0 && <option value="card_to_card">Card to Card</option>}
                          {p.price_usd > 0 && <option value="crypto">Crypto (USD)</option>}
                        </Select>

                        {/* Card-to-Card */}
                        {gateway === "card_to_card" && (
                          <div className="space-y-2.5 rounded-xl bg-surface-2/40 border border-border/40 p-3">
                            <p className="text-[11px] text-fg-muted">{t("portal.plans.cardTransferHint")}</p>
                            {paymentSettings && <CardToCardInfo settings={paymentSettings} />}
                            <div>
                              <label className="block text-[10px] font-semibold text-fg-muted mb-1">
                                {t("portal.plans.receiptRequired")}
                              </label>
                              <input
                                type="file"
                                accept="image/*"
                                onChange={handleProofFile}
                                className="block w-full text-xs text-fg-muted file:mr-2 file:rounded-lg file:border-0 file:bg-primary/10 file:px-3 file:py-1.5 file:text-xs file:font-semibold file:text-primary hover:file:bg-primary/20 transition"
                              />
                              {proofImage && (
                                <img src={proofImage} alt="Receipt" className="mt-2 max-h-32 rounded-lg border border-border/40 object-contain" />
                              )}
                            </div>
                            <div>
                              <label className="block text-[10px] font-semibold text-fg-muted mb-1">Reference (optional)</label>
                              <input
                                type="text"
                                value={txId}
                                onChange={(e) => setTxId(e.target.value)}
                                placeholder="Reference number"
                                className="w-full rounded-lg border border-border/60 bg-surface/60 px-3 py-2 text-xs text-fg placeholder:text-fg-subtle/40 focus:outline-none focus:border-primary/50 transition"
                              />
                            </div>
                          </div>
                        )}

                        {/* Crypto */}
                        {gateway === "crypto" && paymentSettings && (
                          <div className="space-y-2.5 rounded-xl bg-surface-2/40 border border-border/40 p-3">
                            <p className="text-[11px] text-fg-muted">{t("portal.plans.cryptoTransferHint")}</p>
                            <CryptoPaySelector
                              addresses={paymentSettings.crypto_addresses ?? {}}
                              selected={cryptoCoin}
                              onSelect={setCryptoCoin}
                            />
                            <div>
                              <label className="block text-[10px] font-semibold text-fg-muted mb-1">
                                {t("portal.plans.txHashRequired")}
                              </label>
                              <input
                                type="text"
                                value={txId}
                                onChange={(e) => setTxId(e.target.value)}
                                placeholder={t("portal.plans.txHashPlaceholder")}
                                className="w-full rounded-lg border border-border/60 bg-surface/60 px-3 py-2 text-xs text-fg placeholder:text-fg-subtle/40 focus:outline-none focus:border-primary/50 transition"
                              />
                            </div>
                            <div>
                              <label className="block text-[10px] font-semibold text-fg-muted mb-1">
                                {t("portal.plans.cryptoScreenshot")}
                              </label>
                              <input
                                type="file"
                                accept="image/*"
                                onChange={handleProofFile}
                                className="block w-full text-xs text-fg-muted file:mr-2 file:rounded-lg file:border-0 file:bg-primary/10 file:px-3 file:py-1.5 file:text-xs file:font-semibold file:text-primary hover:file:bg-primary/20 transition"
                              />
                              {proofImage && (
                                <img src={proofImage} alt="Screenshot" className="mt-2 max-h-32 rounded-lg border border-border/40 object-contain" />
                              )}
                            </div>
                          </div>
                        )}

                        {/* ZarinPal */}
                        {gateway === "zarinpal" && (
                          <div className="rounded-xl bg-primary/5 border border-primary/20 p-3 text-xs text-fg-muted">
                            You will be redirected to <strong className="text-primary">ZarinPal</strong> to complete the payment.
                          </div>
                        )}

                        <div className="flex gap-2 pt-1">
                          <Button
                            size="sm"
                            className="flex-1"
                            disabled={!gateway || purchasing}
                            onClick={handlePurchase}
                          >
                            {purchasing ? (
                              <span className="flex items-center gap-1.5">
                                <div className="h-3 w-3 animate-spin rounded-full border-2 border-white border-t-transparent" />
                                {t("portal.plans.processing")}
                              </span>
                            ) : (
                              <span className="flex items-center gap-1.5">
                                <Zap size={14} />
                                {t("portal.plans.confirm")}
                              </span>
                            )}
                          </Button>
                          <Button
                            size="sm"
                            variant="ghost"
                            onClick={() => setSelectedPlan(null)}
                          >
                            <X size={14} />
                          </Button>
                        </div>
                      </div>
                    </motion.div>
                  )}
                </AnimatePresence>
              </GlassCard>
            </motion.div>
          );
        })}
        {plans.length === 0 && (
          <p className="col-span-full text-center text-sm text-fg-muted py-12">{t("portal.plans.empty")}</p>
        )}
      </div>
    </div>
  );
}
