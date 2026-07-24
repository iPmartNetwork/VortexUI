import { useState } from "react";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { Play, Eye, History, Users, Calendar, Database, Settings, Shield, Wifi } from "lucide-react";
import { api } from "@/api/client";
import { Button, Select, Badge } from "@/components/ui";
import { GlassCard } from "@/components/veltrix";
import { useConfirm } from "@/components/confirm";
import { useToast } from "@/components/toast";
import { useTitle } from "@/lib/useTitle";

// Types
interface BulkFilter {
  statuses?: string[];
  admin_id?: string;
  groups?: string[];
}

interface BulkPreviewResult {
  affected_count: number;
  summary: Record<string, unknown>;
}

interface BulkOperation {
  id: string;
  admin_id: string;
  operation_type: string;
  parameters: Record<string, unknown>;
  filters: BulkFilter;
  affected_count: number;
  status: string;
  created_at: string;
}

// Operation type categories
const OPERATION_CATEGORIES = [
  {
    label: "Groups",
    icon: Users,
    operations: [
      { value: "add_groups", label: "Add Groups" },
      { value: "remove_groups", label: "Remove Groups" },
    ],
  },
  {
    label: "Expiration",
    icon: Calendar,
    operations: [
      { value: "add_expire_days", label: "Add Expire Days" },
      { value: "sub_expire_days", label: "Subtract Expire Days" },
    ],
  },
  {
    label: "Data Limit",
    icon: Database,
    operations: [
      { value: "add_data_limit", label: "Add Data Limit" },
      { value: "sub_data_limit", label: "Subtract Data Limit" },
    ],
  },
  {
    label: "Proxy Settings",
    icon: Settings,
    operations: [{ value: "update_proxy_settings", label: "Update Proxy Settings" }],
  },
  {
    label: "WireGuard",
    icon: Wifi,
    operations: [
      { value: "allocate_wg_peers", label: "Allocate WG Peers" },
      { value: "repair_wg_peers", label: "Repair WG Peers" },
    ],
  },
  {
    label: "Status",
    icon: Shield,
    operations: [{ value: "change_status", label: "Change Status" }],
  },
] as const;

const USER_STATUSES = ["active", "limited", "expired", "disabled", "on_hold"];

export function BulkOperations() {
  useTitle("Bulk Operations");
  const queryClient = useQueryClient();
  const confirm = useConfirm();
  const toast = useToast();

  const [activeTab, setActiveTab] = useState<"operations" | "history">("operations");
  const [operationType, setOperationType] = useState("");
  const [parameters, setParameters] = useState<Record<string, unknown>>({});
  const [filters, setFilters] = useState<BulkFilter>({});
  const [previewResult, setPreviewResult] = useState<BulkPreviewResult | null>(null);

  // History query
  const { data: historyData, isLoading: historyLoading } = useQuery({
    queryKey: ["bulk-history"],
    queryFn: () =>
      api<{ operations: BulkOperation[]; total: number }>("/api/v2/bulk/history", {
        query: { limit: 50, offset: 0 },
      }),
    enabled: activeTab === "history",
  });

  // Preview mutation
  const previewMutation = useMutation({
    mutationFn: () =>
      api<{ preview: BulkPreviewResult }>("/api/v2/bulk/preview", {
        method: "POST",
        body: { operation_type: operationType, parameters, filters },
      }),
    onSuccess: (data) => {
      setPreviewResult(data.preview);
    },
    onError: (err: Error) => {
      toast.error(err.message || "Preview failed");
    },
  });

  // Execute mutation
  const executeMutation = useMutation({
    mutationFn: () =>
      api<{ operation: BulkOperation }>("/api/v2/bulk/execute", {
        method: "POST",
        body: { operation_type: operationType, parameters, filters },
      }),
    onSuccess: (data) => {
      toast.success(`Operation completed. ${data.operation.affected_count} users affected.`);
      setPreviewResult(null);
      queryClient.invalidateQueries({ queryKey: ["bulk-history"] });
    },
    onError: (err: Error) => {
      toast.error(err.message || "Execute failed");
    },
  });

  async function handleExecute() {
    if (
      await confirm({
        title: "Execute Bulk Operation",
        message: `This will apply "${operationType}" to ${previewResult?.affected_count ?? "unknown"} users. This action cannot be undone.`,
        confirmLabel: "Execute",
        destructive: true,
      })
    ) {
      executeMutation.mutate();
    }
  }

  function updateParam(key: string, value: unknown) {
    setParameters((prev) => ({ ...prev, [key]: value }));
  }

  function toggleStatusFilter(status: string) {
    setFilters((prev) => {
      const current = prev.statuses ?? [];
      const next = current.includes(status) ? current.filter((s) => s !== status) : [...current, status];
      return { ...prev, statuses: next.length > 0 ? next : undefined };
    });
  }

  return (
    <div className="space-y-6 animate-page-enter">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-bold text-fg">Bulk Operations</h1>
      </div>

      {/* Tabs */}
      <div className="flex gap-1 rounded-xl bg-surface-2/60 p-1">
        <button
          onClick={() => setActiveTab("operations")}
          className={`flex items-center gap-2 rounded-lg px-4 py-2 text-sm font-medium transition ${
            activeTab === "operations" ? "bg-surface text-fg shadow-sm" : "text-fg-muted hover:text-fg"
          }`}
        >
          <Play size={16} />
          Operations
        </button>
        <button
          onClick={() => setActiveTab("history")}
          className={`flex items-center gap-2 rounded-lg px-4 py-2 text-sm font-medium transition ${
            activeTab === "history" ? "bg-surface text-fg shadow-sm" : "text-fg-muted hover:text-fg"
          }`}
        >
          <History size={16} />
          History
        </button>
      </div>

      {activeTab === "operations" && (
        <div className="grid grid-cols-1 gap-6 lg:grid-cols-3">
          {/* Operation Selection */}
          <GlassCard className="lg:col-span-2 space-y-4 p-6">
            <h2 className="text-lg font-semibold text-fg">Select Operation</h2>
            <div className="grid grid-cols-2 gap-3 sm:grid-cols-3">
              {OPERATION_CATEGORIES.map((cat) => (
                <div key={cat.label} className="space-y-2">
                  <div className="flex items-center gap-2 text-xs font-medium uppercase text-fg-muted">
                    <cat.icon size={14} />
                    {cat.label}
                  </div>
                  {cat.operations.map((op) => (
                    <button
                      key={op.value}
                      onClick={() => {
                        setOperationType(op.value);
                        setPreviewResult(null);
                      }}
                      className={`w-full rounded-lg border px-3 py-2 text-left text-sm transition ${
                        operationType === op.value
                          ? "border-primary bg-primary/10 text-primary"
                          : "border-border text-fg-muted hover:border-primary/50 hover:text-fg"
                      }`}
                    >
                      {op.label}
                    </button>
                  ))}
                </div>
              ))}
            </div>

            {/* Parameters */}
            {operationType && (
              <div className="space-y-3 border-t border-border pt-4">
                <h3 className="text-sm font-medium text-fg">Parameters</h3>
                {(operationType === "add_expire_days" || operationType === "sub_expire_days") && (
                  <div>
                    <label className="text-xs text-fg-muted">Days</label>
                    <input
                      type="number"
                      min={1}
                      value={(parameters.days as number) || ""}
                      onChange={(e) => updateParam("days", Number(e.target.value))}
                      className="mt-1 w-full rounded-lg border border-border bg-surface px-3 py-2 text-sm text-fg"
                      placeholder="Number of days"
                    />
                  </div>
                )}
                {(operationType === "add_data_limit" || operationType === "sub_data_limit") && (
                  <div>
                    <label className="text-xs text-fg-muted">Data (GB)</label>
                    <input
                      type="number"
                      min={0.1}
                      step={0.1}
                      value={((parameters.bytes as number) || 0) / 1073741824 || ""}
                      onChange={(e) => updateParam("bytes", Number(e.target.value) * 1073741824)}
                      className="mt-1 w-full rounded-lg border border-border bg-surface px-3 py-2 text-sm text-fg"
                      placeholder="Amount in GB"
                    />
                  </div>
                )}
                {operationType === "change_status" && (
                  <Select
                    value={(parameters.status as string) || ""}
                    onChange={(e) => updateParam("status", e.target.value)}
                  >
                    <option value="" disabled>Select target status</option>
                    {USER_STATUSES.map((s) => (
                      <option key={s} value={s}>{s}</option>
                    ))}
                  </Select>
                )}
              </div>
            )}
          </GlassCard>

          {/* Filters Panel */}
          <GlassCard className="space-y-4 p-6">
            <h2 className="text-lg font-semibold text-fg">Filters</h2>

            <div className="space-y-3">
              <div>
                <label className="text-xs font-medium text-fg-muted">Status Filter</label>
                <div className="mt-1 flex flex-wrap gap-2">
                  {USER_STATUSES.map((status) => (
                    <button
                      key={status}
                      onClick={() => toggleStatusFilter(status)}
                      className={`rounded-full px-3 py-1 text-xs font-medium transition ${
                        filters.statuses?.includes(status)
                          ? "bg-primary text-white"
                          : "bg-surface-2 text-fg-muted hover:text-fg"
                      }`}
                    >
                      {status}
                    </button>
                  ))}
                </div>
              </div>

              <div>
                <label className="text-xs font-medium text-fg-muted">Admin ID</label>
                <input
                  type="text"
                  value={filters.admin_id || ""}
                  onChange={(e) => setFilters((prev) => ({ ...prev, admin_id: e.target.value || undefined }))}
                  className="mt-1 w-full rounded-lg border border-border bg-surface px-3 py-2 text-sm text-fg"
                  placeholder="Filter by admin UUID"
                />
              </div>
            </div>

            {/* Preview / Execute Actions */}
            <div className="space-y-3 border-t border-border pt-4">
              <Button
                onClick={() => previewMutation.mutate()}
                disabled={!operationType || previewMutation.isPending}
                className="w-full"
                variant="outline"
              >
                <Eye size={16} className="mr-2" />
                {previewMutation.isPending ? "Previewing..." : "Preview"}
              </Button>

              {previewResult && (
                <div className="rounded-lg border border-border bg-surface-2/50 p-3 text-sm">
                  <div className="flex items-center justify-between">
                    <span className="text-fg-muted">Affected Users:</span>
                    <span className="font-semibold text-fg">{previewResult.affected_count}</span>
                  </div>
                </div>
              )}

              <Button
                onClick={handleExecute}
                disabled={!previewResult || executeMutation.isPending}
                className="w-full"
                variant="destructive"
              >
                <Play size={16} className="mr-2" />
                {executeMutation.isPending ? "Executing..." : "Execute"}
              </Button>
            </div>
          </GlassCard>
        </div>
      )}

      {activeTab === "history" && (
        <GlassCard className="p-6">
          <h2 className="mb-4 text-lg font-semibold text-fg">Operation History</h2>
          {historyLoading ? (
            <div className="py-8 text-center text-fg-muted">Loading...</div>
          ) : (
            <div className="overflow-x-auto">
              <table className="w-full text-sm">
                <thead>
                  <tr className="border-b border-border text-left text-fg-muted">
                    <th className="pb-3 pr-4 font-medium">Type</th>
                    <th className="pb-3 pr-4 font-medium">Affected</th>
                    <th className="pb-3 pr-4 font-medium">Status</th>
                    <th className="pb-3 pr-4 font-medium">Date</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-border">
                  {historyData?.operations.map((op) => (
                    <tr key={op.id} className="text-fg">
                      <td className="py-3 pr-4">
                        <span className="rounded-md bg-surface-2 px-2 py-1 text-xs font-medium">
                          {op.operation_type}
                        </span>
                      </td>
                      <td className="py-3 pr-4">{op.affected_count}</td>
                      <td className="py-3 pr-4">
                        <Badge color={op.status === "completed" ? "active" : "disabled"}>{op.status}</Badge>
                      </td>
                      <td className="py-3 pr-4 text-fg-muted">
                        {new Date(op.created_at).toLocaleString()}
                      </td>
                    </tr>
                  ))}
                  {(!historyData?.operations || historyData.operations.length === 0) && (
                    <tr>
                      <td colSpan={4} className="py-8 text-center text-fg-muted">
                        No operations recorded yet.
                      </td>
                    </tr>
                  )}
                </tbody>
              </table>
            </div>
          )}
        </GlassCard>
      )}
    </div>
  );
}
