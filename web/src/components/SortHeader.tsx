import { ChevronUp, ChevronDown, ChevronsUpDown } from "lucide-react";
import { cn } from "@/lib/utils";

export type SortDir = "asc" | "desc" | null;

interface Props {
  label: string;
  active: boolean;
  dir: SortDir;
  onClick: () => void;
  className?: string;
}

export function SortHeader({ label, active, dir, onClick, className }: Props) {
  return (
    <button onClick={onClick} className={cn("flex items-center gap-1 font-medium hover:text-fg transition", className)}>
      {label}
      {!active && <ChevronsUpDown size={12} className="text-fg-subtle/50" />}
      {active && dir === "asc" && <ChevronUp size={12} className="text-primary" />}
      {active && dir === "desc" && <ChevronDown size={12} className="text-primary" />}
    </button>
  );
}

export function cycleSort(current: SortDir): SortDir {
  if (current === null) return "asc";
  if (current === "asc") return "desc";
  return null;
}
