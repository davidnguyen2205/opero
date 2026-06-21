// Package controlplane owns tenant onboarding and authentication against the
// control-plane database (tenants, users, billing). It never touches a tenant
// database directly; tenant-scoped data is the concern of other modules.
package controlplane

import (
	"errors"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"

	controlplanedb "github.com/davidnguyen2205/opero/backend/gen/sqlc/controlplane"
)

// Sentinel errors mapped to HTTP statuses by the handler.
var (
	ErrValidation         = errors.New("validation failed")
	ErrConflict           = errors.New("already exists")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrNotFound           = errors.New("not found")
)

// Domain types. These are the module's own types, mapped to/from the generated
// sqlc and API types so neither leaks across the boundary.

type Tenant struct {
	ID     uuid.UUID
	Name   string
	Slug   string
	Status string
	Plan   string
}

type User struct {
	ID           uuid.UUID
	TenantID     uuid.UUID
	Email        string
	Role         string
	Status       string
	PasswordHash string // internal only; never serialized to API responses
}

type SignupInput struct {
	CompanyName   string
	Slug          string // optional; derived from CompanyName when empty
	AdminFullName string
	AdminEmail    string
	AdminPassword string
}

type LoginInput struct {
	TenantSlug string
	Email      string
	Password   string
}

type AuthResult struct {
	Token     string
	ExpiresAt time.Time
	User      User
	Tenant    Tenant
}

type CurrentUserResult struct {
	User   User
	Tenant Tenant
}

func tenantFromDB(t controlplanedb.Tenant) Tenant {
	return Tenant{ID: t.ID, Name: t.Name, Slug: t.Slug, Status: t.Status, Plan: t.Plan}
}

func userFromDB(u controlplanedb.User) User {
	return User{
		ID:           u.ID,
		TenantID:     u.TenantID,
		Email:        u.Email,
		Role:         u.Role,
		Status:       u.Status,
		PasswordHash: u.PasswordHash,
	}
}

var slugRe = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)

// DeriveSlug turns a free-text company name into a URL/login-friendly slug.
func DeriveSlug(name string) string {
	name = strings.ToLower(strings.TrimSpace(name))
	var b strings.Builder
	prevHyphen := false
	for _, r := range name {
		switch {
		case (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9'):
			b.WriteRune(r)
			prevHyphen = false
		default:
			if b.Len() > 0 && !prevHyphen {
				b.WriteByte('-')
				prevHyphen = true
			}
		}
	}
	return strings.Trim(b.String(), "-")
}

func validateSlug(slug string) bool {
	return len(slug) >= 2 && len(slug) <= 63 && slugRe.MatchString(slug)
}

// dbNameFromSlug builds the tenant's logical database name. Since a valid slug
// is restricted to [a-z0-9-], replacing hyphens with underscores yields a safe
// Postgres identifier.
func dbNameFromSlug(prefix, slug string) string {
	return prefix + strings.ReplaceAll(slug, "-", "_")
}
