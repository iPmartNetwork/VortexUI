import { useState } from "react";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { Plus, Trash2, Users } from "lucide-react";
import { api } from "@/api/client";
import { Button, Input, Badge } from "@/components/ui";
import { Modal } from "@/components/Modal";
import { GlassCard } from "@/components/veltrix";
import { useToast } from "@/components/toast";
import { useConfirm } from "@/components/confirm";
import { formatBytes } from "@/lib/utils";
import { useTitle } from "@/lib/useTitle";
import { useI18n } from "@/i18n/i18n";

interface FamilyMember {
  id: string;
  user_id: string;
  username: string;
  used_traffic: number;
  label: string;
  joined_at: string;
}

interface FamilyGroup {
  id: string;
  name: string;
  owner_id: string;
  owner_name: string;
  data_limit: number;
  used_traffic: number;
  max_members: number;
  member_quota: number;
  members: FamilyMember[];
  created_at: string;
}

export function FamilyGroups() {
  const { t } = useI18n();
  useTitle(t("nav.familyGroups"));
  const qc = useQueryClient();
  const toast = useToast();
  const confirm = useConfirm();
  const [createOpen, setCreateOpen] = useState(false);
  const [viewId, setViewId] = useState<string | null>(null);

  const { data } = useQuery({
    queryKey: ["family-groups"],
    queryFn: () => api<{ groups: FamilyGroup[]; total: number }>("/api/families"),
  });

  const delMut = useMutation({
    mutationFn: (id: string) => api<void>(`/api/families/${id}`, { method: "DELETE" }),
    onSuccess: () => qc.invalidateQueries({ queryKey: ["family-groups"] }),
  });

  async function remove(g: FamilyGroup) {
    const ok = await confirm({
      title: `${t("family.deleteConfirmPrefix")} "${g.name}"?`,
      confirmLabel: t("common.delete"),
      destructive: true,
    });
    if (!ok) return;
    await delMut.mutateAsync(g.id);
    toast.success(t("common.deleted"));
  }

  const groups = data?.groups ?? [];

  return (
    <div className="space-y-5 animate-page-enter">
      <div className="flex flex-col lg:flex-row lg:items-start justify-between gap-4">
        <div>
          <h1 className="text-2xl font-bold text-fg tracking-tight">{t("nav.familyGroups")}</h1>
          <p className="text-sm text-fg-muted mt-1">{t("family.subtitle")}</p>
        </div>
        <Button onClick={() => setCreateOpen(true)} className="flex-shrink-0">
          <Plus size={14} /> {t("family.newGroup")}
        </Button>
      </div>

      <CreateGroupModal open={createOpen} onClose={() => setCreateOpen(false)} />
      {viewId && <GroupDetailModal groupId={viewId} onClose={() => setViewId(null)} />}

      <div className="grid grid-cols-1 gap-4 lg:grid-cols-2">
        {groups.map((g) => (
          <GlassCard key={g.id} hover className="!p-4 space-y-3" onClick={() => setViewId(g.id)}>
            <div className="flex items-center justify-between gap-2">
              <div className="flex items-center gap-2 min-w-0">
                <div className="h-9 w-9 rounded-xl bg-primary/10 flex items-center justify-center text-primary flex-shrink-0">
                  <Users size={16} />
                </div>
                <div className="min-w-0">
                  <h3 className="text-sm font-bold text-fg truncate">{g.name}</h3>
                  <p className="text-[11px] text-fg-subtle truncate">{t("family.owner")} {g.owner_name}</p>
                </div>
              </div>
              <Badge color="active">{g.members?.length ?? 0}/{g.max_members}</Badge>
            </div>
            <div className="grid grid-cols-2 gap-2 rounded-lg bg-surface-2/50 border border-border/40 px-3 py-2 text-xs">
              <div><span className="text-fg-subtle">{t("family.pool")}</span> <strong className="text-fg">{formatBytes(g.data_limit, false)}</strong></div>
              <div><span className="text-fg-subtle">{t("family.used")}</span> <strong className="text-fg">{formatBytes(g.used_traffic, false)}</strong></div>
            </div>
            <div className="flex justify-end pt-1 border-t border-border/40">
              <Button
                variant="ghost"
                size="sm"
                className="text-danger"
                onClick={(e) => {
                  e.stopPropagation();
                  remove(g);
                }}
              >
                <Trash2 size={13} /> {t("common.delete")}
              </Button>
            </div>
          </GlassCard>
        ))}
        {groups.length === 0 && (
          <p className="col-span-full text-center text-sm text-fg-muted py-8">{t("family.empty")}</p>
        )}
      </div>
    </div>
  );
}

function CreateGroupModal({ open, onClose }: { open: boolean; onClose: () => void }) {
  const { t } = useI18n();
  const qc = useQueryClient();
  const toast = useToast();
  const [f, setF] = useState({ name: "", owner_id: "", data_limit: "100", max_members: "5", member_quota: "0" });
  const [ownerSearch, setOwnerSearch] = useState("");

  const { data: usersData } = useQuery({
    queryKey: ["users-for-family", ownerSearch],
    queryFn: () => api<{ users: { id: string; username: string }[] }>("/api/users", { query: { search: ownerSearch, limit: 10 } }),
    enabled: open,
  });

  const create = useMutation({
    mutationFn: (input: Record<string, unknown>) => api("/api/families", { method: "POST", body: input }),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["family-groups"] });
      onClose();
      toast.success(t("family.groupCreated"));
    },
    onError: (e: unknown) => toast.error(e instanceof Error ? e.message : t("common.failed")),
  });

  return (
    <Modal open={open} onClose={onClose} title={t("family.newGroupTitle")}>
      <form
        onSubmit={(e) => {
          e.preventDefault();
          create.mutate({
            name: f.name,
            owner_id: f.owner_id,
            data_limit: Number(f.data_limit) * 1024 * 1024 * 1024,
            max_members: Number(f.max_members),
            member_quota: Number(f.member_quota) * 1024 * 1024 * 1024,
          });
        }}
        className="space-y-3"
      >
        <Input placeholder={t("family.groupName")} value={f.name} onChange={(e) => setF((s) => ({ ...s, name: e.target.value }))} required />
        <div>
          <Input placeholder={t("family.searchUser")} value={ownerSearch} onChange={(e) => setOwnerSearch(e.target.value)} />
          {usersData?.users && usersData.users.length > 0 && !f.owner_id && (
            <div className="mt-1 max-h-32 overflow-y-auto rounded-lg border border-border/40 bg-surface-2/40">
              {usersData.users.map((u) => (
                <button
                  key={u.id}
                  type="button"
                  className="w-full px-3 py-1.5 text-left text-xs text-fg hover:bg-primary/10"
                  onClick={() => {
                    setF((s) => ({ ...s, owner_id: u.id }));
                    setOwnerSearch(u.username);
                  }}
                >
                  {u.username} <span className="text-fg-muted">({u.id.slice(0, 8)})</span>
                </button>
              ))}
            </div>
          )}
          {f.owner_id && <p className="mt-1 text-[10px] text-fg-subtle">{t("family.selected")} {ownerSearch}</p>}
        </div>
        <div className="grid grid-cols-1 gap-2 sm:grid-cols-3">
          <Input placeholder={t("family.poolGb")} value={f.data_limit} onChange={(e) => setF((s) => ({ ...s, data_limit: e.target.value }))} inputMode="numeric" />
          <Input placeholder={t("family.maxMembers")} value={f.max_members} onChange={(e) => setF((s) => ({ ...s, max_members: e.target.value }))} inputMode="numeric" />
          <Input placeholder={t("family.memberCapGb")} value={f.member_quota} onChange={(e) => setF((s) => ({ ...s, member_quota: e.target.value }))} inputMode="numeric" />
        </div>
        <div className="flex justify-end gap-2 pt-2">
          <Button type="button" variant="ghost" onClick={onClose}>{t("common.cancel")}</Button>
          <Button type="submit" disabled={create.isPending}>{t("common.create")}</Button>
        </div>
      </form>
    </Modal>
  );
}

function GroupDetailModal({ groupId, onClose }: { groupId: string; onClose: () => void }) {
  const { t } = useI18n();
  const qc = useQueryClient();
  const toast = useToast();
  const [newUser, setNewUser] = useState({ user_id: "", label: "" });

  const { data } = useQuery({
    queryKey: ["family-group", groupId],
    queryFn: () => api<{ group: FamilyGroup }>(`/api/families/${groupId}`),
  });

  const addMut = useMutation({
    mutationFn: (input: { user_id: string; label: string }) => api(`/api/families/${groupId}/members`, { method: "POST", body: input }),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["family-group", groupId] });
      setNewUser({ user_id: "", label: "" });
      toast.success(t("family.memberAdded"));
    },
    onError: (e: unknown) => toast.error(e instanceof Error ? e.message : t("common.failed")),
  });

  const removeMut = useMutation({
    mutationFn: (uid: string) => api<void>(`/api/families/${groupId}/members/${uid}`, { method: "DELETE" }),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["family-group", groupId] });
      toast.success(t("family.removed"));
    },
  });

  const group = data?.group;

  return (
    <Modal open onClose={onClose} title={group?.name || t("family.groupDefault")} className="max-w-lg">
      <div className="space-y-4">
        <div className="text-xs text-fg-muted">
          {t("family.owner")} {group?.owner_name} | {t("family.membersLabel")} {group?.members?.length ?? 0}/{group?.max_members}
        </div>
        <div className="space-y-2 max-h-[200px] overflow-y-auto">
          {group?.members?.map((m) => (
            <div key={m.id} className="flex items-center justify-between rounded-lg bg-surface-2/40 px-3 py-2">
              <div>
                <span className="text-sm font-medium text-fg">{m.username}</span>
                {m.label && <span className="ml-2 text-xs text-fg-muted">({m.label})</span>}
                <div className="text-xs text-fg-subtle">{formatBytes(m.used_traffic, false)}</div>
              </div>
              <Button variant="ghost" size="sm" className="text-danger" onClick={() => removeMut.mutate(m.user_id)}>
                <Trash2 size={13} />
              </Button>
            </div>
          ))}
        </div>
        <form
          onSubmit={(e) => {
            e.preventDefault();
            addMut.mutate(newUser);
          }}
          className="flex gap-2"
        >
          <Input placeholder={t("family.userId")} value={newUser.user_id} onChange={(e) => setNewUser((s) => ({ ...s, user_id: e.target.value }))} className="flex-1" />
          <Input placeholder={t("family.label")} value={newUser.label} onChange={(e) => setNewUser((s) => ({ ...s, label: e.target.value }))} className="w-28" />
          <Button type="submit" size="sm" disabled={addMut.isPending || !newUser.user_id}>{t("common.add")}</Button>
        </form>
      </div>
    </Modal>
  );
}
