// Package liveview implements the manager "who's working now" view. It owns no
// tenant tables: it composes the roster, attendance, and identity modules
// through their exported Go interfaces (the sanctioned cross-module mechanism)
// and joins the results in memory. Each dependency resolves the same
// request-context tenant pool, so tenant isolation is preserved.
package liveview

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/davidnguyen2205/opero/backend/internal/attendance"
	"github.com/davidnguyen2205/opero/backend/internal/identity"
	"github.com/davidnguyen2205/opero/backend/internal/roster"
)

// Dependencies, expressed as narrow interfaces over the other modules' exported
// services (satisfied by *roster.Service, *attendance.Service, *identity.Service).
type shiftLister interface {
	ListShifts(ctx context.Context, f roster.ShiftFilter) ([]roster.Shift, error)
}
type attendanceLister interface {
	ListByShiftIDs(ctx context.Context, shiftIDs []uuid.UUID) ([]attendance.Record, error)
}
type employeeLister interface {
	ListEmployees(ctx context.Context, f identity.EmployeeFilter) ([]identity.Employee, error)
}

// Entry is one published shift joined with the employee and their current
// attendance state.
type Entry struct {
	EmployeeID       uuid.UUID
	EmployeeName     string
	Shift            roster.Shift
	AttendanceStatus string // not_checked_in | checked_in | on_break | checked_out
	CheckInAt        *time.Time
	CheckOutAt       *time.Time
	CheckInLat       *float64
	CheckInLng       *float64
	BreakStartedAt   *time.Time
}

type Service struct {
	shifts    shiftLister
	attend    attendanceLister
	employees employeeLister
}

func NewService(shifts shiftLister, attend attendanceLister, employees employeeLister) *Service {
	return &Service{shifts: shifts, attend: attend, employees: employees}
}

// LiveView returns published shifts whose start is in [from, to), each joined
// with the employee name and current attendance state. Shifts come back ordered
// by start time (the roster store orders them), which this preserves.
func (s *Service) LiveView(ctx context.Context, from, to time.Time) ([]Entry, error) {
	published := "published"
	shifts, err := s.shifts.ListShifts(ctx, roster.ShiftFilter{Status: &published, From: &from, To: &to})
	if err != nil {
		return nil, err
	}

	// Index attendance by the shift it is linked to. Matched by shift_id (not a
	// check_in_at window) so an early-arrival or overnight check-in that
	// preceded `from` is still joined to its shift.
	shiftIDs := make([]uuid.UUID, 0, len(shifts))
	for _, sh := range shifts {
		shiftIDs = append(shiftIDs, sh.ID)
	}
	records, err := s.attend.ListByShiftIDs(ctx, shiftIDs)
	if err != nil {
		return nil, err
	}
	byShift := make(map[uuid.UUID]attendance.Record, len(records))
	for _, r := range records {
		if r.ShiftID != nil {
			byShift[*r.ShiftID] = r
		}
	}

	// Employee name lookup.
	emps, err := s.employees.ListEmployees(ctx, identity.EmployeeFilter{})
	if err != nil {
		return nil, err
	}
	nameByID := make(map[uuid.UUID]string, len(emps))
	for _, e := range emps {
		nameByID[e.ID] = e.FullName
	}

	entries := make([]Entry, 0, len(shifts))
	for _, sh := range shifts {
		e := Entry{
			EmployeeID:       sh.EmployeeID,
			EmployeeName:     nameByID[sh.EmployeeID],
			Shift:            sh,
			AttendanceStatus: "not_checked_in",
		}
		if rec, ok := byShift[sh.ID]; ok {
			e.AttendanceStatus = rec.Status
			e.CheckInAt = rec.CheckInAt
			e.CheckOutAt = rec.CheckOutAt
			e.CheckInLat = rec.CheckInLat
			e.CheckInLng = rec.CheckInLng
			e.BreakStartedAt = rec.BreakStartedAt
		}
		entries = append(entries, e)
	}
	return entries, nil
}
