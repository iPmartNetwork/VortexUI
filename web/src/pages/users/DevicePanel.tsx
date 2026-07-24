import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { Smartphone, Trash2 } from "lucide-react";
import { api } from "@/api/client";
import { Button } from "@/components/ui";
import { GlassCard } from "@/components/veltrix";
import { useConfirm } from "@/components/confirm";
import { useToast } from "@/components/toast";

interface Device {
  id: string;
  user_id: string;
  hwid: string;
  os: string;
  last_seen: string;
  created_at: string;
}

interface DevicePanelProps {
  userId: string;
}

/**
 * DevicePanel displays a user's registered devices (HWIDs) and allows
 * revoking individual device registrations.
 */
export function DevicePanel({ userId }: DevicePanelProps) {
  const queryClient = useQueryClient();
  const toast = useToast();
  const confirm = useConfirm();

  const { data, isLoading } = useQuery({
    queryKey: ["user-devices", userId],
    queryFn: () => api<{ devices: Device[] }>(`/api/v2/users/${userId}/devices`),
    enabled: !!userId,
  });

  const revokeMutation = useMutation({
    mutationFn: (hwid: string) =>
      api(`/api/v2/users/${userId}/devices/${encodeURIComponent(hwid)}`, {
        method: "DELETE",
      }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["user-devices", userId] });
      toast.success("Device revoked");
    },
    onError: () => {
      toast.error("Failed to revoke device");
    },
  });

  async function handleRevoke(hwid: string) {
    const ok = await confirm({
      title: "Revoke device?",
      message: `This will remove HWID ${hwid} from the user's allowed devices. The device will need to re-register.`,
      confirmLabel: "Revoke",
      destructive: true,
    });
    if (ok) {
      revokeMutation.mutate(hwid);
    }
  }

  const devices = data?.devices ?? [];

  if (isLoading) {
    return (
      <div className="py-8 text-center text-fg-muted">Loading devices...</div>
    );
  }

  if (devices.length === 0) {
    return (
      <GlassCard className="p-6">
        <div className="flex flex-col items-center gap-2 py-6 text-fg-muted">
          <Smartphone size={32} className="opacity-50" />
          <p className="text-sm">No devices registered</p>
        </div>
      </GlassCard>
    );
  }

  return (
    <GlassCard className="overflow-hidden">
      <div className="overflow-x-auto">
        <table className="w-full text-sm">
          <thead>
            <tr className="border-b border-border/50 text-left text-xs text-fg-muted">
              <th className="px-4 py-3 font-medium">HWID</th>
              <th className="px-4 py-3 font-medium">OS</th>
              <th className="px-4 py-3 font-medium">Last Seen</th>
              <th className="px-4 py-3 font-medium w-16" />
            </tr>
          </thead>
          <tbody>
            {devices.map((device) => (
              <tr
                key={device.id}
                className="border-b border-border/30 last:border-0 hover:bg-surface-2/40 transition-colors"
              >
                <td className="px-4 py-3 font-mono text-xs">{device.hwid}</td>
                <td className="px-4 py-3">{device.os || "unknown"}</td>
                <td className="px-4 py-3 text-fg-muted">
                  {formatLastSeen(device.last_seen)}
                </td>
                <td className="px-4 py-3">
                  <Button
                    variant="ghost"
                    size="sm"
                    onClick={() => handleRevoke(device.hwid)}
                    disabled={revokeMutation.isPending}
                    aria-label={`Revoke device ${device.hwid}`}
                  >
                    <Trash2 size={14} className="text-danger" />
                  </Button>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </GlassCard>
  );
}

function formatLastSeen(iso: string): string {
  if (!iso) return "Never";
  const date = new Date(iso);
  if (isNaN(date.getTime())) return "Unknown";
  const now = Date.now();
  const diff = now - date.getTime();
  if (diff < 60_000) return "Just now";
  if (diff < 3_600_000) return `${Math.floor(diff / 60_000)}m ago`;
  if (diff < 86_400_000) return `${Math.floor(diff / 3_600_000)}h ago`;
  return date.toLocaleDateString();
}
