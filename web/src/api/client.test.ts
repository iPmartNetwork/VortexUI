import { beforeEach, describe, expect, it, vi } from "vitest";
import { api, clearToken, getToken, setToken } from "./client";

const fetchMock = vi.fn();
const localStorageMock = (() => {
  const store = new Map<string, string>();
  return {
    getItem: (key: string) => (store.has(key) ? store.get(key)! : null),
    setItem: (key: string, value: string) => store.set(key, value),
    removeItem: (key: string) => store.delete(key),
    clear: () => store.clear(),
  };
})();

vi.stubGlobal("fetch", fetchMock);
vi.stubGlobal("localStorage", localStorageMock);
vi.stubGlobal("window", {
  location: { origin: "http://localhost" },
  dispatchEvent: vi.fn(),
});

describe("api", () => {
  beforeEach(() => {
    localStorageMock.clear();
    fetchMock.mockReset();
  });

  it("uses plain-text error messages when the server does not return JSON", async () => {
    fetchMock.mockResolvedValue({
      ok: false,
      status: 502,
      text: vi.fn().mockResolvedValue("bad gateway"),
    });

    await expect(api("/test")).rejects.toMatchObject({
      status: 502,
      message: "bad gateway",
    });
  });

  it("attaches the auth token and parses JSON responses", async () => {
    setToken("abc123");
    fetchMock.mockResolvedValue({
      ok: true,
      status: 200,
      headers: { get: () => "application/json" },
      text: vi.fn().mockResolvedValue(JSON.stringify({ ok: true })),
    });

    const result = await api("/test");

    expect(fetchMock).toHaveBeenCalledWith(
      "http://localhost/test",
      expect.objectContaining({
        headers: expect.objectContaining({ Authorization: "Bearer abc123" }),
      }),
    );
    expect(result).toEqual({ ok: true });
  });

  it("falls back to plain text when a JSON response body is malformed", async () => {
    fetchMock.mockResolvedValue({
      ok: true,
      status: 200,
      headers: { get: () => "application/json" },
      text: vi.fn().mockResolvedValue("plain fallback"),
    });

    await expect(api("/test")).resolves.toBe("plain fallback");
  });

  it("clears the token and emits an unauthorized event on 401", async () => {
    setToken("old");
    const dispatch = vi.fn();
    vi.stubGlobal("window", { location: { origin: "http://localhost" }, dispatchEvent: dispatch });
    fetchMock.mockResolvedValue({
      ok: false,
      status: 401,
      text: vi.fn().mockResolvedValue(""),
    });

    await expect(api("/test")).rejects.toMatchObject({ status: 401, message: "unauthorized" });
    expect(getToken()).toBeNull();
    expect(dispatch).toHaveBeenCalledWith(expect.any(Event));
    clearToken();
  });
});
