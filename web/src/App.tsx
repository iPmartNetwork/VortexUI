import { Suspense, lazy, type ComponentType } from "react";
import { Navigate, Route, Routes } from "react-router-dom";
import { useAuth } from "@/auth/auth";
import { Layout } from "@/components/Layout";
import { Login } from "@/pages/Login";
import { Overview } from "@/pages/Overview";
import { SkeletonPage } from "@/components/Skeleton";

const LazyUsers = lazy(() => import("@/pages/Users").then((m) => ({ default: m.Users })));
const LazyUserDetail = lazy(() => import("@/pages/UserDetail").then((m) => ({ default: m.UserDetail })));
const LazyNodes = lazy(() => import("@/pages/Nodes").then((m) => ({ default: m.Nodes })));
const LazyInbounds = lazy(() => import("@/pages/Inbounds").then((m) => ({ default: m.Inbounds })));
const LazyRoutingBalancers = lazy(() => import("@/pages/RoutingBalancers").then((m) => ({ default: m.RoutingBalancers })));
const LazyRouting = lazy(() => import("@/pages/Routing").then((m) => ({ default: m.Routing })));
const LazyResellerDashboard = lazy(() => import("@/pages/ResellerDashboard").then((m) => ({ default: m.ResellerDashboard })));
const LazyResellerAccount = lazy(() => import("@/pages/ResellerAccount").then((m) => ({ default: m.ResellerAccount })));
const LazyResellerQuotaAlerts = lazy(() => import("@/pages/ResellerQuotaAlerts").then((m) => ({ default: m.ResellerQuotaAlerts })));
const LazyAudit = lazy(() => import("@/pages/Audit").then((m) => ({ default: m.Audit })));
const LazyLogs = lazy(() => import("@/pages/Logs").then((m) => ({ default: m.Logs })));
const LazySettings = lazy(() => import("@/pages/Settings").then((m) => ({ default: m.Settings })));
const LazyResellerPlatform = lazy(() => import("@/pages/ResellerPlatform").then((m) => ({ default: m.ResellerPlatform })));
const LazyOrders = lazy(() => import("@/pages/Orders").then((m) => ({ default: m.Orders })));
const LazySecuritySuite = lazy(() => import("@/pages/SecuritySuite").then((m) => ({ default: m.SecuritySuite })));
const LazyMonitor = lazy(() => import("@/pages/Monitor").then((m) => ({ default: m.Monitor })));
const LazySmartQuota = lazy(() => import("@/pages/SmartQuota").then((m) => ({ default: m.SmartQuota })));
const LazyRelayChains = lazy(() => import("@/pages/RelayChains").then((m) => ({ default: m.RelayChains })));
const LazyAnalytics = lazy(() => import("@/pages/Analytics").then((m) => ({ default: m.Analytics })));
const LazyTickets = lazy(() => import("@/pages/Tickets").then((m) => ({ default: m.Tickets })));
const LazyMigration = lazy(() => import("@/pages/Migration").then((m) => ({ default: m.Migration })));
const LazyFamilyGroups = lazy(() => import("@/pages/FamilyGroups").then((m) => ({ default: m.FamilyGroups })));
const LazyReferrals = lazy(() => import("@/pages/Referrals").then((m) => ({ default: m.Referrals })));
const LazyDoHSettings = lazy(() => import("@/pages/DoHSettings").then((m) => ({ default: m.DoHSettings })));
const LazySNIManager = lazy(() => import("@/pages/SNIManager").then((m) => ({ default: m.SNIManager })));
const LazyFingerprint = lazy(() => import("@/pages/Fingerprint").then((m) => ({ default: m.Fingerprint })));
const LazyFederation = lazy(() => import("@/pages/Federation").then((m) => ({ default: m.Federation })));
const LazyDeepLinks = lazy(() => import("@/pages/DeepLinks").then((m) => ({ default: m.DeepLinks })));
const LazyQuotaNotifications = lazy(() => import("@/pages/QuotaNotifications").then((m) => ({ default: m.QuotaNotifications })));
const LazyIPLimit = lazy(() => import("@/pages/IPLimit").then((m) => ({ default: m.IPLimit })));
const LazyResellerPaymentSettings = lazy(() => import("@/pages/ResellerPaymentSettings").then((m) => ({ default: m.ResellerPaymentSettings })));
// PHASE 3 Components
const LazyPerformance = lazy(() => import("@/pages/Performance").then((m) => ({ default: m.Performance })));
const LazySecurityHardening = lazy(() => import("@/pages/SecurityHardening").then((m) => ({ default: m.SecurityHardening })));
const LazyCompliance = lazy(() => import("@/pages/Compliance").then((m) => ({ default: m.Compliance })));
const LazyResellerDetail = lazy(() => import("@/pages/admins/ResellerDetail").then((m) => ({ default: m.ResellerDetail })));
const LazyPortalLogin = lazy(() => import("@/pages/portal/PortalLogin").then((m) => ({ default: m.PortalLogin })));
const LazyPortalLayout = lazy(() => import("@/pages/portal/PortalLayout").then((m) => ({ default: m.PortalLayout })));
const LazyPortalDashboard = lazy(() => import("@/pages/portal/PortalDashboard").then((m) => ({ default: m.PortalDashboard })));
const LazyPortalPlans = lazy(() => import("@/pages/portal/PortalPlans").then((m) => ({ default: m.PortalPlans })));
const LazyPortalTickets = lazy(() => import("@/pages/portal/PortalTickets").then((m) => ({ default: m.PortalTickets })));
const LazyPortalReferral = lazy(() => import("@/pages/portal/PortalReferral").then((m) => ({ default: m.PortalReferral })));

const LazyRoute = ({ component: Component }: { component: ComponentType }) => {
  return (
    <Suspense fallback={<SkeletonPage />}>
      <Component />
    </Suspense>
  );
};

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
        <Route path="/users" element={<LazyRoute component={LazyUsers} />} />
        <Route path="/users/:id" element={<LazyRoute component={LazyUserDetail} />} />
        <Route path="/reseller-dashboard" element={<LazyRoute component={LazyResellerDashboard} />} />
        <Route path="/reseller-account" element={<LazyRoute component={LazyResellerAccount} />} />
        <Route path="/my-quota" element={<Navigate to="/reseller-dashboard" replace />} />
        <Route path="/nodes" element={<LazyRoute component={LazyNodes} />} />
        <Route path="/inbounds" element={<LazyRoute component={LazyInbounds} />} />
        <Route path="/outbounds" element={<Navigate to="/routing?tab=outbounds" replace />} />
        <Route path="/routing" element={<LazyRoute component={LazyRoutingBalancers} />} />
        <Route path="/routing/node-rules" element={<LazyRoute component={LazyRouting} />} />
        <Route path="/routing-packs" element={<Navigate to="/routing?tab=packs" replace />} />
        <Route path="/balancers" element={<Navigate to="/routing?tab=balancers" replace />} />
        <Route path="/admins" element={<Navigate to="/settings?tab=admins" replace />} />
        <Route path="/plans" element={<Navigate to="/wallet-billing?tab=plans" replace />} />
        <Route path="/wallet-billing" element={<LazyRoute component={LazyResellerPlatform} />} />
        <Route path="/orders" element={<LazyRoute component={LazyOrders} />} />
        <Route path="/evasion" element={<LazyRoute component={LazySecuritySuite} />} />
        <Route path="/monitor" element={<LazyRoute component={LazyMonitor} />} />
        <Route path="/reality-scanner" element={<Navigate to="/evasion?tab=reality" replace />} />
        <Route path="/clean-ip" element={<Navigate to="/evasion?tab=cleanip" replace />} />
        <Route path="/smart-quota" element={<LazyRoute component={LazySmartQuota} />} />
        <Route path="/relay-chains" element={<LazyRoute component={LazyRelayChains} />} />
        <Route path="/decoy-website" element={<Navigate to="/evasion?tab=decoy" replace />} />
        <Route path="/analytics" element={<LazyRoute component={LazyAnalytics} />} />
        <Route path="/tickets" element={<LazyRoute component={LazyTickets} />} />
        <Route path="/migration" element={<LazyRoute component={LazyMigration} />} />
        <Route path="/probing-protection" element={<Navigate to="/evasion?tab=decoy" replace />} />
        <Route path="/ip-limit" element={<LazyRoute component={LazyIPLimit} />} />
        <Route path="/family-groups" element={<LazyRoute component={LazyFamilyGroups} />} />
        <Route path="/referrals" element={<LazyRoute component={LazyReferrals} />} />
        <Route path="/doh" element={<LazyRoute component={LazyDoHSettings} />} />
        <Route path="/sni-manager" element={<LazyRoute component={LazySNIManager} />} />
        <Route path="/tls-tricks" element={<Navigate to="/evasion?tab=tls" replace />} />
        <Route path="/fingerprint" element={<LazyRoute component={LazyFingerprint} />} />
        <Route path="/federation" element={<LazyRoute component={LazyFederation} />} />
        <Route path="/deep-links" element={<LazyRoute component={LazyDeepLinks} />} />
        <Route path="/quota-notifications" element={<LazyRoute component={LazyQuotaNotifications} />} />
        <Route path="/reseller-quota-alerts" element={<LazyRoute component={LazyResellerQuotaAlerts} />} />
        <Route path="/reseller-payment" element={<LazyRoute component={LazyResellerPaymentSettings} />} />
        <Route path="/pending-orders" element={<Navigate to="/wallet-billing?tab=orders" replace />} />
        <Route path="/audit" element={<LazyRoute component={LazyAudit} />} />
        <Route path="/logs" element={<LazyRoute component={LazyLogs} />} />
        <Route path="/settings" element={<LazyRoute component={LazySettings} />} />
        <Route path="/settings/admins/:id" element={<LazyRoute component={LazyResellerDetail} />} />
        {/* PHASE 3 Routes */}
        <Route path="/performance" element={<LazyRoute component={LazyPerformance} />} />
        <Route path="/security" element={<LazyRoute component={LazySecurityHardening} />} />
        <Route path="/compliance" element={<LazyRoute component={LazyCompliance} />} />
      </Route>
      <Route path="*" element={<NotFound />} />
      {/* Portal (end-user self-service) */}
      <Route path="/portal/login" element={<LazyRoute component={LazyPortalLogin} />} />
      <Route element={<LazyRoute component={LazyPortalLayout} />}>
        <Route path="/portal/dashboard" element={<LazyRoute component={LazyPortalDashboard} />} />
        <Route path="/portal/plans" element={<LazyRoute component={LazyPortalPlans} />} />
        <Route path="/portal/referral" element={<LazyRoute component={LazyPortalReferral} />} />
        <Route path="/portal/tickets" element={<LazyRoute component={LazyPortalTickets} />} />
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
