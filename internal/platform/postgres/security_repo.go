package postgres

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"log/slog"

	"github.com/vortexui/vortexui/internal/domain"
)

// SecurityThreatRepositoryImpl implements port.SecurityThreatRepository
type SecurityThreatRepositoryImpl struct {
	db *pgxpool.Pool
	log *slog.Logger
}

// NewSecurityThreatRepository creates new threat repository
func NewSecurityThreatRepository(db *pgxpool.Pool, log *slog.Logger) *SecurityThreatRepositoryImpl {
	return &SecurityThreatRepositoryImpl{db: db, log: log}
}

// SaveThreat records detected threat
func (r *SecurityThreatRepositoryImpl) SaveThreat(ctx context.Context, threat *domain.SecurityThreat) error {
	metadataJSON, _ := json.Marshal(threat.Metadata)
	
	query := `
		INSERT INTO security_threats (id, threat_type, severity, source_ip, target_path, payload, detection_method, blocked, metadata, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`

	_, err := r.db.Exec(ctx, query,
		threat.ID, threat.ThreatType, threat.Severity, threat.SourceIP, threat.TargetPath,
		threat.Payload, threat.DetectionMethod, threat.Blocked, string(metadataJSON), threat.CreatedAt,
	)
	if err != nil {
		r.log.Error("failed to save threat", "error", err, "threat_type", threat.ThreatType)
		return fmt.Errorf("save threat: %w", err)
	}

	return nil
}

// GetThreat retrieves threat by ID
func (r *SecurityThreatRepositoryImpl) GetThreat(ctx context.Context, threatID uuid.UUID) (*domain.SecurityThreat, error) {
	threat := &domain.SecurityThreat{}
	metadataStr := ""

	query := `SELECT id, threat_type, severity, source_ip, target_path, payload, detection_method, blocked, metadata, created_at FROM security_threats WHERE id = $1`

	err := r.db.QueryRow(ctx, query, threatID).Scan(
		&threat.ID, &threat.ThreatType, &threat.Severity, &threat.SourceIP, &threat.TargetPath,
		&threat.Payload, &threat.DetectionMethod, &threat.Blocked, &metadataStr, &threat.CreatedAt,
	)

	if err == pgx.ErrNoRows {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		r.log.Error("failed to get threat", "error", err, "threat_id", threatID)
		return nil, fmt.Errorf("get threat: %w", err)
	}

	if len(metadataStr) > 0 {
		_ = json.Unmarshal([]byte(metadataStr), &threat.Metadata)
	}

	return threat, nil
}

// ListThreats retrieves recent threats with filtering
func (r *SecurityThreatRepositoryImpl) ListThreats(ctx context.Context, threatType string, limit, offset int) ([]*domain.SecurityThreat, error) {
	query := `
		SELECT id, threat_type, severity, source_ip, target_path, payload, detection_method, blocked, metadata, created_at
		FROM security_threats
	`

	args := []interface{}{}
	argCount := 1

	if threatType != "" {
		query += fmt.Sprintf(" WHERE threat_type = $%d", argCount)
		args = append(args, threatType)
		argCount++
	}

	query += fmt.Sprintf(" ORDER BY created_at DESC LIMIT $%d OFFSET $%d", argCount, argCount+1)
	args = append(args, limit, offset)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		r.log.Error("failed to query threats", "error", err)
		return nil, fmt.Errorf("query threats: %w", err)
	}
	defer rows.Close()

	threats := make([]*domain.SecurityThreat, 0)
	for rows.Next() {
		threat := &domain.SecurityThreat{}
		metadataStr := ""

		if err := rows.Scan(
			&threat.ID, &threat.ThreatType, &threat.Severity, &threat.SourceIP, &threat.TargetPath,
			&threat.Payload, &threat.DetectionMethod, &threat.Blocked, &metadataStr, &threat.CreatedAt,
		); err != nil {
			r.log.Error("failed to scan threat", "error", err)
			return nil, fmt.Errorf("scan threat: %w", err)
		}

		if len(metadataStr) > 0 {
			_ = json.Unmarshal([]byte(metadataStr), &threat.Metadata)
		}

		threats = append(threats, threat)
	}

	return threats, nil
}

// GetBlockedThreats retrieves blocked threats
func (r *SecurityThreatRepositoryImpl) GetBlockedThreats(ctx context.Context, limit, offset int) ([]*domain.SecurityThreat, error) {
	query := `
		SELECT id, threat_type, severity, source_ip, target_path, payload, detection_method, blocked, metadata, created_at
		FROM security_threats
		WHERE blocked = TRUE
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := r.db.Query(ctx, query, limit, offset)
	if err != nil {
		r.log.Error("failed to query blocked threats", "error", err)
		return nil, fmt.Errorf("query blocked threats: %w", err)
	}
	defer rows.Close()

	threats := make([]*domain.SecurityThreat, 0)
	for rows.Next() {
		threat := &domain.SecurityThreat{}
		metadataStr := ""

		if err := rows.Scan(
			&threat.ID, &threat.ThreatType, &threat.Severity, &threat.SourceIP, &threat.TargetPath,
			&threat.Payload, &threat.DetectionMethod, &threat.Blocked, &metadataStr, &threat.CreatedAt,
		); err != nil {
			r.log.Error("failed to scan threat", "error", err)
			return nil, fmt.Errorf("scan threat: %w", err)
		}

		if len(metadataStr) > 0 {
			_ = json.Unmarshal([]byte(metadataStr), &threat.Metadata)
		}

		threats = append(threats, threat)
	}

	return threats, nil
}

// CountThreats returns total threat count
func (r *SecurityThreatRepositoryImpl) CountThreats(ctx context.Context) (int64, error) {
	var count int64

	query := `SELECT COUNT(*) FROM security_threats`

	err := r.db.QueryRow(ctx, query).Scan(&count)
	if err != nil {
		r.log.Error("failed to count threats", "error", err)
		return 0, fmt.Errorf("count threats: %w", err)
	}

	return count, nil
}

// DeleteOldThreats removes threats older than days
func (r *SecurityThreatRepositoryImpl) DeleteOldThreats(ctx context.Context, daysOld int) error {
	query := `DELETE FROM security_threats WHERE created_at < NOW() - INTERVAL '1 day' * $1`

	result, err := r.db.Exec(ctx, query, daysOld)
	if err != nil {
		r.log.Error("failed to delete old threats", "error", err, "days_old", daysOld)
		return fmt.Errorf("delete old threats: %w", err)
	}

	r.log.Info("deleted old threats", "rows_affected", result.RowsAffected())
	return nil
}

// ========== IP Reputation Repository ==========

// IPReputationRepositoryImpl implements port.IPReputationRepository
type IPReputationRepositoryImpl struct {
	db *pgxpool.Pool
	log *slog.Logger
}

// NewIPReputationRepository creates new IP reputation repository
func NewIPReputationRepository(db *pgxpool.Pool, log *slog.Logger) *IPReputationRepositoryImpl {
	return &IPReputationRepositoryImpl{db: db, log: log}
}

// SaveReputation stores or updates IP reputation
func (r *IPReputationRepositoryImpl) SaveReputation(ctx context.Context, rep *domain.IPReputation) error {
	query := `
		INSERT INTO ip_reputation (id, ip_address, reputation_score, threat_level, failed_logins, blocked_requests, country, is_proxy, is_tor, is_vpn, last_seen, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
		ON CONFLICT (ip_address) DO UPDATE SET
			reputation_score = $3, threat_level = $4, failed_logins = $5, blocked_requests = $6,
			country = $7, is_proxy = $8, is_tor = $9, is_vpn = $10, last_seen = $11, updated_at = $13
	`

	_, err := r.db.Exec(ctx, query,
		rep.ID, rep.IPAddress, rep.ReputationScore, rep.ThreatLevel, rep.FailedLogins,
		rep.BlockedRequests, rep.Country, rep.IsProxy, rep.IsTor, rep.IsVPN, rep.LastSeen, rep.CreatedAt, rep.UpdatedAt,
	)

	if err != nil {
		r.log.Error("failed to save reputation", "error", err, "ip", rep.IPAddress)
		return fmt.Errorf("save reputation: %w", err)
	}

	return nil
}

// GetReputation retrieves reputation for IP
func (r *IPReputationRepositoryImpl) GetReputation(ctx context.Context, ipAddress string) (*domain.IPReputation, error) {
	rep := &domain.IPReputation{}

	query := `
		SELECT id, ip_address, reputation_score, threat_level, failed_logins, blocked_requests,
		       country, is_proxy, is_tor, is_vpn, last_seen, created_at, updated_at
		FROM ip_reputation WHERE ip_address = $1
	`

	err := r.db.QueryRow(ctx, query, ipAddress).Scan(
		&rep.ID, &rep.IPAddress, &rep.ReputationScore, &rep.ThreatLevel, &rep.FailedLogins,
		&rep.BlockedRequests, &rep.Country, &rep.IsProxy, &rep.IsTor, &rep.IsVPN, &rep.LastSeen, &rep.CreatedAt, &rep.UpdatedAt,
	)

	if err == pgx.ErrNoRows {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		r.log.Error("failed to get reputation", "error", err, "ip", ipAddress)
		return nil, fmt.Errorf("get reputation: %w", err)
	}

	return rep, nil
}

// ListMaliciousIPs retrieves blacklisted IPs
func (r *IPReputationRepositoryImpl) ListMaliciousIPs(ctx context.Context, limit, offset int) ([]*domain.IPReputation, error) {
	query := `
		SELECT id, ip_address, reputation_score, threat_level, failed_logins, blocked_requests,
		       country, is_proxy, is_tor, is_vpn, last_seen, created_at, updated_at
		FROM ip_reputation
		WHERE threat_level IN ('suspicious', 'malicious')
		ORDER BY reputation_score ASC
		LIMIT $1 OFFSET $2
	`

	rows, err := r.db.Query(ctx, query, limit, offset)
	if err != nil {
		r.log.Error("failed to list malicious IPs", "error", err)
		return nil, fmt.Errorf("list malicious IPs: %w", err)
	}
	defer rows.Close()

	reps := make([]*domain.IPReputation, 0)
	for rows.Next() {
		rep := &domain.IPReputation{}

		if err := rows.Scan(
			&rep.ID, &rep.IPAddress, &rep.ReputationScore, &rep.ThreatLevel, &rep.FailedLogins,
			&rep.BlockedRequests, &rep.Country, &rep.IsProxy, &rep.IsTor, &rep.IsVPN, &rep.LastSeen, &rep.CreatedAt, &rep.UpdatedAt,
		); err != nil {
			r.log.Error("failed to scan reputation", "error", err)
			return nil, fmt.Errorf("scan reputation: %w", err)
		}

		reps = append(reps, rep)
	}

	return reps, nil
}

// UpdateReputationScore adjusts IP score
func (r *IPReputationRepositoryImpl) UpdateReputationScore(ctx context.Context, ipAddress string, scoreChange float64) error {
	query := `
		UPDATE ip_reputation
		SET reputation_score = LEAST(100.0, GREATEST(0.0, reputation_score + $2)),
		    updated_at = NOW()
		WHERE ip_address = $1
	`

	result, err := r.db.Exec(ctx, query, ipAddress, scoreChange)
	if err != nil {
		r.log.Error("failed to update reputation score", "error", err, "ip", ipAddress)
		return fmt.Errorf("update reputation: %w", err)
	}

	if result.RowsAffected() == 0 {
		return domain.ErrNotFound
	}

	return nil
}

// BlockIP marks IP as blocked
func (r *IPReputationRepositoryImpl) BlockIP(ctx context.Context, ipAddress string) error {
	query := `
		UPDATE ip_reputation
		SET threat_level = 'malicious', reputation_score = 0, updated_at = NOW()
		WHERE ip_address = $1
	`

	result, err := r.db.Exec(ctx, query, ipAddress)
	if err != nil {
		r.log.Error("failed to block IP", "error", err, "ip", ipAddress)
		return fmt.Errorf("block IP: %w", err)
	}

	if result.RowsAffected() == 0 {
		return domain.ErrNotFound
	}

	return nil
}

// UnblockIP removes IP block
func (r *IPReputationRepositoryImpl) UnblockIP(ctx context.Context, ipAddress string) error {
	query := `
		UPDATE ip_reputation
		SET threat_level = 'neutral', reputation_score = 50.0, updated_at = NOW()
		WHERE ip_address = $1
	`

	result, err := r.db.Exec(ctx, query, ipAddress)
	if err != nil {
		r.log.Error("failed to unblock IP", "error", err, "ip", ipAddress)
		return fmt.Errorf("unblock IP: %w", err)
	}

	if result.RowsAffected() == 0 {
		return domain.ErrNotFound
	}

	return nil
}

// ========== Incident Response Repository ==========

// IncidentResponseRepositoryImpl implements port.IncidentResponseRepository
type IncidentResponseRepositoryImpl struct {
	db *pgxpool.Pool
	log *slog.Logger
}

// NewIncidentResponseRepository creates new incident repository
func NewIncidentResponseRepository(db *pgxpool.Pool, log *slog.Logger) *IncidentResponseRepositoryImpl {
	return &IncidentResponseRepositoryImpl{db: db, log: log}
}

// SaveIncident creates new incident record
func (r *IncidentResponseRepositoryImpl) SaveIncident(ctx context.Context, incident *domain.IncidentResponse) error {
	metadataJSON, _ := json.Marshal(incident.Metadata)

	query := `
		INSERT INTO incident_responses (id, incident_type, title, description, severity, status, threat_actors, impacted_systems, root_cause, remediation, metadata, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	`

	_, err := r.db.Exec(ctx, query,
		incident.ID, incident.IncidentType, incident.Title, incident.Description, incident.Severity,
		incident.Status, incident.ThreatActors, incident.ImpactedSystems, incident.RootCause,
		incident.Remediation, string(metadataJSON), incident.CreatedAt,
	)

	if err != nil {
		r.log.Error("failed to save incident", "error", err, "incident_type", incident.IncidentType)
		return fmt.Errorf("save incident: %w", err)
	}

	return nil
}

// GetIncident retrieves incident details
func (r *IncidentResponseRepositoryImpl) GetIncident(ctx context.Context, incidentID uuid.UUID) (*domain.IncidentResponse, error) {
	incident := &domain.IncidentResponse{}
	metadataStr := ""

	query := `
		SELECT id, incident_type, title, description, severity, status, threat_actors, impacted_systems,
		       root_cause, remediation, metadata, created_at, resolved_at
		FROM incident_responses WHERE id = $1
	`

	err := r.db.QueryRow(ctx, query, incidentID).Scan(
		&incident.ID, &incident.IncidentType, &incident.Title, &incident.Description, &incident.Severity,
		&incident.Status, &incident.ThreatActors, &incident.ImpactedSystems, &incident.RootCause,
		&incident.Remediation, &metadataStr, &incident.CreatedAt, &incident.ResolvedAt,
	)

	if err == pgx.ErrNoRows {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		r.log.Error("failed to get incident", "error", err, "incident_id", incidentID)
		return nil, fmt.Errorf("get incident: %w", err)
	}

	if len(metadataStr) > 0 {
		_ = json.Unmarshal([]byte(metadataStr), &incident.Metadata)
	}

	return incident, nil
}

// ListIncidents retrieves active incidents
func (r *IncidentResponseRepositoryImpl) ListIncidents(ctx context.Context, status string, limit, offset int) ([]*domain.IncidentResponse, error) {
	query := `
		SELECT id, incident_type, title, description, severity, status, threat_actors, impacted_systems,
		       root_cause, remediation, metadata, created_at, resolved_at
		FROM incident_responses
	`

	args := []interface{}{}
	argCount := 1

	if status != "" {
		query += fmt.Sprintf(" WHERE status = $%d", argCount)
		args = append(args, status)
		argCount++
	}

	query += fmt.Sprintf(" ORDER BY created_at DESC LIMIT $%d OFFSET $%d", argCount, argCount+1)
	args = append(args, limit, offset)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		r.log.Error("failed to list incidents", "error", err)
		return nil, fmt.Errorf("list incidents: %w", err)
	}
	defer rows.Close()

	incidents := make([]*domain.IncidentResponse, 0)
	for rows.Next() {
		incident := &domain.IncidentResponse{}
		metadataStr := ""

		if err := rows.Scan(
			&incident.ID, &incident.IncidentType, &incident.Title, &incident.Description, &incident.Severity,
			&incident.Status, &incident.ThreatActors, &incident.ImpactedSystems, &incident.RootCause,
			&incident.Remediation, &metadataStr, &incident.CreatedAt, &incident.ResolvedAt,
		); err != nil {
			r.log.Error("failed to scan incident", "error", err)
			return nil, fmt.Errorf("scan incident: %w", err)
		}

		if len(metadataStr) > 0 {
			_ = json.Unmarshal([]byte(metadataStr), &incident.Metadata)
		}

		incidents = append(incidents, incident)
	}

	return incidents, nil
}

// UpdateIncidentStatus changes incident status
func (r *IncidentResponseRepositoryImpl) UpdateIncidentStatus(ctx context.Context, incidentID uuid.UUID, status string) error {
	query := `UPDATE incident_responses SET status = $2 WHERE id = $1`

	result, err := r.db.Exec(ctx, query, incidentID, status)
	if err != nil {
		r.log.Error("failed to update incident status", "error", err, "incident_id", incidentID)
		return fmt.Errorf("update incident status: %w", err)
	}

	if result.RowsAffected() == 0 {
		return domain.ErrNotFound
	}

	return nil
}

// ResolveIncident marks incident as resolved
func (r *IncidentResponseRepositoryImpl) ResolveIncident(ctx context.Context, incidentID uuid.UUID, resolution string) error {
	query := `UPDATE incident_responses SET status = 'resolved', resolved_at = NOW() WHERE id = $1`

	result, err := r.db.Exec(ctx, query, incidentID)
	if err != nil {
		r.log.Error("failed to resolve incident", "error", err, "incident_id", incidentID)
		return fmt.Errorf("resolve incident: %w", err)
	}

	if result.RowsAffected() == 0 {
		return domain.ErrNotFound
	}

	return nil
}

// GetIncidentsByType retrieves incidents of specific type
func (r *IncidentResponseRepositoryImpl) GetIncidentsByType(ctx context.Context, incidentType string) ([]*domain.IncidentResponse, error) {
	query := `
		SELECT id, incident_type, title, description, severity, status, threat_actors, impacted_systems,
		       root_cause, remediation, metadata, created_at, resolved_at
		FROM incident_responses
		WHERE incident_type = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(ctx, query, incidentType)
	if err != nil {
		r.log.Error("failed to get incidents by type", "error", err, "type", incidentType)
		return nil, fmt.Errorf("get incidents by type: %w", err)
	}
	defer rows.Close()

	incidents := make([]*domain.IncidentResponse, 0)
	for rows.Next() {
		incident := &domain.IncidentResponse{}
		metadataStr := ""

		if err := rows.Scan(
			&incident.ID, &incident.IncidentType, &incident.Title, &incident.Description, &incident.Severity,
			&incident.Status, &incident.ThreatActors, &incident.ImpactedSystems, &incident.RootCause,
			&incident.Remediation, &metadataStr, &incident.CreatedAt, &incident.ResolvedAt,
		); err != nil {
			r.log.Error("failed to scan incident", "error", err)
			return nil, fmt.Errorf("scan incident: %w", err)
		}

		if len(metadataStr) > 0 {
			_ = json.Unmarshal([]byte(metadataStr), &incident.Metadata)
		}

		incidents = append(incidents, incident)
	}

	return incidents, nil
}

// ========== Security Policy Repository ==========

// SecurityPolicyRepositoryImpl implements port.SecurityPolicyRepository
type SecurityPolicyRepositoryImpl struct {
	db *pgxpool.Pool
	log *slog.Logger
}

// NewSecurityPolicyRepository creates new policy repository
func NewSecurityPolicyRepository(db *pgxpool.Pool, log *slog.Logger) *SecurityPolicyRepositoryImpl {
	return &SecurityPolicyRepositoryImpl{db: db, log: log}
}

// GetPolicy retrieves current security policy
func (r *SecurityPolicyRepositoryImpl) GetPolicy(ctx context.Context) (*domain.SecurityPolicy, error) {
	policy := &domain.SecurityPolicy{}

	query := `
		SELECT id, name, description, enable_csrf_protection, enable_xss_protection,
		       enable_sql_injection_detection, enable_ddos_protection, require_https,
		       encryption_required, min_password_length, require_mfa, session_timeout,
		       max_concurrent_sessions, allowed_cors_origins, require_content_security, created_at, updated_at
		FROM security_policies LIMIT 1
	`

	err := r.db.QueryRow(ctx, query).Scan(
		&policy.ID, &policy.Name, &policy.Description, &policy.EnableCSRFProtection, &policy.EnableXSSProtection,
		&policy.EnableSQLInjectionDetection, &policy.EnableDDoSProtection, &policy.RequireHTTPS,
		&policy.EncryptionRequired, &policy.MinPasswordLength, &policy.RequireMFA, &policy.SessionTimeout,
		&policy.MaxConcurrentSessions, &policy.AllowedCORSOrigins, &policy.RequireContentSecurity, &policy.CreatedAt, &policy.UpdatedAt,
	)

	if err == pgx.ErrNoRows {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		r.log.Error("failed to get policy", "error", err)
		return nil, fmt.Errorf("get policy: %w", err)
	}

	return policy, nil
}

// SavePolicy creates or updates policy
func (r *SecurityPolicyRepositoryImpl) SavePolicy(ctx context.Context, policy *domain.SecurityPolicy) error {
	query := `
		INSERT INTO security_policies (id, name, description, enable_csrf_protection, enable_xss_protection,
		       enable_sql_injection_detection, enable_ddos_protection, require_https,
		       encryption_required, min_password_length, require_mfa, session_timeout,
		       max_concurrent_sessions, allowed_cors_origins, require_content_security, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17)
		ON CONFLICT (name) DO UPDATE SET
			description = $3, enable_csrf_protection = $4, enable_xss_protection = $5,
			enable_sql_injection_detection = $6, enable_ddos_protection = $7, require_https = $8,
			encryption_required = $9, min_password_length = $10, require_mfa = $11, session_timeout = $12,
			max_concurrent_sessions = $13, allowed_cors_origins = $14, require_content_security = $15, updated_at = $17
	`

	_, err := r.db.Exec(ctx, query,
		policy.ID, policy.Name, policy.Description, policy.EnableCSRFProtection, policy.EnableXSSProtection,
		policy.EnableSQLInjectionDetection, policy.EnableDDoSProtection, policy.RequireHTTPS,
		policy.EncryptionRequired, policy.MinPasswordLength, policy.RequireMFA, policy.SessionTimeout,
		policy.MaxConcurrentSessions, policy.AllowedCORSOrigins, policy.RequireContentSecurity, policy.CreatedAt, policy.UpdatedAt,
	)

	if err != nil {
		r.log.Error("failed to save policy", "error", err)
		return fmt.Errorf("save policy: %w", err)
	}

	return nil
}

// GetDefaultPolicy returns system defaults
func (r *SecurityPolicyRepositoryImpl) GetDefaultPolicy(ctx context.Context) (*domain.SecurityPolicy, error) {
	policy := &domain.SecurityPolicy{}

	query := `
		SELECT id, name, description, enable_csrf_protection, enable_xss_protection,
		       enable_sql_injection_detection, enable_ddos_protection, require_https,
		       encryption_required, min_password_length, require_mfa, session_timeout,
		       max_concurrent_sessions, allowed_cors_origins, require_content_security, created_at, updated_at
		FROM security_policies WHERE name = 'Default Policy'
	`

	err := r.db.QueryRow(ctx, query).Scan(
		&policy.ID, &policy.Name, &policy.Description, &policy.EnableCSRFProtection, &policy.EnableXSSProtection,
		&policy.EnableSQLInjectionDetection, &policy.EnableDDoSProtection, &policy.RequireHTTPS,
		&policy.EncryptionRequired, &policy.MinPasswordLength, &policy.RequireMFA, &policy.SessionTimeout,
		&policy.MaxConcurrentSessions, &policy.AllowedCORSOrigins, &policy.RequireContentSecurity, &policy.CreatedAt, &policy.UpdatedAt,
	)

	if err == pgx.ErrNoRows {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		r.log.Error("failed to get default policy", "error", err)
		return nil, fmt.Errorf("get default policy: %w", err)
	}

	return policy, nil
}

// ValidateCompliance checks if system meets policy requirements
func (r *SecurityPolicyRepositoryImpl) ValidateCompliance(ctx context.Context) (map[string]bool, []string, error) {
	compliant := make(map[string]bool)
	issues := make([]string, 0)

	policy, err := r.GetPolicy(ctx)
	if err != nil {
		r.log.Error("failed to validate compliance", "error", err)
		return compliant, issues, fmt.Errorf("validate compliance: %w", err)
	}

	// Check various compliance metrics
	compliant["csrf_protection"] = policy.EnableCSRFProtection
	compliant["xss_protection"] = policy.EnableXSSProtection
	compliant["sql_injection_protection"] = policy.EnableSQLInjectionDetection
	compliant["ddos_protection"] = policy.EnableDDoSProtection
	compliant["https_required"] = policy.RequireHTTPS
	compliant["encryption_required"] = policy.EncryptionRequired
	compliant["mfa_required"] = policy.RequireMFA

	// Add issues where compliance is not met
	if !policy.EnableCSRFProtection {
		issues = append(issues, "CSRF protection is disabled")
	}
	if !policy.RequireHTTPS {
		issues = append(issues, "HTTPS is not required")
	}
	if !policy.EncryptionRequired {
		issues = append(issues, "Encryption is not required")
	}

	return compliant, issues, nil
}
