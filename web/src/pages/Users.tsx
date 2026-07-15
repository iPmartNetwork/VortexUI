import { useEffect, useRef, useState } from "react";
import { useNavigate } from "react-router-dom";
import {
  User as UserIcon,
  Search,
  Plus,
  Clock,
  QrCode,
  BarChart3,
  Pencil,
  Trash2,
  Layers,
  Download,
  MoreVertical,
} from "lucide-react";
import { useDeleteUser, useUsers } from "@/api/hooks";
import type { User, UserStatus } from "@/api/types";
import { Button } from "@/components/ui";
import { Pagination } from "@/components/Pagination";
import { GlassCard, ProtocolBadge, StatusBadge } from "@/components/veltrix";
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
import { EmptyState } from "@/components/EmptyState";
import { Users as UsersIcon } from "lucide-react";

type StatusFilter = "" | "active" | "warning" | "inactive";

const STATUS_FILTERS: { value: StatusFilter; labelKey: "users.filterAll" | "users.statActive" | "users.filterWarning" | "users.filterInactive" }[] = [
  { value: "", labelKey: "users.filterAll" },
  { value: "active", labelKey: "users.statActive" },
  { value: "warning", labelKey: "users.filterWarning" },
  { value: "inactive", labelKey: "users.filterInactive" },
];

function statusDisplay(status: UserStatus): { type: string; label: string; pulse: boolean } {
  switch (status) {
    case "limited":
      return { type: "warning", label: "WARNING", pulse: false };
    case "expired":
      return { type: "error", label: "ERROR", pulse: false };
    case "disabled":
    case "on_hold":
      return { type: "inactive", label: "INACTIVE", pulse: false };
    default:
      return { type: "active", label: "ACTIVE", pulse: true };
  }
}

function userQueryParams(filter: StatusFilter): { status?: string; status_group?: string } {
  switch (filter) {
    case "active":
      return { status: "active" };
    case "warning":
      return { status_group: "warning" };
    case "inactive":
      return { status_group: "inactive" };
    default:
      return {};
  }
}

function devicesLabel(u: User): string {
  const connected = u.device_count ?? u.allowed_hwids?.length ?? 0;
  if (u.device_limit > 0) return `${connected}/${u.device_limit}`;
  return `${connected}/∞`;
}

function shortUserId(index: number): string {
  return `#${String(index + 1).padStart(4, "0")}`;
}

export function Users() {
  useTitle("Users");
  const { can } = useAuth();
  const canWrite = can("user:write");
  const [search, setSearch] = useState("");
  const [statusFilter, setStatusFilter] = useState<StatusFilter>("");
  const [page, setPage] = useState(0);
  const [pageSize] = useState(20);
  const [modalOpen, setModalOpen] = useState(false);
  const [bulkOpen, setBulkOpen] = useState(false);
  const [importOpen, setImportOpen] = useState(false);
  const [editing, setEditing] = useState<User | null>(null);
  const [viewing, setViewing] = useState<User | null>(null);
  const [subbing, setSubbing] = useState<User | null>(null);
  const [menuUserId, setMenuUserId] = useState<string | null>(null);
  const menuRef = useRef<HTMLDivElement>(null);

  const { data, isLoading, error } = useUsers({
    search,
    limit: pageSize,
    offset: page * pageSize,
    ...userQueryParams(statusFilter),
  });
  const users = data?.users ?? [];
  const total = data?.total ?? 0;
  const showing = users.length;

  const { t } = useI18n();
  const nav = useNavigate();
  const del = useDeleteUser();
  const confirm = useConfirm();
  const toast = useToast();

  useEffect(() => {
    function onDocClick(e: MouseEvent) {
      if (menuRef.current && !menuRef.current.contains(e.target as Node)) {
        setMenuUserId(null);
      }
    }
    document.addEventListener("mousedown", onDocClick);
    return () => document.removeEventListener("mousedown", onDocClick);
  }, []);

  function onSearch(v: string) {
    setSearch(v);
    setPage(0);
  }

  async function remove(u: User) {
    setMenuUserId(null);
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

  const summaryLabel = t("users.showingOf")
    .replace("{count}", String(showing))
    .replace("{total}", String(total));

  return (
    <div className="space-y-5 animate-page-enter">
      <CreateUserModal open={modalOpen} onClose={() => setModalOpen(false)} />
      <BulkCreateModal open={bulkOpen} onClose={() => setBulkOpen(false)} />
      <ImportUsersModal open={importOpen} onClose={() => setImportOpen(false)} />
      <EditUserModal user={editing} onClose={() => setEditing(null)} />
      <UserUsageModal user={viewing} onClose={() => setViewing(null)} />
      <UserSubModal user={subbing} onClose={() => setSubbing(null)} />

      <div className="flex flex-col sm:flex-row sm:items-start justify-between gap-4">
        <div>
          <h1 className="text-2xl font-bold text-fg tracking-tight">{t("users.managementTitle")}</h1>
          <p className="text-sm text-fg-muted mt-1">
            {total} {t("users.totalUsers")}
          </p>
        </div>
        <div className="flex flex-wrap items-center gap-2">
          {canWrite && (
            <>
              <Button variant="outline" size="sm" className="hidden sm:inline-flex" onClick={() => setImportOpen(true)}>
                <Download size={14} /> {t("users.import")}
              </Button>
              <Button variant="outline" size="sm" className="hidden sm:inline-flex" onClick={() => setBulkOpen(true)}>
                <Layers size={14} /> {t("users.bulk")}
              </Button>
            </>
          )}
          <Button onClick={() => setModalOpen(true)} disabled={!canWrite}>
            <Plus size={15} /> {t("users.new")}
          </Button>
        </div>
      </div>

      <GlassCard hover={false} className="!p-0 overflow-hidden">
        <div className="flex flex-col lg:flex-row lg:items-center justify-between gap-3 p-4 border-b border-border/40">
          <div className="relative flex-1 max-w-md">
            <Search size={14} className="absolute start-3 top-1/2 -translate-y-1/2 text-fg-subtle" />
            <input
              type="text"
              placeholder={`${t("common.search")} users...`}
              value={search}
              onChange={(e) => onSearch(e.target.value)}
              className="w-full h-9 rounded-lg bg-surface/80 border border-border/60 ps-9 pe-3 text-sm text-fg placeholder:text-fg-subtle focus:outline-none focus:border-primary/40 focus:ring-1 focus:ring-primary/20 transition-all input-surface"
            />
          </div>
          <div className="flex flex-wrap items-center gap-1">
            {STATUS_FILTERS.map((f) => (
              <button
                key={f.value || "all"}
                type="button"
                onClick={() => {
                  setStatusFilter(f.value);
                  setPage(0);
                }}
                className={cn(
                  "px-3.5 py-1.5 rounded-lg text-xs font-medium transition-all",
                  statusFilter === f.value
                    ? "bg-primary text-primary-fg shadow-sm"
                    : "text-fg-muted hover:text-fg hover:bg-surface/60",
                )}
              >
                {t(f.labelKey)}
              </button>
            ))}
          </div>
        </div>

        {isLoading && <div className="p-8 text-sm text-fg-muted text-center">{t("common.loading")}</div>}
        {error && <div className="p-8 text-sm text-danger text-center">Failed to load users</div>}

        {data && (
          <div className="overflow-x-auto">
            <table className="w-full text-sm">
              <thead>
                <tr className="border-b border-border/40 bg-surface/30">
                  <th className="text-start py-3 px-4 text-xs font-semibold text-fg-subtle uppercase tracking-wide">
                    {t("users.username")}
                  </th>
                  <th className="text-start py-3 px-4 text-xs font-semibold text-fg-subtle uppercase tracking-wide hidden md:table-cell">
                    {t("users.protocol")}
                  </th>
                  <th className="text-start py-3 px-4 text-xs font-semibold text-fg-subtle uppercase tracking-wide">
                    {t("users.dataUsage")}
                  </th>
                  <th className="text-start py-3 px-4 text-xs font-semibold text-fg-subtle uppercase tracking-wide hidden sm:table-cell">
                    {t("users.devices")}
                  </th>
                  <th className="text-start py-3 px-4 text-xs font-semibold text-fg-subtle uppercase tracking-wide">
                    {t("common.status")}
                  </th>
                  <th className="text-start py-3 px-4 text-xs font-semibold text-fg-subtle uppercase tracking-wide hidden lg:table-cell">
                    {t("users.expires")}
                  </th>
                  <th className="text-end py-3 px-4 text-xs font-semibold text-fg-subtle uppercase tracking-wide w-12">
                    {t("common.actions")}
                  </th>
                </tr>
              </thead>
              <tbody>
                {users.map((u, i) => {
                  const usagePct =
                    u.data_limit > 0 ? Math.min(100, (u.used_traffic / u.data_limit) * 100) : 0;
                  const st = statusDisplay(u.status);
                  const globalIndex = page * pageSize + i;
                  return (
                    <tr
                      key={u.id}
                      className="border-b border-border/20 hover:bg-surface/40 transition-colors"
                    >
                      <td className="py-3.5 px-4">
                        <div className="flex items-center gap-3 min-w-[160px]">
                          <div className="h-9 w-9 rounded-full bg-primary/10 flex items-center justify-center text-primary flex-shrink-0">
                            <UserIcon size={16} />
                          </div>
                          <div className="min-w-0">
                            <button
                              type="button"
                              onClick={() => nav(`/users/${u.id}`)}
                              className="font-semibold text-fg text-sm hover:text-primary transition truncate block max-w-[180px]"
                            >
                              {u.username}
                            </button>
                            <p className="text-[11px] text-fg-subtle font-mono">
                              ID: {shortUserId(globalIndex)}
                            </p>
                          </div>
                        </div>
                      </td>
                      <td className="py-3.5 px-4 hidden md:table-cell">
                        <ProtocolBadge label={u.protocol_label ?? "—"} />
                      </td>
                      <td className="py-3.5 px-4">
                        <div className="space-y-1.5 min-w-[140px] max-w-[200px]">
                          <div className="text-xs text-fg-muted whitespace-nowrap">
                            {formatBytes(u.used_traffic, false)} /{" "}
                            {u.data_limit > 0 ? formatBytes(u.data_limit, false) : "∞"}
                          </div>
                          {u.data_limit > 0 && (
                            <div className="h-1.5 rounded-full bg-surface-3 overflow-hidden">
                              <div
                                className={cn(
                                  "h-full rounded-full transition-all duration-500",
                                  usagePct > 90 || u.status === "limited"
                                    ? "bg-warning"
                                    : usagePct > 70
                                      ? "bg-warning/80"
                                      : "bg-primary",
                                )}
                                style={{ width: `${usagePct}%` }}
                              />
                            </div>
                          )}
                        </div>
                      </td>
                      <td className="py-3.5 px-4 hidden sm:table-cell text-fg-muted tabular-nums">
                        {devicesLabel(u)}
                      </td>
                      <td className="py-3.5 px-4">
                        <StatusBadge status={st.type} label={st.label} pulse={st.pulse} />
                      </td>
                      <td className="py-3.5 px-4 hidden lg:table-cell">
                        <div className="flex items-center gap-1.5 text-fg-muted text-xs whitespace-nowrap">
                          <Clock size={12} className="text-fg-subtle" />
                          {u.expire_at ? new Date(u.expire_at).toLocaleDateString() : "Never"}
                        </div>
                      </td>
                      <td className="py-3.5 px-4 text-end">
                        <div className="relative inline-block" ref={menuUserId === u.id ? menuRef : undefined}>
                          <button
                            type="button"
                            onClick={() => setMenuUserId(menuUserId === u.id ? null : u.id)}
                            className="p-1.5 rounded-lg text-fg-muted hover:text-fg hover:bg-surface-2/80 transition"
                            aria-label={t("common.actions")}
                          >
                            <MoreVertical size={16} />
                          </button>
                          {menuUserId === u.id && (
                            <div className="absolute end-0 top-full mt-1 z-20 min-w-[160px] rounded-lg border border-border/60 bg-surface shadow-lg py-1 text-xs">
                              <MenuAction icon={<QrCode size={14} />} onClick={() => { setSubbing(u); setMenuUserId(null); }}>
                                {t("users.subscription")}
                              </MenuAction>
                              <MenuAction icon={<BarChart3 size={14} />} onClick={() => { setViewing(u); setMenuUserId(null); }}>
                                {t("users.usage")}
                              </MenuAction>
                              {canWrite && (
                                <>
                                  <MenuAction icon={<Pencil size={14} />} onClick={() => { setEditing(u); setMenuUserId(null); }}>
                                    {t("common.edit")}
                                  </MenuAction>
                                  <MenuAction
                                    icon={<Trash2 size={14} />}
                                    className="text-danger hover:bg-danger/10"
                                    onClick={() => remove(u)}
                                  >
                                    {t("common.delete")}
                                  </MenuAction>
                                </>
                              )}
                            </div>
                          )}
                        </div>
                      </td>
                    </tr>
                  );
                })}
                {users.length === 0 && (
                  <tr>
                    <td colSpan={7} className="px-5 py-8">
                      <EmptyState
                        icon={UsersIcon}
                        title={search ? "No users match your search" : t("users.none")}
                        description={search ? "Try a different name or clear the filter." : t("users.totalUsers")}
                        compact
                      />
                    </td>
                  </tr>
                )}
              </tbody>
            </table>
          </div>
        )}

        {data && total > 0 && (
          <div className="border-t border-border/40 px-4 py-3">
            <Pagination
              page={page}
              total={total}
              pageSize={pageSize}
              onPageChange={setPage}
              summaryLabel={summaryLabel}
              alwaysShow
            />
          </div>
        )}
      </GlassCard>
    </div>
  );
}

function MenuAction({
  children,
  icon,
  onClick,
  className,
}: {
  children: React.ReactNode;
  icon: React.ReactNode;
  onClick: () => void;
  className?: string;
}) {
  return (
    <button
      type="button"
      onClick={onClick}
      className={cn(
        "w-full flex items-center gap-2 px-3 py-2 text-start text-fg hover:bg-surface-2/60 transition",
        className,
      )}
    >
      {icon}
      {children}
    </button>
  );
}
