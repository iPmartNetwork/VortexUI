package postgres

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/vortexui/vortexui/internal/domain"
)

// Supplemental tables exported/restored with JSON v3 backups (whitelist only).
var supplementalTableOrder = []string{
	"panel_settings",
	"billing_settings",
	"wallet_packages",
	"wallet_deposits",
	"admin_wallet_ledger",
	"reseller_payment_config",
	"orders",
	"admin_quota_notify_config",
	"admin_quota_notify_events",
	"quota_notify_config",
	"quota_policies",
	"sub_settings",
	"sub_hosts",
	"wireguard_peers",
}

// OptionalTrafficTables may be included when IncludeTrafficMetrics is set.
var optionalTrafficTables = []string{
	"traffic_points",
}

func exportSupplementalTables(ctx context.Context, pool *pgxpool.Pool, includeTraffic bool) (map[string][]map[string]any, error) {
	tables := append([]string{}, supplementalTableOrder...)
	if includeTraffic {
		tables = append(tables, optionalTrafficTables...)
	}
	out := make(map[string][]map[string]any, len(tables))
	for _, table := range tables {
		rows, err := exportTableRows(ctx, pool, table)
		if err != nil {
			return nil, fmt.Errorf("export %s: %w", table, err)
		}
		if len(rows) > 0 {
			out[table] = rows
		}
	}
	return out, nil
}

func exportTableRows(ctx context.Context, pool *pgxpool.Pool, table string) ([]map[string]any, error) {
	if !isAllowedBackupTable(table) {
		return nil, fmt.Errorf("table not allowed: %s", table)
	}
	sql := fmt.Sprintf("SELECT * FROM %s", quoteIdent(table))
	pgxRows, err := pool.Query(ctx, sql)
	if err != nil {
		return nil, err
	}
	defer pgxRows.Close()

	descs := pgxRows.FieldDescriptions()
	cols := make([]string, len(descs))
	for i, d := range descs {
		cols[i] = d.Name
	}

	var out []map[string]any
	for pgxRows.Next() {
		vals, err := pgxRows.Values()
		if err != nil {
			return nil, err
		}
		row := make(map[string]any, len(cols))
		for i, col := range cols {
			row[col] = normalizeCell(vals[i])
		}
		out = append(out, row)
	}
	return out, pgxRows.Err()
}

func restoreSupplementalTables(ctx context.Context, tx pgx.Tx, sup *domain.BackupSupplemental) error {
	if sup == nil || len(sup.Tables) == 0 {
		return nil
	}
	// Delete in reverse FK order, then insert in forward order.
	for i := len(supplementalTableOrder) - 1; i >= 0; i-- {
		table := supplementalTableOrder[i]
		if _, ok := sup.Tables[table]; ok {
			if err := deleteAllRows(ctx, tx, table); err != nil {
				return err
			}
		}
	}
	for _, table := range supplementalTableOrder {
		rows, ok := sup.Tables[table]
		if !ok || len(rows) == 0 {
			continue
		}
		if err := insertTableRows(ctx, tx, table, rows); err != nil {
			return fmt.Errorf("restore %s: %w", table, err)
		}
	}
	// Optional tables (traffic) not in supplementalTableOrder.
	for table, rows := range sup.Tables {
		if containsString(supplementalTableOrder, table) || len(rows) == 0 {
			continue
		}
		if !isAllowedBackupTable(table) {
			return fmt.Errorf("table not allowed: %s", table)
		}
		if err := deleteAllRows(ctx, tx, table); err != nil {
			return err
		}
		if err := insertTableRows(ctx, tx, table, rows); err != nil {
			return fmt.Errorf("restore %s: %w", table, err)
		}
	}
	return nil
}

func deleteAllRows(ctx context.Context, tx pgx.Tx, table string) error {
	if !isAllowedBackupTable(table) {
		return fmt.Errorf("table not allowed: %s", table)
	}
	_, err := tx.Exec(ctx, fmt.Sprintf("DELETE FROM %s", quoteIdent(table)))
	return err
}

func insertTableRows(ctx context.Context, tx pgx.Tx, table string, rows []map[string]any) error {
	if !isAllowedBackupTable(table) {
		return fmt.Errorf("table not allowed: %s", table)
	}
	for _, row := range rows {
		cols := make([]string, 0, len(row))
		args := make([]any, 0, len(row))
		placeholders := make([]string, 0, len(row))
		i := 1
		for col, val := range row {
			cols = append(cols, quoteIdent(col))
			args = append(args, denormalizeCell(val))
			placeholders = append(placeholders, fmt.Sprintf("$%d", i))
			i++
		}
		if len(cols) == 0 {
			continue
		}
		sql := fmt.Sprintf(
			"INSERT INTO %s (%s) VALUES (%s)",
			quoteIdent(table),
			strings.Join(cols, ", "),
			strings.Join(placeholders, ", "),
		)
		if _, err := tx.Exec(ctx, sql, args...); err != nil {
			return err
		}
	}
	return nil
}

func wipeOrphanConfigData(ctx context.Context, tx pgx.Tx) error {
	// Tables without FK cascade to users/nodes that would leave orphan rows.
	orphans := []string{
		"traffic_points",
		"sub_hosts",
		"wireguard_peers",
		"orders",
		"admin_quota_notify_events",
	}
	for _, table := range orphans {
		if err := deleteAllRows(ctx, tx, table); err != nil {
			return fmt.Errorf("wipe %s: %w", table, err)
		}
	}
	return nil
}

func exportAdminCredentials(ctx context.Context, pool *pgxpool.Pool) ([]domain.BackupAdminCredential, error) {
	rows, err := pool.Query(ctx, `
		SELECT id, password_hash, totp_secret, webhook_secret FROM admins ORDER BY created_at`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.BackupAdminCredential
	for rows.Next() {
		var c domain.BackupAdminCredential
		if err := rows.Scan(&c.AdminID, &c.PasswordHash, &c.TOTPSecret, &c.WebhookSecret); err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

func isAllowedBackupTable(table string) bool {
	for _, t := range supplementalTableOrder {
		if t == table {
			return true
		}
	}
	for _, t := range optionalTrafficTables {
		if t == table {
			return true
		}
	}
	return false
}

func quoteIdent(name string) string {
	return `"` + strings.ReplaceAll(name, `"`, `""`) + `"`
}

func containsString(list []string, s string) bool {
	for _, v := range list {
		if v == s {
			return true
		}
	}
	return false
}

func normalizeCell(v any) any {
	switch x := v.(type) {
	case []byte:
		if json.Valid(x) {
			var j any
			if json.Unmarshal(x, &j) == nil {
				return j
			}
		}
		return string(x)
	case [16]byte:
		return uuid.UUID(x).String()
	default:
		return v
	}
}

func denormalizeCell(v any) any {
	switch x := v.(type) {
	case string:
		if len(x) == 36 && strings.Count(x, "-") == 4 {
			if id, err := uuid.Parse(x); err == nil {
				return id
			}
		}
		return x
	case map[string]any, []any:
		b, err := json.Marshal(x)
		if err != nil {
			return v
		}
		return b
	default:
		return v
	}
}
