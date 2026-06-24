package leave

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/google/uuid"

	appmw "github.com/davidnguyen2205/opero/backend/internal/platform/middleware"
)

// repo is the tenant-DB persistence the service needs (satisfied by *Store).
// An interface so the service is unit-testable with fakes.
type repo interface {
	Create(ctx context.Context, employeeID uuid.UUID, in CreateInput) (Request, error)
	Get(ctx context.Context, id uuid.UUID) (Request, error)
	List(ctx context.Context, f Filter) ([]Request, error)
	SetStatus(ctx context.Context, id uuid.UUID, status string, reviewedBy *uuid.UUID) (Request, error)
	SumApprovedDays(ctx context.Context, employeeID uuid.UUID, yearStart, yearEnd time.Time) (int, error)
	Entitlement(ctx context.Context, employeeID uuid.UUID, year int) (int, bool, error)
}

// EmployeeResolver maps a control-plane user id to its tenant employee id (false
// when the user has no linked employee). Satisfied by identity.Service; kept as
// an interface so this module needn't import identity.
type EmployeeResolver interface {
	EmployeeIDByUserID(ctx context.Context, userID uuid.UUID) (uuid.UUID, bool, error)
}

// Service holds leave business logic. The tenant store is resolved per request
// from the context pool (placed by TenantMiddleware) via newStore.
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

var validTypes = map[string]bool{"holiday": true, "sick": true, "personal": true}

func (s *Service) validateCreate(in CreateInput) error {
	if !validTypes[in.Type] {
		return fmt.Errorf("%w: type must be holiday, sick or personal", ErrValidation)
	}
	if in.StartDate.IsZero() || in.EndDate.IsZero() {
		return fmt.Errorf("%w: start_date and end_date are required", ErrValidation)
	}
	if in.EndDate.Before(in.StartDate) {
		return fmt.Errorf("%w: end_date must be on or after start_date", ErrValidation)
	}
	if in.Note != nil && strings.TrimSpace(*in.Note) == "" {
		in.Note = nil
	}
	return nil
}

// CreateMyLeave creates a pending request for the employee linked to userID.
func (s *Service) CreateMyLeave(ctx context.Context, userID uuid.UUID, in CreateInput) (Request, error) {
	if err := s.validateCreate(in); err != nil {
		return Request{}, err
	}
	empID, found, err := s.employees.EmployeeIDByUserID(ctx, userID)
	if err != nil {
		return Request{}, err
	}
	if !found {
		return Request{}, fmt.Errorf("%w: no employee is linked to this account", ErrValidation)
	}
	st, err := s.newStore(ctx)
	if err != nil {
		return Request{}, err
	}
	return st.Create(ctx, empID, in)
}

// ListMyLeave returns the caller's own requests (empty if no linked employee).
func (s *Service) ListMyLeave(ctx context.Context, userID uuid.UUID) ([]Request, error) {
	empID, found, err := s.employees.EmployeeIDByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if !found {
		return []Request{}, nil
	}
	st, err := s.newStore(ctx)
	if err != nil {
		return nil, err
	}
	return st.List(ctx, Filter{EmployeeID: &empID})
}

// MyBalance computes the caller's balance for the current (UTC) year. Returns a
// zeroed balance (no error) if the caller has no linked employee.
func (s *Service) MyBalance(ctx context.Context, userID uuid.UUID, now time.Time) (Balance, error) {
	year := now.UTC().Year()
	empID, found, err := s.employees.EmployeeIDByUserID(ctx, userID)
	if err != nil {
		return Balance{}, err
	}
	if !found {
		return Balance{Year: year, EntitledDays: DefaultEntitledDays, RemainingDays: DefaultEntitledDays}, nil
	}
	st, err := s.newStore(ctx)
	if err != nil {
		return Balance{}, err
	}
	entitled, ok, err := st.Entitlement(ctx, empID, year)
	if err != nil {
		return Balance{}, err
	}
	if !ok {
		entitled = DefaultEntitledDays
	}
	yearStart := time.Date(year, time.January, 1, 0, 0, 0, 0, time.UTC)
	yearEnd := time.Date(year, time.December, 31, 0, 0, 0, 0, time.UTC)
	used, err := st.SumApprovedDays(ctx, empID, yearStart, yearEnd)
	if err != nil {
		return Balance{}, err
	}
	return Balance{
		Year:          year,
		EntitledDays:  entitled,
		UsedDays:      used,
		RemainingDays: entitled - used,
	}, nil
}

// List returns all requests for the tenant (manager view).
func (s *Service) List(ctx context.Context, f Filter) ([]Request, error) {
	st, err := s.newStore(ctx)
	if err != nil {
		return nil, err
	}
	return st.List(ctx, f)
}

// review sets a request's status, recording the reviewing user. Idempotent when
// the request is already in the target status; rejects transitions away from a
// already-decided request to avoid silently overturning a decision.
func (s *Service) review(ctx context.Context, id uuid.UUID, target string, reviewerUserID uuid.UUID) (Request, error) {
	st, err := s.newStore(ctx)
	if err != nil {
		return Request{}, err
	}
	cur, err := st.Get(ctx, id)
	if err != nil {
		return Request{}, err
	}
	if cur.Status == target {
		return cur, nil // idempotent
	}
	if cur.Status != StatusPending {
		return Request{}, fmt.Errorf("%w: request already %s", ErrValidation, cur.Status)
	}
	rb := reviewerUserID
	return st.SetStatus(ctx, id, target, &rb)
}

func (s *Service) Approve(ctx context.Context, id, reviewerUserID uuid.UUID) (Request, error) {
	return s.review(ctx, id, StatusApproved, reviewerUserID)
}

func (s *Service) Reject(ctx context.Context, id, reviewerUserID uuid.UUID) (Request, error) {
	return s.review(ctx, id, StatusRejected, reviewerUserID)
}
