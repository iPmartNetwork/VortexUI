import { useState } from "react";
import { Button, Input, Select } from "@/components/ui";
import type { UserTemplate } from "./TemplateListPage";

export interface TemplateFormData {
  name: string;
  data_limit: number;
  expire_duration: number | null;
  device_limit: number;
  reset_strategy: string;
  note: string;
  protocol_settings: Record<string, unknown>;
  groups: string[];
}

const GB = 1024 * 1024 * 1024;
const DAY_SECONDS = 86400;

interface TemplateFormProps {
  initialData: UserTemplate | null;
  onSubmit: (data: TemplateFormData) => Promise<void>;
  onCancel: () => void;
  isPending: boolean;
}

export function TemplateForm({ initialData, onSubmit, onCancel, isPending }: TemplateFormProps) {
  const [name, setName] = useState(initialData?.name ?? "");
  const [dataLimitGB, setDataLimitGB] = useState(
    initialData ? String(initialData.data_limit / GB || "") : "",
  );
  const [expireDays, setExpireDays] = useState(
    initialData?.expire_duration ? String(Math.round(initialData.expire_duration / DAY_SECONDS)) : "",
  );
  const [deviceLimit, setDeviceLimit] = useState(
    initialData ? String(initialData.device_limit || "") : "",
  );
  const [resetStrategy, setResetStrategy] = useState(initialData?.reset_strategy ?? "no_reset");
  const [note, setNote] = useState(initialData?.note ?? "");
  const [groupsInput, setGroupsInput] = useState(initialData?.groups.join(", ") ?? "");
  const [protocolJson, setProtocolJson] = useState(
    initialData?.protocol_settings
      ? JSON.stringify(initialData.protocol_settings, null, 2)
      : "{}",
  );
  const [error, setError] = useState("");

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    setError("");

    // Validate
    if (!name.trim()) {
      setError("Name is required");
      return;
    }

    let parsedProtocol: Record<string, unknown>;
    try {
      parsedProtocol = JSON.parse(protocolJson);
    } catch {
      setError("Protocol settings must be valid JSON");
      return;
    }

    const groups = groupsInput
      .split(",")
      .map((s) => s.trim())
      .filter(Boolean);

    const data: TemplateFormData = {
      name: name.trim(),
      data_limit: dataLimitGB ? Math.round(Number(dataLimitGB) * GB) : 0,
      expire_duration: expireDays ? Math.round(Number(expireDays) * DAY_SECONDS) : null,
      device_limit: deviceLimit ? Number(deviceLimit) : 0,
      reset_strategy: resetStrategy,
      note: note.trim(),
      protocol_settings: parsedProtocol,
      groups,
    };

    try {
      await onSubmit(data);
    } catch (err: unknown) {
      setError(err instanceof Error ? err.message : "Failed to save template");
    }
  }

  return (
    <form onSubmit={handleSubmit} className="space-y-4">
      {/* Name */}
      <div>
        <label className="mb-1 block text-xs font-medium text-fg-muted">Template Name</label>
        <Input
          value={name}
          onChange={(e) => setName(e.target.value)}
          placeholder="e.g. Premium Monthly"
          required
          autoFocus
        />
      </div>

      {/* Data & Expire */}
      <div className="grid grid-cols-2 gap-3">
        <div>
          <label className="mb-1 block text-xs font-medium text-fg-muted">Data Limit (GB)</label>
          <Input
            value={dataLimitGB}
            onChange={(e) => setDataLimitGB(e.target.value)}
            placeholder="0 = unlimited"
            inputMode="decimal"
          />
        </div>
        <div>
          <label className="mb-1 block text-xs font-medium text-fg-muted">Expire Duration (days)</label>
          <Input
            value={expireDays}
            onChange={(e) => setExpireDays(e.target.value)}
            placeholder="Empty = never"
            inputMode="numeric"
          />
        </div>
      </div>

      {/* Device limit & Reset strategy */}
      <div className="grid grid-cols-2 gap-3">
        <div>
          <label className="mb-1 block text-xs font-medium text-fg-muted">Device Limit</label>
          <Input
            value={deviceLimit}
            onChange={(e) => setDeviceLimit(e.target.value)}
            placeholder="0 = unlimited"
            inputMode="numeric"
          />
        </div>
        <div>
          <label className="mb-1 block text-xs font-medium text-fg-muted">Reset Strategy</label>
          <Select value={resetStrategy} onChange={(e) => setResetStrategy(e.target.value)}>
            <option value="no_reset">No Reset</option>
            <option value="daily">Daily</option>
            <option value="weekly">Weekly</option>
            <option value="monthly">Monthly</option>
          </Select>
        </div>
      </div>

      {/* Groups */}
      <div>
        <label className="mb-1 block text-xs font-medium text-fg-muted">
          Groups <span className="text-fg-subtle font-normal">(comma-separated)</span>
        </label>
        <Input
          value={groupsInput}
          onChange={(e) => setGroupsInput(e.target.value)}
          placeholder="e.g. premium, reseller-a"
        />
      </div>

      {/* Note */}
      <div>
        <label className="mb-1 block text-xs font-medium text-fg-muted">Note</label>
        <textarea
          value={note}
          onChange={(e) => setNote(e.target.value)}
          className="field input-surface w-full resize-none"
          rows={2}
          placeholder="Optional description..."
        />
      </div>

      {/* Protocol settings */}
      <div>
        <label className="mb-1 block text-xs font-medium text-fg-muted">Protocol Settings (JSON)</label>
        <textarea
          value={protocolJson}
          onChange={(e) => setProtocolJson(e.target.value)}
          className="field input-surface w-full resize-none font-mono text-xs"
          rows={4}
          placeholder="{}"
        />
      </div>

      {/* Error */}
      {error && <p className="text-sm text-danger">{error}</p>}

      {/* Actions */}
      <div className="flex justify-end gap-2 pt-2">
        <Button type="button" variant="ghost" onClick={onCancel}>
          Cancel
        </Button>
        <Button type="submit" disabled={isPending}>
          {isPending ? "Saving..." : initialData ? "Update" : "Create"}
        </Button>
      </div>
    </form>
  );
}
