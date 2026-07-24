import { useState } from "react";
import { useMutation, useQueryClient } from "@tanstack/react-query";
import { api } from "@/api/client";
import { Button, Input } from "./ui";
import { Modal } from "./Modal";

interface BulkCreateResult {
  usernames: string[];
  created_count: number;
}

interface BulkCreateDialogProps {
  open: boolean;
  onClose: () => void;
  templateId: string;
  templateName: string;
}

export function BulkCreateDialog({ open, onClose, templateId, templateName }: BulkCreateDialogProps) {
  const qc = useQueryClient();
  const [count, setCount] = useState("10");
  const [result, setResult] = useState<BulkCreateResult | null>(null);
  const [error, setError] = useState("");

  const bulkCreate = useMutation({
    mutationFn: (input: { count: number }) =>
      api<BulkCreateResult>(`/api/v2/templates/${templateId}/bulk-create`, {
        method: "POST",
        body: input,
      }),
    onSuccess: (data) => {
      setResult(data);
      qc.invalidateQueries({ queryKey: ["users"] });
    },
  });

  function reset() {
    setCount("10");
    setResult(null);
    setError("");
  }

  function handleClose() {
    reset();
    onClose();
  }

  const n = Math.max(0, Math.min(1000, Number(count) || 0));

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    setError("");

    if (n < 1 || n > 1000) {
      setError("Count must be between 1 and 1000");
      return;
    }

    try {
      await bulkCreate.mutateAsync({ count: n });
    } catch (err: unknown) {
      setError(err instanceof Error ? err.message : "Bulk create failed");
    }
  }

  return (
    <Modal open={open} onClose={handleClose} title={result ? "Bulk Create Complete" : "Bulk Create Users"}>
      {result ? (
        <div className="space-y-4">
          <p className="text-sm text-fg">
            Created <span className="font-semibold text-success">{result.created_count}</span> users
            from template <span className="font-medium">"{templateName}"</span>.
          </p>

          {result.usernames.length > 0 && (
            <div>
              <p className="mb-1 text-xs font-medium text-fg-muted">Created usernames:</p>
              <div className="max-h-48 overflow-auto rounded-lg border border-border/60 bg-surface/40 p-3">
                <div className="space-y-1">
                  {result.usernames.map((username) => (
                    <div
                      key={username}
                      className="rounded-md bg-surface-2/40 px-2 py-1 text-xs font-mono text-fg"
                    >
                      {username}
                    </div>
                  ))}
                </div>
              </div>
            </div>
          )}

          <Button variant="ghost" className="w-full" onClick={handleClose}>
            Done
          </Button>
        </div>
      ) : (
        <form onSubmit={handleSubmit} className="space-y-4">
          <div>
            <label className="mb-1 block text-xs font-medium text-fg-muted">
              Number of users to create
            </label>
            <Input
              type="number"
              min={1}
              max={1000}
              value={count}
              onChange={(e) => setCount(e.target.value)}
              inputMode="numeric"
              autoFocus
            />
            <input
              type="range"
              min={1}
              max={1000}
              value={n}
              onChange={(e) => setCount(e.target.value)}
              className="mt-2 w-full accent-primary"
            />
          </div>

          {/* Preview */}
          <div className="rounded-lg bg-surface-2/40 border border-border/40 px-3 py-2">
            <p className="text-xs text-fg-muted">
              Will create <span className="font-semibold text-fg">{n}</span> user{n !== 1 ? "s" : ""} from
              template "<span className="font-medium text-fg">{templateName}</span>"
            </p>
          </div>

          {/* Error */}
          {error && <p className="text-sm text-danger">{error}</p>}

          {/* Actions */}
          <div className="flex justify-end gap-2 pt-1">
            <Button type="button" variant="ghost" onClick={handleClose}>
              Cancel
            </Button>
            <Button type="submit" disabled={bulkCreate.isPending}>
              {bulkCreate.isPending ? "Creating..." : `Create ${n} User${n !== 1 ? "s" : ""}`}
            </Button>
          </div>
        </form>
      )}
    </Modal>
  );
}
