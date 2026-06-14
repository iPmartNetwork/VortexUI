import { cn } from "@/lib/utils";
import { ChevronLeft, ChevronRight } from "lucide-react";

interface Props {
  page: number;
  total: number;
  pageSize: number;
  onPageChange: (p: number) => void;
  onPageSizeChange?: (s: number) => void;
}

export function Pagination({ page, total, pageSize, onPageChange, onPageSizeChange }: Props) {
  const maxPage = Math.max(0, Math.ceil(total / pageSize) - 1);
  if (total <= pageSize) return null;

  const pages = buildPages(page, maxPage);
  const from = page * pageSize + 1;
  const to = Math.min((page + 1) * pageSize, total);

  return (
    <div className="flex flex-wrap items-center justify-between gap-3 text-sm text-fg-muted">
      <div className="flex items-center gap-2">
        <span>{from}–{to} of {total}</span>
        {onPageSizeChange && (
          <select
            value={pageSize}
            onChange={(e) => onPageSizeChange(Number(e.target.value))}
            className="rounded-lg border border-border bg-surface-2/50 px-2 py-1 text-xs outline-none"
          >
            {[10, 20, 50, 100].map((s) => <option key={s} value={s}>{s}/page</option>)}
          </select>
        )}
      </div>
      <div className="flex items-center gap-1">
        <PgBtn disabled={page === 0} onClick={() => onPageChange(page - 1)}><ChevronLeft size={14} className="rtl:rotate-180" /></PgBtn>
        {pages.map((p, i) =>
          p === "..." ? (
            <span key={`e${i}`} className="px-1 text-fg-subtle">…</span>
          ) : (
            <PgBtn key={p} active={p === page} onClick={() => onPageChange(p as number)}>
              {(p as number) + 1}
            </PgBtn>
          ),
        )}
        <PgBtn disabled={page >= maxPage} onClick={() => onPageChange(page + 1)}><ChevronRight size={14} className="rtl:rotate-180" /></PgBtn>
      </div>
    </div>
  );
}

function PgBtn({ children, active, disabled, onClick }: { children: React.ReactNode; active?: boolean; disabled?: boolean; onClick: () => void }) {
  return (
    <button
      onClick={onClick}
      disabled={disabled}
      className={cn(
        "grid h-8 min-w-[2rem] place-items-center rounded-lg text-xs font-medium transition",
        active ? "bg-primary/15 text-primary ring-1 ring-primary/30" : "text-fg-muted hover:bg-surface-2/60 hover:text-fg",
        disabled && "pointer-events-none opacity-40",
      )}
    >
      {children}
    </button>
  );
}

function buildPages(current: number, max: number): (number | "...")[] {
  if (max <= 6) return Array.from({ length: max + 1 }, (_, i) => i);
  const pages: (number | "...")[] = [0];
  if (current > 2) pages.push("...");
  for (let i = Math.max(1, current - 1); i <= Math.min(max - 1, current + 1); i++) pages.push(i);
  if (current < max - 2) pages.push("...");
  pages.push(max);
  return pages;
}
