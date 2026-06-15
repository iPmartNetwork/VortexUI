import { Navigate, Route, Routes } from "react-router-dom";
import { useAuth } from "@/auth/auth";
import { Layout } from "@/components/Layout";
import { Login } from "@/pages/Login";
import { Overview } from "@/pages/Overview";
import { Users } from "@/pages/Users";
import { UserDetail } from "@/pages/UserDetail";
import { Nodes } from "@/pages/Nodes";
import { Outbounds } from "@/pages/Outbounds";
import { Routing } from "@/pages/Routing";
import { Balancers } from "@/pages/Balancers";
import { Admins } from "@/pages/Admins";
import { Audit } from "@/pages/Audit";
import { Logs } from "@/pages/Logs";
import { Settings } from "@/pages/Settings";

function Protected({ children }: { children: React.ReactNode }) {
  const { isAuthenticated } = useAuth();
  return isAuthenticated ? <>{children}</> : <Navigate to="/login" replace />;
}

export function App() {
  return (
    <Routes>
      <Route path="/login" element={<Login />} />
      <Route
        element={
          <Protected>
            <Layout />
          </Protected>
        }
      >
        <Route path="/overview" element={<Overview />} />
        <Route path="/users" element={<Users />} />
        <Route path="/users/:id" element={<UserDetail />} />
        <Route path="/nodes" element={<Nodes />} />
        <Route path="/outbounds" element={<Outbounds />} />
        <Route path="/routing" element={<Routing />} />
        <Route path="/balancers" element={<Balancers />} />
        <Route path="/admins" element={<Admins />} />
        <Route path="/audit" element={<Audit />} />
        <Route path="/logs" element={<Logs />} />
        <Route path="/settings" element={<Settings />} />
      </Route>
      <Route path="*" element={<NotFound />} />
    </Routes>
  );
}

function NotFound() {
  return (
    <div className="flex min-h-screen items-center justify-center p-8">
      <div className="card max-w-sm p-8 text-center space-y-4 animate-scale-in">
        <div className="text-6xl">404</div>
        <h1 className="text-lg font-bold text-fg">Page not found</h1>
        <p className="text-sm text-fg-muted">The page you're looking for doesn't exist.</p>
        <a href="/overview" className="grad-bg inline-flex items-center gap-2 rounded-xl px-5 py-2.5 text-sm font-medium text-white shadow-lg">
          Go to Dashboard
        </a>
      </div>
    </div>
  );
}
