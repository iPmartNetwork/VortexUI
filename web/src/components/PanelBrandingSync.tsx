import { useEffect } from "react";
import { usePanelSettings } from "@/api/settings-hooks";
import { useAuth } from "@/auth/auth";
import { applyAccentColor } from "@/theme/branding";
import { useTheme } from "@/theme/theme";

/** Applies panel accent color from server settings after login. */
export function PanelBrandingSync() {
  const { isAuthenticated } = useAuth();
  const { data: settings } = usePanelSettings(isAuthenticated);
  const { resolved } = useTheme();

  useEffect(() => {
    if (isAuthenticated && settings?.accent_color) {
      applyAccentColor(settings.accent_color, resolved);
    }
  }, [isAuthenticated, settings?.accent_color, resolved]);

  return null;
}
