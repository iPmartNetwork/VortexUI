import { useState } from "react";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { api } from "@/api/client";
import { Button, Card, Input, PageHeader, Badge, Select } from "@/components/ui";
import { Modal } from "@/components/Modal";
import { useToast } from "@/components/toast";
import { useConfirm } from "@/components/confirm";
import { useI18n } from "@/i18n/i18n";

interface SNIDomain { id: string; inbound_id: string; domain: string; auto_cert: boolean; cert_status: string; expires_at: string | null; }
interface SSLCert { id: string; domain: string; wildcard: boolean; issuer: string; status: string; auto_renew: boolean; expires_at: string | null; }
export interface SNIRoute { id: string; inbound_id: string; sni: string; action: string; target_tag: string; priority: number; enabled: boolean; }

export function SNIManager() {
  const { t } = useI18n();
  const qc = useQueryClient();
  const toast = useToast();
  const confirm = useConfirm();
  const [addDomainOpen, setAddDomainOpen] = useState(false);
  const [addCertOpen, setAddCertOpen] = useState(false);

  const { data: domainsData } = useQuery({ queryKey: ["sni-domains"], queryFn: () => api<{ domains: SNIDomain[] }>("/api/sni/domains") });
  const { data: certsData } = useQuery({ queryKey: ["sni-certs"], queryFn: () => api<{ certificates: SSLCert[] }>("/api/sni/certs") });

  const delDomain = useMutation({ mutationFn: (id: string) => api<void>(`/api/sni/domains/${id}`, { method: "DELETE" }), onSuccess: () => qc.invalidateQueries({ queryKey: ["sni-domains"] }) });
  const delCert = useMutation({ mutationFn: (id: string) => api<void>(`/api/sni/certs/${id}`, { method: "DELETE" }), onSuccess: () => qc.invalidateQueries({ queryKey: ["sni-certs"] }) });
  const renewCert = useMutation({ mutationFn: (id: string) => api(`/api/sni/certs/${id}/renew`, { method: "POST" }), onSuccess: () => { qc.invalidateQueries({ queryKey: ["sni-certs"] }); toast.success("Renewal started"); } });

  return (
    <div className="space-y-6 animate-page-enter">
      <PageHeader title={t("sni.title")} subtitle={t("sni.subtitle")} />

      <div className="rounded-lg border border-border/40 bg-surface-2/20 p-4 text-xs text-fg-muted space-y-2">
        <p className="font-medium text-fg text-sm">{t("sni.infoTitle")}</p>
        <p>{t("sni.infoDesc")}</p>
      </div>

      {/* Domains */}
      <Card>
        <div className="flex items-center justify-between mb-3">
          <h3 className="text-sm font-bold text-fg">{t("sni.domains")}</h3>
          <Button size="sm" onClick={() => setAddDomainOpen(true)}>{t("sni.addDomain")}</Button>
        </div>
        <AddDomainModal open={addDomainOpen} onClose={() => setAddDomainOpen(false)} />
        <div className="space-y-2">
          {domainsData?.domains?.map(d => (
            <div key={d.id} className="flex items-center justify-between rounded-lg bg-surface-2/40 px-3 py-2">
              <div className="flex items-center gap-3">
                <span className="font-mono text-sm text-fg">{d.domain}</span>
                <Badge color={d.cert_status === "active" ? "active" : d.cert_status === "pending" ? "limited" : "expired"}>{d.cert_status}</Badge>
                {d.auto_cert && <span className="text-xs text-fg-subtle">Auto</span>}
              </div>
              <Button variant="ghost" size="sm" className="text-destructive" onClick={async () => { if (await confirm({ title: "Delete domain?", destructive: true })) { delDomain.mutate(d.id); } }}>Delete</Button>
            </div>
          ))}
          {(!domainsData?.domains || domainsData.domains.length === 0) && <p className="text-xs text-fg-muted text-center py-4">{t("sni.noDomains")}</p>}
        </div>
      </Card>

      {/* Certificates */}
      <Card>
        <div className="flex items-center justify-between mb-3">
          <h3 className="text-sm font-bold text-fg">{t("sni.certificates")}</h3>
          <Button size="sm" onClick={() => setAddCertOpen(true)}>{t("sni.issueCert")}</Button>
        </div>
        <AddCertModal open={addCertOpen} onClose={() => setAddCertOpen(false)} />
        <div className="space-y-2">
          {certsData?.certificates?.map(c => (
            <div key={c.id} className="flex items-center justify-between rounded-lg bg-surface-2/40 px-3 py-2">
              <div className="flex items-center gap-3">
                <span className="font-mono text-sm text-fg">{c.wildcard ? "*." : ""}{c.domain}</span>
                <Badge color={c.status === "active" ? "active" : c.status === "pending" ? "limited" : "expired"}>{c.status}</Badge>
                <span className="text-xs text-fg-subtle">{c.issuer}</span>
                {c.expires_at && <span className="text-xs text-fg-muted">exp: {new Date(c.expires_at).toLocaleDateString()}</span>}
              </div>
              <div className="flex gap-1">
                <Button variant="ghost" size="sm" onClick={() => renewCert.mutate(c.id)}>Renew</Button>
                <Button variant="ghost" size="sm" className="text-destructive" onClick={() => delCert.mutate(c.id)}>Delete</Button>
              </div>
            </div>
          ))}
          {(!certsData?.certificates || certsData.certificates.length === 0) && <p className="text-xs text-fg-muted text-center py-4">{t("sni.noCerts")}</p>}
        </div>
      </Card>
    </div>
  );
}

function AddDomainModal({ open, onClose }: { open: boolean; onClose: () => void }) {
  const qc = useQueryClient(); const toast = useToast();
  const [f, setF] = useState({ inbound_id: "", domain: "", auto_cert: true });
  const create = useMutation({ mutationFn: (b: any) => api("/api/sni/domains", { method: "POST", body: b }), onSuccess: () => { qc.invalidateQueries({ queryKey: ["sni-domains"] }); onClose(); toast.success("Domain added"); } });
  return (<Modal open={open} onClose={onClose} title="Add Domain"><form onSubmit={e => { e.preventDefault(); create.mutate(f); }} className="space-y-3"><Input placeholder="Inbound ID" value={f.inbound_id} onChange={e => setF(s => ({...s, inbound_id: e.target.value}))} required /><Input placeholder="domain.com" value={f.domain} onChange={e => setF(s => ({...s, domain: e.target.value}))} required /><label className="flex items-center gap-2 text-sm"><input type="checkbox" checked={f.auto_cert} onChange={e => setF(s => ({...s, auto_cert: e.target.checked}))} /> Auto-provision SSL</label><div className="flex justify-end gap-2"><Button type="button" variant="ghost" onClick={onClose}>Cancel</Button><Button type="submit" disabled={create.isPending}>Add</Button></div></form></Modal>);
}

function AddCertModal({ open, onClose }: { open: boolean; onClose: () => void }) {
  const qc = useQueryClient(); const toast = useToast();
  const [f, setF] = useState({ domain: "", wildcard: false, issuer: "letsencrypt", auto_renew: true });
  const create = useMutation({ mutationFn: (b: any) => api("/api/sni/certs", { method: "POST", body: b }), onSuccess: () => { qc.invalidateQueries({ queryKey: ["sni-certs"] }); onClose(); toast.success("Certificate issued"); } });
  return (<Modal open={open} onClose={onClose} title="Issue Certificate"><form onSubmit={e => { e.preventDefault(); create.mutate(f); }} className="space-y-3"><Input placeholder="domain.com" value={f.domain} onChange={e => setF(s => ({...s, domain: e.target.value}))} required /><Select value={f.issuer} onChange={e => setF(s => ({...s, issuer: e.target.value}))}><option value="letsencrypt">Let's Encrypt</option><option value="zerossl">ZeroSSL</option><option value="custom">Custom</option></Select><label className="flex items-center gap-2 text-sm"><input type="checkbox" checked={f.wildcard} onChange={e => setF(s => ({...s, wildcard: e.target.checked}))} /> Wildcard (*)</label><label className="flex items-center gap-2 text-sm"><input type="checkbox" checked={f.auto_renew} onChange={e => setF(s => ({...s, auto_renew: e.target.checked}))} /> Auto-renew</label><div className="flex justify-end gap-2"><Button type="button" variant="ghost" onClick={onClose}>Cancel</Button><Button type="submit" disabled={create.isPending}>Issue</Button></div></form></Modal>);
}
