import { Navigate, Route, Routes } from "react-router-dom";
import { useAuth } from "@/auth/auth";
import { Layout } from "@/components/Layout";
import { Login } from "@/pages/Login";
import { Users } from "@/pages/Users";
import { Nodes } from "@/pages/Nodes";
import { Admins } from "@/pages/Admins";
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
        <Route path="/users" element={<Users />} />
        <Route path="/nodes" element={<Nodes />} />
        <Route path="/admins" element={<Admins />} />
        <Route path="/settings" element={<Settings />} />
      </Route>
      <Route path="*" element={<Navigate to="/users" replace />} />
    </Routes>
  );
}
