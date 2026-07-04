import { useEffect, useState } from "react";
import {
  useBillingDeposits,
  useBillingSettings,
  useCreateWalletPackage,
  useDeleteWalletPackage,
  useReviewDeposit,
  useSaveBillingSettings,
  useUpdateWalletPackage,
  useWalletPackages,
  formatPrice,
  type WalletDeposit,
  type WalletPackage,
} from "@/api/wallet-billing-hooks";
import { Badge, Button, Card, Input, PageHeader } from "@/components/ui";
import { Modal } from "@/components/Modal";
import { useConfirm } from "@/components/confirm";
import { useToast } from "@/components/toast";
import { useI18n } from "@/i18n/i18n";
import { formatBytes } from "@/lib/utils";
import { CryptoAddressEditor, PackageCurrencyPicker } from "@/components/CryptoCurrencySelector";
import { mergeCryptoAddresses } from "@/lib/crypto-currencies";

const METHODS = ["zarinpal", "card_to_card", "crypto", "nowpayments"] as const;

export function WalletBilling() {
  const { t } = useI18n();
  const toast = useToast();
  const confirm = useConfirm();
  const [tab, setTab] = useState<"deposits" | "packages" | "settings">("deposits");
  const pending = useBillingDeposits("pending");
  const packages = useWalletPackages(false);
  const settings = useBillingSettings();
  const review = useReviewDeposit();
  const saveSettings = useSaveBillingSettings();
  const createPkg = useCreateWalletPackage();
  const updatePkg = useUpdateWalletPackage();
  const delPkg = useDeleteWalletPackage();
  const [reviewDeposit, setReviewDeposit] = useState<WalletDeposit | null>(null);
  const [reviewNote, setReviewNote] = useState("");
  const [pkgOpen, setPkgOpen] = useState(false);
  const [editPkg, setEditPkg] = useState<WalletPackage | null>(null);
  const [form, setForm] = useState({
    card_number: "",
    card_holder: "",
    card_bank: "",
    manual_instructions: "",
    crypto_addresses: mergeCryptoAddresses(),
  });

  useEffect(() => {
    if (!settings.data?.settings) return;
    const s = settings.data.settings;
    setForm({
      card_number: s.card_number,
      card_holder: s.card_holder,
      card_bank: s.card_bank,
      manual_instructions: s.manual_instructions,
      crypto_addresses: mergeCryptoAddresses(s.crypto_addresses),
    });
  }, [settings.data]);

  async function onReview(approve: boolean) {
    if (!reviewDeposit) return;
    try {
      await review.mutateAsync({ id: reviewDeposit.id, action: approve ? "approve" : "reject", note: reviewNote });
      toast.success(approve ? t("billing.reviewApproved") : t("billing.reviewRejected"));
      setReviewDeposit(null);
      setReviewNote("");
    } catch {
      toast.error(t("billing.reviewFailed"));
    }
  }

  async function onSaveSettings() {
    try {
      await saveSettings.mutateAsync({
        card_number: form.card_number,
        card_holder: form.card_holder,
        card_bank: form.card_bank,
        manual_instructions: form.manual_instructions,
        crypto_addresses: form.crypto_addresses,
      });
      toast.success(t("billing.settingsSaved"));
    } catch {
      toast.error(t("billing.settingsFailed"));
    }
  }

  return (
    <div className="space-y-6 animate-page-enter">
      <PageHeader title={t("billing.title")} subtitle={t("billing.subtitle")} />

      <div className="flex flex-wrap gap-2">
        {(["deposits", "packages", "settings"] as const).map((k) => (
          <Button key={k} variant={tab === k ? "primary" : "ghost"} size="sm" onClick={() => setTab(k)}>
            {t(`billing.tab.${k}`)}
          </Button>
        ))}
      </div>

      {tab === "deposits" && (
        <Card className="p-0 overflow-x-auto">
          <table className="w-full text-sm">
            <thead className="border-b text-left text-muted-foreground">
              <tr>
                <th className="px-4 py-2">{t("billing.colReseller")}</th>
                <th className="px-4 py-2">{t("billing.colPackage")}</th>
                <th className="px-4 py-2">{t("billing.colMethod")}</th>
                <th className="px-4 py-2">{t("billing.colAmount")}</th>
                <th className="px-4 py-2">{t("billing.colWhen")}</th>
                <th className="px-4 py-2"></th>
              </tr>
            </thead>
            <tbody>
              {(pending.data?.deposits ?? []).map((d) => (
                <tr key={d.id} className="border-b last:border-0">
                  <td className="px-4 py-2 font-medium">{d.admin_username ?? d.admin_id.slice(0, 8)}</td>
                  <td className="px-4 py-2">{d.package_name ?? "—"}</td>
                  <td className="px-4 py-2"><Badge>{d.method}</Badge></td>
                  <td className="px-4 py-2">{formatPrice(d.amount, d.currency)}</td>
                  <td className="px-4 py-2 text-muted-foreground">{new Date(d.created_at).toLocaleString()}</td>
                  <td className="px-4 py-2 text-right">
                    <Button variant="ghost" size="sm" onClick={() => setReviewDeposit(d)}>{t("billing.review")}</Button>
                  </td>
                </tr>
              ))}
              {(pending.data?.deposits ?? []).length === 0 && (
                <tr><td colSpan={6} className="px-4 py-6 text-muted-foreground">{t("billing.noPending")}</td></tr>
              )}
            </tbody>
          </table>
        </Card>
      )}

      {tab === "packages" && (
        <div className="space-y-4">
          <div className="flex justify-end">
            <Button onClick={() => { setEditPkg(null); setPkgOpen(true); }}>{t("billing.newPackage")}</Button>
          </div>
          <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
            {(packages.data?.packages ?? []).map((p) => (
              <Card key={p.id} className="space-y-2 p-4">
                <div className="flex items-center justify-between">
                  <h3 className="font-semibold">{p.name}</h3>
                  <Badge color={p.enabled ? "active" : "expired"}>{p.enabled ? t("common.enabled") : t("common.disabled")}</Badge>
                </div>
                <p className="text-xs text-muted-foreground">{p.description}</p>
                <div className="text-sm">{formatBytes(p.traffic_bytes, false)} · {p.user_credits} users</div>
                <div className="font-medium">{formatPrice(p.price_amount, p.currency)}</div>
                <div className="flex flex-wrap gap-1">{p.methods.map((m) => <Badge key={m}>{m}</Badge>)}</div>
                <div className="flex gap-2 pt-2">
                  <Button variant="ghost" size="sm" onClick={() => { setEditPkg(p); setPkgOpen(true); }}>{t("common.edit")}</Button>
                  <Button variant="ghost" size="sm" className="text-destructive" onClick={async () => {
                    const ok = await confirm({ title: t("billing.deletePackage"), destructive: true, confirmLabel: t("common.delete") });
                    if (!ok) return;
                    await delPkg.mutateAsync(p.id);
                  }}>{t("common.delete")}</Button>
                </div>
              </Card>
            ))}
          </div>
        </div>
      )}

      {tab === "settings" && (
        <Card className="grid gap-3 p-5 sm:grid-cols-2">
          <div className="sm:col-span-2 text-sm font-medium">{t("billing.cardToCard")}</div>
          <Input placeholder={t("billing.cardNumber")} value={form.card_number} onChange={(e) => setForm({ ...form, card_number: e.target.value })} />
          <Input placeholder={t("billing.cardHolder")} value={form.card_holder} onChange={(e) => setForm({ ...form, card_holder: e.target.value })} />
          <Input placeholder={t("billing.cardBank")} value={form.card_bank} onChange={(e) => setForm({ ...form, card_bank: e.target.value })} />
          <CryptoAddressEditor
            addresses={form.crypto_addresses}
            onChange={(crypto_addresses) => setForm({ ...form, crypto_addresses })}
          />
          <Input className="sm:col-span-2" placeholder={t("billing.manualInstructions")} value={form.manual_instructions} onChange={(e) => setForm({ ...form, manual_instructions: e.target.value })} />
          <Button onClick={onSaveSettings} disabled={saveSettings.isPending}>{t("common.save")}</Button>
        </Card>
      )}

      {reviewDeposit && (
        <Modal open title={t("billing.reviewTitle")} onClose={() => setReviewDeposit(null)}>
          <div className="space-y-3 text-sm">
            <div><strong>{reviewDeposit.admin_username}</strong> · {reviewDeposit.package_name}</div>
            <div>{formatPrice(reviewDeposit.amount, reviewDeposit.currency)} · {reviewDeposit.method}</div>
            {reviewDeposit.crypto_coin && <div>{t("billing.colCrypto")}: <Badge>{reviewDeposit.crypto_coin}</Badge></div>}
            {reviewDeposit.tx_id && <div>TX: <code>{reviewDeposit.tx_id}</code></div>}
            {reviewDeposit.reseller_note && <div>{reviewDeposit.reseller_note}</div>}
            {reviewDeposit.proof_image && (
              <img src={reviewDeposit.proof_image} alt="" className="max-h-64 rounded border object-contain" />
            )}
            <Input placeholder={t("billing.adminNote")} value={reviewNote} onChange={(e) => setReviewNote(e.target.value)} />
            <div className="flex justify-end gap-2">
              <Button variant="ghost" onClick={() => onReview(false)}>{t("billing.reject")}</Button>
              <Button onClick={() => onReview(true)}>{t("billing.approve")}</Button>
            </div>
          </div>
        </Modal>
      )}

      {pkgOpen && (
        <PackageModal
          pkg={editPkg}
          onClose={() => setPkgOpen(false)}
          onSave={async (body) => {
            if (editPkg) await updatePkg.mutateAsync({ ...editPkg, ...body });
            else await createPkg.mutateAsync(body);
            setPkgOpen(false);
            toast.success(t("billing.packageSaved"));
          }}
        />
      )}
    </div>
  );
}

function PackageModal({
  pkg,
  onClose,
  onSave,
}: {
  pkg: WalletPackage | null;
  onClose: () => void;
  onSave: (body: Record<string, unknown>) => Promise<void>;
}) {
  const { t } = useI18n();
  const [name, setName] = useState(pkg?.name ?? "");
  const [description, setDescription] = useState(pkg?.description ?? "");
  const [trafficGb, setTrafficGb] = useState(pkg ? String(pkg.traffic_bytes / (1024 ** 3)) : "10");
  const [userCredits, setUserCredits] = useState(String(pkg?.user_credits ?? 50));
  const [price, setPrice] = useState(String(pkg?.price_amount ?? 100000));
  const [currency, setCurrency] = useState(pkg?.currency ?? "IRR");
  const [methods, setMethods] = useState<string[]>(pkg?.methods ?? ["zarinpal", "card_to_card"]);
  const [enabled, setEnabled] = useState(pkg?.enabled ?? true);

  function toggleMethod(m: string) {
    setMethods((prev) => (prev.includes(m) ? prev.filter((x) => x !== m) : [...prev, m]));
  }

  return (
    <Modal open title={pkg ? t("billing.editPackage") : t("billing.newPackage")} onClose={onClose}>
      <div className="space-y-3">
        <Input placeholder={t("billing.pkgName")} value={name} onChange={(e) => setName(e.target.value)} />
        <Input placeholder={t("billing.pkgDesc")} value={description} onChange={(e) => setDescription(e.target.value)} />
        <div className="grid grid-cols-2 gap-2">
          <Input type="number" placeholder="GB" value={trafficGb} onChange={(e) => setTrafficGb(e.target.value)} />
          <Input type="number" placeholder={t("billing.userCredits")} value={userCredits} onChange={(e) => setUserCredits(e.target.value)} />
          <Input type="number" placeholder={t("billing.price")} value={price} onChange={(e) => setPrice(e.target.value)} />
          <div className="col-span-2">
            <div className="mb-1 text-xs text-muted-foreground">{t("billing.pkgCurrency")}</div>
            <PackageCurrencyPicker value={currency} onChange={setCurrency} />
          </div>
        </div>
        <div className="flex flex-wrap gap-2">
          {METHODS.map((m) => (
            <label key={m} className="flex items-center gap-1 text-xs">
              <input type="checkbox" checked={methods.includes(m)} onChange={() => toggleMethod(m)} /> {m}
            </label>
          ))}
        </div>
        <label className="flex items-center gap-2 text-sm">
          <input type="checkbox" checked={enabled} onChange={(e) => setEnabled(e.target.checked)} /> {t("common.enabled")}
        </label>
        <div className="flex justify-end gap-2">
          <Button variant="ghost" onClick={onClose}>{t("common.cancel")}</Button>
          <Button onClick={() => onSave({
            name,
            description,
            traffic_bytes: Math.round(Number(trafficGb) * 1024 ** 3),
            user_credits: Number(userCredits),
            price_amount: Number(price),
            currency,
            methods,
            enabled,
          })}>{t("common.save")}</Button>
        </div>
      </div>
    </Modal>
  );
}
