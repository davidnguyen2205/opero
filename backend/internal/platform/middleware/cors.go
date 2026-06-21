package middleware

import "net/http"

// CORS returns middleware that allows the configured browser origins to call
// the API with credentials (the bearer token). If allowedOrigins is empty it is
// a no-op (no CORS headers), which is the safe production default for an API
// with no browser clients on other origins.
//
// This is a deliberately small, dependency-free implementation: exact-match
// origin allowlist, the methods/headers Opero uses, and preflight handling.
func CORS(allowedOrigins []string) func(http.Handler) http.Handler {
	allowed := make(map[string]bool, len(allowedOrigins))
	for _, o := range allowedOrigins {
		allowed[o] = true
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")
			if origin != "" && allowed[origin] {
				h := w.Header()
				h.Set("Access-Control-Allow-Origin", origin)
				h.Add("Vary", "Origin")
				h.Set("Access-Control-Allow-Methods", "GET, POST, PATCH, DELETE, OPTIONS")
				h.Set("Access-Control-Allow-Headers", "Authorization, Content-Type")
				h.Set("Access-Control-Max-Age", "300")

				// Preflight: respond immediately, don't fall through to the route.
				if r.Method == http.MethodOptions {
					w.WriteHeader(http.StatusNoContent)
					return
				}
			} else if r.Method == http.MethodOptions && origin != "" {
				// Disallowed origin preflight — reject without leaking headers.
				w.WriteHeader(http.StatusForbidden)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
