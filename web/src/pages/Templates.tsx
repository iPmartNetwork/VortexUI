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

const defaultForm: TemplateForm = { name: "", data_limit: 0, expire_duration: 0, device_limit: 0, reset_strategy: "no_reset", note: "" };

function formatBytes(bytes: number): string {
  if (bytes === 0) return "Unlimited";
  const gb = bytes / (1024 * 1024 * 1024);
  return gb >= 1 ? `${gb.toFixed(1)} GB` : `${(bytes / (1024 * 1024)).toFixed(0)} MB`;
}

function formatDuration(seconds?: number): string {
  if (!seconds) return "Never";
  const days = Math.floor(seconds / 86400);
  return days > 0 ? `${days} days` : `${Math.floor(seconds / 3600)} hours`;
}

export function Templates() {
  const { t: _t } = useI18n();
  useTitle("User Templates");
  const qc = useQueryClient();
  const toast = useToast();
  const confirm = useConfirm();

  const [showForm, setShowForm] = useState(false);
  const [editing, setEditing] = useState<UserTemplate | null>(null);
  const [form, setForm] = useState<TemplateForm>(defaultForm);
  const [showBulk, setShowBulk] = useState<string | null>(null);
  const [bulkCount, setBulkCount] = useState(10);

  const { data: templates, isLoading } = useQuery({
    queryKey: ["user-templates"],
    queryFn: () => api<UserTemplate[]>("/api/v2/templates"),
  });

  const createMut = useMutation({
    mutationFn: (body: TemplateForm) => api("/api/v2/templates", { method: "POST", body }),
    onSuccess: () => { qc.invalidateQueries({ queryKey: ["user-templates"] }); toast.success("Template created"); resetForm(); },
  });

  const updateMut = useMutation({
    mutationFn: ({ id, body }: { id: string; body: TemplateForm }) => api(`/api/v2/templates/${id}`, { method: "PUT", body }),
    onSuccess: () => { qc.invalidateQueries({ queryKey: ["user-templates"] }); toast.success("Template updated"); resetForm(); },
  });

  const deleteMut = useMutation({
    mutationFn: (id: string) => api(`/api/v2/templates/${id}`, { method: "DELETE" }),
    onSuccess: () => { qc.invalidateQueries({ queryKey: ["user-templates"] }); toast.success("Template deleted"); },
  });

  const bulkMut = useMutation({
    mutationFn: ({ id, count }: { id: string; count: number }) => api(`/api/v2/templates/${id}/bulk-create`, { method: "POST", body: { count } }),
    onSuccess: () => { toast.success("Users created"); setShowBulk(null); },
  });

  const resetForm = () => { setForm(defaultForm); setEditing(null); setShowForm(false); };

  const openEdit = (tmpl: UserTemplate) => {
    setEditing(tmpl);
    setForm({ name: tmpl.name, data_limit: tmpl.data_limit, expire_duration: tmpl.expire_duration || 0, device_limit: tmpl.device_limit, reset_strategy: tmpl.reset_strategy, note: tmpl.note });
    setShowForm(true);
  };

  return (
    <div className="space-y-6 p-6">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-bold">User Templates</h1>
        <Button onClick={() => { setEditing(null); setForm(defaultForm); setShowForm(true); }}><Plus className="w-4 h-4 mr-1" />Create</Button>
      </div>

      <GlassCard className="p-4">
        {isLoading ? <p className="text-fg-muted">Loading...</p> : templates && templates.length > 0 ? (
          <div className="overflow-x-auto">
            <table className="w-full text-sm">
              <thead><tr className="border-b"><th className="text-left py-2 px-3">Name</th><th className="text-left py-2 px-3">Data</th><th className="text-left py-2 px-3">Duration</th><th className="text-left py-2 px-3">Devices</th><th className="text-left py-2 px-3">Reset</th><th className="text-right py-2 px-3">Actions</th></tr></thead>
              <tbody>
                {templates.map((tmpl) => (
                  <tr key={tmpl.id} className="border-b hover:bg-surface-2/50">
                    <td className="py-2 px-3 font-medium">{tmpl.name}</td>
                    <td className="py-2 px-3">{formatBytes(tmpl.data_limit)}</td>
                    <td className="py-2 px-3">{formatDuration(tmpl.expire_duration)}</td>
                    <td className="py-2 px-3">{tmpl.device_limit || "∞"}</td>
                    <td className="py-2 px-3"><Badge>{tmpl.reset_strategy}</Badge></td>
                    <td className="py-2 px-3 text-right space-x-1">
                      <Button variant="ghost" size="sm" onClick={() => openEdit(tmpl)}><Pencil className="w-3 h-3" /></Button>
                      <Button variant="ghost" size="sm" onClick={() => setShowBulk(tmpl.id)}><Users className="w-3 h-3" /></Button>
                      <Button variant="ghost" size="sm" onClick={async () => { if (await confirm({ title: "Delete this template?" })) deleteMut.mutate(tmpl.id); }}><Trash2 className="w-3 h-3" /></Button>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        ) : <p className="text-fg-muted text-sm">No templates yet.</p>}
      </GlassCard>

      <Modal open={showForm} onClose={resetForm} title={editing ? "Edit Template" : "Create Template"}>
        <div className="space-y-4">
          <Input placeholder="Template name" value={form.name} onChange={(e) => setForm({ ...form, name: e.target.value })} />
          <div className="grid grid-cols-2 gap-4">
            <Input type="number" placeholder="Data (GB)" value={form.data_limit / (1024*1024*1024)} onChange={(e) => setForm({ ...form, data_limit: Number(e.target.value) * 1024*1024*1024 })} />
            <Input type="number" placeholder="Duration (days)" value={form.expire_duration / 86400} onChange={(e) => setForm({ ...form, expire_duration: Number(e.target.value) * 86400 })} />
          </div>
          <div className="grid grid-cols-2 gap-4">
            <Input type="number" placeholder="Device limit" value={form.device_limit} onChange={(e) => setForm({ ...form, device_limit: Number(e.target.value) })} />
            <select className="field input-surface" value={form.reset_strategy} onChange={(e) => setForm({ ...form, reset_strategy: e.target.value })}>
              <option value="no_reset">No Reset</option><option value="daily">Daily</option><option value="weekly">Weekly</option><option value="monthly">Monthly</option>
            </select>
          </div>
          <Input placeholder="Note (optional)" value={form.note} onChange={(e) => setForm({ ...form, note: e.target.value })} />
          <Button onClick={() => editing ? updateMut.mutate({ id: editing.id, body: form }) : createMut.mutate(form)} disabled={!form.name}>{editing ? "Save" : "Create"}</Button>
        </div>
      </Modal>

      <Modal open={!!showBulk} onClose={() => setShowBulk(null)} title="Bulk Create Users">
        <div className="space-y-4">
          <Input type="number" min={1} max={1000} value={bulkCount} onChange={(e) => setBulkCount(Number(e.target.value))} />
          <Button onClick={() => showBulk && bulkMut.mutate({ id: showBulk, count: bulkCount })} disabled={bulkCount < 1 || bulkCount > 1000}><Copy className="w-4 h-4 mr-1" />Create {bulkCount} Users</Button>
        </div>
      </Modal>
    </div>
  );
}
