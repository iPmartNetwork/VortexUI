import { createContext, useCallback, useContext, useMemo, useState } from "react";
import { api, clearToken, getToken, setToken } from "@/api/client";

interface AuthState {
  isAuthenticated: boolean;
  login: (username: string, password: string, totp?: string) => Promise<void>;
  logout: () => void;
}

const AuthContext = createContext<AuthState | null>(null);

export function AuthProvider({ children }: { children: React.ReactNode }) {
  const [token, setTok] = useState<string | null>(getToken());

  const login = useCallback(async (username: string, password: string, totp?: string) => {
    const res = await api<{ token: string }>("/api/login", {
      method: "POST",
      body: { username, password, totp_code: totp ?? "" },
    });
    setToken(res.token);
    setTok(res.token);
  }, []);

  const logout = useCallback(() => {
    clearToken();
    setTok(null);
  }, []);

  const value = useMemo<AuthState>(
    () => ({ isAuthenticated: !!token, login, logout }),
    [token, login, logout],
  );
  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
}

export function useAuth(): AuthState {
  const ctx = useContext(AuthContext);
  if (!ctx) throw new Error("useAuth must be used within AuthProvider");
  return ctx;
}
