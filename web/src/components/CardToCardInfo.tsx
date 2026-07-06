import { useI18n } from "@/i18n/i18n";
import type { BillingSettings } from "@/api/wallet-billing-hooks";

export function CardToCardInfo({ settings }: { settings: Pick<BillingSettings, "card_number" | "card_holder" | "card_bank" | "manual_instructions"> }) {
  const { t } = useI18n();

  if (!settings.card_number) {
    return (
      <p className="text-xs text-fg-muted rounded-lg border border-border/40 bg-surface-2/40 px-3 py-2">
        {t("billing.cardNotConfigured")}
      </p>
    );
  }

  return (
    <div className="rounded-xl border border-border/40 bg-surface-2/50 px-4 py-3 text-sm space-y-1">
      <div className="text-xs font-medium text-fg-subtle">{t("billing.cardToCard")}</div>
      {settings.card_bank && <div className="text-fg">{settings.card_bank}</div>}
      <div className="font-mono text-base text-fg tracking-wide" dir="ltr">{settings.card_number}</div>
      {settings.card_holder && <div className="text-fg">{settings.card_holder}</div>}
      {settings.manual_instructions && (
        <p className="text-xs text-fg-muted pt-1">{settings.manual_instructions}</p>
      )}
    </div>
  );
}
