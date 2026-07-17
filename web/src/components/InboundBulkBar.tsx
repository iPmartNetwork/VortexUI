import { useState } from "react";
import { Button } from "./ui";
import { Modal } from "./Modal";

interface Props {
  selectedCount: number;
  onEnable: () => void;
  onDisable: () => void;
  onDelete: () => void;
  onClearSelection: () => void;
  isPending: boolean;
}

export function InboundBulkBar({ selectedCount, onEnable, onDisable, onDelete, onClearSelection, isPending }: Props) {
  const [confirmDelete, setConfirmDelete] = useState(false);

  if (selectedCount === 0) return null;

  return (
    <>
      <div className="fixed bottom-6 left-1/2 -translate-x-1/2 z-50 flex items-center gap-3 rounded-2xl bg-bg-elevated border border-border/60 shadow-2xl px-5 py-3 animate-slide-up">
        <span className="text-sm font-semibold text-fg">{selectedCount} selected</span>
        <div className="h-4 w-px bg-border/50" />
        <Button variant="ghost" size="sm" onClick={onEnable} disabled={isPending}>Enable</Button>
        <Button variant="ghost" size="sm" onClick={onDisable} disabled={isPending}>Disable</Button>
        <Button variant="ghost" size="sm" onClick={() => setConfirmDelete(true)} disabled={isPending} className="text-danger hover:text-danger">Delete</Button>
        <div className="h-4 w-px bg-border/50" />
        <button type="button" onClick={onClearSelection} className="text-xs text-fg-subtle hover:text-fg transition-colors">
          Clear
        </button>
      </div>

      <Modal open={confirmDelete} onClose={() => setConfirmDelete(false)} title="Confirm Delete">
        <p className="text-sm text-fg-muted mb-4">
          Are you sure you want to delete {selectedCount} inbound{selectedCount !== 1 ? "s" : ""}? This action cannot be undone.
        </p>
        <div className="flex justify-end gap-2">
          <Button variant="ghost" onClick={() => setConfirmDelete(false)}>Cancel</Button>
          <Button variant="destructive" onClick={() => { onDelete(); setConfirmDelete(false); }}>
            Delete {selectedCount}
          </Button>
        </div>
      </Modal>
    </>
  );
}
