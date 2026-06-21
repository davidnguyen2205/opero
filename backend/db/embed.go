// Package dbassets embeds the SQL migration files so the binary can run
// migrations (control-plane and per-tenant fan-out) without shipping loose
// files. The raw .sql files remain the source of truth under db/migrations.
package dbassets

import "embed"

// FS contains the migrations tree:
//
//	migrations/controlplane/*.sql
//	migrations/tenant/*.sql
//
//go:embed all:migrations
var FS embed.FS

// Migration directories within FS.
const (
	ControlPlaneDir = "migrations/controlplane"
	TenantDir       = "migrations/tenant"
)
