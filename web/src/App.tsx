import { Navigate, Route, Routes } from "react-router-dom";
import { useAuth } from "@/auth/auth";
import { Layout } from "@/components/Layout";
import { Login } from "@/pages/Login";
import { Overview } from "@/pages/Overview";
import { Users } from "@/pages/Users";
import { UserDetail } from "@/pages/UserDetail";
import { Nodes } from "@/pages/Nodes";
import { Inbounds } from "@/pages/Inbounds";
import { RoutingBalancers } from "@/pages/RoutingBalancers";
import { Routing } from "@/pages/Routing";
import { ResellerDashboard } from "@/pages/ResellerDashboard";
import { ResellerAccount } from "@/pages/ResellerAccount";
import { ResellerQuotaAlerts } from "@/pages/ResellerQuotaAlerts";
import { Logs } from "@/pages/Logs";
import { Settings } from "@/pages/Settings";
import { ResellerPlatform } from "@/pages/ResellerPlatform";
import { Orders } from "@/pages/Orders";
import { SecuritySuite } from "@/pages/SecuritySuite";
import { Monitor } from "@/pages/Monitor";
import { SmartQuota } from "@/pages/SmartQuota";
import { RelayChains } from "@/pages/RelayChains";
import { Analytics } from "@/pages/Analytics";
import { Tickets } from "@/pages/Tickets";
import { Migration } from "@/pages/Migration";
import { FamilyGroups } from "@/pages/FamilyGroups";
import { Referrals } from "@/pages/Referrals";
import { DoHSettings } from "@/pages/DoHSettings";
import { SNIManager } from "@/pages/SNIManager";
import { Fingerprint } from "@/pages/Fingerprint";
import { Federation } from "@/pages/Federation";
import { DeepLinks } from "@/pages/DeepLinks";
import { QuotaNotifications } from "@/pages/QuotaNotifications";
import { IPLimit } from "@/pages/IPLimit";
import { ResellerPaymentSettings } from "@/pages/ResellerPaymentSettings";
import { ResellerDetail } from "@/pages/admins/ResellerDetail";
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
        <Route path="/reseller-dashboard" element={<ResellerDashboard />} />
        <Route path="/reseller-account" element={<ResellerAccount />} />
        <Route path="/my-quota" element={<Navigate to="/reseller-dashboard" replace />} />
        <Route path="/nodes" element={<Nodes />} />
        <Route path="/inbounds" element={<Inbounds />} />
        <Route path="/outbounds" element={<Navigate to="/routing?tab=outbounds" replace />} />
        <Route path="/routing" element={<RoutingBalancers />} />
        <Route path="/routing/node-rules" element={<Routing />} />
        <Route path="/routing-packs" element={<Navigate to="/routing?tab=packs" replace />} />
        <Route path="/balancers" element={<Navigate to="/routing?tab=balancers" replace />} />
        <Route path="/admins" element={<Navigate to="/settings?tab=admins" replace />} />
        <Route path="/plans" element={<Navigate to="/wallet-billing?tab=plans" replace />} />
        <Route path="/wallet-billing" element={<ResellerPlatform />} />
        <Route path="/orders" element={<Orders />} />
        <Route path="/evasion" element={<SecuritySuite />} />
        <Route path="/monitor" element={<Monitor />} />
        <Route path="/reality-scanner" element={<Navigate to="/evasion?tab=reality" replace />} />
        <Route path="/clean-ip" element={<Navigate to="/evasion?tab=cleanip" replace />} />
        <Route path="/smart-quota" element={<SmartQuota />} />
        <Route path="/relay-chains" element={<RelayChains />} />
        <Route path="/decoy-website" element={<Navigate to="/evasion?tab=decoy" replace />} />
        <Route path="/analytics" element={<Analytics />} />
        <Route path="/tickets" element={<Tickets />} />
        <Route path="/migration" element={<Migration />} />
        <Route path="/probing-protection" element={<Navigate to="/evasion?tab=decoy" replace />} />
        <Route path="/ip-limit" element={<IPLimit />} />
        <Route path="/family-groups" element={<FamilyGroups />} />
        <Route path="/referrals" element={<Referrals />} />
        <Route path="/doh" element={<DoHSettings />} />
        <Route path="/sni-manager" element={<SNIManager />} />
        <Route path="/tls-tricks" element={<Navigate to="/evasion?tab=tls" replace />} />
        <Route path="/fingerprint" element={<Fingerprint />} />
        <Route path="/federation" element={<Federation />} />
        <Route path="/deep-links" element={<DeepLinks />} />
        <Route path="/quota-notifications" element={<QuotaNotifications />} />
        <Route path="/reseller-quota-alerts" element={<ResellerQuotaAlerts />} />
        <Route path="/reseller-payment" element={<ResellerPaymentSettings />} />
        <Route path="/pending-orders" element={<Navigate to="/wallet-billing?tab=orders" replace />} />
        <Route path="/audit" element={<Navigate to="/overview" replace />} />
        <Route path="/logs" element={<Logs />} />
        <Route path="/settings" element={<Settings />} />
        <Route path="/settings/admins/:id" element={<ResellerDetail />} />
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
        <a href="/overview" className="grad-bg inline-flex items-center gap-2 rounded-xl px-5 py-2.5 text-sm font-medium text-primary-fg shadow-lg">
          Go to Dashboard
        </a>
      </div>
    </div>
  );
}
