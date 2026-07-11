import { useState } from "react";
import { AlertTriangle } from "lucide-react";
import {
    useSecurityThreats,
    useBlockedThreats,
    useThreatCount,
    useSecurityScore,
    useSecurityPolicy,
    useComplianceValidation,
} from "@/api/security-hooks";
import { Button, Badge } from "@/components/ui";
import { GlassCard } from "@/components/veltrix";
import { useTitle } from "@/lib/useTitle";
import { useAuth } from "@/auth/auth";

export function SecurityHardening() {
    useTitle("Security Hardening");
    const { can } = useAuth();
    const canWrite = can("security:write");
    const [threatFilter, setThreatFilter] = useState("");

    const threats = useSecurityThreats(threatFilter, 20, 0);
    const blocked = useBlockedThreats(10);
    const threatCount = useThreatCount();
    const score = useSecurityScore();
    const policy = useSecurityPolicy();
    const compliance = useComplianceValidation();

    const severityColor = (severity: string) => {
        switch (severity) {
            case "critical":
                return "error";
            case "high":
                return "warning";
            case "medium":
                return "info";
            default:
                return "success";
        }
    };

    const threatTypes = ["sql_injection", "xss", "csrf", "ddos", "brute_force", "anomaly"];

    return (
        <div className="space-y-6 animate-page-enter">
            <div>
                <h1 className="text-2xl font-bold text-fg tracking-tight">Security Hardening</h1>
                <p className="text-sm text-fg-secondary mt-1">Threat detection, policies & compliance</p>
            </div>

            {/* Security Score & Status */}
            <div className="grid grid-cols-1 lg:grid-cols-2 gap-4">
                {score.data && (
                    <GlassCard className="p-6">
                        <div className="flex items-start justify-between">
                            <div>
                                <h2 className="text-lg font-semibold text-fg mb-4">Overall Score</h2>
                                <div className="flex items-center gap-2">
                                    <div className="relative w-24 h-24">
                                        <svg className="w-full h-full transform -rotate-90">
                                            <circle
                                                cx="48"
                                                cy="48"
                                                r="40"
                                                fill="none"
                                                stroke="currentColor"
                                                strokeWidth="2"
                                                className="text-fg-secondary/20"
                                            />
                                            <circle
                                                cx="48"
                                                cy="48"
                                                r="40"
                                                fill="none"
                                                stroke="currentColor"
                                                strokeWidth="2"
                                                strokeDasharray={`${(score.data.overall_score / 100) * 251} 251`}
                                                className={
                                                    score.data.overall_score >= 80
                                                        ? "text-green-500"
                                                        : score.data.overall_score >= 60
                                                            ? "text-yellow-500"
                                                            : "text-red-500"
                                                }
                                            />
                                        </svg>
                                        <div className="absolute inset-0 flex items-center justify-center">
                                            <span className="text-2xl font-bold text-fg">
                                                {score.data.overall_score.toFixed(0)}
                                            </span>
                                        </div>
                                    </div>
                                    <div className="flex-1 space-y-2">
                                        <div>
                                            <p className="text-xs text-fg-secondary">Threat Detection</p>
                                            <div className="flex items-center gap-2">
                                                <div className="flex-1 h-2 bg-fg-secondary/20 rounded overflow-hidden">
                                                    <div
                                                        className="h-full bg-blue-500 transition-all"
                                                        style={{ width: `${score.data.components.threat_detection}%` }}
                                                    />
                                                </div>
                                                <span className="text-xs font-bold text-fg w-8">
                                                    {score.data.components.threat_detection}%
                                                </span>
                                            </div>
                                        </div>
                                        <div>
                                            <p className="text-xs text-fg-secondary">Compliance</p>
                                            <div className="flex items-center gap-2">
                                                <div className="flex-1 h-2 bg-fg-secondary/20 rounded overflow-hidden">
                                                    <div
                                                        className="h-full bg-green-500 transition-all"
                                                        style={{ width: `${score.data.components.policy_compliance}%` }}
                                                    />
                                                </div>
                                                <span className="text-xs font-bold text-fg w-8">
                                                    {score.data.components.policy_compliance}%
                                                </span>
                                            </div>
                                        </div>
                                        <div>
                                            <p className="text-xs text-fg-secondary">Encryption</p>
                                            <div className="flex items-center gap-2">
                                                <div className="flex-1 h-2 bg-fg-secondary/20 rounded overflow-hidden">
                                                    <div
                                                        className="h-full bg-purple-500 transition-all"
                                                        style={{ width: `${score.data.components.encryption}%` }}
                                                    />
                                                </div>
                                                <span className="text-xs font-bold text-fg w-8">
                                                    {score.data.components.encryption}%
                                                </span>
                                            </div>
                                        </div>
                                    </div>
                                </div>
                            </div>
                        </div>
                    </GlassCard>
                )}

                {/* Threat Count Cards */}
                <div className="grid grid-cols-2 gap-3">
                    <GlassCard>
                        <div>
                            <p className="text-xs text-fg-secondary uppercase tracking-wide">Total Threats</p>
                            <p className="text-3xl font-bold text-fg mt-2">{threatCount.data?.count || 0}</p>
                        </div>
                    </GlassCard>
                    <GlassCard>
                        <div>
                            <p className="text-xs text-fg-secondary uppercase tracking-wide">Blocked</p>
                            <p className="text-3xl font-bold text-green-500 mt-2">{blocked.data?.threats.length || 0}</p>
                        </div>
                    </GlassCard>
                </div>
            </div>

            {/* Security Threats */}
            <GlassCard className="p-6">
                <div className="flex items-center justify-between mb-4">
                    <h2 className="text-lg font-semibold text-fg flex items-center gap-2">
                        <AlertTriangle className="w-5 h-5 text-orange-500" />
                        Recent Threats
                    </h2>
                </div>

                {/* Threat Type Filter */}
                <div className="flex gap-2 mb-4 overflow-x-auto pb-2">
                    <Button
                        size="sm"
                        variant={threatFilter === "" ? "primary" : "ghost"}
                        onClick={() => setThreatFilter("")}
                    >
                        All
                    </Button>
                    {threatTypes.map((type) => (
                        <Button
                            key={type}
                            size="sm"
                            variant={threatFilter === type ? "primary" : "ghost"}
                            onClick={() => setThreatFilter(type)}
                        >
                            {type}
                        </Button>
                    ))}
                </div>

                {/* Threats List */}
                <div className="space-y-2 max-h-96 overflow-y-auto">
                    {threats.isLoading && <p className="text-sm text-fg-secondary">Loading threats...</p>}
                    {threats.data?.threats && threats.data.threats.length === 0 && (
                        <p className="text-sm text-fg-secondary">No threats detected</p>
                    )}
                    {threats.data?.threats.map((threat) => (
                        <div
                            key={threat.id}
                            className="p-3 bg-fg/5 rounded border border-fg-secondary/20 hover:bg-fg/10 transition"
                        >
                            <div className="flex items-start justify-between gap-2">
                                <div className="flex-1">
                                    <div className="flex items-center gap-2">
                                        <div className="flex h-6 items-center gap-1 rounded px-2.5 text-xs font-medium"
                                            style={{ backgroundColor: severityColor(threat.severity) === "error" ? 'rgba(239, 68, 68, 0.1)' : severityColor(threat.severity) === "warning" ? 'rgba(202, 138, 4, 0.1)' : 'rgba(34, 197, 94, 0.1)', color: severityColor(threat.severity) === "error" ? 'rgb(239, 68, 68)' : severityColor(threat.severity) === "warning" ? 'rgb(202, 138, 4)' : 'rgb(34, 197, 94)' }}>
                                            {threat.severity}
                                        </div>
                                        <p className="text-sm font-medium text-fg">{threat.threat_type}</p>
                                        <div className="flex h-6 items-center gap-1 rounded px-2.5 text-xs font-medium"
                                            style={{ backgroundColor: threat.blocked ? 'rgba(34, 197, 94, 0.1)' : 'rgba(202, 138, 4, 0.1)', color: threat.blocked ? 'rgb(34, 197, 94)' : 'rgb(202, 138, 4)' }}>
                                            {threat.blocked ? "Blocked" : "Detected"}
                                        </div>
                                    </div>
                                    <p className="text-xs text-fg-secondary mt-1">
                                        {threat.source_ip} → {threat.target_path}
                                    </p>
                                    {threat.payload && (
                                        <p className="text-xs text-fg-secondary font-mono mt-1 truncate">
                                            {threat.payload.substring(0, 60)}...
                                        </p>
                                    )}
                                    <p className="text-xs text-fg-secondary mt-1">{threat.created_at}</p>
                                </div>
                            </div>
                        </div>
                    ))}
                </div>
            </GlassCard>

            {/* Security Policy */}
            {policy.data && (
                <GlassCard className="p-6">
                    <div className="flex items-center justify-between mb-4">
                        <h2 className="text-lg font-semibold text-fg">Security Policy</h2>
                        {canWrite && (
                            <Button size="sm" variant="outline">
                                Edit
                            </Button>
                        )}
                    </div>

                    <div className="grid grid-cols-2 md:grid-cols-3 gap-4">
                        {[
                            { label: "CSRF Protection", value: policy.data.enable_csrf_protection },
                            { label: "XSS Protection", value: policy.data.enable_xss_protection },
                            { label: "SQL Injection Detection", value: policy.data.enable_sql_injection_detection },
                            { label: "DDoS Protection", value: policy.data.enable_ddos_protection },
                            { label: "Require HTTPS", value: policy.data.require_https },
                            { label: "Encryption Required", value: policy.data.encryption_required },
                            { label: "Require MFA", value: policy.data.require_mfa },
                        ].map((item) => (
                            <div key={item.label} className="p-3 bg-fg/5 rounded border border-fg-secondary/20">
                                <p className="text-xs text-fg-secondary mb-2">{item.label}</p>
                                <div className="flex h-6 items-center gap-1 rounded px-2.5 text-xs font-medium"
                                    style={{ backgroundColor: item.value ? 'rgba(34, 197, 94, 0.1)' : 'rgba(107, 114, 128, 0.1)', color: item.value ? 'rgb(34, 197, 94)' : 'rgb(107, 114, 128)' }}>
                                    {item.value ? "Enabled" : "Disabled"}
                                </div>
                            </div>
                        ))}
                    </div>

                    <div className="grid grid-cols-3 gap-4 mt-4 pt-4 border-t border-fg-secondary/20">
                        <div>
                            <p className="text-xs text-fg-secondary">Min Password Length</p>
                            <p className="text-lg font-bold text-fg mt-1">{policy.data.min_password_length}</p>
                        </div>
                        <div>
                            <p className="text-xs text-fg-secondary">Session Timeout</p>
                            <p className="text-lg font-bold text-fg mt-1">{policy.data.session_timeout}m</p>
                        </div>
                        <div>
                            <p className="text-xs text-fg-secondary">Max Concurrent Sessions</p>
                            <p className="text-lg font-bold text-fg mt-1">{policy.data.max_concurrent_sessions}</p>
                        </div>
                    </div>
                </GlassCard>
            )}

            {/* Compliance Status */}
            {compliance.data && (
                <GlassCard className="p-6">
                    <h2 className="text-lg font-semibold text-fg mb-4">
                        Compliance Status
                    </h2>
                    <div className="space-y-2">
                        {Object.entries(compliance.data.compliant).map(([key, value]: [string, unknown]) => (
                            <div key={key} className="flex items-center justify-between p-2 hover:bg-fg/5 rounded">
                                <span className="text-sm capitalize text-fg">{key.replace(/_/g, " ")}</span>
                                <Badge color={value ? "success" : "error"}>
                                    {value ? "Compliant" : "Non-compliant"}
                                </Badge>
                            </div>
                        ))}
                    </div>

                    {compliance.data.issues.length > 0 && (
                        <div className="mt-4 p-3 bg-red-500/10 border border-red-500/20 rounded">
                            <p className="text-xs font-semibold text-red-500 mb-2">Issues Found:</p>
                            <ul className="space-y-1">
                                {compliance.data.issues.map((issue: string, i: number) => (
                                    <li key={i} className="text-xs text-fg-secondary">
                                        • {issue}
                                    </li>
                                ))}
                            </ul>
                        </div>
                    )}
                </GlassCard>
            )}
        </div>
    );
}
