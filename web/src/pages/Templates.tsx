import { useState } from "react";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { Plus, Trash2, Pencil, Copy, Users } from "lucide-react";
import { api } from "@/api/client";
import { Button, Input, Badge } from "@/components/ui";
import { GlassCard } from "@/components/veltrix";
import { Modal } from "@/components/Modal";
import { useToast } from "@/components/toast";
import { useConfirm } from "@/components/confirm";
import { useTitle } from "@/lib/useTitle";
import { useI18n } from "@/i18n/i18n";

// Types
interface UserTemplate {
  id: string;
  name: string;
  data_limit: number;
  expire_duration?: number;
  device_limit: number;
  reset_strategy: string;
  note: string;
  groups: string[];
  created_at: string;
  updated_at: string;
}

interface TemplateForm {
  name: string;
  data_limit: number;
  expire_duration: number;
  device_limit: number;
  reset_strategy: string;
  note: string;
}

const defaultForm: TemplateForm = {
  name: "",
  data_limit: 0,
  expire_duration: 0,
  device_limit: 0,
  reset_strategy: "no_reset",
  note: "",
};

function formatBytes(bytes: number): string {
  if (bytes === 0) return "Unlimited";
  const gb = bytes / (1024 * 1024 * 1024);
  if (gb >= 1) return `${gb.toFixed(1)} GB`;
  const mb = bytes / (1024 * 1024);
  return `${mb.toFixed(0)} MB`;
}

function formatDuration(seconds?: number): string {
  if (!seconds) return "Never";
  const days = Math.floor(seconds / 86400);
  if (days > 0) return `${days} days`;
  const hours = Math.floor(seconds / 3600);
  return `${hours} hours`;
}

export function Templates() {
  const { t } = useI18n();
  useTitle(t("templates.title", "User Templates"));
  const queryClient = useQueryClient();
  const toast = useToast();
  const confirm = useConfirm();

  const [showForm, setShowForm] = useState(false);
  const [editing, setEditing] = useState<UserTemplate | null>(null);
  const [form, setForm] = useState<TemplateForm>(defaultForm);
  const [showBulk, setShowBulk] = useState<string | null>(null);
  const [bulkCount, setBulkCount] = useState(10);

  // Fetch templates
  const { data: templates, isLoading } = useQuery({
    queryKey: ["user-templates"],
    queryFn: () =>
      api.get("/api/v2/templates").then((r) => r.data as UserTemplate[]),
  });

  // Create
  const createMutation = useMutation({
    mutationFn: (body: TemplateForm) =>
      api.post("/api/v2/templates", body),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["user-templates"] });
      toast.success(t("templates.created", "Template created"));
      resetForm();
    },
  });

  // Update
  const updateMutation = useMutation({
    mutationFn: ({ id, body }: { id: string; body: TemplateForm }) =>
      api.put(`/api/v2/templates/${id}`, body),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["user-templates"] });
      toast.success(t("templates.updated", "Template updated"));
      resetForm();
    },
  });

  // Delete
  const deleteMutation = useMutation({
    mutationFn: (id: string) => api.delete(`/api/v2/templates/${id}`),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["user-templates"] });
      toast.success(t("templates.deleted", "Template deleted"));
    },
  });

  // Bulk create
  const bulkCreateMutation = useMutation({
    mutationFn: ({ id, count }: { id: string; count: number }) =>
      api.post(`/api/v2/templates/${id}/bulk-create`, { count }),
    onSuccess: (res) => {
      toast.success(t("templates.bulkCreated", "Users created successfully"));
      setShowBulk(null);
    },
  });

  const resetForm = () => {
    setForm(defaultForm);
    setEditing(null);
    setShowForm(false);
  };

  const openEdit = (tmpl: UserTemplate) => {
    setEditing(tmpl);
    setForm({
      name: tmpl.name,
      data_limit: tmpl.data_limit,
      expire_duration: tmpl.expire_duration || 0,
      device_limit: tmpl.device_limit,
      reset_strategy: tmpl.reset_strategy,
      note: tmpl.note,
    });
    setShowForm(true);
  };

  const handleSubmit = () => {
    if (editing) {
      updateMutation.mutate({ id: editing.id, body: form });
    } else {
      createMutation.mutate(form);
    }
  };

  return (
    <div className="space-y-6 p-6">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-bold">
          {t("templates.title", "User Templates")}
        </h1>
        <Button onClick={() => { setEditing(null); setForm(defaultForm); setShowForm(true); }}>
          <Plus className="w-4 h-4 mr-1" />
          {t("templates.create", "Create Template")}
        </Button>
      </div>

      {/* Template list */}
      <GlassCard className="p-4">
        {isLoading ? (
          <p className="text-muted-foreground">{t("common.loading", "Loading...")}</p>
        ) : templates && templates.length > 0 ? (
          <div className="overflow-x-auto">
            <table className="w-full text-sm">
              <thead>
                <tr className="border-b">
                  <th className="text-left py-2 px-3">Name</th>
                  <th className="text-left py-2 px-3">Data Limit</th>
                  <th className="text-left py-2 px-3">Duration</th>
                  <th className="text-left py-2 px-3">Devices</th>
                  <th className="text-left py-2 px-3">Reset</th>
                  <th className="text-right py-2 px-3">Actions</th>
                </tr>
              </thead>
              <tbody>
                {templates.map((tmpl) => (
                  <tr key={tmpl.id} className="border-b hover:bg-muted/50">
                    <td className="py-2 px-3 font-medium">{tmpl.name}</td>
                    <td className="py-2 px-3">{formatBytes(tmpl.data_limit)}</td>
                    <td className="py-2 px-3">{formatDuration(tmpl.expire_duration)}</td>
                    <td className="py-2 px-3">{tmpl.device_limit || "∞"}</td>
                    <td className="py-2 px-3">
                      <Badge variant="outline">{tmpl.reset_strategy}</Badge>
                    </td>
                    <td className="py-2 px-3 text-right space-x-1">
                      <Button variant="ghost" size="sm" onClick={() => openEdit(tmpl)}>
                        <Pencil className="w-3 h-3" />
                      </Button>
                      <Button variant="ghost" size="sm" onClick={() => setShowBulk(tmpl.id)}>
                        <Users className="w-3 h-3" />
                      </Button>
                      <Button
                        variant="ghost"
                        size="sm"
                        onClick={async () => {
                          const ok = await confirm(t("templates.confirmDelete", "Delete this template?"));
                          if (ok) deleteMutation.mutate(tmpl.id);
                        }}
                      >
                        <Trash2 className="w-3 h-3" />
                      </Button>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        ) : (
          <p className="text-muted-foreground text-sm">
            {t("templates.empty", "No templates yet. Create one to get started.")}
          </p>
        )}
      </GlassCard>

      {/* Create/Edit Modal */}
      <Modal open={showForm} onClose={resetForm} title={editing ? "Edit Template" : "Create Template"}>
        <div className="space-y-4">
          <div>
            <label className="block text-sm font-medium mb-1">Name</label>
            <Input value={form.name} onChange={(e) => setForm({ ...form, name: e.target.value })} placeholder="e.g. Monthly 50GB" />
          </div>
          <div className="grid grid-cols-2 gap-4">
            <div>
              <label className="block text-sm font-medium mb-1">Data Limit (GB)</label>
              <Input type="number" value={form.data_limit / (1024*1024*1024)} onChange={(e) => setForm({ ...form, data_limit: Number(e.target.value) * 1024*1024*1024 })} />
            </div>
            <div>
              <label className="block text-sm font-medium mb-1">Duration (days)</label>
              <Input type="number" value={form.expire_duration / 86400} onChange={(e) => setForm({ ...form, expire_duration: Number(e.target.value) * 86400 })} />
            </div>
          </div>
          <div className="grid grid-cols-2 gap-4">
            <div>
              <label className="block text-sm font-medium mb-1">Device Limit</label>
              <Input type="number" value={form.device_limit} onChange={(e) => setForm({ ...form, device_limit: Number(e.target.value) })} />
            </div>
            <div>
              <label className="block text-sm font-medium mb-1">Reset Strategy</label>
              <select className="w-full rounded border p-2 bg-background" value={form.reset_strategy} onChange={(e) => setForm({ ...form, reset_strategy: e.target.value })}>
                <option value="no_reset">No Reset</option>
                <option value="daily">Daily</option>
                <option value="weekly">Weekly</option>
                <option value="monthly">Monthly</option>
                <option value="yearly">Yearly</option>
              </select>
            </div>
          </div>
          <div>
            <label className="block text-sm font-medium mb-1">Note</label>
            <Input value={form.note} onChange={(e) => setForm({ ...form, note: e.target.value })} placeholder="Optional note" />
          </div>
          <Button onClick={handleSubmit} disabled={!form.name || createMutation.isPending || updateMutation.isPending}>
            {editing ? t("common.save", "Save") : t("templates.create", "Create Template")}
          </Button>
        </div>
      </Modal>

      {/* Bulk Create Modal */}
      <Modal open={!!showBulk} onClose={() => setShowBulk(null)} title={t("templates.bulkCreate", "Bulk Create Users")}>
        <div className="space-y-4">
          <div>
            <label className="block text-sm font-medium mb-1">{t("templates.userCount", "Number of users (1-1000)")}</label>
            <Input type="number" min={1} max={1000} value={bulkCount} onChange={(e) => setBulkCount(Number(e.target.value))} />
          </div>
          <Button onClick={() => showBulk && bulkCreateMutation.mutate({ id: showBulk, count: bulkCount })} disabled={bulkCount < 1 || bulkCount > 1000 || bulkCreateMutation.isPending}>
            <Copy className="w-4 h-4 mr-1" />
            {t("templates.bulkCreate", "Bulk Create Users")}
          </Button>
        </div>
      </Modal>
    </div>
  );
}
