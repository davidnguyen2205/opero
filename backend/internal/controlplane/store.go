package controlplane

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	controlplanedb "github.com/davidnguyen2205/opero/backend/gen/sqlc/controlplane"
)

// Store is the only place that touches the control-plane database. It wraps the
// sqlc-generated queries and maps DB errors to the module's sentinel errors.
type Store struct {
	q *controlplanedb.Queries
}

func NewStore(pool *pgxpool.Pool) *Store {
	return &Store{q: controlplanedb.New(pool)}
}

func mapErr(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrNotFound
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" { // unique_violation
		return ErrConflict
	}
	return err
}

func (s *Store) CreateTenant(ctx context.Context, name, slug, dbName, plan string) (Tenant, error) {
	t, err := s.q.CreateTenant(ctx, controlplanedb.CreateTenantParams{
		Name:   name,
		Slug:   slug,
		DbName: dbName,
		Status: "provisioning",
		Plan:   plan,
	})
	if err != nil {
		return Tenant{}, fmt.Errorf("create tenant: %w", mapErr(err))
	}
	return tenantFromDB(t), nil
}

func (s *Store) GetTenantByID(ctx context.Context, id uuid.UUID) (Tenant, error) {
	t, err := s.q.GetTenantByID(ctx, id)
	if err != nil {
		return Tenant{}, fmt.Errorf("get tenant by id: %w", mapErr(err))
	}
	return tenantFromDB(t), nil
}

func (s *Store) GetTenantBySlug(ctx context.Context, slug string) (Tenant, error) {
	t, err := s.q.GetTenantBySlug(ctx, slug)
	if err != nil {
		return Tenant{}, fmt.Errorf("get tenant by slug: %w", mapErr(err))
	}
	return tenantFromDB(t), nil
}

func (s *Store) SetTenantStatus(ctx context.Context, id uuid.UUID, status string) (Tenant, error) {
	t, err := s.q.SetTenantStatus(ctx, controlplanedb.SetTenantStatusParams{ID: id, Status: status})
	if err != nil {
		return Tenant{}, fmt.Errorf("set tenant status: %w", mapErr(err))
	}
	return tenantFromDB(t), nil
}

func (s *Store) DeleteTenant(ctx context.Context, id uuid.UUID) error {
	if err := s.q.DeleteTenant(ctx, id); err != nil {
		return fmt.Errorf("delete tenant: %w", mapErr(err))
	}
	return nil
}

func (s *Store) CreateUser(ctx context.Context, tenantID uuid.UUID, email, passwordHash, role, status string) (User, error) {
	u, err := s.q.CreateUser(ctx, controlplanedb.CreateUserParams{
		TenantID:     tenantID,
		Email:        email,
		PasswordHash: passwordHash,
		Role:         role,
		Status:       status,
	})
	if err != nil {
		return User{}, fmt.Errorf("create user: %w", mapErr(err))
	}
	return userFromDB(u), nil
}

func (s *Store) GetUserByID(ctx context.Context, id uuid.UUID) (User, error) {
	u, err := s.q.GetUserByID(ctx, id)
	if err != nil {
		return User{}, fmt.Errorf("get user by id: %w", mapErr(err))
	}
	return userFromDB(u), nil
}

func (s *Store) DeleteUser(ctx context.Context, id uuid.UUID) error {
	if err := s.q.DeleteUser(ctx, id); err != nil {
		return fmt.Errorf("delete user: %w", mapErr(err))
	}
	return nil
}

func (s *Store) GetUserByTenantAndEmail(ctx context.Context, tenantID uuid.UUID, email string) (User, error) {
	u, err := s.q.GetUserByTenantAndEmail(ctx, controlplanedb.GetUserByTenantAndEmailParams{
		TenantID: tenantID,
		Lower:    email,
	})
	if err != nil {
		return User{}, fmt.Errorf("get user by tenant and email: %w", mapErr(err))
	}
	return userFromDB(u), nil
}

// TenantDBName returns the logical database name for a tenant. It satisfies the
// TenantRegistry interface used by TenantMiddleware (M2+).
func (s *Store) TenantDBName(ctx context.Context, tenantID uuid.UUID) (string, error) {
	t, err := s.q.GetTenantByID(ctx, tenantID)
	if err != nil {
		return "", fmt.Errorf("lookup tenant db name: %w", mapErr(err))
	}
	return t.DbName, nil
}
