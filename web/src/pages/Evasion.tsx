import { useState } from "react";
import { Shield } from "lucide-react";
import { Button, Card, Input, PageHeader, Select } from "@/components/ui";
import { Modal } from "@/components/Modal";
import { useToast } from "@/components/toast";

interface EvasionProfile {
  id: string;
  name: string;
  description: string;
  fragment_enabled: boolean;
  fragment_length: string;
  fragment_interval: string;
  fingerprint: string;
  mux_enabled: boolean;
  mux_protocol: string;
  enabled: boolean;
}

// Prebuilt defaults for demo (real data comes from API)
const DEFAULTS: EvasionProfile[] = [
  { id: "1", name: "Iran (Fragment + Chrome)", description: "TLS fragment + Chrome fingerprint", fragment_enabled: true, fragment_length: "10-30", fragment_interval: "10-20", fingerprint: "chrome", mux_enabled: false, mux_protocol: "", enabled: true },
  { id: "2", name: "China (Mux + Random)", description: "Multiplexed + randomized fingerprint", fragment_enabled: false, fragment_length: "", fragment_interval: "", fingerprint: "randomized", mux_enabled: true, mux_protocol: "h2mux", enabled: true },
  { id: "3", name: "Russia (Fragment + Firefox)", description: "Fragment for TSPU bypass", fragment_enabled: true, fragment_length: "1-3", fragment_interval: "5-10", fingerprint: "firefox", mux_enabled: false, mux_protocol: "", enabled: true },
];

export function Evasion() {
  const [profiles, setProfiles] = useState<EvasionProfile[]>(DEFAULTS);
  const [createOpen, setCreateOpen] = useState(false);
  const toast = useToast();

  function remove(id: string) {
    setProfiles(p => p.filter(x => x.id !== id));
    toast.success("Profile removed");
  }

  return (
    <div className="space-y-6 animate-fade-in">
      <div className="flex items-center justify-between">
        <PageHeader title="Evasion Profiles" />
        <Button onClick={() => setCreateOpen(true)}>New profile</Button>
      </div>

      <p className="text-sm text-fg-muted">Anti-DPI presets. Link a profile to inbounds for one-click evasion hardening.</p>

      <div className="grid grid-cols-1 gap-4 md:grid-cols-2 lg:grid-cols-3">
        {profiles.map((p) => (
          <Card key={p.id} className="space-y-3">
            <div className="flex items-center justify-between">
              <div className="flex items-center gap-2">
                <Shield size={14} className="text-primary" />
                <h3 className="text-sm font-bold text-fg">{p.name}</h3>
              </div>
              <span className={`h-2 w-2 rounded-full ${p.enabled ? "bg-success" : "bg-fg-subtle"}`} />
            </div>
            {p.description && <p className="text-xs text-fg-muted">{p.description}</p>}
            <div className="flex flex-wrap gap-1.5">
              {p.fragment_enabled && <span className="rounded-md bg-accent/10 px-2 py-0.5 text-[10px] font-medium text-accent">Fragment {p.fragment_length}</span>}
              {p.mux_enabled && <span className="rounded-md bg-warning/10 px-2 py-0.5 text-[10px] font-medium text-warning">Mux {p.mux_protocol}</span>}
              {p.fingerprint && <span className="rounded-md bg-primary/10 px-2 py-0.5 text-[10px] font-medium text-primary">FP: {p.fingerprint}</span>}
            </div>
            <div className="flex justify-end">
              <Button variant="ghost" className="text-destructive text-xs" onClick={() => remove(p.id)}>Delete</Button>
            </div>
          </Card>
        ))}
      </div>

      {createOpen && <CreateEvasionModal onClose={() => setCreateOpen(false)} onCreate={(p) => { setProfiles(prev => [...prev, p]); setCreateOpen(false); }} />}
    </div>
  );
}

function CreateEvasionModal({ onClose, onCreate }: { onClose: () => void; onCreate: (p: EvasionProfile) => void }) {
  const [name, setName] = useState("");
  const [fp, setFp] = useState("chrome");
  const [frag, setFrag] = useState(true);
  const [fragLen, setFragLen] = useState("10-30");
  const [mux, setMux] = useState(false);
  const [muxProto, setMuxProto] = useState("smux");

  function submit(e: React.FormEvent) {
    e.preventDefault();
    onCreate({
      id: crypto.randomUUID(),
      name, description: "",
      fragment_enabled: frag, fragment_length: fragLen, fragment_interval: "10-20",
      fingerprint: fp, mux_enabled: mux, mux_protocol: muxProto, enabled: true,
    });
  }

  return (
    <Modal open onClose={onClose} title="New Evasion Profile">
      <form onSubmit={submit} className="space-y-3">
        <Input placeholder="Profile name" value={name} onChange={(e) => setName(e.target.value)} required />
        <Select value={fp} onChange={(e) => setFp(e.target.value)}>
          <option value="chrome">Chrome</option>
          <option value="firefox">Firefox</option>
          <option value="safari">Safari</option>
          <option value="random">Random</option>
          <option value="randomized">Randomized</option>
        </Select>
        <label className="flex items-center gap-2 text-sm">
          <input type="checkbox" checked={frag} onChange={(e) => setFrag(e.target.checked)} />
          Enable Fragment
        </label>
        {frag && <Input placeholder="Fragment length (e.g. 10-30)" value={fragLen} onChange={(e) => setFragLen(e.target.value)} />}
        <label className="flex items-center gap-2 text-sm">
          <input type="checkbox" checked={mux} onChange={(e) => setMux(e.target.checked)} />
          Enable Mux
        </label>
        {mux && (
          <Select value={muxProto} onChange={(e) => setMuxProto(e.target.value)}>
            <option value="smux">smux</option>
            <option value="yamux">yamux</option>
            <option value="h2mux">h2mux</option>
          </Select>
        )}
        <div className="flex justify-end gap-2 pt-2">
          <Button type="button" variant="ghost" onClick={onClose}>Cancel</Button>
          <Button type="submit">Create</Button>
        </div>
      </form>
    </Modal>
  );
}
