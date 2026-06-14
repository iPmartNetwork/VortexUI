import { Component, type ReactNode } from "react";

interface Props { children: ReactNode; }
interface State { error: Error | null; }

export class ErrorBoundary extends Component<Props, State> {
  state: State = { error: null };
  static getDerivedStateFromError(error: Error) { return { error }; }
  render() {
    if (this.state.error) {
      return (
        <div className="flex min-h-screen items-center justify-center p-8">
          <div className="card max-w-md p-8 text-center space-y-4">
            <div className="text-4xl">⚠️</div>
            <h1 className="text-lg font-bold text-fg">Something went wrong</h1>
            <p className="text-sm text-fg-muted">{this.state.error.message}</p>
            <button onClick={() => window.location.reload()} className="grad-bg rounded-xl px-5 py-2.5 text-sm font-medium text-white shadow-lg">
              Reload
            </button>
          </div>
        </div>
      );
    }
    return this.props.children;
  }
}
