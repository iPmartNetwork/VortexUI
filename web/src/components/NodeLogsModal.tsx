import { useNodeLogs } from "@/api/policy-hooks";
import type { Node } from "@/api/types";
import { Modal } from "./Modal";
import { useI18n } from "@/i18n/i18n";

export function NodeLogsModal({ node, onClose }: { node: Node | null; onClose: () => void }) {
  const { t } = useI18n();
  const { data, isLoading } = useNodeLogs(node?.id ?? null);
  if (!node) return null;
  return (
    <Modal open={!!node} onClose={onClose} title={`${node.name} — ${t("nav.logs")}`} className="max-w-2xl">
      <div className="max-h-[60vh] overflow-auto rounded-lg bg-black/30 p-3 font-mono text-xs leading-5 text-fg-muted">
        {isLoading && <p className="text-fg-subtle">{t("common.loading")}</p>}
        {data?.lines?.length === 0 && <p className="text-fg-subtle">{t("common.none")}</p>}
        {data?.lines?.map((line, i) => (
          <div key={i} className="hover:text-fg">{line}</div>
        ))}
      </div>
    </Modal>
  );
}
