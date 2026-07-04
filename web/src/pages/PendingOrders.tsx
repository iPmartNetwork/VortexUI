import { useState } from "react";
import { usePendingOrders, useReviewOrder } from "@/api/reseller-payment-hooks";
import { Badge, Button, Card, PageHeader } from "@/components/ui";
import { useToast } from "@/components/toast";
import { useConfirm } from "@/components/confirm";
import { useI18n } from "@/i18n/i18n";
import { CheckCircle2, XCircle } from "lucide-react";

export function PendingOrders() {
  const { t } = useI18n();
  const toast = useToast();
  const confirm = useConfirm();
  const orders = usePendingOrders();
  const review = useReviewOrder();
  const [processingId, setProcessingId] = useState<string | null>(null);

  async function handleApprove(id: string) {
    setProcessingId(id);
    try {
      await review.mutateAsync({ id, action: "approve" });
      toast.success(t("pendingOrders.approved"));
    } catch {
      toast.error(t("pendingOrders.reviewFailed"));
    } finally {
      setProcessingId(null);
    }
  }

  async function handleReject(id: string) {
    const ok = await confirm({ title: t("pendingOrders.confirmReject"), destructive: true });
    if (!ok) return;
    setProcessingId(id);
    try {
      await review.mutateAsync({ id, action: "reject" });
      toast.success(t("pendingOrders.rejected"));
    } catch {
      toast.error(t("pendingOrders.reviewFailed"));
    } finally {
      setProcessingId(null);
    }
  }

  if (orders.isLoading) {
    return <div className="p-8 text-center text-fg-muted">{t("common.loading")}</div>;
  }

  return (
    <div className="space-y-6 animate-page-enter">
      <PageHeader title={t("pendingOrders.title")} subtitle={t("pendingOrders.subtitle")} />

      <Card className="p-0 overflow-hidden">
        <table className="w-full text-sm">
          <thead className="border-b bg-surface-2/30 text-left text-xs text-fg-muted">
            <tr>
              <th className="px-4 py-3 font-medium">{t("pendingOrders.colUsername")}</th>
              <th className="px-4 py-3 font-medium">{t("pendingOrders.colAmount")}</th>
              <th className="px-4 py-3 font-medium">{t("pendingOrders.colGateway")}</th>
              <th className="px-4 py-3 font-medium">{t("pendingOrders.colTxId")}</th>
              <th className="px-4 py-3 font-medium">Proof</th>
              <th className="px-4 py-3 font-medium">{t("pendingOrders.colDate")}</th>
              <th className="px-4 py-3 font-medium">{t("common.actions")}</th>
            </tr>
          </thead>
          <tbody className="divide-y divide-border/30">
            {orders.data?.orders?.map((o) => (
              <tr key={o.id} className="hover:bg-surface-2/20">
                <td className="px-4 py-3 font-medium">{o.username || "—"}</td>
                <td className="px-4 py-3 font-mono tabular-nums">
                  {o.currency === "IRR" ? `${o.amount.toLocaleString()} ت` : `$${(o.amount / 100).toFixed(2)}`}
                </td>
                <td className="px-4 py-3">
                  <Badge color="muted">{o.gateway}</Badge>
                </td>
                <td className="px-4 py-3 text-xs text-fg-subtle font-mono max-w-[140px] truncate" title={o.gateway_id}>
                  {o.gateway_id || "—"}
                </td>
                <td className="px-4 py-3">
                  {o.proof_image ? (
                    <a href={o.proof_image} target="_blank" rel="noopener noreferrer">
                      <img src={o.proof_image} alt="Receipt" className="h-10 w-10 rounded object-cover border border-border/40 hover:opacity-80 transition" />
                    </a>
                  ) : (
                    <span className="text-fg-subtle text-xs">—</span>
                  )}
                </td>
                <td className="px-4 py-3 text-xs text-fg-subtle">
                  {new Date(o.created_at).toLocaleDateString()}
                </td>
                <td className="px-4 py-3">
                  <div className="flex items-center gap-1.5">
                    <Button
                      size="sm"
                      variant="outline"
                      disabled={processingId === o.id}
                      onClick={() => handleApprove(o.id)}
                      className="!text-success hover:!bg-success/10"
                    >
                      <CheckCircle2 size={14} />
                    </Button>
                    <Button
                      size="sm"
                      variant="outline"
                      disabled={processingId === o.id}
                      onClick={() => handleReject(o.id)}
                      className="!text-danger hover:!bg-danger/10"
                    >
                      <XCircle size={14} />
                    </Button>
                  </div>
                </td>
              </tr>
            ))}
            {(!orders.data?.orders || orders.data.orders.length === 0) && (
              <tr>
                <td colSpan={7} className="px-4 py-8 text-center text-fg-muted">
                  {t("pendingOrders.empty")}
                </td>
              </tr>
            )}
          </tbody>
        </table>
      </Card>
    </div>
  );
}
