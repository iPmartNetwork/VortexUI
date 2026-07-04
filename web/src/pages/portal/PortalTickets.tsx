import { useState } from "react";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { portalApi } from "./portalApi";
import { Button, Card, Input, Badge, Select, PageHeader } from "@/components/ui";
import { Modal } from "@/components/Modal";
import { useToast } from "@/components/toast";
import { useI18n } from "@/i18n/i18n";

interface Ticket {
  id: string;
  subject: string;
  status: string;
  priority: string;
  created_at: string;
  updated_at: string;
}

interface TicketMessage {
  id: string;
  sender: string;
  body: string;
  created_at: string;
}

interface TicketDetail extends Ticket {
  messages: TicketMessage[];
}

export function PortalTickets() {
  const [createOpen, setCreateOpen] = useState(false);
  const [viewId, setViewId] = useState<string | null>(null);
  const { t } = useI18n();

  const { data } = useQuery({
    queryKey: ["portal-tickets"],
    queryFn: () => portalApi<{ tickets: Ticket[] }>("/api/portal/tickets"),
  });

  return (
    <div className="space-y-6">
      <PageHeader
        title={t("portal.ticketsTitle")}
        subtitle={t("portal.ticketsCount").replace("{count}", String(data?.tickets?.length ?? 0))}
      >
        <Button onClick={() => setCreateOpen(true)}>{t("portal.newTicket")}</Button>
      </PageHeader>

      <CreateTicketModal open={createOpen} onClose={() => setCreateOpen(false)} />
      {viewId && <ViewTicketModal ticketId={viewId} onClose={() => setViewId(null)} />}

      <div className="space-y-3">
        {data?.tickets?.map((t) => (
          <Card key={t.id} className="flex items-center justify-between cursor-pointer hover:ring-1 hover:ring-primary/30" onClick={() => setViewId(t.id)}>
            <div className="space-y-1">
              <h3 className="text-sm font-medium text-fg">{t.subject}</h3>
              <div className="flex gap-2">
                <Badge color={t.status === "open" ? "active" : t.status === "answered" ? "limited" : "disabled"}>
                  {t.status}
                </Badge>
                <span className="text-xs text-fg-subtle">{new Date(t.updated_at).toLocaleDateString()}</span>
              </div>
            </div>
          </Card>
        ))}
        {(!data?.tickets || data.tickets.length === 0) && (
          <p className="text-center text-sm text-fg-muted py-8">No tickets yet.</p>
        )}
      </div>
    </div>
  );
}

function CreateTicketModal({ open, onClose }: { open: boolean; onClose: () => void }) {
  const qc = useQueryClient();
  const toast = useToast();
  const [f, setF] = useState({ subject: "", body: "", priority: "medium" });
  const create = useMutation({
    mutationFn: (input: Record<string, string>) => portalApi("/api/portal/tickets", { method: "POST", body: input }),
    onSuccess: () => { qc.invalidateQueries({ queryKey: ["portal-tickets"] }); onClose(); toast.success("Ticket created"); },
  });

  return (
    <Modal open={open} onClose={onClose} title="New Ticket">
      <form onSubmit={(e) => { e.preventDefault(); create.mutate(f); }} className="space-y-3">
        <Input placeholder="Subject" value={f.subject} onChange={(e) => setF(s => ({ ...s, subject: e.target.value }))} required />
        <textarea
          placeholder="Describe your issue..."
          value={f.body}
          onChange={(e) => setF(s => ({ ...s, body: e.target.value }))}
          className="field min-h-[100px] resize-y"
          required
        />
        <Select value={f.priority} onChange={(e) => setF(s => ({ ...s, priority: e.target.value }))}>
          <option value="low">Low</option>
          <option value="medium">Medium</option>
          <option value="high">High</option>
        </Select>
        <div className="flex justify-end gap-2 pt-2">
          <Button type="button" variant="ghost" onClick={onClose}>Cancel</Button>
          <Button type="submit" disabled={create.isPending}>Submit</Button>
        </div>
      </form>
    </Modal>
  );
}

function ViewTicketModal({ ticketId, onClose }: { ticketId: string; onClose: () => void }) {
  const qc = useQueryClient();
  const toast = useToast();
  const [reply, setReply] = useState("");

  const { data } = useQuery({
    queryKey: ["portal-ticket", ticketId],
    queryFn: () => portalApi<{ ticket: TicketDetail }>(`/api/portal/tickets/${ticketId}`),
  });

  const replyMut = useMutation({
    mutationFn: (body: string) => portalApi(`/api/portal/tickets/${ticketId}/reply`, { method: "POST", body: { body } }),
    onSuccess: () => { qc.invalidateQueries({ queryKey: ["portal-ticket", ticketId] }); setReply(""); toast.success("Reply sent"); },
  });

  const ticket = data?.ticket;

  return (
    <Modal open={true} onClose={onClose} title={ticket?.subject || "Ticket"} className="max-w-lg">
      <div className="space-y-4 max-h-[60vh] overflow-y-auto">
        {ticket?.messages?.map((m) => (
          <div key={m.id} className={`rounded-lg p-3 text-sm ${m.sender === "admin" ? "bg-primary/10 ml-4" : "bg-surface-2 mr-4"}`}>
            <div className="flex justify-between items-center mb-1">
              <span className="text-xs font-medium text-fg-subtle capitalize">{m.sender}</span>
              <span className="text-xs text-fg-subtle">{new Date(m.created_at).toLocaleString()}</span>
            </div>
            <p className="text-fg whitespace-pre-wrap">{m.body}</p>
          </div>
        ))}
      </div>
      {ticket?.status !== "closed" && (
        <form onSubmit={(e) => { e.preventDefault(); if (reply.trim()) replyMut.mutate(reply); }} className="mt-4 flex gap-2">
          <Input placeholder="Type a reply..." value={reply} onChange={(e) => setReply(e.target.value)} className="flex-1" />
          <Button type="submit" disabled={replyMut.isPending || !reply.trim()} size="sm">Send</Button>
        </form>
      )}
    </Modal>
  );
}
