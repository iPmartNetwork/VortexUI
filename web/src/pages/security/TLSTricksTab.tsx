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
  mux_enabled?: boolean;
  ech_enabled: boolean;
  enabled: boolean;
}

interface PresetRow {
  isp: string;
  name: string;
  defaults: TLSProfile;
}

interface ProfileFormState {
  id?: string;
  name: string;
  isp: string;
  fragment_enabled: boolean;
  fragment_size: string;
  fragment_interval: string;
  utls_fingerprint: string;
  mux_enabled: boolean;
  enabled: boolean;
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

function profileToForm(p: TLSProfile): ProfileFormState {
  return {
    id: p.id,
    name: p.name,
    isp: p.isp || "custom",
    fragment_enabled: p.fragment_enabled,
    fragment_size: p.fragment_size,
    fragment_interval: p.fragment_interval,
    utls_fingerprint: p.utls_fingerprint,
    mux_enabled: p.mux_enabled ?? true,
    enabled: p.enabled,
  };
}

function defaultsToForm(defaults: TLSProfile, isp: string): ProfileFormState {
  return {
    name: defaults.name,
    isp,
    fragment_enabled: defaults.fragment_enabled,
    fragment_size: defaults.fragment_size,
    fragment_interval: defaults.fragment_interval,
    utls_fingerprint: defaults.utls_fingerprint,
    mux_enabled: defaults.mux_enabled ?? true,
    enabled: true,
  };
}

export function TLSTricksTab() {
  const { t } = useI18n();
  const toast = useToast();
  const confirm = useConfirm();
  const qc = useQueryClient();
  const { can } = useAuth();
  const canWrite = can("inbound:write");
  const [createOpen, setCreateOpen] = useState(false);
  const [editForm, setEditForm] = useState<ProfileFormState | null>(null);

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

  const applyToInbounds = useMutation({
    mutationFn: (id: string) => api<{ applied: number }>(`/api/tls-tricks/${id}/apply`, { method: "POST", body: {} }),
    onSuccess: (res) => {
      qc.invalidateQueries({ queryKey: ["inbounds"] });
      qc.invalidateQueries({ queryKey: ["inbounds-fleet"] });
      toast.success(
        t("security.tls.appliedInbounds").replace("{count}", String(res.applied ?? 0)),
      );
    },
    onError: () => toast.error(t("common.failed")),
  });

  function profileForISP(isp: string, presetName?: string): TLSProfile | undefined {
    return (
      profiles.find((p) => p.isp === isp) ??
      (presetName ? profiles.find((p) => p.name === presetName) : undefined)
    );
  }

  function openEdit(existing: TLSProfile | undefined, defaults: TLSProfile, isp: string) {
    setEditForm(existing ? profileToForm(existing) : defaultsToForm(defaults, isp));
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
          const existing = profileForISP(preset.isp, preset.name);
          const active = !!existing?.enabled;
          const applied = existing ? inboundCounts[existing.id] ?? 0 : 0;
          const d = existing ?? preset.defaults;

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
                <button
                  type="button"
                  onClick={() => openEdit(existing, preset.defaults, preset.isp)}
                  className={cn(
                    "h-9 w-9 rounded-xl flex items-center justify-center flex-shrink-0 transition-colors",
                    "text-primary bg-primary/10 hover:bg-primary/20",
                  )}
                  aria-label={t("security.tls.editProfile")}
                  title={t("security.tls.editProfile")}
                >
                  <SlidersHorizontal size={16} />
                </button>
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
                      <>
                        <Button
                          size="sm"
                          onClick={() => applyToInbounds.mutate(existing.id)}
                          disabled={applyToInbounds.isPending}
                        >
                          {t("security.tls.applyToInbounds")}
                        </Button>
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
                      </>
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
                <div className="flex items-center gap-2">
                  <StatusBadge status={p.enabled ? "active" : "inactive"} label={p.enabled ? "ACTIVE" : "OFF"} pulse={false} />
                  <button
                    type="button"
                    onClick={() => setEditForm(profileToForm(p))}
                    className="h-8 w-8 rounded-lg flex items-center justify-center text-primary bg-primary/10 hover:bg-primary/20 transition-colors"
                    aria-label={t("security.tls.editProfile")}
                  >
                    <SlidersHorizontal size={14} />
                  </button>
                </div>
              </GlassCard>
            ))}
          </div>
        </div>
      )}

      <CreateProfileModal open={createOpen} onClose={() => setCreateOpen(false)} />
      {editForm && (
        <ProfileEditModal
          key={editForm.id ?? editForm.isp}
          open
          initial={editForm}
          readOnly={!canWrite}
          onClose={() => setEditForm(null)}
        />
      )}
    </div>
  );
}

function ProfileEditModal({
  open,
  onClose,
  initial,
  readOnly,
}: {
  open: boolean;
  onClose: () => void;
  initial: ProfileFormState;
  readOnly?: boolean;
}) {
  const qc = useQueryClient();
  const toast = useToast();
  const { t } = useI18n();
  const [f, setF] = useState(initial);

  const save = useMutation({
    mutationFn: async (body: ProfileFormState) => {
      const payload = {
        name: body.name,
        isp: body.isp,
        fragment_enabled: body.fragment_enabled,
        fragment_size: body.fragment_size,
        fragment_interval: body.fragment_interval,
        utls_fingerprint: body.utls_fingerprint,
        mux_enabled: body.mux_enabled,
        enabled: body.enabled,
      };
      if (body.id) {
        return api(`/api/tls-tricks/${body.id}`, { method: "PUT", body: payload });
      }
      return api("/api/tls-tricks", { method: "POST", body: payload });
    },
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["tls-tricks"] });
      onClose();
      toast.success(t("security.tls.profileSaved"));
    },
  });

  return (
    <Modal open={open} onClose={onClose} title={t("security.tls.editProfile")} className="max-w-lg">
      <form
        onSubmit={(e) => {
          e.preventDefault();
          if (!readOnly) save.mutate(f);
        }}
        className="space-y-3"
      >
        <Input
          placeholder={t("security.tls.profileName")}
          value={f.name}
          onChange={(e) => setF((s) => ({ ...s, name: e.target.value }))}
          required
          disabled={readOnly}
        />
        <div className="grid grid-cols-2 gap-2">
          <div>
            <label className="text-xs text-fg-subtle">{t("security.tls.fragment")} size</label>
            <Input
              value={f.fragment_size}
              onChange={(e) => setF((s) => ({ ...s, fragment_size: e.target.value }))}
              placeholder="10-50"
              disabled={readOnly || !f.fragment_enabled}
            />
          </div>
          <div>
            <label className="text-xs text-fg-subtle">{t("security.tls.fragment")} interval (ms)</label>
            <Input
              value={f.fragment_interval}
              onChange={(e) => setF((s) => ({ ...s, fragment_interval: e.target.value }))}
              placeholder="10-20"
              disabled={readOnly || !f.fragment_enabled}
            />
          </div>
        </div>
        <div>
          <label className="text-xs text-fg-subtle">uTLS fingerprint</label>
          <Select
            value={f.utls_fingerprint}
            onChange={(e) => setF((s) => ({ ...s, utls_fingerprint: e.target.value }))}
            disabled={readOnly}
          >
            <option value="chrome">Chrome</option>
            <option value="firefox">Firefox</option>
            <option value="safari">Safari</option>
            <option value="random">Random</option>
          </Select>
        </div>
        <div className="space-y-2 rounded-lg border border-border/40 px-3 py-2">
          <label className="flex items-center gap-2 text-sm text-fg">
            <input
              type="checkbox"
              checked={f.fragment_enabled}
              onChange={(e) => setF((s) => ({ ...s, fragment_enabled: e.target.checked }))}
              disabled={readOnly}
              className="rounded"
            />
            {t("security.tls.fragment")} enabled
          </label>
          <label className="flex items-center gap-2 text-sm text-fg">
            <input
              type="checkbox"
              checked={f.mux_enabled}
              onChange={(e) => setF((s) => ({ ...s, mux_enabled: e.target.checked }))}
              disabled={readOnly}
              className="rounded"
            />
            Multiplexing (mux)
          </label>
          <label className="flex items-center gap-2 text-sm text-fg">
            <input
              type="checkbox"
              checked={f.enabled}
              onChange={(e) => setF((s) => ({ ...s, enabled: e.target.checked }))}
              disabled={readOnly}
              className="rounded"
            />
            {t("common.enabled")}
          </label>
        </div>
        <div className="flex justify-end gap-2 pt-2">
          <Button type="button" variant="ghost" onClick={onClose}>{readOnly ? t("common.close") : t("common.cancel")}</Button>
          {!readOnly && (
            <Button type="submit" disabled={save.isPending}>{t("common.save")}</Button>
          )}
        </div>
      </form>
    </Modal>
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
