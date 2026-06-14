import { useState } from "react";
import { QRCodeSVG } from "qrcode.react";
import { useConfirmTOTP, useDisableTOTP, useSetupTOTP } from "@/api/admin-hooks";
import { Button, Card, Input } from "@/components/ui";

export function Settings() {
  const setup = useSetupTOTP();
  const confirm = useConfirmTOTP();
  const disable = useDisableTOTP();

  const [url, setUrl] = useState("");
  const [code, setCode] = useState("");
  const [enabled, setEnabled] = useState(false);
  const [msg, setMsg] = useState("");

  async function begin() {
    setMsg("");
    const res = await setup.mutateAsync();
    setUrl(res.url);
  }

  async function activate(e: React.FormEvent) {
    e.preventDefault();
    setMsg("");
    try {
      await confirm.mutateAsync(code);
      setEnabled(true);
      setUrl("");
      setCode("");
    } catch {
      setMsg("Invalid code — try the current one from your app.");
    }
  }

  async function turnOff(e: React.FormEvent) {
    e.preventDefault();
    setMsg("");
    try {
      await disable.mutateAsync(code);
      setEnabled(false);
      setCode("");
    } catch {
      setMsg("Invalid code.");
    }
  }

  return (
    <div className="max-w-lg space-y-6">
      <h1 className="text-2xl font-bold tracking-tight">Settings</h1>

      <Card className="space-y-4">
        <div>
          <h2 className="font-semibold">Two-factor authentication</h2>
          <p className="text-sm text-muted-foreground">
            Protect your account with a time-based one-time code.
          </p>
        </div>

        {enabled ? (
          <form onSubmit={turnOff} className="space-y-3">
            <p className="text-sm text-green-400">2FA is enabled on your account.</p>
            <div className="flex gap-2">
              <Input placeholder="Current 6-digit code" value={code} onChange={(e) => setCode(e.target.value)} inputMode="numeric" />
              <Button type="submit" variant="destructive" disabled={disable.isPending}>Disable</Button>
            </div>
          </form>
        ) : url ? (
          <form onSubmit={activate} className="space-y-3">
            <p className="text-sm text-muted-foreground">Scan with your authenticator, then enter the code to confirm.</p>
            <div className="flex justify-center rounded-lg bg-white p-4">
              <QRCodeSVG value={url} size={160} />
            </div>
            <div className="flex gap-2">
              <Input placeholder="6-digit code" value={code} onChange={(e) => setCode(e.target.value)} inputMode="numeric" autoFocus />
              <Button type="submit" disabled={confirm.isPending}>Confirm</Button>
            </div>
          </form>
        ) : (
          <Button onClick={begin} disabled={setup.isPending}>
            {setup.isPending ? "Preparing…" : "Enable 2FA"}
          </Button>
        )}
        {msg && <p className="text-sm text-destructive">{msg}</p>}
      </Card>
    </div>
  );
}
