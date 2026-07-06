import { useEffect, useMemo, useState } from "react";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { Eye, Globe, Shield } from "lucide-react";
import { api } from "@/api/client";
import { Button, Input, Select, Switch } from "@/components/ui";
import { GlassCard, StatusBadge } from "@/components/veltrix";
import { useToast } from "@/components/toast";
import { useI18n } from "@/i18n/i18n";
import { useAuth } from "@/auth/auth";

interface ProbingPolicy {
  enabled: boolean;
  action: string;
  block_duration: number;
  max_probe_per_min: number;
  whitelisted_ips: string[];
  notify_telegram: boolean;
}

interface ProbeEvent {
  id: string;
  source_ip: string;
  port: number;
  method: string;
  action: string;
  created_at: string;
}

interface DecoySite {
  id: string;
  node_id: string | null;
  mode: string;
  target_url: string;
  static_html: string;
  enabled: boolean;
}

export function DecoyProbingTab() {
  const { t } = useI18n();
  const toast = useToast();
  const qc = useQueryClient();
  const { can, sudo } = useAuth();
  const canProbing = sudo || can("admin:manage");
  const canDecoy = can("node:write");

  const [policyForm, setPolicyForm] = useState<ProbingPolicy | null>(null);
  const [logOpen, setLogOpen] = useState(false);

  const { data: policyData } = useQuery({
    queryKey: ["probing-policy"],
    queryFn: () => api<{ policy: ProbingPolicy }>("/api/probing/policy"),
    enabled: canProbing,
  });
  const { data: eventsData } = useQuery({
    queryKey: ["probing-events"],
    queryFn: () => api<{ events: ProbeEvent[] }>("/api/probing/events"),
    enabled: canProbing,
  });
  const { data: decoyData } = useQuery({
    queryKey: ["decoys"],
    queryFn: () => api<{ decoys: DecoySite[] }>("/api/decoys"),
    enabled: canDecoy,
  });

  const policy = policyForm ?? policyData?.policy;
  const globalDecoy = decoyData?.decoys?.find((d) => !d.node_id) ?? decoyData?.decoys?.[0];
  const [decoyUrl, setDecoyUrl] = useState("");
  const [decoyEnabled, setDecoyEnabled] = useState(true);

  useEffect(() => {
    if (globalDecoy) {
      setDecoyUrl(globalDecoy.target_url || "");
      setDecoyEnabled(globalDecoy.enabled);
    }
  }, [globalDecoy?.id, globalDecoy?.target_url, globalDecoy?.enabled]);

  const blocked24h = useMemo(() => {
    const since = Date.now() - 24 * 60 * 60 * 1000;
    return (eventsData?.events ?? []).filter(
      (e) => e.action === "block" && new Date(e.created_at).getTime() >= since,
    ).length;
  }, [eventsData]);

  const savePolicy = useMutation({
    mutationFn: (p: ProbingPolicy) => api("/api/probing/policy", { method: "PUT", body: p }),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["probing-policy"] });
      toast.success(t("common.save"));
    },
  });

  const saveDecoy = useMutation({
    mutationFn: async () => {
      if (globalDecoy) {
        return api(`/api/decoys/${globalDecoy.id}`, {
          method: "PUT",
          body: {
            mode: globalDecoy.mode,
            target_url: decoyUrl,
            static_html: globalDecoy.static_html,
            enabled: decoyEnabled,
          },
        });
      }
      return api("/api/decoys", {
        method: "POST",
        body: { mode: "proxy", target_url: decoyUrl, enabled: decoyEnabled, static_html: "" },
      });
    },
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["decoys"] });
      toast.success(t("common.save"));
    },
  });

  function updatePolicy(field: keyof ProbingPolicy, value: unknown) {
    setPolicyForm((prev) => ({ ...(prev ?? policyData?.policy ?? {} as ProbingPolicy), [field]: value }));
  }

  const actionLabel =
    policy?.action === "block"
      ? t("security.decoy.actionBlock").replace("{hours}", String(Math.max(1, Math.round((policy.block_duration || 7200) / 3600))))
      : policy?.action === "honeypot"
        ? t("security.decoy.actionHoneypot")
        : t("security.decoy.actionLog");

  return (
    <div className="grid grid-cols-1 gap-4 lg:grid-cols-2">
      {canProbing && policy && (
        <GlassCard hover={false} className="!p-5 space-y-4">
          <div className="flex items-start justify-between gap-3">
            <div className="flex items-start gap-3">
              <div className="h-10 w-10 rounded-xl bg-danger/10 text-danger flex items-center justify-center flex-shrink-0">
                <Shield size={18} />
              </div>
              <div>
                <h3 className="text-sm font-bold text-fg">{t("security.decoy.probingTitle")}</h3>
                <p className="text-[11px] text-fg-subtle mt-0.5">{t("security.decoy.probingSubtitle")}</p>
              </div>
            </div>
            <StatusBadge status={policy.enabled ? "active" : "inactive"} label={policy.enabled ? "ENABLED" : "OFF"} />
          </div>

          <p className="text-xs text-fg-muted leading-relaxed">{t("security.decoy.probingDesc")}</p>

          <div className="space-y-2 text-xs border-t border-border/40 pt-3">
            <div className="flex justify-between gap-2">
              <span className="text-fg-subtle">{t("security.decoy.blocked24h")}</span>
              <span className="font-bold text-danger tabular-nums">{blocked24h.toLocaleString()} IPs</span>
            </div>
            <div className="flex justify-between gap-2">
              <span className="text-fg-subtle">{t("probing.action")}</span>
              <span className="text-fg text-end">{actionLabel}</span>
            </div>
          </div>

          <label className="flex items-center gap-2 text-xs">
            <input
              type="checkbox"
              checked={policy.enabled}
              onChange={(e) => updatePolicy("enabled", e.target.checked)}
              className="rounded"
            />
            {t("security.decoy.enableProbing")}
          </label>

          <div className="grid grid-cols-2 gap-2">
            <Select value={policy.action} onChange={(e) => updatePolicy("action", e.target.value)}>
              <option value="block">Block</option>
              <option value="honeypot">Honeypot</option>
              <option value="log">Log only</option>
            </Select>
            <Input
              value={policy.block_duration}
              onChange={(e) => updatePolicy("block_duration", Number(e.target.value))}
              inputMode="numeric"
              placeholder={t("probing.blockDuration")}
            />
          </div>

          <div className="flex flex-col gap-2">
            <Button variant="outline" size="sm" onClick={() => setLogOpen((v) => !v)}>
              <Eye size={14} /> {t("security.decoy.viewLog")}
            </Button>
            <Button size="sm" onClick={() => policy && savePolicy.mutate(policy)} disabled={savePolicy.isPending}>
              {t("common.save")}
            </Button>
          </div>

          {logOpen && (
            <div className="max-h-48 overflow-y-auto space-y-1.5 rounded-lg bg-surface-2/40 p-2">
              {(eventsData?.events ?? []).slice(0, 30).map((e) => (
                <div key={e.id} className="flex justify-between text-[11px] font-mono px-1">
                  <span>{e.source_ip}:{e.port}</span>
                  <span className="text-fg-muted">{e.action}</span>
                </div>
              ))}
              {(eventsData?.events?.length ?? 0) === 0 && (
                <p className="text-xs text-fg-muted text-center py-2">{t("probing.noProbes")}</p>
              )}
            </div>
          )}
        </GlassCard>
      )}

      {!canProbing && (
        <GlassCard hover={false} className="!p-5">
          <p className="text-sm text-fg-muted">{t("security.decoy.noProbingPerm")}</p>
        </GlassCard>
      )}

      {canDecoy && (
        <GlassCard hover={false} className="!p-5 space-y-4">
          <div className="flex items-start justify-between gap-3">
            <div className="flex items-start gap-3">
              <div className="h-10 w-10 rounded-xl bg-success/10 text-success flex items-center justify-center flex-shrink-0">
                <Globe size={18} />
              </div>
              <div>
                <h3 className="text-sm font-bold text-fg">{t("security.decoy.decoyTitle")}</h3>
                <p className="text-[11px] text-fg-subtle mt-0.5">{t("security.decoy.decoySubtitle")}</p>
              </div>
            </div>
            <Switch checked={decoyEnabled} onCheckedChange={setDecoyEnabled} />
          </div>

          <p className="text-xs text-fg-muted leading-relaxed">{t("security.decoy.decoyDesc")}</p>

          <div>
            <label className="text-xs text-fg-subtle">{t("security.decoy.targetUrl")}</label>
            <Input
              className="mt-1 font-mono text-xs"
              value={decoyUrl}
              onChange={(e) => setDecoyUrl(e.target.value)}
              placeholder="https://www.example.com"
              dir="ltr"
            />
          </div>

          <div className="flex flex-col gap-2">
            <Button
              variant="outline"
              size="sm"
              onClick={() => decoyUrl && window.open(decoyUrl, "_blank", "noopener,noreferrer")}
              disabled={!decoyUrl}
            >
              <Eye size={14} /> {t("security.decoy.preview")}
            </Button>
            <Button size="sm" onClick={() => saveDecoy.mutate()} disabled={saveDecoy.isPending || !decoyUrl}>
              {t("common.save")}
            </Button>
          </div>
        </GlassCard>
      )}

      {!canDecoy && (
        <GlassCard hover={false} className="!p-5">
          <p className="text-sm text-fg-muted">{t("security.decoy.noDecoyPerm")}</p>
        </GlassCard>
      )}
    </div>
  );
}
