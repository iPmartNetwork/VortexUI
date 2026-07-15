import { useMemo, useState } from "react";
import { useQuery } from "@tanstack/react-query";
import { CreditCard, History, Search, Receipt } from "lucide-react";
import { api } from "@/api/client";
import { Input, Select } from "@/components/ui";
import { GlassCard, StatsCard, StatusBadge } from "@/components/veltrix";
import { useI18n } from "@/i18n/i18n";
import { EmptyState } from "@/components/EmptyState";
import { useTitle } from "@/lib/useTitle";

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

function statusOf(s: string): string {
  if (s === "paid") return "active";
  if (s === "pending") return "warning";
  if (s === "failed") return "error";
  if (s === "refunded") return "info";
  return "inactive";
}

function formatAmount(o: Order): string {
  return o.currency === "IRR" ? `${o.amount.toLocaleString()} T` : `$${(o.amount / 100).toFixed(2)}`;
}

export function Orders() {
  useTitle("Orders");
  const { t } = useI18n();
  const orders = useOrders();
  const [status, setStatus] = useState("all");
  const [query, setQuery] = useState("");

  const list = orders.data?.orders ?? [];
  const filtered = useMemo(() => {
    return list.filter((o) => {
      if (status !== "all" && o.status !== status) return false;
      if (query && !o.username?.toLowerCase().includes(query.toLowerCase()) && !o.gateway_id?.toLowerCase().includes(query.toLowerCase())) return false;
      return true;
    });
  }, [list, status, query]);

  const paidCount = list.filter((o) => o.status === "paid").length;
  const pendingCount = list.filter((o) => o.status === "pending").length;

  return (
    <div className="space-y-5 animate-page-enter">
      <div>
        <h1 className="text-2xl font-bold text-fg tracking-tight">{t("nav.orders")}</h1>
        <p className="text-sm text-fg-muted mt-1">All purchase orders across the platform</p>
      </div>

      <div className="grid grid-cols-1 gap-4 sm:grid-cols-3">
        <StatsCard title="Total orders" value={list.length} icon={<History size={18} />} color="cyan" />
        <StatsCard title="Paid" value={paidCount} icon={<CreditCard size={18} />} color="green" />
        <StatsCard title="Pending" value={pendingCount} icon={<CreditCard size={18} />} color="orange" />
      </div>

      <div className="flex flex-col sm:flex-row gap-3">
        <div className="relative flex-1">
          <Search size={14} className="absolute start-3 top-1/2 -translate-y-1/2 text-fg-subtle" />
          <Input placeholder="Search username or gateway ID…" value={query} onChange={(e) => setQuery(e.target.value)} className="ps-9" />
        </div>
        <Select value={status} onChange={(e) => setStatus(e.target.value)} className="sm:w-44 flex-shrink-0">
          <option value="all">All statuses</option>
          <option value="pending">Pending</option>
          <option value="paid">Paid</option>
          <option value="failed">Failed</option>
          <option value="cancelled">Cancelled</option>
          <option value="refunded">Refunded</option>
        </Select>
      </div>

      <GlassCard hover={false} className="!p-0 overflow-hidden">
        <div className="overflow-x-auto">
          <table className="w-full text-sm">
            <thead>
              <tr className="border-b border-border/40 bg-surface-2/30 text-left text-[11px] uppercase tracking-wide text-fg-subtle">
                <th className="px-4 py-3 font-medium">Username</th>
                <th className="px-4 py-3 font-medium">Gateway</th>
                <th className="px-4 py-3 font-medium">Amount</th>
                <th className="px-4 py-3 font-medium">Status</th>
                <th className="px-4 py-3 font-medium">Date</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-border/20">
              {filtered.map((o) => (
                <tr key={o.id} className="hover:bg-surface-2/40">
                  <td className="px-4 py-3 font-medium text-fg">{o.username || "—"}</td>
                  <td className="px-4 py-3 capitalize text-fg-muted">{o.gateway.replace(/_/g, " ")}</td>
                  <td className="px-4 py-3 font-mono tabular-nums">{formatAmount(o)}</td>
                  <td className="px-4 py-3"><StatusBadge status={statusOf(o.status)} label={o.status} pulse={o.status === "pending"} /></td>
                  <td className="px-4 py-3 text-xs text-fg-subtle">{new Date(o.created_at).toLocaleDateString()}</td>
                </tr>
              ))}
              {filtered.length === 0 && (
                <tr><td colSpan={5} className="px-4 py-8"><EmptyState icon={Receipt} title={query || status !== "all" ? "No matching orders" : t("common.none")} compact /></td></tr>
              )}
            </tbody>
          </table>
        </div>
      </GlassCard>
    </div>
  );
}
