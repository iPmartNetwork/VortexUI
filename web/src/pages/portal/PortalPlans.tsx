import { useState, useEffect } from "react";
import { useQuery } from "@tanstack/react-query";
import { portalApi } from "./portalApi";
import { Button, Select } from "@/components/ui";
import { GlassCard } from "@/components/veltrix";
import { useToast } from "@/components/toast";
import { formatBytes } from "@/lib/utils";

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

export function PortalPlans() {
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

  // Reset state when selectedPlan or gateway changes
  useEffect(() => {
    setTxId("");
    setProofImage("");
    setCryptoCoin("USDT");
  }, [selectedPlan?.id, gateway]);

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

    // Validate per method
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
    // Default gateway selection
    if (plan.price_toman > 0) setGateway("zarinpal");
    else if (plan.price_usd > 0) setGateway("crypto");
    else setGateway("");
  }

  if (isLoading) return <div className="p-8 text-center text-fg-muted">Loading plans...</div>;

  return (
    <div className="space-y-6 animate-page-enter">
      <h1 className="text-xl font-bold text-fg">Available Plans</h1>
      <p className="text-sm text-fg-muted">
        Purchase a plan to extend your subscription (traffic + duration are added to your current balance).
      </p>

      <div className="grid grid-cols-1 gap-4 md:grid-cols-2 lg:grid-cols-3">
        {data?.plans?.map((p) => (
          <GlassCard key={p.id} className="space-y-4">
            <h3 className="text-sm font-bold text-fg">{p.name}</h3>
            {p.description && <p className="text-xs text-fg-muted">{p.description}</p>}
            <div className="grid grid-cols-2 gap-2 text-xs">
              <div><span className="text-fg-subtle">Data:</span> <strong>{formatBytes(p.data_limit, false)}</strong></div>
              <div><span className="text-fg-subtle">Duration:</span> <strong>{p.duration_days}d</strong></div>
              <div><span className="text-fg-subtle">Devices:</span> <strong>{p.device_limit || "\u221E"}</strong></div>
            </div>
            <div className="flex items-center justify-between border-t border-border/40 pt-3">
              <div className="text-sm font-bold text-primary">
                {p.price_toman > 0 ? `${p.price_toman.toLocaleString()} \u062A\u0648\u0645\u0627\u0646` : p.price_usd > 0 ? `$${p.price_usd}` : "Free"}
              </div>
              <Button size="sm" onClick={() => openGatewaySelector(p)}>
                Purchase
              </Button>
            </div>

            {/* Inline gateway selector */}
            {selectedPlan?.id === p.id && (
              <div className="space-y-3 border-t border-border/40 pt-3 animate-page-enter">
                <label className="block text-xs font-medium text-fg-muted">Payment Gateway</label>
                <Select value={gateway} onChange={(e) => setGateway(e.target.value)}>
                  <option value="" disabled>Select gateway...</option>
                  {p.price_toman > 0 && <option value="zarinpal">ZarinPal (\u062A\u0648\u0645\u0627\u0646)</option>}
                  {p.price_toman > 0 && <option value="card_to_card">Card to Card (\u06A9\u0627\u0631\u062A \u0628\u0647 \u06A9\u0627\u0631\u062A)</option>}
                  {p.price_usd > 0 && <option value="crypto">Crypto (USD)</option>}
                </Select>

                {/* Card-to-Card form */}
                {gateway === "card_to_card" && (
                  <div className="space-y-2">
                    <p className="text-[11px] text-fg-subtle">Transfer the amount to the card shown below, then upload your receipt.</p>
                    <label className="block text-xs font-medium text-fg-muted">Upload receipt image (required)</label>
                    <input
                      type="file"
                      accept="image/*"
                      onChange={handleProofFile}
                      className="block w-full text-xs text-fg-muted file:mr-2 file:rounded file:border-0 file:bg-surface-2 file:px-3 file:py-1.5 file:text-xs file:text-fg"
                    />
                    {proofImage && (
                      <img src={proofImage} alt="Receipt preview" className="mt-2 max-w-full h-auto max-h-40 rounded-lg border border-border/40 object-contain" />
                    )}
                    <label className="block text-xs font-medium text-fg-muted">Reference number (optional)</label>
                    <input
                      type="text"
                      value={txId}
                      onChange={(e) => setTxId(e.target.value)}
                      placeholder="Reference number (optional)"
                      className="w-full rounded-lg border border-border bg-surface px-3 py-2 text-xs text-fg"
                    />
                  </div>
                )}

                {/* Crypto form */}
                {gateway === "crypto" && (
                  <div className="space-y-2">
                    <p className="text-[11px] text-fg-subtle">Send the exact amount to the address, then provide the transaction hash.</p>
                    <label className="block text-xs font-medium text-fg-muted">Transaction Hash (TX ID) *</label>
                    <input
                      type="text"
                      value={txId}
                      onChange={(e) => setTxId(e.target.value)}
                      placeholder="Enter transaction hash"
                      className="w-full rounded-lg border border-border bg-surface px-3 py-2 text-xs text-fg"
                    />
                    <label className="block text-xs font-medium text-fg-muted">Coin</label>
                    <Select value={cryptoCoin} onChange={(e) => setCryptoCoin(e.target.value)}>
                      <option value="USDT">USDT</option>
                      <option value="BTC">BTC</option>
                      <option value="ETH">ETH</option>
                      <option value="TRX">TRX</option>
                    </Select>
                    <label className="block text-xs font-medium text-fg-muted">Upload transfer screenshot (optional)</label>
                    <input
                      type="file"
                      accept="image/*"
                      onChange={handleProofFile}
                      className="block w-full text-xs text-fg-muted file:mr-2 file:rounded file:border-0 file:bg-surface-2 file:px-3 file:py-1.5 file:text-xs file:text-fg"
                    />
                    {proofImage && (
                      <img src={proofImage} alt="Screenshot preview" className="mt-2 max-w-full h-auto max-h-40 rounded-lg border border-border/40 object-contain" />
                    )}
                  </div>
                )}

                {/* ZarinPal info */}
                {gateway === "zarinpal" && (
                  <p className="text-[11px] text-fg-subtle">You will be redirected to ZarinPal to complete the payment.</p>
                )}

                <div className="flex gap-2">
                  <Button
                    size="sm"
                    disabled={!gateway || purchasing}
                    onClick={handlePurchase}
                  >
                    {purchasing ? "Processing..." : "Confirm"}
                  </Button>
                  <Button
                    size="sm"
                    variant="ghost"
                    onClick={() => setSelectedPlan(null)}
                  >
                    Cancel
                  </Button>
                </div>
              </div>
            )}
          </GlassCard>
        ))}
        {(!data?.plans || data.plans.length === 0) && (
          <p className="col-span-full text-center text-sm text-fg-muted py-8">No plans available.</p>
        )}
      </div>
    </div>
  );
}
