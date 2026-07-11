import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { api } from "./client";

// ========== PHASE 3B: Performance Optimization Hooks ==========

export interface QueryMetric {
    id: string;
    query: string;
    execution_time_ms: number;
    rows_affected: number;
    is_slow: boolean;
    index_used: string;
    slow_threshold_ms: number;
    created_at: string;
}

export interface CacheStats {
    hit_count: number;
    miss_count: number;
    hit_rate: number;
    current_size: number;
    max_size: number;
    eviction_count: number;
}

export interface RateLimitRule {
    id: string;
    name: string;
    endpoint: string;
    method: string;
    requests_per_min: number;
    burst_size: number;
    enabled: boolean;
    created_at: string;
}

export interface PerformanceHealth {
    cache: CacheStats;
    slow_queries_last_hour: number;
    avg_query_time_ms: number;
    connections_active: number;
    connections_max: number;
}

export function usePerformanceHealth() {
    return useQuery({
        queryKey: ["performance-health"],
        queryFn: () => api<PerformanceHealth>("/api/performance/health"),
        refetchInterval: 10000,
    });
}

export function useSlowQueries(limit = 50, offset = 0) {
    return useQuery({
        queryKey: ["slow-queries", limit, offset],
        queryFn: () =>
            api<{ queries: QueryMetric[]; count: number }>("/api/performance/queries/slow", {
                query: { limit, offset },
            }),
    });
}

export function useQueryStats() {
    return useQuery({
        queryKey: ["query-stats"],
        queryFn: () =>
            api<{
                avg_time_ms: number;
                max_time_ms: number;
                min_time_ms: number;
                total_queries: number;
            }>("/api/performance/queries/stats"),
    });
}

export function usePerformanceAlerts() {
    return useQuery({
        queryKey: ["performance-alerts"],
        queryFn: () => api<{ alerts: any[] }>("/api/performance/alerts"),
        refetchInterval: 15000,
    });
}

export function useResolveAlert() {
    const qc = useQueryClient();
    return useMutation({
        mutationFn: (alertId: string) =>
            api(`/api/performance/alerts/${alertId}/resolve`, { method: "PUT" }),
        onSuccess: () => {
            qc.invalidateQueries({ queryKey: ["performance-alerts"] });
        },
    });
}

export function useRateLimitRules() {
    return useQuery({
        queryKey: ["rate-limit-rules"],
        queryFn: () => api<{ rules: RateLimitRule[] }>("/api/performance/rate-limits/rules"),
    });
}

export function useRateLimitViolations(clientIp: string, minutesBack = 60) {
    return useQuery({
        queryKey: ["rate-limit-violations", clientIp, minutesBack],
        queryFn: () =>
            api<{ violations: any[] }>("/api/performance/rate-limits/violations", {
                query: { client_ip: clientIp, minutes_back: minutesBack },
            }),
        enabled: !!clientIp,
    });
}

// ========== PHASE 3D: Security Hardening & Defense Hooks ==========

export interface SecurityThreat {
    id: string;
    threat_type: string;
    severity: "low" | "medium" | "high" | "critical";
    source_ip: string;
    target_path: string;
    payload?: string;
    detection_method: string;
    blocked: boolean;
    created_at: string;
}

export interface IPReputation {
    id: string;
    ip_address: string;
    reputation_score: number;
    threat_level: "trusted" | "neutral" | "suspicious" | "malicious";
    failed_logins: number;
    blocked_requests: number;
    is_proxy: boolean;
    is_tor: boolean;
    is_vpn: boolean;
    last_seen: string;
}

export interface SecurityPolicy {
    id: string;
    name: string;
    enable_csrf_protection: boolean;
    enable_xss_protection: boolean;
    enable_sql_injection_detection: boolean;
    enable_ddos_protection: boolean;
    require_https: boolean;
    encryption_required: boolean;
    min_password_length: number;
    require_mfa: boolean;
    session_timeout: number;
    max_concurrent_sessions: number;
}

export interface WAFRule {
    id: string;
    name: string;
    rule_type: string;
    pattern: string;
    action: "allow" | "block" | "challenge" | "log";
    enabled: boolean;
    priority: number;
}

export interface SecurityScore {
    overall_score: number;
    components: {
        threat_detection: number;
        policy_compliance: number;
        encryption: number;
        access_control: number;
    };
}

export function useSecurityThreats(threatType = "", limit = 50, offset = 0) {
    return useQuery({
        queryKey: ["security-threats", threatType, limit, offset],
        queryFn: () =>
            api<{ threats: SecurityThreat[]; count: number }>("/api/security/threats", {
                query: { type: threatType, limit, offset },
            }),
        refetchInterval: 15000,
    });
}

export function useBlockedThreats(limit = 50) {
    return useQuery({
        queryKey: ["blocked-threats", limit],
        queryFn: () =>
            api<{ threats: SecurityThreat[] }>("/api/security/threats/blocked", {
                query: { limit },
            }),
        refetchInterval: 20000,
    });
}

export function useThreatCount() {
    return useQuery({
        queryKey: ["threat-count"],
        queryFn: () => api<{ count: number }>("/api/security/threats/count"),
        refetchInterval: 30000,
    });
}

export function useSecurityPolicy() {
    return useQuery({
        queryKey: ["security-policy"],
        queryFn: () => api<SecurityPolicy>("/api/security/policy"),
    });
}

export function useUpdateSecurityPolicy() {
    const qc = useQueryClient();
    return useMutation({
        mutationFn: (policy: Partial<SecurityPolicy>) =>
            api<SecurityPolicy>("/api/security/policy", { method: "PUT", body: policy }),
        onSuccess: () => {
            qc.invalidateQueries({ queryKey: ["security-policy"] });
        },
    });
}

export function useSecurityScore() {
    return useQuery({
        queryKey: ["security-score"],
        queryFn: () => api<SecurityScore>("/api/security/score"),
        refetchInterval: 60000,
    });
}

export function useComplianceValidation() {
    return useQuery({
        queryKey: ["compliance-validation"],
        queryFn: () =>
            api<{ compliant: Record<string, boolean>; issues: string[] }>(
                "/api/security/compliance/validate",
            ),
        refetchInterval: 60000,
    });
}

export function useSecurityIPReputation(ipAddress: string) {
    return useQuery({
        queryKey: ["ip-reputation", ipAddress],
        queryFn: () => api<IPReputation>(`/api/security/reputation/${ipAddress}`),
        enabled: !!ipAddress,
    });
}
