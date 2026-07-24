import { useState } from "react";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { Plus, Trash2, Pencil, Check, X, Eye } from "lucide-react";
import { api } from "@/api/client";
import { Button, Input, Badge } from "@/components/ui";
import { GlassCard } from "@/components/veltrix";
import { Modal } from "@/components/Modal";
import { useToast } from "@/components/toast";
import { useConfirm } from "@/components/confirm";
import { useTitle } from "@/lib/useTitle";

// Types
interface ClientTemplate {
  id: string;
  name: string;
  client_pattern: string;
  routing_rules: unknown[];
  dns_settings: Record<string, unknown>;
  custom_outbounds: unknown[];
  priority: number;
  enabled: boolean;
  created_at: string;
  updated_at: string;
}

interface SubscriptionApproval {
  id: string;
  user_id: string;
  request_data: Record<string, unknown>;
  status: string;
  admin_id?: string;
  created_at: string;
  resolved_at?: string;
}

interface TemplateForm {
  name: string;
  client_pattern: string;
  routing_rules: string;
  dns_settings: string;
  custom_outbounds: string;
  priority: number;
  enabled: boolean;
}

interface TemplatesResponse {
  templates: ClientTemplate[];
}

interface ApprovalsResponse {
  approvals: SubscriptionApproval[];
}

interface PreviewResponse {
  preview: string;
}

const defaultForm: TemplateForm = {
  name: "",
  client_pattern: "",
  routing_rules: "[]",
  dns_settings: "{}",
  custom_outbounds: "[]",
  priority: 0,
  enabled: true,
};

export function ClientTemplates() {
  useTitle("Client Templates");
  const queryClient = useQueryClient();
  const toast = useToast();
  const confirm = useConfirm();

  const [showForm, setShowForm] = useState(false);
  const [editingTemplate, setEditingTemplate] = useState<ClientTemplate | null>(null);
  const [form, setForm] = useState<TemplateForm>(defaultForm);
  const [activeTab, setActiveTab] = useState<"templates" | "approvals" | "preview">("templates");
  const [previewTemplate, setPreviewTemplate] = useState("");
  const [previewResult, setPreviewResult] = useState("");

  // --- Queries ---
  const { data: templates = [] } = useQuery<ClientTemplate[]>({
    queryKey: ["client-templates"],
    queryFn: async () => {
      const res = await api<TemplatesResponse>("/api/v2/client-templates");
      return res.templates;
    },
  });

  const { data: approvals = [] } = useQuery<SubscriptionApproval[]>({
    queryKey: ["approvals"],
    queryFn: async () => {
      const res = await api<ApprovalsResponse>("/api/v2/approvals");
      return res.approvals;
    },
    enabled: activeTab === "approvals",
  });

  // --- Mutations ---
  const createMutation = useMutation({
    mutationFn: async (data: TemplateForm) => {
      return api("/api/v2/client-templates", {
        method: "POST",
        body: {
          name: data.name,
          client_pattern: data.client_pattern,
          routing_rules: JSON.parse(data.routing_rules),
          dns_settings: JSON.parse(data.dns_settings),
          custom_outbounds: JSON.parse(data.custom_outbounds),
          priority: data.priority,
          enabled: data.enabled,
        },
      });
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["client-templates"] });
      toast.success("Template created");
      closeForm();
    },
    onError: () => toast.error("Failed to create template"),
  });

  const updateMutation = useMutation({
    mutationFn: async ({ id, data }: { id: string; data: TemplateForm }) => {
      return api(`/api/v2/client-templates/${id}`, {
        method: "PUT",
        body: {
          name: data.name,
          client_pattern: data.client_pattern,
          routing_rules: JSON.parse(data.routing_rules),
          dns_settings: JSON.parse(data.dns_settings),
          custom_outbounds: JSON.parse(data.custom_outbounds),
          priority: data.priority,
          enabled: data.enabled,
        },
      });
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["client-templates"] });
      toast.success("Template updated");
      closeForm();
    },
    onError: () => toast.error("Failed to update template"),
  });

  const deleteMutation = useMutation({
    mutationFn: (id: string) =>
      api(`/api/v2/client-templates/${id}`, { method: "DELETE" }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["client-templates"] });
      toast.success("Template deleted");
    },
    onError: () => toast.error("Failed to delete template"),
  });

  const approveMutation = useMutation({
    mutationFn: (id: string) =>
      api(`/api/v2/approvals/${id}/approve`, { method: "POST" }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["approvals"] });
      toast.success("Request approved");
    },
    onError: () => toast.error("Failed to approve"),
  });

  const rejectMutation = useMutation({
    mutationFn: (id: string) =>
      api(`/api/v2/approvals/${id}/reject`, { method: "POST" }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["approvals"] });
      toast.success("Request rejected");
    },
    onError: () => toast.error("Failed to reject"),
  });

  const previewMutation = useMutation({
    mutationFn: async (template: string) => {
      const res = await api<PreviewResponse>("/api/v2/templates/preview", {
        method: "POST",
        body: { template },
      });
      return res.preview;
    },
    onSuccess: (result) => setPreviewResult(result),
    onError: () => toast.error("Preview failed"),
  });

  // --- Helpers ---
  function closeForm() {
    setShowForm(false);
    setEditingTemplate(null);
    setForm(defaultForm);
  }

  function openEdit(tpl: ClientTemplate) {
    setEditingTemplate(tpl);
    setForm({
      name: tpl.name,
      client_pattern: tpl.client_pattern,
      routing_rules: JSON.stringify(tpl.routing_rules, null, 2),
      dns_settings: JSON.stringify(tpl.dns_settings, null, 2),
      custom_outbounds: JSON.stringify(tpl.custom_outbounds, null, 2),
      priority: tpl.priority,
      enabled: tpl.enabled,
    });
    setShowForm(true);
  }

  function handleSubmit() {
    if (editingTemplate) {
      updateMutation.mutate({ id: editingTemplate.id, data: form });
    } else {
      createMutation.mutate(form);
    }
  }

  async function handleDelete(id: string) {
    const ok = await confirm({ title: "Delete this template?" });
    if (ok) deleteMutation.mutate(id);
  }

  return (
    <div className="space-y-6">
      {/* Tab navigation */}
      <div className="flex gap-2">
        <Button
          variant={activeTab === "templates" ? "primary" : "ghost"}
          onClick={() => setActiveTab("templates")}
        >
          Templates
        </Button>
        <Button
          variant={activeTab === "approvals" ? "primary" : "ghost"}
          onClick={() => setActiveTab("approvals")}
        >
          Approval Queue
        </Button>
        <Button
          variant={activeTab === "preview" ? "primary" : "ghost"}
          onClick={() => setActiveTab("preview")}
        >
          Live Preview
        </Button>
      </div>

      {/* Templates Tab */}
      {activeTab === "templates" && (
        <GlassCard>
          <div className="flex items-center justify-between mb-4">
            <h2 className="text-lg font-semibold">Client Templates</h2>
            <Button onClick={() => setShowForm(true)}>
              <Plus className="w-4 h-4 mr-1" /> New Template
            </Button>
          </div>

          <div className="overflow-x-auto">
            <table className="w-full text-sm">
              <thead>
                <tr className="border-b border-white/10">
                  <th className="text-left p-2">Name</th>
                  <th className="text-left p-2">Pattern</th>
                  <th className="text-left p-2">Priority</th>
                  <th className="text-left p-2">Status</th>
                  <th className="text-right p-2">Actions</th>
                </tr>
              </thead>
              <tbody>
                {templates.map((tpl) => (
                  <tr key={tpl.id} className="border-b border-white/5 hover:bg-white/5">
                    <td className="p-2 font-medium">{tpl.name}</td>
                    <td className="p-2">
                      <code className="text-xs bg-white/10 px-1 rounded">{tpl.client_pattern}</code>
                    </td>
                    <td className="p-2">{tpl.priority}</td>
                    <td className="p-2">
                      <Badge color={tpl.enabled ? "active" : "disabled"}>
                        {tpl.enabled ? "Enabled" : "Disabled"}
                      </Badge>
                    </td>
                    <td className="p-2 text-right space-x-1">
                      <Button variant="ghost" size="sm" onClick={() => openEdit(tpl)}>
                        <Pencil className="w-3 h-3" />
                      </Button>
                      <Button variant="ghost" size="sm" onClick={() => handleDelete(tpl.id)}>
                        <Trash2 className="w-3 h-3 text-red-400" />
                      </Button>
                    </td>
                  </tr>
                ))}
                {templates.length === 0 && (
                  <tr>
                    <td colSpan={5} className="p-4 text-center text-white/50">
                      No client templates configured
                    </td>
                  </tr>
                )}
              </tbody>
            </table>
          </div>
        </GlassCard>
      )}

      {/* Approvals Tab */}
      {activeTab === "approvals" && (
        <GlassCard>
          <h2 className="text-lg font-semibold mb-4">Pending Approvals</h2>
          <div className="space-y-3">
            {approvals.map((a) => (
              <div
                key={a.id}
                className="flex items-center justify-between p-3 rounded-lg bg-white/5"
              >
                <div>
                  <div className="font-medium text-sm">User: {a.user_id}</div>
                  <div className="text-xs text-white/50">
                    Submitted: {new Date(a.created_at).toLocaleString()}
                  </div>
                  <pre className="text-xs mt-1 text-white/60 max-w-md truncate">
                    {JSON.stringify(a.request_data)}
                  </pre>
                </div>
                <div className="flex gap-2">
                  <Button
                    variant="ghost"
                    size="sm"
                    onClick={() => approveMutation.mutate(a.id)}
                    className="text-green-400"
                  >
                    <Check className="w-4 h-4" /> Approve
                  </Button>
                  <Button
                    variant="ghost"
                    size="sm"
                    onClick={() => rejectMutation.mutate(a.id)}
                    className="text-red-400"
                  >
                    <X className="w-4 h-4" /> Reject
                  </Button>
                </div>
              </div>
            ))}
            {approvals.length === 0 && (
              <p className="text-center text-white/50 py-4">No pending approvals</p>
            )}
          </div>
        </GlassCard>
      )}

      {/* Preview Tab */}
      {activeTab === "preview" && (
        <GlassCard>
          <h2 className="text-lg font-semibold mb-4">Live Preview</h2>
          <div className="space-y-4">
            <div>
              <label className="block text-sm mb-1">Template String</label>
              <textarea
                className="w-full h-32 bg-black/20 border border-white/10 rounded-lg p-3 text-sm font-mono resize-y"
                placeholder="Enter template with variables like {USERNAME}, {NODE_NAME}..."
                value={previewTemplate}
                onChange={(e) => setPreviewTemplate(e.target.value)}
              />
            </div>
            <Button onClick={() => previewMutation.mutate(previewTemplate)}>
              <Eye className="w-4 h-4 mr-1" /> Preview
            </Button>
            {previewResult && (
              <div className="mt-4 p-3 bg-black/20 rounded-lg border border-white/10">
                <label className="block text-xs text-white/50 mb-1">Result</label>
                <pre className="text-sm font-mono whitespace-pre-wrap">{previewResult}</pre>
              </div>
            )}
          </div>
        </GlassCard>
      )}

      {/* Create/Edit Modal */}
      <Modal
        open={showForm}
        onClose={closeForm}
        title={editingTemplate ? "Edit Template" : "New Client Template"}
      >
        <div className="space-y-4">
          <div>
            <label className="block text-sm mb-1">Name</label>
            <Input
              value={form.name}
              onChange={(e) => setForm({ ...form, name: e.target.value })}
              placeholder="e.g., Clash Pro Settings"
            />
          </div>

          <div>
            <label className="block text-sm mb-1">Client Pattern (regex)</label>
            <Input
              value={form.client_pattern}
              onChange={(e) => setForm({ ...form, client_pattern: e.target.value })}
              placeholder="e.g., clash|mihomo"
            />
            <p className="text-xs text-white/40 mt-1">
              Regex matched against User-Agent (case-insensitive)
            </p>
          </div>

          <div>
            <label className="block text-sm mb-1">Priority</label>
            <Input
              type="number"
              value={form.priority}
              onChange={(e) => setForm({ ...form, priority: Number(e.target.value) })}
            />
          </div>

          <div>
            <label className="block text-sm mb-1">Routing Rules (JSON)</label>
            <textarea
              className="w-full h-24 bg-black/20 border border-white/10 rounded-lg p-2 text-sm font-mono resize-y"
              value={form.routing_rules}
              onChange={(e) => setForm({ ...form, routing_rules: e.target.value })}
            />
          </div>

          <div>
            <label className="block text-sm mb-1">DNS Settings (JSON)</label>
            <textarea
              className="w-full h-24 bg-black/20 border border-white/10 rounded-lg p-2 text-sm font-mono resize-y"
              value={form.dns_settings}
              onChange={(e) => setForm({ ...form, dns_settings: e.target.value })}
            />
          </div>

          <div>
            <label className="block text-sm mb-1">Custom Outbounds (JSON)</label>
            <textarea
              className="w-full h-24 bg-black/20 border border-white/10 rounded-lg p-2 text-sm font-mono resize-y"
              value={form.custom_outbounds}
              onChange={(e) => setForm({ ...form, custom_outbounds: e.target.value })}
            />
          </div>

          <div className="flex items-center gap-2">
            <input
              type="checkbox"
              id="enabled"
              checked={form.enabled}
              onChange={(e) => setForm({ ...form, enabled: e.target.checked })}
              className="rounded"
            />
            <label htmlFor="enabled" className="text-sm">
              Enabled
            </label>
          </div>

          <div className="flex justify-end gap-2 pt-2">
            <Button variant="ghost" onClick={closeForm}>
              Cancel
            </Button>
            <Button onClick={handleSubmit}>
              {editingTemplate ? "Update" : "Create"}
            </Button>
          </div>
        </div>
      </Modal>
    </div>
  );
}
