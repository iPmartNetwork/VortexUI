import { useState, useMemo } from "react";
import { cn } from "@/lib/utils";
import { ChevronDown, ChevronUp, ChevronsUpDown, ChevronLeft, ChevronRight } from "lucide-react";
import { Input, Button } from "@/components/ui";

export interface Column<T> {
  key: string;
  label: string;
  sortable?: boolean;
  render?: (row: T, index: number) => React.ReactNode;
  className?: string;
}

interface DataTableProps<T> {
  data: T[];
  columns: Column<T>[];
  pageSize?: number;
  searchable?: boolean;
  searchPlaceholder?: string;
  onRowClick?: (row: T) => void;
  emptyMessage?: string;
  className?: string;
}

type SortDir = "asc" | "desc" | null;

export function DataTable<T extends Record<string, any>>({
  data,
  columns,
  pageSize = 15,
  searchable = true,
  searchPlaceholder = "Search...",
  onRowClick,
  emptyMessage = "No data found.",
  className,
}: DataTableProps<T>) {
  const [search, setSearch] = useState("");
  const [sortKey, setSortKey] = useState<string | null>(null);
  const [sortDir, setSortDir] = useState<SortDir>(null);
  const [page, setPage] = useState(0);

  // Filter
  const filtered = useMemo(() => {
    if (!search.trim()) return data;
    const q = search.toLowerCase();
    return data.filter(row =>
      columns.some(col => {
        const val = row[col.key];
        return val != null && String(val).toLowerCase().includes(q);
      })
    );
  }, [data, search, columns]);

  // Sort
  const sorted = useMemo(() => {
    if (!sortKey || !sortDir) return filtered;
    return [...filtered].sort((a, b) => {
      const av = a[sortKey] ?? "";
      const bv = b[sortKey] ?? "";
      if (typeof av === "number" && typeof bv === "number") {
        return sortDir === "asc" ? av - bv : bv - av;
      }
      const cmp = String(av).localeCompare(String(bv));
      return sortDir === "asc" ? cmp : -cmp;
    });
  }, [filtered, sortKey, sortDir]);

  // Paginate
  const totalPages = Math.max(1, Math.ceil(sorted.length / pageSize));
  const paged = sorted.slice(page * pageSize, (page + 1) * pageSize);

  function handleSort(key: string) {
    if (sortKey === key) {
      if (sortDir === "asc") setSortDir("desc");
      else if (sortDir === "desc") { setSortKey(null); setSortDir(null); }
    } else {
      setSortKey(key);
      setSortDir("asc");
    }
    setPage(0);
  }

  function SortIcon({ col }: { col: string }) {
    if (sortKey !== col) return <ChevronsUpDown size={12} className="text-fg-subtle/50" />;
    if (sortDir === "asc") return <ChevronUp size={12} className="text-primary" />;
    return <ChevronDown size={12} className="text-primary" />;
  }

  return (
    <div className={cn("card p-0 overflow-hidden", className)}>
      {/* Toolbar */}
      {searchable && (
        <div className="flex items-center justify-between gap-3 border-b border-border/40 px-4 py-3">
          <Input
            placeholder={searchPlaceholder}
            value={search}
            onChange={e => { setSearch(e.target.value); setPage(0); }}
            className="max-w-xs"
          />
          <span className="text-xs text-fg-subtle">{filtered.length} results</span>
        </div>
      )}

      {/* Table */}
      <div className="overflow-x-auto">
        <table className="w-full text-sm">
          <thead>
            <tr className="border-b border-border/40 bg-surface-2/20">
              {columns.map(col => (
                <th
                  key={col.key}
                  className={cn(
                    "px-4 py-2.5 text-left text-xs font-semibold uppercase tracking-wider text-fg-subtle",
                    col.sortable && "cursor-pointer select-none hover:text-fg",
                    col.className,
                  )}
                  onClick={col.sortable ? () => handleSort(col.key) : undefined}
                >
                  <span className="inline-flex items-center gap-1">
                    {col.label}
                    {col.sortable && <SortIcon col={col.key} />}
                  </span>
                </th>
              ))}
            </tr>
          </thead>
          <tbody>
            {paged.map((row, i) => (
              <tr
                key={i}
                onClick={onRowClick ? () => onRowClick(row) : undefined}
                className={cn(
                  "border-b border-border/20 transition",
                  onRowClick && "cursor-pointer hover:bg-surface-2/40",
                )}
              >
                {columns.map(col => (
                  <td key={col.key} className={cn("px-4 py-2.5", col.className)}>
                    {col.render ? col.render(row, i) : row[col.key]}
                  </td>
                ))}
              </tr>
            ))}
            {paged.length === 0 && (
              <tr>
                <td colSpan={columns.length} className="px-4 py-8 text-center text-sm text-fg-muted">
                  {emptyMessage}
                </td>
              </tr>
            )}
          </tbody>
        </table>
      </div>

      {/* Pagination */}
      {totalPages > 1 && (
        <div className="flex items-center justify-between border-t border-border/40 px-4 py-2.5">
          <span className="text-xs text-fg-subtle">
            Page {page + 1} of {totalPages}
          </span>
          <div className="flex items-center gap-1">
            <Button
              variant="ghost"
              size="sm"
              disabled={page === 0}
              onClick={() => setPage(p => p - 1)}
            >
              <ChevronLeft size={14} />
            </Button>
            <Button
              variant="ghost"
              size="sm"
              disabled={page >= totalPages - 1}
              onClick={() => setPage(p => p + 1)}
            >
              <ChevronRight size={14} />
            </Button>
          </div>
        </div>
      )}
    </div>
  );
}
