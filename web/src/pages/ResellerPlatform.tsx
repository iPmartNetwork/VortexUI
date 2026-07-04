import { useSearchParams } from "react-router-dom";
import { ClipboardList, List, Wallet as WalletIcon } from "lucide-react";
import { useBillingDeposits } from "@/api/wallet-billing-hooks";
import { usePendingOrders } from "@/api/reseller-payment-hooks";
import { Badge } from "@/components/ui";
import { useI18n } from "@/i18n/i18n";
import type { TKey } from "@/i18n/dict";
import { useTitle } from "@/lib/useTitle";
import { cn } from "@/lib/utils";
import { SummaryWidgets } from "@/pages/reseller/SummaryWidgets";
import { OrdersTab } from "@/pages/reseller/OrdersTab";
import { PlansTab } from "@/pages/reseller/PlansTab";
import { WalletTab } from "@/pages/reseller/WalletTab";

export type ResellerTab = "orders" | "plans" | "wallet";

const TABS: { id: ResellerTab; icon: typeof ClipboardList; labelKey: TKey }[] = [
  { id: "orders", icon: ClipboardList, labelKey: "reseller.tabOrders" },
  { id: "plans", icon: List, labelKey: "reseller.tabPlans" },
  { id: "wallet", icon: WalletIcon, labelKey: "reseller.tabWallet" },
];

function parseTab(raw: string | null): ResellerTab {
  if (raw === "plans" || raw === "wallet") return raw;
  return "orders";
}

export function ResellerPlatform() {
  useTitle("Reseller");
  const { t } = useI18n();
  const [searchParams, setSearchParams] = useSearchParams();
  const tab = parseTab(searchParams.get("tab"));

  const pendingOrders = usePendingOrders();
  const pendingDeposits = useBillingDeposits("pending");
  const pendingCount =
    (pendingOrders.data?.orders?.length ?? 0) + (pendingDeposits.data?.deposits?.length ?? 0);

  function setTab(next: ResellerTab) {
    if (next === "orders") setSearchParams({}, { replace: true });
    else setSearchParams({ tab: next }, { replace: true });
  }

  return (
    <div className="space-y-5 animate-page-enter">
      <div className="flex flex-col xl:flex-row xl:items-start justify-between gap-4">
        <div>
          <div className="flex flex-wrap items-center gap-2">
            <h1 className="text-2xl font-bold text-fg tracking-tight">{t("reseller.pageTitle")}</h1>
            <Badge color="active">{t("reseller.badge")}</Badge>
          </div>
          <p className="text-sm text-fg-muted mt-1 max-w-2xl">{t("reseller.pageSubtitle")}</p>
        </div>
        <div className="flex flex-wrap rounded-xl border border-border/70 bg-surface-2/50 p-0.5 gap-0.5">
          {TABS.map(({ id, icon: Icon, labelKey }) => (
            <button
              key={id}
              type="button"
              onClick={() => setTab(id)}
              className={cn(
                "relative flex items-center gap-1.5 px-3 py-2 rounded-lg text-xs font-semibold transition-colors whitespace-nowrap",
                tab === id ? "bg-primary text-primary-fg shadow-sm" : "text-fg-muted hover:text-fg",
              )}
            >
              <Icon size={14} />
              {t(labelKey)}
              {id === "orders" && pendingCount > 0 && (
                <span className="absolute -top-1 -end-1 min-w-[18px] h-[18px] px-1 rounded-full bg-danger text-[10px] font-bold text-white flex items-center justify-center">
                  {pendingCount > 99 ? "99+" : pendingCount}
                </span>
              )}
            </button>
          ))}
        </div>
      </div>

      <SummaryWidgets onDeposit={() => setTab("wallet")} />

      {tab === "orders" && <OrdersTab />}
      {tab === "plans" && <PlansTab />}
      {tab === "wallet" && <WalletTab />}
    </div>
  );
}

/** @deprecated — use ResellerPlatform */
export function WalletBilling() {
  return <ResellerPlatform />;
}
