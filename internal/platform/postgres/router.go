package postgres

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Querier is the common interface for query execution (pool or conn).
type Querier interface {
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
}

// QueryRouter routes read queries to a read replica (when available) and write
// queries to the primary. If no replica is configured, all queries go to primary.
type QueryRouter struct {
	writePool *pgxpool.Pool
	readPool  *pgxpool.Pool
}

// NewQueryRouter creates a router. readPool can be nil (all queries to primary).
func NewQueryRouter(writePool, readPool *pgxpool.Pool) *QueryRouter {
	return &QueryRouter{
		writePool: writePool,
		readPool:  readPool,
	}
}

// ReadPool returns the pool for read-heavy queries.
// Falls back to writePool if no replica is configured.
func (r *QueryRouter) ReadPool() *pgxpool.Pool {
	if r.readPool != nil {
		return r.readPool
	}
	return r.writePool
}

// WritePool returns the primary pool for write operations.
func (r *QueryRouter) WritePool() *pgxpool.Pool {
	return r.writePool
}

// Reader returns a Querier for read operations.
func (r *QueryRouter) Reader() Querier {
	return r.ReadPool()
}

// Writer returns a Querier for write operations.
func (r *QueryRouter) Writer() Querier {
	return r.writePool
}

// Close closes both pools.
func (r *QueryRouter) Close() {
	r.writePool.Close()
	if r.readPool != nil && r.readPool != r.writePool {
		r.readPool.Close()
	}
}
