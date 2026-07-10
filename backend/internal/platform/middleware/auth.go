package middleware

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"

	"github.com/davidnguyen2205/opero/backend/internal/platform/auth"
)

type contextKey int

const (
	claimsContextKey contextKey = iota
	tenantPoolContextKey
)

// Authenticator validates the bearer token and places the claims in the request
// context. It always enforces a valid token, so apply it only to routes that
// require authentication (see the spec-driven gating in the httpserver package).
func Authenticator(tm *auth.TokenManager, logger *slog.Logger) func(http.Handler) http.Handler {
	return authenticator(tm, logger, "")
}

// TenantAuthenticator validates a tenant token and rejects platform tokens.
func TenantAuthenticator(tm *auth.TokenManager, logger *slog.Logger) func(http.Handler) http.Handler {
	return authenticator(tm, logger, "tenant")
}

// PlatformAuthenticator validates a platform token and rejects tenant tokens.
func PlatformAuthenticator(tm *auth.TokenManager, logger *slog.Logger) func(http.Handler) http.Handler {
	return authenticator(tm, logger, "platform")
}

func authenticator(tm *auth.TokenManager, logger *slog.Logger, requiredKind string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			raw := bearerToken(r.Header.Get("Authorization"))
			if raw == "" {
				WriteUnauthorized(w)
				return
			}
			claims, err := tm.Parse(raw)
			if err != nil {
				logger.DebugContext(r.Context(), "token parse failed", slog.Any("error", err))
				WriteUnauthorized(w)
				return
			}
			if requiredKind != "" && claims.Kind != requiredKind {
				logger.DebugContext(r.Context(), "token kind rejected",
					slog.String("required_kind", requiredKind),
					slog.String("actual_kind", claims.Kind))
				WriteUnauthorized(w)
				return
			}
			ctx := context.WithValue(r.Context(), claimsContextKey, claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// ClaimsFromContext returns the authenticated claims, if present.
func ClaimsFromContext(ctx context.Context) (*auth.Claims, bool) {
	c, ok := ctx.Value(claimsContextKey).(*auth.Claims)
	return c, ok
}

func bearerToken(header string) string {
	if after, ok := strings.CutPrefix(header, "Bearer "); ok {
		return strings.TrimSpace(after)
	}
	return ""
}

// WriteUnauthorized writes a 401 JSON error.
func WriteUnauthorized(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	_ = json.NewEncoder(w).Encode(map[string]string{
		"code":    "unauthorized",
		"message": "authentication required",
	})
}

// WriteForbidden writes a 403 JSON error for an authenticated caller whose role
// does not satisfy the operation's minimum-role requirement.
func WriteForbidden(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusForbidden)
	_ = json.NewEncoder(w).Encode(map[string]string{
		"code":    "forbidden",
		"message": "insufficient role",
	})
}
