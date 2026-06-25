// Package config loads all runtime configuration from the environment.
// Per the project guardrails, configuration comes from the environment only —
// no secrets in code or committed files. See .env.example for the full set.
package config

import (
	"fmt"
	"log/slog"
	"net"
	"net/url"
	"os"
	"strings"
	"time"
)

// Config holds all settings needed to run the API server.
//
// The DB connection parameters are shared between the control-plane database
// and every tenant database; they differ only by database name. This is what
// lets the TenantResolver build a per-tenant DSN from a tenant's db_name.
type Config struct {
	HTTPAddr string
	LogLevel slog.Level

	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBSSLMode  string

	ControlPlaneDBName string

	// TenantDBPrefix is prepended to a sanitized tenant slug to form the
	// per-tenant logical database name (e.g. "opero_tenant_saigon_tours").
	TenantDBPrefix string

	JWTSecret string
	JWTIssuer string
	JWTTTL    time.Duration

	// CORSAllowedOrigins is the list of browser origins allowed to call the API
	// (comma-separated in CORS_ALLOWED_ORIGINS). Empty = CORS disabled (no
	// headers emitted), which is the safe default for production.
	CORSAllowedOrigins []string

	// Storage holds the object-store settings for media uploads (check-in
	// photos). Backed by MinIO locally; any S3-compatible service in prod.
	Storage StorageConfig
}

// StorageConfig holds object-storage (S3/MinIO) settings for media uploads.
type StorageConfig struct {
	Endpoint  string
	AccessKey string
	SecretKey string
	Bucket    string
	UseSSL    bool

	// PublicURL, if set, is used as the base for returned object URLs instead
	// of presigned GET URLs. Empty = presigned URLs (the local/dev default).
	PublicURL string

	// PresignExpiry is the validity window for presigned GET URLs.
	PresignExpiry time.Duration
}

// Load reads configuration from the environment, applying sensible local
// defaults so the server runs against the Dockerized Postgres out of the box.
func Load() (*Config, error) {
	cfg := &Config{
		HTTPAddr:           getEnv("HTTP_ADDR", ":8080"),
		DBHost:             getEnv("DB_HOST", "localhost"),
		DBPort:             getEnv("DB_PORT", "5432"),
		DBUser:             getEnv("DB_USER", "opero"),
		DBPassword:         getEnv("DB_PASSWORD", "opero"),
		DBSSLMode:          getEnv("DB_SSLMODE", "disable"),
		ControlPlaneDBName: getEnv("CONTROLPLANE_DB_NAME", "opero_control"),
		TenantDBPrefix:     getEnv("TENANT_DB_PREFIX", "opero_tenant_"),
		JWTSecret:          getEnv("JWT_SECRET", ""),
		JWTIssuer:          getEnv("JWT_ISSUER", "opero"),
	}
	cfg.LogLevel = parseLevel(getEnv("LOG_LEVEL", "info"))
	cfg.CORSAllowedOrigins = parseList(getEnv("CORS_ALLOWED_ORIGINS", ""))

	ttl, err := time.ParseDuration(getEnv("JWT_TTL", "24h"))
	if err != nil {
		return nil, fmt.Errorf("config: invalid JWT_TTL: %w", err)
	}
	cfg.JWTTTL = ttl

	presignExpiry, err := time.ParseDuration(getEnv("STORAGE_PRESIGN_EXPIRY", "168h"))
	if err != nil {
		return nil, fmt.Errorf("config: invalid STORAGE_PRESIGN_EXPIRY: %w", err)
	}
	cfg.Storage = StorageConfig{
		Endpoint:      getEnv("STORAGE_ENDPOINT", "localhost:9000"),
		AccessKey:     getEnv("STORAGE_ACCESS_KEY", "opero"),
		SecretKey:     getEnv("STORAGE_SECRET_KEY", "opero-secret"),
		Bucket:        getEnv("STORAGE_BUCKET", "opero-media"),
		UseSSL:        strings.EqualFold(getEnv("STORAGE_USE_SSL", "false"), "true"),
		PublicURL:     getEnv("STORAGE_PUBLIC_URL", ""),
		PresignExpiry: presignExpiry,
	}

	if cfg.DBHost == "" || cfg.DBUser == "" {
		return nil, fmt.Errorf("config: DB_HOST and DB_USER must not be empty")
	}
	if cfg.JWTSecret == "" {
		return nil, fmt.Errorf("config: JWT_SECRET is required")
	}
	return cfg, nil
}

// DSN builds a Postgres connection string for the given database name using
// the shared connection parameters. Used for both the control-plane pool and
// each tenant pool.
func (c *Config) DSN(dbName string) string {
	u := url.URL{
		Scheme: "postgres",
		User:   url.UserPassword(c.DBUser, c.DBPassword),
		Host:   net.JoinHostPort(c.DBHost, c.DBPort),
		Path:   "/" + dbName,
	}
	q := url.Values{}
	q.Set("sslmode", c.DBSSLMode)
	u.RawQuery = q.Encode()
	return u.String()
}

// ControlPlaneDSN is the DSN for the shared control-plane database.
func (c *Config) ControlPlaneDSN() string { return c.DSN(c.ControlPlaneDBName) }

func getEnv(key, def string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return def
}

// parseList splits a comma-separated env value into trimmed, non-empty items.
func parseList(s string) []string {
	if strings.TrimSpace(s) == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if p = strings.TrimSpace(p); p != "" {
			out = append(out, p)
		}
	}
	return out
}

func parseLevel(s string) slog.Level {
	switch strings.ToLower(s) {
	case "debug":
		return slog.LevelDebug
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
