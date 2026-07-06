import { describe, expect, it } from "vitest";
import { mergePanelSettings, DEFAULT_PANEL_SETTINGS, sanitizePanelSettings } from "./settings-hooks";

describe("mergePanelSettings", () => {
  it("merges patch into defaults when current is undefined", () => {
    const result = mergePanelSettings(undefined, { panel_name: "TestPanel" });
    expect(result.panel_name).toBe("TestPanel");
    expect(result.accent_color).toBe(DEFAULT_PANEL_SETTINGS.accent_color);
  });

  it("preserves existing values not in patch", () => {
    const current = { ...DEFAULT_PANEL_SETTINGS, panel_name: "Keep", debug_mode: true };
    const result = mergePanelSettings(current, { panel_name: "New" });
    expect(result.panel_name).toBe("New");
    expect(result.debug_mode).toBe(true);
  });
});

describe("sanitizePanelSettings", () => {
  it("trims values and corrects invalid defaults", () => {
    const result = sanitizePanelSettings({
      panel_name: "  Test Panel  ",
      panel_domain: " example.com ",
      sub_url_template: " https://{domain}/sub/{token} ",
      accent_color: "#123456",
      auto_backup_interval_hours: 0,
    });

    expect(result.panel_name).toBe("Test Panel");
    expect(result.panel_domain).toBe("example.com");
    expect(result.sub_url_template).toBe("https://{domain}/sub/{token}");
    expect(result.accent_color).toBe("#123456");
    expect(result.auto_backup_interval_hours).toBe(1);
  });

  it("falls back to defaults for malformed URL templates", () => {
    const result = sanitizePanelSettings({ sub_url_template: "https://example.com/sub" });
    expect(result.sub_url_template).toBe(DEFAULT_PANEL_SETTINGS.sub_url_template);
  });
});
