import { useState } from "react";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { motion, AnimatePresence } from "framer-motion";
import { Plus, MessageSquare, Clock, ChevronRight, Send, AlertCircle } from "lucide-react";
import { portalApi } from "./portalApi";
import { Button, Input } from "@/components/ui";
import { GlassCard } from "@/components/vortexui";
import { Modal } from "@/components/Modal";
import { useToast } from "@/components/toast";
import { useI18n } from "@/i18n/i18n";
import { cn } from "@/lib/utils";

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

const STATUS_STYLES: Record<string, { label: string; dot: string; bg: string }> = {
  open: { label: "Open", dot: "bg-green-500", bg: "bg-green-500/10 border-green-500/30 text-green-400" },
  answered: { label: "Answered", dot: "bg-blue-500", bg: "bg-blue-500/10 border-blue-500/30 text-blue-400" },
  closed: { label: "Closed", dot: "bg-fg-subtle/50", bg: "bg-surface-2/50 border-border/40 text-fg-muted" },
  pending: { label: "Pending", dot: "bg-amber-500", bg: "bg-amber-500/10 border-amber-500/30 text-amber-400" },
};

export function PortalTickets() {
  const [createOpen, setCreateOpen] = useState(false);
  const [viewId, setViewId] = useState<string | null>(null);
  const { t } = useI18n();

  const { data } = useQuery({
    queryKey: ["portal-tickets"],
    queryFn: () => portalApi<{ tickets: Ticket[] }>("/api/portal/tickets"),
  });

  const tickets = data?.tickets ?? [];

  return (
    <div className="space-y-6 animate-page-enter">
      {/* ── Header ── */}
      <div className="relative overflow-hidden rounded-2xl border border-border/60 bg-gradient-to-br from-bg-elevated via-surface to-primary/[0.03] p-5 md:p-6">
        <div className="absolute top-0 end-0 w-56 h-56 rounded-full bg-primary/8 blur-3xl pointer-events-none" />
        <div className="absolute bottom-0 start-0 w-40 h-40 rounded-full bg-accent/8 blur-3xl pointer-events-none" />
        
        <div className="relative z-10 flex flex-col sm:flex-row items-start justify-between gap-4">
          <div className="space-y-1">
            <h1 className="text-xl md:text-2xl font-black text-fg tracking-tight flex items-center gap-2">
              <MessageSquare size={22} className="text-primary" />
              {t("portal.ticketsTitle")}
            </h1>
            <p className="text-[13px] text-fg-muted">
              {t("portal.ticketsCount").replace("{count}", String(tickets.length))}
            </p>
          </div>
          <Button onClick={() => setCreateOpen(true)} className="gap-1.5">
            <Plus size={15} />
            {t("portal.newTicket")}
          </Button>
        </div>
      </div>

      {/* ── Create Modal ── */}
      <CreateTicketModal open={createOpen} onClose={() => setCreateOpen(false)} />
      
      {/* ── View Modal ── */}
      {viewId && <ViewTicketModal ticketId={viewId} onClose={() => setViewId(null)} />}

      {/* ── Tickets List ── */}
      <div className="space-y-2.5">
        {tickets.length === 0 && (
          <div className="flex flex-col items-center justify-center py-16 text-center gap-3">
            <div className="h-12 w-12 rounded-full bg-surface-2/50 flex items-center justify-center">
              <MessageSquare size={22} className="text-fg-subtle" />
            </div>
            <p className="text-sm text-fg-muted">No tickets yet.</p>
            <Button size="sm" variant="glass" onClick={() => setCreateOpen(true)}>
              <Plus size={14} /> Create your first ticket
            </Button>
          </div>
        )}

        <AnimatePresence>
          {tickets.map((t, i) => {
            const st = STATUS_STYLES[t.status] ?? STATUS_STYLES.pending;
            return (
              <motion.div
                key={t.id}
                initial={{ opacity: 0, y: 8 }}
                animate={{ opacity: 1, y: 0 }}
                transition={{ delay: i * 0.04 }}
              >
                <GlassCard
                  hover
                  glow
                  className="cursor-pointer transition-all duration-200 group"
                  onClick={() => setViewId(t.id)}
                >
                  <div className="flex items-center justify-between gap-4">
                    <div className="space-y-1.5 min-w-0 flex-1">
                      <div className="flex items-center gap-2">
                        <h3 className="text-sm font-bold text-fg truncate group-hover:text-primary transition-colors">
                          {t.subject}
                        </h3>
                      </div>
                      <div className="flex items-center gap-2.5">
                        <span className={cn("inline-flex items-center gap-1 rounded-full px-2 py-0.5 text-[9px] font-bold uppercase tracking-wider border", st.bg)}>
                          <span className={cn("h-1.5 w-1.5 rounded-full", st.dot)} />
                          {st.label}
                        </span>
                        <span className="text-[10px] text-fg-subtle flex items-center gap-1">
                          <Clock size={10} />
                          {new Date(t.updated_at).toLocaleDateString([], {
                            month: "short",
                            day: "numeric",
                          })}
                        </span>
                        {t.priority === "high" && (
                          <span className="text-[9px] font-bold text-danger flex items-center gap-0.5">
                            <AlertCircle size={10} />
                            High
                          </span>
                        )}
                      </div>
                    </div>
                    <ChevronRight size={16} className="text-fg-subtle group-hover:text-primary transition-colors flex-shrink-0" />
                  </div>
                </GlassCard>
              </motion.div>
            );
          })}
        </AnimatePresence>
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
    onSuccess: () => { qc.invalidateQueries({ queryKey: ["portal-tickets"] }); onClose(); toast.success("Ticket created"); setF({ subject: "", body: "", priority: "medium" }); },
  });

  return (
    <Modal open={open} onClose={onClose} title="New Ticket" size="md">
      <form onSubmit={(e) => { e.preventDefault(); create.mutate(f); }} className="space-y-4">
        <div className="space-y-1.5">
          <label className="text-[10px] font-bold uppercase tracking-wider text-fg-subtle/70">Subject</label>
          <Input
            placeholder="Brief title for your issue"
            value={f.subject}
            onChange={(e) => setF(s => ({ ...s, subject: e.target.value }))}
            required
          />
        </div>
        <div className="space-y-1.5">
          <label className="text-[10px] font-bold uppercase tracking-wider text-fg-subtle/70">Description</label>
          <textarea
            placeholder="Describe your issue in detail..."
            value={f.body}
            onChange={(e) => setF(s => ({ ...s, body: e.target.value }))}
            className="w-full rounded-xl border border-border/60 bg-surface/60 px-3.5 py-2.5 text-sm text-fg placeholder:text-fg-subtle/40 min-h-[120px] resize-y focus:outline-none focus:border-primary/50 transition"
            required
          />
        </div>
        <div className="space-y-1.5">
          <label className="text-[10px] font-bold uppercase tracking-wider text-fg-subtle/70">Priority</label>
          <div className="flex gap-2">
            {(["low", "medium", "high"] as const).map((p) => (
              <button
                key={p}
                type="button"
                onClick={() => setF(s => ({ ...s, priority: p }))}
                className={cn(
                  "flex-1 rounded-lg px-3 py-2 text-xs font-semibold transition-all duration-200 border",
                  f.priority === p
                    ? p === "high"
                      ? "bg-danger/10 border-danger/30 text-danger"
                      : p === "medium"
                        ? "bg-warning/10 border-warning/30 text-warning"
                        : "bg-success/10 border-success/30 text-success"
                    : "bg-surface-2/40 border-border/40 text-fg-muted hover:text-fg hover:bg-surface-2",
                )}
              >
                {p.charAt(0).toUpperCase() + p.slice(1)}
              </button>
            ))}
          </div>
        </div>
        <div className="flex justify-end gap-2 pt-2">
          <Button type="button" variant="ghost" onClick={onClose}>Cancel</Button>
          <Button type="submit" disabled={create.isPending} className="gap-1.5">
            {create.isPending ? (
              <><div className="h-3 w-3 animate-spin rounded-full border-2 border-white border-t-transparent" /> Submitting</>
            ) : (
              <><Send size={14} /> Submit</>
            )}
          </Button>
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
    <Modal open={true} onClose={onClose} title={ticket?.subject || "Ticket"} size="lg">
      <div className="space-y-4 max-h-[55vh] overflow-y-auto pr-1">
        {ticket?.messages?.map((m) => {
          const isAdmin = m.sender === "admin";
          return (
            <motion.div
              key={m.id}
              initial={{ opacity: 0, y: 6 }}
              animate={{ opacity: 1, y: 0 }}
              className={cn(
                "rounded-xl px-4 py-3 text-sm border max-w-[85%]",
                isAdmin
                  ? "bg-primary/8 border-primary/20 ml-auto"
                  : "bg-surface-2/60 border-border/40",
              )}
            >
              <div className="flex items-center justify-between gap-3 mb-1.5">
                <span className={cn(
                  "text-[10px] font-bold uppercase tracking-wider",
                  isAdmin ? "text-primary" : "text-fg-subtle",
                )}>
                  {isAdmin ? "Support" : "You"}
                </span>
                <span className="text-[10px] text-fg-subtle/70">
                  {new Date(m.created_at).toLocaleString([], {
                    month: "short",
                    day: "numeric",
                    hour: "2-digit",
                    minute: "2-digit",
                  })}
                </span>
              </div>
              <p className="text-fg whitespace-pre-wrap leading-relaxed text-[13px]">{m.body}</p>
            </motion.div>
          );
        })}
      </div>

      {ticket?.status !== "closed" && (
        <form
          onSubmit={(e) => { e.preventDefault(); if (reply.trim()) replyMut.mutate(reply); }}
          className="mt-4 pt-4 border-t border-border/40"
        >
          <div className="flex gap-2">
            <Input
              placeholder="Type a reply..."
              value={reply}
              onChange={(e) => setReply(e.target.value)}
              className="flex-1"
            />
            <Button
              type="submit"
              disabled={replyMut.isPending || !reply.trim()}
              size="sm"
              className="gap-1"
            >
              {replyMut.isPending ? (
                <div className="h-3 w-3 animate-spin rounded-full border-2 border-white border-t-transparent" />
              ) : (
                <Send size={14} />
              )}
            </Button>
          </div>
        </form>
      )}
    </Modal>
  );
}
