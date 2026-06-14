import { useState } from "react";
import { useCreateNode } from "@/api/hooks";
import { Button, Input, Select } from "./ui";
import { Modal } from "./Modal";
import { useToast } from "./toast";

export function CreateNodeModal({ open, onClose }: { open: boolean; onClose: () => void }) {
  const create = useCreateNode();
  const toast = useToast();
  const [name, setName] = useState("");
  const [address, setAddress] = useState("");
  const [core, setCore] = useState("xray");
  const [error, setError] = useState("");

  function close() {
    setName("");
    setAddress("");
    setCore("xray");
    setError("");
    onClose();
  }

  async function submit(e: React.FormEvent) {
    e.preventDefault();
    setError("");
    try {
      await create.mutateAsync({ name, address, core });
      toast.success(`Node ${name} added`);
      close();
    } catch {
      setError("Could not create node (name taken?)");
    }
  }

  return (
    <Modal open={open} onClose={close} title="New node">
      <form onSubmit={submit} className="space-y-3">
        <Input placeholder="Name (e.g. de-1)" value={name} onChange={(e) => setName(e.target.value)} required autoFocus />
        <Input
          placeholder="Agent address (host:50051)"
          value={address}
          onChange={(e) => setAddress(e.target.value)}
          required
        />
        <Select value={core} onChange={(e) => setCore(e.target.value)}>
          <option value="xray">Xray-core</option>
          <option value="singbox">sing-box</option>
        </Select>
        {error && <p className="text-sm text-destructive">{error}</p>}
        <div className="flex justify-end gap-2 pt-1">
          <Button type="button" variant="ghost" onClick={close}>
            Cancel
          </Button>
          <Button type="submit" disabled={create.isPending}>
            {create.isPending ? "Adding…" : "Add node"}
          </Button>
        </div>
      </form>
    </Modal>
  );
}
