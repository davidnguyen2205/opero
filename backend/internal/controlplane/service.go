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
	SetTenantStatus(ctx context.Context, id uuid.UUID, status string) (Tenant, error)
	DeleteTenant(ctx context.Context, id uuid.UUID) error
	CreateUser(ctx context.Context, tenantID uuid.UUID, email, passwordHash, role, status string) (User, error)
	DeleteUser(ctx context.Context, id uuid.UUID) error
	GetUserByID(ctx context.Context, id uuid.UUID) (User, error)
	GetUserByTenantAndEmail(ctx context.Context, tenantID uuid.UUID, email string) (User, error)
}

var validRoles = map[string]bool{"admin": true, "manager": true, "employee": true}

type provisioner interface {
	Create(ctx context.Context, dbName string) error
	Drop(ctx context.Context, dbName string) error
}

type tokenIssuer interface {
	Issue(userID, tenantID uuid.UUID, role string, now time.Time) (string, time.Time, error)
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

func (s *Service) issue(user User, tenant Tenant) (AuthResult, error) {
	token, expiresAt, err := s.tokens.Issue(user.ID, tenant.ID, user.Role, s.now())
	if err != nil {
		return AuthResult{}, fmt.Errorf("issue token: %w", err)
	}
	return AuthResult{Token: token, ExpiresAt: expiresAt, User: user, Tenant: tenant}, nil
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
