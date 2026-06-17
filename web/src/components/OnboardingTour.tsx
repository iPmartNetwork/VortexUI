import { useState, useEffect } from "react";
import { cn } from "@/lib/utils";
import { Button } from "@/components/ui";
import { X, ArrowRight, ArrowLeft, Sparkles } from "lucide-react";

const TOUR_KEY = "vortex.onboarding.done";

interface TourStep {
  title: string;
  description: string;
  target?: string; // CSS selector to highlight
  position?: "center" | "top" | "bottom";
}

const STEPS: TourStep[] = [
  {
    title: "Welcome to VortexUI! 🎉",
    description: "Let's take a quick tour of your new panel. This will only take a minute.",
    position: "center",
  },
  {
    title: "Dashboard Overview",
    description: "Your command center — real-time stats, traffic charts, node health, and system metrics all at a glance.",
    target: "[data-tour='dashboard']",
    position: "bottom",
  },
  {
    title: "User Management",
    description: "Create and manage VPN users, set quotas, track usage, and handle subscriptions from one place.",
    target: "[data-tour='users']",
    position: "bottom",
  },
  {
    title: "Node Fleet",
    description: "Add servers, monitor health, configure inbounds/outbounds, and manage your entire infrastructure.",
    target: "[data-tour='network']",
    position: "bottom",
  },
  {
    title: "Security Tools",
    description: "TLS tricks, reality scanner, probing protection, fingerprint validation — all your anti-censorship tools.",
    target: "[data-tour='security']",
    position: "bottom",
  },
  {
    title: "Quick Search (⌘K)",
    description: "Press Ctrl+K (or ⌘K on Mac) anytime to instantly search and navigate to any page, user, or setting.",
    position: "center",
  },
  {
    title: "You're all set!",
    description: "Explore your panel. You can always re-run this tour from Settings. Happy managing!",
    position: "center",
  },
];

export function OnboardingTour() {
  const [active, setActive] = useState(false);
  const [step, setStep] = useState(0);

  useEffect(() => {
    // Show on first visit
    if (!localStorage.getItem(TOUR_KEY)) {
      const timer = setTimeout(() => setActive(true), 1000);
      return () => clearTimeout(timer);
    }
  }, []);

  // Listen for manual re-trigger
  useEffect(() => {
    function handler() { setStep(0); setActive(true); }
    window.addEventListener("vortex:start-tour", handler);
    return () => window.removeEventListener("vortex:start-tour", handler);
  }, []);

  function finish() {
    setActive(false);
    localStorage.setItem(TOUR_KEY, "true");
  }

  function next() {
    if (step >= STEPS.length - 1) finish();
    else setStep(s => s + 1);
  }

  function prev() {
    if (step > 0) setStep(s => s - 1);
  }

  if (!active) return null;

  const current = STEPS[step];
  const isCenter = !current.target || current.position === "center";

  return (
    <div className="fixed inset-0 z-[300]">
      {/* Overlay */}
      <div className="absolute inset-0 bg-black/60 backdrop-blur-sm animate-fade-in" />

      {/* Tooltip */}
      <div className={cn(
        "absolute z-10 animate-scale-in",
        isCenter ? "inset-0 flex items-center justify-center p-4" : "bottom-8 left-1/2 -translate-x-1/2",
      )}>
        <div className="card w-full max-w-md p-6 shadow-2xl space-y-4">
          {/* Header */}
          <div className="flex items-start justify-between">
            <div className="flex items-center gap-2">
              <Sparkles size={18} className="text-primary" />
              <h3 className="text-base font-bold text-fg">{current.title}</h3>
            </div>
            <button onClick={finish} className="grid h-7 w-7 place-items-center rounded-lg text-fg-subtle hover:bg-surface-2 hover:text-fg">
              <X size={14} />
            </button>
          </div>

          {/* Body */}
          <p className="text-sm text-fg-muted leading-relaxed">{current.description}</p>

          {/* Progress dots */}
          <div className="flex items-center justify-center gap-1.5">
            {STEPS.map((_, i) => (
              <div key={i} className={cn(
                "h-1.5 rounded-full transition-all duration-300",
                i === step ? "w-6 bg-primary" : "w-1.5 bg-border",
              )} />
            ))}
          </div>

          {/* Navigation */}
          <div className="flex items-center justify-between pt-2">
            <button onClick={finish} className="text-xs text-fg-subtle hover:text-fg-muted transition">
              Skip tour
            </button>
            <div className="flex gap-2">
              {step > 0 && (
                <Button variant="ghost" size="sm" onClick={prev}>
                  <ArrowLeft size={14} /> Back
                </Button>
              )}
              <Button size="sm" onClick={next}>
                {step >= STEPS.length - 1 ? "Done" : <>Next <ArrowRight size={14} /></>}
              </Button>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
