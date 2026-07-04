import { useMemo, useState } from "react";
import { useQuery } from "@tanstack/react-query";
import { CheckCircle2, Eye, XCircle } from "lucide-react";
import { api } from "@/api/client";
import { useReviewOrder } from "@/api/reseller-payment-hooks";
import {
  useBillingDeposits,
  useReviewDeposit,
  formatPrice,
  type WalletDeposit,
} from "@/api/wallet-billing-hooks";
import { Badge, Button, Input } from "@/components/ui";
import { Modal } from "@/components/Modal";
import { GlassCard, StatusBadge } from "@/components/veltrix";
import { useConfirm } from "@/components/confirm";
import { useToast } from "@/components/toast";
import { useI18n } from "@/i18n/i18n";

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
  proof_image?: string;
  created_at: string;
  paid_at: string | null;
}

interface Plan {
  id: string;
  name: string;
}

function useAllOrders() {
  return useQuery({
    queryKey: ["orders"],
    queryFn: () => api<{ orders: Order[] }>("/api/orders"),
    refetchInterval: 15000,
  });
}

function usePlansMap() {
  return useQuery({
    queryKey: ["plans"],
    queryFn: () => api<{ plans: Plan[] }>("/api/plans"),
  });
}

function orderRef(id: string) {
  return `ORD-${id.replace(/-/g, "").slice(0, 4).toUpperCase()}`;
}

function formatAmount(amount: number, currency: string) {
  if (currency === "IRR") return `${amount.toLocaleString()} T`;
  return `$${(amount / 100).toFixed(2)} ${currency === "USD" ? "USDT" : currency}`;
}

function gatewayLabel(g: string) {
  return g.replace(/_/g, " ").toUpperCase();
}

export function OrdersTab() {
  const { t } = useI18n();
  const toast = useToast();
  const confirm = useConfirm();
  const orders = useAllOrders();
  const plans = usePlansMap();
  const reviewOrder = useReviewOrder();
  const pendingDeposits = useBillingDeposits("pending");
  const reviewDeposit = useReviewDeposit();

  const [processingId, setProcessingId] = useState<string | null>(null);
  const [reviewDepositRow, setReviewDepositRow] = useState<WalletDeposit | null>(null);
  const [reviewNote, setReviewNote] = useState("");
  const [proofUrl, setProofUrl] = useState<string | null>(null);

  const planMap = useMemo(() => {
    const m: Record<string, string> = {};
    for (const p of plans.data?.plans ?? []) m[p.id] = p.name;
    return m;
  }, [plans.data]);

  const list = orders.data?.orders ?? [];
  const pendingOrders = list.filter((o) => o.status === "pending");
  const deposits = pendingDeposits.data?.deposits ?? [];

  async function approveOrder(id: string) {
    setProcessingId(id);
    try {
      await reviewOrder.mutateAsync({ id, action: "approve" });
      toast.success(t("pendingOrders.approved"));
    } catch {
      toast.error(t("pendingOrders.reviewFailed"));
    } finally {
      setProcessingId(null);
    }
  }

  async function rejectOrder(id: string) {
    if (!(await confirm({ title: t("pendingOrders.confirmReject"), destructive: true }))) return;
    setProcessingId(id);
    try {
      await reviewOrder.mutateAsync({ id, action: "reject" });
      toast.success(t("pendingOrders.rejected"));
    } catch {
      toast.error(t("pendingOrders.reviewFailed"));
    } finally {
      setProcessingId(null);
    }
  }

  async function onReviewDeposit(approve: boolean) {
    if (!reviewDepositRow) return;
    try {
      await reviewDeposit.mutateAsync({
        id: reviewDepositRow.id,
        action: approve ? "approve" : "reject",
        note: reviewNote,
      });
      toast.success(approve ? t("billing.reviewApproved") : t("billing.reviewRejected"));
      setReviewDepositRow(null);
      setReviewNote("");
    } catch {
      toast.error(t("billing.reviewFailed"));
    }
  }

  return (
    <div className="space-y-4">
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-2">
        <div>
          <h2 className="text-sm font-bold text-fg">{t("reseller.ordersTitle")}</h2>
          <p className="text-xs text-fg-muted mt-0.5">{t("reseller.ordersDesc")}</p>
        </div>
        {(pendingOrders.length + deposits.length) > 0 && (
          <Badge color="warning">{pendingOrders.length + deposits.length} {t("reseller.pendingBadge")}</Badge>
        )}
      </div>

      <GlassCard hover={false} className="!p-0 overflow-hidden">
        <div className="overflow-x-auto">
          <table className="w-full text-sm">
            <thead>
              <tr className="border-b border-border/40 text-[11px] uppercase tracking-wide text-fg-subtle bg-surface-2/30">
                <th className="py-3 px-4 text-left">{t("reseller.colOrder")}</th>
                <th className="py-3 px-4 text-left">{t("reseller.colUser")}</th>
                <th className="py-3 px-4 text-left">{t("reseller.colPlan")}</th>
                <th className="py-3 px-4 text-left">{t("reseller.colAmount")}</th>
                <th className="py-3 px-4 text-left">{t("reseller.colGateway")}</th>
                <th className="py-3 px-4 text-center">{t("reseller.colStatus")}</th>
                <th className="py-3 px-4 text-right">{t("common.actions")}</th>
              </tr>
            </thead>
            <tbody>
              {list.map((o) => (
                <tr key={o.id} className="border-b border-border/20 hover:bg-surface-2/40">
                  <td className="py-3 px-4 font-mono text-xs font-bold">{orderRef(o.id)}</td>
                  <td className="py-3 px-4 text-xs">{o.username || "—"}</td>
                  <td className="py-3 px-4 text-xs text-fg-muted">{planMap[o.plan_id] ?? o.plan_id.slice(0, 8)}</td>
                  <td className="py-3 px-4 font-mono text-xs tabular-nums">{formatAmount(o.amount, o.currency)}</td>
                  <td className="py-3 px-4">
                    <div className="flex flex-col gap-1">
                      <Badge color="muted">{gatewayLabel(o.gateway)}</Badge>
                      {o.proof_image && (
                        <button
                          type="button"
                          className="text-[10px] text-primary hover:underline inline-flex items-center gap-1"
                          onClick={() => setProofUrl(o.proof_image!)}
                        >
                          <Eye size={11} /> {t("reseller.viewReceipt")}
                        </button>
                      )}
                      {o.gateway_id && !o.proof_image && (
                        <span className="text-[10px] font-mono text-fg-subtle truncate max-w-[120px]" title={o.gateway_id}>
                          {o.gateway_id}
                        </span>
                      )}
                    </div>
                  </td>
                  <td className="py-3 px-4 text-center">
                    <StatusBadge
                      status={o.status === "paid" ? "active" : o.status === "pending" ? "warning" : "inactive"}
                      label={o.status === "paid" ? "APPROVED" : o.status === "pending" ? "PENDING REVIEW" : o.status.toUpperCase()}
                      pulse={o.status === "pending"}
                    />
                  </td>
                  <td className="py-3 px-4 text-right">
                    {o.status === "pending" ? (
                      <div className="flex justify-end gap-1">
                        <Button
                          size="sm"
                          disabled={processingId === o.id}
                          onClick={() => approveOrder(o.id)}
                          className="!bg-success/15 !text-success hover:!bg-success/25"
                        >
                          <CheckCircle2 size={14} /> {t("billing.approve")}
                        </Button>
                        <Button
                          size="sm"
                          variant="ghost"
                          disabled={processingId === o.id}
                          onClick={() => rejectOrder(o.id)}
                          className="!text-danger"
                        >
                          <XCircle size={14} />
                        </Button>
                      </div>
                    ) : (
                      <span className="text-xs text-fg-subtle">{t("reseller.completed")}</span>
                    )}
                  </td>
                </tr>
              ))}
              {deposits.map((d) => (
                <tr key={`dep-${d.id}`} className="border-b border-border/20 hover:bg-surface-2/40 bg-warning/5">
                  <td className="py-3 px-4 font-mono text-xs font-bold">DEP-{d.id.slice(0, 4).toUpperCase()}</td>
                  <td className="py-3 px-4 text-xs">{d.admin_username ?? d.admin_id.slice(0, 8)}</td>
                  <td className="py-3 px-4 text-xs text-fg-muted">{d.package_name ?? t("reseller.walletDeposit")}</td>
                  <td className="py-3 px-4 font-mono text-xs">{formatPrice(d.amount, d.currency)}</td>
                  <td className="py-3 px-4">
                    <Badge color="muted">{gatewayLabel(d.method)}</Badge>
                    {d.proof_image && (
                      <button type="button" className="text-[10px] text-primary hover:underline ms-1" onClick={() => setProofUrl(d.proof_image!)}>
                        <Eye size={11} />
                      </button>
                    )}
                  </td>
                  <td className="py-3 px-4 text-center">
                    <StatusBadge status="warning" label="PENDING REVIEW" pulse />
                  </td>
                  <td className="py-3 px-4 text-right">
                    <Button size="sm" variant="outline" onClick={() => setReviewDepositRow(d)}>{t("billing.review")}</Button>
                  </td>
                </tr>
              ))}
              {list.length === 0 && deposits.length === 0 && (
                <tr>
                  <td colSpan={7} className="py-10 text-center text-sm text-fg-muted">{t("reseller.ordersEmpty")}</td>
                </tr>
              )}
            </tbody>
          </table>
        </div>
      </GlassCard>

      {proofUrl && (
        <Modal open title={t("reseller.viewReceipt")} onClose={() => setProofUrl(null)}>
          <img src={proofUrl} alt="" className="max-h-[70vh] mx-auto rounded border object-contain" />
        </Modal>
      )}

      {reviewDepositRow && (
        <Modal open title={t("billing.reviewTitle")} onClose={() => setReviewDepositRow(null)}>
          <div className="space-y-3 text-sm">
            <div><strong>{reviewDepositRow.admin_username}</strong> · {reviewDepositRow.package_name}</div>
            <div>{formatPrice(reviewDepositRow.amount, reviewDepositRow.currency)} · {reviewDepositRow.method}</div>
            {reviewDepositRow.proof_image && (
              <img src={reviewDepositRow.proof_image} alt="" className="max-h-48 rounded border object-contain" />
            )}
            <Input placeholder={t("billing.adminNote")} value={reviewNote} onChange={(e) => setReviewNote(e.target.value)} />
            <div className="flex justify-end gap-2">
              <Button variant="ghost" onClick={() => onReviewDeposit(false)}>{t("billing.reject")}</Button>
              <Button onClick={() => onReviewDeposit(true)}>{t("billing.approve")}</Button>
            </div>
          </div>
        </Modal>
      )}
    </div>
  );
}
