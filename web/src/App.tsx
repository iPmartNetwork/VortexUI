import { Suspense, lazy, type ComponentType, useState } from "react";
import { Navigate, Route, Routes } from "react-router-dom";
import { motion } from "framer-motion";
import { useAuth } from "@/auth/auth";
import { Layout } from "@/components/Layout";
import { Login } from "@/pages/Login";
import { ErrorBoundary } from "@/components/ErrorBoundary";
import { SplashScreen } from "@/components/SplashScreen";
import {
  SkeletonPage,
  SkeletonListPage,
  SkeletonDetailPage,
  SkeletonDashboardPage,
  SkeletonAnalyticsPage,
  SkeletonPortalPage,
  SkeletonFormPage,
} from "@/components/Skeleton";

const LazyOverview = lazy(() => import("@/pages/Overview").then((m) => ({ default: m.Overview })));
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
const LazyConnectionQuality = lazy(() => import("@/pages/ConnectionQuality").then((m) => ({ default: m.ConnectionQuality })));
const LazySwitchAnalytics = lazy(() => import("@/pages/SwitchAnalytics").then((m) => ({ default: m.SwitchAnalytics })));
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

// Enterprise v1.4.1 Components
const LazyTemplates = lazy(() => import("@/pages/Templates").then((m) => ({ default: m.Templates })));
const LazyBulkOperations = lazy(() => import("@/pages/BulkOperations").then((m) => ({ default: m.BulkOperations })));
const LazyNotifications = lazy(() => import("@/pages/Notifications").then((m) => ({ default: m.Notifications })));
const LazyDashboardPro = lazy(() => import("@/pages/DashboardPro").then((m) => ({ default: m.DashboardPro })));
const LazyClientTemplates = lazy(() => import("@/pages/ClientTemplates").then((m) => ({ default: m.ClientTemplates })));
const LazyConfigManagement = lazy(() => import("@/pages/ConfigManagement").then((m) => ({ default: m.ConfigManagement })));
const LazyWireGuard = lazy(() => import("@/pages/WireGuard").then((m) => ({ default: m.WireGuard })));
const LazyAdvancedSecurity = lazy(() => import("@/pages/security/AdvancedSecurity").then((m) => ({ default: m.AdvancedSecurity })));
const LazyApiDocs = lazy(() => import("@/pages/ApiDocs").then((m) => ({ default: m.ApiDocs })));
const LazyTunnels = lazy(() => import("@/pages/Tunnels").then((m) => ({ default: m.Tunnels })));
const LazyGeoExits = lazy(() => import("@/pages/GeoExits").then((m) => ({ default: m.GeoExits })));
const LazyIPRotation = lazy(() => import("@/pages/IPRotation").then((m) => ({ default: m.IPRotation })));
const LazyPortalSpeedTest = lazy(() => import("@/portal/pages/SpeedTestPage").then((m) => ({ default: m.SpeedTestPage })));
const LazyPortalGuides = lazy(() => import("@/portal/pages/GuidesPage").then((m) => ({ default: m.GuidesPage })));
const LazyPortalSetupWizard = lazy(() => import("@/portal/pages/SetupWizardPage").then((m) => ({ default: m.SetupWizardPage })));

// ─── Page transition animation ──────────────────────────────────
const pageVariants = {
  hidden: { opacity: 0, y: 10 },
  visible: { opacity: 1, y: 0 },
};

// ─── LazyRoute with custom skeleton support ─────────────────────
type SkeletonType =
  | "default"
  | "list"
  | "detail"
  | "dashboard"
  | "analytics"
  | "portal"
  | "form";

const skeletonComponents: Record<SkeletonType, ComponentType> = {
  default: SkeletonPage,
  list: SkeletonListPage,
  detail: SkeletonDetailPage,
  dashboard: SkeletonDashboardPage,
  analytics: SkeletonAnalyticsPage,
  portal: SkeletonPortalPage,
  form: SkeletonFormPage,
};

const LazyRoute = ({
  component: Component,
  skeleton = "default",
}: {
  component: ComponentType;
  skeleton?: SkeletonType;
}) => {
  const SkeletonComponent = skeletonComponents[skeleton];
  return (
    <ErrorBoundary>
      <Suspense fallback={<SkeletonComponent />}>
        <motion.div
          variants={pageVariants}
          initial="hidden"
          animate="visible"
          transition={{ duration: 0.3, ease: [0.16, 1, 0.3, 1] }}
        >
          <Component />
        </motion.div>
      </Suspense>
    </ErrorBoundary>
  );
};

function Protected({ children }: { children: React.ReactNode }) {
  const { isAuthenticated, loading } = useAuth();
  if (loading) return null;
  return isAuthenticated ? <>{children}</> : <Navigate to="/login" replace />;
}

export function App() {
  const [splashDone, setSplashDone] = useState(false);

  // Show splash during initial load (handles auth timing internally)
  if (!splashDone) {
    return <SplashScreen onFinish={() => setSplashDone(true)} />;
  }

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
        {/* Dashboard / Overview */}
        <Route path="/overview" element={<LazyRoute component={LazyOverview} skeleton="dashboard" />} />

        {/* List pages */}
        <Route path="/users" element={<LazyRoute component={LazyUsers} skeleton="list" />} />
        <Route path="/nodes" element={<LazyRoute component={LazyNodes} skeleton="list" />} />
        <Route path="/inbounds" element={<LazyRoute component={LazyInbounds} skeleton="list" />} />
        <Route path="/tickets" element={<LazyRoute component={LazyTickets} skeleton="list" />} />
        <Route path="/monitor" element={<LazyRoute component={LazyMonitor} skeleton="list" />} />
        <Route path="/orders" element={<LazyRoute component={LazyOrders} skeleton="list" />} />
        <Route path="/audit" element={<LazyRoute component={LazyAudit} skeleton="list" />} />
        <Route path="/logs" element={<LazyRoute component={LazyLogs} skeleton="list" />} />
        <Route path="/migration" element={<LazyRoute component={LazyMigration} skeleton="list" />} />
        <Route path="/family-groups" element={<LazyRoute component={LazyFamilyGroups} skeleton="list" />} />
        <Route path="/referrals" element={<LazyRoute component={LazyReferrals} skeleton="list" />} />
        <Route path="/ip-limit" element={<LazyRoute component={LazyIPLimit} skeleton="list" />} />
        <Route path="/connection-quality" element={<LazyRoute component={LazyConnectionQuality} skeleton="list" />} />
        <Route path="/switch-analytics" element={<LazyRoute component={LazySwitchAnalytics} skeleton="list" />} />
        <Route path="/quota-notifications" element={<LazyRoute component={LazyQuotaNotifications} skeleton="list" />} />
        <Route path="/templates" element={<LazyRoute component={LazyTemplates} skeleton="list" />} />
        <Route path="/bulk-operations" element={<LazyRoute component={LazyBulkOperations} skeleton="list" />} />
        <Route path="/notifications" element={<LazyRoute component={LazyNotifications} skeleton="list" />} />
        <Route path="/client-templates" element={<LazyRoute component={LazyClientTemplates} skeleton="list" />} />

        {/* Detail pages */}
        <Route path="/users/:id" element={<LazyRoute component={LazyUserDetail} skeleton="detail" />} />
        <Route path="/settings/admins/:id" element={<LazyRoute component={LazyResellerDetail} skeleton="detail" />} />

        {/* Analytics pages */}
        <Route path="/analytics" element={<LazyRoute component={LazyAnalytics} skeleton="analytics" />} />
        <Route path="/dashboard-pro" element={<LazyRoute component={LazyDashboardPro} skeleton="analytics" />} />

        {/* Form/Config pages */}
        <Route path="/doh" element={<LazyRoute component={LazyDoHSettings} skeleton="form" />} />
        <Route path="/sni-manager" element={<LazyRoute component={LazySNIManager} skeleton="form" />} />
        <Route path="/fingerprint" element={<LazyRoute component={LazyFingerprint} skeleton="form" />} />
        <Route path="/federation" element={<LazyRoute component={LazyFederation} skeleton="form" />} />
        <Route path="/deep-links" element={<LazyRoute component={LazyDeepLinks} skeleton="form" />} />
        <Route path="/config-management" element={<LazyRoute component={LazyConfigManagement} skeleton="form" />} />
        <Route path="/wireguard" element={<LazyRoute component={LazyWireGuard} skeleton="form" />} />
        <Route path="/api-docs" element={<LazyRoute component={LazyApiDocs} skeleton="form" />} />
        <Route path="/tunnels" element={<LazyRoute component={LazyTunnels} skeleton="form" />} />
        <Route path="/geo-exits" element={<LazyRoute component={LazyGeoExits} skeleton="form" />} />
        <Route path="/ip-rotation" element={<LazyRoute component={LazyIPRotation} skeleton="form" />} />
        <Route path="/advanced-security" element={<LazyRoute component={LazyAdvancedSecurity} skeleton="form" />} />

        {/* Routing (complex hybrid) */}
        <Route path="/routing" element={<LazyRoute component={LazyRoutingBalancers} skeleton="list" />} />
        <Route path="/routing/node-rules" element={<LazyRoute component={LazyRouting} skeleton="list" />} />

        {/* Reseller */}
        <Route path="/reseller-dashboard" element={<LazyRoute component={LazyResellerDashboard} skeleton="dashboard" />} />
        <Route path="/reseller-account" element={<LazyRoute component={LazyResellerAccount} skeleton="detail" />} />
        <Route path="/reseller-quota-alerts" element={<LazyRoute component={LazyResellerQuotaAlerts} skeleton="list" />} />
        <Route path="/reseller-payment" element={<LazyRoute component={LazyResellerPaymentSettings} skeleton="form" />} />
        <Route path="/my-quota" element={<Navigate to="/reseller-dashboard" replace />} />

        {/* Wallet & Billing */}
        <Route path="/wallet-billing" element={<LazyRoute component={LazyResellerPlatform} skeleton="list" />} />

        {/* Security (complex page with sub-tabs) */}
        <Route path="/evasion" element={<LazyRoute component={LazySecuritySuite} skeleton="list" />} />
        <Route path="/reality-scanner" element={<Navigate to="/evasion?tab=reality" replace />} />
        <Route path="/clean-ip" element={<Navigate to="/evasion?tab=cleanip" replace />} />
        <Route path="/decoy-website" element={<Navigate to="/evasion?tab=decoy" replace />} />
        <Route path="/probing-protection" element={<Navigate to="/evasion?tab=decoy" replace />} />
        <Route path="/tls-tricks" element={<Navigate to="/evasion?tab=tls" replace />} />

        {/* Phase 3 */}
        <Route path="/performance" element={<LazyRoute component={LazyPerformance} skeleton="dashboard" />} />
        <Route path="/security" element={<LazyRoute component={LazySecurityHardening} skeleton="detail" />} />
        <Route path="/compliance" element={<LazyRoute component={LazyCompliance} skeleton="detail" />} />

        {/* Other */}
        <Route path="/smart-quota" element={<LazyRoute component={LazySmartQuota} skeleton="dashboard" />} />
        <Route path="/relay-chains" element={<LazyRoute component={LazyRelayChains} skeleton="list" />} />

        {/* Redirected routes */}
        <Route path="/outbounds" element={<Navigate to="/routing?tab=outbounds" replace />} />
        <Route path="/routing-packs" element={<Navigate to="/routing?tab=packs" replace />} />
        <Route path="/balancers" element={<Navigate to="/routing?tab=balancers" replace />} />
        <Route path="/admins" element={<Navigate to="/settings?tab=admins" replace />} />
        <Route path="/plans" element={<Navigate to="/wallet-billing?tab=plans" replace />} />
        <Route path="/pending-orders" element={<Navigate to="/wallet-billing?tab=orders" replace />} />

        {/* Settings */}
        <Route path="/settings" element={<LazyRoute component={LazySettings} skeleton="form" />} />

        {/* PHASE 3 Routes */}
        {/* Enterprise v1.4.1 Routes */}
      </Route>
      <Route path="*" element={<NotFound />} />
      {/* Portal (end-user self-service) */}
      <Route path="/portal/login" element={<LazyRoute component={LazyPortalLogin} skeleton="portal" />} />
      <Route element={<LazyRoute component={LazyPortalLayout} skeleton="portal" />}>
        <Route path="/portal/dashboard" element={<LazyRoute component={LazyPortalDashboard} skeleton="portal" />} />
        <Route path="/portal/plans" element={<LazyRoute component={LazyPortalPlans} skeleton="portal" />} />
        <Route path="/portal/referral" element={<LazyRoute component={LazyPortalReferral} skeleton="portal" />} />
        <Route path="/portal/tickets" element={<LazyRoute component={LazyPortalTickets} skeleton="portal" />} />
        <Route path="/portal/speed-test" element={<LazyRoute component={LazyPortalSpeedTest} skeleton="portal" />} />
        <Route path="/portal/guides" element={<LazyRoute component={LazyPortalGuides} skeleton="portal" />} />
        <Route path="/portal/setup" element={<LazyRoute component={LazyPortalSetupWizard} skeleton="portal" />} />
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
