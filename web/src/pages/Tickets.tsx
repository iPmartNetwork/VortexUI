import { useEffect, useMemo, useState } from "react";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { RefreshCw, Send, X } from "lucide-react";
import { api } from "@/api/client";
import { Badge, Button } from "@/components/ui";
import { GlassCard, StatusBadge } from "@/components/veltrix";
import { useToast } from "@/components/toast";
import { useI18n } from "@/i18n/i18n";
import { useTitle } from "@/lib/useTitle";
import { cn } from "@/lib/utils";

interface Ticket {
  id: string;
  user_id: string;
  username?: string;
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

type Filter = "" | "open" | "answered" | "closed";

const CANNED_KEYS = [
  "tickets.canned.hamrah",
  "tickets.canned.gaming",
  "tickets.canned.reset",
  "tickets.canned.check",
] as const;

function ticketRef(id: string) {
  const n = parseInt(id.replace(/\D/g, "").slice(-3), 10);
  return `TCK-${Number.isFinite(n) ? n : id.slice(0, 3).toUpperCase()}`;
}

function relativeTime(iso: string) {
  const diff = Date.now() - new Date(iso).getTime();
  const mins = Math.floor(diff / 60000);
  if (mins < 1) return "just now";
  if (mins < 60) return `${mins} mins ago`;
  const hrs = Math.floor(mins / 60);
  if (hrs < 24) return `${hrs}h ago`;
  return `${Math.floor(hrs / 24)}d ago`;
}

function statusType(status: string): "active" | "warning" | "inactive" {
  if (status === "open") return "active";
  if (status === "answered") return "warning";
  return "inactive";
}

export function Tickets() {
  useTitle("Tickets");
  const { t } = useI18n();
  const toast = useToast();
  const qc = useQueryClient();
  const [filter, setFilter] = useState<Filter>("");
  const [selectedId, setSelectedId] = useState<string | null>(null);
  const [reply, setReply] = useState("");

  const { data, isFetching, refetch } = useQuery({
    queryKey: ["admin-tickets", filter],
    queryFn: () => api<{ tickets: Ticket[]; total: number }>("/api/tickets", { query: { status: filter } }),
    refetchInterval: 15_000,
  });

  const { data: summary } = useQuery({
    queryKey: ["admin-tickets-summary"],
    queryFn: () => api<{ tickets: Ticket[] }>("/api/tickets"),
    refetchInterval: 15_000,
  });

  const tickets = data?.tickets ?? [];

  const counts = useMemo(() => {
    const all = summary?.tickets ?? [];
    return {
      open: all.filter((x) => x.status === "open").length,
      answered: all.filter((x) => x.status === "answered").length,
    };
  }, [summary]);

  useEffect(() => {
    if (tickets.length === 0) {
      setSelectedId(null);
      return;
    }
    if (!selectedId || !tickets.some((x) => x.id === selectedId)) {
      setSelectedId(tickets[0].id);
    }
  }, [tickets, selectedId]);

  const { data: detailData, isLoading: detailLoading } = useQuery({
    queryKey: ["admin-ticket-detail", selectedId],
    queryFn: () => api<{ ticket: Ticket & { messages: TicketMessage[] } }>(`/api/tickets/${selectedId}`),
    enabled: !!selectedId,
    refetchInterval: 10_000,
  });

  const ticket = detailData?.ticket;

  const replyMut = useMutation({
    mutationFn: (body: string) => api(`/api/tickets/${selectedId}/reply`, { method: "POST", body: { body } }),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["admin-ticket-detail", selectedId] });
      qc.invalidateQueries({ queryKey: ["admin-tickets"] });
      setReply("");
      toast.success(t("tickets.replySent"));
    },
  });

  const closeMut = useMutation({
    mutationFn: () => api(`/api/tickets/${selectedId}/close`, { method: "POST" }),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["admin-tickets"] });
      qc.invalidateQueries({ queryKey: ["admin-ticket-detail", selectedId] });
      toast.success(t("tickets.closed"));
    },
  });

  const filters: { id: Filter; label: string }[] = [
    { id: "", label: t("tickets.filterAll") },
    { id: "open", label: t("tickets.filterOpen") },
    { id: "answered", label: t("tickets.filterAnswered") },
    { id: "closed", label: t("tickets.filterClosed") },
  ];

  return (
    <div className="space-y-4 animate-page-enter">
      <div className="flex flex-col lg:flex-row lg:items-start justify-between gap-4">
        <div>
          <div className="flex flex-wrap items-center gap-2">
            <h1 className="text-2xl font-bold text-fg tracking-tight">{t("tickets.pageTitle")}</h1>
            <span className="inline-flex items-center gap-1">
              <Badge color="active">
                <RefreshCw size={11} className={cn(isFetching && "animate-spin")} />
                {t("tickets.liveReload")}
              </Badge>
            </span>
          </div>
          <p className="text-sm text-fg-muted mt-1 max-w-2xl">{t("tickets.pageSubtitle")}</p>
        </div>
        <div className="flex items-center gap-4 text-xs text-fg-muted">
          <span><strong className="text-danger">{t("tickets.openCount")}:</strong> {counts.open}</span>
          <span><strong className="text-warning">{t("tickets.answeredCount")}:</strong> {counts.answered}</span>
          <Button variant="ghost" size="sm" onClick={() => refetch()}>
            <RefreshCw size={14} />
          </Button>
        </div>
      </div>

      <div className="grid grid-cols-1 gap-4 lg:grid-cols-[minmax(280px,340px)_1fr] xl:grid-cols-[360px_1fr] min-h-[520px]">
        {/* Left: ticket list */}
        <GlassCard hover={false} className="!p-0 flex flex-col overflow-hidden">
          <div className="flex flex-wrap gap-1 p-3 border-b border-border/40 bg-surface-2/30">
            {filters.map((f) => (
              <button
                key={f.id || "all"}
                type="button"
                onClick={() => setFilter(f.id)}
                className={cn(
                  "px-2.5 py-1 rounded-lg text-[11px] font-semibold transition",
                  filter === f.id ? "bg-primary text-primary-fg" : "text-fg-muted hover:text-fg hover:bg-surface-2",
                )}
              >
                {f.label}
              </button>
            ))}
          </div>
          <div className="flex-1 overflow-y-auto divide-y divide-border/30">
            {tickets.map((tk) => (
              <button
                key={tk.id}
                type="button"
                onClick={() => setSelectedId(tk.id)}
                className={cn(
                  "w-full text-start px-4 py-3 transition hover:bg-surface-2/50",
                  selectedId === tk.id && "bg-primary/5 border-s-2 border-s-primary",
                )}
              >
                <div className="flex items-start justify-between gap-2 mb-1">
                  <span className="font-mono text-[11px] font-bold text-fg-subtle">{ticketRef(tk.id)}</span>
                  <StatusBadge
                    status={statusType(tk.status)}
                    label={tk.status.toUpperCase()}
                    pulse={tk.status === "open"}
                  />
                </div>
                <p className="text-sm font-medium text-fg line-clamp-2">{tk.subject}</p>
                <p className="text-[11px] text-fg-muted mt-1">
                  {tk.username || tk.user_id.slice(0, 8)} · {relativeTime(tk.updated_at)}
                </p>
              </button>
            ))}
            {tickets.length === 0 && (
              <p className="text-sm text-fg-muted text-center py-12 px-4">{t("tickets.empty")}</p>
            )}
          </div>
        </GlassCard>

        {/* Right: detail + chat */}
        <GlassCard hover={false} className="!p-0 flex flex-col overflow-hidden min-h-[480px]">
          {!selectedId ? (
            <div className="flex-1 flex items-center justify-center text-sm text-fg-muted p-8">
              {t("tickets.selectOne")}
            </div>
          ) : detailLoading && !ticket ? (
            <div className="flex-1 flex items-center justify-center text-sm text-fg-muted">{t("common.loading")}</div>
          ) : ticket ? (
            <>
              <div className="flex items-start justify-between gap-3 p-4 border-b border-border/40 bg-surface-2/20">
                <div className="min-w-0">
                  <div className="flex flex-wrap items-center gap-2">
                    <span className="font-mono text-xs font-bold text-primary">{ticketRef(ticket.id)}</span>
                    <StatusBadge status={statusType(ticket.status)} label={ticket.status.toUpperCase()} pulse={false} />
                  </div>
                  <h2 className="text-sm font-bold text-fg mt-1">{ticket.subject}</h2>
                  <p className="text-[11px] text-fg-muted mt-0.5">
                    {t("tickets.client")}: <span className="font-mono">{ticket.username || ticket.user_id.slice(0, 12)}</span>
                    {" · "}
                    {t("tickets.priority")}: <strong className="uppercase">{ticket.priority}</strong>
                  </p>
                </div>
                {ticket.status !== "closed" && (
                  <Button variant="outline" size="sm" onClick={() => closeMut.mutate()} disabled={closeMut.isPending}>
                    <X size={14} /> {t("tickets.closeTicket")}
                  </Button>
                )}
              </div>

              <div className="flex-1 overflow-y-auto p-4 space-y-3 bg-surface-1/30">
                {(ticket.messages ?? []).map((m) => {
                  const isAdmin = m.sender === "admin";
                  return (
                    <div key={m.id} className={cn("flex flex-col max-w-[85%]", isAdmin ? "ms-auto items-end" : "items-start")}>
                      <span className="text-[10px] text-fg-subtle mb-1 px-1">
                        {isAdmin ? t("tickets.supportEngineer") : (ticket.username || t("tickets.client"))}
                        {" · "}
                        {new Date(m.created_at).toLocaleString()}
                      </span>
                      <div
                        className={cn(
                          "rounded-2xl px-3.5 py-2.5 text-sm leading-relaxed whitespace-pre-wrap",
                          isAdmin ? "bg-primary/15 text-fg rounded-ee-sm" : "bg-surface-2 border border-border/40 text-fg rounded-es-sm",
                        )}
                      >
                        {m.body}
                      </div>
                    </div>
                  );
                })}
                {(ticket.messages ?? []).length === 0 && (
                  <p className="text-xs text-fg-muted text-center py-8">{t("tickets.noMessages")}</p>
                )}
              </div>

              {ticket.status !== "closed" && (
                <div className="border-t border-border/40 p-4 space-y-3 bg-bg-elevated">
                  <div>
                    <p className="text-[10px] font-semibold text-fg-subtle uppercase tracking-wide mb-1.5">
                      {t("tickets.cannedLabel")}
                    </p>
                    <div className="flex flex-wrap gap-1.5">
                      {CANNED_KEYS.map((key) => (
                        <button
                          key={key}
                          type="button"
                          onClick={() => setReply(t(key))}
                          className="text-[10px] px-2 py-1 rounded-md border border-border/60 bg-surface-2/50 hover:border-primary/40 hover:bg-primary/5 transition text-fg-muted hover:text-fg max-w-[200px] truncate"
                          title={t(key)}
                        >
                          {t(key)}
                        </button>
                      ))}
                    </div>
                  </div>
                  <textarea
                    value={reply}
                    onChange={(e) => setReply(e.target.value)}
                    placeholder={t("tickets.replyPlaceholder")}
                    rows={3}
                    className="field w-full resize-none text-sm"
                  />
                  <div className="flex justify-end">
                    <Button
                      onClick={() => reply.trim() && replyMut.mutate(reply.trim())}
                      disabled={replyMut.isPending || !reply.trim()}
                    >
                      <Send size={14} /> {t("tickets.sendReply")}
                    </Button>
                  </div>
                </div>
              )}
            </>
          ) : null}
        </GlassCard>
      </div>
    </div>
  );
}
