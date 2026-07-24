import { useState } from "react";
import { RefreshCw } from "lucide-react";
import { Button } from "@/components/ui";

interface DynamicQRCodeProps { subscriptionUrl: string; }

export function DynamicQRCode({ subscriptionUrl }: DynamicQRCodeProps) {
  const [key, setKey] = useState(0);
  const qrUrl = `https://api.qrserver.com/v1/create-qr-code/?size=256x256&data=${encodeURIComponent(subscriptionUrl)}&t=${key}`;

  return (
    <div className="flex flex-col items-center gap-3">
      <img src={qrUrl} alt="QR Code" className="w-64 h-64 rounded" />
      <Button variant="outline" size="sm" onClick={() => setKey((k) => k + 1)}>
        <RefreshCw className="w-3 h-3 mr-1" />Refresh QR
      </Button>
    </div>
  );
}
