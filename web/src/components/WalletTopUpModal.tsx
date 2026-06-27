import { useState } from "react";
import { useTopUpAdminWallet } from "@/api/reseller-hooks";
import { Modal } from "@/components/Modal";
import { Button, Input } from "@/components/ui";
import { useToast } from "@/components/toast";
import { useI18n } from "@/i18n/i18n";

export function WalletTopUpModal({
  open,
  onClose,
  adminId,
  username,
}: {
  open: boolean;
  onClose: () => void;
  adminId: string;
  username: string;
}) {
  const { t } = useI18n();
  const toast = useToast();
  const topUp = useTopUpAdminWallet();
  const [trafficGb, setTrafficGb] = useState("");
  const [userCredits, setUserCredits] = useState("");
  const [reason, setReason] = useState("");

  async function submit() {
    const gb = Number(trafficGb);
    const users = Number(userCredits);
    if ((!trafficGb || gb <= 0) && (!userCredits || users <= 0)) return;
    try {
      await topUp.mutateAsync({
        adminId,
        traffic_bytes: gb > 0 ? Math.round(gb * 1024 * 1024 * 1024) : 0,
        user_credits: users > 0 ? Math.round(users) : 0,
        reason: reason.trim() || undefined,
      });
      toast.success(t("reseller.admins.walletTopUpOk"));
      setTrafficGb("");
      setUserCredits("");
      setReason("");
      onClose();
    } catch {
      toast.error(t("reseller.admins.walletTopUpFail"));
    }
  }

  return (
    <Modal
      open={open}
      onClose={onClose}
      title={`${t("reseller.admins.walletTopUpTitle")} — ${username}`}
    >
      <div className="space-y-3">
        <Input
          type="number"
          min={0}
          step={0.1}
          placeholder={t("reseller.admins.walletTrafficGb")}
          value={trafficGb}
          onChange={(e) => setTrafficGb(e.target.value)}
        />
        <Input
          type="number"
          min={0}
          step={1}
          placeholder={t("reseller.admins.walletUserCredits")}
          value={userCredits}
          onChange={(e) => setUserCredits(e.target.value)}
        />
        <Input
          placeholder={t("reseller.admins.walletReason")}
          value={reason}
          onChange={(e) => setReason(e.target.value)}
        />
        <div className="flex justify-end gap-2 pt-2">
          <Button variant="ghost" onClick={onClose}>{t("common.cancel")}</Button>
          <Button onClick={submit} disabled={topUp.isPending}>{t("reseller.admins.walletTopUp")}</Button>
        </div>
      </div>
    </Modal>
  );
}
