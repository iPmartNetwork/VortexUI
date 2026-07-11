package postgres

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/vortexui/vortexui/internal/domain"
)

// IPAccessRuleRepository implements IP access rule repository using pgx
type IPAccessRuleRepository struct {
	pool *pgxpool.Pool
	log  *slog.Logger
}

// NewIPAccessRuleRepository creates a new IP access rule repository
func NewIPAccessRuleRepository(pool *pgxpool.Pool, log *slog.Logger) *IPAccessRuleRepository {
	if log == nil {
		log = slog.Default()
	}
	return &IPAccessRuleRepository{
		pool: pool,
		log:  log,
	}
}

// SaveRule saves a new IP access rule
func (r *IPAccessRuleRepository) SaveRule(ctx context.Context, rule *domain.IPAccessRule) error {
	query := `
		INSERT INTO ip_access_rules (id, admin_id, rule_type, ip_address, description, active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	_, err := r.pool.Exec(ctx, query,
		rule.ID, rule.AdminID, rule.RuleType, rule.IPAddress, rule.Description, rule.Active,
		rule.CreatedAt, rule.UpdatedAt,
	)

	if err != nil {
		r.log.Error("failed to save IP access rule", "admin_id", rule.AdminID, "error", err)
		return fmt.Errorf("save IP access rule: %w", err)
	}

	return nil
}

// GetRule retrieves a rule by ID
func (r *IPAccessRuleRepository) GetRule(ctx context.Context, ruleID uuid.UUID) (*domain.IPAccessRule, error) {
	rule := &domain.IPAccessRule{}

	query := `
		SELECT id, admin_id, rule_type, ip_address, description, active, created_at, updated_at
		FROM ip_access_rules
		WHERE id = $1
	`

	err := r.pool.QueryRow(ctx, query, ruleID).Scan(
		&rule.ID, &rule.AdminID, &rule.RuleType, &rule.IPAddress, &rule.Description,
		&rule.Active, &rule.CreatedAt, &rule.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.NewDomainError(
				domain.ErrCodeNotFound,
				"IP access rule not found",
				err,
			)
		}
		r.log.Error("failed to get IP access rule", "rule_id", ruleID, "error", err)
		return nil, fmt.Errorf("get IP access rule: %w", err)
	}

	return rule, nil
}

// ListRules retrieves rules with optional filtering
func (r *IPAccessRuleRepository) ListRules(ctx context.Context, adminID *uuid.UUID, ruleType *string) ([]domain.IPAccessRule, error) {
	query := `
		SELECT id, admin_id, rule_type, ip_address, description, active, created_at, updated_at
		FROM ip_access_rules
		WHERE 1=1
	`

	args := []interface{}{}
	argIdx := 1

	if adminID != nil {
		query += fmt.Sprintf(" AND admin_id = $%d", argIdx)
		args = append(args, *adminID)
		argIdx++
	}

	if ruleType != nil {
		query += fmt.Sprintf(" AND rule_type = $%d", argIdx)
		args = append(args, *ruleType)
		argIdx++
	}

	query += " ORDER BY created_at DESC"

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		r.log.Error("failed to list IP access rules", "error", err)
		return nil, fmt.Errorf("list IP access rules: %w", err)
	}
	defer rows.Close()

	rules := []domain.IPAccessRule{}
	for rows.Next() {
		rule := domain.IPAccessRule{}
		if err := rows.Scan(
			&rule.ID, &rule.AdminID, &rule.RuleType, &rule.IPAddress, &rule.Description,
			&rule.Active, &rule.CreatedAt, &rule.UpdatedAt,
		); err != nil {
			r.log.Error("failed to scan IP access rule", "error", err)
			continue
		}
		rules = append(rules, rule)
	}

	if err = rows.Err(); err != nil {
		r.log.Error("error reading IP access rule rows", "error", err)
		return nil, err
	}

	return rules, nil
}

// UpdateRule updates an existing rule
func (r *IPAccessRuleRepository) UpdateRule(ctx context.Context, rule *domain.IPAccessRule) error {
	query := `
		UPDATE ip_access_rules
		SET rule_type = $1, ip_address = $2, description = $3, active = $4, updated_at = $5
		WHERE id = $6
	`

	result, err := r.pool.Exec(ctx, query,
		rule.RuleType, rule.IPAddress, rule.Description, rule.Active, rule.UpdatedAt, rule.ID,
	)

	if err != nil {
		r.log.Error("failed to update IP access rule", "rule_id", rule.ID, "error", err)
		return fmt.Errorf("update IP access rule: %w", err)
	}

	if result.RowsAffected() == 0 {
		return domain.NewDomainError(
			domain.ErrCodeNotFound,
			"IP access rule not found",
			errors.New("no rows affected"),
		)
	}

	return nil
}

// DeleteRule deletes a rule
func (r *IPAccessRuleRepository) DeleteRule(ctx context.Context, ruleID uuid.UUID) error {
	query := `DELETE FROM ip_access_rules WHERE id = $1`

	result, err := r.pool.Exec(ctx, query, ruleID)
	if err != nil {
		r.log.Error("failed to delete IP access rule", "rule_id", ruleID, "error", err)
		return fmt.Errorf("delete IP access rule: %w", err)
	}

	if result.RowsAffected() == 0 {
		return domain.NewDomainError(
			domain.ErrCodeNotFound,
			"IP access rule not found",
			errors.New("no rows affected"),
		)
	}

	return nil
}

// GetGlobalRules retrieves all global (admin_id = NULL) rules
func (r *IPAccessRuleRepository) GetGlobalRules(ctx context.Context) ([]domain.IPAccessRule, error) {
	query := `
		SELECT id, admin_id, rule_type, ip_address, description, active, created_at, updated_at
		FROM ip_access_rules
		WHERE admin_id IS NULL AND active = TRUE
		ORDER BY created_at DESC
	`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		r.log.Error("failed to get global IP access rules", "error", err)
		return nil, fmt.Errorf("get global IP access rules: %w", err)
	}
	defer rows.Close()

	rules := []domain.IPAccessRule{}
	for rows.Next() {
		rule := domain.IPAccessRule{}
		if err := rows.Scan(
			&rule.ID, &rule.AdminID, &rule.RuleType, &rule.IPAddress, &rule.Description,
			&rule.Active, &rule.CreatedAt, &rule.UpdatedAt,
		); err != nil {
			r.log.Error("failed to scan global IP access rule", "error", err)
			continue
		}
		rules = append(rules, rule)
	}

	if err = rows.Err(); err != nil {
		r.log.Error("error reading global IP access rule rows", "error", err)
		return nil, err
	}

	return rules, nil
}
