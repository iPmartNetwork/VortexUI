import { QRCodeSVG } from "qrcode.react";
import { RotateCcw, KeyRound, Wifi } from "lucide-react";
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
  if (!user) return null;
  const d = sub.data;
  const on = online.data;

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
