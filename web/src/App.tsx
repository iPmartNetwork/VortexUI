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
import { RoutingPacks } from "@/pages/RoutingPacks";
import { Balancers } from "@/pages/Balancers";
import { Admins } from "@/pages/Admins";
import { MyQuota } from "@/pages/MyQuota";
import { Audit } from "@/pages/Audit";
import { Logs } from "@/pages/Logs";
import { Settings } from "@/pages/Settings";
import { Plans } from "@/pages/Plans";
import { Orders } from "@/pages/Orders";
import { Evasion } from "@/pages/Evasion";
import { Monitor } from "@/pages/Monitor";
import { RealityScanner } from "@/pages/RealityScanner";
import { CleanIPScanner } from "@/pages/CleanIPScanner";
import { SmartQuota } from "@/pages/SmartQuota";
import { RelayChains } from "@/pages/RelayChains";
import { DecoyWebsite } from "@/pages/DecoyWebsite";
import { Analytics } from "@/pages/Analytics";
import { Tickets } from "@/pages/Tickets";
import { Migration } from "@/pages/Migration";
import { ProbingProtection } from "@/pages/ProbingProtection";
import { FamilyGroups } from "@/pages/FamilyGroups";
import { Referrals } from "@/pages/Referrals";
import { DoHSettings } from "@/pages/DoHSettings";
import { SNIManager } from "@/pages/SNIManager";
import { TLSTricks } from "@/pages/TLSTricks";
import { Fingerprint } from "@/pages/Fingerprint";
import { Federation } from "@/pages/Federation";
import { DeepLinks } from "@/pages/DeepLinks";
import { QuotaNotifications } from "@/pages/QuotaNotifications";
import { IPLimit } from "@/pages/IPLimit";
import { PortalLogin } from "@/pages/portal/PortalLogin";
import { PortalLayout } from "@/pages/portal/PortalLayout";
import { PortalDashboard } from "@/pages/portal/PortalDashboard";
import { PortalPlans } from "@/pages/portal/PortalPlans";
import { PortalTickets } from "@/pages/portal/PortalTickets";

function Protected({ children }: { children: React.ReactNode }) {
  const { isAuthenticated, loading } = useAuth();
  if (loading) return null;
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
        <Route path="/my-quota" element={<MyQuota />} />
        <Route path="/nodes" element={<Nodes />} />
        <Route path="/outbounds" element={<Outbounds />} />
        <Route path="/routing" element={<Routing />} />
        <Route path="/routing-packs" element={<RoutingPacks />} />
        <Route path="/balancers" element={<Balancers />} />
        <Route path="/admins" element={<Admins />} />
        <Route path="/plans" element={<Plans />} />
        <Route path="/orders" element={<Orders />} />
        <Route path="/evasion" element={<Evasion />} />
        <Route path="/monitor" element={<Monitor />} />
        <Route path="/reality-scanner" element={<RealityScanner />} />
        <Route path="/clean-ip" element={<CleanIPScanner />} />
        <Route path="/smart-quota" element={<SmartQuota />} />
        <Route path="/relay-chains" element={<RelayChains />} />
        <Route path="/decoy-website" element={<DecoyWebsite />} />
        <Route path="/analytics" element={<Analytics />} />
        <Route path="/tickets" element={<Tickets />} />
        <Route path="/migration" element={<Migration />} />
        <Route path="/probing-protection" element={<ProbingProtection />} />
        <Route path="/ip-limit" element={<IPLimit />} />
        <Route path="/family-groups" element={<FamilyGroups />} />
        <Route path="/referrals" element={<Referrals />} />
        <Route path="/doh" element={<DoHSettings />} />
        <Route path="/sni-manager" element={<SNIManager />} />
        <Route path="/tls-tricks" element={<TLSTricks />} />
        <Route path="/fingerprint" element={<Fingerprint />} />
        <Route path="/federation" element={<Federation />} />
        <Route path="/deep-links" element={<DeepLinks />} />
        <Route path="/quota-notifications" element={<QuotaNotifications />} />
        <Route path="/audit" element={<Audit />} />
        <Route path="/logs" element={<Logs />} />
        <Route path="/settings" element={<Settings />} />
      </Route>
      <Route path="*" element={<NotFound />} />
      {/* Portal (end-user self-service) */}
      <Route path="/portal/login" element={<PortalLogin />} />
      <Route element={<PortalLayout />}>
        <Route path="/portal/dashboard" element={<PortalDashboard />} />
        <Route path="/portal/plans" element={<PortalPlans />} />
        <Route path="/portal/tickets" element={<PortalTickets />} />
      </Route>
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
