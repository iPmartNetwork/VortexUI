import { useQuery } from "@tanstack/react-query";
import { portalApi } from "./portalApi";
import { Card, Badge } from "@/components/ui";
import { formatBytes } from "@/lib/utils";

interface DashboardData {
  username: string;
  status: string;
  data_limit: number;
  used_traffic: number;
  expire_at: string | null;
  device_limit: number;
  reset_strategy: string;
  sub_token: string;
  created_at: string;
}

export function PortalDashboard() {
  const { data, isLoading } = useQuery({
    queryKey: ["portal-dashboard"],
    queryFn: () => portalApi<DashboardData>("/api/portal/dashboard"),
  });

  if (isLoading) return <div className="p-8 text-center text-fg-muted">Loading...</div>;
  if (!data) return null;

  const usagePercent = data.data_limit > 0 ? Math.min((data.used_traffic / data.data_limit) * 100, 100) : 0;

  return (
    <div className="space-y-6 animate-fade-in">
      <h1 className="text-xl font-bold text-fg">Welcome, {data.username}</h1>

      <div className="grid grid-cols-1 gap-4 md:grid-cols-2 lg:grid-cols-3">
        <Card className="space-y-3">
          <div className="text-xs text-fg-subtle uppercase">Status</div>
          <Badge color={data.status}>{data.status}</Badge>
        </Card>

        <Card className="space-y-3">
          <div className="text-xs text-fg-subtle uppercase">Data Usage</div>
          <div className="text-lg font-bold text-fg">
            {formatBytes(data.used_traffic, false)} / {data.data_limit > 0 ? formatBytes(data.data_limit, false) : "∞"}
          </div>
          {data.data_limit > 0 && (
            <div className="h-2 rounded-full bg-surface-2 overflow-hidden">
              <div
                className="h-full rounded-full grad-bg transition-all"
                style={{ width: `${usagePercent}%` }}
              />
            </div>
          )}
        </Card>

        <Card className="space-y-3">
          <div className="text-xs text-fg-subtle uppercase">Expires</div>
          <div className="text-sm font-medium text-fg">
            {data.expire_at ? new Date(data.expire_at).toLocaleDateString() : "Never"}
          </div>
        </Card>

        <Card className="space-y-3">
          <div className="text-xs text-fg-subtle uppercase">Device Limit</div>
          <div className="text-sm font-medium text-fg">{data.device_limit || "Unlimited"}</div>
        </Card>

        <Card className="space-y-3">
          <div className="text-xs text-fg-subtle uppercase">Reset Strategy</div>
          <div className="text-sm font-medium text-fg capitalize">{data.reset_strategy.replace("_", " ")}</div>
        </Card>

        <Card className="space-y-3">
          <div className="text-xs text-fg-subtle uppercase">Subscription</div>
          <div className="text-xs font-mono text-fg-muted break-all">{data.sub_token}</div>
        </Card>
      </div>
    </div>
  );
}
