import { AlertTriangle, Info } from "lucide-react";
import { useI18n } from "@/i18n/i18n";

interface UsageAlertProps {
  usedBytes: number;
  limitBytes: number;
}

function formatBytes(bytes: number): string {
  if (bytes === 0) return "0 B";
  const units = ["B", "KB", "MB", "GB", "TB"];
  const i = Math.floor(Math.log(bytes) / Math.log(1024));
  return (bytes / Math.pow(1024, i)).toFixed(1) + " " + units[i];
}

export function UsageAlert({ usedBytes, limitBytes }: UsageAlertProps) {
  const { t } = useI18n();

  if (limitBytes <= 0) return null;

  const percentage = (usedBytes / limitBytes) * 100;
  const isWarning = percentage >= 80;
  const isCritical = percentage >= 90;

  if (percentage < 80) return null;

  return (
    <div
      className={`flex items-center gap-3 p-3 rounded-lg ${
        isCritical
          ? "bg-destructive/10 text-destructive border border-destructive/30"
          : isWarning
          ? "bg-yellow-500/10 text-yellow-700 dark:text-yellow-400 border border-yellow-500/30"
          : "bg-blue-500/10 text-blue-700 dark:text-blue-400 border border-blue-500/30"
      }`}
    >
      {isCritical ? (
        <AlertTriangle className="w-5 h-5 flex-shrink-0" />
      ) : (
        <Info className="w-5 h-5 flex-shrink-0" />
      )}
      <div className="flex-1">
        <p className="text-sm font-medium">
          {isCritical
            ? t("portal.usageCritical", "Traffic limit almost reached!")
            : t("portal.usageWarning", "Traffic usage is high")}
        </p>
        <p className="text-xs mt-0.5">
          {formatBytes(usedBytes)} / {formatBytes(limitBytes)} ({percentage.toFixed(0)}%)
        </p>
      </div>
    </div>
  );
}
