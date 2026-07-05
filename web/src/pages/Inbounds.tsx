import { useMemo, useState } from "react";
import { useNavigate } from "react-router-dom";
import { ChevronDown, Globe, Info, Network, Plus, Server, Settings } from "lucide-react";
import {
  useInboundsFleet,
  useNodes,
  useSubHosts,
  type HostSecurity,
  type Inbound,
  type InboundFleetRow,
  type SubHost,
} from "@/api/hooks";
import type { Node } from "@/api/types";
import { Badge, Button } from "@/components/ui";
import { NodeInboundsModal } from "@/components/NodeInboundsModal";
import { SubHostsModal } from "@/components/SubHostsModal";
import { GlassCard, ProtocolBadge, StatusBadge } from "@/components/veltrix";
import { useI18n } from "@/i18n/i18n";
import { useTitle } from "@/lib/useTitle";
import { useAuth } from "@/auth/auth";
import { cn } from "@/lib/utils";

function transportLabel(ib: InboundFleetRow): string {
  const net = ib.network || "tcp";
  return net.toUpperCase();
}

// securityColor mirrors SubHostsModal's mapping so REALITY/TLS/none read
// consistently everywhere a security mode is shown.
function securityColor(s?: string): string {
  switch (s) {
    case "reality":
      return "on_hold";
    case "tls":
      return "active";
    case "none":
    case "":
    case undefined:
      return "disabled";
    default:
      return "muted";
  }
}

function securityLabel(s?: string): string {
  if (!s || s === "none") return "none";
  if (s === "inbound_default") return "default";
  return s;
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
  const [pendingEdit, setPendingEdit] = useState<Inbound | null>(null);
  const [subHostsFor, setSubHostsFor] = useState<Inbound | null>(null);
  const [expandedId, setExpandedId] = useState<string | null>(null);

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

  function openNodeManager(nodeName: string, edit?: Inbound) {
    const node = nodeList.find((n) => n.name === nodeName);
    if (node) {
      setPendingEdit(edit ?? null);
      setManaging(node);
    }
  }

  return (
    <div className="space-y-5 animate-page-enter">
      <NodeInboundsModal
        node={managing}
        initialEdit={pendingEdit}
        onClose={() => {
          setManaging(null);
          setPendingEdit(null);
        }}
      />
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

      <GlassCard hover={false} className="!p-3.5 bg-primary/5 border-primary/20">
        <div className="flex items-start gap-2.5">
          <div className="h-7 w-7 rounded-lg bg-primary/15 flex items-center justify-center text-primary flex-shrink-0">
            <Info size={14} />
          </div>
          <div className="min-w-0">
            <p className="text-xs font-semibold text-fg">{t("hosts.title")}</p>
            <p className="text-xs text-fg-muted mt-0.5 leading-relaxed">{t("hosts.help")}</p>
          </div>
        </div>
      </GlassCard>

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
          <div className="divide-y divide-border/30">
            {filtered.map((ib) => (
              <InboundRow
                key={ib.id}
                ib={ib}
                showNode={!nodeFilter}
                canWrite={canWrite}
                expanded={expandedId === ib.id}
                onToggleExpand={() => setExpandedId((cur) => (cur === ib.id ? null : ib.id))}
                onAddSubHost={() => setSubHostsFor(ib)}
                onEdit={() => openNodeManager(ib.node_name, ib)}
              />
            ))}
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

function InboundRow({
  ib,
  showNode,
  canWrite,
  expanded,
  onToggleExpand,
  onAddSubHost,
  onEdit,
}: {
  ib: InboundFleetRow;
  showNode: boolean;
  canWrite: boolean;
  expanded: boolean;
  onToggleExpand: () => void;
  onAddSubHost: () => void;
  onEdit: () => void;
}) {
  const { t } = useI18n();
  const sniList = (ib.sni ?? []).filter(Boolean);

  return (
    <div className="p-4 hover:bg-surface/30 transition-colors">
      <div className="flex flex-wrap items-center justify-between gap-3">
        <div className="flex items-center gap-3 min-w-0">
          <div className="h-9 w-9 rounded-xl bg-primary/10 flex items-center justify-center text-primary flex-shrink-0">
            <Globe size={16} />
          </div>
          <div className="min-w-0">
            <div className="flex flex-wrap items-center gap-1.5">
              <span className="font-semibold text-fg text-sm truncate">{ib.tag}</span>
              <ProtocolBadge label={ib.protocol} />
              <span className="inline-flex items-center rounded-md border border-border/60 bg-surface-2 px-2 py-0.5 text-[10px] font-semibold uppercase tracking-wide text-fg-muted">
                {transportLabel(ib)}
              </span>
              {showNode && (
                <span className="inline-flex items-center gap-1 rounded-md bg-surface-2/70 px-2 py-0.5 text-[10px] font-medium text-fg-subtle">
                  <Server size={10} /> {ib.node_name}
                </span>
              )}
            </div>
            <div className="mt-1.5 flex flex-wrap items-center gap-x-3 gap-y-1 text-xs text-fg-muted">
              <span>
                {t("hosts.port")}: <span className="font-mono text-fg">{ib.port}</span>
              </span>
              <span className="flex items-center gap-1.5">
                {t("hosts.security")}:{" "}
                <Badge color={securityColor(ib.security)}>{securityLabel(ib.security)}</Badge>
              </span>
              {sniList.length > 0 && (
                <span className="truncate max-w-[320px]" title={sniList.join(", ")}>
                  SNI: <span className="text-fg">{sniList.join(", ")}</span>
                </span>
              )}
            </div>
          </div>
        </div>

        <div className="flex items-center gap-1.5 flex-shrink-0">
          <StatusBadge
            status={ib.enabled ? "active" : "inactive"}
            label={ib.enabled ? "ENABLED" : "DISABLED"}
            pulse={false}
          />
          <Button variant="outline" size="sm" onClick={onAddSubHost}>
            <Plus size={13} /> {t("hosts.add")}
          </Button>
          <Button
            variant="ghost"
            size="sm"
            onClick={onToggleExpand}
            aria-label={t("hosts.title")}
            className="!px-2"
          >
            <ChevronDown size={15} className={cn("transition-transform", expanded && "rotate-180")} />
          </Button>
          {canWrite && (
            <Button variant="ghost" size="sm" onClick={onEdit} className="!px-2" aria-label={t("common.edit")}>
              <Settings size={15} />
            </Button>
          )}
        </div>
      </div>

      {expanded && <SubHostsPreview inboundId={ib.id} onOpen={onAddSubHost} />}
    </div>
  );
}

function SubHostsPreview({ inboundId, onOpen }: { inboundId: string; onOpen: () => void }) {
  const { t } = useI18n();
  const { data, isLoading } = useSubHosts(inboundId);
  const hosts = data?.hosts ?? [];

  return (
    <div className="mt-3 ms-12 space-y-1.5">
      {isLoading && <p className="text-xs text-fg-subtle">{t("common.loading")}</p>}
      {!isLoading && hosts.length === 0 && (
        <button
          type="button"
          onClick={onOpen}
          className="text-xs text-fg-subtle hover:text-primary transition-colors"
        >
          {t("hosts.empty")}
        </button>
      )}
      {hosts.map((h: SubHost) => (
        <button
          key={h.id}
          type="button"
          onClick={onOpen}
          className="w-full flex items-center justify-between gap-2 rounded-lg bg-surface/60 border border-border/30 px-3 py-1.5 text-xs hover:bg-surface-2/60 hover:border-primary/20 transition-colors text-start"
        >
          <span className="flex items-center gap-2 min-w-0">
            <span
              className={cn(
                "h-1.5 w-1.5 rounded-full flex-shrink-0",
                h.enabled ? "bg-success" : "bg-fg-subtle",
              )}
            />
            <span className="font-medium text-fg truncate">{h.remark || "—"}</span>
            <span className="text-fg-subtle truncate" dir="ltr">
              {h.address}
              {h.port ? `:${h.port}` : ""}
            </span>
          </span>
          <Badge color={securityColor(h.security as HostSecurity)}>{securityLabel(h.security)}</Badge>
        </button>
      ))}
    </div>
  );
}
