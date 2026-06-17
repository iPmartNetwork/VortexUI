import { useState } from "react";
import { useNavigate } from "react-router-dom";
import { Button, Input } from "@/components/ui";
import { useToast } from "@/components/toast";

const PORTAL_TOKEN_KEY = "vortex.portal.token";

export function setPortalToken(t: string) { localStorage.setItem(PORTAL_TOKEN_KEY, t); }
export function getPortalToken() { return localStorage.getItem(PORTAL_TOKEN_KEY); }
export function clearPortalToken() { localStorage.removeItem(PORTAL_TOKEN_KEY); }

export function PortalLogin() {
  const [token, setToken] = useState("");
  const [loading, setLoading] = useState(false);
  const navigate = useNavigate();
  const toast = useToast();

  async function submit(e: React.FormEvent) {
    e.preventDefault();
    setLoading(true);
    try {
      const res = await fetch("/api/portal/login", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ token }),
      });
      if (!res.ok) {
        const err = await res.json().catch(() => ({}));
        throw new Error(err.message || "Login failed");
      }
      const data = await res.json();
      setPortalToken(data.token);
      navigate("/portal/dashboard");
    } catch (err: any) {
      toast.error(err.message || "Invalid token");
    } finally {
      setLoading(false);
    }
  }

  return (
    <div className="flex min-h-screen items-center justify-center p-4 bg-bg">
      <div className="card w-full max-w-sm p-8 space-y-6 animate-scale-in">
        <div className="text-center space-y-2">
          <h1 className="text-2xl font-bold text-fg">User Portal</h1>
          <p className="text-sm text-fg-muted">Enter your subscription token to login</p>
        </div>
        <form onSubmit={submit} className="space-y-4">
          <Input
            placeholder="Subscription token"
            value={token}
            onChange={(e) => setToken(e.target.value)}
            required
            className="text-center"
          />
          <Button type="submit" className="w-full" disabled={loading || !token}>
            {loading ? "Logging in..." : "Login"}
          </Button>
        </form>
      </div>
    </div>
  );
}
