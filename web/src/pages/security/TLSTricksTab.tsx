import { useMemo, useState } from "react";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { Plus, SlidersHorizontal } from "lucide-react";
import { api } from "@/api/client";
import { useInboundsFleet } from "@/api/hooks";
import { Badge, Button, Input, Select } from "@/components/ui";
import { Modal } from "@/components/Modal";
import { GlassCard, StatusBadge } from "@/components/veltrix";
import { useConfirm } from "@/components/confirm";
import { useToast } from "@/components/toast";
import { useI18n } from "@/i18n/i18n";
import type { TKey } from "@/i18n/dict";
import { useAuth } from "@/auth/auth";
import { cn } from "@/lib/utils";

interface TLSProfile {
  id: string;
  name: string;
  isp: string;
  description?: string;
  fragment_enabled: boolean;
  fragment_size: string;
  fragment_interval: string;
  utls_fingerprint: string;
  ech_enabled: boolean;
  enabled: boolean;
}

interface PresetRow {
  isp: string;
  name: string;
  defaults: TLSProfile;
}

const PRESET_META: Record<string, { badge: string; target: string; descKey: TKey }> = {
  hamrah_aval: { badge: "Most Popular", target: "Mobile (MCI)", descKey: "security.preset.hamrah_aval" },
  irancell: { badge: "Optimized", target: "Mobile (MTN)", descKey: "security.preset.irancell" },
  mokhaberat: { badge: "Stable", target: "Fixed Line (TCI)", descKey: "security.preset.mokhaberat" },
  shatel: { badge: "Ultra Fast", target: "Fixed Line", descKey: "security.preset.shatel" },
  asiatech: { badge: "Ultra Fast", target: "Fixed Line", descKey: "security.preset.asiatech" },
};

function alpnFromProfile(p: TLSProfile): string {
  if (p.ech_enabled) return "h2";
  if (p.utls_fingerprint === "safari") return "http/1.1";
  return "h2, http/1.1";
}

export function TLSTricksTab() {
  const { t } = useI18n();
  const toast = useToast();
  const confirm = useConfirm();
  const qc = useQueryClient();
  const { can } = useAuth();
  const canWrite = can("inbound:write");
  const [createOpen, setCreateOpen] = useState(false);

  const { data } = useQuery({
    queryKey: ["tls-tricks"],
    queryFn: () => api<{ profiles: TLSProfile[] }>("/api/tls-tricks"),
  });
  const { data: presetsData } = useQuery({
    queryKey: ["tls-presets"],
    queryFn: () => api<{ presets: PresetRow[] }>("/api/tls-tricks/presets"),
  });
  const inbounds = useInboundsFleet();

  const profiles = data?.profiles ?? [];
  const presets = (presetsData?.presets ?? []).filter((p) => p.isp !== "custom");

  const inboundCounts = useMemo(() => {
    const map: Record<string, number> = {};
    for (const ib of inbounds.data?.inbounds ?? []) {
      if (ib.evasion_profile_id) {
        map[ib.evasion_profile_id] = (map[ib.evasion_profile_id] ?? 0) + 1;
      }
    }
    return map;
  }, [inbounds.data]);

  const del = useMutation({
    mutationFn: (id: string) => api<void>(`/api/tls-tricks/${id}`, { method: "DELETE" }),
    onSuccess: () => qc.invalidateQueries({ queryKey: ["tls-tricks"] }),
  });

  const fromPreset = useMutation({
    mutationFn: (isp: string) => api("/api/tls-tricks/preset", { method: "POST", body: { isp } }),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["tls-tricks"] });
      toast.success(t("security.tls.profileApplied"));
    },
  });

  function profileForISP(isp: string): TLSProfile | undefined {
    return profiles.find((p) => p.isp === isp);
  }

  return (
    <div className="space-y-4">
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-3">
        <div>
          <h2 className="text-sm font-bold text-fg">{t("security.tls.sectionTitle")}</h2>
          <p className="text-xs text-fg-muted mt-0.5 max-w-2xl">{t("security.tls.sectionDesc")}</p>
        </div>
        {canWrite && (
          <Button size="sm" onClick={() => setCreateOpen(true)}>
            <Plus size={14} /> {t("security.tls.customProfile")}
          </Button>
        )}
      </div>

      <div className="grid grid-cols-1 gap-4 lg:grid-cols-2">
        {presets.map((preset) => {
          const meta = PRESET_META[preset.isp] ?? { badge: "Preset", target: preset.isp, descKey: "security.tls.sectionDesc" as TKey };
          const existing = profileForISP(preset.isp);
          const active = !!existing?.enabled;
          const applied = existing ? inboundCounts[existing.id] ?? 0 : 0;
          const d = preset.defaults;

          return (
            <GlassCard key={preset.isp} hover={false} className="!p-5 space-y-3">
              <div className="flex items-start justify-between gap-2">
                <div className="min-w-0">
                  <div className="flex flex-wrap items-center gap-2">
                    <Badge color="muted">{meta.badge}</Badge>
                    <StatusBadge status={active ? "active" : "inactive"} label={active ? "ACTIVE" : "STANDBY"} pulse={active} />
                  </div>
                  <h3 className="text-sm font-bold text-fg mt-2">{preset.name}</h3>
                  <p className="text-[11px] text-fg-subtle">{meta.target}</p>
                </div>
                <div className={cn("h-9 w-9 rounded-xl flex items-center justify-center flex-shrink-0", "text-primary bg-primary/10")}>
                  <SlidersHorizontal size={16} />
                </div>
              </div>

              <p className="text-xs text-fg-muted leading-relaxed">
                {t(meta.descKey)}
              </p>

              <div className="rounded-lg bg-surface-2/50 border border-border/40 px-3 py-2 grid grid-cols-2 gap-2 text-[11px]">
                <div>
                  <span className="text-fg-subtle">ALPN</span>
                  <p className="font-mono text-fg mt-0.5">{alpnFromProfile(d)}</p>
                </div>
                <div>
                  <span className="text-fg-subtle">{t("security.tls.fragment")}</span>
                  <p className="font-mono text-fg mt-0.5">
                    {d.fragment_enabled ? `${d.fragment_interval}ms` : t("security.tls.disabled")}
                  </p>
                </div>
              </div>

              <div className="flex items-center justify-between gap-2 pt-1 border-t border-border/40">
                {existing && applied > 0 ? (
                  <span className="text-xs text-fg-muted">
                    {t("security.tls.appliedInbounds").replace("{count}", String(applied))}
                  </span>
                ) : (
                  <span className="text-xs text-fg-subtle">uTLS: {d.utls_fingerprint}</span>
                )}
                {canWrite && (
                  <div className="flex gap-1">
                    {!existing ? (
                      <Button size="sm" onClick={() => fromPreset.mutate(preset.isp)} disabled={fromPreset.isPending}>
                        {t("security.tls.applyProfile")}
                      </Button>
                    ) : (
                      <Button
                        size="sm"
                        variant="ghost"
                        className="text-danger text-xs"
                        onClick={async () => {
                          if (await confirm({ title: t("common.delete"), destructive: true })) {
                            await del.mutateAsync(existing.id);
                            toast.success(t("common.delete"));
                          }
                        }}
                      >
                        {t("common.delete")}
                      </Button>
                    )}
                  </div>
                )}
              </div>
            </GlassCard>
          );
        })}
      </div>

      {profiles.filter((p) => p.isp === "custom").length > 0 && (
        <div className="space-y-2">
          <h3 className="text-xs font-semibold text-fg-subtle uppercase tracking-wide">{t("security.tls.customProfiles")}</h3>
          <div className="grid grid-cols-1 gap-3 md:grid-cols-2">
            {profiles.filter((p) => p.isp === "custom").map((p) => (
              <GlassCard key={p.id} hover={false} className="!p-4 flex items-center justify-between gap-2">
                <div>
                  <p className="text-sm font-bold text-fg">{p.name}</p>
                  <p className="text-xs text-fg-muted">Fragment: {p.fragment_enabled ? p.fragment_size : "off"}</p>
                </div>
                <StatusBadge status={p.enabled ? "active" : "inactive"} label={p.enabled ? "ACTIVE" : "OFF"} pulse={false} />
              </GlassCard>
            ))}
          </div>
        </div>
      )}

      <CreateProfileModal open={createOpen} onClose={() => setCreateOpen(false)} />
    </div>
  );
}

function CreateProfileModal({ open, onClose }: { open: boolean; onClose: () => void }) {
  const qc = useQueryClient();
  const toast = useToast();
  const { t } = useI18n();
  const [f, setF] = useState({
    name: "",
    isp: "custom",
    fragment_enabled: true,
    fragment_size: "10-50",
    fragment_interval: "10-20",
    fragment_packets: "tlshello",
    mux_enabled: true,
    mux_concurrency: 8,
    utls_fingerprint: "chrome",
    padding_enabled: true,
    padding_size: "100-200",
    ech_enabled: false,
    enabled: true,
  });

  const create = useMutation({
    mutationFn: (b: typeof f) => api("/api/tls-tricks", { method: "POST", body: b }),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["tls-tricks"] });
      onClose();
      toast.success(t("common.create"));
    },
  });

  return (
    <Modal open={open} onClose={onClose} title={t("security.tls.customProfile")} className="max-w-lg">
      <form
        onSubmit={(e) => {
          e.preventDefault();
          create.mutate(f);
        }}
        className="space-y-3"
      >
        <Input placeholder={t("security.tls.profileName")} value={f.name} onChange={(e) => setF((s) => ({ ...s, name: e.target.value }))} required />
        <div className="grid grid-cols-2 gap-2">
          <Input value={f.fragment_size} onChange={(e) => setF((s) => ({ ...s, fragment_size: e.target.value }))} placeholder="Fragment size" />
          <Select value={f.utls_fingerprint} onChange={(e) => setF((s) => ({ ...s, utls_fingerprint: e.target.value }))}>
            <option value="chrome">Chrome</option>
            <option value="firefox">Firefox</option>
            <option value="safari">Safari</option>
            <option value="random">Random</option>
          </Select>
        </div>
        <div className="flex justify-end gap-2 pt-2">
          <Button type="button" variant="ghost" onClick={onClose}>{t("common.cancel")}</Button>
          <Button type="submit" disabled={create.isPending}>{t("common.create")}</Button>
        </div>
      </form>
    </Modal>
  );
}
