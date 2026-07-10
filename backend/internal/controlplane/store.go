package controlplane

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	controlplanedb "github.com/davidnguyen2205/opero/backend/gen/sqlc/controlplane"
)

// Store is the only place that touches the control-plane database. It wraps the
// sqlc-generated queries and maps DB errors to the module's sentinel errors.
type Store struct {
	// pool is nil on a transaction-scoped store returned inside InTx.
	pool *pgxpool.Pool
	q    *controlplanedb.Queries
}

func NewStore(pool *pgxpool.Pool) *Store {
	return &Store{pool: pool, q: controlplanedb.New(pool)}
}

// InTx runs fn inside a single control-plane transaction. The repo passed to fn
// is scoped to that transaction, so a mutation and its audit event commit
// together or roll back together. Committing only after fn succeeds is what
// guarantees a state change is never persisted without its audit record.
func (s *Store) InTx(ctx context.Context, fn func(repo) error) error {
	if s.pool == nil {
		return fmt.Errorf("InTx: store is not pool-backed")
	}
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx) //nolint:errcheck // no-op once committed
	if err := fn(&Store{q: s.q.WithTx(tx)}); err != nil {
		return err
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit tx: %w", err)
	}
	return nil
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

func pgUUIDFromPtr(id *uuid.UUID) pgtype.UUID {
	if id == nil {
		return pgtype.UUID{}
	}
	return pgtype.UUID{Bytes: *id, Valid: true}
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

func (s *Store) ListTenants(ctx context.Context) ([]Tenant, error) {
	rows, err := s.q.ListTenants(ctx)
	if err != nil {
		return nil, fmt.Errorf("list tenants: %w", mapErr(err))
	}
	tenants := make([]Tenant, 0, len(rows))
	for _, row := range rows {
		tenants = append(tenants, tenantFromDB(row))
	}
	return tenants, nil
}

func (s *Store) SetTenantStatus(ctx context.Context, id uuid.UUID, status string) (Tenant, error) {
	t, err := s.q.SetTenantStatus(ctx, controlplanedb.SetTenantStatusParams{ID: id, Status: status})
	if err != nil {
		return Tenant{}, fmt.Errorf("set tenant status: %w", mapErr(err))
	}
	return tenantFromDB(t), nil
}

func (s *Store) UpdateTenantPlatform(ctx context.Context, id uuid.UUID, name, status, plan *string) (Tenant, error) {
	t, err := s.q.UpdateTenantPlatform(ctx, controlplanedb.UpdateTenantPlatformParams{
		ID:     id,
		Name:   name,
		Status: status,
		Plan:   plan,
	})
	if err != nil {
		return Tenant{}, fmt.Errorf("update tenant platform: %w", mapErr(err))
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

func (s *Store) ListUsersPlatform(ctx context.Context, tenantID *uuid.UUID, role, status *string) ([]PlatformTenantUser, error) {
	rows, err := s.q.ListUsersPlatform(ctx, controlplanedb.ListUsersPlatformParams{
		TenantID: pgUUIDFromPtr(tenantID),
		Role:     role,
		Status:   status,
	})
	if err != nil {
		return nil, fmt.Errorf("list users platform: %w", mapErr(err))
	}
	users := make([]PlatformTenantUser, 0, len(rows))
	for _, row := range rows {
		users = append(users, platformTenantUserFromDB(row))
	}
	return users, nil
}

func (s *Store) UpdateUserStatusPlatform(ctx context.Context, id uuid.UUID, status string) (User, error) {
	u, err := s.q.UpdateUserStatusPlatform(ctx, controlplanedb.UpdateUserStatusPlatformParams{
		ID:     id,
		Status: status,
	})
	if err != nil {
		return User{}, fmt.Errorf("update user status platform: %w", mapErr(err))
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

func (s *Store) CreatePlatformUser(ctx context.Context, email, passwordHash, role, status string) (PlatformUser, error) {
	u, err := s.q.CreatePlatformUser(ctx, controlplanedb.CreatePlatformUserParams{
		Email:        email,
		PasswordHash: passwordHash,
		Role:         role,
		Status:       status,
	})
	if err != nil {
		return PlatformUser{}, fmt.Errorf("create platform user: %w", mapErr(err))
	}
	return platformUserFromDB(u), nil
}

func (s *Store) GetPlatformUserByID(ctx context.Context, id uuid.UUID) (PlatformUser, error) {
	u, err := s.q.GetPlatformUserByID(ctx, id)
	if err != nil {
		return PlatformUser{}, fmt.Errorf("get platform user by id: %w", mapErr(err))
	}
	return platformUserFromDB(u), nil
}

func (s *Store) GetPlatformUserByEmail(ctx context.Context, email string) (PlatformUser, error) {
	u, err := s.q.GetPlatformUserByEmail(ctx, email)
	if err != nil {
		return PlatformUser{}, fmt.Errorf("get platform user by email: %w", mapErr(err))
	}
	return platformUserFromDB(u), nil
}

func (s *Store) ListSubscriptionsPlatform(ctx context.Context, tenantID *uuid.UUID, plan, status *string) ([]PlatformSubscription, error) {
	rows, err := s.q.ListSubscriptionsPlatform(ctx, controlplanedb.ListSubscriptionsPlatformParams{
		TenantID: pgUUIDFromPtr(tenantID),
		Plan:     plan,
		Status:   status,
	})
	if err != nil {
		return nil, fmt.Errorf("list subscriptions platform: %w", mapErr(err))
	}
	subscriptions := make([]PlatformSubscription, 0, len(rows))
	for _, row := range rows {
		subscriptions = append(subscriptions, platformSubscriptionFromDB(row))
	}
	return subscriptions, nil
}

func (s *Store) UpdateSubscriptionPlatform(ctx context.Context, id uuid.UUID, plan, status *string) (PlatformSubscription, error) {
	sub, err := s.q.UpdateSubscriptionPlatform(ctx, controlplanedb.UpdateSubscriptionPlatformParams{
		ID:     id,
		Plan:   plan,
		Status: status,
	})
	if err != nil {
		return PlatformSubscription{}, fmt.Errorf("update subscription platform: %w", mapErr(err))
	}
	return subscriptionFromDB(sub), nil
}

func (s *Store) CreateSuperAdminAuditEvent(ctx context.Context, actorID uuid.UUID, action, targetType string, targetID, tenantID *uuid.UUID, metadata map[string]any) error {
	if metadata == nil {
		metadata = map[string]any{}
	}
	raw, err := json.Marshal(metadata)
	if err != nil {
		return fmt.Errorf("marshal audit metadata: %w", err)
	}
	_, err = s.q.CreateSuperAdminAuditEvent(ctx, controlplanedb.CreateSuperAdminAuditEventParams{
		ActorPlatformUserID: actorID,
		Action:              action,
		TargetType:          targetType,
		TargetID:            pgUUIDFromPtr(targetID),
		TenantID:            pgUUIDFromPtr(tenantID),
		Metadata:            raw,
	})
	if err != nil {
		return fmt.Errorf("create super admin audit event: %w", mapErr(err))
	}
	return nil
}

func (s *Store) ListSuperAdminAuditEvents(ctx context.Context, tenantID, actorID *uuid.UUID, action *string, limit int32) ([]SuperAdminAuditEvent, error) {
	rows, err := s.q.ListSuperAdminAuditEvents(ctx, controlplanedb.ListSuperAdminAuditEventsParams{
		TenantID:            pgUUIDFromPtr(tenantID),
		ActorPlatformUserID: pgUUIDFromPtr(actorID),
		Action:              action,
		Limit:               limit,
	})
	if err != nil {
		return nil, fmt.Errorf("list super admin audit events: %w", mapErr(err))
	}
	events := make([]SuperAdminAuditEvent, 0, len(rows))
	for _, row := range rows {
		events = append(events, auditEventFromDB(row))
	}
	return events, nil
}

func (s *Store) CountTenantsByStatus(ctx context.Context) (map[string]int, error) {
	rows, err := s.q.CountTenantsByStatus(ctx)
	if err != nil {
		return nil, fmt.Errorf("count tenants by status: %w", mapErr(err))
	}
	counts := make(map[string]int, len(rows))
	for _, row := range rows {
		counts[row.Status] = int(row.Count)
	}
	return counts, nil
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
