import { createContext, useCallback, useContext, useEffect, useMemo, useState } from "react";
import { api, clearToken, getToken, setToken } from "@/api/client";
import type { Admin } from "@/api/types";
import { hasPermission } from "./permissions";

interface Session {
  admin: Admin;
  permissions: Set<string>;
}

interface AuthState {
  isAuthenticated: boolean;
  session: Session | null;
  loading: boolean;
  sudo: boolean;
  permissions: Set<string>;
  can: (perm: string) => boolean;
  login: (username: string, password: string, totp?: string) => Promise<void>;
  logout: () => void;
  refreshSession: () => Promise<void>;
}

const AuthContext = createContext<AuthState | null>(null);

export function AuthProvider({ children }: { children: React.ReactNode }) {
  const [token, setTok] = useState<string | null>(getToken());
  const [session, setSession] = useState<Session | null>(null);
  const [loading, setLoading] = useState(!!getToken());

  const refreshSession = useCallback(async () => {
    if (!getToken()) {
      setSession(null);
      setLoading(false);
      return;
    }
    setLoading(true);
    try {
      const res = await api<{ admin: Admin; permissions: string[] }>("/api/account");
      setSession({ admin: res.admin, permissions: new Set(res.permissions) });
    } catch {
      clearToken();
      setTok(null);
      setSession(null);
    } finally {
      setLoading(false);
    }
  }, []);

  const login = useCallback(async (username: string, password: string, totp?: string) => {
    const res = await api<{ token: string }>("/api/login", {
      method: "POST",
      body: { username, password, totp_code: totp ?? "" },
    });
    setToken(res.token);
    setTok(res.token);
    const account = await api<{ admin: Admin; permissions: string[] }>("/api/account");
    setSession({ admin: account.admin, permissions: new Set(account.permissions) });
    setLoading(false);
  }, []);

  const logout = useCallback(() => {
    clearToken();
    setTok(null);
    setSession(null);
    setLoading(false);
  }, []);

  useEffect(() => {
    if (!token) {
      setSession(null);
      setLoading(false);
      return;
    }
    void refreshSession();
  }, [token, refreshSession]);

  useEffect(() => {
    const onUnauthorized = () => {
      clearToken();
      setTok(null);
      setSession(null);
      setLoading(false);
    };
    window.addEventListener("vortex:unauthorized", onUnauthorized);
    return () => window.removeEventListener("vortex:unauthorized", onUnauthorized);
  }, []);

  const sudo = session?.admin.sudo ?? false;
  const permissions = session?.permissions ?? new Set<string>();

  const value = useMemo<AuthState>(
    () => ({
      isAuthenticated: !!token,
      session,
      loading,
      sudo,
      permissions,
      can: (perm: string) => hasPermission(sudo, permissions, perm),
      login,
      logout,
      refreshSession,
    }),
    [token, session, loading, sudo, permissions, login, logout, refreshSession],
  );
  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
}

export function useAuth(): AuthState {
  const ctx = useContext(AuthContext);
  if (!ctx) throw new Error("useAuth must be used within AuthProvider");
  return ctx;
}
