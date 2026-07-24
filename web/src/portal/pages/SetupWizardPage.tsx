import { useQuery } from "@tanstack/react-query";
import { CheckCircle, Circle, ArrowRight } from "lucide-react";
import { api } from "@/api/client";
import { Button } from "@/components/ui";
import { GlassCard } from "@/components/veltrix";

interface WizardStep { step: number; title: string; description: string; action: string; completed: boolean; }

export function SetupWizardPage() {
  const { data: steps } = useQuery({
    queryKey: ["setup-wizard"],
    queryFn: () => api<WizardStep[]>("/api/v2/portal/setup-wizard"),
  });

  return (
    <div className="space-y-6">
      <h2 className="text-xl font-bold">Setup Wizard</h2>
      <p className="text-fg-muted">Follow these steps to get connected.</p>
      {steps && (
        <div className="space-y-3">{steps.map((s) => (
          <GlassCard key={s.step} className={`p-4 flex items-center gap-4 ${s.completed?"opacity-70":""}`}>
            <div className="flex-shrink-0">{s.completed ? <CheckCircle className="w-6 h-6 text-success"/> : <Circle className="w-6 h-6 text-fg-muted"/>}</div>
            <div className="flex-1"><h3 className="font-medium">{s.step}. {s.title}</h3><p className="text-sm text-fg-muted">{s.description}</p></div>
            {!s.completed && <Button variant="ghost" size="sm"><ArrowRight className="w-4 h-4"/></Button>}
          </GlassCard>
        ))}</div>
      )}
    </div>
  );
}
