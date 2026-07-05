import { useQuery } from "@tanstack/react-query";
import { Activity, Globe, Radio, Users, Wifi } from "lucide-react";
import { api } from "@/api/client";
import { Badge } from "@/components/ui";
import { GlassCard, StatsCard } from "@/components/veltrix";
import { useI18n } from "@/i18n/i18n";
import { useTitle } from "@/lib/useTitle";

interface LiveConnection {
  user_id: string;
  username: string;
  node_name: string;
  ip: string;
  protocol: string;
  connected_since: string;
}

function useLiveConnections() {
  return useQuery({
    queryKey: ["live-connections"],
    queryFn: () => api<{ connections: LiveConnection[]; total: number }>("/api/monitor/connections"),
    refetchInterval: 3000,
  });
}

function timeSince(iso: string): string {
  const sec = Math.floor((Date.now() - new Date(iso).getTime()) / 1000);
  if (sec < 60) return `${sec}s`;
  const min = Math.floor(sec / 60);
  if (min < 60) return `${min}m`;
  return `${Math.floor(min / 60)}h ${min % 60}m`;
}

export function Monitor() {
  useTitle("Monitor");
  const { t } = useI18n();
  const { data, isLoading, isError } = useLiveConnections();
  const connections = data?.connections ?? [];
  const total = data?.total ?? 0;

  return (
    <div className="space-y-5 animate-page-enter">
      <div className="flex flex-col lg:flex-row lg:items-start justify-between gap-4">
        <div>
          <div className="flex flex-wrap items-center gap-2">
            <h1 className="text-2xl font-bold text-fg tracking-tight">{t("monitor.title")}</h1>
            <span className="inline-flex items-center gap-1.5 rounded-full bg-success/10 px-2.5 py-1 text-[11px] font-semibold text-success">
              <span className="relative flex h-1.5 w-1.5">
                <span className="absolute inline-flex h-full w-full animate-ping rounded-full bg-success/60" />
                <span className="inline-flex h-1.5 w-1.5 rounded-full bg-success" />
              </span>
              {total} {t("monitor.connections")}
            </span>
          </div>
          <p className="text-sm text-fg-muted mt-1 max-w-2xl">{t("monitor.subtitle")}</p>
        </div>
      </div>

      <div className="grid grid-cols-2 gap-4 md:grid-cols-4">
        <StatsCard
          title={t("monitor.usersOnline")}
          value={new Set(connections.map((c) => c.user_id)).size}
          icon={<Users size={18} />}
          color="blue"
        />
        <StatsCard
          title={t("monitor.connections")}
          value={total}
          icon={<Wifi size={18} />}
          color="cyan"
        />
        <StatsCard
          title={t("monitor.uniqueIPs")}
          value={new Set(connections.map((c) => c.ip)).size}
          icon={<Globe size={18} />}
          color="purple"
        />
        <StatsCard
          title={t("monitor.nodesActive")}
          value={new Set(connections.map((c) => c.node_name)).size}
          icon={<Activity size={18} />}
          color="green"
        />
      </div>

      <GlassCard hover={false} className="!p-0 overflow-hidden">
        <div className="overflow-x-auto">
          <table className="w-full text-sm">
            <thead>
              <tr className="border-b border-border/40 text-[11px] uppercase tracking-wide text-fg-subtle bg-surface-2/30">
                <th className="py-3 px-4 text-left">{t("monitor.colUser")}</th>
                <th className="py-3 px-4 text-left">{t("monitor.colNode")}</th>
                <th className="py-3 px-4 text-left">{t("monitor.colIP")}</th>
                <th className="py-3 px-4 text-center">{t("monitor.colProtocol")}</th>
                <th className="py-3 px-4 text-right">{t("monitor.colDuration")}</th>
              </tr>
            </thead>
            <tbody>
              {connections.map((c, i) => (
                <tr key={`${c.user_id}-${c.ip}-${i}`} className="border-b border-border/20 hover:bg-surface-2/40">
                  <td className="py-3 px-4 font-medium text-fg">{c.username}</td>
                  <td className="py-3 px-4 text-fg-muted flex items-center gap-1.5">
                    <Radio size={12} className="text-success flex-shrink-0" />
                    {c.node_name}
                  </td>
                  <td className="py-3 px-4 font-mono text-xs text-fg-subtle" dir="ltr">{c.ip}</td>
                  <td className="py-3 px-4 text-center">
                    <Badge color="muted">{c.protocol}</Badge>
                  </td>
                  <td className="py-3 px-4 text-right text-xs tabular-nums text-fg-muted">{timeSince(c.connected_since)}</td>
                </tr>
              ))}
              {isLoading && (
                <tr>
                  <td colSpan={5} className="py-8 px-4 text-center text-fg-muted">…</td>
                </tr>
              )}
              {isError && (
                <tr>
                  <td colSpan={5} className="py-8 px-4 text-center text-fg-muted">{t("monitor.apiError")}</td>
                </tr>
              )}
              {!isLoading && !isError && connections.length === 0 && (
                <tr>
                  <td colSpan={5} className="py-8 px-4 text-center text-fg-muted">{t("monitor.noActive")}</td>
                </tr>
              )}
            </tbody>
          </table>
        </div>
      </GlassCard>
    </div>
  );
}
