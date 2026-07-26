import { cn } from "@/lib/utils";

// ════════════════════════════════════════════════════════════════
// BASE — Glass-styled shimmer block
// ════════════════════════════════════════════════════════════════
export function Skeleton({ className }: { className?: string }) {
  return (
    <div
      className={cn(
        "animate-shimmer rounded-xl bg-gradient-to-r from-surface-2/20 via-surface-2/50 to-surface-2/20 bg-[length:200%_100%]",
        className,
      )}
      aria-hidden="true"
    />
  );
}

/** Line of text — mimics a real text line height */
export function SkeletonLine({ width = "w-full", className }: { width?: string; className?: string }) {
  return <Skeleton className={cn("h-3.5 rounded-md", width, className)} />;
}

/** Short line — for labels, subtitles */
export function SkeletonLabel({ className }: { className?: string }) {
  return <Skeleton className={cn("h-2.5 w-16 rounded-md", className)} />;
}

/** Circle placeholder — for avatars, icons */
export function SkeletonCircle({ size = 10 }: { size?: number }) {
  return (
    <div
      className="animate-shimmer rounded-full bg-gradient-to-r from-surface-2/20 via-surface-2/50 to-surface-2/20 bg-[length:200%_100%] flex-shrink-0"
      style={{ width: size * 4, height: size * 4 }}
      aria-hidden="true"
    />
  );
}

/** Icon box placeholder — square with rounded corners */
export function SkeletonIcon({ className }: { className?: string }) {
  return <Skeleton className={cn("h-10 w-10 rounded-xl", className)} />;
}

// ════════════════════════════════════════════════════════════════
// GLASS CARD SKELETON
// ════════════════════════════════════════════════════════════════
export function SkeletonCard({ className }: { className?: string }) {
  return (
    <div
      className={cn(
        "rounded-2xl bg-bg-elevated border border-border p-5 space-y-3",
        className,
      )}
      aria-hidden="true"
    >
      <Skeleton className="h-3.5 w-1/3" />
      <Skeleton className="h-7 w-2/3" />
      <Skeleton className="h-3 w-1/2" />
    </div>
  );
}

/** Stat card skeleton — mimics the StatsCard component layout */
export function SkeletonStatCard({ className }: { className?: string }) {
  return (
    <div
      className={cn(
        "rounded-2xl bg-bg-elevated border border-border p-5",
        className,
      )}
      aria-hidden="true"
    >
      <div className="flex items-start justify-between gap-3">
        <div className="min-w-0 flex-1 space-y-2">
          <Skeleton className="h-2.5 w-14 rounded-md" />
          <Skeleton className="h-7 w-24 rounded-lg" />
          <Skeleton className="h-2.5 w-20 rounded-md" />
        </div>
        <Skeleton className="h-10 w-10 rounded-xl flex-shrink-0" />
      </div>
    </div>
  );
}

/** Badge/pill skeleton */
export function SkeletonBadge({ className }: { className?: string }) {
  return <Skeleton className={cn("h-5 w-14 rounded-full", className)} />;
}

/** Button skeleton */
export function SkeletonButton({ className }: { className?: string }) {
  return <Skeleton className={cn("h-9 w-24 rounded-xl", className)} />;
}

/** Input field skeleton */
export function SkeletonInput({ className }: { className?: string }) {
  return <Skeleton className={cn("h-10 w-full rounded-xl", className)} />;
}

// ════════════════════════════════════════════════════════════════
// TABLE SKELETONS
// ════════════════════════════════════════════════════════════════
export function SkeletonRow({ cols = 5 }: { cols?: number }) {
  return (
    <div className="flex items-center gap-4 py-3.5 border-b border-border/20">
      {Array.from({ length: cols }).map((_, i) => (
        <Skeleton
          key={i}
          className={cn("h-4 rounded-md", i === 0 ? "w-32" : "flex-1")}
        />
      ))}
    </div>
  );
}

export function SkeletonTable({
  rows = 6,
  cols = 5,
}: {
  rows?: number;
  cols?: number;
}) {
  return (
    <div className="rounded-2xl bg-bg-elevated border border-border p-5 space-y-1">
      {/* Header row */}
      <div className="flex items-center gap-4 pb-3 border-b border-border/40">
        {Array.from({ length: cols }).map((_, i) => (
          <Skeleton
            key={i}
            className={cn("h-3 rounded-md", i === 0 ? "w-24" : "flex-1")}
          />
        ))}
      </div>
      {/* Body rows */}
      {Array.from({ length: rows }).map((_, i) => (
        <SkeletonRow key={i} cols={cols} />
      ))}
    </div>
  );
}

// ════════════════════════════════════════════════════════════════
// CHART SKELETONS
// ════════════════════════════════════════════════════════════════
export function SkeletonChart({ className }: { className?: string }) {
  return (
    <div className={cn("rounded-2xl bg-bg-elevated border border-border p-5 space-y-4", className)} aria-hidden="true">
      <div className="flex items-center justify-between">
        <Skeleton className="h-4 w-32 rounded-md" />
        <Skeleton className="h-4 w-12 rounded-md" />
      </div>
      <Skeleton className="h-44 w-full rounded-xl" />
    </div>
  );
}

/** Donut chart skeleton */
export function SkeletonDonut({ className }: { className?: string }) {
  return (
    <div className={cn("rounded-2xl bg-bg-elevated border border-border p-5 space-y-4", className)} aria-hidden="true">
      <Skeleton className="h-4 w-28 rounded-md" />
      <div className="flex items-center justify-center py-4">
        <div className="relative h-32 w-32">
          {/* Circular skeleton */}
          <div className="absolute inset-0 animate-shimmer rounded-full bg-gradient-to-r from-surface-2/20 via-surface-2/50 to-surface-2/20 bg-[length:200%_100%]" />
          {/* Center hole illusion */}
          <div className="absolute inset-4 rounded-full bg-bg" />
        </div>
      </div>
      {/* Legend items */}
      <div className="space-y-2">
        {Array.from({ length: 3 }).map((_, i) => (
          <div key={i} className="flex items-center gap-2">
            <Skeleton className="h-2.5 w-2.5 rounded-full" />
            <Skeleton className="h-3 w-20 rounded-md" />
            <Skeleton className="h-3 w-8 rounded-md ms-auto" />
          </div>
        ))}
      </div>
    </div>
  );
}

// ════════════════════════════════════════════════════════════════
// GRID / LIST SKELETONS
// ════════════════════════════════════════════════════════════════
export function SkeletonGrid({ count = 6 }: { count?: number }) {
  return (
    <div className="grid grid-cols-1 gap-4 md:grid-cols-2 lg:grid-cols-3">
      {Array.from({ length: count }).map((_, i) => (
        <SkeletonCard key={i} />
      ))}
    </div>
  );
}

/** Filter/search bar skeleton */
export function SkeletonFilterBar({ className }: { className?: string }) {
  return (
    <div className={cn("flex items-center gap-3 flex-wrap", className)} aria-hidden="true">
      <Skeleton className="h-10 w-56 rounded-xl" />
      <Skeleton className="h-10 w-32 rounded-xl" />
      <Skeleton className="h-10 w-10 rounded-xl" />
      <Skeleton className="h-10 w-10 rounded-xl ms-auto" />
    </div>
  );
}

// ════════════════════════════════════════════════════════════════
// 🔷 PAGE-TYPE SKELETONS
// ════════════════════════════════════════════════════════════════

/** Standard page — header + 4 stat cards + table */
export function SkeletonPage() {
  return (
    <div className="space-y-6 animate-fade-in" aria-label="Loading page content">
      {/* Page header */}
      <div className="space-y-2">
        <Skeleton className="h-7 w-48 rounded-lg" />
        <Skeleton className="h-4 w-72 rounded-md" />
      </div>
      {/* Stat cards */}
      <div className="grid grid-cols-2 gap-4 lg:grid-cols-4">
        <SkeletonStatCard />
        <SkeletonStatCard />
        <SkeletonStatCard />
        <SkeletonStatCard />
      </div>
      {/* Table */}
      <SkeletonTable />
    </div>
  );
}

/** List page — header + filter bar + table (for Users, Nodes, Inbounds, etc.) */
export function SkeletonListPage() {
  return (
    <div className="space-y-6 animate-fade-in" aria-label="Loading list page">
      {/* Header with actions */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
        <div className="space-y-2">
          <Skeleton className="h-7 w-40 rounded-lg" />
          <Skeleton className="h-4 w-60 rounded-md" />
        </div>
        <SkeletonButton />
      </div>
      {/* Filter/search bar */}
      <SkeletonFilterBar />
      {/* Table */}
      <SkeletonTable rows={8} />
      {/* Pagination */}
      <div className="flex items-center justify-between">
        <Skeleton className="h-4 w-32 rounded-md" />
        <div className="flex items-center gap-2">
          <Skeleton className="h-8 w-8 rounded-lg" />
          <Skeleton className="h-8 w-8 rounded-lg" />
          <Skeleton className="h-8 w-8 rounded-lg" />
        </div>
      </div>
    </div>
  );
}

/** Detail page — header with back button + profile card + detail grid (for UserDetail, etc.) */
export function SkeletonDetailPage() {
  return (
    <div className="space-y-6 animate-fade-in" aria-label="Loading detail page">
      {/* Back button + header */}
      <div className="flex items-center gap-3">
        <Skeleton className="h-8 w-8 rounded-xl" />
        <div className="space-y-1.5">
          <Skeleton className="h-6 w-48 rounded-lg" />
          <Skeleton className="h-3.5 w-32 rounded-md" />
        </div>
        <div className="ms-auto flex gap-2">
          <Skeleton className="h-9 w-20 rounded-xl" />
          <Skeleton className="h-9 w-20 rounded-xl" />
        </div>
      </div>
      {/* Profile card */}
      <div className="rounded-2xl bg-bg-elevated border border-border p-6 space-y-4">
        <div className="flex items-center gap-4">
          <SkeletonCircle size={14} />
          <div className="space-y-2">
            <Skeleton className="h-5 w-40 rounded-lg" />
            <Skeleton className="h-3.5 w-56 rounded-md" />
          </div>
          <SkeletonBadge className="ms-auto" />
        </div>
        <div className="h-px bg-border/40" />
        {/* Info grid */}
        <div className="grid grid-cols-2 md:grid-cols-3 gap-4">
          {Array.from({ length: 6 }).map((_, i) => (
            <div key={i} className="space-y-1.5">
              <Skeleton className="h-2.5 w-16 rounded-md" />
              <Skeleton className="h-4 w-28 rounded-md" />
            </div>
          ))}
        </div>
      </div>
      {/* Tabs */}
      <div className="flex gap-1 rounded-xl bg-surface-2/50 p-1">
        {Array.from({ length: 4 }).map((_, i) => (
          <Skeleton key={i} className="h-8 flex-1 rounded-lg" />
        ))}
      </div>
      {/* Tab content placeholder */}
      <SkeletonTable rows={4} cols={4} />
    </div>
  );
}

/** Dashboard / Overview page — stat cards row + charts row + recent activity */
export function SkeletonDashboardPage() {
  return (
    <div className="space-y-6 animate-fade-in" aria-label="Loading dashboard">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div className="space-y-2">
          <Skeleton className="h-7 w-52 rounded-lg" />
          <Skeleton className="h-4 w-64 rounded-md" />
        </div>
        <Skeleton className="h-9 w-28 rounded-xl" />
      </div>
      {/* Live status bar */}
      <div className="flex items-center gap-2">
        <SkeletonBadge />
        <Skeleton className="h-3 w-24 rounded-md" />
      </div>
      {/* Stat cards row */}
      <div className="grid grid-cols-2 gap-4 lg:grid-cols-4">
        {Array.from({ length: 4 }).map((_, i) => (
          <SkeletonStatCard key={i} />
        ))}
      </div>
      {/* Main chart */}
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-4">
        <SkeletonChart className="lg:col-span-2" />
        <div className="space-y-4">
          <SkeletonDonut />
        </div>
      </div>
      {/* Recent activity */}
      <div className="rounded-2xl bg-bg-elevated border border-border p-5 space-y-3">
        <div className="flex items-center justify-between">
          <Skeleton className="h-4 w-28 rounded-md" />
          <Skeleton className="h-4 w-16 rounded-md" />
        </div>
        {Array.from({ length: 4 }).map((_, i) => (
          <div key={i} className="flex items-center gap-3 py-2">
            <Skeleton className="h-2 w-2 rounded-full" />
            <Skeleton className="h-3.5 w-48 rounded-md" />
            <Skeleton className="h-3 w-16 rounded-md ms-auto" />
          </div>
        ))}
      </div>
    </div>
  );
}

/** Analytics page — date range selector + stat cards + large chart + small charts */
export function SkeletonAnalyticsPage() {
  return (
    <div className="space-y-6 animate-fade-in" aria-label="Loading analytics">
      {/* Header */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
        <div className="space-y-2">
          <Skeleton className="h-7 w-44 rounded-lg" />
          <Skeleton className="h-4 w-56 rounded-md" />
        </div>
        <Skeleton className="h-9 w-64 rounded-xl" />
      </div>
      {/* Stat cards */}
      <div className="grid grid-cols-2 gap-4 lg:grid-cols-5">
        {Array.from({ length: 5 }).map((_, i) => (
          <SkeletonStatCard key={i} />
        ))}
      </div>
      {/* Large chart */}
      <SkeletonChart />
      {/* Small charts row */}
      <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
        <SkeletonChart />
        <SkeletonChart />
      </div>
    </div>
  );
}

/** Portal page skeleton — header + stat cards */
export function SkeletonPortalPage() {
  return (
    <div className="space-y-6 animate-fade-in" aria-label="Loading portal page">
      {/* Hero section */}
      <div className="rounded-2xl bg-bg-elevated border border-border p-6 space-y-4">
        <div className="flex items-center justify-between">
          <div className="space-y-2">
            <Skeleton className="h-6 w-40 rounded-lg" />
            <Skeleton className="h-4 w-64 rounded-md" />
          </div>
          <SkeletonBadge />
        </div>
        <Skeleton className="h-3 w-full rounded-md" />
      </div>
      {/* Cards grid */}
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
        {Array.from({ length: 3 }).map((_, i) => (
          <SkeletonCard key={i} />
        ))}
      </div>
      {/* Table */}
      <SkeletonTable rows={4} cols={4} />
    </div>
  );
}

/** Form page skeleton — header + form fields */
export function SkeletonFormPage() {
  return (
    <div className="space-y-6 animate-fade-in max-w-2xl" aria-label="Loading form">
      {/* Header */}
      <div className="space-y-2">
        <Skeleton className="h-7 w-40 rounded-lg" />
        <Skeleton className="h-4 w-64 rounded-md" />
      </div>
      {/* Form card */}
      <div className="rounded-2xl bg-bg-elevated border border-border p-6 space-y-5">
        {Array.from({ length: 4 }).map((_, i) => (
          <div key={i} className="space-y-2">
            <Skeleton className="h-3.5 w-24 rounded-md" />
            <SkeletonInput />
          </div>
        ))}
        <div className="flex items-center gap-3 pt-2">
          <SkeletonButton />
          <SkeletonButton className="w-20" />
        </div>
      </div>
    </div>
  );
}
