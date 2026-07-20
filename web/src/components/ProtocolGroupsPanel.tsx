import { useState } from "react";
import {
  useProtocolGroups,
  useDeleteProtocolGroup,
  useSwitchSummary,
  type ProtocolGroup,
} from "@/api/protocol-group-hooks";
import { useI18n } from "@/i18n/i18n";
import { Button, Card, Badge } from "./ui";
import { useToast } from "./toast";
import { ProtocolGroupModal } from "./ProtocolGroupModal";
import { ISPProfileModal } from "./ISPProfileModal";
import { Plus, Pencil, Trash2, Network, Globe } from "lucide-react";

interface Props {
  nodeId: string;
}

export function ProtocolGroupsPanel({ nodeId }: Props) {
  const { t } = useI18n();
  const toast = useToast();
  const groupsQuery = useProtocolGroups(nodeId);
  const summaryQuery = useSwitchSummary({ node_id: nodeId });
  const deleteGroup = useDeleteProtocolGroup();

  const [groupModalOpen, setGroupModalOpen] = useState(false);
  const [editGroup, setEditGroup] = useState<ProtocolGroup | null>(null);
  const [ispModalGroup, setIspModalGroup] = useState<ProtocolGroup | null>(null);

  const groups = groupsQuery.data ?? [];
  const summary = summaryQuery.data;

  function handleCreate() {
    setEditGroup(null);
    setGroupModalOpen(true);
  }

  function handleEdit(g: ProtocolGroup) {
    setEditGroup(g);
    setGroupModalOpen(true);
  }

  async function handleDelete(g: ProtocolGroup) {
    if (!confirm(t("protocolGroups.deleteConfirm"))) return;
    try {
      await deleteGroup.mutateAsync(g.id);
      toast.success(t("protocolGroups.deleted"));
    } catch (e: any) {
      toast.error(e?.message ?? "Failed");
    }
  }

  return (
    <div className="space-y-4">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h3 className="text-sm font-semibold text-fg">{t("protocolGroups.title")}</h3>
          <p className="text-[11px] text-fg-subtle">{t("protocolGroups.subtitle")}</p>
        </div>
        <Button size="sm" onClick={handleCreate}>
          <Plus size={14} />
          {t("protocolGroups.create")}
        </Button>
      </div>

      {/* Switch Summary Stats */}
      {summary && summary.total_switches > 0 && (
        <Card className="p-3">
          <p className="text-xs font-medium text-fg-muted mb-2">
            {t("protocolGroups.switchSummary")}
          </p>
          <div className="grid grid-cols-3 gap-3">
            <div>
              <p className="text-lg font-bold text-fg">{summary.total_switches}</p>
              <p className="text-[10px] text-fg-subtle">{t("protocolGroups.totalSwitches")}</p>
            </div>
            <div>
              <p className="text-[10px] text-fg-subtle mb-1">{t("protocolGroups.byProtocol")}</p>
              <div className="space-y-0.5">
                {Object.entries(summary.by_protocol ?? {})
                  .slice(0, 3)
                  .map(([proto, count]) => (
                    <div key={proto} className="flex justify-between text-[10px]">
                      <span className="text-fg-muted">{proto}</span>
                      <span className="text-fg font-medium">{count}</span>
                    </div>
                  ))}
              </div>
            </div>
            <div>
              <p className="text-[10px] text-fg-subtle mb-1">{t("protocolGroups.byISP")}</p>
              <div className="space-y-0.5">
                {Object.entries(summary.by_isp ?? {})
                  .slice(0, 3)
                  .map(([isp, count]) => (
                    <div key={isp} className="flex justify-between text-[10px]">
                      <span className="text-fg-muted">{isp}</span>
                      <span className="text-fg font-medium">{count}</span>
                    </div>
                  ))}
              </div>
            </div>
          </div>
        </Card>
      )}

      {/* Groups List */}
      {groups.length === 0 ? (
        <p className="text-xs text-fg-subtle text-center py-6">
          {t("protocolGroups.noGroups")}
        </p>
      ) : (
        <div className="space-y-2">
          {groups.map((g) => (
            <Card key={g.id} className="p-3">
              <div className="flex items-start justify-between">
                <div className="flex-1 min-w-0">
                  <div className="flex items-center gap-2">
                    <Network size={14} className="text-primary flex-shrink-0" />
                    <span className="text-sm font-medium text-fg truncate">{g.name}</span>
                    <Badge>
                      {g.inbound_ids?.length ?? 0} inbounds
                    </Badge>
                  </div>
                  <div className="flex items-center gap-3 mt-1.5 text-[10px] text-fg-subtle">
                    <span>Probe: {g.probe_interval}s</span>
                    <span>Timeout: {g.probe_timeout}s</span>
                    <span>Retries: {g.max_retries}</span>
                  </div>
                </div>
                <div className="flex items-center gap-1 flex-shrink-0">
                  <button
                    type="button"
                    onClick={() => setIspModalGroup(g)}
                    className="p-1.5 rounded-md hover:bg-surface-2 text-fg-muted hover:text-fg transition"
                    title={t("ispProfiles.title")}
                  >
                    <Globe size={13} />
                  </button>
                  <button
                    type="button"
                    onClick={() => handleEdit(g)}
                    className="p-1.5 rounded-md hover:bg-surface-2 text-fg-muted hover:text-fg transition"
                    title={t("protocolGroups.edit")}
                  >
                    <Pencil size={13} />
                  </button>
                  <button
                    type="button"
                    onClick={() => handleDelete(g)}
                    className="p-1.5 rounded-md hover:bg-surface-2 text-danger/70 hover:text-danger transition"
                    title="Delete"
                  >
                    <Trash2 size={13} />
                  </button>
                </div>
              </div>
            </Card>
          ))}
        </div>
      )}

      {/* Modals */}
      <ProtocolGroupModal
        open={groupModalOpen}
        onClose={() => setGroupModalOpen(false)}
        nodeId={nodeId}
        group={editGroup}
      />
      {ispModalGroup && (
        <ISPProfileModal
          open={!!ispModalGroup}
          onClose={() => setIspModalGroup(null)}
          groupId={ispModalGroup.id}
          groupName={ispModalGroup.name}
        />
      )}
    </div>
  );
}
