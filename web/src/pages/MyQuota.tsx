import { Gauge, Users, HardDrive } from "lucide-react";
import { useAccountQuota } from "@/api/quota-hooks";
import { useAuth } from "@/auth/auth";
import { Card, PageHeader } from "@/components/ui";
import { formatBytes, pct } from "@/lib/utils";

function QuotaBar({ label, used, limit, format = "number" }: { label: string; used: number; limit: number; format?: "number" | "bytes" }) {
  const unlimited = limit <= 0;
  const displayUsed = format === "bytes" ? formatBytes(used, false) : String(used);
  const displayLimit = unlimited ? "∞" : format === "bytes" ? formatBytes(limit, false) : String(limit);
  const displayRem = unlimited ? "∞" : format === "bytes" ? formatBytes(Math.max(0, limit - used), false) : String(Math.max(0, limit - used));
  const p = unlimited ? 0 : pct(used, limit);
  return (
    <Card className="space-y-3 p-5">
      <div className="flex items-center justify-between text-sm">
        <span className="font-medium">{label}</span>
        <span className="text-muted-foreground">{displayUsed} / {displayLimit}</span>
      </div>
      {!unlimited && (
        <div className="h-2 rounded-full bg-muted">
          <div className="h-full rounded-full bg-primary transition-all" style={{ width: `${p}%` }} />
        </div>
      )}
      <p className="text-xs text-muted-foreground">Remaining: {displayRem}</p>
    </Card>
  );
}

export function MyQuota() {
  const { session, sudo } = useAuth();
  const { data, isLoading } = useAccountQuota();
  const admin = session?.admin;

  if (sudo) {
    return (
      <div className="space-y-4">
        <PageHeader title="My quota" subtitle="Sudo admins have no reseller limits" />
        <Card className="p-6 text-sm text-muted-foreground">Full access — no account or traffic pool limits apply.</Card>
      </div>
    );
  }

  const u = data?.usage;

  return (
    <div className="space-y-6">
      <PageHeader title="My quota" subtitle={admin?.username ? `Reseller · ${admin.username}` : "Reseller limits"} />

      {isLoading && <p className="text-sm text-muted-foreground">Loading…</p>}

      {u && (
        <>
          <div className="grid gap-4 md:grid-cols-3">
            <Card className="flex items-center gap-3 p-5">
              <Users className="text-primary" size={22} />
              <div>
                <div className="text-xs text-muted-foreground">Accounts</div>
                <div className="text-xl font-bold">{u.user_count}{u.user_quota > 0 ? ` / ${u.user_quota}` : ""}</div>
              </div>
            </Card>
            <Card className="flex items-center gap-3 p-5">
              <HardDrive className="text-accent" size={22} />
              <div>
                <div className="text-xs text-muted-foreground">Traffic assigned</div>
                <div className="text-xl font-bold">{formatBytes(u.traffic_allocated, false)}</div>
              </div>
            </Card>
            <Card className="flex items-center gap-3 p-5">
              <Gauge className="text-success" size={22} />
              <div>
                <div className="text-xs text-muted-foreground">Traffic consumed</div>
                <div className="text-xl font-bold">{formatBytes(u.traffic_used, false)}</div>
              </div>
            </Card>
          </div>

          <div className="grid gap-4 md:grid-cols-2">
            <QuotaBar label="User accounts" used={u.user_count} limit={u.user_quota} />
            <QuotaBar label="Traffic pool (assigned to users)" used={u.traffic_allocated} limit={u.traffic_quota} format="bytes" />
          </div>

          <Card className="p-4 text-xs text-muted-foreground">
            Traffic pool is the total data limit you can assign across your users. Consumed traffic is what your users have actually used.
          </Card>
        </>
      )}
    </div>
  );
}
