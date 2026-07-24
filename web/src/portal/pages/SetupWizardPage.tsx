import { useQuery } from "@tanstack/react-query";
import { CheckCircle, Circle, ArrowRight } from "lucide-react";
import { api } from "@/api/client";
import { Button } from "@/components/ui";
import { GlassCard } from "@/components/veltrix";
import { useI18n } from "@/i18n/i18n";

interface WizardStep {
  step: number;
  title: string;
  description: string;
  action: string;
  completed: boolean;
}

export function SetupWizardPage() {
  const { t } = useI18n();

  const { data: steps } = useQuery({
    queryKey: ["setup-wizard"],
    queryFn: () =>
      api.get("/api/v2/portal/setup-wizard").then((r) => r.data as WizardStep[]),
  });

  return (
    <div className="space-y-6">
      <h2 className="text-xl font-bold">
        {t("portal.setupWizard", "Setup Wizard")}
      </h2>
      <p className="text-muted-foreground">
        {t("portal.setupWizardDesc", "Follow these steps to get connected.")}
      </p>

      {steps && (
        <div className="space-y-3">
          {steps.map((step) => (
            <GlassCard
              key={step.step}
              className={`p-4 flex items-center gap-4 ${
                step.completed ? "opacity-70" : ""
              }`}
            >
              <div className="flex-shrink-0">
                {step.completed ? (
                  <CheckCircle className="w-6 h-6 text-green-500" />
                ) : (
                  <Circle className="w-6 h-6 text-muted-foreground" />
                )}
              </div>
              <div className="flex-1">
                <h3 className="font-medium">
                  {step.step}. {step.title}
                </h3>
                <p className="text-sm text-muted-foreground">
                  {step.description}
                </p>
              </div>
              {!step.completed && (
                <Button variant="ghost" size="sm">
                  <ArrowRight className="w-4 h-4" />
                </Button>
              )}
            </GlassCard>
          ))}
        </div>
      )}
    </div>
  );
}
