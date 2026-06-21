// Package httpserver wires the chi router with cross-cutting middleware, the
// /health route, and the oapi-generated API routes.
package httpserver

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/davidnguyen2205/opero/backend/gen/oapi"
	"github.com/davidnguyen2205/opero/backend/internal/platform/auth"
	appmw "github.com/davidnguyen2205/opero/backend/internal/platform/middleware"
)

// Pinger is the minimal database surface the health check needs.
type Pinger interface {
	Ping(ctx context.Context) error
}

// Deps are the dependencies needed to build the HTTP handler.
type Deps struct {
	Logger         *slog.Logger
	ControlPlane   Pinger
	API            oapi.ServerInterface
	Tokens         *auth.TokenManager
	TenantRegistry appmw.TenantRegistry
	TenantPools    appmw.TenantPools

	// CORSAllowedOrigins are browser origins allowed to call the API. Empty
	// disables CORS (no headers).
	CORSAllowedOrigins []string

	// TenantRoutePrefixes are the path prefixes of tenant-data routes, which
	// additionally get TenantMiddleware. This is an explicit allowlist on
	// purpose: a tenant route omitted here fails loud (its service gets no
	// tenant pool and returns an error) rather than a control-plane route
	// silently acquiring one. Control-plane secured routes (e.g. /auth/me) are
	// simply absent from this list and get authentication only.
	TenantRoutePrefixes []string
}

// New builds the HTTP handler: recovery + request logging, /health, and the API
// routes from the generated spec. Authentication is enforced on operations the
// spec marks secured; tenant-data routes additionally get tenant resolution.
func New(d Deps) http.Handler {
	r := chi.NewRouter()

	r.Use(appmw.CORS(d.CORSAllowedOrigins))
	r.Use(appmw.Recoverer(d.Logger))
	r.Use(appmw.RequestLogger(d.Logger))

	r.Get("/health", healthHandler(d.ControlPlane))

	oapi.HandlerWithOptions(d.API, oapi.ChiServerOptions{
		BaseRouter:       r,
		Middlewares:      []oapi.MiddlewareFunc{securedChain(d)},
		ErrorHandlerFunc: paramErrorHandler,
	})

	return r
}

// paramErrorHandler normalizes oapi request-binding failures (bad path/query
// params) to the same JSON Error shape the handlers use, instead of the
// generated default's plain-text response.
func paramErrorHandler(w http.ResponseWriter, _ *http.Request, err error) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)
	_ = json.NewEncoder(w).Encode(oapi.Error{Code: "invalid_request", Message: err.Error()})
}

// securedChain enforces auth (and, for tenant-data routes, tenant resolution)
// only on operations the spec marks secured — the generated wrapper sets
// BearerAuthScopes in context for those. A route gets TenantMiddleware iff its
// path matches one of Deps.TenantRoutePrefixes (see that field's doc).
func securedChain(d Deps) oapi.MiddlewareFunc {
	authn := appmw.Authenticator(d.Tokens, d.Logger)
	tenant := appmw.TenantResolver(d.TenantRegistry, d.TenantPools, d.Logger)
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if _, secured := r.Context().Value(oapi.BearerAuthScopes).([]string); !secured {
				next.ServeHTTP(w, r)
				return
			}
			h := next
			if isTenantRoute(d.TenantRoutePrefixes, r.URL.Path) {
				h = tenant(h)
			}
			authn(h).ServeHTTP(w, r)
		})
	}
}

func isTenantRoute(prefixes []string, path string) bool {
	for _, p := range prefixes {
		if strings.HasPrefix(path, p) {
			return true
		}
	}
	return false
}

func healthHandler(controlPlane Pinger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()

		status := "ok"
		code := http.StatusOK
		if err := controlPlane.Ping(ctx); err != nil {
			status = "degraded"
			code = http.StatusServiceUnavailable
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(code)
		_ = json.NewEncoder(w).Encode(map[string]string{"status": status})
	}
}
