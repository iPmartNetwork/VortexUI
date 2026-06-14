import { useUserUsage } from "@/api/hooks";
import type { User } from "@/api/types";
import { formatBytes } from "@/lib/utils";
import { Modal } from "./Modal";
import { UsageChart } from "./UsageChart";

export function UserUsageModal({ user, onClose }: { user: User | null; onClose: () => void }) {
  const usage = useUserUsage(user?.id ?? null);
  if (!user) return null;

  return (
    <Modal open={!!user} onClose={onClose} title={`Usage · ${user.username}`} className="max-w-lg">
      <div className="mb-4 flex gap-6 text-sm">
        <div>
          <div className="text-xs text-muted-foreground">Total used</div>
          <div className="text-lg font-semibold">{formatBytes(user.used_traffic)}</div>
        </div>
        <div>
          <div className="text-xs text-muted-foreground">Limit</div>
          <div className="text-lg font-semibold">{formatBytes(user.data_limit)}</div>
        </div>
      </div>
      {usage.isLoading && <p className="py-8 text-center text-sm text-muted-foreground">Loading…</p>}
      {usage.data && <UsageChart points={usage.data.points} />}
    </Modal>
  );
}
