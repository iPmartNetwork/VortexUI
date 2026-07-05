import { useParams, useNavigate } from "react-router-dom";
import { useQuery } from "@tanstack/react-query";
import { ArrowLeft, RotateCcw, KeyRound, Wifi, ShieldAlert, Globe, Database, CalendarClock, Smartphone } from "lucide-react";
import { api } from "@/api/client";
import { useUserUsage, useRoutingPacks, useUserRoutingPack, useSetUserRoutingPack } from "@/api/hooks";
import { useResetUser, useRevokeSub, useUserSub, useUserOnline, useUserOnlineIPs } from "@/api/policy-hooks";
import type { User } from "@/api/types";
import { Badge, Button, Select } from "@/components/ui";
import { GlassCard, StatsCard } from "@/components/veltrix";
import { UsageChart } from "@/components/UsageChart";
import { CopyField } from "@/components/CopyField";
import { useConfirm } from "@/components/confirm";
import { useToast } from "@/components/toast";
import { useAuth } from "@/auth/auth";
import { useTitle } from "@/lib/useTitle";
import { formatBytes } from "@/lib/utils";
import { useI18n } from "@/i18n/i18n";
import { QRCodeSVG } from "qrcode.react";

export function UserDetail() {
  const { t } = useI18n();
  useTitle(t("userDetail.title"));
  const { can } = useAuth();
  const canWrite = can("user:write");
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const toast = useToast();
  const confirm = useConfirm();

  const { data: detail, isLoading } = useQuery({
    queryKey: ["user", id],
    queryFn: () => api<{ user: User; inbound_ids: string[] }>(`/api/users/${id}`),
    enabled: !!id,
  });
  const user = detail?.user;
  const usage = useUserUsage(id ?? null);
  const sub = useUserSub(id ?? null);
  const online = useUserOnline(id ?? null);
  const onlineIPs = useUserOnlineIPs(id ?? null);
  const reset = useResetUser();
  const revoke = useRevokeSub();

  if (isLoading || !user) {
    return <div className="py-20 text-center text-fg-muted">{t("common.loading")}</div>;
  }

  const on = online.data;
  const d = sub.data;

  async function doReset() {
    if (await confirm({ title: `${t("userDetail.resetConfirm")} ${user!.username}?`, confirmLabel: t("userDetail.reset") })) {
      await reset.mutateAsync(user!.id);
      toast.success(t("userDetail.usageReset"));
    }
  }
  async function doRevoke() {
    if (await confirm({
      title: t("userDetail.revokeConfirm"),
      message: t("userDetail.revokeMessage"),
      confirmLabel: t("userDetail.revoke"),
      destructive: true,
    })) {
      await revoke.mutateAsync(user!.id);
      toast.success(t("userDetail.linkRevoked"));
    }
  }

  return (
    <div className="space-y-6 animate-page-enter">
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
                {on.live_connections} {t("userDetail.live")}
              </span>
            )}
          </div>
        </div>
        {canWrite && (
          <div className="flex gap-2">
            <Button variant="outline" size="sm" onClick={doReset}><RotateCcw size={14} /> {t("userDetail.reset")}</Button>
            <Button variant="outline" size="sm" className="text-danger" onClick={doRevoke}><KeyRound size={14} /> {t("userDetail.revoke")}</Button>
          </div>
        )}
      </div>

      <div className="grid grid-cols-2 gap-4 md:grid-cols-4">
        <StatsCard title={t("userDetail.used")} value={formatBytes(user.used_traffic, false)} icon={<Database size={18} />} color="blue" />
        <StatsCard title={t("userDetail.limit")} value={formatBytes(user.data_limit)} icon={<Database size={18} />} color="purple" />
        <StatsCard title={t("userDetail.expires")} value={user.expire_at ? new Date(user.expire_at).toLocaleDateString() : t("common.never")} icon={<CalendarClock size={18} />} color="orange" />
        <StatsCard title={t("userDetail.devices")} value={`${on?.active_devices ?? 0} / ${user.device_limit || "∞"}`} icon={<Smartphone size={18} />} color="cyan" />
      </div>

      <GlassCard hover={false} className="!p-5">
        <h3 className="mb-3 text-xs font-semibold uppercase tracking-wider text-fg-subtle">{t("userDetail.traffic7d")}</h3>
        {usage.isLoading && <div className="h-40 animate-pulse rounded-lg bg-surface-2/50" />}
        {usage.data && <UsageChart points={usage.data.points} />}
      </GlassCard>

      {onlineIPs.data?.tracking && (
        <GlassCard hover={false} className="!p-5">
          <div className="mb-3 flex items-center justify-between">
            <h3 className="text-xs font-semibold uppercase tracking-wider text-fg-subtle">{t("userDetail.onlineIPs")}</h3>
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
            <p className="py-4 text-center text-sm text-fg-muted">{t("userDetail.noConnections")}</p>
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
        </GlassCard>
      )}

      <RoutingPackSelector userId={id ?? null} />

      {d && (
        <GlassCard hover={false} className="!p-5 space-y-4">
          <h3 className="text-xs font-semibold uppercase tracking-wider text-fg-subtle">{t("userDetail.subscription")}</h3>
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
              <p className="mb-1.5 text-xs font-medium text-fg-muted">{t("userDetail.configs")} ({d.links.length})</p>
              <div className="max-h-32 space-y-1.5 overflow-auto">
                {d.links.map((l, i) => <CopyField key={i} value={l} />)}
              </div>
            </div>
          )}
        </GlassCard>
      )}
    </div>
  );
}

function RoutingPackSelector({ userId }: { userId: string | null }) {
  const { t } = useI18n();
  const packs = useRoutingPacks();
  const current = useUserRoutingPack(userId);
  const setPack = useSetUserRoutingPack(userId);
  const toast = useToast();

  const value = current.data?.pack_id ?? "";

  async function change(packId: string) {
    await setPack.mutateAsync(packId);
    toast.success(t("userDetail.routingPackUpdated"));
  }

  return (
    <GlassCard hover={false} className="!p-5 space-y-3">
      <h3 className="text-xs font-semibold uppercase tracking-wider text-fg-subtle">{t("userDetail.routingPack")}</h3>
      <p className="text-xs text-fg-muted">{t("userDetail.routingPackHint")}</p>
      <Select
        className="w-full sm:w-72"
        value={value}
        disabled={current.isLoading || setPack.isPending}
        onChange={(e) => change(e.target.value)}
      >
        <option value="">— {t("common.none")} —</option>
        {(packs.data?.packs ?? []).map((p) => (
          <option key={p.id} value={p.id}>
            {p.name}{p.builtin ? ` (${t("userDetail.builtin")})` : ""}
          </option>
        ))}
      </Select>
    </GlassCard>
  );
}
