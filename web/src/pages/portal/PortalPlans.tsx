import { useQuery } from "@tanstack/react-query";
import { portalApi } from "./portalApi";
import { Card, Button } from "@/components/ui";
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

export function PortalPlans() {
  const { data, isLoading } = useQuery({
    queryKey: ["portal-plans"],
    queryFn: () => portalApi<{ plans: Plan[] }>("/api/portal/plans"),
  });

  if (isLoading) return <div className="p-8 text-center text-fg-muted">Loading plans...</div>;

  return (
    <div className="space-y-6 animate-fade-in">
      <h1 className="text-xl font-bold text-fg">Available Plans</h1>
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
              <Button size="sm" onClick={() => window.open(`/api/shop/purchase?plan_id=${p.id}`, "_blank")}>
                Purchase
              </Button>
            </div>
          </Card>
        ))}
        {(!data?.plans || data.plans.length === 0) && (
          <p className="col-span-full text-center text-sm text-fg-muted py-8">No plans available.</p>
        )}
      </div>
    </div>
  );
}
