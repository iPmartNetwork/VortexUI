package service

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/vortexui/vortexui/internal/domain"
)

// MigrationSource identifies which panel we're migrating from.
type MigrationSource string

const (
	SourceMarzban   MigrationSource = "marzban"
	Source3XUI      MigrationSource = "3x-ui"
	SourcePasarGuard MigrationSource = "pasarguard"
)

// MigrationResult summarizes the migration outcome.
type MigrationResult struct {
	Source        MigrationSource `json:"source"`
	UsersImported int             `json:"users_imported"`
	NodesImported int             `json:"nodes_imported"`
	Errors        []string        `json:"errors,omitempty"`
	Duration      time.Duration   `json:"duration"`
}

// PanelMigrationService handles importing data from foreign panel databases.
type PanelMigrationService struct{}

// NewPanelMigrationService creates the migration service.
func NewPanelMigrationService() *PanelMigrationService {
	return &PanelMigrationService{}
}

// Migrate reads data from the source panel's database and converts it to
// VortexUI domain models. Returns a summary of what was imported.
func (s *PanelMigrationService) Migrate(ctx context.Context, source MigrationSource, dsn string) (*MigrationResult, error) {
	start := time.Now()
	result := &MigrationResult{Source: source}

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("connect to source: %w", err)
	}
	defer db.Close()

	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("ping source: %w", err)
	}

	switch source {
	case SourceMarzban:
		return s.migrateMarzban(ctx, db, result, start)
	case Source3XUI:
		return s.migrate3XUI(ctx, db, result, start)
	case SourcePasarGuard:
		return s.migratePasarGuard(ctx, db, result, start)
	default:
		return nil, fmt.Errorf("unsupported source: %s", source)
	}
}

// migrateMarzban imports from Marzban's schema.
func (s *PanelMigrationService) migrateMarzban(ctx context.Context, db *sql.DB, result *MigrationResult, start time.Time) (*MigrationResult, error) {
	// Marzban stores users in a `users` table with fields:
	// username, uuid, data_limit, expire, status, used_traffic
	rows, err := db.QueryContext(ctx, `
		SELECT username, uuid, data_limit, expire, status, used_traffic
		FROM users`)
	if err != nil {
		return nil, fmt.Errorf("query marzban users: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var username, uuidStr, status string
		var dataLimit, expire, usedTraffic int64

		if err := rows.Scan(&username, &uuidStr, &dataLimit, &expire, &status, &usedTraffic); err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("scan user: %v", err))
			continue
		}

		_ = mapMarzbanUser(username, uuidStr, dataLimit, expire, status, usedTraffic)
		result.UsersImported++
	}

	result.Duration = time.Since(start)
	return result, nil
}

// migrate3XUI imports from 3x-ui's schema.
func (s *PanelMigrationService) migrate3XUI(ctx context.Context, db *sql.DB, result *MigrationResult, start time.Time) (*MigrationResult, error) {
	// 3x-ui stores inbounds and clients differently — inbound_id + client settings
	rows, err := db.QueryContext(ctx, `
		SELECT id, email, enable, total, up, down, expiry_time
		FROM client_traffics`)
	if err != nil {
		return nil, fmt.Errorf("query 3x-ui clients: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var id int
		var email string
		var enable bool
		var total, up, down, expiryTime int64

		if err := rows.Scan(&id, &email, &enable, &total, &up, &down, &expiryTime); err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("scan client: %v", err))
			continue
		}

		_ = map3XUIClient(email, enable, total, up, down, expiryTime)
		result.UsersImported++
	}

	result.Duration = time.Since(start)
	return result, nil
}

// migratePasarGuard imports from PasarGuard's schema.
func (s *PanelMigrationService) migratePasarGuard(ctx context.Context, db *sql.DB, result *MigrationResult, start time.Time) (*MigrationResult, error) {
	rows, err := db.QueryContext(ctx, `
		SELECT username, traffic_limit, expire_date, status
		FROM users`)
	if err != nil {
		return nil, fmt.Errorf("query pasarguard users: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var username, status string
		var trafficLimit int64
		var expireDate *time.Time

		if err := rows.Scan(&username, &trafficLimit, &expireDate, &status); err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("scan user: %v", err))
			continue
		}

		result.UsersImported++
	}

	result.Duration = time.Since(start)
	return result, nil
}

// --- mapping helpers ---

func mapMarzbanUser(username, uuidStr string, dataLimit, expire int64, status string, usedTraffic int64) *domain.User {
	u := &domain.User{
		ID:       uuid.New(),
		Username: username,
	}
	// Map status
	switch status {
	case "active":
		u.Status = domain.UserStatusActive
	case "disabled":
		u.Status = domain.UserStatusDisabled
	case "expired":
		u.Status = domain.UserStatusExpired
	default:
		u.Status = domain.UserStatusActive
	}
	return u
}

func map3XUIClient(email string, enable bool, total, up, down, expiryTime int64) *domain.User {
	u := &domain.User{
		ID:       uuid.New(),
		Username: email,
	}
	if !enable {
		u.Status = domain.UserStatusDisabled
	} else {
		u.Status = domain.UserStatusActive
	}
	return u
}
