import { useState, useMemo } from "react";
import { useNavigate } from "react-router-dom";
import { motion } from "framer-motion";
import {
  Users as UsersIcon,
  Search,
  Plus,
  Clock,
  QrCode,
  BarChart3,
  Pencil,
  Trash2,
  Layers,
  Download,
} from "lucide-react";
import { useBulkDeleteUsers, useDeleteUser, useUsers } from "@/api/hooks";
import { useOverview } from "@/api/policy-hooks";
import type { User, UserStatus } from "@/api/types";
import { Button, Select } from "@/components/ui";
import { Pagination } from "@/components/Pagination";
import { SortHeader, cycleSort, type SortDir } from "@/components/SortHeader";
import { GlassCard, StatsCard, StatusBadge } from "@/components/veltrix";
import { useI18n } from "@/i18n/i18n";
import { useTitle } from "@/lib/useTitle";
import { CreateUserModal } from "@/components/CreateUserModal";
import { BulkCreateModal } from "@/components/BulkCreateModal";
import { ImportUsersModal } from "@/components/ImportUsersModal";
import { EditUserModal } from "@/components/EditUserModal";
import { UserUsageModal } from "@/components/UserUsageModal";
import { UserSubModal } from "@/components/UserSubModal";
import { useConfirm } from "@/components/confirm";
import { useToast } from "@/components/toast";
import { useAuth } from "@/auth/auth";
import { cn, formatBytes } from "@/lib/utils";

const STATUS_FILTERS: { value: string; label: string }[] = [
  { value: "", label: "all" },
  { value: "active", label: "active" },
  { value: "limited", label: "limited" },
  { value: "expired", label: "expired" },
  { value: "disabled", label: "disabled" },
  { value: "on_hold", label: "on_hold" },
];

function statusFilterLabel(value: string, t: (key: import("@/i18n/dict").TKey) => string): string {
  switch (value) {
    case "":
      return t("users.filterAll");
    case "active":
      return t("users.statActive");
    case "limited":
      return t("users.statLimited");
    case "expired":
      return t("users.statExpired");
    case "disabled":
      return t("common.disabled");
    case "on_hold":
      return t("users.filterOnHold");
    default:
      return value;
  }
}

function statusBadgeType(status: UserStatus): string {
  if (status === "limited") return "warning";
  if (status === "expired") return "error";
  if (status === "disabled") return "inactive";
  if (status === "on_hold") return "info";
  return "active";
}

export function Users() {
  useTitle("Users");
  const { can } = useAuth();
  const canWrite = can("user:write");
  const [search, setSearch] = useState("");
  const [statusFilter, setStatusFilter] = useState("");
  const [page, setPage] = useState(0);
  const [pageSize, setPageSize] = useState(20);
  const [modalOpen, setModalOpen] = useState(false);
  const [bulkOpen, setBulkOpen] = useState(false);
  const [importOpen, setImportOpen] = useState(false);
  const [editing, setEditing] = useState<User | null>(null);
  const [viewing, setViewing] = useState<User | null>(null);
  const [subbing, setSubbing] = useState<User | null>(null);
  const [selected, setSelected] = useState<Set<string>>(new Set());
  const [sortKey, setSortKey] = useState<string | null>(null);
  const [sortDir, setSortDir] = useState<SortDir>(null);

  const { data, isLoading, error } = useUsers({
    search,
    status: statusFilter,
    limit: pageSize,
    offset: page * pageSize,
  });
  const overview = useOverview().data;
  const byStatus = overview?.users.by_status ?? {};
  const total = data?.total ?? 0;

  const { t } = useI18n();
  const nav = useNavigate();
  const del = useDeleteUser();
  const bulkDel = useBulkDeleteUsers();
  const confirm = useConfirm();
  const toast = useToast();

  function toggleSort(key: string) {
    if (sortKey === key) {
      const next = cycleSort(sortDir);
      setSortDir(next);
      if (next === null) setSortKey(null);
    } else {
      setSortKey(key);
      setSortDir("asc");
    }
  }

  const sortedUsers = useMemo(() => {
    const list = data?.users ?? [];
    if (!sortKey || !sortDir) return list;
    return [...list].sort((a, b) => {
      let av: string | number;
      let bv: string | number;
      switch (sortKey) {
        case "username":
          av = a.username.toLowerCase();
          bv = b.username.toLowerCase();
          break;
        case "status":
          av = a.status;
          bv = b.status;
          break;
        case "usage":
          av = a.used_traffic;
          bv = b.used_traffic;
          break;
        case "expires":
          av = a.expire_at ?? "";
          bv = b.expire_at ?? "";
          break;
        default:
          return 0;
      }
      if (av < bv) return sortDir === "asc" ? -1 : 1;
      if (av > bv) return sortDir === "asc" ? 1 : -1;
      return 0;
    });
  }, [data?.users, sortKey, sortDir]);

  function onSearch(v: string) {
    setSearch(v);
    setPage(0);
  }

  async function remove(u: User) {
    const ok = await confirm({
      title: `Delete ${u.username}?`,
      message: "This removes the user and revokes their access on all nodes.",
      confirmLabel: "Delete",
      destructive: true,
    });
    if (!ok) return;
    try {
      await del.mutateAsync(u.id);
      toast.success(`Deleted ${u.username}`);
    } catch {
      toast.error("Delete failed");
    }
  }

  return (
    <div className="space-y-6 animate-page-enter">
      <CreateUserModal open={modalOpen} onClose={() => setModalOpen(false)} />
      <BulkCreateModal open={bulkOpen} onClose={() => setBulkOpen(false)} />
      <ImportUsersModal open={importOpen} onClose={() => setImportOpen(false)} />
      <EditUserModal user={editing} onClose={() => setEditing(null)} />
      <UserUsageModal user={viewing} onClose={() => setViewing(null)} />
      <UserSubModal user={subbing} onClose={() => setSubbing(null)} />

      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
        <div>
          <h1 className="text-xl font-bold text-fg">{t("users.title")}</h1>
          <p className="text-sm text-fg-muted mt-0.5">
            {total} {t("common.total")}
            {statusFilter ? ` · ${statusFilter}` : ""}
          </p>
        </div>
        <div className="flex flex-wrap items-center gap-2">
          <Button variant="outline" onClick={() => setImportOpen(true)} disabled={!canWrite}>
            <Download size={15} /> {t("users.import")}
          </Button>
          <Button variant="outline" onClick={() => setBulkOpen(true)} disabled={!canWrite}>
            <Layers size={15} /> {t("users.bulk")}
          </Button>
          <Button onClick={() => setModalOpen(true)} disabled={!canWrite}>
            <Plus size={15} /> {t("users.new")}
          </Button>
        </div>
      </div>

      <div className="grid grid-cols-2 lg:grid-cols-4 gap-3">
        <StatsCard
          title={t("users.statTotal")}
          value={overview?.users.total ?? "—"}
          icon={<UsersIcon size={18} />}
          color="cyan"
          delay={0.05}
        />
        <StatsCard
          title={t("users.statActive")}
          value={byStatus.active ?? 0}
          icon={<UsersIcon size={18} />}
          color="green"
          delay={0.1}
        />
        <StatsCard
          title={t("users.statLimited")}
          value={byStatus.limited ?? 0}
          icon={<UsersIcon size={18} />}
          color="orange"
          delay={0.15}
        />
        <StatsCard
          title={t("users.statExpired")}
          value={byStatus.expired ?? 0}
          icon={<UsersIcon size={18} />}
          color="red"
          delay={0.2}
        />
      </div>

      <GlassCard hover={false} className="!p-3">
        <div className="flex flex-col sm:flex-row gap-3 items-start sm:items-center">
          <div className="relative flex-1 max-w-sm w-full">
            <Search size={14} className="absolute start-3 top-1/2 -translate-y-1/2 text-fg-subtle" />
            <input
              type="text"
              placeholder={t("common.search")}
              value={search}
              onChange={(e) => onSearch(e.target.value)}
              className="w-full h-8 rounded-lg bg-surface/80 border border-border/60 ps-8 pe-3 text-xs text-fg placeholder:text-fg-subtle focus:outline-none focus:border-primary/40 focus:ring-1 focus:ring-primary/20 transition-all input-surface"
            />
          </div>
          <div className="flex flex-wrap items-center gap-1 bg-surface/60 rounded-lg p-0.5 border border-border/40">
            {STATUS_FILTERS.map((f) => (
              <button
                key={f.value || "all"}
                type="button"
                onClick={() => {
                  setStatusFilter(f.value);
                  setPage(0);
                }}
                className={cn(
                  "px-3 py-1.5 rounded-md text-[11px] font-medium transition-all capitalize",
                  statusFilter === f.value
                    ? "bg-primary/15 text-primary"
                    : "text-fg-muted hover:text-fg",
                )}
              >
                {statusFilterLabel(f.value, t)}
              </button>
            ))}
          </div>
          <Select
            className="w-28 h-8 text-xs hidden lg:block"
            value={String(pageSize)}
            onChange={(e) => {
              setPageSize(Number(e.target.value));
              setPage(0);
            }}
          >
            {[10, 20, 50, 100].map((s) => (
              <option key={s} value={s}>
                {s}/page
              </option>
            ))}
          </Select>
        </div>
      </GlassCard>

      {canWrite && selected.size > 0 && (
        <GlassCard hover={false} className="!py-3 !px-4">
          <div className="flex items-center gap-3">
            <span className="text-sm font-medium text-fg">{selected.size} selected</span>
            <div className="ms-auto flex gap-2">
              <Button
                variant="outline"
                size="sm"
                onClick={async () => {
                  if (
                    await confirm({
                      title: `Delete ${selected.size} users?`,
                      confirmLabel: "Delete",
                      destructive: true,
                    })
                  ) {
                    try {
                      const res = await bulkDel.mutateAsync([...selected]);
                      setSelected(new Set());
                      toast.success(`Deleted ${res.deleted} users`);
                      if (res.failures.length > 0) toast.error(`${res.failures.length} deletions failed`);
                    } catch {
                      toast.error("Bulk delete failed");
                    }
                  }
                }}
              >
                Delete selected
              </Button>
              <Button variant="ghost" size="sm" onClick={() => setSelected(new Set())}>
                Clear
              </Button>
            </div>
          </div>
        </GlassCard>
      )}

      <GlassCard hover={false} className="!p-0 overflow-hidden">
        {isLoading && <div className="p-6 text-sm text-fg-muted">{t("common.loading")}</div>}
        {error && <div className="p-6 text-sm text-danger">Failed to load users</div>}
        {data && (
          <div className="overflow-x-auto">
            <table className="w-full text-xs">
              <thead>
                <tr className="border-b border-border/40">
                  <th className="py-3 px-3 w-10">
                    {canWrite && (
                      <input
                        type="checkbox"
                        className="rounded"
                        checked={selected.size === data.users.length && selected.size > 0}
                        onChange={(e) =>
                          setSelected(
                            e.target.checked ? new Set(data.users.map((u) => u.id)) : new Set(),
                          )
                        }
                      />
                    )}
                  </th>
                  <th className="text-start py-3 px-3 text-fg-subtle font-medium">
                    <SortHeader
                      label={t("users.username")}
                      active={sortKey === "username"}
                      dir={sortKey === "username" ? sortDir : null}
                      onClick={() => toggleSort("username")}
                    />
                  </th>
                  <th className="text-start py-3 px-3 text-fg-subtle font-medium hidden md:table-cell">
                    {t("users.devices")}
                  </th>
                  <th className="text-start py-3 px-3 text-fg-subtle font-medium">
                    <SortHeader
                      label={t("users.usage")}
                      active={sortKey === "usage"}
                      dir={sortKey === "usage" ? sortDir : null}
                      onClick={() => toggleSort("usage")}
                    />
                  </th>
                  <th className="text-start py-3 px-3 text-fg-subtle font-medium">
                    <SortHeader
                      label={t("common.status")}
                      active={sortKey === "status"}
                      dir={sortKey === "status" ? sortDir : null}
                      onClick={() => toggleSort("status")}
                    />
                  </th>
                  <th className="text-start py-3 px-3 text-fg-subtle font-medium hidden lg:table-cell">
                    <SortHeader
                      label={t("users.expires")}
                      active={sortKey === "expires"}
                      dir={sortKey === "expires" ? sortDir : null}
                      onClick={() => toggleSort("expires")}
                    />
                  </th>
                  <th className="text-end py-3 px-3 text-fg-subtle font-medium">{t("common.actions")}</th>
                </tr>
              </thead>
              <tbody>
                {sortedUsers.map((u, i) => {
                  const usagePct =
                    u.data_limit > 0 ? Math.min(100, (u.used_traffic / u.data_limit) * 100) : 0;
                  return (
                    <motion.tr
                      key={u.id}
                      initial={{ opacity: 0, y: 5 }}
                      animate={{ opacity: 1, y: 0 }}
                      transition={{ delay: i * 0.02 }}
                      className="border-b border-border/20 hover:bg-surface/40 transition-colors group"
                    >
                      <td className="py-3 px-3">
                        {canWrite && (
                          <input
                            type="checkbox"
                            className="rounded"
                            checked={selected.has(u.id)}
                            onChange={(e) => {
                              const s = new Set(selected);
                              if (e.target.checked) s.add(u.id);
                              else s.delete(u.id);
                              setSelected(s);
                            }}
                          />
                        )}
                      </td>
                      <td className="py-3 px-3">
                        <div className="flex items-center gap-2.5">
                          <div className="h-8 w-8 rounded-lg bg-surface-2 flex items-center justify-center text-fg-subtle flex-shrink-0">
                            {u.username.slice(0, 2).toUpperCase()}
                          </div>
                          <div className="min-w-0">
                            <button
                              type="button"
                              onClick={() => nav(`/users/${u.id}`)}
                              className="font-medium text-fg font-mono text-[11px] hover:text-primary transition truncate block max-w-[140px] sm:max-w-none"
                            >
                              {u.username}
                            </button>
                            {u.note && (
                              <p className="text-[10px] text-fg-subtle truncate max-w-[160px]">{u.note}</p>
                            )}
                          </div>
                        </div>
                      </td>
                      <td className="py-3 px-3 hidden md:table-cell text-fg-muted">
                        {u.device_limit > 0 ? u.device_limit : "∞"}
                      </td>
                      <td className="py-3 px-3">
                        <div className="space-y-1 w-28">
                          <div className="flex justify-between text-[10px]">
                            <span className="text-fg-muted">{formatBytes(u.used_traffic, false)}</span>
                            <span className="text-fg-subtle">
                              {u.data_limit > 0 ? formatBytes(u.data_limit, false) : "∞"}
                            </span>
                          </div>
                          {u.data_limit > 0 && (
                            <div className="h-1.5 rounded-full bg-surface-3 overflow-hidden">
                              <div
                                className={cn(
                                  "h-full rounded-full transition-all duration-500",
                                  usagePct > 90
                                    ? "bg-danger"
                                    : usagePct > 70
                                      ? "bg-warning"
                                      : "bg-primary",
                                )}
                                style={{ width: `${usagePct}%` }}
                              />
                            </div>
                          )}
                        </div>
                      </td>
                      <td className="py-3 px-3">
                        <StatusBadge
                          status={statusBadgeType(u.status)}
                          label={u.status}
                          pulse={u.status === "active"}
                        />
                      </td>
                      <td className="py-3 px-3 hidden lg:table-cell">
                        <div className="flex items-center gap-1 text-fg-muted">
                          <Clock size={11} />
                          {u.expire_at ? new Date(u.expire_at).toLocaleDateString() : "Never"}
                        </div>
                      </td>
                      <td className="py-3 px-3">
                        <div className="flex items-center justify-end gap-0.5 opacity-100 sm:opacity-0 sm:group-hover:opacity-100 transition-opacity">
                          <Button variant="ghost" size="sm" onClick={() => setSubbing(u)} title="Subscription / QR">
                            <QrCode size={14} />
                          </Button>
                          <Button variant="ghost" size="sm" onClick={() => setViewing(u)} title={t("users.usage")}>
                            <BarChart3 size={14} />
                          </Button>
                          {canWrite && (
                            <>
                              <Button variant="ghost" size="sm" onClick={() => setEditing(u)} title={t("common.edit")}>
                                <Pencil size={14} />
                              </Button>
                              <Button
                                variant="ghost"
                                size="sm"
                                className="text-danger"
                                onClick={() => remove(u)}
                                title={t("common.delete")}
                              >
                                <Trash2 size={14} />
                              </Button>
                            </>
                          )}
                        </div>
                      </td>
                    </motion.tr>
                  );
                })}
                {sortedUsers.length === 0 && (
                  <tr>
                    <td colSpan={7} className="px-5 py-10 text-center text-fg-muted">
                      No users found
                    </td>
                  </tr>
                )}
              </tbody>
            </table>
          </div>
        )}
        {data && total > pageSize && (
          <div className="border-t border-border/40 px-4 py-3">
            <Pagination
              page={page}
              total={total}
              pageSize={pageSize}
              onPageChange={setPage}
              onPageSizeChange={(s) => {
                setPageSize(s);
                setPage(0);
              }}
            />
          </div>
        )}
      </GlassCard>
    </div>
  );
}
