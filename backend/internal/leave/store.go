package leave

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"

	leavedb "github.com/davidnguyen2205/opero/backend/gen/sqlc/leave"
)

// Store is the only place that touches the tenant database for this module. It
// is constructed per request from the tenant-scoped pool.
type Store struct {
	q *leavedb.Queries
}

func NewStore(db leavedb.DBTX) *Store {
	return &Store{q: leavedb.New(db)}
}

func mapErr(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrNotFound
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case "23503": // foreign_key_violation — bad employee_id reference
			return fmt.Errorf("%w: referenced record does not exist", ErrValidation)
		case "23514": // check_violation — type/status/date order
			return fmt.Errorf("%w: invalid field value", ErrValidation)
		}
	}
	return err
}

func toPgUUID(p *uuid.UUID) pgtype.UUID {
	if p == nil {
		return pgtype.UUID{}
	}
	return pgtype.UUID{Bytes: *p, Valid: true}
}

func fromPgUUID(v pgtype.UUID) *uuid.UUID {
	if !v.Valid {
		return nil
	}
	u := uuid.UUID(v.Bytes)
	return &u
}

func fromPgTimestamptz(v pgtype.Timestamptz) *time.Time {
	if !v.Valid {
		return nil
	}
	t := v.Time
	return &t
}

func requestFromDB(r leavedb.LeaveRequest) Request {
	return Request{
		ID:         r.ID,
		EmployeeID: r.EmployeeID,
		Type:       r.Type,
		StartDate:  r.StartDate,
		EndDate:    r.EndDate,
		Note:       r.Note,
		Status:     r.Status,
		ReviewedBy: fromPgUUID(r.ReviewedBy),
		ReviewedAt: fromPgTimestamptz(r.ReviewedAt),
		CreatedAt:  r.CreatedAt,
		UpdatedAt:  r.UpdatedAt,
	}
}

func (s *Store) Create(ctx context.Context, employeeID uuid.UUID, in CreateInput) (Request, error) {
	r, err := s.q.CreateLeaveRequest(ctx, leavedb.CreateLeaveRequestParams{
		EmployeeID: employeeID,
		Type:       in.Type,
		StartDate:  in.StartDate,
		EndDate:    in.EndDate,
		Note:       in.Note,
	})
	if err != nil {
		return Request{}, fmt.Errorf("create leave request: %w", mapErr(err))
	}
	return requestFromDB(r), nil
}

func (s *Store) Get(ctx context.Context, id uuid.UUID) (Request, error) {
	r, err := s.q.GetLeaveRequest(ctx, id)
	if err != nil {
		return Request{}, fmt.Errorf("get leave request: %w", mapErr(err))
	}
	return requestFromDB(r), nil
}

func (s *Store) List(ctx context.Context, f Filter) ([]Request, error) {
	rows, err := s.q.ListLeaveRequests(ctx, leavedb.ListLeaveRequestsParams{
		EmployeeID: toPgUUID(f.EmployeeID),
		Status:     f.Status,
	})
	if err != nil {
		return nil, fmt.Errorf("list leave requests: %w", mapErr(err))
	}
	out := make([]Request, 0, len(rows))
	for _, r := range rows {
		out = append(out, requestFromDB(r))
	}
	return out, nil
}

func (s *Store) SetStatus(ctx context.Context, id uuid.UUID, status string, reviewedBy *uuid.UUID) (Request, error) {
	r, err := s.q.SetLeaveStatus(ctx, leavedb.SetLeaveStatusParams{
		Status:     status,
		ReviewedBy: toPgUUID(reviewedBy),
		ID:         id,
	})
	if err != nil {
		return Request{}, fmt.Errorf("set leave status: %w", mapErr(err))
	}
	return requestFromDB(r), nil
}

// SumApprovedDays returns calendar days (inclusive) consumed by approved
// requests starting within [yearStart, yearEnd].
func (s *Store) SumApprovedDays(ctx context.Context, employeeID uuid.UUID, yearStart, yearEnd time.Time) (int, error) {
	n, err := s.q.SumApprovedLeaveDays(ctx, leavedb.SumApprovedLeaveDaysParams{
		EmployeeID: employeeID,
		YearStart:  yearStart,
		YearEnd:    yearEnd,
	})
	if err != nil {
		return 0, fmt.Errorf("sum approved leave days: %w", mapErr(err))
	}
	return int(n), nil
}

// Entitlement returns the employee's entitled days for the year and whether a
// balance row exists. When false, callers apply DefaultEntitledDays.
func (s *Store) Entitlement(ctx context.Context, employeeID uuid.UUID, year int) (int, bool, error) {
	n, err := s.q.GetLeaveEntitlement(ctx, leavedb.GetLeaveEntitlementParams{
		EmployeeID: employeeID,
		Year:       int32(year),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, false, nil
		}
		return 0, false, fmt.Errorf("get leave entitlement: %w", mapErr(err))
	}
	return int(n), true, nil
}
