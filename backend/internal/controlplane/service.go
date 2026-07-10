package controlplane

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/mail"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/davidnguyen2205/opero/backend/internal/platform/auth"
)

// repo is the control-plane persistence the service depends on (satisfied by
// *Store). Declared as an interface so the service can be unit-tested with fakes.
type repo interface {
	CreateTenant(ctx context.Context, name, slug, dbName, plan string) (Tenant, error)
	GetTenantByID(ctx context.Context, id uuid.UUID) (Tenant, error)
	GetTenantBySlug(ctx context.Context, slug string) (Tenant, error)
	ListTenants(ctx context.Context) ([]Tenant, error)
	SetTenantStatus(ctx context.Context, id uuid.UUID, status string) (Tenant, error)
	UpdateTenantPlatform(ctx context.Context, id uuid.UUID, name, status, plan *string) (Tenant, error)
	DeleteTenant(ctx context.Context, id uuid.UUID) error
	CreateUser(ctx context.Context, tenantID uuid.UUID, email, passwordHash, role, status string) (User, error)
	DeleteUser(ctx context.Context, id uuid.UUID) error
	GetUserByID(ctx context.Context, id uuid.UUID) (User, error)
	GetUserByTenantAndEmail(ctx context.Context, tenantID uuid.UUID, email string) (User, error)
	ListUsersPlatform(ctx context.Context, tenantID *uuid.UUID, role, status *string) ([]PlatformTenantUser, error)
	UpdateUserStatusPlatform(ctx context.Context, id uuid.UUID, status string) (User, error)
	CreatePlatformUser(ctx context.Context, email, passwordHash, role, status string) (PlatformUser, error)
	GetPlatformUserByID(ctx context.Context, id uuid.UUID) (PlatformUser, error)
	GetPlatformUserByEmail(ctx context.Context, email string) (PlatformUser, error)
	ListSubscriptionsPlatform(ctx context.Context, tenantID *uuid.UUID, plan, status *string) ([]PlatformSubscription, error)
	UpdateSubscriptionPlatform(ctx context.Context, id uuid.UUID, plan, status *string) (PlatformSubscription, error)
	CreateSuperAdminAuditEvent(ctx context.Context, actorID uuid.UUID, action, targetType string, targetID, tenantID *uuid.UUID, metadata map[string]any) error
	ListSuperAdminAuditEvents(ctx context.Context, tenantID, actorID *uuid.UUID, action *string, limit int32) ([]SuperAdminAuditEvent, error)
	CountTenantsByStatus(ctx context.Context) (map[string]int, error)
	// InTx runs fn inside a single transaction, passing a tx-scoped repo. Used
	// to make a platform mutation and its audit event atomic.
	InTx(ctx context.Context, fn func(repo) error) error
}

var validRoles = map[string]bool{"admin": true, "manager": true, "employee": true}
var validPlatformRoles = map[string]bool{"super_admin": true, "support": true, "ops": true}
var validUserStatuses = map[string]bool{"active": true, "disabled": true}
var validTenantStatuses = map[string]bool{"active": true, "suspended": true, "provisioning": true}

type provisioner interface {
	Create(ctx context.Context, dbName string) error
	Drop(ctx context.Context, dbName string) error
}

type tokenIssuer interface {
	Issue(userID, tenantID uuid.UUID, role string, now time.Time) (string, time.Time, error)
	IssuePlatform(platformUserID uuid.UUID, role string, now time.Time) (string, time.Time, error)
}

// Service holds all control-plane business logic.
type Service struct {
	repo     repo
	prov     provisioner
	tokens   tokenIssuer
	dbPrefix string
	logger   *slog.Logger
	now      func() time.Time
}

func NewService(r repo, p provisioner, tokens tokenIssuer, dbPrefix string, logger *slog.Logger) *Service {
	return &Service{
		repo:     r,
		prov:     p,
		tokens:   tokens,
		dbPrefix: dbPrefix,
		logger:   logger,
		now:      time.Now,
	}
}

// Signup creates a tenant, provisions its isolated database, creates the first
// admin user, and returns an access token.
//
// Steps span the control-plane DB and a physical CREATE DATABASE, which cannot
// be one atomic transaction. On failure after the tenant row is created, we
// best-effort clean up (drop DB, delete tenant row) so a failed signup does not
// leave an orphan.
func (s *Service) Signup(ctx context.Context, in SignupInput) (AuthResult, error) {
	slug := in.Slug
	if slug == "" {
		slug = DeriveSlug(in.CompanyName)
	}
	if strings.TrimSpace(in.CompanyName) == "" || !validateSlug(slug) {
		return AuthResult{}, fmt.Errorf("%w: company_name and a valid slug are required", ErrValidation)
	}
	email := strings.TrimSpace(in.AdminEmail)
	if _, err := mail.ParseAddress(email); err != nil {
		return AuthResult{}, fmt.Errorf("%w: admin_email is not a valid email address", ErrValidation)
	}
	if len(in.AdminPassword) < 8 {
		return AuthResult{}, fmt.Errorf("%w: admin_password must be at least 8 characters", ErrValidation)
	}

	dbName := dbNameFromSlug(s.dbPrefix, slug)

	tenant, err := s.repo.CreateTenant(ctx, in.CompanyName, slug, dbName, "free")
	if err != nil {
		return AuthResult{}, err // ErrConflict on duplicate slug
	}

	if err := s.prov.Create(ctx, dbName); err != nil {
		s.cleanup(ctx, tenant.ID, dbName)
		return AuthResult{}, fmt.Errorf("provision tenant database: %w", err)
	}

	hash, err := auth.HashPassword(in.AdminPassword)
	if err != nil {
		s.cleanup(ctx, tenant.ID, dbName)
		return AuthResult{}, err
	}

	user, err := s.repo.CreateUser(ctx, tenant.ID, email, hash, "admin", "active")
	if err != nil {
		s.cleanup(ctx, tenant.ID, dbName)
		return AuthResult{}, err
	}

	tenant, err = s.repo.SetTenantStatus(ctx, tenant.ID, "active")
	if err != nil {
		s.cleanup(ctx, tenant.ID, dbName)
		return AuthResult{}, err
	}

	return s.issue(user, tenant)
}

// Login authenticates a user within a tenant identified by slug.
func (s *Service) Login(ctx context.Context, in LoginInput) (AuthResult, error) {
	tenant, err := s.repo.GetTenantBySlug(ctx, in.TenantSlug)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return AuthResult{}, ErrInvalidCredentials
		}
		return AuthResult{}, err
	}

	user, err := s.repo.GetUserByTenantAndEmail(ctx, tenant.ID, strings.TrimSpace(in.Email))
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return AuthResult{}, ErrInvalidCredentials
		}
		return AuthResult{}, err
	}

	if !auth.CheckPassword(user.PasswordHash, in.Password) {
		return AuthResult{}, ErrInvalidCredentials
	}
	if user.Status != "active" {
		return AuthResult{}, ErrInvalidCredentials
	}

	return s.issue(user, tenant)
}

// PlatformLogin authenticates an Opero staff user. It does not resolve or
// select a tenant database.
func (s *Service) PlatformLogin(ctx context.Context, in PlatformLoginInput) (PlatformAuthResult, error) {
	user, err := s.repo.GetPlatformUserByEmail(ctx, strings.TrimSpace(in.Email))
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return PlatformAuthResult{}, ErrInvalidCredentials
		}
		return PlatformAuthResult{}, err
	}
	if !auth.CheckPassword(user.PasswordHash, in.Password) {
		return PlatformAuthResult{}, ErrInvalidCredentials
	}
	if user.Status != "active" {
		return PlatformAuthResult{}, ErrInvalidCredentials
	}
	return s.issuePlatform(user)
}

// CurrentPlatformUser returns the authenticated Opero staff user.
func (s *Service) CurrentPlatformUser(ctx context.Context, platformUserID uuid.UUID) (PlatformUser, error) {
	return s.repo.GetPlatformUserByID(ctx, platformUserID)
}

// CreatePlatformUser provisions an Opero staff login. This is intentionally
// service-only for now; bootstrap/admin CLI code can use it without exposing an
// API endpoint that creates platform users.
func (s *Service) CreatePlatformUser(ctx context.Context, email, password, role string) (uuid.UUID, error) {
	email = strings.TrimSpace(email)
	if _, err := mail.ParseAddress(email); err != nil {
		return uuid.Nil, fmt.Errorf("%w: invalid email", ErrValidation)
	}
	if len(password) < 12 {
		return uuid.Nil, fmt.Errorf("%w: platform password must be at least 12 characters", ErrValidation)
	}
	if role == "" {
		role = "support"
	}
	if !validPlatformRoles[role] {
		return uuid.Nil, fmt.Errorf("%w: invalid platform role", ErrValidation)
	}
	hash, err := auth.HashPassword(password)
	if err != nil {
		return uuid.Nil, err
	}
	user, err := s.repo.CreatePlatformUser(ctx, email, hash, role, "active")
	if err != nil {
		return uuid.Nil, err
	}
	return user.ID, nil
}

// CreateUser provisions a control-plane login for a tenant and returns its id.
// Used by the identity module (via the UserCreator interface) to give an
// employee a login. Returns ErrConflict if the email is already taken.
func (s *Service) CreateUser(ctx context.Context, tenantID uuid.UUID, email, password, role string) (uuid.UUID, error) {
	email = strings.TrimSpace(email)
	if _, err := mail.ParseAddress(email); err != nil {
		return uuid.Nil, fmt.Errorf("%w: invalid email", ErrValidation)
	}
	if len(password) < 8 {
		return uuid.Nil, fmt.Errorf("%w: password must be at least 8 characters", ErrValidation)
	}
	if role == "" {
		role = "employee"
	}
	if !validRoles[role] {
		return uuid.Nil, fmt.Errorf("%w: invalid role", ErrValidation)
	}
	hash, err := auth.HashPassword(password)
	if err != nil {
		return uuid.Nil, err
	}
	user, err := s.repo.CreateUser(ctx, tenantID, email, hash, role, "active")
	if err != nil {
		return uuid.Nil, err
	}
	return user.ID, nil
}

// DeleteUser removes a control-plane user (used for compensation when linking a
// freshly created login to an employee fails).
func (s *Service) DeleteUser(ctx context.Context, userID uuid.UUID) error {
	return s.repo.DeleteUser(ctx, userID)
}

// CurrentUser returns the authenticated user and their tenant.
func (s *Service) CurrentUser(ctx context.Context, userID uuid.UUID) (CurrentUserResult, error) {
	user, err := s.repo.GetUserByID(ctx, userID)
	if err != nil {
		return CurrentUserResult{}, err
	}
	tenant, err := s.repo.GetTenantByID(ctx, user.TenantID)
	if err != nil {
		return CurrentUserResult{}, err
	}
	return CurrentUserResult{User: user, Tenant: tenant}, nil
}

func (s *Service) PlatformListTenants(ctx context.Context) ([]Tenant, error) {
	return s.repo.ListTenants(ctx)
}

func (s *Service) PlatformGetTenant(ctx context.Context, id uuid.UUID) (Tenant, error) {
	return s.repo.GetTenantByID(ctx, id)
}

func (s *Service) PlatformUpdateTenant(ctx context.Context, actorID, id uuid.UUID, name, status, plan *string) (Tenant, error) {
	name = cleanStringPtr(name)
	status = cleanStringPtr(status)
	plan = cleanStringPtr(plan)
	if name == nil && status == nil && plan == nil {
		return Tenant{}, fmt.Errorf("%w: at least one field is required", ErrValidation)
	}
	if status != nil && !validTenantStatuses[*status] {
		return Tenant{}, fmt.Errorf("%w: invalid tenant status", ErrValidation)
	}
	if name != nil && *name == "" {
		return Tenant{}, fmt.Errorf("%w: tenant name cannot be empty", ErrValidation)
	}
	if plan != nil && *plan == "" {
		return Tenant{}, fmt.Errorf("%w: plan cannot be empty", ErrValidation)
	}
	var tenant Tenant
	err := s.repo.InTx(ctx, func(r repo) error {
		var err error
		tenant, err = r.UpdateTenantPlatform(ctx, id, name, status, plan)
		if err != nil {
			return err
		}
		metadata := changedFields(name, status, plan)
		return r.CreateSuperAdminAuditEvent(ctx, actorID, "tenant.updated", "tenant", &tenant.ID, &tenant.ID, metadata)
	})
	if err != nil {
		return Tenant{}, err
	}
	return tenant, nil
}

func (s *Service) PlatformListUsers(ctx context.Context, tenantID *uuid.UUID, role, status *string) ([]PlatformTenantUser, error) {
	role = cleanStringPtr(role)
	status = cleanStringPtr(status)
	if role != nil && !validRoles[*role] {
		return nil, fmt.Errorf("%w: invalid role", ErrValidation)
	}
	if status != nil && !validUserStatuses[*status] {
		return nil, fmt.Errorf("%w: invalid status", ErrValidation)
	}
	return s.repo.ListUsersPlatform(ctx, tenantID, role, status)
}

func (s *Service) PlatformUpdateUser(ctx context.Context, actorID, userID uuid.UUID, status string) (User, error) {
	status = strings.TrimSpace(status)
	if !validUserStatuses[status] {
		return User{}, fmt.Errorf("%w: invalid status", ErrValidation)
	}
	var user User
	err := s.repo.InTx(ctx, func(r repo) error {
		var err error
		user, err = r.UpdateUserStatusPlatform(ctx, userID, status)
		if err != nil {
			return err
		}
		return r.CreateSuperAdminAuditEvent(ctx, actorID, "user.status_updated", "user", &user.ID, &user.TenantID, map[string]any{"status": status})
	})
	if err != nil {
		return User{}, err
	}
	return user, nil
}

func (s *Service) PlatformListSubscriptions(ctx context.Context, tenantID *uuid.UUID, plan, status *string) ([]PlatformSubscription, error) {
	return s.repo.ListSubscriptionsPlatform(ctx, tenantID, cleanStringPtr(plan), cleanStringPtr(status))
}

func (s *Service) PlatformUpdateSubscription(ctx context.Context, actorID, id uuid.UUID, plan, status *string) (PlatformSubscription, error) {
	plan = cleanStringPtr(plan)
	status = cleanStringPtr(status)
	if plan == nil && status == nil {
		return PlatformSubscription{}, fmt.Errorf("%w: at least one field is required", ErrValidation)
	}
	if plan != nil && *plan == "" {
		return PlatformSubscription{}, fmt.Errorf("%w: plan cannot be empty", ErrValidation)
	}
	if status != nil && *status == "" {
		return PlatformSubscription{}, fmt.Errorf("%w: status cannot be empty", ErrValidation)
	}
	var sub PlatformSubscription
	err := s.repo.InTx(ctx, func(r repo) error {
		var err error
		sub, err = r.UpdateSubscriptionPlatform(ctx, id, plan, status)
		if err != nil {
			return err
		}
		metadata := changedFields(nil, status, plan)
		return r.CreateSuperAdminAuditEvent(ctx, actorID, "subscription.updated", "subscription", &sub.ID, &sub.TenantID, metadata)
	})
	if err != nil {
		return PlatformSubscription{}, err
	}
	// Enrich with tenant display fields (read-only, outside the write tx).
	tenant, err := s.repo.GetTenantByID(ctx, sub.TenantID)
	if err != nil {
		return PlatformSubscription{}, err
	}
	sub.TenantName = tenant.Name
	sub.TenantSlug = tenant.Slug
	return sub, nil
}

func (s *Service) PlatformSystemHealth(ctx context.Context) (SystemHealth, error) {
	counts, err := s.repo.CountTenantsByStatus(ctx)
	if err != nil {
		return SystemHealth{}, err
	}
	return SystemHealth{ControlPlane: "ok", TenantsByStatus: counts}, nil
}

func (s *Service) PlatformListAuditEvents(ctx context.Context, tenantID, actorID *uuid.UUID, action *string, limit int32) ([]SuperAdminAuditEvent, error) {
	if limit <= 0 {
		limit = 100
	}
	if limit > 200 {
		limit = 200
	}
	return s.repo.ListSuperAdminAuditEvents(ctx, tenantID, actorID, cleanStringPtr(action), limit)
}

func (s *Service) issue(user User, tenant Tenant) (AuthResult, error) {
	token, expiresAt, err := s.tokens.Issue(user.ID, tenant.ID, user.Role, s.now())
	if err != nil {
		return AuthResult{}, fmt.Errorf("issue token: %w", err)
	}
	return AuthResult{Token: token, ExpiresAt: expiresAt, User: user, Tenant: tenant}, nil
}

func (s *Service) issuePlatform(user PlatformUser) (PlatformAuthResult, error) {
	token, expiresAt, err := s.tokens.IssuePlatform(user.ID, user.Role, s.now())
	if err != nil {
		return PlatformAuthResult{}, fmt.Errorf("issue platform token: %w", err)
	}
	return PlatformAuthResult{Token: token, ExpiresAt: expiresAt, User: user}, nil
}

func cleanStringPtr(value *string) *string {
	if value == nil {
		return nil
	}
	clean := strings.TrimSpace(*value)
	return &clean
}

func changedFields(name, status, plan *string) map[string]any {
	metadata := map[string]any{}
	if name != nil {
		metadata["name"] = *name
	}
	if status != nil {
		metadata["status"] = *status
	}
	if plan != nil {
		metadata["plan"] = *plan
	}
	return metadata
}

// cleanup is best-effort compensation for a failed signup. It only logs
// failures, so a Drop that fails leaves an orphan tenant DB (acceptable for
// M1). NOTE for M2+: once signup opens a pooled tenant connection (e.g. to seed
// the admin employee), DROP DATABASE will fail while that pool is cached — the
// resolver's pool for dbName must be evicted/closed before Drop.
func (s *Service) cleanup(ctx context.Context, tenantID uuid.UUID, dbName string) {
	if err := s.prov.Drop(ctx, dbName); err != nil {
		s.logger.ErrorContext(ctx, "cleanup: drop tenant db failed",
			slog.String("db", dbName), slog.Any("error", err))
	}
	if err := s.repo.DeleteTenant(ctx, tenantID); err != nil {
		s.logger.ErrorContext(ctx, "cleanup: delete tenant row failed",
			slog.Any("error", err))
	}
}
