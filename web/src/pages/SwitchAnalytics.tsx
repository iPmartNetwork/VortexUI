import { useState } from "react";
import { ArrowRightLeft, Globe, Server, TrendingUp } from "lucide-react";
import { GlassCard, StatsCard } from "@/components/veltrix";
import { useI18n } from "@/i18n/i18n";
import { useTitle } from "@/lib/useTitle";
import { EmptyState } from "@/components/EmptyState";
import { useSwitchSummary } from "@/api/protocol-group-hooks";
import { useNodes } from "@/api/hooks";
import { Input, Select } from "@/components/ui";

export function SwitchAnalytics() {
  useTitle("Switch Analytics");
  const { t } = useI18n();
  const nodesQ = useNodes();
  const nodes = nodesQ.data?.nodes ?? [];

  const [nodeId, setNodeId] = useState("");
  const [isp, setIsp] = useState("");
  const [from, setFrom] = useState("");
  const [to, setTo] = useState("");

  const params: Record<string, string> = {};
  if (nodeId) params.node_id = nodeId;
  if (isp) params.isp = isp;
  if (from) params.from = from;
  if (to) params.to = to;

  const { data, isLoading } = useSwitchSummary(params);

  return (
    <div className="space-y-6">
      {/* Header */}
      <div>
        <h1 className="text-lg font-bold text-fg">{t("switchAnalytics.title")}</h1>
        <p className="text-sm text-fg-muted mt-0.5">{t("switchAnalytics.subtitle")}</p>
      </div>

      {/* Filters */}
      <GlassCard hover={false} className="!p-4">
        <div className="grid grid-cols-1 gap-3 sm:grid-cols-2 lg:grid-cols-4">
          <div>
            <label className="text-[10px] font-medium text-fg-subtle uppercase tracking-wider mb-1 block">
              {t("switchAnalytics.filterNode")}
            </label>
            <Select value={nodeId} onChange={(e) => setNodeId(e.target.value)}>
              <option value="">{t("switchAnalytics.allNodes")}</option>
              {nodes.map((n) => (
                <option key={n.id} value={n.id}>
                  {n.name}
                </option>
              ))}
            </Select>
          </div>
          <div>
            <label className="text-[10px] font-medium text-fg-subtle uppercase tracking-wider mb-1 block">
              {t("switchAnalytics.filterISP")}
            </label>
            <Input
              value={isp}
              onChange={(e) => setIsp(e.target.value)}
              placeholder={t("switchAnalytics.allISPs")}
            />
          </div>
          <div>
            <label className="text-[10px] font-medium text-fg-subtle uppercase tracking-wider mb-1 block">
              {t("switchAnalytics.filterFrom")}
            </label>
            <Input type="date" value={from} onChange={(e) => setFrom(e.target.value)} />
          </div>
          <div>
            <label className="text-[10px] font-medium text-fg-subtle uppercase tracking-wider mb-1 block">
              {t("switchAnalytics.filterTo")}
            </label>
            <Input type="date" value={to} onChange={(e) => setTo(e.target.value)} />
          </div>
        </div>
      </GlassCard>

      {/* Stats Cards */}
      {data && (
        <div className="grid grid-cols-2 gap-3 lg:grid-cols-4">
          <StatsCard
            title={t("switchAnalytics.totalSwitches")}
            value={data.total_switches}
            icon={<TrendingUp size={18} />}
            color="blue"
          />
          <StatsCard
            title={t("switchAnalytics.byProtocol")}
            value={Object.keys(data.by_protocol ?? {}).length}
            icon={<ArrowRightLeft size={18} />}
            color="orange"
          />
          <StatsCard
            title={t("switchAnalytics.byISP")}
            value={Object.keys(data.by_isp ?? {}).length}
            icon={<Globe size={18} />}
            color="green"
          />
          <StatsCard
            title={t("switchAnalytics.byNode")}
            value={Object.keys(data.by_node ?? {}).length}
            icon={<Server size={18} />}
            color="cyan"
          />
        </div>
      )}

      {/* No data state */}
      {!isLoading && (!data || data.total_switches === 0) && (
        <EmptyState icon={ArrowRightLeft} title={t("switchAnalytics.noData")} />
      )}

      {/* Breakdown */}
      {data && data.total_switches > 0 && (
        <div className="grid grid-cols-1 gap-4 lg:grid-cols-2">
          {/* By Protocol */}
          <GlassCard hover={false} className="!p-5">
            <h3 className="text-xs font-semibold uppercase tracking-wider text-fg-subtle mb-3">
              {t("switchAnalytics.byProtocol")}
            </h3>
            <div className="space-y-2">
              {Object.entries(data.by_protocol ?? {})
                .sort(([, a], [, b]) => b - a)
                .map(([proto, count]) => (
                  <div key={proto} className="flex items-center justify-between">
                    <span className="text-sm text-fg-muted">{proto}</span>
                    <div className="flex items-center gap-2">
                      <div className="w-20 h-1.5 rounded-full bg-surface-3 overflow-hidden">
                        <div
                          className="h-full rounded-full grad-bg"
                          style={{
                            width: `${Math.min(100, (count / data.total_switches) * 100)}%`,
                          }}
                        />
                      </div>
                      <span className="text-xs font-semibold text-fg tabular-nums w-8 text-end">
                        {count}
                      </span>
                    </div>
                  </div>
                ))}
            </div>
          </GlassCard>

          {/* By ISP */}
          <GlassCard hover={false} className="!p-5">
            <h3 className="text-xs font-semibold uppercase tracking-wider text-fg-subtle mb-3">
              {t("switchAnalytics.byISP")}
            </h3>
            <div className="space-y-2">
              {Object.entries(data.by_isp ?? {})
                .sort(([, a], [, b]) => b - a)
                .map(([ispName, count]) => (
                  <div key={ispName} className="flex items-center justify-between">
                    <span className="text-sm text-fg-muted">{ispName}</span>
                    <div className="flex items-center gap-2">
                      <div className="w-20 h-1.5 rounded-full bg-surface-3 overflow-hidden">
                        <div
                          className="h-full rounded-full grad-bg"
                          style={{
                            width: `${Math.min(100, (count / data.total_switches) * 100)}%`,
                          }}
                        />
                      </div>
                      <span className="text-xs font-semibold text-fg tabular-nums w-8 text-end">
                        {count}
                      </span>
                    </div>
                  </div>
                ))}
            </div>
          </GlassCard>

          {/* By Node */}
          <GlassCard hover={false} className="!p-5">
            <h3 className="text-xs font-semibold uppercase tracking-wider text-fg-subtle mb-3">
              {t("switchAnalytics.byNode")}
            </h3>
            <div className="space-y-2">
              {Object.entries(data.by_node ?? {})
                .sort(([, a], [, b]) => b - a)
                .map(([nodeKey, count]) => {
                  const nodeName = nodes.find((n) => n.id === nodeKey)?.name ?? nodeKey;
                  return (
                    <div key={nodeKey} className="flex items-center justify-between">
                      <span className="text-sm text-fg-muted">{nodeName}</span>
                      <span className="text-xs font-semibold text-fg tabular-nums">{count}</span>
                    </div>
                  );
                })}
            </div>
          </GlassCard>

          {/* Top Switch Pairs */}
          {data.top_switch_pairs && data.top_switch_pairs.length > 0 && (
            <GlassCard hover={false} className="!p-5">
              <h3 className="text-xs font-semibold uppercase tracking-wider text-fg-subtle mb-3">
                {t("switchAnalytics.topPairs")}
              </h3>
              <div className="overflow-x-auto">
                <table className="w-full text-sm">
                  <thead>
                    <tr className="text-[10px] uppercase tracking-wider text-fg-subtle border-b border-border/40">
                      <th className="text-start py-2 font-medium">{t("switchAnalytics.source")}</th>
                      <th className="text-start py-2 font-medium">{t("switchAnalytics.target")}</th>
                      <th className="text-end py-2 font-medium">{t("switchAnalytics.count")}</th>
                    </tr>
                  </thead>
                  <tbody>
                    {data.top_switch_pairs.slice(0, 10).map((pair, i) => (
                      <tr key={i} className="border-b border-border/20">
                        <td className="py-2 text-fg-muted">{pair.source}</td>
                        <td className="py-2 text-fg-muted">{pair.target}</td>
                        <td className="py-2 text-end font-semibold text-fg tabular-nums">{pair.count}</td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            </GlassCard>
          )}
        </div>
      )}
    </div>
  );
}
