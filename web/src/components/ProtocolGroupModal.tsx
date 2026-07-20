import { useState, useEffect } from "react";
import { useNodeInbounds } from "@/api/hooks";
import {
  useCreateProtocolGroup,
  useUpdateProtocolGroup,
  useReorderGroupInbounds,
  type ProtocolGroup,
  type CreateProtocolGroupInput,
} from "@/api/protocol-group-hooks";
import { useI18n } from "@/i18n/i18n";
import { Button, Input } from "./ui";
import { Modal } from "./Modal";
import { useToast } from "./toast";
import { GripVertical } from "lucide-react";

interface Props {
  open: boolean;
  onClose: () => void;
  nodeId: string;
  group?: ProtocolGroup | null; // null = create mode
}

export function ProtocolGroupModal({ open, onClose, nodeId, group }: Props) {
  const { t } = useI18n();
  const toast = useToast();
  const inboundsQuery = useNodeInbounds(nodeId);
  const create = useCreateProtocolGroup();
  const update = useUpdateProtocolGroup();
  const reorder = useReorderGroupInbounds();

  const [name, setName] = useState("");
  const [selectedIds, setSelectedIds] = useState<string[]>([]);
  const [probeUrl, setProbeUrl] = useState("https://www.gstatic.com/generate_204");
  const [probeInterval, setProbeInterval] = useState(90);
  const [probeTimeout, setProbeTimeout] = useState(5);
  const [maxRetries, setMaxRetries] = useState(3);

  // Reset form when opening
  useEffect(() => {
    if (open && group) {
      setName(group.name);
      setSelectedIds(group.inbound_ids ?? []);
      setProbeUrl(group.probe_url || "https://www.gstatic.com/generate_204");
      setProbeInterval(group.probe_interval || 90);
      setProbeTimeout(group.probe_timeout || 5);
      setMaxRetries(group.max_retries || 3);
    } else if (open) {
      setName("");
      setSelectedIds([]);
      setProbeUrl("https://www.gstatic.com/generate_204");
      setProbeInterval(90);
      setProbeTimeout(5);
      setMaxRetries(3);
    }
  }, [open, group]);

  const inbounds = inboundsQuery.data?.inbounds ?? [];

  function toggleInbound(id: string) {
    setSelectedIds((prev) =>
      prev.includes(id) ? prev.filter((x) => x !== id) : [...prev, id],
    );
  }

  function moveUp(idx: number) {
    if (idx === 0) return;
    const next = [...selectedIds];
    [next[idx - 1], next[idx]] = [next[idx], next[idx - 1]];
    setSelectedIds(next);
  }

  function moveDown(idx: number) {
    if (idx >= selectedIds.length - 1) return;
    const next = [...selectedIds];
    [next[idx], next[idx + 1]] = [next[idx + 1], next[idx]];
    setSelectedIds(next);
  }

  async function handleSubmit() {
    const input: CreateProtocolGroupInput = {
      node_id: nodeId,
      name,
      inbound_ids: selectedIds,
      probe_url: probeUrl,
      probe_interval: probeInterval,
      probe_timeout: probeTimeout,
      max_retries: maxRetries,
    };

    try {
      if (group) {
        await update.mutateAsync({ id: group.id, ...input });
        // If order changed, also call reorder
        if (JSON.stringify(group.inbound_ids) !== JSON.stringify(selectedIds)) {
          await reorder.mutateAsync({ groupId: group.id, inboundIds: selectedIds });
        }
        toast.success(t("protocolGroups.updated"));
      } else {
        await create.mutateAsync(input);
        toast.success(t("protocolGroups.created"));
      }
      onClose();
    } catch (e: any) {
      toast.error(e?.message ?? "Failed");
    }
  }

  const isLoading = create.isPending || update.isPending || reorder.isPending;
  const inboundMap = Object.fromEntries(inbounds.map((ib) => [ib.id, ib]));

  return (
    <Modal
      open={open}
      onClose={onClose}
      title={group ? t("protocolGroups.edit") : t("protocolGroups.create")}
      className="max-w-lg"
    >
      <div className="space-y-4">
        {/* Name */}
        <div>
          <label className="text-xs text-fg-muted mb-1 block">{t("protocolGroups.name")}</label>
          <Input
            value={name}
            onChange={(e) => setName(e.target.value)}
            placeholder={t("protocolGroups.namePlaceholder")}
          />
        </div>

        {/* Inbound Selection + Reorder */}
        <div>
          <label className="text-xs text-fg-muted mb-1 block">
            {t("protocolGroups.inbounds")}
          </label>
          <p className="text-[11px] text-fg-subtle mb-2">{t("protocolGroups.inboundsHint")}</p>

          {/* Selected inbounds (ordered) */}
          {selectedIds.length > 0 && (
            <div className="space-y-1 mb-2 border border-border/40 rounded-lg p-2 bg-surface/40">
              {selectedIds.map((id, idx) => {
                const ib = inboundMap[id];
                return (
                  <div
                    key={id}
                    className="flex items-center gap-2 py-1 px-2 rounded-md bg-bg-elevated text-xs"
                  >
                    <GripVertical size={12} className="text-fg-subtle" />
                    <span className="font-medium text-fg flex-1">
                      {ib ? `${ib.tag} (${ib.protocol}/${ib.network})` : id.slice(0, 8)}
                    </span>
                    <span className="text-fg-subtle">#{idx + 1}</span>
                    <button
                      type="button"
                      onClick={() => moveUp(idx)}
                      disabled={idx === 0}
                      className="text-xs px-1 text-fg-muted hover:text-fg disabled:opacity-30"
                    >
                      ↑
                    </button>
                    <button
                      type="button"
                      onClick={() => moveDown(idx)}
                      disabled={idx === selectedIds.length - 1}
                      className="text-xs px-1 text-fg-muted hover:text-fg disabled:opacity-30"
                    >
                      ↓
                    </button>
                    <button
                      type="button"
                      onClick={() => toggleInbound(id)}
                      className="text-xs px-1 text-danger"
                    >
                      ✕
                    </button>
                  </div>
                );
              })}
            </div>
          )}

          {/* Available inbounds to add */}
          <div className="max-h-36 overflow-y-auto space-y-1 border border-border/30 rounded-lg p-2">
            {inbounds
              .filter((ib) => !selectedIds.includes(ib.id))
              .map((ib) => (
                <button
                  key={ib.id}
                  type="button"
                  onClick={() => toggleInbound(ib.id)}
                  className="w-full text-start text-xs py-1.5 px-2 rounded-md hover:bg-surface-2 text-fg-muted hover:text-fg transition"
                >
                  + {ib.tag} ({ib.protocol}/{ib.network}) :{ib.port}
                </button>
              ))}
            {inbounds.length === 0 && (
              <p className="text-[11px] text-fg-subtle py-2 text-center">Loading...</p>
            )}
          </div>
        </div>

        {/* Probe Settings */}
        <div className="grid grid-cols-2 gap-3">
          <div className="col-span-2">
            <label className="text-xs text-fg-muted mb-1 block">{t("protocolGroups.probeUrl")}</label>
            <Input value={probeUrl} onChange={(e) => setProbeUrl(e.target.value)} />
          </div>
          <div>
            <label className="text-xs text-fg-muted mb-1 block">{t("protocolGroups.probeInterval")}</label>
            <Input
              type="number"
              min={30}
              max={600}
              value={probeInterval}
              onChange={(e) => setProbeInterval(Number(e.target.value))}
            />
          </div>
          <div>
            <label className="text-xs text-fg-muted mb-1 block">{t("protocolGroups.probeTimeout")}</label>
            <Input
              type="number"
              min={1}
              max={30}
              value={probeTimeout}
              onChange={(e) => setProbeTimeout(Number(e.target.value))}
            />
          </div>
          <div>
            <label className="text-xs text-fg-muted mb-1 block">{t("protocolGroups.maxRetries")}</label>
            <Input
              type="number"
              min={1}
              max={10}
              value={maxRetries}
              onChange={(e) => setMaxRetries(Number(e.target.value))}
            />
          </div>
        </div>

        {/* Actions */}
        <div className="flex justify-end gap-2 pt-3">
          <Button variant="ghost" onClick={onClose}>
            {t("common.cancel") || "Cancel"}
          </Button>
          <Button onClick={handleSubmit} disabled={!name || selectedIds.length < 2 || isLoading}>
            {isLoading ? "..." : group ? t("common.save") || "Save" : t("protocolGroups.create")}
          </Button>
        </div>
      </div>
    </Modal>
  );
}
