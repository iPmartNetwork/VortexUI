import { useQuery } from "@tanstack/react-query";
import { QrCode, RefreshCw } from "lucide-react";
import { api } from "@/api/client";
import { Button } from "@/components/ui";
import { useI18n } from "@/i18n/i18n";

interface DynamicQRCodeProps {
  subscriptionUrl: string;
  refreshInterval?: number;
}

export function DynamicQRCode({ subscriptionUrl, refreshInterval = 30000 }: DynamicQRCodeProps) {
  const { t } = useI18n();

  const { data: qrUrl, refetch, isLoading } = useQuery({
    queryKey: ["dynamic-qr", subscriptionUrl],
    queryFn: async () => {
      // Generate QR code from subscription URL
      // Uses a QR generation endpoint or client-side library
      const encoded = encodeURIComponent(subscriptionUrl);
      return `https://api.qrserver.com/v1/create-qr-code/?size=256x256&data=${encoded}`;
    },
    refetchInterval: refreshInterval,
  });

  return (
    <div className="flex flex-col items-center gap-3">
      <div className="relative">
        {isLoading ? (
          <div className="w-64 h-64 flex items-center justify-center bg-muted rounded">
            <RefreshCw className="w-8 h-8 animate-spin text-muted-foreground" />
          </div>
        ) : qrUrl ? (
          <img src={qrUrl} alt="Subscription QR Code" className="w-64 h-64 rounded" />
        ) : (
          <div className="w-64 h-64 flex items-center justify-center bg-muted rounded">
            <QrCode className="w-16 h-16 text-muted-foreground" />
          </div>
        )}
      </div>
      <Button variant="outline" size="sm" onClick={() => refetch()}>
        <RefreshCw className="w-3 h-3 mr-1" />
        {t("portal.refreshQR", "Refresh QR")}
      </Button>
      <p className="text-xs text-muted-foreground text-center">
        {t("portal.qrAutoRefresh", "QR code refreshes automatically to stay up-to-date.")}
      </p>
    </div>
  );
}
