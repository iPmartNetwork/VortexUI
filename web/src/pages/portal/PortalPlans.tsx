import { useState } from "react";
import { useQuery } from "@tanstack/react-query";
import { portalApi } from "./portalApi";
import { Card, Button, Select } from "@/components/ui";
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

export function PortalPlans() {
  const toast = useToast();
  const [selectedPlan, setSelectedPlan] = useState<Plan | null>(null);
  const [gateway, setGateway] = useState<string>("");
  const [purchasing, setPurchasing] = useState(false);

  const { data, isLoading } = useQuery({
    queryKey: ["portal-plans"],
    queryFn: () => portalApi<{ plans: Plan[] }>("/api/portal/plans"),
  });

  const { data: dashData } = useQuery({
    queryKey: ["portal-dashboard"],
    queryFn: () => portalApi<DashboardData>("/api/portal/dashboard"),
  });

  async function handlePurchase() {
    if (!selectedPlan || !gateway) return;
    const subToken = dashData?.sub_token;
    if (!subToken) {
      toast.error("Could not resolve subscription token. Please re-login.");
      return;
    }
    setPurchasing(true);
    try {
      const res = await fetch("/api/shop/purchase", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ plan_id: selectedPlan.id, sub_token: subToken, gateway }),
      });
      if (!res.ok) {
        const err = await res.json().catch(() => ({}));
        throw new Error(err.message || `Purchase failed (${res.status})`);
      }
      const result = await res.json();
      if (result.redirect_url) {
        window.location.href = result.redirect_url;
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
    <div className="space-y-6 animate-fade-in">
      <h1 className="text-xl font-bold text-fg">Available Plans</h1>
      <p className="text-sm text-fg-muted">
        Purchase a plan to extend your subscription (traffic + duration are added to your current balance).
      </p>

      <div className="grid grid-cols-1 gap-4 md:grid-cols-2 lg:grid-cols-3">
        {data?.plans?.map((p) => (
          <Card key={p.id} className="space-y-4">
            <h3 className="text-sm font-bold text-fg">{p.name}</h3>
            {p.description && <p className="text-xs text-fg-muted">{p.description}</p>}
            <div className="grid grid-cols-2 gap-2 text-xs">
              <div><span className="text-fg-subtle">Data:</span> <strong>{formatBytes(p.data_limit, false)}</strong></div>
              <div><span className="text-fg-subtle">Duration:</span> <strong>{p.duration_days}d</strong></div>
              <div><span className="text-fg-subtle">Devices:</span> <strong>{p.device_limit || "∞"}</strong></div>
            </div>
            <div className="flex items-center justify-between border-t border-border/40 pt-3">
              <div className="text-sm font-bold text-primary">
                {p.price_toman > 0 ? `${p.price_toman.toLocaleString()} تومان` : p.price_usd > 0 ? `$${p.price_usd}` : "Free"}
              </div>
              <Button size="sm" onClick={() => openGatewaySelector(p)}>
                Purchase
              </Button>
            </div>

            {/* Inline gateway selector */}
            {selectedPlan?.id === p.id && (
              <div className="space-y-3 border-t border-border/40 pt-3 animate-fade-in">
                <label className="block text-xs font-medium text-fg-muted">Payment Gateway</label>
                <Select value={gateway} onChange={(e) => setGateway(e.target.value)}>
                  <option value="" disabled>Select gateway...</option>
                  {p.price_toman > 0 && <option value="zarinpal">ZarinPal (تومان)</option>}
                  {p.price_usd > 0 && <option value="crypto">Crypto (USD)</option>}
                </Select>
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
          </Card>
        ))}
        {(!data?.plans || data.plans.length === 0) && (
          <p className="col-span-full text-center text-sm text-fg-muted py-8">No plans available.</p>
        )}
      </div>
    </div>
  );
}
