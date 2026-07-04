import { useQuery } from "@tanstack/react-query";
import { api } from "@/api/client";
import { Badge, Card, PageHeader } from "@/components/ui";

interface Order {
  id: string;
  user_id: string | null;
  plan_id: string;
  username: string;
  status: string;
  gateway: string;
  gateway_id: string;
  amount: number;
  currency: string;
  created_at: string;
  paid_at: string | null;
}

function useOrders() {
  return useQuery({ queryKey: ["orders"], queryFn: () => api<{ orders: Order[] }>("/api/orders"), refetchInterval: 10000 });
}

const statusColor: Record<string, string> = {
  pending: "warning",
  paid: "active",
  failed: "danger",
  cancelled: "muted",
  refunded: "info",
};

export function Orders() {
  const orders = useOrders();

  return (
    <div className="space-y-6 animate-page-enter">
      <PageHeader title="Orders" />

      <Card className="p-0 overflow-hidden">
        <table className="w-full text-sm">
          <thead className="border-b bg-surface-2/30 text-left text-xs text-fg-muted">
            <tr>
              <th className="px-4 py-3 font-medium">Username</th>
              <th className="px-4 py-3 font-medium">Gateway</th>
              <th className="px-4 py-3 font-medium">Amount</th>
              <th className="px-4 py-3 font-medium">Status</th>
              <th className="px-4 py-3 font-medium">Date</th>
            </tr>
          </thead>
          <tbody className="divide-y divide-border/30">
            {orders.data?.orders?.map((o) => (
              <tr key={o.id} className="hover:bg-surface-2/20">
                <td className="px-4 py-3 font-medium">{o.username || "—"}</td>
                <td className="px-4 py-3 capitalize text-fg-muted">{o.gateway}</td>
                <td className="px-4 py-3 font-mono tabular-nums">
                  {o.currency === "IRR" ? `${o.amount.toLocaleString()} ت` : `$${(o.amount / 100).toFixed(2)}`}
                </td>
                <td className="px-4 py-3">
                  <Badge color={statusColor[o.status] || "muted"}>{o.status}</Badge>
                </td>
                <td className="px-4 py-3 text-xs text-fg-subtle">{new Date(o.created_at).toLocaleDateString()}</td>
              </tr>
            ))}
            {orders.data?.orders?.length === 0 && (
              <tr><td colSpan={5} className="px-4 py-8 text-center text-fg-muted">No orders yet</td></tr>
            )}
          </tbody>
        </table>
      </Card>
    </div>
  );
}
