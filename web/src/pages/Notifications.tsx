import { useState } from "react";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { Plus, Trash2, Bell, Send, Pencil } from "lucide-react";
import { api } from "@/api/client";
import { Button, Input, Select, Badge } from "@/components/ui";
import { GlassCard } from "@/components/veltrix";
import { Modal } from "@/components/Modal";
import { useToast } from "@/components/toast";
import { useConfirm } from "@/components/confirm";
import { useTitle } from "@/lib/useTitle";
import { useI18n } from "@/i18n/i18n";

// Types
interface NotificationChannel {
  id: string;
  name: string;
  type: "telegram" | "webhook";
  config: Record<string, string>;
  scope_type: string;
  scope_id?: string;
  events: string[];
  template: string;
  enabled: boolean;
  created_at: string;
  updated_at: string;
}

interface ChannelForm {
  name: string;
  type: "telegram" | "webhook";
  config: Record<string, string>;
  scope_type: string;
  scope_id: string;
  events: string[];
  template: string;
  enabled: boolean;
}

const EVENT_TYPES = [
  "user_created",
  "user_deleted",
  "user_limited",
  "user_expired",
  "user_expiry_warning",
  "user_reset",
  "node_down",
  "node_up",
  "node_disconnect_alert",
  "security_probe",
  "protocol_switch",
  "admin_quota_warning",
] as const;

const SCOPE_TYPES = [
  { value: "global", label: "Global" },
  { value: "node", label: "Node" },
  { value: "admin", label: "Admin" },
  { value: "group", label: "Group" },
] as const;

const defaultForm: ChannelForm = {
  name: "",
  type: "telegram",
  config: {},
  scope_type: "global",
  scope_id: "",
  events: [],
  template: "",
  enabled: true,
};

export function Notifications() {
  const { t } = useI18n();
  useTitle(t("notifications.title", "Notifications"));
  const queryClient = useQueryClient();
  const toast = useToast();
  const confirm = useConfirm();

  const [showForm, setShowForm] = useState(false);
  const [editingChannel, setEditingChannel] = useState<NotificationChannel | null>(null);
  const [form, setForm] = useState<ChannelForm>(defaultForm);

  // Fetch channels
  const { data, isLoading } = useQuery({
    queryKey: ["notification-channels"],
    queryFn: () =>
      api<{ channels: NotificationChannel[] }>("/api/v2/notifications/channels"),
  });

  const channels = data?.channels ?? [];

  // Create channel
  const createMutation = useMutation({
    mutationFn: (body: ChannelForm) =>
      api("/api/v2/notifications/channels", { method: "POST", body }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["notification-channels"] });
      toast.success(t("notifications.created", "Channel created"));
      resetForm();
    },
    onError: () => toast.error(t("notifications.createFailed", "Failed to create channel")),
  });

  // Update channel
  const updateMutation = useMutation({
    mutationFn: ({ id, body }: { id: string; body: ChannelForm }) =>
      api(`/api/v2/notifications/channels/${id}`, { method: "PUT", body }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["notification-channels"] });
      toast.success(t("notifications.updated", "Channel updated"));
      resetForm();
    },
    onError: () => toast.error(t("notifications.updateFailed", "Failed to update channel")),
  });

  // Delete channel
  const deleteMutation = useMutation({
    mutationFn: (id: string) =>
      api(`/api/v2/notifications/channels/${id}`, { method: "DELETE" }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["notification-channels"] });
      toast.success(t("notifications.deleted", "Channel deleted"));
    },
    onError: () => toast.error(t("notifications.deleteFailed", "Failed to delete channel")),
  });

  // Test notification
  const testMutation = useMutation({
    mutationFn: (channelId: string) =>
      api("/api/v2/notifications/test", {
        method: "POST",
        body: { channel_id: channelId, message: "Test notification from VortexUI" },
      }),
    onSuccess: () => toast.success(t("notifications.testSent", "Test notification sent")),
    onError: () => toast.error(t("notifications.testFailed", "Test notification failed")),
  });

  function resetForm() {
    setForm(defaultForm);
    setEditingChannel(null);
    setShowForm(false);
  }

  function openCreate() {
    setForm(defaultForm);
    setEditingChannel(null);
    setShowForm(true);
  }

  function openEdit(ch: NotificationChannel) {
    setForm({
      name: ch.name,
      type: ch.type,
      config: ch.config,
      scope_type: ch.scope_type,
      scope_id: ch.scope_id ?? "",
      events: ch.events,
      template: ch.template,
      enabled: ch.enabled,
    });
    setEditingChannel(ch);
    setShowForm(true);
  }

  async function handleDelete(ch: NotificationChannel) {
    const ok = await confirm({
      title: t("notifications.deleteConfirm", "Delete Channel"),
      message: t("notifications.deleteMessage", `Delete "${ch.name}"? This cannot be undone.`),
    });
    if (ok) deleteMutation.mutate(ch.id);
  }

  function handleSubmit() {
    if (editingChannel) {
      updateMutation.mutate({ id: editingChannel.id, body: form });
    } else {
      createMutation.mutate(form);
    }
  }

  function toggleEvent(event: string) {
    setForm((prev) => ({
      ...prev,
      events: prev.events.includes(event)
        ? prev.events.filter((e) => e !== event)
        : [...prev.events, event],
    }));
  }

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-white">
            {t("notifications.title", "Notifications")}
          </h1>
          <p className="text-sm text-gray-400 mt-1">
            {t("notifications.subtitle", "Manage notification channels and event routing")}
          </p>
        </div>
        <Button onClick={openCreate} className="gap-2">
          <Plus className="w-4 h-4" />
          {t("notifications.addChannel", "Add Channel")}
        </Button>
      </div>

      {/* Channel list */}
      <GlassCard>
        {isLoading ? (
          <div className="p-8 text-center text-gray-400">Loading...</div>
        ) : channels.length === 0 ? (
          <div className="p-8 text-center text-gray-400">
            <Bell className="w-12 h-12 mx-auto mb-3 opacity-50" />
            <p>{t("notifications.empty", "No notification channels configured")}</p>
          </div>
        ) : (
          <div className="overflow-x-auto">
            <table className="w-full text-sm">
              <thead>
                <tr className="border-b border-white/10 text-left text-gray-400">
                  <th className="py-3 px-4">{t("notifications.name", "Name")}</th>
                  <th className="py-3 px-4">{t("notifications.type", "Type")}</th>
                  <th className="py-3 px-4">{t("notifications.scope", "Scope")}</th>
                  <th className="py-3 px-4">{t("notifications.events", "Events")}</th>
                  <th className="py-3 px-4">{t("notifications.status", "Status")}</th>
                  <th className="py-3 px-4">{t("notifications.actions", "Actions")}</th>
                </tr>
              </thead>
              <tbody>
                {channels.map((ch) => (
                  <tr key={ch.id} className="border-b border-white/5 hover:bg-white/5">
                    <td className="py-3 px-4 font-medium text-white">{ch.name}</td>
                    <td className="py-3 px-4">
                      <Badge variant={ch.type === "telegram" ? "info" : "default"}>
                        {ch.type}
                      </Badge>
                    </td>
                    <td className="py-3 px-4 text-gray-300">
                      {ch.scope_type}
                      {ch.scope_id && <span className="text-gray-500 ml-1">({ch.scope_id})</span>}
                    </td>
                    <td className="py-3 px-4 text-gray-300">
                      {ch.events.length} event{ch.events.length !== 1 ? "s" : ""}
                    </td>
                    <td className="py-3 px-4">
                      <Badge variant={ch.enabled ? "success" : "warning"}>
                        {ch.enabled ? "Enabled" : "Disabled"}
                      </Badge>
                    </td>
                    <td className="py-3 px-4">
                      <div className="flex items-center gap-2">
                        <Button
                          size="sm"
                          variant="ghost"
                          onClick={() => testMutation.mutate(ch.id)}
                          title="Test"
                        >
                          <Send className="w-4 h-4" />
                        </Button>
                        <Button
                          size="sm"
                          variant="ghost"
                          onClick={() => openEdit(ch)}
                          title="Edit"
                        >
                          <Pencil className="w-4 h-4" />
                        </Button>
                        <Button
                          size="sm"
                          variant="ghost"
                          onClick={() => handleDelete(ch)}
                          title="Delete"
                          className="text-red-400 hover:text-red-300"
                        >
                          <Trash2 className="w-4 h-4" />
                        </Button>
                      </div>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </GlassCard>

      {/* Create/Edit Modal */}
      {showForm && (
        <Modal
          title={editingChannel ? t("notifications.editChannel", "Edit Channel") : t("notifications.addChannel", "Add Channel")}
          onClose={resetForm}
        >
          <div className="space-y-4">
            {/* Name */}
            <div>
              <label className="block text-sm font-medium text-gray-300 mb-1">
                {t("notifications.name", "Name")}
              </label>
              <Input
                value={form.name}
                onChange={(e) => setForm({ ...form, name: e.target.value })}
                placeholder="My Notification Channel"
              />
            </div>

            {/* Type */}
            <div>
              <label className="block text-sm font-medium text-gray-300 mb-1">
                {t("notifications.type", "Type")}
              </label>
              <Select
                value={form.type}
                onChange={(e) =>
                  setForm({ ...form, type: e.target.value as "telegram" | "webhook", config: {} })
                }
              >
                <option value="telegram">Telegram</option>
                <option value="webhook">Webhook</option>
              </Select>
            </div>

            {/* Config - Telegram */}
            {form.type === "telegram" && (
              <div>
                <label className="block text-sm font-medium text-gray-300 mb-1">
                  Chat ID
                </label>
                <Input
                  value={form.config.chat_id ?? ""}
                  onChange={(e) =>
                    setForm({ ...form, config: { ...form.config, chat_id: e.target.value } })
                  }
                  placeholder="-1001234567890"
                />
              </div>
            )}

            {/* Config - Webhook */}
            {form.type === "webhook" && (
              <>
                <div>
                  <label className="block text-sm font-medium text-gray-300 mb-1">
                    Webhook URL
                  </label>
                  <Input
                    value={form.config.url ?? ""}
                    onChange={(e) =>
                      setForm({ ...form, config: { ...form.config, url: e.target.value } })
                    }
                    placeholder="https://example.com/webhook"
                  />
                </div>
                <div>
                  <label className="block text-sm font-medium text-gray-300 mb-1">
                    HMAC Secret
                  </label>
                  <Input
                    value={form.config.hmac_secret ?? ""}
                    onChange={(e) =>
                      setForm({ ...form, config: { ...form.config, hmac_secret: e.target.value } })
                    }
                    placeholder="Optional HMAC secret for signing"
                  />
                </div>
              </>
            )}

            {/* Scope */}
            <div className="grid grid-cols-2 gap-3">
              <div>
                <label className="block text-sm font-medium text-gray-300 mb-1">
                  {t("notifications.scope", "Scope")}
                </label>
                <Select
                  value={form.scope_type}
                  onChange={(e) => setForm({ ...form, scope_type: e.target.value })}
                >
                  {SCOPE_TYPES.map((s) => (
                    <option key={s.value} value={s.value}>
                      {s.label}
                    </option>
                  ))}
                </Select>
              </div>
              {form.scope_type !== "global" && (
                <div>
                  <label className="block text-sm font-medium text-gray-300 mb-1">
                    Scope ID
                  </label>
                  <Input
                    value={form.scope_id}
                    onChange={(e) => setForm({ ...form, scope_id: e.target.value })}
                    placeholder="node-id or admin-id"
                  />
                </div>
              )}
            </div>

            {/* Event types */}
            <div>
              <label className="block text-sm font-medium text-gray-300 mb-2">
                {t("notifications.events", "Events")}
              </label>
              <div className="grid grid-cols-2 gap-2 max-h-48 overflow-y-auto">
                {EVENT_TYPES.map((event) => (
                  <label
                    key={event}
                    className="flex items-center gap-2 text-sm text-gray-300 cursor-pointer"
                  >
                    <input
                      type="checkbox"
                      checked={form.events.includes(event)}
                      onChange={() => toggleEvent(event)}
                      className="rounded border-gray-600 bg-gray-800 text-blue-500 focus:ring-blue-500"
                    />
                    {event.replace(/_/g, " ")}
                  </label>
                ))}
              </div>
            </div>

            {/* Template */}
            <div>
              <label className="block text-sm font-medium text-gray-300 mb-1">
                {t("notifications.template", "Template")}
              </label>
              <textarea
                value={form.template}
                onChange={(e) => setForm({ ...form, template: e.target.value })}
                className="w-full rounded-lg border border-white/10 bg-gray-800 px-3 py-2 text-sm text-white placeholder-gray-500 focus:border-blue-500 focus:outline-none"
                rows={3}
                placeholder="{EVENT}: {USERNAME} - {MESSAGE}"
              />
              <p className="text-xs text-gray-500 mt-1">
                Available variables: {"{EVENT}"}, {"{USERNAME}"}, {"{NODE_NAME}"}, {"{MESSAGE}"}
              </p>
            </div>

            {/* Enabled */}
            <label className="flex items-center gap-2 text-sm text-gray-300 cursor-pointer">
              <input
                type="checkbox"
                checked={form.enabled}
                onChange={(e) => setForm({ ...form, enabled: e.target.checked })}
                className="rounded border-gray-600 bg-gray-800 text-blue-500 focus:ring-blue-500"
              />
              {t("notifications.enabled", "Enabled")}
            </label>

            {/* Actions */}
            <div className="flex justify-end gap-3 pt-2">
              <Button variant="ghost" onClick={resetForm}>
                {t("common.cancel", "Cancel")}
              </Button>
              <Button
                onClick={handleSubmit}
                disabled={!form.name || createMutation.isPending || updateMutation.isPending}
              >
                {editingChannel
                  ? t("common.save", "Save")
                  : t("common.create", "Create")}
              </Button>
            </div>
          </div>
        </Modal>
      )}
    </div>
  );
}
