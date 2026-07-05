import { describe, expect, it } from "vitest";
import { mergePanelSettings, DEFAULT_PANEL_SETTINGS } from "./settings-hooks";

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
