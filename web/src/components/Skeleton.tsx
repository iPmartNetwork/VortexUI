import { cn } from "@/lib/utils";

/** Base shimmer skeleton block */
export function Skeleton({ className }: { className?: string }) {
  return (
    <div className={cn("animate-shimmer rounded-xl bg-gradient-to-r from-surface-2/40 via-surface-2/80 to-surface-2/40 bg-[length:200%_100%]", className)} />
  );
}

/** Card-shaped skeleton */
export function SkeletonCard({ className }: { className?: string }) {
  return (
    <div className={cn("card p-5 space-y-3", className)}>
      <Skeleton className="h-4 w-1/3" />
      <Skeleton className="h-8 w-2/3" />
      <Skeleton className="h-3 w-1/2" />
    </div>
  );
}

/** Table row skeleton */
export function SkeletonRow({ cols = 5 }: { cols?: number }) {
  return (
    <div className="flex items-center gap-4 py-3 border-b border-border/20">
      {Array.from({ length: cols }).map((_, i) => (
        <Skeleton key={i} className={cn("h-4", i === 0 ? "w-32" : "flex-1")} />
      ))}
    </div>
  );
}

/** Table skeleton with header + rows */
export function SkeletonTable({ rows = 6, cols = 5 }: { rows?: number; cols?: number }) {
  return (
    <div className="card p-5 space-y-1">
      <div className="flex items-center gap-4 pb-3 border-b border-border/40">
        {Array.from({ length: cols }).map((_, i) => (
          <Skeleton key={i} className={cn("h-3", i === 0 ? "w-24" : "flex-1")} />
        ))}
      </div>
      {Array.from({ length: rows }).map((_, i) => (
        <SkeletonRow key={i} cols={cols} />
      ))}
    </div>
  );
}

/** Page-level skeleton: header + stat cards + table */
export function SkeletonPage() {
  return (
    <div className="space-y-6 animate-fade-in">
      {/* Header */}
      <div className="space-y-2">
        <Skeleton className="h-7 w-48" />
        <Skeleton className="h-4 w-72" />
      </div>
      {/* Stat cards */}
      <div className="grid grid-cols-2 gap-4 lg:grid-cols-4">
        <SkeletonCard />
        <SkeletonCard />
        <SkeletonCard />
        <SkeletonCard />
      </div>
      {/* Table */}
      <SkeletonTable />
    </div>
  );
}

/** Chart skeleton */
export function SkeletonChart({ className }: { className?: string }) {
  return (
    <div className={cn("card p-5 space-y-3", className)}>
      <Skeleton className="h-4 w-40" />
      <Skeleton className="h-40 w-full rounded-xl" />
    </div>
  );
}

/** Grid of cards skeleton */
export function SkeletonGrid({ count = 6 }: { count?: number }) {
  return (
    <div className="grid grid-cols-1 gap-4 md:grid-cols-2 lg:grid-cols-3">
      {Array.from({ length: count }).map((_, i) => (
        <SkeletonCard key={i} />
      ))}
    </div>
  );
}
