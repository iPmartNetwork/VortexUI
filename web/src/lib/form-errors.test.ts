import { describe, expect, it } from "vitest";
import { getApiErrorMessage } from "./form-errors";

const t = (key: string) => `tr:${key}`;

describe("getApiErrorMessage", () => {
  it("returns a status-aware message for throttled requests", () => {
    expect(getApiErrorMessage({ status: 429 }, "Invalid credentials", t)).toBe("tr:errors.tooManyRequests");
  });

  it("prefers the server message when present", () => {
    expect(getApiErrorMessage({ message: "custom failure" }, "fallback", t)).toBe("custom failure");
  });

  it("falls back to the provided string when no details exist", () => {
    expect(getApiErrorMessage(undefined, "fallback", t)).toBe("fallback");
  });

  it("maps generic unauthorized to a translated session message", () => {
    expect(getApiErrorMessage({ status: 401, message: "unauthorized" }, "fallback", t)).toBe("tr:errors.sessionExpired");
  });
});
