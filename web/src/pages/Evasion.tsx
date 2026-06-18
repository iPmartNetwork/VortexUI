import { useState } from "react";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { Shield } from "lucide-react";
import { api } from "@/api/client";
import { Button, Card, Input, PageHeader, Select } from "@/components/ui";
import { Modal } from "@/components/Modal";
import { useToast } from "@/components/toast";
import { useI18n } from "@/i18n/i18n";

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

export function Evasion() {
  const { t } = useI18n();
  const qc = useQueryClient();
  const [createOpen, setCreateOpen] = useState(false);
  const toast = useToast();

  const { data } = useQuery({
    queryKey: ["tls-tricks"],
    queryFn: () => api<{ profiles: EvasionProfile[] }>("/api/tls-tricks"),
  });

  const delMut = useMutation({
    mutationFn: (id: string) => api<void>(`/api/tls-tricks/${id}`, { method: "DELETE" }),
    onSuccess: () => { qc.invalidateQueries({ queryKey: ["tls-tricks"] }); toast.success("Profile removed"); },
  });

  const profiles = data?.profiles ?? [];

  return (
    <div className="space-y-6 animate-fade-in">
      <div className="flex items-center justify-between">
        <PageHeader title={t("evasion.title")} />
        <Button onClick={() => setCreateOpen(true)}>{t("evasion.newProfile")}</Button>
      </div>

      <div className="rounded-lg border border-border/40 bg-surface-2/20 p-4 text-xs text-fg-muted space-y-2">
        <p className="font-medium text-fg text-sm">{t("evasion.infoTitle")}</p>
        <p>{t("evasion.infoDesc")}</p>
        <ul className="list-disc pl-4 space-y-1">
          <li><strong>Fragment</strong> — {t("evasion.fragment")}</li>
          <li><strong>Mux</strong> — {t("evasion.mux")}</li>
          <li><strong>Fingerprint</strong> — {t("evasion.fingerprint")}</li>
        </ul>
      </div>

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
              <Button variant="ghost" className="text-destructive text-xs" onClick={() => delMut.mutate(p.id)}>Delete</Button>
            </div>
          </Card>
        ))}
      </div>

      {createOpen && <CreateEvasionModal onClose={() => setCreateOpen(false)} />}
    </div>
  );
}

function CreateEvasionModal({ onClose }: { onClose: () => void }) {
  const qc = useQueryClient();
  const toast = useToast();
  const [name, setName] = useState("");
  const [fp, setFp] = useState("chrome");
  const [frag, setFrag] = useState(true);
  const [fragLen, setFragLen] = useState("10-30");
  const [mux, setMux] = useState(false);
  const [muxProto, setMuxProto] = useState("smux");

  const create = useMutation({
    mutationFn: (body: any) => api("/api/tls-tricks", { method: "POST", body }),
    onSuccess: () => { qc.invalidateQueries({ queryKey: ["tls-tricks"] }); onClose(); toast.success("Profile created"); },
    onError: (e: any) => toast.error(e.message),
  });

  function submit(e: React.FormEvent) {
    e.preventDefault();
    create.mutate({
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
