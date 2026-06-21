// Package middleware holds cross-cutting HTTP middleware. In M0: recovery and
// request logging. Auth and tenant resolution arrive in M1.
package middleware

import (
	"log/slog"
	"net/http"
)

// Recoverer is the single catch-all for panics in the request path. It logs
// the panic via slog and returns a 500 without leaking internal detail.
func Recoverer(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if rec := recover(); rec != nil {
					logger.LogAttrs(r.Context(), slog.LevelError, "panic_recovered",
						slog.Any("panic", rec),
						slog.String("method", r.Method),
						slog.String("path", r.URL.Path),
					)
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusInternalServerError)
					_, _ = w.Write([]byte(`{"error":"internal server error"}`))
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}
