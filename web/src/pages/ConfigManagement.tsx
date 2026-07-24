import { useState } from "react";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import {
  RefreshCw,
  Download,
  Upload,
  RotateCcw,
  CheckCircle,
  XCircle,
  Globe,
  GitCompare,
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
interface ConfigVersion {
  id: string;
  inbound_id: string;
  version: number;
  config_data: Record<string, unknown>;
  comment: string;
  admin_id?: string;
  created_at: string;
}

interface ConfigValidationError {
  field: string;
  message: string;
}

interface ConfigDiff {
  inbound_id: string;
  old_version: number;
  new_version: number;
  old_config: Record<string, unknown>;
  new_config: Record<string, unknown>;
  changes: ConfigChange[];
}

interface ConfigChange {
  path: string;
  old_value: unknown;
  new_value: unknown;
  type: "added" | "removed" | "modified";
}

interface ConnectedIP {
  ip: string;
  country: string;
  user_id?: string;
  username?: string;
}

const protocols = [
  "vmess", "vless", "trojan", "shadowsocks", "hysteria2",
  "tuic", "wireguard", "shadowtls", "naive",
];

const networks = ["tcp", "ws", "grpc", "httpupgrade", "xhttp"];
const securities = ["none", "tls", "reality"];

export function ConfigManagement() {
  const { t } = useI18n();
  useTitle(t("configManagement.title", "Config Management"));
  const queryClient = useQueryClient();
  const toast = useToast();
  const confirm = useConfirm();

  const [selectedInbound, setSelectedInbound] = useState<string>("");
  const [showDiff, setShowDiff] = useState(false);
  const [diffData, setDiffData] = useState<ConfigDiff | null>(null);
  const [showImport, setShowImport] = useState(false);
  const [showConnectedIPs, setShowConnectedIPs] = useState(false);

  // Validation form state
  const [validateForm, setValidateForm] = useState({
    protocol: "vless",
    network: "tcp",
    security: "reality",
    config: "{}",
  });

  // Fetch versions for selected inbound
  const { data: versions } = useQuery({
    queryKey: ["config-versions", selectedInbound],
    queryFn: () =>
      api.get(`/api/v2/inbounds/${selectedInbound}/versions`).then((r) => r.data as ConfigVersion[]),
    enabled: !!selectedInbound,
  });

  // Fetch connected IPs
  const { data: connectedIPs } = useQuery({
    queryKey: ["connected-ips", selectedInbound],
    queryFn: () =>
      api.get(`/api/v2/inbounds/${selectedInbound}/connected-ips`).then((r) => r.data as ConnectedIP[]),
    enabled: !!selectedInbound && showConnectedIPs,
  });

  // Validate config
  const validateMutation = useMutation({
    mutationFn: (data: { protocol: string; network: string; security: string; config: Record<string, unknown> }) =>
      api.post(`/api/v2/inbounds/${selectedInbound}/validate`, data),
    onSuccess: (res) => {
      if (res.data.valid) {
        toast.success(t("configManagement.validConfig", "Configuration is valid"));
      }
    },
  });

  // Get defaults
  const defaultsMutation = useMutation({
    mutationFn: (params: { protocol: string; network: string; security: string }) =>
      api.get(`/api/v2/inbounds/${selectedInbound}/defaults`, { params }),
    onSuccess: (res) => {
      setValidateForm((f) => ({ ...f, config: JSON.stringify(res.data.config, null, 2) }));
      toast.success(t("configManagement.defaultsLoaded", "Defaults loaded"));
    },
  });

  // Rollback
  const rollbackMutation = useMutation({
    mutationFn: (version: number) =>
      api.post(`/api/v2/inbounds/${selectedInbound}/rollback`, { version }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["config-versions", selectedInbound] });
      toast.success(t("configManagement.rollbackSuccess", "Rollback successful"));
    },
  });

  // Export
  const handleExport = async () => {
    try {
      const res = await api.get(`/api/v2/inbounds/${selectedInbound}/export`, {
        responseType: "blob",
      });
      const url = URL.createObjectURL(res.data);
      const a = document.createElement("a");
      a.href = url;
      a.download = `config-${selectedInbound}.json`;
      a.click();
      URL.revokeObjectURL(url);
    } catch {
      toast.error(t("configManagement.exportFailed", "Export failed"));
    }
  };

  // Diff
  const handleDiff = async (oldVersion: number, newVersion: number) => {
    try {
      const res = await api.get(`/api/v2/inbounds/${selectedInbound}/diff`, {
        params: { old_version: oldVersion, new_version: newVersion },
      });
      setDiffData(res.data);
      setShowDiff(true);
    } catch {
      toast.error(t("configManagement.diffFailed", "Could not compute diff"));
    }
  };

  // Validate handler
  const handleValidate = () => {
    try {
      const config = JSON.parse(validateForm.config);
      validateMutation.mutate({
        protocol: validateForm.protocol,
        network: validateForm.network,
        security: validateForm.security,
        config,
      });
    } catch {
      toast.error(t("configManagement.invalidJSON", "Invalid JSON in config"));
    }
  };

  return (
    <div className="space-y-6 p-6">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-bold">
          {t("configManagement.title", "Config Management")}
        </h1>
      </div>

      {/* Inbound Selector */}
      <GlassCard className="p-4">
        <label className="block text-sm font-medium mb-2">
          {t("configManagement.selectInbound", "Select Inbound")}
        </label>
        <Input
          placeholder="Inbound UUID"
          value={selectedInbound}
          onChange={(e) => setSelectedInbound(e.target.value)}
        />
      </GlassCard>

      {selectedInbound && (
        <>
          {/* Validation Panel */}
          <GlassCard className="p-4 space-y-4">
            <h2 className="text-lg font-semibold flex items-center gap-2">
              <CheckCircle className="w-5 h-5" />
              {t("configManagement.validate", "Validate Configuration")}
            </h2>
            <div className="grid grid-cols-3 gap-4">
              <div>
                <label className="block text-xs mb-1">Protocol</label>
                <select
                  className="w-full rounded border p-2 bg-background"
                  value={validateForm.protocol}
                  onChange={(e) => setValidateForm((f) => ({ ...f, protocol: e.target.value }))}
                >
                  {protocols.map((p) => (
                    <option key={p} value={p}>{p}</option>
                  ))}
                </select>
              </div>
              <div>
                <label className="block text-xs mb-1">Network</label>
                <select
                  className="w-full rounded border p-2 bg-background"
                  value={validateForm.network}
                  onChange={(e) => setValidateForm((f) => ({ ...f, network: e.target.value }))}
                >
                  {networks.map((n) => (
                    <option key={n} value={n}>{n}</option>
                  ))}
                </select>
              </div>
              <div>
                <label className="block text-xs mb-1">Security</label>
                <select
                  className="w-full rounded border p-2 bg-background"
                  value={validateForm.security}
                  onChange={(e) => setValidateForm((f) => ({ ...f, security: e.target.value }))}
                >
                  {securities.map((s) => (
                    <option key={s} value={s}>{s}</option>
                  ))}
                </select>
              </div>
            </div>
            <div>
              <label className="block text-xs mb-1">Config JSON</label>
              <textarea
                className="w-full h-40 rounded border p-2 font-mono text-sm bg-background"
                value={validateForm.config}
                onChange={(e) => setValidateForm((f) => ({ ...f, config: e.target.value }))}
              />
            </div>
            <div className="flex gap-2">
              <Button onClick={handleValidate} disabled={validateMutation.isPending}>
                <CheckCircle className="w-4 h-4 mr-1" />
                {t("configManagement.validate", "Validate")}
              </Button>
              <Button
                variant="outline"
                onClick={() =>
                  defaultsMutation.mutate({
                    protocol: validateForm.protocol,
                    network: validateForm.network,
                    security: validateForm.security,
                  })
                }
              >
                <RefreshCw className="w-4 h-4 mr-1" />
                {t("configManagement.loadDefaults", "Load Defaults")}
              </Button>
            </div>
            {validateMutation.data && !validateMutation.data.data.valid && (
              <div className="mt-2 p-3 bg-destructive/10 rounded">
                <p className="font-medium text-destructive flex items-center gap-1">
                  <XCircle className="w-4 h-4" />
                  {t("configManagement.validationErrors", "Validation Errors")}
                </p>
                <ul className="mt-1 list-disc list-inside text-sm">
                  {(validateMutation.data.data.errors as ConfigValidationError[]).map((e, i) => (
                    <li key={i}>
                      <code>{e.field}</code>: {e.message}
                    </li>
                  ))}
                </ul>
              </div>
            )}
          </GlassCard>

          {/* Version History */}
          <GlassCard className="p-4 space-y-4">
            <div className="flex items-center justify-between">
              <h2 className="text-lg font-semibold">
                {t("configManagement.versions", "Version History")}
              </h2>
              <div className="flex gap-2">
                <Button variant="outline" size="sm" onClick={handleExport}>
                  <Download className="w-4 h-4 mr-1" />
                  {t("configManagement.export", "Export")}
                </Button>
                <Button variant="outline" size="sm" onClick={() => setShowImport(true)}>
                  <Upload className="w-4 h-4 mr-1" />
                  {t("configManagement.import", "Import")}
                </Button>
                <Button variant="outline" size="sm" onClick={() => setShowConnectedIPs(true)}>
                  <Globe className="w-4 h-4 mr-1" />
                  {t("configManagement.connectedIPs", "Connected IPs")}
                </Button>
              </div>
            </div>

            {versions && versions.length > 0 ? (
              <div className="overflow-x-auto">
                <table className="w-full text-sm">
                  <thead>
                    <tr className="border-b">
                      <th className="text-left py-2 px-3">Version</th>
                      <th className="text-left py-2 px-3">Comment</th>
                      <th className="text-left py-2 px-3">Created</th>
                      <th className="text-right py-2 px-3">Actions</th>
                    </tr>
                  </thead>
                  <tbody>
                    {versions.map((v, i) => (
                      <tr key={v.id} className="border-b hover:bg-muted/50">
                        <td className="py-2 px-3">
                          <Badge variant="secondary">v{v.version}</Badge>
                        </td>
                        <td className="py-2 px-3">{v.comment || "—"}</td>
                        <td className="py-2 px-3">
                          {new Date(v.created_at).toLocaleString()}
                        </td>
                        <td className="py-2 px-3 text-right space-x-1">
                          {i < versions.length - 1 && (
                            <Button
                              variant="ghost"
                              size="sm"
                              onClick={() => handleDiff(versions[i + 1].version, v.version)}
                            >
                              <GitCompare className="w-3 h-3" />
                            </Button>
                          )}
                          <Button
                            variant="ghost"
                            size="sm"
                            onClick={async () => {
                              const ok = await confirm(
                                t("configManagement.confirmRollback", "Rollback to version {{v}}?").replace("{{v}}", String(v.version))
                              );
                              if (ok) rollbackMutation.mutate(v.version);
                            }}
                          >
                            <RotateCcw className="w-3 h-3" />
                          </Button>
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            ) : (
              <p className="text-muted-foreground text-sm">
                {t("configManagement.noVersions", "No configuration versions yet.")}
              </p>
            )}
          </GlassCard>

          {/* Connected IPs Panel */}
          {showConnectedIPs && (
            <GlassCard className="p-4 space-y-4">
              <div className="flex items-center justify-between">
                <h2 className="text-lg font-semibold flex items-center gap-2">
                  <Globe className="w-5 h-5" />
                  {t("configManagement.connectedIPs", "Connected IPs")}
                </h2>
                <Button variant="ghost" size="sm" onClick={() => setShowConnectedIPs(false)}>
                  <XCircle className="w-4 h-4" />
                </Button>
              </div>
              {connectedIPs && connectedIPs.length > 0 ? (
                <div className="overflow-x-auto">
                  <table className="w-full text-sm">
                    <thead>
                      <tr className="border-b">
                        <th className="text-left py-2 px-3">IP</th>
                        <th className="text-left py-2 px-3">Country</th>
                        <th className="text-left py-2 px-3">User</th>
                      </tr>
                    </thead>
                    <tbody>
                      {connectedIPs.map((ip, i) => (
                        <tr key={i} className="border-b hover:bg-muted/50">
                          <td className="py-2 px-3 font-mono">{ip.ip}</td>
                          <td className="py-2 px-3">{ip.country || "—"}</td>
                          <td className="py-2 px-3">{ip.username || ip.user_id || "—"}</td>
                        </tr>
                      ))}
                    </tbody>
                  </table>
                </div>
              ) : (
                <p className="text-muted-foreground text-sm">
                  {t("configManagement.noConnections", "No active connections.")}
                </p>
              )}
            </GlassCard>
          )}
        </>
      )}

      {/* Diff Modal */}
      <Modal open={showDiff} onClose={() => setShowDiff(false)} title="Config Diff">
        {diffData && (
          <div className="space-y-3 max-h-96 overflow-auto">
            <p className="text-sm text-muted-foreground">
              v{diffData.old_version} → v{diffData.new_version}
            </p>
            {diffData.changes.length === 0 ? (
              <p className="text-sm">{t("configManagement.noChanges", "No changes")}</p>
            ) : (
              <div className="space-y-2">
                {diffData.changes.map((change, i) => (
                  <div
                    key={i}
                    className={`p-2 rounded text-sm font-mono ${
                      change.type === "added"
                        ? "bg-green-500/10 text-green-700 dark:text-green-400"
                        : change.type === "removed"
                        ? "bg-red-500/10 text-red-700 dark:text-red-400"
                        : "bg-yellow-500/10 text-yellow-700 dark:text-yellow-400"
                    }`}
                  >
                    <span className="font-semibold">{change.path}</span>
                    {change.type === "modified" && (
                      <>
                        <div className="text-red-600 dark:text-red-400">
                          - {JSON.stringify(change.old_value)}
                        </div>
                        <div className="text-green-600 dark:text-green-400">
                          + {JSON.stringify(change.new_value)}
                        </div>
                      </>
                    )}
                    {change.type === "added" && (
                      <div>+ {JSON.stringify(change.new_value)}</div>
                    )}
                    {change.type === "removed" && (
                      <div>- {JSON.stringify(change.old_value)}</div>
                    )}
                  </div>
                ))}
              </div>
            )}
          </div>
        )}
      </Modal>

      {/* Import Modal */}
      <Modal open={showImport} onClose={() => setShowImport(false)} title="Import Configuration">
        <ImportForm
          inboundId={selectedInbound}
          onSuccess={() => {
            setShowImport(false);
            queryClient.invalidateQueries({ queryKey: ["config-versions", selectedInbound] });
            toast.success(t("configManagement.importSuccess", "Configuration imported"));
          }}
          onError={(msg) => toast.error(msg)}
        />
      </Modal>
    </div>
  );
}

// --- Import Form sub-component ---

function ImportForm({
  inboundId,
  onSuccess,
  onError,
}: {
  inboundId: string;
  onSuccess: () => void;
  onError: (msg: string) => void;
}) {
  const { t } = useI18n();
  const [file, setFile] = useState<File | null>(null);
  const [protocol, setProtocol] = useState("vless");
  const [network, setNetwork] = useState("tcp");
  const [security, setSecurity] = useState("reality");

  const importMutation = useMutation({
    mutationFn: async () => {
      if (!file) throw new Error("No file selected");
      const text = await file.text();
      return api.post("/api/v2/inbounds/import", {
        inbound_id: inboundId,
        protocol,
        network,
        security,
        config: JSON.parse(text).config || JSON.parse(text),
      });
    },
    onSuccess: () => onSuccess(),
    onError: (err: Error) => onError(err.message),
  });

  return (
    <div className="space-y-4">
      <div className="grid grid-cols-3 gap-3">
        <div>
          <label className="block text-xs mb-1">Protocol</label>
          <select
            className="w-full rounded border p-2 bg-background"
            value={protocol}
            onChange={(e) => setProtocol(e.target.value)}
          >
            {protocols.map((p) => (
              <option key={p} value={p}>{p}</option>
            ))}
          </select>
        </div>
        <div>
          <label className="block text-xs mb-1">Network</label>
          <select
            className="w-full rounded border p-2 bg-background"
            value={network}
            onChange={(e) => setNetwork(e.target.value)}
          >
            {networks.map((n) => (
              <option key={n} value={n}>{n}</option>
            ))}
          </select>
        </div>
        <div>
          <label className="block text-xs mb-1">Security</label>
          <select
            className="w-full rounded border p-2 bg-background"
            value={security}
            onChange={(e) => setSecurity(e.target.value)}
          >
            {securities.map((s) => (
              <option key={s} value={s}>{s}</option>
            ))}
          </select>
        </div>
      </div>
      <div>
        <label className="block text-xs mb-1">
          {t("configManagement.selectFile", "Select JSON file")}
        </label>
        <input
          type="file"
          accept=".json"
          onChange={(e) => setFile(e.target.files?.[0] || null)}
          className="text-sm"
        />
      </div>
      <Button onClick={() => importMutation.mutate()} disabled={!file || importMutation.isPending}>
        <Upload className="w-4 h-4 mr-1" />
        {t("configManagement.import", "Import")}
      </Button>
    </div>
  );
}
