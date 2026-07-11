-- PHASE 3C: Audit & Compliance Infrastructure
-- Migration for comprehensive audit event logging, compliance tracking, and report generation

-- Audit Events Table: Records all security-related events
CREATE TABLE IF NOT EXISTS audit_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    admin_id UUID NOT NULL REFERENCES admins(id) ON DELETE CASCADE,
    event_type VARCHAR(50) NOT NULL,
    severity VARCHAR(20) NOT NULL CHECK (severity IN ('info', 'warning', 'error', 'critical')),
    target_type VARCHAR(100) NOT NULL,
    target_id UUID,
    description TEXT NOT NULL,
    ip_address INET NOT NULL,
    user_agent TEXT,
    old_value TEXT,
    new_value TEXT,
    status VARCHAR(20) NOT NULL DEFAULT 'pending' CHECK (status IN ('success', 'failure', 'pending', 'alert')),
    error_message TEXT,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_audit_events_admin_id ON audit_events(admin_id);
CREATE INDEX idx_audit_events_event_type ON audit_events(event_type);
CREATE INDEX idx_audit_events_severity ON audit_events(severity);
CREATE INDEX idx_audit_events_created_at ON audit_events(created_at);
CREATE INDEX idx_audit_events_admin_created ON audit_events(admin_id, created_at DESC);
CREATE INDEX idx_audit_events_type_created ON audit_events(event_type, created_at DESC);

-- Compliance Events Table: Tracks compliance status and verifications
CREATE TABLE IF NOT EXISTS compliance_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    event_type VARCHAR(50) NOT NULL,
    category VARCHAR(100) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'pending_review' CHECK (status IN ('compliant', 'non_compliant', 'pending_review')),
    description TEXT NOT NULL,
    evidence TEXT,
    audit_event_id UUID REFERENCES audit_events(id) ON DELETE SET NULL,
    admin_id UUID NOT NULL REFERENCES admins(id) ON DELETE CASCADE,
    verified_at TIMESTAMP,
    expires_at TIMESTAMP,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_compliance_events_event_type ON compliance_events(event_type);
CREATE INDEX idx_compliance_events_status ON compliance_events(status);
CREATE INDEX idx_compliance_events_admin_id ON compliance_events(admin_id);
CREATE INDEX idx_compliance_events_expires_at ON compliance_events(expires_at);
CREATE INDEX idx_compliance_events_created_at ON compliance_events(created_at DESC);

-- Audit Reports Table: Stores generated audit reports
CREATE TABLE IF NOT EXISTS audit_reports (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    title VARCHAR(255) NOT NULL,
    description TEXT,
    report_type VARCHAR(50) NOT NULL DEFAULT 'custom' CHECK (report_type IN ('daily', 'weekly', 'monthly', 'custom', 'incident', 'compliance')),
    status VARCHAR(20) NOT NULL DEFAULT 'draft' CHECK (status IN ('draft', 'pending_review', 'approved', 'rejected', 'exported')),
    start_date TIMESTAMP NOT NULL,
    end_date TIMESTAMP NOT NULL,
    scope VARCHAR(100) NOT NULL,
    filters JSONB DEFAULT '{}',
    event_count INTEGER NOT NULL DEFAULT 0,
    file_path VARCHAR(500),
    created_by UUID NOT NULL REFERENCES admins(id) ON DELETE CASCADE,
    approved_by UUID REFERENCES admins(id) ON DELETE SET NULL,
    approved_at TIMESTAMP,
    exported_at TIMESTAMP,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_audit_reports_report_type ON audit_reports(report_type);
CREATE INDEX idx_audit_reports_status ON audit_reports(status);
CREATE INDEX idx_audit_reports_created_by ON audit_reports(created_by);
CREATE INDEX idx_audit_reports_created_at ON audit_reports(created_at DESC);
CREATE INDEX idx_audit_reports_date_range ON audit_reports(start_date, end_date);

-- Audit Policies Table: Defines audit retention and compliance rules
CREATE TABLE IF NOT EXISTS audit_policies (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    retention_days INTEGER NOT NULL DEFAULT 90,
    alert_on_event_types TEXT[] DEFAULT '{}',
    require_approval_for TEXT[] DEFAULT '{}',
    auto_archive_after_days INTEGER NOT NULL DEFAULT 30,
    compliance_frameworks TEXT[] DEFAULT '{"SOC2", "GDPR"}',
    encryption_required BOOLEAN NOT NULL DEFAULT TRUE,
    tamper_detection_enabled BOOLEAN NOT NULL DEFAULT TRUE,
    exportable_formats TEXT[] DEFAULT '{"json", "csv", "pdf"}',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_audit_policies_name ON audit_policies(name);

-- Audit Log Archives Table: Stores metadata for archived audit logs
CREATE TABLE IF NOT EXISTS audit_log_archives (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    archive_name VARCHAR(255) NOT NULL UNIQUE,
    file_path VARCHAR(500) NOT NULL,
    start_date TIMESTAMP NOT NULL,
    end_date TIMESTAMP NOT NULL,
    event_count INTEGER NOT NULL,
    checksum_sha256 VARCHAR(64) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP
);

CREATE INDEX idx_audit_log_archives_created_at ON audit_log_archives(created_at DESC);
CREATE INDEX idx_audit_log_archives_expires_at ON audit_log_archives(expires_at);

-- Audit Event Alert Rules Table: Defines which events trigger alerts
CREATE TABLE IF NOT EXISTS audit_alert_rules (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    event_types TEXT[] NOT NULL,
    severity VARCHAR(20),
    enabled BOOLEAN NOT NULL DEFAULT TRUE,
    alert_channels TEXT[] DEFAULT '{"email", "webhook"}',
    webhook_url VARCHAR(500),
    email_recipients TEXT[] DEFAULT '{}',
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_audit_alert_rules_enabled ON audit_alert_rules(enabled);

-- Audit Event Acknowledgments Table: Track when admins acknowledge events
CREATE TABLE IF NOT EXISTS audit_event_acknowledgments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    event_id UUID NOT NULL REFERENCES audit_events(id) ON DELETE CASCADE,
    acknowledged_by UUID NOT NULL REFERENCES admins(id) ON DELETE CASCADE,
    acknowledged_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    notes TEXT,
    resolved BOOLEAN NOT NULL DEFAULT FALSE,
    resolved_at TIMESTAMP,
    resolution_notes TEXT
);

CREATE INDEX idx_audit_event_acknowledgments_event_id ON audit_event_acknowledgments(event_id);
CREATE INDEX idx_audit_event_acknowledgments_acknowledged_by ON audit_event_acknowledgments(acknowledged_by);
CREATE INDEX idx_audit_event_acknowledgments_resolved ON audit_event_acknowledgments(resolved);

-- Create trigger to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_compliance_events_updated_at BEFORE UPDATE ON compliance_events
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER trg_audit_reports_updated_at BEFORE UPDATE ON audit_reports
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER trg_audit_policies_updated_at BEFORE UPDATE ON audit_policies
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER trg_audit_alert_rules_updated_at BEFORE UPDATE ON audit_alert_rules
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Create function to clean up old audit events (called periodically)
CREATE OR REPLACE FUNCTION cleanup_old_audit_events(retention_days INTEGER DEFAULT 90)
RETURNS INTEGER AS $$
DECLARE
    deleted_count INTEGER;
BEGIN
    DELETE FROM audit_events WHERE created_at < CURRENT_TIMESTAMP - INTERVAL '1 day' * retention_days;
    GET DIAGNOSTICS deleted_count = ROW_COUNT;
    RETURN deleted_count;
END;
$$ LANGUAGE plpgsql;

-- Create function to archive audit events
CREATE OR REPLACE FUNCTION archive_audit_events(start_date TIMESTAMP, end_date TIMESTAMP)
RETURNS TABLE(archived_count INTEGER, archive_id UUID) AS $$
DECLARE
    v_archive_id UUID;
    v_event_count INTEGER;
BEGIN
    SELECT COUNT(*) INTO v_event_count FROM audit_events 
    WHERE created_at >= start_date AND created_at <= end_date;
    
    v_archive_id := gen_random_uuid();
    
    INSERT INTO audit_log_archives (id, archive_name, file_path, start_date, end_date, event_count, checksum_sha256)
    VALUES (
        v_archive_id,
        'audit-logs-' || DATE(start_date) || '-to-' || DATE(end_date),
        's3://audit-archives/audit-logs-' || v_archive_id::TEXT || '.tar.gz',
        start_date,
        end_date,
        v_event_count,
        md5(CONCAT_WS(',', start_date, end_date))
    );
    
    RETURN QUERY SELECT v_event_count, v_archive_id;
END;
$$ LANGUAGE plpgsql;

-- Insert default audit policy
INSERT INTO audit_policies (
    name,
    description,
    retention_days,
    alert_on_event_types,
    require_approval_for,
    auto_archive_after_days,
    compliance_frameworks,
    encryption_required,
    tamper_detection_enabled,
    exportable_formats
) VALUES (
    'Default Policy',
    'Default audit policy for compliance and event tracking',
    90,
    ARRAY['auth_failure', 'brute_force_attempt', 'suspicious_activity', 'ip_blocked'],
    ARRAY['admin_created', 'admin_deleted', 'admin_permission_changed'],
    30,
    ARRAY['SOC2', 'GDPR', 'ISO27001'],
    TRUE,
    TRUE,
    ARRAY['json', 'csv', 'pdf']
) ON CONFLICT DO NOTHING;

-- Add comment to explain audit event table
COMMENT ON TABLE audit_events IS 'Comprehensive audit events table for logging all security-related actions and events';
COMMENT ON TABLE compliance_events IS 'Compliance tracking table for framework compliance verification and documentation';
COMMENT ON TABLE audit_reports IS 'Generated audit reports for compliance and incident analysis';
COMMENT ON TABLE audit_policies IS 'System-wide audit and compliance policies';
COMMENT ON TABLE audit_log_archives IS 'Archived audit logs metadata for long-term retention and archival';
