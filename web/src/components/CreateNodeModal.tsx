import { useEffect, useState } from "react";
import { Check, Copy, Loader2 } from "lucide-react";
import { QRCodeSVG } from "qrcode.react";
import { useCreateNode, useNodeEnrollment, useTestNodeConnection } from "@/api/hooks";
import type { NodeDiagnostics, NodeEnrollmentPhase } from "@/api/types";
import { SERVER_LOCATIONS } from "@/lib/serverLocations";
import type { CoreType } from "@/lib/coreTypes";
import { EnabledCoresPicker } from "./EnabledCoresPicker";
import { Button, Input, Select } from "./ui";
import { Modal } from "./Modal";
import { useToast } from "./toast";

const AUTO_LOCATION = "__auto__";

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

function phaseLabel(p: NodeEnrollmentPhase): string {
  switch (p) {
    case "synced": return "Synced";
    case "connected": return "Connected";
    default: return "Pending";
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
  const [core, setCore] = useState<CoreType>("xray");
  const [enabledCores, setEnabledCores] = useState<CoreType[]>(["xray"]);
  const [endpoint, setEndpoint] = useState("");
  const [locationKey, setLocationKey] = useState(AUTO_LOCATION);
  const [error, setError] = useState("");
  const [createdId, setCreatedId] = useState<string | null>(null);
  const [diag, setDiag] = useState<NodeDiagnostics | null>(null);
  const [phase, setPhase] = useState<NodeEnrollmentPhase>("pending");
  const [panelCA, setPanelCA] = useState("");
  const [caMatch, setCaMatch] = useState<boolean | null>(null);
  const [copied, setCopied] = useState(false);

  useEffect(() => {
    if (open && step === 0) enroll.refetch();
  }, [open, step, enroll]);

  function close() {
    setStep(0);
    setName("");
    setAddress("");
    setCore("xray");
    setEnabledCores(["xray"]);
    setEndpoint("");
    setLocationKey(AUTO_LOCATION);
    setError("");
    setCreatedId(null);
    setDiag(null);
    setPhase("pending");
    setPanelCA("");
    setCaMatch(null);
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
      const preset = SERVER_LOCATIONS.find((p) => `${p.code}-${p.city}` === locationKey);
      const res = await create.mutateAsync({
        name, address, core, enabled_cores: enabledCores, endpoint: endpoint || undefined,
        location_auto: !preset,
        ...(preset ? { region: preset.label, country_code: preset.code } : {}),
      });
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
    setCaMatch(null);
    try {
      const res = await test.mutateAsync(nodeId);
      setDiag(res.diagnostics);
      setPanelCA(res.panel_ca_fingerprint ?? enroll.data?.ca_fingerprint ?? "");
      setCaMatch(res.ca_match ?? res.diagnostics.ca_match ?? null);
      if (res.enrollment_phase) setPhase(res.enrollment_phase);
    } catch {
      setDiag({ code: "unknown", message: "Test request failed" });
    }
  }

  const bundle = enroll.data;
  const bundleLarge = (bundle?.bundle.length ?? 0) > 2400;

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
            Copy the enrollment bundle or scan the QR on the new server during install (option 2).
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
              <div className="flex flex-col items-center gap-2 rounded-xl bg-surface-2/40 p-4">
                <QRCodeSVG value={bundle.bundle} size={bundleLarge ? 200 : 160} level="L" />
                {bundleLarge && (
                  <p className="text-center text-[10px] text-fg-subtle">
                    Large bundle — prefer Copy if the QR does not scan.
                  </p>
                )}
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
            <li>Paste the bundle or scan the QR when asked (option 2).</li>
            <li>Ensure port <span className="font-mono">50051</span> is open to the panel.</li>
            {enabledCores.length > 1 && (
              <li>
                For dual-core, set on the node:{" "}
                <span className="font-mono text-fg">VORTEX_ENABLED_CORES=xray,singbox</span>
                {" "}plus separate config paths (<span className="font-mono">VORTEX_XRAY_CONFIG</span>,{" "}
                <span className="font-mono">VORTEX_SINGBOX_CONFIG</span>).
              </li>
            )}
            <li>On the node, verify CA fingerprint matches the panel value shown in step 1.</li>
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
          <EnabledCoresPicker
            value={enabledCores}
            onChange={setEnabledCores}
            defaultCore={core}
            onDefaultCoreChange={setCore}
          />
          <Input placeholder="Endpoint (optional — tunnel/CDN IP or domain)" value={endpoint} onChange={(e) => setEndpoint(e.target.value)} />
          <label className="block text-xs text-fg-subtle">
            Server location
            <Select className="mt-1" value={locationKey} onChange={(e) => setLocationKey(e.target.value)}>
              <option value={AUTO_LOCATION}>Auto-detect from IP (GeoIP)</option>
              {SERVER_LOCATIONS.map((loc) => (
                <option key={`${loc.code}-${loc.city}`} value={`${loc.code}-${loc.city}`}>{loc.label}</option>
              ))}
            </Select>
          </label>
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
              <div className="flex flex-wrap items-center gap-2">
                <span className={`inline-flex rounded-full px-2.5 py-0.5 text-[11px] font-semibold uppercase ring-1 ring-inset ${
                  diag.code === "ok" ? "bg-success/12 text-success ring-success/20" : "bg-danger/12 text-danger ring-danger/20"
                }`}>
                  {diagLabel(diag.code)}
                </span>
                <span className="inline-flex rounded-full bg-surface-2/80 px-2.5 py-0.5 text-[11px] font-semibold uppercase text-fg-muted">
                  {phaseLabel(phase)}
                </span>
                {diag.network_reachable && (
                  <span className="inline-flex rounded-full bg-success/12 px-2.5 py-0.5 text-[11px] font-semibold uppercase text-success ring-1 ring-inset ring-success/20">
                    Network OK
                  </span>
                )}
                {caMatch === true && (
                  <span className="inline-flex rounded-full bg-success/12 px-2.5 py-0.5 text-[11px] font-semibold uppercase text-success ring-1 ring-inset ring-success/20">
                    CA match
                  </span>
                )}
                {caMatch === false && (
                  <span className="inline-flex rounded-full bg-danger/12 px-2.5 py-0.5 text-[11px] font-semibold uppercase text-danger ring-1 ring-inset ring-danger/20">
                    CA mismatch
                  </span>
                )}
              </div>
              {panelCA && (
                <p className="text-[10px] text-fg-subtle" dir="ltr">
                  Panel CA: <span className="font-mono">{panelCA}</span>
                </p>
              )}
              {diag.message && <p className="text-xs text-fg-muted">{diag.message}</p>}
              {diag.code === "mtls_fail" && (
                <p className="text-[10px] text-fg-subtle">
                  Network may be OK but certs differ — re-copy from the panel (<span className="font-mono">vortexui node-bundle</span>) and restart <span className="font-mono">vortexui-node</span>.
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

export { diagLabel, diagColor, phaseLabel };
