import { QRCodeSVG } from "qrcode.react";
import { useEffect, useState } from "react";
import { RotateCcw, KeyRound, Wifi, Download, Copy } from "lucide-react";
import { useResetUser, useRevokeSub, useUserSub, useUserOnline } from "@/api/policy-hooks";
import type { User } from "@/api/types";
import { Button } from "./ui";
import { Modal } from "./Modal";
import { CopyField } from "./CopyField";
import { useConfirm } from "./confirm";
import { useToast } from "./toast";

export function UserSubModal({ user, onClose }: { user: User | null; onClose: () => void }) {
  const sub = useUserSub(user?.id ?? null);
  const online = useUserOnline(user?.id ?? null);
  const reset = useResetUser();
  const revoke = useRevokeSub();
  const confirm = useConfirm();
  const toast = useToast();

  // WireGuard config text (null = none / not loaded). Fetched once per link from
  // the public token-authed endpoint; a 404 means the user has no WG config and
  // the section stays hidden.
  const subUrl = sub.data?.subscription_url ?? null;
  const [wgConf, setWgConf] = useState<string | null>(null);
  useEffect(() => {
    setWgConf(null);
    if (!subUrl) return;
    let cancelled = false;
    fetch(`${subUrl}/wireguard`)
      .then((r) => (r.ok ? r.text() : null))
      .then((text) => {
        if (!cancelled && text) setWgConf(text);
      })
      .catch(() => {});
    return () => {
      cancelled = true;
    };
  }, [subUrl]);

  if (!user) return null;
  const d = sub.data;
  const on = online.data;

  function downloadWG() {
    if (!wgConf) return;
    const blob = new Blob([wgConf], { type: "text/plain" });
    const url = URL.createObjectURL(blob);
    const a = document.createElement("a");
    a.href = url;
    a.download = `${user?.username ?? "wireguard"}.conf`;
    document.body.appendChild(a);
    a.click();
    a.remove();
    URL.revokeObjectURL(url);
  }
  async function copyWG() {
    if (!wgConf) return;
    await navigator.clipboard.writeText(wgConf);
    toast.success("WireGuard config copied");
  }

  async function doReset() {
    if (!user) return;
    if (await confirm({ title: `Reset usage for ${user.username}?`, confirmLabel: "Reset" })) {
      await reset.mutateAsync(user.id);
      toast.success("Usage reset");
    }
  }
  async function doRevoke() {
    if (!user) return;
    if (await confirm({ title: `Revoke & regenerate ${user.username}'s link?`, message: "Old links stop working immediately.", confirmLabel: "Revoke", destructive: true })) {
      await revoke.mutateAsync(user.id);
      toast.success("Subscription revoked");
      onClose();
    }
  }

  return (
    <Modal open={!!user} onClose={onClose} title={user.username} className="max-w-lg">
      {!d ? (
        <p className="py-8 text-center text-sm text-fg-muted">Loading…</p>
      ) : (
        <div className="space-y-4">
          <div className="flex justify-center rounded-xl bg-white p-4">
            <QRCodeSVG value={d.subscription_url} size={150} />
          </div>

          {on && (on.live_tracking || on.device_tracking) && (
            <div className="flex items-center gap-4 rounded-lg border border-white/[0.06] bg-white/[0.02] px-4 py-2.5 text-sm">
              {on.live_tracking && (
                <span className="flex items-center gap-1.5">
                  <Wifi size={14} className={on.live_connections > 0 ? "text-success" : "text-fg-subtle"} />
                  <span className="font-medium">{on.live_connections}</span>
                  <span className="text-fg-muted">live</span>
                </span>
              )}
              {on.device_tracking && (
                <span className="flex items-center gap-1.5">
                  <span className="font-medium">{on.active_devices}</span>
                  <span className="text-fg-muted">devices</span>
                </span>
              )}
            </div>
          )}

          <div>
            <p className="mb-1.5 text-xs font-medium text-fg-muted">Subscription link</p>
            <CopyField value={d.subscription_url} />
          </div>

          <div>
            <p className="mb-1.5 text-xs font-medium text-fg-muted">Formats</p>
            <div className="grid grid-cols-2 gap-2">
              {(["clash", "singbox", "base64"] as const).map((k) => (
                <CopyField key={k} value={d.formats[k]} />
              ))}
            </div>
          </div>

          {d.links.length > 0 && (
            <div>
              <p className="mb-1.5 text-xs font-medium text-fg-muted">Configs ({d.links.length})</p>
              <div className="max-h-40 space-y-2 overflow-auto">
                {d.links.map((l, i) => (
                  <CopyField key={i} value={l} />
                ))}
              </div>
            </div>
          )}

          {wgConf && (
            <div>
              <p className="mb-1.5 flex items-center gap-1.5 text-xs font-medium text-fg-muted">
                <Wifi size={14} /> WireGuard
              </p>
              <div className="flex justify-center rounded-xl bg-white p-4">
                <QRCodeSVG value={wgConf} size={150} />
              </div>
              <div className="mt-2 flex gap-2">
                <Button variant="outline" className="flex-1" onClick={downloadWG}>
                  <Download size={15} /> Download
                </Button>
                <Button variant="outline" className="flex-1" onClick={copyWG}>
                  <Copy size={15} /> Copy config
                </Button>
              </div>
            </div>
          )}

          <div className="flex gap-2 border-t border-white/[0.06] pt-4">
            <Button variant="outline" className="flex-1" onClick={doReset}>
              <RotateCcw size={15} /> Reset usage
            </Button>
            <Button variant="outline" className="flex-1 text-danger" onClick={doRevoke}>
              <KeyRound size={15} /> Revoke link
            </Button>
          </div>
        </div>
      )}
    </Modal>
  );
}
