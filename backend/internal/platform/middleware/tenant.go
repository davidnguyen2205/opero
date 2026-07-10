package middleware

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// TenantRegistry resolves a tenant's logical database name. Satisfied by the
// controlplane Store.
type TenantRegistry interface {
	TenantDBName(ctx context.Context, tenantID uuid.UUID) (string, error)
}

// TenantPools returns a cached connection pool for a tenant database. Satisfied
// by the platform db.TenantResolver.
type TenantPools interface {
	Pool(ctx context.Context, dbName string) (*pgxpool.Pool, error)
}

// TenantResolver resolves the caller's tenant database from their token claims
// and attaches the tenant-scoped pool to the request context. This is the only
// sanctioned way a tenant pool enters a request — services read it from the
// context and must never open one ad hoc.
//
// Apply this only to tenant-scoped routes, after Authenticator. It is wired to
// routes starting in M2; no M1 endpoint reads tenant data.
func TenantResolver(reg TenantRegistry, pools TenantPools, logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims, ok := ClaimsFromContext(r.Context())
			if !ok {
				WriteUnauthorized(w)
				return
			}
			dbName, err := reg.TenantDBName(r.Context(), claims.TenantIDValue())
			if err != nil {
				logger.ErrorContext(r.Context(), "tenant db lookup failed", slog.Any("error", err))
				writeServiceUnavailable(w)
				return
			}
			pool, err := pools.Pool(r.Context(), dbName)
			if err != nil {
				logger.ErrorContext(r.Context(), "tenant pool resolve failed", slog.Any("error", err))
				writeServiceUnavailable(w)
				return
			}
			ctx := context.WithValue(r.Context(), tenantPoolContextKey, pool)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// TenantPoolFromContext returns the tenant-scoped pool placed by TenantResolver.
func TenantPoolFromContext(ctx context.Context) (*pgxpool.Pool, bool) {
	p, ok := ctx.Value(tenantPoolContextKey).(*pgxpool.Pool)
	return p, ok
}

func writeServiceUnavailable(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusServiceUnavailable)
	_ = json.NewEncoder(w).Encode(map[string]string{
		"code":    "unavailable",
		"message": "tenant temporarily unavailable",
	})
}
