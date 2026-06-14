import { useState, useMemo } from "react";
import { useNavigate } from "react-router-dom";
import { useDeleteUser, useUsers } from "@/api/hooks";
import type { User } from "@/api/types";
import { Badge, Button, Card, Input, PageHeader, Select } from "@/components/ui";
import { Pagination } from "@/components/Pagination";
import { SortHeader, cycleSort, type SortDir } from "@/components/SortHeader";
import { useI18n } from "@/i18n/i18n";
import { QrCode, BarChart3, Pencil, Trash2, Layers } from "lucide-react";
import { CreateUserModal } from "@/components/CreateUserModal";
import { BulkCreateModal } from "@/components/BulkCreateModal";
import { EditUserModal } from "@/components/EditUserModal";
import { UserUsageModal } from "@/components/UserUsageModal";
import { UserSubModal } from "@/components/UserSubModal";
import { useConfirm } from "@/components/confirm";
import { useToast } from "@/components/toast";
import { formatBytes } from "@/lib/utils";

export function Users() {
  const [search, setSearch] = useState("");
  const [statusFilter, setStatusFilter] = useState("");
  const [page, setPage] = useState(0);
  const [pageSize, setPageSize] = useState(20);
  const [modalOpen, setModalOpen] = useState(false);
  const [bulkOpen, setBulkOpen] = useState(false);
  const [editing, setEditing] = useState<User | null>(null);
  const [viewing, setViewing] = useState<User | null>(null);
  const [subbing, setSubbing] = useState<User | null>(null);
  const [selected, setSelected] = useState<Set<string>>(new Set());
  const [sortKey, setSortKey] = useState<string | null>(null);
  const [sortDir, setSortDir] = useState<SortDir>(null);

  const { data, isLoading, error } = useUsers({ search, status: statusFilter, limit: pageSize, offset: page * pageSize });
  const total = data?.total ?? 0;

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
      let av: any, bv: any;
      switch (sortKey) {
        case "username": av = a.username.toLowerCase(); bv = b.username.toLowerCase(); break;
        case "status": av = a.status; bv = b.status; break;
        case "usage": av = a.used_traffic; bv = b.used_traffic; break;
        case "expires": av = a.expire_at ?? ""; bv = b.expire_at ?? ""; break;
        default: return 0;
      }
      if (av < bv) return sortDir === "asc" ? -1 : 1;
      if (av > bv) return sortDir === "asc" ? 1 : -1;
      return 0;
    });
  }, [data?.users, sortKey, sortDir]);

  function onSearch(v: string) {
    setSearch(v);
    setPage(0); // a new query resets to the first page
  }
  const del = useDeleteUser();
  const confirm = useConfirm();
  const toast = useToast();
  const { t } = useI18n();
  const nav = useNavigate();

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
    <div className="space-y-6">
      <CreateUserModal open={modalOpen} onClose={() => setModalOpen(false)} />
      <BulkCreateModal open={bulkOpen} onClose={() => setBulkOpen(false)} />
      <EditUserModal user={editing} onClose={() => setEditing(null)} />
      <UserUsageModal user={viewing} onClose={() => setViewing(null)} />
      <UserSubModal user={subbing} onClose={() => setSubbing(null)} />
      <PageHeader title={t("users.title")} subtitle={`${data?.total ?? 0} ${t("common.total")}`}>
        <Select className="w-32" value={statusFilter} onChange={(e) => { setStatusFilter(e.target.value); setPage(0); }}>
          <option value="">All</option>
          <option value="active">Active</option>
          <option value="limited">Limited</option>
          <option value="expired">Expired</option>
          <option value="disabled">Disabled</option>
          <option value="on_hold">On hold</option>
        </Select>
        <Input
          className="w-52"
          placeholder={t("common.search")}
          value={search}
          onChange={(e) => onSearch(e.target.value)}
        />
        <Button variant="outline" onClick={() => setBulkOpen(true)}><Layers size={15} /> {t("users.bulk")}</Button>
        <Button onClick={() => setModalOpen(true)}>{t("users.new")}</Button>
      </PageHeader>

      {selected.size > 0 && (
        <div className="card flex items-center gap-3 px-5 py-3">
          <span className="text-sm font-medium text-fg">{selected.size} selected</span>
          <div className="ms-auto flex gap-2">
            <Button variant="outline" size="sm" onClick={async () => {
              if (await confirm({ title: `Delete ${selected.size} users?`, confirmLabel: "Delete", destructive: true })) {
                for (const id of selected) await del.mutateAsync(id);
                setSelected(new Set());
                toast.success(`Deleted ${selected.size} users`);
              }
            }}>Delete selected</Button>
            <Button variant="ghost" size="sm" onClick={() => setSelected(new Set())}>Clear</Button>
          </div>
        </div>
      )}

      <Card className="p-0 overflow-x-auto">
        {isLoading && <div className="p-6 text-sm text-fg-muted">{t("common.loading")}</div>}
        {error && <div className="p-6 text-sm text-danger">Failed to load users</div>}
        {data && (
          <table className="w-full text-sm">
            <thead className="border-b text-start text-fg-muted">
              <tr>
                <th className="px-3 py-3 w-10">
                  <input type="checkbox" className="rounded" checked={selected.size === (data?.users.length ?? 0) && selected.size > 0} onChange={(e) => setSelected(e.target.checked ? new Set(data!.users.map(u => u.id)) : new Set())} />
                </th>
                <th className="px-5 py-3 text-start"><SortHeader label={t("users.username")} active={sortKey === "username"} dir={sortKey === "username" ? sortDir : null} onClick={() => toggleSort("username")} /></th>
                <th className="px-5 py-3 text-start"><SortHeader label={t("common.status")} active={sortKey === "status"} dir={sortKey === "status" ? sortDir : null} onClick={() => toggleSort("status")} /></th>
                <th className="px-5 py-3 text-start"><SortHeader label={t("users.usage")} active={sortKey === "usage"} dir={sortKey === "usage" ? sortDir : null} onClick={() => toggleSort("usage")} /></th>
                <th className="px-5 py-3 text-start"><SortHeader label={t("users.expires")} active={sortKey === "expires"} dir={sortKey === "expires" ? sortDir : null} onClick={() => toggleSort("expires")} /></th>
                <th className="px-5 py-3"></th>
              </tr>
            </thead>
            <tbody>
              {sortedUsers.map((u) => (
                <tr key={u.id} className="border-b last:border-0 hover:bg-muted/40">
                  <td className="px-3 py-3">
                    <input type="checkbox" className="rounded" checked={selected.has(u.id)} onChange={(e) => { const s = new Set(selected); e.target.checked ? s.add(u.id) : s.delete(u.id); setSelected(s); }} />
                  </td>
                  <td className="px-5 py-3 font-medium">
                    <button onClick={() => nav(`/users/${u.id}`)} className="text-fg hover:text-primary transition">{u.username}</button>
                  </td>
                  <td className="px-5 py-3">
                    <Badge color={u.status}>{u.status}</Badge>
                  </td>
                  <td className="px-5 py-3 text-muted-foreground">
                    {formatBytes(u.used_traffic)} / {formatBytes(u.data_limit)}
                  </td>
                  <td className="px-5 py-3 text-muted-foreground">
                    {u.expire_at ? new Date(u.expire_at).toLocaleDateString() : "Never"}
                  </td>
                  <td className="px-5 py-3">
                    <div className="flex items-center justify-end gap-0.5">
                      <Button variant="ghost" size="sm" onClick={() => setSubbing(u)} title="Subscription / QR">
                        <QrCode size={16} />
                      </Button>
                      <Button variant="ghost" size="sm" onClick={() => setViewing(u)} title={t("users.usage")}>
                        <BarChart3 size={16} />
                      </Button>
                      <Button variant="ghost" size="sm" onClick={() => setEditing(u)} title={t("common.edit")}>
                        <Pencil size={16} />
                      </Button>
                      <Button variant="ghost" size="sm" className="text-danger" onClick={() => remove(u)} title={t("common.delete")}>
                        <Trash2 size={16} />
                      </Button>
                    </div>
                  </td>
                </tr>
              ))}
              {sortedUsers.length === 0 && (
                <tr>
                  <td colSpan={6} className="px-5 py-8 text-center text-muted-foreground">
                    No users found
                  </td>
                </tr>
              )}
            </tbody>
          </table>
        )}
      </Card>

      <Pagination page={page} total={total} pageSize={pageSize} onPageChange={setPage} onPageSizeChange={(s) => { setPageSize(s); setPage(0); }} />
    </div>
  );
}
