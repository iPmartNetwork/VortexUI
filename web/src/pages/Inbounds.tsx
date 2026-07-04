import { useMemo, useState } from "react";
import { useNavigate } from "react-router-dom";
import { Globe, Network, Plus, Server } from "lucide-react";
import { useInboundsFleet, useNodes, type Inbound, type InboundFleetRow } from "@/api/hooks";
import type { Node } from "@/api/types";
import { Button } from "@/components/ui";
import { NodeInboundsModal } from "@/components/NodeInboundsModal";
import { SubHostsModal } from "@/components/SubHostsModal";
import { GlassCard, ProtocolBadge, StatusBadge } from "@/components/veltrix";
import { useI18n } from "@/i18n/i18n";
import { useTitle } from "@/lib/useTitle";
import { useAuth } from "@/auth/auth";
import { cn } from "@/lib/utils";

function transportLabel(ib: InboundFleetRow): string {
  if (!ib.network && !ib.security) return "—";
  const net = ib.network || "tcp";
  const sec = ib.security || "none";
  return `${net}/${sec}`;
}

export function Inbounds() {
  useTitle("Inbounds");
  const { can } = useAuth();
  const canWrite = can("inbound:write");
  const { t } = useI18n();
  const navigate = useNavigate();
  const fleet = useInboundsFleet();
  const nodes = useNodes();
  const [nodeFilter, setNodeFilter] = useState("");
  const [managing, setManaging] = useState<Node | null>(null);
  const [subHostsFor, setSubHostsFor] = useState<Inbound | null>(null);

  const inbounds = fleet.data?.inbounds ?? [];
  const nodeList = nodes.data?.nodes ?? [];

  const filtered = useMemo(
    () => (nodeFilter ? inbounds.filter((ib) => ib.node_name === nodeFilter) : inbounds),
    [inbounds, nodeFilter],
  );

  const nodeNames = useMemo(
    () => [...new Set(inbounds.map((ib) => ib.node_name))].sort(),
    [inbounds],
  );

  function openNodeManager(nodeName: string) {
    const node = nodeList.find((n) => n.name === nodeName);
    if (node) setManaging(node);
  }

  return (
    <div className="space-y-5 animate-page-enter">
      <NodeInboundsModal node={managing} onClose={() => setManaging(null)} />
      <SubHostsModal inbound={subHostsFor} onClose={() => setSubHostsFor(null)} />

      <div className="flex flex-col sm:flex-row sm:items-start justify-between gap-4">
        <div>
          <h1 className="text-2xl font-bold text-fg tracking-tight">{t("nav.inboundsSubhosts")}</h1>
          <p className="text-sm text-fg-muted mt-1">
            {inbounds.length} {t("nodes.inboundsTotal")}
          </p>
        </div>
        {canWrite && nodeList.length > 0 && (
          <Button
            onClick={() => {
              const target = nodeFilter
                ? nodeList.find((n) => n.name === nodeFilter)
                : nodeList[0];
              if (target) setManaging(target);
            }}
          >
            <Plus size={15} /> {t("nodes.manageInbounds")}
          </Button>
        )}
      </div>

      <GlassCard hover={false} className="!p-0 overflow-hidden">
        <div className="flex flex-wrap items-center gap-2 p-4 border-b border-border/40">
          <button
            type="button"
            onClick={() => setNodeFilter("")}
            className={cn(
              "px-3.5 py-1.5 rounded-lg text-xs font-medium transition-all",
              !nodeFilter
                ? "bg-primary text-primary-fg shadow-sm"
                : "text-fg-muted hover:text-fg hover:bg-surface/60",
            )}
          >
            {t("users.filterAll")}
          </button>
          {nodeNames.map((name) => (
            <button
              key={name}
              type="button"
              onClick={() => setNodeFilter(name)}
              className={cn(
                "px-3.5 py-1.5 rounded-lg text-xs font-medium transition-all",
                nodeFilter === name
                  ? "bg-primary text-primary-fg shadow-sm"
                  : "text-fg-muted hover:text-fg hover:bg-surface/60",
              )}
            >
              {name}
            </button>
          ))}
        </div>

        {fleet.isLoading && (
          <div className="p-8 text-sm text-fg-muted text-center">{t("common.loading")}</div>
        )}

        {!fleet.isLoading && filtered.length === 0 && (
          <div className="flex flex-col items-center gap-3 py-16 text-center">
            <Globe size={32} className="text-fg-subtle" />
            <p className="text-sm text-fg-muted">{t("nodes.noInbounds")}</p>
            {canWrite && nodeList.length > 0 && (
              <Button variant="outline" size="sm" onClick={() => setManaging(nodeList[0])}>
                {t("nodes.manageInbounds")}
              </Button>
            )}
            {nodeList.length === 0 && (
              <Button variant="outline" size="sm" onClick={() => navigate("/nodes")}>
                {t("nodes.new")}
              </Button>
            )}
          </div>
        )}

        {!fleet.isLoading && filtered.length > 0 && (
          <div className="overflow-x-auto">
            <table className="w-full text-sm">
              <thead>
                <tr className="border-b border-border/40 bg-surface/30">
                  <th className="text-start py-3 px-4 text-xs font-semibold text-fg-subtle uppercase tracking-wide">
                    {t("nodes.title")}
                  </th>
                  <th className="text-start py-3 px-4 text-xs font-semibold text-fg-subtle uppercase tracking-wide">
                    Tag
                  </th>
                  <th className="text-start py-3 px-4 text-xs font-semibold text-fg-subtle uppercase tracking-wide hidden md:table-cell">
                    {t("users.protocol")}
                  </th>
                  <th className="text-start py-3 px-4 text-xs font-semibold text-fg-subtle uppercase tracking-wide hidden sm:table-cell">
                    Port
                  </th>
                  <th className="text-start py-3 px-4 text-xs font-semibold text-fg-subtle uppercase tracking-wide hidden lg:table-cell">
                    Transport
                  </th>
                  <th className="text-start py-3 px-4 text-xs font-semibold text-fg-subtle uppercase tracking-wide">
                    {t("common.status")}
                  </th>
                  <th className="text-end py-3 px-4 text-xs font-semibold text-fg-subtle uppercase tracking-wide">
                    {t("common.actions")}
                  </th>
                </tr>
              </thead>
              <tbody>
                {filtered.map((ib) => (
                  <tr
                    key={ib.id}
                    className="border-b border-border/20 hover:bg-surface/40 transition-colors"
                  >
                    <td className="py-3.5 px-4">
                      <div className="flex items-center gap-2.5 min-w-[120px]">
                        <div className="h-8 w-8 rounded-full bg-primary/10 flex items-center justify-center text-primary flex-shrink-0">
                          <Server size={14} />
                        </div>
                        <span className="font-medium text-fg text-sm">{ib.node_name}</span>
                      </div>
                    </td>
                    <td className="py-3.5 px-4 font-mono text-xs text-fg">{ib.tag}</td>
                    <td className="py-3.5 px-4 hidden md:table-cell">
                      <ProtocolBadge label={ib.protocol} />
                    </td>
                    <td className="py-3.5 px-4 hidden sm:table-cell tabular-nums text-fg-muted">
                      {ib.port}
                    </td>
                    <td className="py-3.5 px-4 hidden lg:table-cell text-xs text-fg-muted font-mono">
                      {transportLabel(ib)}
                    </td>
                    <td className="py-3.5 px-4">
                      <StatusBadge
                        status={ib.enabled ? "active" : "inactive"}
                        label={ib.enabled ? "ENABLED" : "DISABLED"}
                        pulse={false}
                      />
                    </td>
                    <td className="py-3.5 px-4 text-end">
                      <div className="flex items-center justify-end gap-1">
                        <Button variant="ghost" size="sm" onClick={() => setSubHostsFor(ib)}>
                          SubHosts
                        </Button>
                        {canWrite && (
                          <Button variant="ghost" size="sm" onClick={() => openNodeManager(ib.node_name)}>
                            {t("common.edit")}
                          </Button>
                        )}
                      </div>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}

        {!fleet.isLoading && filtered.length > 0 && (
          <div className="border-t border-border/40 px-4 py-3 text-sm text-fg-muted flex items-center gap-2">
            <Network size={14} className="text-fg-subtle" />
            {t("users.showingOf")
              .replace("{count}", String(filtered.length))
              .replace("{total}", String(inbounds.length))}
          </div>
        )}
      </GlassCard>
    </div>
  );
}
