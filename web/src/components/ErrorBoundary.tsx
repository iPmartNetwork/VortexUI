import { Component } from "react";
import { Button } from "@/components/ui";
import { AlertTriangle, RotateCcw } from "lucide-react";

interface Props {
  children: React.ReactNode;
  fallback?: React.ReactNode;
}

interface State {
  hasError: boolean;
  error: Error | null;
}

export class ErrorBoundary extends Component<Props, State> {
  state: State = { hasError: false, error: null };

  static getDerivedStateFromError(error: Error): State {
    return { hasError: true, error };
  }

  render() {
    if (this.state.hasError) {
      if (this.props.fallback) return this.props.fallback;
      return (
        <div className="flex min-h-[400px] items-center justify-center p-8 animate-fade-in">
          <div className="card max-w-md p-8 text-center space-y-4">
            <div className="mx-auto grid h-14 w-14 place-items-center rounded-2xl bg-danger/10 text-danger">
              <AlertTriangle size={28} />
            </div>
            <h2 className="text-lg font-bold text-fg">Something went wrong</h2>
            <p className="text-sm text-fg-muted">
              {this.state.error?.message || "An unexpected error occurred."}
            </p>
            <Button onClick={() => this.setState({ hasError: false, error: null })}>
              <RotateCcw size={14} /> Try Again
            </Button>
          </div>
        </div>
      );
    }
    return this.props.children;
  }
}
