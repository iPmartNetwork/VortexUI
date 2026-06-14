import { useParams, useNavigate } from "react-router-dom";
import { useQuery } from "@tanstack/react-query";
import { ArrowLeft, RotateCcw, KeyRound, Wifi, ShieldAlert, Globe } from "lucide-react";
import { api } from "@/api/client";
import { useUserUsage } from "@/api/hooks";
import { useResetUser, useRevokeSub, useUserSub, useUserOnline, useUserOnlineIPs } from "@/api/policy-hooks";
import type { User } from "@/api/types";
import { Badge, Button, Card } from "@/components/ui";
import { UsageChart } from "@/components/UsageChart";
import { CopyField } from "@/components/CopyField";
import { useConfirm } from "@/components/confirm";
import { useToast } from "@/components/toast";
import { formatBytes } from "@/lib/utils";
import { QRCodeSVG } from "qrcode.react";

export function UserDetail() {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const toast = useToast();
  const confirm = useConfirm();

  const { data: user, isLoading } = useQuery({
    queryKey: ["user", id],
    queryFn: () => api<User>(`/api/users/${id}`),
    enabled: !!id,
  });
  const usage = useUserUsage(id ?? null);
  const sub = useUserSub(id ?? null);
  const online = useUserOnline(id ?? null);
  const onlineIPs = useUserOnlineIPs(id ?? null);
  const reset = useResetUser();
  const revoke = useRevokeSub();

  if (isLoading || !user) {
    return <div className="py-20 text-center text-fg-muted">Loading…</div>;
  }

  const on = online.data;
  const d = sub.data;

  async function doReset() {
    if (await confirm({ title: `Reset usage for ${user!.username}?`, confirmLabel: "Reset" })) {
      await reset.mutateAsync(user!.id);
      toast.success("Usage reset");
    }
  }
  async function doRevoke() {
    if (await confirm({ title: `Revoke subscription link?`, message: "Old links stop working immediately.", confirmLabel: "Revoke", destructive: true })) {
      await revoke.mutateAsync(user!.id);
      toast.success("Link revoked");
    }
  }

  return (
    <div className="space-y-6 animate-fade-in">
      {/* Header */}
      <div className="flex items-center gap-4">
        <button onClick={() => navigate("/users")} className="grid h-9 w-9 place-items-center rounded-xl text-fg-muted transition hover:bg-surface-2/60 hover:text-fg">
          <ArrowLeft size={18} />
        </button>
        <div className="flex-1">
          <h1 className="text-xl font-bold text-fg">{user.username}</h1>
          <div className="mt-0.5 flex items-center gap-2">
            <Badge color={user.status}>{user.status}</Badge>
            {on?.live_tracking && (
              <span className="flex items-center gap-1 text-xs text-fg-muted">
                <Wifi size={12} className={on.live_connections > 0 ? "text-success" : "text-fg-subtle"} />
                {on.live_connections} live
              </span>
            )}
          </div>
        </div>
        <div className="flex gap-2">
          <Button variant="outline" size="sm" onClick={doReset}><RotateCcw size={14} /> Reset</Button>
          <Button variant="outline" size="sm" className="text-danger" onClick={doRevoke}><KeyRound size={14} /> Revoke</Button>
        </div>
      </div>

      {/* Stats row */}
      <div className="grid grid-cols-2 gap-4 md:grid-cols-4">
        <Card className="p-4 text-center">
          <div className="text-xs text-fg-muted">Used</div>
          <div className="mt-1 text-lg font-bold text-fg">{formatBytes(user.used_traffic, false)}</div>
        </Card>
        <Card className="p-4 text-center">
          <div className="text-xs text-fg-muted">Limit</div>
          <div className="mt-1 text-lg font-bold text-fg">{formatBytes(user.data_limit)}</div>
        </Card>
        <Card className="p-4 text-center">
          <div className="text-xs text-fg-muted">Expires</div>
          <div className="mt-1 text-sm font-bold text-fg">{user.expire_at ? new Date(user.expire_at).toLocaleDateString() : "Never"}</div>
        </Card>
        <Card className="p-4 text-center">
          <div className="text-xs text-fg-muted">Devices</div>
          <div className="mt-1 text-lg font-bold text-fg">{on?.active_devices ?? 0} / {user.device_limit || "∞"}</div>
        </Card>
      </div>

      {/* Chart */}
      <Card>
        <h3 className="mb-3 text-xs font-semibold uppercase tracking-wider text-fg-subtle">Traffic (7 days)</h3>
        {usage.isLoading && <div className="h-40 animate-pulse rounded-lg bg-surface-2/50" />}
        {usage.data && <UsageChart points={usage.data.points} />}
      </Card>

      {/* Online IPs — account-sharing detection */}
      {onlineIPs.data?.tracking && (
        <Card>
          <div className="mb-3 flex items-center justify-between">
            <h3 className="text-xs font-semibold uppercase tracking-wider text-fg-subtle">Online IPs</h3>
            {(() => {
              const count = onlineIPs.data.count;
              const limit = user.device_limit || 0;
              const over = limit > 0 && count > limit;
              return (
                <span className={`flex items-center gap-1 text-xs font-medium ${over ? "text-danger" : "text-fg-muted"}`}>
                  {over ? <ShieldAlert size={13} /> : <Globe size={13} />}
                  {count}{limit > 0 ? ` / ${limit}` : ""}
                </span>
              );
            })()}
          </div>
          {onlineIPs.data.count === 0 ? (
            <p className="py-4 text-center text-sm text-fg-muted">No active connections</p>
          ) : (
            <div className="space-y-1.5">
              {onlineIPs.data.ips.map((e) => (
                <div key={e.ip} className="flex items-center justify-between rounded-lg bg-surface-2/40 px-3 py-2 text-sm">
                  <span className="font-mono text-fg" dir="ltr">{e.ip}</span>
                  <span className="text-xs text-fg-subtle">{new Date(e.last_seen).toLocaleTimeString()}</span>
                </div>
              ))}
            </div>
          )}
        </Card>
      )}

      {/* Subscription */}
      {d && (
        <Card className="space-y-4">
          <h3 className="text-xs font-semibold uppercase tracking-wider text-fg-subtle">Subscription</h3>
          <div className="flex flex-col items-center gap-4 sm:flex-row">
            <div className="rounded-xl bg-white p-3">
              <QRCodeSVG value={d.subscription_url} size={120} />
            </div>
            <div className="flex-1 space-y-2">
              <CopyField value={d.subscription_url} />
              <div className="grid grid-cols-2 gap-2">
                {(["clash", "singbox", "base64"] as const).map((k) => (
                  <CopyField key={k} value={d.formats[k]} />
                ))}
              </div>
            </div>
          </div>
          {d.links.length > 0 && (
            <div>
              <p className="mb-1.5 text-xs font-medium text-fg-muted">Configs ({d.links.length})</p>
              <div className="max-h-32 space-y-1.5 overflow-auto">
                {d.links.map((l, i) => <CopyField key={i} value={l} />)}
              </div>
            </div>
          )}
        </Card>
      )}
    </div>
  );
}
