import { useState, useMemo } from "react";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { Plus, FileText, Trash2, Pencil, Users, Search } from "lucide-react";
import { api } from "@/api/client";
import { Button, Input, PageHeader, Badge } from "@/components/ui";
import { GlassCard } from "@/components/veltrix";
import { EmptyState } from "@/components/EmptyState";
import { Modal } from "@/components/Modal";
import { useTitle } from "@/lib/useTitle";
import { formatBytes } from "@/lib/utils";
import { TemplateForm, type TemplateFormData } from "./TemplateForm";
import { BulkCreateDialog } from "@/components/BulkCreateDialog";

// --- Types ---

export interface UserTemplate {
  id: string;
  name: string;
  data_limit: number;
  expire_duration: number | null;
  device_limit: number;
  reset_strategy: string;
  note: string;
  protocol_settings: Record<string, unknown>;
  groups: string[];
  allowed_admins: string[] | null;
  created_at: string;
  updated_at: string;
}

// --- API Hooks ---

function useTemplates() {
  return useQuery({
    queryKey: ["templates"],
    queryFn: () => api<{ templates: UserTemplate[] }>("/api/v2/templates"),
    select: (d) => d.templates ?? [],
  });
}

function useCreateTemplate() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (input: TemplateFormData) =>
      api<{ template: UserTemplate }>("/api/v2/templates", { method: "POST", body: input }),
    onSuccess: () => qc.invalidateQueries({ queryKey: ["templates"] }),
  });
}

function useUpdateTemplate() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ id, input }: { id: string; input: TemplateFormData }) =>
      api<{ template: UserTemplate }>(`/api/v2/templates/${id}`, { method: "PUT", body: input }),
    onSuccess: () => qc.invalidateQueries({ queryKey: ["templates"] }),
  });
}

function useDeleteTemplate() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => api<void>(`/api/v2/templates/${id}`, { method: "DELETE" }),
    onSuccess: () => qc.invalidateQueries({ queryKey: ["templates"] }),
  });
}

// --- Helpers ---

function formatDuration(seconds: number | null): string {
  if (!seconds) return "Never";
  const days = Math.round(seconds / 86400);
  if (days >= 365) return `${Math.round(days / 365)}y`;
  if (days >= 30) return `${Math.round(days / 30)}mo`;
  return `${days}d`;
}

function resetStrategyLabel(s: string): string {
  switch (s) {
    case "no_reset": return "No Reset";
    case "daily": return "Daily";
    case "weekly": return "Weekly";
    case "monthly": return "Monthly";
    default: return s;
  }
}

// --- Page Component ---

export function TemplateListPage() {
  useTitle("Templates");

  const templates = useTemplates();
  const deleteMutation = useDeleteTemplate();

  const [search, setSearch] = useState("");
  const [formOpen, setFormOpen] = useState(false);
  const [editing, setEditing] = useState<UserTemplate | null>(null);
  const [bulkTarget, setBulkTarget] = useState<{ id: string; name: string } | null>(null);
  const [deleteConfirm, setDeleteConfirm] = useState<string | null>(null);

  const filtered = useMemo(() => {
    const list = templates.data ?? [];
    if (!search) return list;
    const q = search.toLowerCase();
    return list.filter(
      (t) =>
        t.name.toLowerCase().includes(q) ||
        t.groups.some((g) => g.toLowerCase().includes(q)),
    );
  }, [templates.data, search]);

  function openCreate() {
    setEditing(null);
    setFormOpen(true);
  }

  function openEdit(t: UserTemplate) {
    setEditing(t);
    setFormOpen(true);
  }

  function closeForm() {
    setFormOpen(false);
    setEditing(null);
  }

  function handleDelete(id: string) {
    deleteMutation.mutate(id, { onSuccess: () => setDeleteConfirm(null) });
  }

  return (
    <div className="space-y-5 animate-page-enter">
      {/* Header */}
      <PageHeader title="User Templates" subtitle="Reusable blueprints for user provisioning">
        <Button onClick={openCreate}>
          <Plus size={15} /> New Template
        </Button>
      </PageHeader>

      {/* Search */}
      <GlassCard hover={false} className="!p-4">
        <div className="relative">
          <Search size={14} className="absolute start-3 top-1/2 -translate-y-1/2 text-fg-subtle pointer-events-none" />
          <Input
            placeholder="Search templates by name or group..."
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            className="ps-9"
          />
        </div>
      </GlassCard>

      {/* Loading */}
      {templates.isLoading && (
        <GlassCard hover={false} className="!p-8">
          <div className="space-y-4">
            {Array.from({ length: 3 }).map((_, i) => (
              <div key={i} className="flex items-center gap-4 animate-shimmer bg-gradient-to-r from-surface-2/40 via-surface-2/80 to-surface-2/40 bg-[length:200%_100%] rounded-xl h-16" />
            ))}
          </div>
        </GlassCard>
      )}

      {/* Empty state */}
      {!templates.isLoading && filtered.length === 0 && (
        <GlassCard hover={false}>
          <EmptyState
            icon={FileText}
            title={search ? "No templates match your search" : "No templates yet"}
            description={search ? "Try adjusting your search." : "Create your first template to quickly provision users with consistent settings."}
            action={!search ? { label: "Create Template", onClick: openCreate } : undefined}
          />
        </GlassCard>
      )}

      {/* Template table */}
      {!templates.isLoading && filtered.length > 0 && (
        <GlassCard hover={false} className="!p-0 overflow-hidden">
          <div className="overflow-x-auto">
            <table className="w-full text-sm">
              <thead>
                <tr className="border-b border-border/40 text-xs font-semibold uppercase tracking-wider text-fg-subtle">
                  <th className="px-4 py-3 text-start">Name</th>
                  <th className="px-4 py-3 text-start">Data Limit</th>
                  <th className="px-4 py-3 text-start">Expire</th>
                  <th className="px-4 py-3 text-start">Devices</th>
                  <th className="px-4 py-3 text-start">Reset</th>
                  <th className="px-4 py-3 text-start">Groups</th>
                  <th className="px-4 py-3 text-end">Actions</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-border/20">
                {filtered.map((t) => (
                  <tr key={t.id} className="hover:bg-surface/30 transition-colors">
                    <td className="px-4 py-3 font-medium text-fg">{t.name}</td>
                    <td className="px-4 py-3 text-fg-muted tabular-nums">
                      {formatBytes(t.data_limit)}
                    </td>
                    <td className="px-4 py-3 text-fg-muted tabular-nums">
                      {formatDuration(t.expire_duration)}
                    </td>
                    <td className="px-4 py-3 text-fg-muted tabular-nums">
                      {t.device_limit === 0 ? "∞" : t.device_limit}
                    </td>
                    <td className="px-4 py-3">
                      <Badge color="muted">{resetStrategyLabel(t.reset_strategy)}</Badge>
                    </td>
                    <td className="px-4 py-3">
                      <div className="flex flex-wrap gap-1">
                        {t.groups.length === 0 && <span className="text-fg-subtle text-xs">—</span>}
                        {t.groups.slice(0, 3).map((g) => (
                          <Badge key={g} color="on_hold">{g}</Badge>
                        ))}
                        {t.groups.length > 3 && (
                          <span className="text-xs text-fg-muted">+{t.groups.length - 3}</span>
                        )}
                      </div>
                    </td>
                    <td className="px-4 py-3">
                      <div className="flex items-center justify-end gap-1">
                        <button
                          type="button"
                          onClick={() => openEdit(t)}
                          className="h-8 w-8 rounded-lg flex items-center justify-center text-fg-subtle hover:text-fg hover:bg-surface-2/60 transition-all"
                          title="Edit template"
                        >
                          <Pencil size={14} />
                        </button>
                        <button
                          type="button"
                          onClick={() => setBulkTarget({ id: t.id, name: t.name })}
                          className="h-8 w-8 rounded-lg flex items-center justify-center text-fg-subtle hover:text-primary hover:bg-primary/10 transition-all"
                          title="Bulk create users"
                        >
                          <Users size={14} />
                        </button>
                        <button
                          type="button"
                          onClick={() => setDeleteConfirm(t.id)}
                          className="h-8 w-8 rounded-lg flex items-center justify-center text-fg-subtle hover:text-danger hover:bg-danger/10 transition-all"
                          title="Delete template"
                        >
                          <Trash2 size={14} />
                        </button>
                      </div>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
          <div className="border-t border-border/40 px-4 py-3 text-xs text-fg-muted">
            {filtered.length} template{filtered.length !== 1 ? "s" : ""}
          </div>
        </GlassCard>
      )}

      {/* Create/Edit Form Modal */}
      <TemplateFormModal
        open={formOpen}
        onClose={closeForm}
        editing={editing}
      />

      {/* Bulk Create Dialog */}
      {bulkTarget && (
        <BulkCreateDialog
          open={!!bulkTarget}
          onClose={() => setBulkTarget(null)}
          templateId={bulkTarget.id}
          templateName={bulkTarget.name}
        />
      )}

      {/* Delete Confirmation */}
      <Modal open={!!deleteConfirm} onClose={() => setDeleteConfirm(null)} title="Delete Template">
        <div className="space-y-4">
          <p className="text-sm text-fg-muted">
            Are you sure you want to delete this template? This action cannot be undone.
          </p>
          <div className="flex justify-end gap-2">
            <Button variant="ghost" onClick={() => setDeleteConfirm(null)}>Cancel</Button>
            <Button
              variant="destructive"
              onClick={() => deleteConfirm && handleDelete(deleteConfirm)}
              disabled={deleteMutation.isPending}
            >
              {deleteMutation.isPending ? "Deleting..." : "Delete"}
            </Button>
          </div>
        </div>
      </Modal>
    </div>
  );
}

// --- Template Form Modal Wrapper ---

function TemplateFormModal({
  open,
  onClose,
  editing,
}: {
  open: boolean;
  onClose: () => void;
  editing: UserTemplate | null;
}) {
  const createMutation = useCreateTemplate();
  const updateMutation = useUpdateTemplate();

  async function handleSubmit(data: TemplateFormData) {
    if (editing) {
      await updateMutation.mutateAsync({ id: editing.id, input: data });
    } else {
      await createMutation.mutateAsync(data);
    }
    onClose();
  }

  const isPending = createMutation.isPending || updateMutation.isPending;

  return (
    <Modal
      open={open}
      onClose={onClose}
      title={editing ? "Edit Template" : "Create Template"}
      className="max-w-lg"
    >
      <TemplateForm
        initialData={editing}
        onSubmit={handleSubmit}
        onCancel={onClose}
        isPending={isPending}
      />
    </Modal>
  );
}
