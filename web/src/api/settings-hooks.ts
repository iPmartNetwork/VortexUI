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

export function usePanelSettings(enabled = true) {
  return useQuery({
    queryKey: ["panel-settings"],
    enabled,
    queryFn: async () => {
      const res = await api<{ settings: PanelSettings }>("/api/settings");
      return { ...DEFAULT_PANEL_SETTINGS, ...res.settings };
    },
  });
}

export function useSavePanelSettings() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (settings: PanelSettings) =>
      api<{ settings: PanelSettings }>("/api/settings", { method: "PUT", body: settings }),
    onSuccess: (res) => {
      qc.setQueryData(["panel-settings"], { ...DEFAULT_PANEL_SETTINGS, ...res.settings });
    },
  });
}

export function mergePanelSettings(
  current: PanelSettings | undefined,
  patch: Partial<PanelSettings>,
): PanelSettings {
  return { ...(current ?? DEFAULT_PANEL_SETTINGS), ...patch };
}
