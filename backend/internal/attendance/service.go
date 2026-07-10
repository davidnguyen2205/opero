package attendance

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/google/uuid"

	appmw "github.com/davidnguyen2205/opero/backend/internal/platform/middleware"
)

type repo interface {
	GetByClientID(ctx context.Context, clientID uuid.UUID) (Record, error)
	CreateCheckIn(ctx context.Context, employeeID uuid.UUID, in CheckInInput) (Record, error)
	CheckOut(ctx context.Context, in CheckOutInput) (Record, error)
	SetStatus(ctx context.Context, clientID uuid.UUID, status string) (Record, error)
	List(ctx context.Context, f AttendanceFilter) ([]Record, error)
	ListByShiftIDs(ctx context.Context, shiftIDs []uuid.UUID) ([]Record, error)
}

// EmployeeResolver maps a control-plane user id to its tenant employee id.
// Satisfied by identity.Service. The bool is false when the user has no linked
// employee (kept as a bool rather than a cross-package sentinel error so this
// module needn't import identity).
type EmployeeResolver interface {
	EmployeeIDByUserID(ctx context.Context, userID uuid.UUID) (uuid.UUID, bool, error)
}

// ShiftResolver reports who a shift belongs to and its status. Satisfied by
// roster.Service. Not-found is a bool rather than a cross-package sentinel
// error so this module needn't import roster.
type ShiftResolver interface {
	ShiftOwnerByID(ctx context.Context, id uuid.UUID) (employeeID uuid.UUID, status string, found bool, err error)
}

type Service struct {
	newStore  func(ctx context.Context) (repo, error)
	employees EmployeeResolver
	shifts    ShiftResolver
	logger    *slog.Logger
}

func NewService(logger *slog.Logger, employees EmployeeResolver, shifts ShiftResolver) *Service {
	s := &Service{logger: logger, employees: employees, shifts: shifts}
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

// resolveEmployee maps the authenticated user to their tenant employee id.
func (s *Service) resolveEmployee(ctx context.Context, userID uuid.UUID) (uuid.UUID, error) {
	empID, found, err := s.employees.EmployeeIDByUserID(ctx, userID)
	if err != nil {
		return uuid.Nil, err
	}
	if !found {
		return uuid.Nil, fmt.Errorf("%w: no employee is linked to this account", ErrValidation)
	}
	return empID, nil
}

// validateOwnShift verifies the shift exists, belongs to empID, and is
// published. All three failures return one indistinguishable validation error
// so check-in can't be used as an existence oracle for other employees' shifts.
func (s *Service) validateOwnShift(ctx context.Context, empID, shiftID uuid.UUID) error {
	ownerID, status, found, err := s.shifts.ShiftOwnerByID(ctx, shiftID)
	if err != nil {
		return fmt.Errorf("resolving shift for check-in: %w", err)
	}
	if !found || ownerID != empID || status != "published" {
		return fmt.Errorf("%w: shift_id does not reference one of your published shifts", ErrValidation)
	}
	return nil
}

// CheckIn records (idempotently) a check-in for the authenticated user's
// employee. Returns created=true only when a new record was inserted.
func (s *Service) CheckIn(ctx context.Context, userID uuid.UUID, in CheckInInput) (Record, bool, error) {
	if in.ClientID == uuid.Nil {
		return Record{}, false, fmt.Errorf("%w: client_id is required", ErrValidation)
	}
	empID, err := s.resolveEmployee(ctx, userID)
	if err != nil {
		return Record{}, false, err
	}
	st, err := s.newStore(ctx)
	if err != nil {
		return Record{}, false, err
	}

	existing, err := st.GetByClientID(ctx, in.ClientID)
	switch {
	case err == nil:
		if existing.EmployeeID != empID {
			return Record{}, false, fmt.Errorf("%w: client_id already used by another employee", ErrConflict)
		}
		return existing, false, nil // idempotent replay
	case !errors.Is(err, ErrNotFound):
		return Record{}, false, err
	}

	// Only validate before the insert; the idempotent-replay path above returns
	// a record that was already ownership-checked at creation time.
	if in.ShiftID != nil {
		if err := s.validateOwnShift(ctx, empID, *in.ShiftID); err != nil {
			return Record{}, false, err
		}
	}

	rec, err := st.CreateCheckIn(ctx, empID, in)
	if err != nil {
		// Lost a race with a concurrent identical check-in: fetch and return it.
		if errors.Is(err, ErrConflict) {
			if existing, getErr := st.GetByClientID(ctx, in.ClientID); getErr == nil && existing.EmployeeID == empID {
				return existing, false, nil
			}
		}
		return Record{}, false, err
	}
	return rec, true, nil
}

// CheckOut records check-out against the record identified by client_id. The
// record must belong to the authenticated user's employee. Idempotent.
func (s *Service) CheckOut(ctx context.Context, userID uuid.UUID, in CheckOutInput) (Record, error) {
	if in.ClientID == uuid.Nil {
		return Record{}, fmt.Errorf("%w: client_id is required", ErrValidation)
	}
	empID, err := s.resolveEmployee(ctx, userID)
	if err != nil {
		return Record{}, err
	}
	st, err := s.newStore(ctx)
	if err != nil {
		return Record{}, err
	}

	rec, err := st.GetByClientID(ctx, in.ClientID)
	if err != nil {
		return Record{}, err // ErrNotFound
	}
	if rec.EmployeeID != empID {
		// Don't reveal another employee's record.
		return Record{}, ErrNotFound
	}
	if rec.Status == "checked_out" {
		return rec, nil // idempotent
	}
	return st.CheckOut(ctx, in)
}

// SetBreak toggles the break state of the authenticated user's open attendance
// record (checked_in ⇄ on_break). Idempotent on the target state.
func (s *Service) SetBreak(ctx context.Context, userID, clientID uuid.UUID, onBreak bool) (Record, error) {
	if clientID == uuid.Nil {
		return Record{}, fmt.Errorf("%w: client_id is required", ErrValidation)
	}
	empID, err := s.resolveEmployee(ctx, userID)
	if err != nil {
		return Record{}, err
	}
	st, err := s.newStore(ctx)
	if err != nil {
		return Record{}, err
	}
	rec, err := st.GetByClientID(ctx, clientID)
	if err != nil {
		return Record{}, err // ErrNotFound
	}
	if rec.EmployeeID != empID {
		return Record{}, ErrNotFound
	}
	target := "checked_in"
	if onBreak {
		target = "on_break"
	}
	if rec.Status == target {
		return rec, nil // idempotent
	}
	if rec.Status != "checked_in" && rec.Status != "on_break" {
		return Record{}, fmt.Errorf("%w: shift is not open", ErrValidation)
	}
	return st.SetStatus(ctx, clientID, target)
}

func (s *Service) List(ctx context.Context, f AttendanceFilter) ([]Record, error) {
	st, err := s.newStore(ctx)
	if err != nil {
		return nil, err
	}
	return st.List(ctx, f)
}

// ListByShiftIDs returns attendance records linked to any of the given shifts,
// independent of check-in time (for the live view join).
func (s *Service) ListByShiftIDs(ctx context.Context, shiftIDs []uuid.UUID) ([]Record, error) {
	st, err := s.newStore(ctx)
	if err != nil {
		return nil, err
	}
	return st.ListByShiftIDs(ctx, shiftIDs)
}
