import { CheckCircle, AlertCircle } from "lucide-react";
import { GlassCard } from "@/components/veltrix";
import { useTitle } from "@/lib/useTitle";

export function Compliance() {
    useTitle("Compliance");

    // Placeholder data - replace with actual API calls
    const complianceData = {
        compliant: {
            csrf_protection: true,
            xss_protection: true,
            sql_injection_detection: true,
            ddos_protection: true,
            https_required: false,
            encryption_required: true,
            mfa_required: true,
        },
        issues: ["HTTPS is not enabled", "Consider enabling HTTPS for all admin communications"],
    };

    const policyData = {
        enable_csrf_protection: true,
        enable_xss_protection: true,
        enable_sql_injection_detection: true,
        enable_ddos_protection: true,
        require_https: false,
        encryption_required: true,
        require_mfa: true,
        require_content_security: true,
        min_password_length: 12,
        session_timeout: 30,
        max_concurrent_sessions: 5,
    };

    const isCompliant = Object.values(complianceData.compliant).every(Boolean);
    const complianceScore = Math.round(
        (Object.values(complianceData.compliant).filter(Boolean).length /
            Object.values(complianceData.compliant).length) *
        100
    );

    return (
        <div className="space-y-6 animate-page-enter">
            <div>
                <h1 className="text-2xl font-bold text-fg tracking-tight">Compliance</h1>
                <p className="text-sm text-fg-secondary mt-1">System compliance status and policy enforcement</p>
            </div>

            {/* Overall Status */}
            <div className="grid grid-cols-1 lg:grid-cols-3 gap-4">
                <GlassCard className="p-6 lg:col-span-2">
                    <div className="flex items-start justify-between">
                        <div>
                            <h2 className="text-lg font-semibold text-fg mb-2">Overall Compliance</h2>
                            <p className="text-sm text-fg-secondary mb-4">
                                {isCompliant
                                    ? "✓ Your system is fully compliant with all policies"
                                    : "⚠ There are non-compliant areas that need attention"}
                            </p>
                        </div>
                        <div className="flex items-center gap-3">
                            {isCompliant ? (
                                <CheckCircle className="w-12 h-12 text-green-500" />
                            ) : (
                                <AlertCircle className="w-12 h-12 text-yellow-500" />
                            )}
                            <div>
                                <p className="text-4xl font-bold text-fg">{complianceScore.toFixed(0)}%</p>
                                <p className="text-xs text-fg-secondary">Compliant</p>
                            </div>
                        </div>
                    </div>

                    {/* Compliance Bar */}
                    <div className="mt-4 h-3 bg-fg/10 rounded-full overflow-hidden">
                        <div
                            className={`h-full transition-all ${complianceScore >= 80 ? "bg-green-500" : complianceScore >= 50 ? "bg-yellow-500" : "bg-red-500"
                                }`}
                            style={{ width: `${complianceScore}%` }}
                        />
                    </div>
                </GlassCard>

                <GlassCard className="p-6">
                    <div>
                        <p className="text-xs text-fg-secondary uppercase tracking-wide mb-2">Status Summary</p>
                        <div className="space-y-2">
                            <div className="flex items-center justify-between">
                                <span className="text-sm text-fg">✓ Compliant</span>
                                <span className="font-bold text-green-500">
                                    {Object.values(complianceData.compliant).filter(Boolean).length}
                                </span>
                            </div>
                            <div className="flex items-center justify-between">
                                <span className="text-sm text-fg">✗ Non-Compliant</span>
                                <span className="font-bold text-red-500">
                                    {Object.values(complianceData.compliant).filter((v) => !v).length}
                                </span>
                            </div>
                        </div>
                    </div>
                </GlassCard>
            </div>

            {/* Detailed Compliance Items */}
            <GlassCard className="p-6">
                <h2 className="text-lg font-semibold text-fg mb-4">Compliance Checklist</h2>
                <div className="space-y-3">
                    {complianceData.compliant &&
                        Object.entries(complianceData.compliant).map(([key, value]) => (
                            <div
                                key={key}
                                className="flex items-center justify-between p-4 rounded-lg border border-fg-secondary/20 hover:bg-fg/5 transition"
                            >
                                <div className="flex items-center gap-3">
                                    {value ? (
                                        <CheckCircle className="w-5 h-5 text-green-500" />
                                    ) : (
                                        <AlertCircle className="w-5 h-5 text-red-500" />
                                    )}
                                    <div>
                                        <p className="font-medium text-fg capitalize">{key.replace(/_/g, " ")}</p>
                                        <p className="text-xs text-fg-secondary">
                                            {value ? "All requirements met" : "Action required"}
                                        </p>
                                    </div>
                                </div>
                                <div className="flex h-6 items-center gap-1 rounded px-2.5 text-xs font-medium"
                                    style={{ backgroundColor: value ? 'rgba(34, 197, 94, 0.1)' : 'rgba(239, 68, 68, 0.1)', color: value ? 'rgb(34, 197, 94)' : 'rgb(239, 68, 68)' }}>
                                    {value ? "Compliant" : "Non-Compliant"}
                                </div>
                            </div>
                        ))}
                </div>
            </GlassCard>

            {/* Issues */}
            {complianceData.issues && complianceData.issues.length > 0 && (
                <GlassCard className="p-6 border-l-4 border-red-500">
                    <h2 className="text-lg font-semibold text-fg mb-4 flex items-center gap-2">
                        <AlertCircle className="w-5 h-5 text-red-500" />
                        Issues Detected
                    </h2>
                    <div className="space-y-3">
                        {complianceData.issues.map((issue: string, i: number) => (
                            <div key={i} className="p-3 bg-red-500/10 rounded-lg border border-red-500/20">
                                <p className="text-sm text-fg">• {issue}</p>
                            </div>
                        ))}
                    </div>
                </GlassCard>
            )}

            {/* Security Policy Status */}
            {policyData && (
                <GlassCard className="p-6">
                    <h2 className="text-lg font-semibold text-fg mb-4 flex items-center gap-2">
                        Security Policies
                    </h2>

                    <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                        {[
                            { label: "CSRF Protection", enabled: policyData.enable_csrf_protection },
                            { label: "XSS Protection", enabled: policyData.enable_xss_protection },
                            { label: "SQL Injection Detection", enabled: policyData.enable_sql_injection_detection },
                            { label: "DDoS Protection", enabled: policyData.enable_ddos_protection },
                            { label: "HTTPS Required", enabled: policyData.require_https },
                            { label: "Encryption Required", enabled: policyData.encryption_required },
                            { label: "MFA Required", enabled: policyData.require_mfa },
                            { label: "Content Security", enabled: policyData.require_content_security },
                        ].map((item) => (
                            <div key={item.label} className="flex items-center justify-between p-3 rounded-lg hover:bg-fg/5">
                                <span className="text-sm text-fg">{item.label}</span>
                                <div className="flex h-6 items-center gap-1 rounded px-2.5 text-xs font-medium"
                                    style={{ backgroundColor: item.enabled ? 'rgba(34, 197, 94, 0.1)' : 'rgba(107, 114, 128, 0.1)', color: item.enabled ? 'rgb(34, 197, 94)' : 'rgb(107, 114, 128)' }}>
                                    {item.enabled ? "Enabled" : "Disabled"}
                                </div>
                            </div>
                        ))}
                    </div>

                    <div className="mt-4 pt-4 border-t border-fg-secondary/20 space-y-2">
                        <div className="flex justify-between p-2">
                            <span className="text-sm text-fg-secondary">Min Password Length</span>
                            <span className="font-bold text-fg">{policyData.min_password_length} chars</span>
                        </div>
                        <div className="flex justify-between p-2">
                            <span className="text-sm text-fg-secondary">Session Timeout</span>
                            <span className="font-bold text-fg">{policyData.session_timeout} minutes</span>
                        </div>
                        <div className="flex justify-between p-2">
                            <span className="text-sm text-fg-secondary">Max Concurrent Sessions</span>
                            <span className="font-bold text-fg">{policyData.max_concurrent_sessions}</span>
                        </div>
                    </div>
                </GlassCard>
            )}

            {/* Recommendations */}
            <GlassCard className="p-6 bg-blue-500/10 border-blue-500/20">
                <h2 className="text-lg font-semibold text-fg mb-3 flex items-center gap-2">
                    <CheckCircle className="w-5 h-5 text-blue-500" />
                    Recommendations
                </h2>
                <ul className="space-y-2 text-sm text-fg-secondary">
                    <li>✓ Review security policies quarterly</li>
                    <li>✓ Monitor audit logs for suspicious activities</li>
                    <li>✓ Keep all security patches up to date</li>
                    <li>✓ Enable MFA for all admin accounts</li>
                    <li>✓ Use strong, unique passwords</li>
                    <li>✓ Regularly backup your data</li>
                </ul>
            </GlassCard>
        </div>
    );
}
