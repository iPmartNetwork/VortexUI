import { useState } from "react";
import { cn } from "@/lib/utils";
import { Button } from "@/components/ui";
import { Check } from "lucide-react";

export interface WizardStep {
  title: string;
  description?: string;
  content: React.ReactNode;
  validate?: () => boolean;
}

interface WizardProps {
  steps: WizardStep[];
  onComplete: () => void;
  onCancel?: () => void;
  completeLabel?: string;
}

export function Wizard({ steps, onComplete, onCancel, completeLabel = "Complete" }: WizardProps) {
  const [current, setCurrent] = useState(0);
  const step = steps[current];
  const isLast = current === steps.length - 1;

  function next() {
    if (step.validate && !step.validate()) return;
    if (isLast) { onComplete(); return; }
    setCurrent(c => c + 1);
  }

  function prev() {
    if (current > 0) setCurrent(c => c - 1);
  }

  return (
    <div className="space-y-6">
      {/* Step indicators */}
      <div className="flex items-center gap-2">
        {steps.map((_, i) => (
          <div key={i} className="flex items-center gap-2">
            <div className={cn(
              "grid h-7 w-7 place-items-center rounded-full text-xs font-bold transition",
              i < current ? "bg-success text-white" :
              i === current ? "bg-primary text-white" :
              "bg-surface-2 text-fg-subtle",
            )}>
              {i < current ? <Check size={14} /> : i + 1}
            </div>
            {i < steps.length - 1 && (
              <div className={cn("h-0.5 w-8 rounded-full transition", i < current ? "bg-success" : "bg-border/40")} />
            )}
          </div>
        ))}
      </div>

      {/* Step header */}
      <div>
        <h3 className="text-sm font-bold text-fg">{step.title}</h3>
        {step.description && <p className="text-xs text-fg-muted mt-1">{step.description}</p>}
      </div>

      {/* Step content */}
      <div className="animate-fade-in" key={current}>
        {step.content}
      </div>

      {/* Navigation */}
      <div className="flex items-center justify-between border-t border-border/40 pt-4">
        <div>
          {onCancel && <Button variant="ghost" onClick={onCancel}>Cancel</Button>}
        </div>
        <div className="flex gap-2">
          {current > 0 && <Button variant="outline" onClick={prev}>Back</Button>}
          <Button onClick={next}>{isLast ? completeLabel : "Next"}</Button>
        </div>
      </div>
    </div>
  );
}
