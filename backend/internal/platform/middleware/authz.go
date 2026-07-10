package middleware

import (
	"log/slog"
	"net/http"
)

// roleRank is the coarse tenant-role hierarchy: admin > manager > employee.
// An unknown or empty role has rank 0 and therefore satisfies no requirement.
var roleRank = map[string]int{
	"employee": 1,
	"manager":  2,
	"admin":    3,
}

// roleSatisfies reports whether a caller with the actual role meets the required
// minimum-role scopes. An empty requirement means "any authenticated user" and
// always passes. When multiple scopes are present the most restrictive (highest
// rank) wins. An unknown/empty actual role never satisfies a non-empty
// requirement.
func roleSatisfies(actual string, required []string) bool {
	if len(required) == 0 {
		return true
	}
	need := 0
	for _, r := range required {
		if rank := roleRank[r]; rank > need {
			need = rank
		}
	}
	return roleRank[actual] >= need
}

// Authorizer returns middleware that enforces the coarse minimum-role gate for
// an operation. The required scopes come from the spec (threaded by the
// generated wrapper and read in the httpserver package) and are passed in as a
// plain slice so this package stays free of the generated API dependency.
//
// It must run AFTER authentication so the caller's claims are present in
// context. No claims -> 401; role below the requirement -> 403.
func Authorizer(logger *slog.Logger, required []string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims, ok := ClaimsFromContext(r.Context())
			if !ok {
				WriteUnauthorized(w)
				return
			}
			if !roleSatisfies(claims.Role, required) {
				logger.DebugContext(r.Context(), "role authorization denied",
					slog.String("actual_role", claims.Role),
					slog.Any("required_roles", required))
				WriteForbidden(w)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
