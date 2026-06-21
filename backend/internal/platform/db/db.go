// Package db owns database connectivity: the shared control-plane pool and a
// resolver that returns one cached connection pool per tenant database.
//
// Tenant isolation guarantee: the resolver maps a tenant's logical database
// name (db_name, read from the control-plane registry by TenantMiddleware) to
// a pool. It never decides which tenant a request belongs to. Services receive
// the tenant pool through the request context and must never open one ad hoc.
package db

import (
	"context"
	"fmt"
	"sync"

	"github.com/jackc/pgx/v5/pgxpool"
)

// NewControlPlanePool creates and verifies the shared control-plane pool.
func NewControlPlanePool(ctx context.Context, dsn string) (*pgxpool.Pool, error) {
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, fmt.Errorf("create control-plane pool: %w", err)
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping control-plane db: %w", err)
	}
	return pool, nil
}

// TenantResolver resolves and caches one *pgxpool.Pool per tenant database,
// keyed by the tenant's logical database name.
type TenantResolver struct {
	buildDSN func(dbName string) string

	mu    sync.RWMutex
	pools map[string]*pgxpool.Pool
}

// NewTenantResolver builds a resolver. buildDSN turns a tenant's db_name into a
// connection string (typically config.Config.DSN).
func NewTenantResolver(buildDSN func(dbName string) string) *TenantResolver {
	return &TenantResolver{
		buildDSN: buildDSN,
		pools:    make(map[string]*pgxpool.Pool),
	}
}

// Pool returns a cached pool for dbName, creating and verifying one on first
// use. Safe for concurrent use.
func (r *TenantResolver) Pool(ctx context.Context, dbName string) (*pgxpool.Pool, error) {
	if dbName == "" {
		return nil, fmt.Errorf("tenant resolver: empty db name")
	}

	r.mu.RLock()
	p, ok := r.pools[dbName]
	r.mu.RUnlock()
	if ok {
		return p, nil
	}

	r.mu.Lock()
	defer r.mu.Unlock()
	// Re-check: another goroutine may have created it between locks.
	if p, ok := r.pools[dbName]; ok {
		return p, nil
	}

	pool, err := pgxpool.New(ctx, r.buildDSN(dbName))
	if err != nil {
		return nil, fmt.Errorf("create tenant pool for %q: %w", dbName, err)
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping tenant db %q: %w", dbName, err)
	}
	r.pools[dbName] = pool
	return pool, nil
}

// Close closes all cached tenant pools.
func (r *TenantResolver) Close() {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, p := range r.pools {
		p.Close()
	}
	r.pools = make(map[string]*pgxpool.Pool)
}
