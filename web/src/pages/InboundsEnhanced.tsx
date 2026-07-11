import { useMemo, useState } from "react";
import { ChevronDown, Globe, Info, Plus, Settings, Search, Copy, CheckCircle } from "lucide-react";
import {
    useInboundsFleet,
    useNodes,
    type Inbound,
    type InboundFleetRow,
} from "@/api/hooks";
import type { Node } from "@/api/types";
import { Button, Input } from "@/components/ui";
import { NodeInboundsModal } from "@/components/NodeInboundsModal";
import { SubHostsModal } from "@/components/SubHostsModal";
import { GlassCard } from "@/components/veltrix";
import { useTitle } from "@/lib/useTitle";
import { useAuth } from "@/auth/auth";
import { cn } from "@/lib/utils";

function transportLabel(ib: InboundFleetRow): string {
    const net = ib.network || "tcp";
    return net.toUpperCase();
}

function securityLabel(s?: string): string {
    if (!s || s === "none") return "none";
    if (s === "inbound_default") return "default";
    return s;
}

export function InboundsEnhanced() {
    useTitle("Inbounds & Hosts");
    const { can } = useAuth();
    const canWrite = can("inbound:write");
    const fleet = useInboundsFleet();
    const nodes = useNodes();

    const [nodeFilter, setNodeFilter] = useState("");
    const [searchQuery, setSearchQuery] = useState("");
    const [managing, setManaging] = useState<Node | null>(null);
    const [pendingEdit, setPendingEdit] = useState<Inbound | null>(null);
    const [subHostsFor, setSubHostsFor] = useState<Inbound | null>(null);
    const [expandedId, setExpandedId] = useState<string | null>(null);
    const [copiedId, setCopiedId] = useState<string | null>(null);

    const inbounds = fleet.data?.inbounds ?? [];
    const nodeList = nodes.data?.nodes ?? [];

    const filtered = useMemo(() => {
        let result = inbounds;

        if (nodeFilter) {
            result = result.filter((ib) => ib.node_name === nodeFilter);
        }

        if (searchQuery) {
            const q = searchQuery.toLowerCase();
            result = result.filter(
                (ib) =>
                    ib.tag?.toLowerCase().includes(q) ||
                    ib.node_name.toLowerCase().includes(q),
            );
        }

        return result;
    }, [inbounds, nodeFilter, searchQuery]);

    const nodeNames = useMemo(
        () => [...new Set(inbounds.map((ib) => ib.node_name))].sort(),
        [inbounds],
    );

    const stats = useMemo(
        () => ({
            total: inbounds.length,
            byProtocol: inbounds.reduce(
                (acc, ib) => {
                    const proto = ib.protocol || "unknown";
                    acc[proto] = (acc[proto] || 0) + 1;
                    return acc;
                },
                {} as Record<string, number>,
            ),
            byNode: inbounds.reduce(
                (acc, ib) => {
                    acc[ib.node_name] = (acc[ib.node_name] || 0) + 1;
                    return acc;
                },
                {} as Record<string, number>,
            ),
            active: inbounds.filter((ib) => ib.enabled !== false).length,
        }),
        [inbounds],
    );

    const copyToClipboard = (text: string, id: string) => {
        navigator.clipboard.writeText(text);
        setCopiedId(id);
        setTimeout(() => setCopiedId(null), 2000);
    };

    const openNodeManager = (nodeName: string, edit?: Inbound) => {
        const node = nodeList.find((n) => n.name === nodeName);
        if (node) {
            setPendingEdit(edit ?? null);
            setManaging(node);
        }
    };

    return (
        <div className="space-y-6 animate-page-enter">
            <NodeInboundsModal
                node={managing}
                initialEdit={pendingEdit}
                onClose={() => {
                    setManaging(null);
                    setPendingEdit(null);
                }}
            />
            <SubHostsModal inbound={subHostsFor} onClose={() => setSubHostsFor(null)} />

            {/* Header */}
            <div className="flex flex-col sm:flex-row sm:items-start justify-between gap-4">
                <div>
                    <h1 className="text-2xl font-bold text-fg tracking-tight">Inbounds & Hosts</h1>
                    <p className="text-sm text-fg-secondary mt-1">
                        Manage {stats.total} inbound{stats.total !== 1 ? "s" : ""} across {nodeNames.length} node
                        {nodeNames.length !== 1 ? "s" : ""}
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
                        <Plus size={16} className="mr-2" />
                        Add Inbound
                    </Button>
                )}
            </div>

            {/* Statistics */}
            <div className="grid grid-cols-2 sm:grid-cols-4 gap-3">
                <GlassCard className="p-4">
                    <p className="text-xs text-fg-secondary uppercase tracking-wide">Total Inbounds</p>
                    <p className="text-2xl font-bold text-fg mt-1">{stats.total}</p>
                </GlassCard>
                <GlassCard className="p-4">
                    <p className="text-xs text-fg-secondary uppercase tracking-wide">Active</p>
                    <p className="text-2xl font-bold text-green-500 mt-1">{stats.active}</p>
                </GlassCard>
                <GlassCard className="p-4">
                    <p className="text-xs text-fg-secondary uppercase tracking-wide">Nodes</p>
                    <p className="text-2xl font-bold text-blue-500 mt-1">{nodeNames.length}</p>
                </GlassCard>
                <GlassCard className="p-4">
                    <p className="text-xs text-fg-secondary uppercase tracking-wide">Protocols</p>
                    <p className="text-2xl font-bold text-purple-500 mt-1">
                        {Object.keys(stats.byProtocol).length}
                    </p>
                </GlassCard>
            </div>

            {/* Info Box */}
            <GlassCard className="!p-3.5 bg-blue-500/10 border-blue-500/20">
                <div className="flex items-start gap-2.5">
                    <div className="h-7 w-7 rounded-lg bg-blue-500/15 flex items-center justify-center text-blue-500 flex-shrink-0">
                        <Info size={14} />
                    </div>
                    <div className="min-w-0">
                        <p className="text-xs font-semibold text-fg">Subscription Hosts</p>
                        <p className="text-xs text-fg-secondary mt-0.5 leading-relaxed">
                            Create alternate domain/SNI combinations for your inbounds. Users can choose which host
                            to use in their subscriptions.
                        </p>
                    </div>
                </div>
            </GlassCard>

            {/* Search & Filters */}
            <GlassCard className="!p-0 overflow-hidden">
                <div className="space-y-3 p-4 border-b border-fg-secondary/20">
                    {/* Search */}
                    <div className="relative">
                        <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 w-4 h-4 text-fg-secondary pointer-events-none" />
                        <Input
                            placeholder="Search by name, node, address..."
                            value={searchQuery}
                            onChange={(e) => setSearchQuery(e.target.value)}
                            className="pl-9"
                        />
                    </div>

                    <div className="flex flex-wrap gap-2">
                        <button
                            type="button"
                            onClick={() => setNodeFilter("")}
                            className={cn(
                                "px-3 py-1.5 rounded text-xs font-medium transition-all",
                                !nodeFilter
                                    ? "bg-blue-500 text-white"
                                    : "text-fg-secondary hover:text-fg hover:bg-fg/5 border border-fg-secondary/20",
                            )}
                        >
                            All ({inbounds.length})
                        </button>
                        {nodeNames.map((name) => (
                            <button
                                key={name}
                                type="button"
                                onClick={() => setNodeFilter(name)}
                                className={cn(
                                    "px-3 py-1.5 rounded text-xs font-medium transition-all",
                                    nodeFilter === name
                                        ? "bg-blue-500 text-white"
                                        : "text-fg-secondary hover:text-fg hover:bg-fg/5 border border-fg-secondary/20",
                                )}
                            >
                                {name} ({stats.byNode[name]})
                            </button>
                        ))}
                    </div>
                </div>

                {/* Results */}
                {fleet.isLoading && (
                    <div className="p-8 text-sm text-fg-secondary text-center">Loading inbounds...</div>
                )}

                {!fleet.isLoading && filtered.length === 0 && (
                    <div className="flex flex-col items-center gap-3 py-12 text-center">
                        <Globe size={32} className="text-fg-secondary/40" />
                        <p className="text-sm text-fg-secondary">
                            {searchQuery ? "No inbounds match your search" : "No inbounds configured"}
                        </p>
                        {canWrite && nodeList.length > 0 && (
                            <Button variant="outline" size="sm" onClick={() => setManaging(nodeList[0])}>
                                <Plus size={14} className="mr-2" />
                                Create Inbound
                            </Button>
                        )}
                    </div>
                )}

                {!fleet.isLoading && filtered.length > 0 && (
                    <div className="divide-y divide-fg-secondary/10">
                        {filtered.map((ib) => (
                            <div key={ib.id} className="hover:bg-fg/5 transition p-4">
                                <div className="flex items-start justify-between gap-4">
                                    <div className="flex-1 min-w-0">
                                        <div className="flex items-center gap-2 mb-2">
                                            <p className="font-bold text-fg truncate">{ib.tag || ib.protocol}</p>
                                            <div className="flex h-6 items-center gap-1 rounded px-2.5 text-xs font-medium bg-green-500/10 text-green-600">
                                                active
                                            </div>
                                            <div className="flex h-6 items-center gap-1 rounded px-2.5 text-xs font-medium"
                                                style={{ backgroundColor: 'rgba(34, 197, 94, 0.1)', color: 'rgb(34, 197, 94)' }}>
                                                {securityLabel(ib.security)}
                                            </div>
                                        </div>
                                        <div className="flex flex-wrap items-center gap-2 text-xs text-fg-secondary mb-2">
                                            <span className="bg-fg/10 px-2 py-1 rounded">{ib.node_name}</span>
                                            <span>
                                                0.0.0.0:{ib.port}
                                            </span>
                                            <span className="bg-fg/10 px-2 py-1 rounded">{ib.protocol}</span>
                                            <span className="text-fg-secondary">{transportLabel(ib)}</span>
                                        </div>
                                        {ib.sni && <p className="text-xs text-fg-secondary">SNI: {ib.sni}</p>}
                                    </div>

                                    <div className="flex items-center gap-2">
                                        {canWrite && (
                                            <>
                                                <Button
                                                    size="sm"
                                                    variant="ghost"
                                                    onClick={() => openNodeManager(ib.node_name, ib)}
                                                >
                                                    <Settings size={16} />
                                                </Button>
                                                {/* <Button
                          size="sm"
                          variant="ghost"
                          onClick={() => setSubHostsFor(ib)}
                        >
                          <Globe size={16} />
                        </Button> */}
                                            </>
                                        )}

                                        <button
                                            onClick={() =>
                                                copyToClipboard(`0.0.0.0:${ib.port}`, ib.id)
                                            }
                                            className="p-2 hover:bg-fg/10 rounded transition"
                                            title="Copy address"
                                        >
                                            {copiedId === ib.id ? (
                                                <CheckCircle size={16} className="text-green-500" />
                                            ) : (
                                                <Copy size={16} className="text-fg-secondary" />
                                            )}
                                        </button>

                                        <button
                                            onClick={() => setExpandedId((cur) => (cur === ib.id ? null : ib.id))}
                                            className={cn(
                                                "p-2 hover:bg-fg/10 rounded transition",
                                                expandedId === ib.id && "bg-fg/10",
                                            )}
                                        >
                                            <ChevronDown
                                                size={16}
                                                className={cn("transition", expandedId === ib.id && "rotate-180")}
                                            />
                                        </button>
                                    </div>
                                </div>

                                {/* Expanded Details */}
                                {expandedId === ib.id && (
                                    <div className="mt-3 pt-3 border-t border-fg-secondary/10 space-y-2 text-xs">
                                        <div className="flex justify-between">
                                            <span className="text-fg-secondary">Tag:</span>
                                            <code className="text-fg font-mono">{ib.tag || "N/A"}</code>
                                        </div>
                                        <div className="flex justify-between">
                                            <span className="text-fg-secondary">Protocol:</span>
                                            <code className="text-fg font-mono">{ib.protocol}</code>
                                        </div>
                                        {Object.keys(ib.raw || {}).length > 0 && (
                                            <div className="mt-2 pt-2 border-t border-fg-secondary/10">
                                                <p className="text-fg-secondary mb-1">Raw Config:</p>
                                                <pre className="bg-fg/5 p-2 rounded text-xs overflow-x-auto max-h-24">
                                                    {JSON.stringify(ib.raw, null, 2)}
                                                </pre>
                                            </div>
                                        )}
                                    </div>
                                )}
                            </div>
                        ))}
                    </div>
                )}
            </GlassCard>

            {/* Protocol Distribution */}
            {Object.keys(stats.byProtocol).length > 0 && (
                <GlassCard className="p-4">
                    <h3 className="font-semibold text-fg mb-3 text-sm">Protocol Distribution</h3>
                    <div className="space-y-2">
                        {Object.entries(stats.byProtocol)
                            .sort((a, b) => b[1] - a[1])
                            .map(([proto, count]) => (
                                <div key={proto} className="flex items-center justify-between">
                                    <div className="flex items-center gap-2">
                                        <span className="inline-flex h-6 items-center gap-1 rounded border border-fg-secondary/20 px-2.5 text-xs font-medium">{proto.toUpperCase()}</span>
                                        <span className="text-sm text-fg-secondary">{proto}</span>
                                    </div>
                                    <div className="flex items-center gap-3">
                                        <div className="w-32 h-2 bg-fg/10 rounded overflow-hidden">
                                            <div
                                                className="h-full bg-blue-500 transition-all"
                                                style={{
                                                    width: `${(count / stats.total) * 100}%`,
                                                }}
                                            />
                                        </div>
                                        <span className="text-sm font-bold text-fg w-6 text-right">{count}</span>
                                    </div>
                                </div>
                            ))}
                    </div>
                </GlassCard>
            )}
        </div>
    );
}

export { InboundsEnhanced as Inbounds };
