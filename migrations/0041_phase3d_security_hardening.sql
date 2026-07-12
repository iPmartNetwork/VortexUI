-- +goose Up
-- PHASE 3D: Security Hardening & Defense Schema
-- Migration: 0041_phase3d_security_hardening.sql

-- Security Threats Table
CREATE TABLE IF NOT EXISTS security_threats (
    id UUID PRIMARY KEY,
    threat_type VARCHAR(100) NOT NULL CHECK (threat_type IN ('sql_injection', 'xss', 'csrf', 'ddos', 'brute_force', 'malware', 'anomaly')),
    severity VARCHAR(50) NOT NULL CHECK (severity IN ('low', 'medium', 'high', 'critical')),
    source_ip INET,
    target_path VARCHAR(500),
    payload TEXT,
    detection_method VARCHAR(100),
    blocked BOOLEAN DEFAULT FALSE,
    block_reason VARCHAR(500),
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_security_threats_type ON security_threats(threat_type);
CREATE INDEX IF NOT EXISTS idx_security_threats_severity ON security_threats(severity);
CREATE INDEX IF NOT EXISTS idx_security_threats_source_ip ON security_threats(source_ip);
CREATE INDEX IF NOT EXISTS idx_security_threats_created_at ON security_threats(created_at DESC);

-- IP Reputation Table
CREATE TABLE IF NOT EXISTS ip_reputation (
    id UUID PRIMARY KEY,
    ip_address INET NOT NULL UNIQUE,
    reputation_score FLOAT DEFAULT 50.0,
    threat_level VARCHAR(50) DEFAULT 'neutral' CHECK (threat_level IN ('trusted', 'neutral', 'suspicious', 'malicious')),
    failed_logins INTEGER DEFAULT 0,
    blocked_requests INTEGER DEFAULT 0,
    country VARCHAR(100),
    is_proxy BOOLEAN DEFAULT FALSE,
    is_tor BOOLEAN DEFAULT FALSE,
    is_vpn BOOLEAN DEFAULT FALSE,
    last_seen TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_ip_reputation_threat_level ON ip_reputation(threat_level);
CREATE INDEX IF NOT EXISTS idx_ip_reputation_last_seen ON ip_reputation(last_seen DESC);

-- Security Policies Table
CREATE TABLE IF NOT EXISTS security_policies (
    id UUID PRIMARY KEY,
    name VARCHAR(255) NOT NULL UNIQUE,
    description TEXT,
    enable_csrf_protection BOOLEAN DEFAULT TRUE,
    enable_xss_protection BOOLEAN DEFAULT TRUE,
    enable_sql_injection_detection BOOLEAN DEFAULT TRUE,
    enable_ddos_protection BOOLEAN DEFAULT TRUE,
    require_https BOOLEAN DEFAULT TRUE,
    encryption_required BOOLEAN DEFAULT TRUE,
    min_password_length INTEGER DEFAULT 12,
    require_mfa BOOLEAN DEFAULT TRUE,
    session_timeout INTEGER DEFAULT 30,  -- minutes
    max_concurrent_sessions INTEGER DEFAULT 5,
    allowed_cors_origins TEXT[] DEFAULT ARRAY[]::TEXT[],
    require_content_security BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Default Security Policy
INSERT INTO security_policies (
    id, name, description, enable_csrf_protection, enable_xss_protection,
    enable_sql_injection_detection, enable_ddos_protection, require_https,
    encryption_required, min_password_length, require_mfa, session_timeout,
    max_concurrent_sessions, created_at, updated_at
) VALUES (
    '00000000-0000-0000-0000-000000000001'::UUID,
    'Default Policy',
    'System default security policy',
    TRUE, TRUE, TRUE, TRUE, TRUE, TRUE, 12, TRUE, 30, 5, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP
) ON CONFLICT DO NOTHING;

-- Security Headers Table
CREATE TABLE IF NOT EXISTS security_headers (
    id UUID PRIMARY KEY,
    name VARCHAR(255) NOT NULL UNIQUE,
    value TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Default Security Headers
INSERT INTO security_headers (id, name, value, created_at) VALUES
('00000000-0000-0000-0000-000000000010'::UUID, 'X-Frame-Options', 'DENY', CURRENT_TIMESTAMP),
('00000000-0000-0000-0000-000000000011'::UUID, 'X-Content-Type-Options', 'nosniff', CURRENT_TIMESTAMP),
('00000000-0000-0000-0000-000000000012'::UUID, 'X-XSS-Protection', '1; mode=block', CURRENT_TIMESTAMP),
('00000000-0000-0000-0000-000000000013'::UUID, 'Strict-Transport-Security', 'max-age=31536000; includeSubDomains', CURRENT_TIMESTAMP),
('00000000-0000-0000-0000-000000000014'::UUID, 'Content-Security-Policy', 'default-src ''self''; script-src ''self''', CURRENT_TIMESTAMP),
('00000000-0000-0000-0000-000000000015'::UUID, 'Referrer-Policy', 'strict-origin-when-cross-origin', CURRENT_TIMESTAMP)
ON CONFLICT DO NOTHING;

-- Encryption Keys Table
CREATE TABLE IF NOT EXISTS encryption_keys (
    id UUID PRIMARY KEY,
    key_type VARCHAR(100) NOT NULL,
    key_algorithm VARCHAR(100),
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    rotated_at TIMESTAMP,
    expires_at TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_encryption_keys_is_active ON encryption_keys(is_active);
CREATE INDEX IF NOT EXISTS idx_encryption_keys_key_type ON encryption_keys(key_type);

-- Anomaly Detection Table
CREATE TABLE IF NOT EXISTS anomaly_detections (
    id UUID PRIMARY KEY,
    admin_id UUID REFERENCES admins(id) ON DELETE CASCADE,
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    anomaly_type VARCHAR(100) NOT NULL,
    severity VARCHAR(50) DEFAULT 'medium' CHECK (severity IN ('low', 'medium', 'high', 'critical')),
    description TEXT,
    context JSONB DEFAULT '{}',
    flagged BOOLEAN DEFAULT FALSE,
    investigated BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_anomaly_detections_anomaly_type ON anomaly_detections(anomaly_type);
CREATE INDEX IF NOT EXISTS idx_anomaly_detections_severity ON anomaly_detections(severity);
CREATE INDEX IF NOT EXISTS idx_anomaly_detections_user_id ON anomaly_detections(user_id);
CREATE INDEX IF NOT EXISTS idx_anomaly_detections_created_at ON anomaly_detections(created_at DESC);

-- Vulnerability Assessment Table
CREATE TABLE IF NOT EXISTS vulnerability_assessments (
    id UUID PRIMARY KEY,
    scan_type VARCHAR(100) NOT NULL,
    status VARCHAR(50) DEFAULT 'pending' CHECK (status IN ('pending', 'running', 'completed', 'failed')),
    vuln_count INTEGER DEFAULT 0,
    critical_count INTEGER DEFAULT 0,
    high_count INTEGER DEFAULT 0,
    medium_count INTEGER DEFAULT 0,
    low_count INTEGER DEFAULT 0,
    cves_found TEXT[] DEFAULT ARRAY[]::TEXT[],
    report_path VARCHAR(500),
    started_at TIMESTAMP NOT NULL,
    completed_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_vulnerability_assessments_scan_type ON vulnerability_assessments(scan_type);
CREATE INDEX IF NOT EXISTS idx_vulnerability_assessments_status ON vulnerability_assessments(status);
CREATE INDEX IF NOT EXISTS idx_vulnerability_assessments_created_at ON vulnerability_assessments(created_at DESC);

-- WAF Rules Table
CREATE TABLE IF NOT EXISTS waf_rules (
    id UUID PRIMARY KEY,
    name VARCHAR(255) NOT NULL UNIQUE,
    description TEXT,
    rule_type VARCHAR(100) NOT NULL,
    pattern TEXT,
    action VARCHAR(50) NOT NULL DEFAULT 'log' CHECK (action IN ('allow', 'block', 'challenge', 'log')),
    enabled BOOLEAN DEFAULT TRUE,
    priority INTEGER DEFAULT 100,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_waf_rules_enabled ON waf_rules(enabled);
CREATE INDEX IF NOT EXISTS idx_waf_rules_priority ON waf_rules(priority);

-- Default WAF Rules
INSERT INTO waf_rules (id, name, description, rule_type, pattern, action, enabled, priority, created_at, updated_at) VALUES
('00000000-0000-0000-0000-000000000020'::UUID, 'Block SQL Injection', 'Detect SQL injection attempts', 'pattern', 'union.*select|drop.*table|insert.*into', 'block', TRUE, 10, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
('00000000-0000-0000-0000-000000000021'::UUID, 'Block XSS Attempts', 'Detect XSS attacks', 'pattern', '<script|javascript:|onerror=|onclick=', 'block', TRUE, 11, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
('00000000-0000-0000-0000-000000000022'::UUID, 'Block Path Traversal', 'Detect directory traversal', 'pattern', '\.\./|\.\.\\.', 'block', TRUE, 12, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
ON CONFLICT DO NOTHING;

-- Incident Response Table
CREATE TABLE IF NOT EXISTS incident_responses (
    id UUID PRIMARY KEY,
    incident_type VARCHAR(100) NOT NULL,
    title VARCHAR(500) NOT NULL,
    description TEXT,
    severity VARCHAR(50) NOT NULL CHECK (severity IN ('low', 'medium', 'high', 'critical')),
    status VARCHAR(50) DEFAULT 'open' CHECK (status IN ('open', 'investigating', 'contained', 'resolved', 'closed')),
    threat_actors TEXT[] DEFAULT ARRAY[]::TEXT[],
    impacted_systems TEXT[] DEFAULT ARRAY[]::TEXT[],
    root_cause TEXT,
    remediation TEXT,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    resolved_at TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_incident_responses_severity ON incident_responses(severity);
CREATE INDEX IF NOT EXISTS idx_incident_responses_status ON incident_responses(status);
CREATE INDEX IF NOT EXISTS idx_incident_responses_created_at ON incident_responses(created_at DESC);

-- Auto-update trigger
DROP TRIGGER IF EXISTS trigger_security_policies_updated_at ON security_policies;
CREATE TRIGGER trigger_security_policies_updated_at
BEFORE UPDATE ON security_policies
FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column();

DROP TRIGGER IF EXISTS trigger_waf_rules_updated_at ON waf_rules;
CREATE TRIGGER trigger_waf_rules_updated_at
BEFORE UPDATE ON waf_rules
FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column();

DROP TRIGGER IF EXISTS trigger_ip_reputation_updated_at ON ip_reputation;
CREATE TRIGGER trigger_ip_reputation_updated_at
BEFORE UPDATE ON ip_reputation
FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column();
