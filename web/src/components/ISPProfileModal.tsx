import { useState, useEffect } from "react";
import {
  useISPProfiles,
  useCreateISPProfile,
  useUpdateISPProfile,
  useDeleteISPProfile,
  type ISPProfile,
} from "@/api/protocol-group-hooks";
import { useI18n } from "@/i18n/i18n";
import { Button, Input } from "./ui";
import { Modal } from "./Modal";
import { useToast } from "./toast";
import { Trash2 } from "lucide-react";

interface Props {
  open: boolean;
  onClose: () => void;
  groupId: string;
  groupName: string;
}

export function ISPProfileModal({ open, onClose, groupId, groupName }: Props) {
  const { t } = useI18n();
  const toast = useToast();
  const profilesQuery = useISPProfiles(open ? groupId : null);
  const create = useCreateISPProfile();
  const update = useUpdateISPProfile();
  const del = useDeleteISPProfile();

  const [editId, setEditId] = useState<string | null>(null);
  const [isp, setIsp] = useState("");
  const [country, setCountry] = useState("IR");
  const [protocols, setProtocols] = useState("");

  const profiles = profilesQuery.data ?? [];

  function resetForm() {
    setEditId(null);
    setIsp("");
    setCountry("IR");
    setProtocols("");
  }

  function startEdit(p: ISPProfile) {
    setEditId(p.id);
    setIsp(p.isp_identifier);
    setCountry(p.country_code);
    setProtocols(p.preferred_protocols?.join(", ") ?? "");
  }

  useEffect(() => {
    if (open) resetForm();
  }, [open]);

  async function handleSubmit() {
    const protoList = protocols
      .split(",")
      .map((s) => s.trim())
      .filter(Boolean);

    try {
      if (editId) {
        await update.mutateAsync({
          id: editId,
          isp_identifier: isp,
          country_code: country,
          preferred_protocols: protoList,
        });
        toast.success(t("ispProfiles.updated"));
      } else {
        await create.mutateAsync({
          group_id: groupId,
          isp_identifier: isp,
          country_code: country,
          preferred_protocols: protoList,
        });
        toast.success(t("ispProfiles.created"));
      }
      resetForm();
    } catch (e: any) {
      toast.error(e?.message ?? "Failed");
    }
  }

  async function handleDelete(id: string) {
    try {
      await del.mutateAsync(id);
      toast.success(t("ispProfiles.deleted"));
      if (editId === id) resetForm();
    } catch (e: any) {
      toast.error(e?.message ?? "Failed");
    }
  }

  const isLoading = create.isPending || update.isPending;

  return (
    <Modal
      open={open}
      onClose={onClose}
      title={`${t("ispProfiles.title")} — ${groupName}`}
      className="max-w-lg"
    >
      <div className="space-y-4">
        {/* Existing profiles list */}
        {profiles.length > 0 ? (
          <div className="space-y-1 border border-border/40 rounded-lg p-2 bg-surface/40 max-h-40 overflow-y-auto">
            {profiles.map((p) => (
              <div
                key={p.id}
                className="flex items-center gap-2 py-1.5 px-2 rounded-md hover:bg-bg-elevated text-xs cursor-pointer transition"
                onClick={() => startEdit(p)}
              >
                <span className="font-medium text-fg flex-1">
                  {p.isp_identifier}
                  <span className="text-fg-subtle ms-1">({p.country_code})</span>
                </span>
                <span className="text-fg-subtle truncate max-w-[140px]">
                  {p.preferred_protocols?.join(", ")}
                </span>
                <button
                  type="button"
                  onClick={(e) => {
                    e.stopPropagation();
                    handleDelete(p.id);
                  }}
                  className="text-danger hover:text-danger/80 p-0.5"
                >
                  <Trash2 size={12} />
                </button>
              </div>
            ))}
          </div>
        ) : (
          <p className="text-xs text-fg-subtle text-center py-3">
            {t("ispProfiles.noProfiles")}
          </p>
        )}

        {/* Form */}
        <div className="border-t border-border/30 pt-3 space-y-3">
          <p className="text-xs font-medium text-fg-muted">
            {editId ? t("ispProfiles.edit") : t("ispProfiles.create")}
          </p>
          <div className="grid grid-cols-2 gap-3">
            <div>
              <label className="text-xs text-fg-muted mb-1 block">{t("ispProfiles.isp")}</label>
              <Input
                value={isp}
                onChange={(e) => setIsp(e.target.value)}
                placeholder={t("ispProfiles.ispHint")}
              />
            </div>
            <div>
              <label className="text-xs text-fg-muted mb-1 block">{t("ispProfiles.country")}</label>
              <Input
                value={country}
                onChange={(e) => setCountry(e.target.value)}
                placeholder={t("ispProfiles.countryHint")}
                maxLength={2}
              />
            </div>
          </div>
          <div>
            <label className="text-xs text-fg-muted mb-1 block">
              {t("ispProfiles.preferredProtocols")}
            </label>
            <Input
              value={protocols}
              onChange={(e) => setProtocols(e.target.value)}
              placeholder={t("ispProfiles.preferredHint")}
            />
          </div>
        </div>

        {/* Actions */}
        <div className="flex justify-end gap-2 pt-2">
          {editId && (
            <Button variant="ghost" size="sm" onClick={resetForm}>
              {t("common.cancel") || "Cancel"}
            </Button>
          )}
          <Button size="sm" onClick={handleSubmit} disabled={!isp || !protocols || isLoading}>
            {isLoading ? "..." : editId ? t("common.save") || "Save" : t("ispProfiles.create")}
          </Button>
        </div>
      </div>
    </Modal>
  );
}
