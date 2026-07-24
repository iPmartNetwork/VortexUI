import { useQuery } from "@tanstack/react-query";
import { BookOpen } from "lucide-react";
import { api } from "@/api/client";
import { GlassCard } from "@/components/veltrix";

interface Guide { id: string; app_name: string; platform: string; icon_url: string; content: string; }

export function GuidesPage() {
  const { data: guides } = useQuery({
    queryKey: ["portal-guides"],
    queryFn: () => api<Guide[]>("/api/v2/portal/guides"),
  });

  return (
    <div className="space-y-6">
      <h2 className="text-xl font-bold flex items-center gap-2"><BookOpen className="w-5 h-5"/>Connection Guides</h2>
      {guides && guides.length > 0 ? (
        <div className="grid gap-4 sm:grid-cols-2">{guides.map((g) => (
          <GlassCard key={g.id} className="p-4">
            <div className="flex items-center gap-3 mb-3">
              {g.icon_url && <img src={g.icon_url} alt="" className="w-8 h-8 rounded"/>}
              <div><h3 className="font-medium">{g.app_name}</h3><span className="text-xs text-fg-muted">{g.platform}</span></div>
            </div>
            <div className="prose prose-sm dark:prose-invert" dangerouslySetInnerHTML={{ __html: g.content }}/>
          </GlassCard>
        ))}</div>
      ) : <p className="text-fg-muted text-sm">No guides available.</p>}
    </div>
  );
}
