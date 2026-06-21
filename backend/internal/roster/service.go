package roster

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/google/uuid"

	appmw "github.com/davidnguyen2205/opero/backend/internal/platform/middleware"
)

const statusDraft = "draft"

// repo is the tenant-DB persistence the service needs (satisfied by *Store).
// Declared as an interface so the service is unit-testable with fakes.
type repo interface {
	CreateLocation(ctx context.Context, in CreateLocationInput) (Location, error)
	GetLocation(ctx context.Context, id uuid.UUID) (Location, error)
	ListLocations(ctx context.Context) ([]Location, error)
	UpdateLocation(ctx context.Context, id uuid.UUID, in UpdateLocationInput) (Location, error)
	DeleteLocation(ctx context.Context, id uuid.UUID) error
	CreateShift(ctx context.Context, in CreateShiftInput, status string) (Shift, error)
	GetShift(ctx context.Context, id uuid.UUID) (Shift, error)
	ListShifts(ctx context.Context, f ShiftFilter) ([]Shift, error)
	UpdateShift(ctx context.Context, id uuid.UUID, in UpdateShiftInput) (Shift, error)
	PublishShift(ctx context.Context, id uuid.UUID) (Shift, error)
	DeleteShift(ctx context.Context, id uuid.UUID) error
}

// EmployeeResolver maps a control-plane user id to its tenant employee id.
// Satisfied by identity.Service. The bool is false when the user has no linked
// employee (kept as a bool rather than a cross-package sentinel so this module
// needn't import identity). Used by ListMyShifts to scope to the caller.
type EmployeeResolver interface {
	EmployeeIDByUserID(ctx context.Context, userID uuid.UUID) (uuid.UUID, bool, error)
}

// Service holds roster business logic. The tenant store is resolved per request
// from the context pool (placed by TenantMiddleware) via newStore, honouring the
// rule that services use only the request-scoped tenant handle.
type Service struct {
	newStore  func(ctx context.Context) (repo, error)
	employees EmployeeResolver
	logger    *slog.Logger
}

func NewService(logger *slog.Logger, employees EmployeeResolver) *Service {
	s := &Service{logger: logger, employees: employees}
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

// --- locations ---

func (s *Service) CreateLocation(ctx context.Context, in CreateLocationInput) (Location, error) {
	if strings.TrimSpace(in.Name) == "" {
		return Location{}, fmt.Errorf("%w: name is required", ErrValidation)
	}
	st, err := s.newStore(ctx)
	if err != nil {
		return Location{}, err
	}
	return st.CreateLocation(ctx, in)
}

func (s *Service) GetLocation(ctx context.Context, id uuid.UUID) (Location, error) {
	st, err := s.newStore(ctx)
	if err != nil {
		return Location{}, err
	}
	return st.GetLocation(ctx, id)
}

func (s *Service) ListLocations(ctx context.Context) ([]Location, error) {
	st, err := s.newStore(ctx)
	if err != nil {
		return nil, err
	}
	return st.ListLocations(ctx)
}

func (s *Service) UpdateLocation(ctx context.Context, id uuid.UUID, in UpdateLocationInput) (Location, error) {
	if in.Name != nil && strings.TrimSpace(*in.Name) == "" {
		return Location{}, fmt.Errorf("%w: name must not be empty", ErrValidation)
	}
	st, err := s.newStore(ctx)
	if err != nil {
		return Location{}, err
	}
	return st.UpdateLocation(ctx, id, in)
}

func (s *Service) DeleteLocation(ctx context.Context, id uuid.UUID) error {
	st, err := s.newStore(ctx)
	if err != nil {
		return err
	}
	return st.DeleteLocation(ctx, id)
}

// --- shifts ---

func (s *Service) CreateShift(ctx context.Context, in CreateShiftInput) (Shift, error) {
	if in.EmployeeID == uuid.Nil {
		return Shift{}, fmt.Errorf("%w: employee_id is required", ErrValidation)
	}
	if in.StartsAt.IsZero() {
		return Shift{}, fmt.Errorf("%w: starts_at is required", ErrValidation)
	}
	if in.EndsAt.IsZero() {
		return Shift{}, fmt.Errorf("%w: ends_at is required", ErrValidation)
	}
	if !in.EndsAt.After(in.StartsAt) {
		return Shift{}, fmt.Errorf("%w: ends_at must be after starts_at", ErrValidation)
	}
	st, err := s.newStore(ctx)
	if err != nil {
		return Shift{}, err
	}
	return st.CreateShift(ctx, in, statusDraft)
}

func (s *Service) GetShift(ctx context.Context, id uuid.UUID) (Shift, error) {
	st, err := s.newStore(ctx)
	if err != nil {
		return Shift{}, err
	}
	return st.GetShift(ctx, id)
}

func (s *Service) ListShifts(ctx context.Context, f ShiftFilter) ([]Shift, error) {
	st, err := s.newStore(ctx)
	if err != nil {
		return nil, err
	}
	return st.ListShifts(ctx, f)
}

// ListMyShifts lists shifts for the employee linked to the given user. Returns
// an empty slice (no error) if the user has no linked employee record.
func (s *Service) ListMyShifts(ctx context.Context, userID uuid.UUID, f ShiftFilter) ([]Shift, error) {
	empID, found, err := s.employees.EmployeeIDByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if !found {
		return []Shift{}, nil
	}
	f.EmployeeID = &empID
	st, err := s.newStore(ctx)
	if err != nil {
		return nil, err
	}
	return st.ListShifts(ctx, f)
}

func (s *Service) UpdateShift(ctx context.Context, id uuid.UUID, in UpdateShiftInput) (Shift, error) {
	// When both bounds are supplied we can validate ordering up front for a
	// clean 400. Partial updates that leave one bound implicit are still guarded
	// by the DB CHECK (shifts_time_order), mapped to ErrValidation in the store.
	if in.StartsAt != nil && in.EndsAt != nil && !in.EndsAt.After(*in.StartsAt) {
		return Shift{}, fmt.Errorf("%w: ends_at must be after starts_at", ErrValidation)
	}
	st, err := s.newStore(ctx)
	if err != nil {
		return Shift{}, err
	}
	return st.UpdateShift(ctx, id, in)
}

func (s *Service) PublishShift(ctx context.Context, id uuid.UUID) (Shift, error) {
	st, err := s.newStore(ctx)
	if err != nil {
		return Shift{}, err
	}
	return st.PublishShift(ctx, id)
}

func (s *Service) DeleteShift(ctx context.Context, id uuid.UUID) error {
	st, err := s.newStore(ctx)
	if err != nil {
		return err
	}
	return st.DeleteShift(ctx, id)
}
