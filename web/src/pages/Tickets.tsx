import { useState } from "react";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { api } from "@/api/client";
import { Button, Card, Input, PageHeader, Badge, Select } from "@/components/ui";
import { Modal } from "@/components/Modal";
import { useToast } from "@/components/toast";

interface Ticket {
  id: string;
  user_id: string;
  subject: string;
  status: string;
  priority: string;
  created_at: string;
  updated_at: string;
}

interface TicketMessage {
  id: string;
  sender: string;
  sender_id: string;
  body: string;
  created_at: string;
}

export function Tickets() {
  const [viewId, setViewId] = useState<string | null>(null);
  const [filter, setFilter] = useState("");

  const { data } = useQuery({
    queryKey: ["admin-tickets", filter],
    queryFn: () => api<{ tickets: Ticket[]; total: number }>("/api/tickets", { query: { status: filter } }),
  });

  return (
    <div className="space-y-6 animate-fade-in">
      <div className="flex items-center justify-between">
        <PageHeader title="Support Tickets" subtitle="Manage user support requests" />
        <Select value={filter} onChange={(e) => setFilter(e.target.value)}>
          <option value="">All</option>
          <option value="open">Open</option>
          <option value="answered">Answered</option>
          <option value="closed">Closed</option>
        </Select>
      </div>

      {viewId && <TicketDetailModal ticketId={viewId} onClose={() => setViewId(null)} />}

      <div className="space-y-3">
        {data?.tickets?.map((t) => (
          <Card key={t.id} className="flex items-center justify-between cursor-pointer hover:ring-1 hover:ring-primary/30" onClick={() => setViewId(t.id)}>
            <div className="space-y-1">
              <h3 className="text-sm font-medium text-fg">{t.subject}</h3>
              <div className="flex gap-2 items-center">
                <Badge color={t.status === "open" ? "active" : t.status === "answered" ? "limited" : "disabled"}>{t.status}</Badge>
                <Badge color={t.priority === "high" ? "expired" : "muted"}>{t.priority}</Badge>
                <span className="text-xs text-fg-subtle">{new Date(t.updated_at).toLocaleString()}</span>
              </div>
            </div>
          </Card>
        ))}
        {(!data?.tickets || data.tickets.length === 0) && (
          <p className="text-center text-sm text-fg-muted py-8">No tickets found.</p>
        )}
      </div>
    </div>
  );
}

function TicketDetailModal({ ticketId, onClose }: { ticketId: string; onClose: () => void }) {
  const qc = useQueryClient();
  const toast = useToast();
  const [reply, setReply] = useState("");

  // Reuse the portal ticket endpoint via admin API
  const { data } = useQuery({
    queryKey: ["admin-ticket-detail", ticketId],
    queryFn: async () => {
      // Admin views messages via direct query — for now fetch list + messages from portal endpoint concept
      // In production this would be a dedicated admin ticket detail endpoint
      return api<{ ticket: Ticket & { messages: TicketMessage[] } }>(`/api/tickets/${ticketId}` as any);
    },
    retry: false,
  });

  const replyMut = useMutation({
    mutationFn: (body: string) => api(`/api/tickets/${ticketId}/reply`, { method: "POST", body: { body } }),
    onSuccess: () => { qc.invalidateQueries({ queryKey: ["admin-ticket-detail", ticketId] }); setReply(""); toast.success("Reply sent"); },
  });

  const closeMut = useMutation({
    mutationFn: () => api(`/api/tickets/${ticketId}/close`, { method: "POST" }),
    onSuccess: () => { qc.invalidateQueries({ queryKey: ["admin-tickets"] }); onClose(); toast.success("Ticket closed"); },
  });

  const ticket = data?.ticket;

  return (
    <Modal open={true} onClose={onClose} title={ticket?.subject || "Ticket"} className="max-w-lg">
      <div className="space-y-4 max-h-[60vh] overflow-y-auto">
        {ticket?.messages?.map((m: TicketMessage) => (
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
        <div className="mt-4 space-y-2">
          <form onSubmit={(e) => { e.preventDefault(); if (reply.trim()) replyMut.mutate(reply); }} className="flex gap-2">
            <Input placeholder="Type a reply..." value={reply} onChange={(e) => setReply(e.target.value)} className="flex-1" />
            <Button type="submit" disabled={replyMut.isPending || !reply.trim()} size="sm">Reply</Button>
          </form>
          <Button variant="outline" size="sm" className="w-full" onClick={() => closeMut.mutate()} disabled={closeMut.isPending}>
            Close Ticket
          </Button>
        </div>
      )}
    </Modal>
  );
}
