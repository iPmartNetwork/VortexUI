import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { api } from "./client";

export interface PanelSettings {
  panel_name: string;
  panel_domain: string;
  sub_url_template: string;
  auto_sync_nodes: boolean;
  debug_mode: boolean;
  clash_rules_extra: string;
  singbox_dns_extra: string;
  accent_color: string;
  logo_url: string;
  footer_text: string;
  ip_whitelist: string;
  ip_blacklist: string;
  push_notifications: boolean;
  email_alerts: boolean;
  notify_telegram_token: string;
  require_2fa: boolean;
  api_access_enabled: boolean;
  auto_backup_enabled: boolean;
  auto_backup_interval_hours: number;
  auto_backup_telegram_chat_id: string;
  auto_backup_s3_endpoint: string;
  auto_backup_s3_bucket: string;
  auto_backup_s3_access_key: string;
  auto_backup_s3_secret_key: string;
}

export const DEFAULT_PANEL_SETTINGS: PanelSettings = {
  panel_name: "VortexUI",
  panel_domain: "",
  sub_url_template: "https://{domain}/sub/{token}",
  auto_sync_nodes: true,
  debug_mode: false,
  clash_rules_extra: "",
  singbox_dns_extra: "",
  accent_color: "#6366f1",
  logo_url: "",
  footer_text: "© 2026 iPmart Network. All rights reserved.",
  ip_whitelist: "",
  ip_blacklist: "",
  push_notifications: true,
  email_alerts: false,
  notify_telegram_token: "",
  require_2fa: false,
  api_access_enabled: true,
  auto_backup_enabled: false,
  auto_backup_interval_hours: 24,
  auto_backup_telegram_chat_id: "",
  auto_backup_s3_endpoint: "",
  auto_backup_s3_bucket: "",
  auto_backup_s3_access_key: "",
  auto_backup_s3_secret_key: "",
};

function clampBackupInterval(value: number | undefined) {
  if (typeof value !== "number" || !Number.isFinite(value)) {
    return DEFAULT_PANEL_SETTINGS.auto_backup_interval_hours;
  }
  return Math.min(168, Math.max(1, Math.round(value)));
}

function normalizeString(value: string | undefined, fallback: string) {
  const trimmed = value?.trim();
  return trimmed ? trimmed : fallback;
}

function isValidAccentColor(value: string | undefined) {
  return /^#([0-9a-f]{3}|[0-9a-f]{6})$/i.test(value?.trim() ?? "");
}

function isValidSubUrlTemplate(value: string | undefined) {
  const trimmed = value?.trim();
  if (!trimmed) return false;
  return trimmed.includes("{domain}") && trimmed.includes("{token}");
}

export function sanitizePanelSettings(input?: Partial<PanelSettings>): PanelSettings {
  const sanitized = {
    ...DEFAULT_PANEL_SETTINGS,
    ...(input ?? {}),
  } as PanelSettings;

  return {
    ...sanitized,
    panel_name: normalizeString(sanitized.panel_name, DEFAULT_PANEL_SETTINGS.panel_name),
    panel_domain: sanitized.panel_domain.trim(),
    sub_url_template: isValidSubUrlTemplate(sanitized.sub_url_template)
      ? sanitized.sub_url_template.trim()
      : DEFAULT_PANEL_SETTINGS.sub_url_template,
    accent_color: isValidAccentColor(sanitized.accent_color)
      ? sanitized.accent_color.trim().toLowerCase()
      : DEFAULT_PANEL_SETTINGS.accent_color,
    auto_backup_interval_hours: clampBackupInterval(sanitized.auto_backup_interval_hours),
    footer_text: normalizeString(sanitized.footer_text, DEFAULT_PANEL_SETTINGS.footer_text),
    logo_url: sanitized.logo_url.trim(),
    notify_telegram_token: sanitized.notify_telegram_token.trim(),
    auto_backup_telegram_chat_id: sanitized.auto_backup_telegram_chat_id.trim(),
    auto_backup_s3_endpoint: sanitized.auto_backup_s3_endpoint.trim(),
    auto_backup_s3_bucket: sanitized.auto_backup_s3_bucket.trim(),
    auto_backup_s3_access_key: sanitized.auto_backup_s3_access_key.trim(),
    auto_backup_s3_secret_key: sanitized.auto_backup_s3_secret_key.trim(),
    clash_rules_extra: sanitized.clash_rules_extra.trim(),
    singbox_dns_extra: sanitized.singbox_dns_extra.trim(),
    ip_whitelist: sanitized.ip_whitelist.trim(),
    ip_blacklist: sanitized.ip_blacklist.trim(),
  };
}

export function usePanelSettings(enabled = true) {
  return useQuery({
    queryKey: ["panel-settings"],
    enabled,
    queryFn: async () => {
      const res = await api<{ settings: PanelSettings }>("/api/settings");
      return sanitizePanelSettings(res.settings);
    },
  });
}

export function useSavePanelSettings() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (settings: PanelSettings) =>
      api<{ settings: PanelSettings }>('/api/settings', { method: 'PUT', body: sanitizePanelSettings(settings) }),
    onSuccess: (res) => {
      qc.setQueryData(['panel-settings'], sanitizePanelSettings(res.settings));
    },
  });
}

export function mergePanelSettings(
  current: PanelSettings | undefined,
  patch: Partial<PanelSettings>,
): PanelSettings {
  return sanitizePanelSettings({ ...(current ?? DEFAULT_PANEL_SETTINGS), ...patch });
}
