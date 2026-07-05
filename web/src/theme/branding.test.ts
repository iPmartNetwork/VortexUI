import { describe, expect, it } from "vitest";
import { hexToHslComponents } from "./branding";

describe("hexToHslComponents", () => {
  it("parses 6-digit hex with hash", () => {
    expect(hexToHslComponents("#6366f1")).toBe("239 84% 67%");
  });

  it("parses hex without hash", () => {
    expect(hexToHslComponents("6366f1")).toBe("239 84% 67%");
  });

  it("returns null for invalid input", () => {
    expect(hexToHslComponents("not-a-color")).toBeNull();
    expect(hexToHslComponents("#fff")).toBeNull();
  });
});
