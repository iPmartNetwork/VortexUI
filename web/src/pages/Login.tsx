import { useState } from "react";
import { useNavigate } from "react-router-dom";
import { useAuth } from "@/auth/auth";
import { Button, Card, Input } from "@/components/ui";

export function Login() {
  const { login } = useAuth();
  const navigate = useNavigate();
  const [username, setUsername] = useState("");
  const [password, setPassword] = useState("");
  const [totp, setTotp] = useState("");
  const [error, setError] = useState("");
  const [busy, setBusy] = useState(false);

  async function submit(e: React.FormEvent) {
    e.preventDefault();
    setError("");
    setBusy(true);
    try {
      await login(username, password, totp);
      navigate("/users");
    } catch {
      setError("Invalid credentials");
    } finally {
      setBusy(false);
    }
  }

  return (
    <div className="flex min-h-screen items-center justify-center p-4">
      <Card className="w-full max-w-sm">
        <h1 className="mb-1 text-xl font-bold tracking-tight">
          Vortex<span className="text-primary">UI</span>
        </h1>
        <p className="mb-6 text-sm text-muted-foreground">Sign in to the panel</p>
        <form onSubmit={submit} className="space-y-3">
          <Input placeholder="Username" value={username} onChange={(e) => setUsername(e.target.value)} autoFocus />
          <Input
            type="password"
            placeholder="Password"
            value={password}
            onChange={(e) => setPassword(e.target.value)}
          />
          <Input
            placeholder="2FA code (if enabled)"
            value={totp}
            onChange={(e) => setTotp(e.target.value)}
            inputMode="numeric"
          />
          {error && <p className="text-sm text-destructive">{error}</p>}
          <Button type="submit" className="w-full" disabled={busy}>
            {busy ? "Signing in…" : "Sign in"}
          </Button>
        </form>
      </Card>
    </div>
  );
}
