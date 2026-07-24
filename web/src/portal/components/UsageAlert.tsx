import { AlertTriangle, Info } from "lucide-react";

interface UsageAlertProps { usedBytes: number; limitBytes: number; }

function fmtB(b: number): string { if (!b) return "0 B"; const u=["B","KB","MB","GB","TB"]; const i=Math.floor(Math.log(b)/Math.log(1024)); return (b/Math.pow(1024,i)).toFixed(1)+" "+u[i]; }

export function UsageAlert({ usedBytes, limitBytes }: UsageAlertProps) {
  if (limitBytes <= 0) return null;
  const pct = (usedBytes / limitBytes) * 100;
  if (pct < 80) return null;
  const critical = pct >= 90;

  return (
    <div className={`flex items-center gap-3 p-3 rounded-lg ${critical ? "bg-danger/10 text-danger border border-danger/30" : "bg-warning/10 text-warning border border-warning/30"}`}>
      {critical ? <AlertTriangle className="w-5 h-5 flex-shrink-0"/> : <Info className="w-5 h-5 flex-shrink-0"/>}
      <div className="flex-1">
        <p className="text-sm font-medium">{critical ? "Traffic limit almost reached!" : "Traffic usage is high"}</p>
        <p className="text-xs mt-0.5">{fmtB(usedBytes)} / {fmtB(limitBytes)} ({pct.toFixed(0)}%)</p>
      </div>
    </div>
  );
}
