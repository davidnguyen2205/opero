package config

import (
	"testing"
	"time"
)

func TestLoadDefaults(t *testing.T) {
	t.Setenv("JWT_SECRET", "test-secret")
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.HTTPAddr != ":8080" {
		t.Errorf("HTTPAddr = %q, want :8080", cfg.HTTPAddr)
	}
	if cfg.ControlPlaneDBName != "opero_control" {
		t.Errorf("ControlPlaneDBName = %q, want opero_control", cfg.ControlPlaneDBName)
	}
	if cfg.JWTTTL != 24*time.Hour {
		t.Errorf("JWTTTL = %v, want 24h", cfg.JWTTTL)
	}
}

func TestLoadEnvOverride(t *testing.T) {
	t.Setenv("JWT_SECRET", "test-secret")
	t.Setenv("HTTP_ADDR", ":9999")
	t.Setenv("CONTROLPLANE_DB_NAME", "cp")
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.HTTPAddr != ":9999" {
		t.Errorf("HTTPAddr = %q, want :9999", cfg.HTTPAddr)
	}
	if cfg.ControlPlaneDBName != "cp" {
		t.Errorf("ControlPlaneDBName = %q, want cp", cfg.ControlPlaneDBName)
	}
}

func TestLoadRequiresJWTSecret(t *testing.T) {
	t.Setenv("JWT_SECRET", "")
	if _, err := Load(); err == nil {
		t.Fatal("expected error when JWT_SECRET is empty")
	}
}

func TestDSN(t *testing.T) {
	cfg := &Config{DBUser: "u", DBPassword: "p", DBHost: "h", DBPort: "5432", DBSSLMode: "disable"}
	got := cfg.DSN("mydb")
	want := "postgres://u:p@h:5432/mydb?sslmode=disable"
	if got != want {
		t.Errorf("DSN = %q, want %q", got, want)
	}
}
