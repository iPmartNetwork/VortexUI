package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// PoolConfig holds configurable connection pool settings.
type PoolConfig struct {
	// DSN is the PostgreSQL connection string.
	DSN string
	// MinConns is the minimum number of idle connections.
	MinConns int32
	// MaxConns is the maximum number of connections.
	MaxConns int32
	// MaxConnLifetime is the maximum time a connection stays open.
	MaxConnLifetime time.Duration
	// MaxConnIdleTime is the maximum time a connection stays idle before being closed.
	MaxConnIdleTime time.Duration
	// HealthCheckPeriod is how often idle connections are health-checked.
	HealthCheckPeriod time.Duration
}

// DefaultPoolConfig returns sensible defaults for a production pool.
func DefaultPoolConfig(dsn string) PoolConfig {
	return PoolConfig{
		DSN:               dsn,
		MinConns:          5,
		MaxConns:          25,
		MaxConnLifetime:   30 * time.Minute,
		MaxConnIdleTime:   5 * time.Minute,
		HealthCheckPeriod: 30 * time.Second,
	}
}

// NewPool creates a configured pgxpool.Pool from PoolConfig.
func NewPool(ctx context.Context, cfg PoolConfig) (*pgxpool.Pool, error) {
	poolCfg, err := pgxpool.ParseConfig(cfg.DSN)
	if err != nil {
		return nil, fmt.Errorf("parse pool config: %w", err)
	}

	if cfg.MinConns > 0 {
		poolCfg.MinConns = cfg.MinConns
	}
	if cfg.MaxConns > 0 {
		poolCfg.MaxConns = cfg.MaxConns
	}
	if cfg.MaxConnLifetime > 0 {
		poolCfg.MaxConnLifetime = cfg.MaxConnLifetime
	}
	if cfg.MaxConnIdleTime > 0 {
		poolCfg.MaxConnIdleTime = cfg.MaxConnIdleTime
	}
	if cfg.HealthCheckPeriod > 0 {
		poolCfg.HealthCheckPeriod = cfg.HealthCheckPeriod
	}

	pool, err := pgxpool.NewWithConfig(ctx, poolCfg)
	if err != nil {
		return nil, fmt.Errorf("create pool: %w", err)
	}

	// Verify connectivity
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping pool: %w", err)
	}

	return pool, nil
}
