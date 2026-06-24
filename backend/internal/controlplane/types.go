// Package controlplane owns tenant onboarding and authentication against the
// control-plane database (tenants, users, billing). It never touches a tenant
// database directly; tenant-scoped data is the concern of other modules.
package controlplane

import (
	"encoding/json"
	"errors"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

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
	ID        uuid.UUID
	Name      string
	Slug      string
	DBName    string
	Status    string
	Plan      string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type User struct {
	ID           uuid.UUID
	TenantID     uuid.UUID
	Email        string
	Role         string
	Status       string
	CreatedAt    time.Time
	UpdatedAt    time.Time
	PasswordHash string // internal only; never serialized to API responses
}

type PlatformUser struct {
	ID           uuid.UUID
	Email        string
	Role         string
	Status       string
	CreatedAt    time.Time
	UpdatedAt    time.Time
	PasswordHash string // internal only; never serialized to API responses
}

type PlatformTenantUser struct {
	ID         uuid.UUID
	TenantID   uuid.UUID
	TenantName string
	TenantSlug string
	Email      string
	Role       string
	Status     string
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

type PlatformSubscription struct {
	ID         uuid.UUID
	TenantID   uuid.UUID
	TenantName string
	TenantSlug string
	Plan       string
	Status     string
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

type SuperAdminAuditEvent struct {
	ID                  uuid.UUID
	ActorPlatformUserID uuid.UUID
	ActorEmail          string
	Action              string
	TargetType          string
	TargetID            *uuid.UUID
	TenantID            *uuid.UUID
	TenantName          *string
	TenantSlug          *string
	Metadata            map[string]any
	CreatedAt           time.Time
}

type SystemHealth struct {
	ControlPlane    string
	TenantsByStatus map[string]int
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

type PlatformLoginInput struct {
	Email    string
	Password string
}

type PlatformAuthResult struct {
	Token     string
	ExpiresAt time.Time
	User      PlatformUser
}

type CurrentUserResult struct {
	User   User
	Tenant Tenant
}

func tenantFromDB(t controlplanedb.Tenant) Tenant {
	return Tenant{
		ID:        t.ID,
		Name:      t.Name,
		Slug:      t.Slug,
		DBName:    t.DbName,
		Status:    t.Status,
		Plan:      t.Plan,
		CreatedAt: pgTime(t.CreatedAt),
		UpdatedAt: pgTime(t.UpdatedAt),
	}
}

func userFromDB(u controlplanedb.User) User {
	return User{
		ID:           u.ID,
		TenantID:     u.TenantID,
		Email:        u.Email,
		Role:         u.Role,
		Status:       u.Status,
		CreatedAt:    pgTime(u.CreatedAt),
		UpdatedAt:    pgTime(u.UpdatedAt),
		PasswordHash: u.PasswordHash,
	}
}

func platformUserFromDB(u controlplanedb.PlatformUser) PlatformUser {
	return PlatformUser{
		ID:           u.ID,
		Email:        u.Email,
		Role:         u.Role,
		Status:       u.Status,
		CreatedAt:    pgTime(u.CreatedAt),
		UpdatedAt:    pgTime(u.UpdatedAt),
		PasswordHash: u.PasswordHash,
	}
}

func platformTenantUserFromDB(u controlplanedb.ListUsersPlatformRow) PlatformTenantUser {
	return PlatformTenantUser{
		ID:         u.ID,
		TenantID:   u.TenantID,
		TenantName: u.TenantName,
		TenantSlug: u.TenantSlug,
		Email:      u.Email,
		Role:       u.Role,
		Status:     u.Status,
		CreatedAt:  pgTime(u.CreatedAt),
		UpdatedAt:  pgTime(u.UpdatedAt),
	}
}

func platformSubscriptionFromDB(s controlplanedb.ListSubscriptionsPlatformRow) PlatformSubscription {
	return PlatformSubscription{
		ID:         s.ID,
		TenantID:   s.TenantID,
		TenantName: s.TenantName,
		TenantSlug: s.TenantSlug,
		Plan:       s.Plan,
		Status:     s.Status,
		CreatedAt:  pgTime(s.CreatedAt),
		UpdatedAt:  pgTime(s.UpdatedAt),
	}
}

func subscriptionFromDB(s controlplanedb.Subscription) PlatformSubscription {
	return PlatformSubscription{
		ID:        s.ID,
		TenantID:  s.TenantID,
		Plan:      s.Plan,
		Status:    s.Status,
		CreatedAt: pgTime(s.CreatedAt),
		UpdatedAt: pgTime(s.UpdatedAt),
	}
}

func auditEventFromDB(e controlplanedb.ListSuperAdminAuditEventsRow) SuperAdminAuditEvent {
	var metadata map[string]any
	if len(e.Metadata) > 0 {
		_ = json.Unmarshal(e.Metadata, &metadata)
	}
	if metadata == nil {
		metadata = map[string]any{}
	}
	return SuperAdminAuditEvent{
		ID:                  e.ID,
		ActorPlatformUserID: e.ActorPlatformUserID,
		ActorEmail:          e.ActorEmail,
		Action:              e.Action,
		TargetType:          e.TargetType,
		TargetID:            pgUUIDPtr(e.TargetID),
		TenantID:            pgUUIDPtr(e.TenantID),
		TenantName:          e.TenantName,
		TenantSlug:          e.TenantSlug,
		Metadata:            metadata,
		CreatedAt:           pgTime(e.CreatedAt),
	}
}

func pgTime(t pgtype.Timestamptz) time.Time {
	if !t.Valid {
		return time.Time{}
	}
	return t.Time
}

func pgUUIDPtr(u pgtype.UUID) *uuid.UUID {
	if !u.Valid {
		return nil
	}
	id := uuid.UUID(u.Bytes)
	return &id
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
