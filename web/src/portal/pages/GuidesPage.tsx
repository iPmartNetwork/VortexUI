import { useQuery } from "@tanstack/react-query";
import { motion } from "framer-motion";
import { BookOpen, ExternalLink, Monitor, Smartphone, Globe } from "lucide-react";
import { api } from "@/api/client";
import { GlassCard } from "@/components/vortexui";

interface Guide { id: string; app_name: string; platform: string; icon_url: string; content: string; }

const PLATFORM_ICONS: Record<string, React.ReactNode> = {
  windows: <Monitor size={14} />,
  android: <Smartphone size={14} />,
  ios: <Smartphone size={14} />,
  mac: <Monitor size={14} />,
  linux: <Monitor size={14} />,
  web: <Globe size={14} />,
};

function PlatformIcon({ platform }: { platform: string }) {
  const key = platform.toLowerCase();
  return PLATFORM_ICONS[key] ?? <ExternalLink size={14} />;
}

export function GuidesPage() {
  const { data: guides, isLoading } = useQuery({
    queryKey: ["portal-guides"],
    queryFn: () => api<Guide[]>("/api/v2/portal/guides"),
  });

  return (
    <div className="space-y-6 animate-page-enter">
      {/* ── Header ── */}
      <div className="relative overflow-hidden rounded-2xl border border-border/60 bg-gradient-to-br from-bg-elevated via-surface to-primary/[0.03] p-5 md:p-6">
        <div className="absolute top-0 end-0 w-56 h-56 rounded-full bg-primary/8 blur-3xl pointer-events-none" />
        <div className="absolute bottom-0 start-0 w-40 h-40 rounded-full bg-accent/8 blur-3xl pointer-events-none" />
        
        <div className="relative z-10 flex items-center gap-3">
          <div className="h-10 w-10 rounded-xl bg-primary/15 flex items-center justify-center">
            <BookOpen size={20} className="text-primary" />
          </div>
          <div>
            <h1 className="text-xl md:text-2xl font-black text-fg tracking-tight">
              Connection Guides
            </h1>
            <p className="text-[13px] text-fg-muted">
              {guides ? `${guides.length} guides available` : "Setup instructions for all platforms"}
            </p>
          </div>
        </div>
      </div>

      {/* ── Guides Grid ── */}
      {isLoading ? (
        <div className="grid gap-4 sm:grid-cols-2">
          {[1, 2].map(i => (
            <div key={i} className="h-48 animate-pulse rounded-xl bg-surface-2/50" />
          ))}
        </div>
      ) : guides && guides.length > 0 ? (
        <div className="grid gap-4 sm:grid-cols-2">
          {guides.map((g, i) => (
            <motion.div
              key={g.id}
              initial={{ opacity: 0, y: 10 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ delay: i * 0.05 }}
            >
              <GlassCard glow hover className="p-5 transition-all duration-300 group h-full">
                {/* App header */}
                <div className="flex items-center gap-3 mb-4">
                  {g.icon_url ? (
                    <div className="h-10 w-10 rounded-xl bg-white/10 flex items-center justify-center overflow-hidden ring-1 ring-border/30">
                      <img src={g.icon_url} alt="" className="w-7 h-7 object-contain" />
                    </div>
                  ) : (
                    <div className="h-10 w-10 rounded-xl bg-gradient-to-br from-primary/20 to-accent/20 flex items-center justify-center">
                      <BookOpen size={18} className="text-primary" />
                    </div>
                  )}
                  <div className="min-w-0">
                    <h3 className="text-sm font-bold text-fg group-hover:text-primary transition-colors truncate">
                      {g.app_name}
                    </h3>
                    <span className="inline-flex items-center gap-1 text-[10px] font-semibold text-fg-subtle uppercase tracking-wider">
                      <PlatformIcon platform={g.platform} />
                      {g.platform}
                    </span>
                  </div>
                </div>

                {/* Content */}
                <div
                  className="prose prose-sm dark:prose-invert max-w-none text-[13px] leading-relaxed text-fg-muted
                    [&_a]:text-primary [&_a]:font-medium [&_a]:no-underline [&_a:hover]:underline
                    [&_code]:bg-surface-2/80 [&_code]:px-1.5 [&_code]:py-0.5 [&_code]:rounded [&_code]:text-xs [&_code]:text-fg
                    [&_pre]:bg-bg-elevated [&_pre]:border [&_pre]:border-border/40 [&_pre]:rounded-xl [&_pre]:p-3 [&_pre]:text-xs
                    [&_img]:rounded-xl [&_img]:border [&_img]:border-border/30
                    [&_h1]:text-fg [&_h2]:text-fg [&_h3]:text-fg [&_h4]:text-fg
                    [&_strong]:text-fg"
                  dangerouslySetInnerHTML={{ __html: g.content }}
                />
              </GlassCard>
            </motion.div>
          ))}
        </div>
      ) : (
        <div className="flex flex-col items-center justify-center py-16 gap-3">
          <div className="h-12 w-12 rounded-full bg-surface-2/50 flex items-center justify-center">
            <BookOpen size={22} className="text-fg-subtle" />
          </div>
          <p className="text-sm text-fg-muted">No guides available.</p>
        </div>
      )}
    </div>
  );
}
