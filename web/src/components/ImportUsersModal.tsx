import { useState } from "react";
import { useImportUsers } from "@/api/hooks";
import { Button, Select } from "./ui";
import { Modal } from "./Modal";
import { UploadCloud } from "lucide-react";

type Result = { parsed: number; created_count: number; failures: { username: string; error: string }[] };

// ImportUsersModal migrates users from a 3x-ui or Marzban export file. The raw
// JSON is sent to the server, which parses the foreign format and creates users
// with fresh credentials.
export function ImportUsersModal({ open, onClose }: { open: boolean; onClose: () => void }) {
  const imp = useImportUsers();
  const [source, setSource] = useState<"3xui" | "marzban">("3xui");
  const [fileName, setFileName] = useState("");
  const [data, setData] = useState<unknown>(null);
  const [result, setResult] = useState<Result | null>(null);
  const [error, setError] = useState("");

  function reset() {
    setSource("3xui"); setFileName(""); setData(null); setResult(null); setError("");
  }
  function close() { reset(); onClose(); }

  async function onFile(e: React.ChangeEvent<HTMLInputElement>) {
    setError("");
    const file = e.target.files?.[0];
    if (!file) return;
    setFileName(file.name);
    try {
      setData(JSON.parse(await file.text()));
    } catch {
      setError("Not valid JSON");
      setData(null);
    }
  }

  async function submit() {
    setError("");
    if (data == null) { setError("Choose an export file first"); return; }
    try {
      const res = await imp.mutateAsync({ source, data });
      setResult({ parsed: res.parsed, created_count: res.created_count, failures: res.failures });
    } catch {
      setError("Import failed — check the file and source format");
    }
  }

  return (
    <Modal open={open} onClose={close} title={result ? "Import finished" : "Import users"}>
      {result ? (
        <div className="space-y-4">
          <p className="text-sm text-fg">
            Parsed <span className="font-semibold">{result.parsed}</span> · created{" "}
            <span className="font-semibold text-success">{result.created_count}</span>.
          </p>
          {result.failures.length > 0 && (
            <div>
              <p className="mb-1 text-xs font-medium text-danger">{result.failures.length} skipped</p>
              <div className="max-h-32 space-y-1 overflow-auto rounded-md border border-border/60 p-2 text-xs">
                {result.failures.map((f) => (
                  <div key={f.username} className="flex justify-between gap-2">
                    <span className="font-mono">{f.username}</span>
                    <span className="text-fg-muted">{f.error}</span>
                  </div>
                ))}
              </div>
            </div>
          )}
          <Button variant="ghost" className="w-full" onClick={close}>Done</Button>
        </div>
      ) : (
        <div className="space-y-4">
          <div>
            <label className="mb-1 block text-xs font-medium text-fg-muted">Source panel</label>
            <Select value={source} onChange={(e) => setSource(e.target.value as "3xui" | "marzban")}>
              <option value="3xui">3x-ui (inbound / clients export)</option>
              <option value="marzban">Marzban (users export)</option>
            </Select>
          </div>

          <label className="flex cursor-pointer flex-col items-center justify-center gap-2 rounded-xl border border-dashed border-border/70 px-4 py-8 text-center transition hover:border-primary/50 hover:bg-surface-2/30">
            <UploadCloud size={22} className="text-fg-muted" />
            <span className="text-sm text-fg-muted">{fileName || "Choose a JSON export file"}</span>
            <input type="file" accept="application/json,.json" className="hidden" onChange={onFile} />
          </label>

          <p className="text-xs text-fg-subtle">
            Credentials are regenerated on import — users receive fresh subscription links. Assign inbounds after importing.
          </p>

          {error && <p className="text-sm text-danger">{error}</p>}
          <div className="flex justify-end gap-2 pt-1">
            <Button type="button" variant="ghost" onClick={close}>Cancel</Button>
            <Button onClick={submit} disabled={imp.isPending || data == null}>
              {imp.isPending ? "Importing…" : "Import"}
            </Button>
          </div>
        </div>
      )}
    </Modal>
  );
}
