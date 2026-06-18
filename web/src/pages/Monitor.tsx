import { useQuery } from "@tanstack/react-query";
import { Activity, Globe, Users, Wifi } from "lucide-react";
import { api } from "@/api/client";
import { Card, PageHeader } from "@/components/ui";
import { useI18n } from "@/i18n/i18n";

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

export function Monitor() {
  const { t } = useI18n();
  const { data, isLoading, isError } = useLiveConnections();
  const connections = data?.connections ?? [];
  const total = data?.total ?? 0;

  return (
    <div className="space-y-6 animate-fade-in">
      <div className="flex items-center justify-between">
        <PageHeader title={t("monitor.title")} />
        <div className="flex items-center gap-2 rounded-full bg-success/10 px-3 py-1.5 text-xs font-medium text-success">
          <span className="relative flex h-2 w-2">
            <span className="absolute inline-flex h-full w-full animate-ping rounded-full bg-success/60" />
            <span className="inline-flex h-2 w-2 rounded-full bg-success" />
          </span>
          {total} {t("monitor.connections")}
        </div>
      </div>

      {/* Stats cards */}
      <div className="grid grid-cols-2 gap-4 md:grid-cols-4">
        <StatCard icon={<Users size={16} />} label={t("monitor.usersOnline")} value={new Set(connections.map(c => c.user_id)).size} />
        <StatCard icon={<Wifi size={16} />} label={t("monitor.connections")} value={total} />
        <StatCard icon={<Globe size={16} />} label={t("monitor.uniqueIPs")} value={new Set(connections.map(c => c.ip)).size} />
        <StatCard icon={<Activity size={16} />} label={t("monitor.nodesActive")} value={new Set(connections.map(c => c.node_name)).size} />
      </div>

      {/* Connection table */}
      <Card className="p-0 overflow-hidden">
        <table className="w-full text-sm">
          <thead className="border-b bg-surface-2/30 text-left text-xs text-fg-muted">
            <tr>
              <th className="px-4 py-3 font-medium">User</th>
              <th className="px-4 py-3 font-medium">Node</th>
              <th className="px-4 py-3 font-medium">IP</th>
              <th className="px-4 py-3 font-medium">Protocol</th>
              <th className="px-4 py-3 font-medium">Duration</th>
            </tr>
          </thead>
          <tbody className="divide-y divide-border/30">
            {connections.map((c, i) => (
              <tr key={i} className="hover:bg-surface-2/20 transition">
                <td className="px-4 py-2.5 font-medium">{c.username}</td>
                <td className="px-4 py-2.5 text-fg-muted">{c.node_name}</td>
                <td className="px-4 py-2.5 font-mono text-xs text-fg-subtle">{c.ip}</td>
                <td className="px-4 py-2.5"><span className="rounded bg-primary/10 px-1.5 py-0.5 text-[10px] font-medium text-primary uppercase">{c.protocol}</span></td>
                <td className="px-4 py-2.5 text-xs text-fg-subtle">{timeSince(c.connected_since)}</td>
              </tr>
            ))}
            {isLoading && <tr><td colSpan={5} className="px-4 py-8 text-center text-fg-muted">Loading...</td></tr>}
            {isError && <tr><td colSpan={5} className="px-4 py-8 text-center text-fg-muted">{t("monitor.apiError")}</td></tr>}
            {!isLoading && !isError && connections.length === 0 && <tr><td colSpan={5} className="px-4 py-8 text-center text-fg-muted">{t("monitor.noActive")}</td></tr>}
          </tbody>
        </table>
      </Card>
    </div>
  );
}

function StatCard({ icon, label, value }: { icon: React.ReactNode; label: string; value: number }) {
  return (
    <Card className="flex items-center gap-3 p-4">
      <div className="grid h-9 w-9 place-items-center rounded-xl bg-primary/10 text-primary">{icon}</div>
      <div>
        <div className="text-lg font-bold text-fg">{value}</div>
        <div className="text-[10px] text-fg-subtle uppercase">{label}</div>
      </div>
    </Card>
  );
}

function timeSince(iso: string): string {
  const sec = Math.floor((Date.now() - new Date(iso).getTime()) / 1000);
  if (sec < 60) return `${sec}s`;
  const min = Math.floor(sec / 60);
  if (min < 60) return `${min}m`;
  return `${Math.floor(min / 60)}h ${min % 60}m`;
}
