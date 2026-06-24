package tours

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/google/uuid"

	appmw "github.com/davidnguyen2205/opero/backend/internal/platform/middleware"
)

// repo is the tenant-DB persistence the service needs (satisfied by *Store).
type repo interface {
	Create(ctx context.Context, in CreateInput) (Tour, error)
	Get(ctx context.Context, id uuid.UUID) (Tour, error)
	List(ctx context.Context, f Filter) ([]Tour, error)
	Update(ctx context.Context, id uuid.UUID, in UpdateInput) (Tour, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

// Service holds tour business logic. The tenant store is resolved per request
// from the context pool (placed by TenantMiddleware) via newStore.
type Service struct {
	newStore func(ctx context.Context) (repo, error)
	logger   *slog.Logger
}

func NewService(logger *slog.Logger) *Service {
	s := &Service{logger: logger}
	s.newStore = s.tenantStore
	return s
}

func (s *Service) tenantStore(ctx context.Context) (repo, error) {
	pool, ok := appmw.TenantPoolFromContext(ctx)
	if !ok {
		return nil, ErrNoTenant
	}
	return NewStore(pool), nil
}

func (s *Service) Create(ctx context.Context, in CreateInput) (Tour, error) {
	if strings.TrimSpace(in.Name) == "" {
		return Tour{}, fmt.Errorf("%w: name is required", ErrValidation)
	}
	if !ValidCategory(in.Category) {
		return Tour{}, fmt.Errorf("%w: invalid category", ErrValidation)
	}
	st, err := s.newStore(ctx)
	if err != nil {
		return Tour{}, err
	}
	return st.Create(ctx, in)
}

func (s *Service) Get(ctx context.Context, id uuid.UUID) (Tour, error) {
	st, err := s.newStore(ctx)
	if err != nil {
		return Tour{}, err
	}
	return st.Get(ctx, id)
}

func (s *Service) List(ctx context.Context, f Filter) ([]Tour, error) {
	st, err := s.newStore(ctx)
	if err != nil {
		return nil, err
	}
	return st.List(ctx, f)
}

func (s *Service) Update(ctx context.Context, id uuid.UUID, in UpdateInput) (Tour, error) {
	if in.Name != nil && strings.TrimSpace(*in.Name) == "" {
		return Tour{}, fmt.Errorf("%w: name must not be empty", ErrValidation)
	}
	if in.Category != nil && !ValidCategory(*in.Category) {
		return Tour{}, fmt.Errorf("%w: invalid category", ErrValidation)
	}
	st, err := s.newStore(ctx)
	if err != nil {
		return Tour{}, err
	}
	return st.Update(ctx, id, in)
}

func (s *Service) Delete(ctx context.Context, id uuid.UUID) error {
	st, err := s.newStore(ctx)
	if err != nil {
		return err
	}
	return st.Delete(ctx, id)
}
