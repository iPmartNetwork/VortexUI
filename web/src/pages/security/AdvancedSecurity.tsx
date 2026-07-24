import { useState } from "react";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { Shield, Key, Globe, Clock, AlertTriangle, Trash2, Plus, Ban, Activity } from "lucide-react";
import { api } from "@/api/client";
import { Button, Input, Badge } from "@/components/ui";
import { GlassCard } from "@/components/veltrix";
import { Modal } from "@/components/Modal";
import { useToast } from "@/components/toast";
import { useConfirm } from "@/components/confirm";
import { useTitle } from "@/lib/useTitle";
import { useI18n } from "@/i18n/i18n";

interface Session { id: string; ip_address: string; user_agent: string; country: string; last_active: string; revoked: boolean; }
interface LoginEntry { id: string; username: string; ip_address: string; country: string; success: boolean; failure_reason: string; created_at: string; }
interface AuditEntry { id: string; operation: string; resource: string; ip_address: string; created_at: string; }
interface WhitelistEntry { id: string; cidr: string; description: string; }
interface BanEntry { id: string; ip_address: string; reason: string; expires_at?: string; }

type Tab = "sessions"|"login"|"audit"|"whitelist"|"bans"|"tokens";

export function AdvancedSecurity() {
  const { t: _t } = useI18n();
  useTitle("Advanced Security");
  const qc = useQueryClient();
  const toast = useToast();
  const confirm = useConfirm();
  const [tab, setTab] = useState<Tab>("sessions");
  const [showWL, setShowWL] = useState(false);

  const { data: sessions } = useQuery({ queryKey: ["sec-sessions"], queryFn: () => api<Session[]>("/api/v2/security/sessions"), enabled: tab==="sessions" });
  const { data: logins } = useQuery({ queryKey: ["sec-logins"], queryFn: () => api<LoginEntry[]>("/api/v2/security/login-audit"), enabled: tab==="login" });
  const { data: audits } = useQuery({ queryKey: ["sec-audit"], queryFn: () => api<AuditEntry[]>("/api/v2/security/audit-log"), enabled: tab==="audit" });
  const { data: whitelist } = useQuery({ queryKey: ["sec-wl"], queryFn: () => api<WhitelistEntry[]>("/api/v2/security/ip-whitelist"), enabled: tab==="whitelist" });
  const { data: bans } = useQuery({ queryKey: ["sec-bans"], queryFn: () => api<BanEntry[]>("/api/v2/security/bans"), enabled: tab==="bans" });

  const revokeMut = useMutation({ mutationFn: (id: string) => api(`/api/v2/security/sessions/${id}`, { method: "DELETE" }), onSuccess: () => { qc.invalidateQueries({ queryKey: ["sec-sessions"] }); toast.success("Revoked"); } });
  const revokeAllMut = useMutation({ mutationFn: () => api("/api/v2/security/sessions", { method: "DELETE" }), onSuccess: () => { qc.invalidateQueries({ queryKey: ["sec-sessions"] }); toast.success("All revoked"); } });
  const delWL = useMutation({ mutationFn: (id: string) => api(`/api/v2/security/ip-whitelist/${id}`, { method: "DELETE" }), onSuccess: () => { qc.invalidateQueries({ queryKey: ["sec-wl"] }); toast.success("Removed"); } });
  const delBan = useMutation({ mutationFn: (id: string) => api(`/api/v2/security/bans/${id}`, { method: "DELETE" }), onSuccess: () => { qc.invalidateQueries({ queryKey: ["sec-bans"] }); toast.success("Ban removed"); } });

  const tabs: { key: Tab; label: string; icon: React.ReactNode }[] = [
    { key: "sessions", label: "Sessions", icon: <Activity className="w-4 h-4"/> },
    { key: "login", label: "Login Audit", icon: <Clock className="w-4 h-4"/> },
    { key: "audit", label: "Audit Log", icon: <Shield className="w-4 h-4"/> },
    { key: "whitelist", label: "IP Whitelist", icon: <Globe className="w-4 h-4"/> },
    { key: "bans", label: "IP Bans", icon: <Ban className="w-4 h-4"/> },
    { key: "tokens", label: "API Tokens", icon: <Key className="w-4 h-4"/> },
  ];

  return (
    <div className="space-y-6 p-6">
      <h1 className="text-2xl font-bold flex items-center gap-2"><Shield className="w-6 h-6"/>Advanced Security</h1>
      <div className="flex gap-2 border-b border-border pb-2 overflow-x-auto">
        {tabs.map((t) => <Button key={t.key} variant={tab===t.key?"primary":"ghost"} size="sm" onClick={()=>setTab(t.key)} className="flex items-center gap-1">{t.icon}{t.label}</Button>)}
      </div>

      {tab === "sessions" && (
        <GlassCard className="p-4 space-y-4">
          <div className="flex items-center justify-between"><h2 className="text-lg font-semibold">Active Sessions</h2><Button variant="destructive" size="sm" onClick={async()=>{if(await confirm({title:"Revoke all?"}))revokeAllMut.mutate()}}>Revoke All</Button></div>
          {sessions && sessions.length > 0 ? (
            <table className="w-full text-sm"><thead><tr className="border-b"><th className="text-left py-2 px-3">IP</th><th className="text-left py-2 px-3">Country</th><th className="text-left py-2 px-3">Last Active</th><th className="text-left py-2 px-3">Status</th><th className="text-right py-2 px-3">Act</th></tr></thead>
            <tbody>{sessions.map((s)=>(
              <tr key={s.id} className="border-b hover:bg-surface-2/50"><td className="py-2 px-3 font-mono">{s.ip_address}</td><td className="py-2 px-3">{s.country||"—"}</td><td className="py-2 px-3">{new Date(s.last_active).toLocaleString()}</td><td className="py-2 px-3"><Badge color={s.revoked?"expired":"active"}>{s.revoked?"Revoked":"Active"}</Badge></td><td className="py-2 px-3 text-right">{!s.revoked&&<Button variant="ghost" size="sm" onClick={()=>revokeMut.mutate(s.id)}><Trash2 className="w-3 h-3"/></Button>}</td></tr>
            ))}</tbody></table>
          ) : <p className="text-fg-muted text-sm">No sessions.</p>}
        </GlassCard>
      )}

      {tab === "login" && (
        <GlassCard className="p-4 space-y-4">
          <h2 className="text-lg font-semibold">Login Audit</h2>
          {logins && logins.length > 0 ? (
            <table className="w-full text-sm"><thead><tr className="border-b"><th className="text-left py-2 px-3">User</th><th className="text-left py-2 px-3">IP</th><th className="text-left py-2 px-3">Result</th><th className="text-left py-2 px-3">Time</th></tr></thead>
            <tbody>{logins.map((e)=>(
              <tr key={e.id} className="border-b hover:bg-surface-2/50"><td className="py-2 px-3">{e.username}</td><td className="py-2 px-3 font-mono">{e.ip_address}</td><td className="py-2 px-3"><Badge color={e.success?"active":"expired"}>{e.success?"OK":"Fail"}</Badge></td><td className="py-2 px-3">{new Date(e.created_at).toLocaleString()}</td></tr>
            ))}</tbody></table>
          ) : <p className="text-fg-muted text-sm">No entries.</p>}
        </GlassCard>
      )}

      {tab === "audit" && (
        <GlassCard className="p-4 space-y-4">
          <h2 className="text-lg font-semibold">Security Audit</h2>
          {audits && audits.length > 0 ? (
            <table className="w-full text-sm"><thead><tr className="border-b"><th className="text-left py-2 px-3">Operation</th><th className="text-left py-2 px-3">Resource</th><th className="text-left py-2 px-3">IP</th><th className="text-left py-2 px-3">Time</th></tr></thead>
            <tbody>{audits.map((e)=>(
              <tr key={e.id} className="border-b hover:bg-surface-2/50"><td className="py-2 px-3"><Badge>{e.operation}</Badge></td><td className="py-2 px-3">{e.resource||"—"}</td><td className="py-2 px-3 font-mono">{e.ip_address}</td><td className="py-2 px-3">{new Date(e.created_at).toLocaleString()}</td></tr>
            ))}</tbody></table>
          ) : <p className="text-fg-muted text-sm">No entries.</p>}
        </GlassCard>
      )}

      {tab === "whitelist" && (
        <GlassCard className="p-4 space-y-4">
          <div className="flex items-center justify-between"><h2 className="text-lg font-semibold">IP Whitelist</h2><Button size="sm" onClick={()=>setShowWL(true)}><Plus className="w-4 h-4 mr-1"/>Add</Button></div>
          {whitelist && whitelist.length > 0 ? whitelist.map((e)=>(
            <div key={e.id} className="flex items-center justify-between border border-border rounded-xl p-3"><div><span className="font-mono">{e.cidr}</span>{e.description&&<span className="ml-2 text-sm text-fg-muted">{e.description}</span>}</div><Button variant="ghost" size="sm" onClick={()=>delWL.mutate(e.id)}><Trash2 className="w-3 h-3"/></Button></div>
          )) : <div className="flex items-center gap-2 text-fg-muted text-sm"><AlertTriangle className="w-4 h-4"/>No whitelist. All IPs allowed.</div>}
        </GlassCard>
      )}

      {tab === "bans" && (
        <GlassCard className="p-4 space-y-4">
          <h2 className="text-lg font-semibold">IP Bans</h2>
          {bans && bans.length > 0 ? bans.map((b)=>(
            <div key={b.id} className="flex items-center justify-between border border-border rounded-xl p-3"><div><span className="font-mono">{b.ip_address}</span><span className="ml-2 text-sm text-fg-muted">{b.reason}</span>{b.expires_at&&<span className="ml-2 text-xs text-fg-muted">Exp: {new Date(b.expires_at).toLocaleString()}</span>}</div><Button variant="ghost" size="sm" onClick={()=>delBan.mutate(b.id)}><Trash2 className="w-3 h-3"/></Button></div>
          )) : <p className="text-fg-muted text-sm">No bans.</p>}
        </GlassCard>
      )}

      {tab === "tokens" && (
        <GlassCard className="p-4 space-y-4">
          <h2 className="text-lg font-semibold">Scoped API Tokens</h2>
          <p className="text-fg-muted text-sm">Create tokens with restricted scopes for integrations.</p>
        </GlassCard>
      )}

      <Modal open={showWL} onClose={()=>setShowWL(false)} title="Add Whitelist Entry"><AddWLForm onDone={()=>{setShowWL(false);qc.invalidateQueries({queryKey:["sec-wl"]});toast.success("Added")}}/></Modal>
    </div>
  );
}

function AddWLForm({ onDone }: { onDone: () => void }) {
  const [cidr, setCidr] = useState("");
  const [desc, setDesc] = useState("");
  const mut = useMutation({ mutationFn: () => api("/api/v2/security/ip-whitelist", { method: "POST", body: { cidr, description: desc } }), onSuccess: onDone });
  return (
    <div className="space-y-4">
      <Input value={cidr} onChange={(e)=>setCidr(e.target.value)} placeholder="192.168.1.0/24"/>
      <Input value={desc} onChange={(e)=>setDesc(e.target.value)} placeholder="Office"/>
      <Button onClick={()=>mut.mutate()} disabled={!cidr||mut.isPending}><Plus className="w-4 h-4 mr-1"/>Add</Button>
    </div>
  );
}
