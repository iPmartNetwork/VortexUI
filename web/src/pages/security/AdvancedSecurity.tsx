import { useState } from "react";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import {
  Shield,
  Key,
  Globe,
  Clock,
  AlertTriangle,
  Trash2,
  Plus,
  Ban,
  Activity,
} from "lucide-react";
import { api } from "@/api/client";
import { Button, Input, Badge } from "@/components/ui";
import { GlassCard } from "@/components/veltrix";
import { Modal } from "@/components/Modal";
import { useToast } from "@/components/toast";
import { useConfirm } from "@/components/confirm";
import { useTitle } from "@/lib/useTitle";
import { useI18n } from "@/i18n/i18n";

// Types
interface AdminSession {
  id: string;
  admin_id: string;
  ip_address: string;
  user_agent: string;
  country: string;
  last_active: string;
  created_at: string;
  revoked: boolean;
}

interface LoginAuditEntry {
  id: string;
  admin_id?: string;
  username: string;
  ip_address: string;
  user_agent: string;
  country: string;
  success: boolean;
  failure_reason: string;
  created_at: string;
}

interface SecurityAuditEntry {
  id: string;
  admin_id?: string;
  operation: string;
  resource: string;
  before_state?: Record<string, unknown>;
  after_state?: Record<string, unknown>;
  ip_address: string;
  created_at: string;
}

interface IPWhitelistEntry {
  id: string;
  admin_id?: string;
  cidr: string;
  description: string;
  created_at: string;
}

interface IPBan {
  id: string;
  ip_address: string;
  reason: string;
  expires_at?: string;
  created_at: string;
}

type Tab = "sessions" | "login-audit" | "security-audit" | "whitelist" | "bans" | "tokens";

export function AdvancedSecurity() {
  const { t } = useI18n();
  useTitle(t("security.title", "Advanced Security"));
  const queryClient = useQueryClient();
  const toast = useToast();
  const confirm = useConfirm();

  const [activeTab, setActiveTab] = useState<Tab>("sessions");
  const [showAddWhitelist, setShowAddWhitelist] = useState(false);
  const [showCreateToken, setShowCreateToken] = useState(false);

  // Queries
  const { data: sessions } = useQuery({
    queryKey: ["security-sessions"],
    queryFn: () => api.get("/api/v2/security/sessions").then((r) => r.data as AdminSession[]),
    enabled: activeTab === "sessions",
  });

  const { data: loginAudit } = useQuery({
    queryKey: ["security-login-audit"],
    queryFn: () => api.get("/api/v2/security/login-audit").then((r) => r.data as LoginAuditEntry[]),
    enabled: activeTab === "login-audit",
  });

  const { data: securityAudit } = useQuery({
    queryKey: ["security-audit-log"],
    queryFn: () => api.get("/api/v2/security/audit-log").then((r) => r.data as SecurityAuditEntry[]),
    enabled: activeTab === "security-audit",
  });

  const { data: whitelist } = useQuery({
    queryKey: ["security-whitelist"],
    queryFn: () => api.get("/api/v2/security/ip-whitelist").then((r) => r.data as IPWhitelistEntry[]),
    enabled: activeTab === "whitelist",
  });

  const { data: bans } = useQuery({
    queryKey: ["security-bans"],
    queryFn: () => api.get("/api/v2/security/bans").then((r) => r.data as IPBan[]),
    enabled: activeTab === "bans",
  });

  // Mutations
  const revokeSessionMutation = useMutation({
    mutationFn: (id: string) => api.delete(`/api/v2/security/sessions/${id}`),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["security-sessions"] });
      toast.success(t("security.sessionRevoked", "Session revoked"));
    },
  });

  const revokeAllSessionsMutation = useMutation({
    mutationFn: () => api.delete("/api/v2/security/sessions"),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["security-sessions"] });
      toast.success(t("security.allSessionsRevoked", "All sessions revoked"));
    },
  });

  const removeWhitelistMutation = useMutation({
    mutationFn: (id: string) => api.delete(`/api/v2/security/ip-whitelist/${id}`),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["security-whitelist"] });
      toast.success(t("security.whitelistRemoved", "Whitelist entry removed"));
    },
  });

  const removeBanMutation = useMutation({
    mutationFn: (id: string) => api.delete(`/api/v2/security/bans/${id}`),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["security-bans"] });
      toast.success(t("security.banRemoved", "Ban removed"));
    },
  });

  const tabs: { key: Tab; label: string; icon: React.ReactNode }[] = [
    { key: "sessions", label: t("security.sessions", "Sessions"), icon: <Activity className="w-4 h-4" /> },
    { key: "login-audit", label: t("security.loginAudit", "Login Audit"), icon: <Clock className="w-4 h-4" /> },
    { key: "security-audit", label: t("security.auditLog", "Audit Log"), icon: <Shield className="w-4 h-4" /> },
    { key: "whitelist", label: t("security.whitelist", "IP Whitelist"), icon: <Globe className="w-4 h-4" /> },
    { key: "bans", label: t("security.bans", "IP Bans"), icon: <Ban className="w-4 h-4" /> },
    { key: "tokens", label: t("security.tokens", "API Tokens"), icon: <Key className="w-4 h-4" /> },
  ];

  return (
    <div className="space-y-6 p-6">
      <h1 className="text-2xl font-bold flex items-center gap-2">
        <Shield className="w-6 h-6" />
        {t("security.title", "Advanced Security")}
      </h1>

      {/* Tab navigation */}
      <div className="flex gap-2 border-b pb-2 overflow-x-auto">
        {tabs.map((tab) => (
          <Button
            key={tab.key}
            variant={activeTab === tab.key ? "default" : "ghost"}
            size="sm"
            onClick={() => setActiveTab(tab.key)}
            className="flex items-center gap-1"
          >
            {tab.icon}
            {tab.label}
          </Button>
        ))}
      </div>

      {/* Sessions */}
      {activeTab === "sessions" && (
        <GlassCard className="p-4 space-y-4">
          <div className="flex items-center justify-between">
            <h2 className="text-lg font-semibold">{t("security.activeSessions", "Active Sessions")}</h2>
            <Button
              variant="destructive"
              size="sm"
              onClick={async () => {
                const ok = await confirm(t("security.revokeAllConfirm", "Revoke all sessions?"));
                if (ok) revokeAllSessionsMutation.mutate();
              }}
            >
              {t("security.revokeAll", "Revoke All")}
            </Button>
          </div>
          {sessions && sessions.length > 0 ? (
            <div className="overflow-x-auto">
              <table className="w-full text-sm">
                <thead>
                  <tr className="border-b">
                    <th className="text-left py-2 px-3">IP</th>
                    <th className="text-left py-2 px-3">Country</th>
                    <th className="text-left py-2 px-3">User Agent</th>
                    <th className="text-left py-2 px-3">Last Active</th>
                    <th className="text-left py-2 px-3">Status</th>
                    <th className="text-right py-2 px-3">Actions</th>
                  </tr>
                </thead>
                <tbody>
                  {sessions.map((s) => (
                    <tr key={s.id} className="border-b hover:bg-muted/50">
                      <td className="py-2 px-3 font-mono">{s.ip_address}</td>
                      <td className="py-2 px-3">{s.country || "—"}</td>
                      <td className="py-2 px-3 truncate max-w-[200px]">{s.user_agent}</td>
                      <td className="py-2 px-3">{new Date(s.last_active).toLocaleString()}</td>
                      <td className="py-2 px-3">
                        <Badge variant={s.revoked ? "destructive" : "default"}>
                          {s.revoked ? "Revoked" : "Active"}
                        </Badge>
                      </td>
                      <td className="py-2 px-3 text-right">
                        {!s.revoked && (
                          <Button
                            variant="ghost"
                            size="sm"
                            onClick={() => revokeSessionMutation.mutate(s.id)}
                          >
                            <Trash2 className="w-3 h-3" />
                          </Button>
                        )}
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          ) : (
            <p className="text-muted-foreground text-sm">{t("security.noSessions", "No active sessions.")}</p>
          )}
        </GlassCard>
      )}

      {/* Login Audit */}
      {activeTab === "login-audit" && (
        <GlassCard className="p-4 space-y-4">
          <h2 className="text-lg font-semibold">{t("security.loginAudit", "Login Audit")}</h2>
          {loginAudit && loginAudit.length > 0 ? (
            <div className="overflow-x-auto">
              <table className="w-full text-sm">
                <thead>
                  <tr className="border-b">
                    <th className="text-left py-2 px-3">Username</th>
                    <th className="text-left py-2 px-3">IP</th>
                    <th className="text-left py-2 px-3">Country</th>
                    <th className="text-left py-2 px-3">Result</th>
                    <th className="text-left py-2 px-3">Time</th>
                  </tr>
                </thead>
                <tbody>
                  {loginAudit.map((entry) => (
                    <tr key={entry.id} className="border-b hover:bg-muted/50">
                      <td className="py-2 px-3">{entry.username}</td>
                      <td className="py-2 px-3 font-mono">{entry.ip_address}</td>
                      <td className="py-2 px-3">{entry.country || "—"}</td>
                      <td className="py-2 px-3">
                        <Badge variant={entry.success ? "default" : "destructive"}>
                          {entry.success ? "Success" : "Failed"}
                        </Badge>
                        {entry.failure_reason && (
                          <span className="ml-1 text-xs text-muted-foreground">{entry.failure_reason}</span>
                        )}
                      </td>
                      <td className="py-2 px-3">{new Date(entry.created_at).toLocaleString()}</td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          ) : (
            <p className="text-muted-foreground text-sm">{t("security.noLoginAudit", "No login audit entries.")}</p>
          )}
        </GlassCard>
      )}

      {/* Security Audit */}
      {activeTab === "security-audit" && (
        <GlassCard className="p-4 space-y-4">
          <h2 className="text-lg font-semibold">{t("security.auditLog", "Security Audit Log")}</h2>
          {securityAudit && securityAudit.length > 0 ? (
            <div className="overflow-x-auto">
              <table className="w-full text-sm">
                <thead>
                  <tr className="border-b">
                    <th className="text-left py-2 px-3">Operation</th>
                    <th className="text-left py-2 px-3">Resource</th>
                    <th className="text-left py-2 px-3">IP</th>
                    <th className="text-left py-2 px-3">Time</th>
                  </tr>
                </thead>
                <tbody>
                  {securityAudit.map((entry) => (
                    <tr key={entry.id} className="border-b hover:bg-muted/50">
                      <td className="py-2 px-3">
                        <Badge variant="outline">{entry.operation}</Badge>
                      </td>
                      <td className="py-2 px-3">{entry.resource || "—"}</td>
                      <td className="py-2 px-3 font-mono">{entry.ip_address}</td>
                      <td className="py-2 px-3">{new Date(entry.created_at).toLocaleString()}</td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          ) : (
            <p className="text-muted-foreground text-sm">{t("security.noAudit", "No audit entries.")}</p>
          )}
        </GlassCard>
      )}

      {/* IP Whitelist */}
      {activeTab === "whitelist" && (
        <GlassCard className="p-4 space-y-4">
          <div className="flex items-center justify-between">
            <h2 className="text-lg font-semibold">{t("security.whitelist", "IP Whitelist")}</h2>
            <Button size="sm" onClick={() => setShowAddWhitelist(true)}>
              <Plus className="w-4 h-4 mr-1" />
              {t("security.addWhitelist", "Add Entry")}
            </Button>
          </div>
          {whitelist && whitelist.length > 0 ? (
            <div className="space-y-2">
              {whitelist.map((entry) => (
                <div key={entry.id} className="flex items-center justify-between border rounded p-3">
                  <div>
                    <span className="font-mono">{entry.cidr}</span>
                    {entry.description && (
                      <span className="ml-2 text-sm text-muted-foreground">{entry.description}</span>
                    )}
                  </div>
                  <Button
                    variant="ghost"
                    size="sm"
                    onClick={() => removeWhitelistMutation.mutate(entry.id)}
                  >
                    <Trash2 className="w-3 h-3" />
                  </Button>
                </div>
              ))}
            </div>
          ) : (
            <div className="flex items-center gap-2 text-muted-foreground text-sm">
              <AlertTriangle className="w-4 h-4" />
              {t("security.noWhitelist", "No whitelist configured. All IPs are allowed.")}
            </div>
          )}
        </GlassCard>
      )}

      {/* IP Bans */}
      {activeTab === "bans" && (
        <GlassCard className="p-4 space-y-4">
          <h2 className="text-lg font-semibold">{t("security.bans", "IP Bans")}</h2>
          {bans && bans.length > 0 ? (
            <div className="space-y-2">
              {bans.map((ban) => (
                <div key={ban.id} className="flex items-center justify-between border rounded p-3">
                  <div>
                    <span className="font-mono">{ban.ip_address}</span>
                    <span className="ml-2 text-sm text-muted-foreground">{ban.reason}</span>
                    {ban.expires_at && (
                      <Badge variant="outline" className="ml-2">
                        Expires: {new Date(ban.expires_at).toLocaleString()}
                      </Badge>
                    )}
                  </div>
                  <Button
                    variant="ghost"
                    size="sm"
                    onClick={() => removeBanMutation.mutate(ban.id)}
                  >
                    <Trash2 className="w-3 h-3" />
                  </Button>
                </div>
              ))}
            </div>
          ) : (
            <p className="text-muted-foreground text-sm">{t("security.noBans", "No active IP bans.")}</p>
          )}
        </GlassCard>
      )}

      {/* Tokens */}
      {activeTab === "tokens" && (
        <GlassCard className="p-4 space-y-4">
          <div className="flex items-center justify-between">
            <h2 className="text-lg font-semibold">{t("security.scopedTokens", "Scoped API Tokens")}</h2>
            <Button size="sm" onClick={() => setShowCreateToken(true)}>
              <Plus className="w-4 h-4 mr-1" />
              {t("security.createToken", "Create Token")}
            </Button>
          </div>
          <p className="text-muted-foreground text-sm">
            {t("security.tokenDesc", "Create API tokens with restricted scopes for third-party integrations.")}
          </p>
        </GlassCard>
      )}

      {/* Add Whitelist Modal */}
      <Modal open={showAddWhitelist} onClose={() => setShowAddWhitelist(false)} title={t("security.addWhitelist", "Add Whitelist Entry")}>
        <AddWhitelistForm
          onSuccess={() => {
            setShowAddWhitelist(false);
            queryClient.invalidateQueries({ queryKey: ["security-whitelist"] });
            toast.success(t("security.whitelistAdded", "Whitelist entry added"));
          }}
          onError={(msg) => toast.error(msg)}
        />
      </Modal>

      {/* Create Token Modal */}
      <Modal open={showCreateToken} onClose={() => setShowCreateToken(false)} title={t("security.createToken", "Create Scoped Token")}>
        <CreateTokenForm
          onSuccess={() => {
            setShowCreateToken(false);
            toast.success(t("security.tokenCreated", "Token created"));
          }}
          onError={(msg) => toast.error(msg)}
        />
      </Modal>
    </div>
  );
}

// --- Sub-components ---

function AddWhitelistForm({ onSuccess, onError }: { onSuccess: () => void; onError: (msg: string) => void }) {
  const { t } = useI18n();
  const [cidr, setCidr] = useState("");
  const [description, setDescription] = useState("");

  const mutation = useMutation({
    mutationFn: () => api.post("/api/v2/security/ip-whitelist", { cidr, description }),
    onSuccess: () => onSuccess(),
    onError: (err: Error) => onError(err.message),
  });

  return (
    <div className="space-y-4">
      <div>
        <label className="block text-sm font-medium mb-1">CIDR / IP</label>
        <Input value={cidr} onChange={(e) => setCidr(e.target.value)} placeholder="192.168.1.0/24 or 1.2.3.4/32" />
      </div>
      <div>
        <label className="block text-sm font-medium mb-1">{t("security.description", "Description")}</label>
        <Input value={description} onChange={(e) => setDescription(e.target.value)} placeholder="Office network" />
      </div>
      <Button onClick={() => mutation.mutate()} disabled={!cidr || mutation.isPending}>
        <Plus className="w-4 h-4 mr-1" />
        {t("common.add", "Add")}
      </Button>
    </div>
  );
}

const availableScopes = [
  "users:read", "users:write",
  "nodes:read", "nodes:write",
  "inbounds:read", "inbounds:write",
  "admin:read", "admin:write",
  "settings:read", "settings:write",
  "subscription:read", "security:manage",
];

function CreateTokenForm({ onSuccess, onError }: { onSuccess: () => void; onError: (msg: string) => void }) {
  const { t } = useI18n();
  const [name, setName] = useState("");
  const [scopes, setScopes] = useState<string[]>([]);

  const mutation = useMutation({
    mutationFn: () => api.post("/api/v2/api-tokens", { name, scopes }),
    onSuccess: () => onSuccess(),
    onError: (err: Error) => onError(err.message),
  });

  const toggleScope = (scope: string) => {
    setScopes((prev) =>
      prev.includes(scope) ? prev.filter((s) => s !== scope) : [...prev, scope]
    );
  };

  return (
    <div className="space-y-4">
      <div>
        <label className="block text-sm font-medium mb-1">{t("security.tokenName", "Token Name")}</label>
        <Input value={name} onChange={(e) => setName(e.target.value)} placeholder="my-integration" />
      </div>
      <div>
        <label className="block text-sm font-medium mb-1">{t("security.scopes", "Scopes")}</label>
        <div className="grid grid-cols-2 gap-2 mt-2">
          {availableScopes.map((scope) => (
            <label key={scope} className="flex items-center gap-2 text-sm cursor-pointer">
              <input
                type="checkbox"
                checked={scopes.includes(scope)}
                onChange={() => toggleScope(scope)}
                className="rounded"
              />
              <code className="text-xs">{scope}</code>
            </label>
          ))}
        </div>
      </div>
      <Button onClick={() => mutation.mutate()} disabled={!name || scopes.length === 0 || mutation.isPending}>
        <Key className="w-4 h-4 mr-1" />
        {t("security.createToken", "Create Token")}
      </Button>
    </div>
  );
}
