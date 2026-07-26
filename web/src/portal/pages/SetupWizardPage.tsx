import { useQuery } from "@tanstack/react-query";
import { motion } from "framer-motion";
import { CheckCircle, ArrowRight, Sparkles, Smartphone } from "lucide-react";
import { api } from "@/api/client";
import { Button } from "@/components/ui";
import { GlassCard } from "@/components/vortexui";
import { cn } from "@/lib/utils";

interface WizardStep { step: number; title: string; description: string; action: string; completed: boolean; }

export function SetupWizardPage() {
  const { data: steps, isLoading } = useQuery({
    queryKey: ["setup-wizard"],
    queryFn: () => api<WizardStep[]>("/api/v2/portal/setup-wizard"),
  });

  const completedCount = steps?.filter(s => s.completed).length ?? 0;
  const totalCount = steps?.length ?? 0;
  const progressPct = totalCount > 0 ? (completedCount / totalCount) * 100 : 0;

  return (
    <div className="space-y-6 animate-page-enter">
      {/* ── Header ── */}
      <div className="relative overflow-hidden rounded-2xl border border-border/60 bg-gradient-to-br from-bg-elevated via-surface to-primary/[0.03] p-5 md:p-6">
        <div className="absolute top-0 end-0 w-56 h-56 rounded-full bg-primary/8 blur-3xl pointer-events-none" />
        <div className="absolute bottom-0 start-0 w-40 h-40 rounded-full bg-accent/8 blur-3xl pointer-events-none" />
        
        <div className="relative z-10">
          <div className="flex items-center gap-2 mb-2">
            <div className="h-8 w-8 rounded-lg bg-primary/15 flex items-center justify-center">
              <Sparkles size={16} className="text-primary" />
            </div>
            <h1 className="text-xl md:text-2xl font-black text-fg tracking-tight">
              Setup Wizard
            </h1>
          </div>
          <p className="text-[13px] text-fg-muted">Follow these steps to get connected.</p>

          {/* Progress bar */}
          {totalCount > 0 && (
            <div className="mt-4 flex items-center gap-3">
              <div className="flex-1 h-1.5 rounded-full bg-surface-3/60 overflow-hidden">
                <div
                  className="h-full rounded-full bg-gradient-to-r from-primary to-accent transition-all duration-700"
                  style={{ width: `${progressPct}%` }}
                />
              </div>
              <span className="text-xs font-bold tabular-nums text-fg-muted">
                {completedCount}/{totalCount}
              </span>
            </div>
          )}
        </div>
      </div>

      {/* ── Steps ── */}
      {isLoading ? (
        <div className="space-y-3">
          {[1, 2, 3].map(i => (
            <div key={i} className="h-20 animate-pulse rounded-xl bg-surface-2/50" />
          ))}
        </div>
      ) : steps && steps.length > 0 ? (
        <div className="space-y-3">
          {steps.map((s, i) => (
            <motion.div
              key={s.step}
              initial={{ opacity: 0, y: 10 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ delay: i * 0.06 }}
            >
              <GlassCard
                glow
                hover
                className={cn(
                  "relative overflow-hidden transition-all duration-300",
                  s.completed && "opacity-70 hover:opacity-100",
                )}
              >
                {/* Connector line (not last) */}
                {i < steps.length - 1 && (
                  <div className="absolute start-8 top-12 bottom-0 w-px bg-border/40" />
                )}

                <div className="flex items-start gap-4 relative z-10">
                  {/* Step circle */}
                  <div className={cn(
                    "h-8 w-8 rounded-full flex items-center justify-center flex-shrink-0 relative z-10 transition-all duration-300",
                    s.completed
                      ? "bg-success/15 text-success ring-2 ring-success/30"
                      : "bg-surface-2/60 text-fg-muted ring-1 ring-border/40",
                  )}>
                    {s.completed ? <CheckCircle size={16} /> : <span className="text-xs font-black">{s.step}</span>}
                  </div>

                  {/* Content */}
                  <div className="flex-1 min-w-0">
                    <h3 className="text-sm font-bold text-fg">{s.title}</h3>
                    <p className="text-xs text-fg-muted mt-0.5 leading-relaxed">{s.description}</p>
                  </div>

                  {/* Action */}
                  {!s.completed && (
                    <Button variant="glass" size="sm" className="flex-shrink-0 gap-1">
                      {s.action || "Continue"} <ArrowRight size={14} />
                    </Button>
                  )}
                  {s.completed && (
                    <span className="text-[10px] font-bold text-success flex items-center gap-1 flex-shrink-0">
                      <CheckCircle size={12} /> Done
                    </span>
                  )}
                </div>
              </GlassCard>
            </motion.div>
          ))}
        </div>
      ) : (
        <div className="flex flex-col items-center justify-center py-16 gap-3">
          <div className="h-12 w-12 rounded-full bg-surface-2/50 flex items-center justify-center">
            <Smartphone size={22} className="text-fg-subtle" />
          </div>
          <p className="text-sm text-fg-muted">No setup steps available.</p>
        </div>
      )}
    </div>
  );
}
