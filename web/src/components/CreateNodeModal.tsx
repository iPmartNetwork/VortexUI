import { useEffect, useState } from "react";
import { Check, Copy, Loader2 } from "lucide-react";
import { useCreateNode, useNodeEnrollment, useTestNodeConnection } from "@/api/hooks";
import type { NodeDiagnostics } from "@/api/types";
import { Button, Input, Select } from "./ui";
import { Modal } from "./Modal";
import { useToast } from "./toast";

const STEPS = ["Bundle", "Install", "Register", "Test"] as const;

function diagLabel(code: NodeDiagnostics["code"]): string {
  switch (code) {
    case "ok": return "Connected";
    case "mtls_fail": return "mTLS fail";
    case "unreachable": return "Unreachable";
    case "core_down": return "Core down";
    default: return "Unknown";
  }
}

function diagColor(code: NodeDiagnostics["code"]): string {
  switch (code) {
    case "ok": return "running";
    case "mtls_fail": return "down";
    case "unreachable": return "down";
    case "core_down": return "on_hold";
    default: return "muted";
  }
}

export function CreateNodeModal({ open, onClose }: { open: boolean; onClose: () => void }) {
  const create = useCreateNode();
  const enroll = useNodeEnrollment();
  const test = useTestNodeConnection();
  const toast = useToast();
  const [step, setStep] = useState(0);
  const [name, setName] = useState("");
  const [address, setAddress] = useState("");
  const [core, setCore] = useState("xray");
  const [endpoint, setEndpoint] = useState("");
  const [error, setError] = useState("");
  const [createdId, setCreatedId] = useState<string | null>(null);
  const [diag, setDiag] = useState<NodeDiagnostics | null>(null);
  const [copied, setCopied] = useState(false);

  useEffect(() => {
    if (open && step === 0) enroll.refetch();
  }, [open, step, enroll]);

  function close() {
    setStep(0);
    setName("");
    setAddress("");
    setCore("xray");
    setEndpoint("");
    setError("");
    setCreatedId(null);
    setDiag(null);
    setCopied(false);
    onClose();
  }

  async function copyBundle() {
    const b = enroll.data?.bundle;
    if (!b) return;
    await navigator.clipboard.writeText(b);
    setCopied(true);
    toast.success("Bundle copied");
    setTimeout(() => setCopied(false), 2000);
  }

  async function register(e: React.FormEvent) {
    e.preventDefault();
    setError("");
    try {
      const res = await create.mutateAsync({ name, address, core, endpoint: endpoint || undefined });
      setCreatedId(res.node.id);
      setStep(3);
      runTest(res.node.id);
    } catch {
      setError("Could not create node (name taken?)");
    }
  }

  async function runTest(id?: string) {
    const nodeId = id ?? createdId;
    if (!nodeId) return;
    setDiag(null);
    try {
      const res = await test.mutateAsync(nodeId);
      setDiag(res.diagnostics);
    } catch {
      setDiag({ code: "unknown", message: "Test request failed" });
    }
  }

  const bundle = enroll.data;

  return (
    <Modal open={open} onClose={close} title="Add node" className="max-w-lg">
      <div className="mb-4 flex gap-1">
        {STEPS.map((label, i) => (
          <div
            key={label}
            className={`flex-1 rounded-md py-1 text-center text-[10px] font-semibold uppercase tracking-wide ${
              i === step ? "bg-primary/15 text-primary" : i < step ? "bg-success/10 text-success" : "bg-surface-2/60 text-fg-subtle"
            }`}
          >
            {label}
          </div>
        ))}
      </div>

      {step === 0 && (
        <div className="space-y-3">
          <p className="text-sm text-fg-muted">
            Copy the enrollment bundle. You will paste it on the new server during install (option 2).
          </p>
          {enroll.isFetching && (
            <div className="flex items-center gap-2 text-sm text-fg-muted">
              <Loader2 size={14} className="animate-spin" /> Loading bundle…
            </div>
          )}
          {enroll.isError && <p className="text-sm text-destructive">Could not load bundle — check panel certs.</p>}
          {bundle && (
            <>
              <div className="rounded-xl bg-surface-2/50 p-3">
                <div className="text-[10px] font-semibold uppercase text-fg-subtle">Panel CA fingerprint</div>
                <div className="mt-1 break-all font-mono text-[11px] text-fg" dir="ltr">{bundle.ca_fingerprint}</div>
              </div>
              <Button type="button" variant="outline" className="w-full" onClick={copyBundle}>
                {copied ? <Check size={14} /> : <Copy size={14} />}
                {copied ? "Copied" : "Copy enrollment bundle"}
              </Button>
              <p className="text-[10px] text-fg-subtle">
                CLI alternative: run <span className="font-mono">vortexui node-bundle</span> on the panel server.
              </p>
            </>
          )}
          <div className="flex justify-end gap-2 pt-1">
            <Button type="button" variant="ghost" onClick={close}>Cancel</Button>
            <Button type="button" disabled={!bundle} onClick={() => setStep(1)}>Next</Button>
          </div>
        </div>
      )}

      {step === 1 && (
        <div className="space-y-3 text-sm text-fg-muted">
          <ol className="list-decimal space-y-2 ps-4">
            <li>SSH into the new server and run the VortexUI installer.</li>
            <li>Choose <strong>Node</strong> when prompted.</li>
            <li>Paste the enrollment bundle when asked (option 2).</li>
            <li>Ensure port <span className="font-mono">50051</span> is open to the panel.</li>
            <li>On the node, verify CA fingerprint: <span className="font-mono text-[11px]">openssl x509 -in /etc/vortexui/certs/ca.crt -noout -fingerprint -sha256</span></li>
          </ol>
          <p className="text-[10px] text-fg-subtle">After install, run <span className="font-mono">vortexui doctor</span> on the node to verify certs and the agent.</p>
          <div className="flex justify-between gap-2 pt-1">
            <Button type="button" variant="ghost" onClick={() => setStep(0)}>Back</Button>
            <Button type="button" onClick={() => setStep(2)}>Next</Button>
          </div>
        </div>
      )}

      {step === 2 && (
        <form onSubmit={register} className="space-y-3">
          <p className="text-sm text-fg-muted">Register the node in the panel using its public IP and agent port.</p>
          <Input placeholder="Name (e.g. fr-1)" value={name} onChange={(e) => setName(e.target.value)} required autoFocus />
          <Input placeholder="Agent address (host:50051)" value={address} onChange={(e) => setAddress(e.target.value)} required />
          <Select value={core} onChange={(e) => setCore(e.target.value)}>
            <option value="xray">Xray-core</option>
            <option value="singbox">sing-box</option>
          </Select>
          <Input placeholder="Endpoint (optional — tunnel/CDN IP or domain)" value={endpoint} onChange={(e) => setEndpoint(e.target.value)} />
          {error && <p className="text-sm text-destructive">{error}</p>}
          <div className="flex justify-between gap-2 pt-1">
            <Button type="button" variant="ghost" onClick={() => setStep(1)}>Back</Button>
            <Button type="submit" disabled={create.isPending}>{create.isPending ? "Adding…" : "Add & test"}</Button>
          </div>
        </form>
      )}

      {step === 3 && (
        <div className="space-y-3">
          <p className="text-sm text-fg-muted">Connectivity test from the panel to <span className="font-mono text-fg">{address}</span></p>
          {test.isPending && (
            <div className="flex items-center gap-2 text-sm text-fg-muted">
              <Loader2 size={14} className="animate-spin" /> Testing connection…
            </div>
          )}
          {diag && (
            <div className="rounded-xl bg-surface-2/50 p-3 space-y-2">
              <div className="flex items-center gap-2">
                <span className={`inline-flex rounded-full px-2.5 py-0.5 text-[11px] font-semibold uppercase ring-1 ring-inset ${
                  diag.code === "ok" ? "bg-success/12 text-success ring-success/20" : "bg-danger/12 text-danger ring-danger/20"
                }`}>
                  {diagLabel(diag.code)}
                </span>
              </div>
              {diag.message && <p className="text-xs text-fg-muted">{diag.message}</p>}
              {diag.code === "mtls_fail" && (
                <p className="text-[10px] text-fg-subtle">
                  Re-copy certs from the panel (<span className="font-mono">vortexui node-bundle</span>) and restart <span className="font-mono">vortexui-node</span>.
                </p>
              )}
            </div>
          )}
          <div className="flex justify-between gap-2 pt-1">
            <Button type="button" variant="ghost" onClick={() => runTest()} disabled={test.isPending}>Retest</Button>
            <Button type="button" onClick={() => { toast.success(`Node ${name} added`); close(); }}>
              {diag?.code === "ok" ? "Done" : "Close anyway"}
            </Button>
          </div>
        </div>
      )}
    </Modal>
  );
}

export { diagLabel, diagColor };
