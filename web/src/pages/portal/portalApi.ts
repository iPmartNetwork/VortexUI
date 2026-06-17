import { getPortalToken, clearPortalToken } from "./PortalLogin";

export class PortalApiError extends Error {
  status: number;
  constructor(status: number, message: string) {
    super(message);
    this.status = status;
  }
}

export async function portalApi<T>(path: string, opts: { method?: string; body?: unknown } = {}): Promise<T> {
  const headers: Record<string, string> = {};
  const token = getPortalToken();
  if (token) headers["Authorization"] = `Bearer ${token}`;
  if (opts.body !== undefined) headers["Content-Type"] = "application/json";

  const res = await fetch(path, {
    method: opts.method ?? "GET",
    headers,
    body: opts.body !== undefined ? JSON.stringify(opts.body) : undefined,
  });

  if (res.status === 401) {
    clearPortalToken();
    window.location.href = "/portal/login";
    throw new PortalApiError(401, "unauthorized");
  }
  if (!res.ok) {
    const detail = await res.json().catch(() => ({}));
    throw new PortalApiError(res.status, detail.message ?? `request failed (${res.status})`);
  }
  if (res.status === 204) return undefined as T;
  return (await res.json()) as T;
}
