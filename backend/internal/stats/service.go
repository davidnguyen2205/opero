// Package stats implements the field-app "my activity" aggregate (GET /me/stats).
// It owns no tenant tables: like liveview it composes the roster, attendance and
// identity modules through their exported Go interfaces and computes the numbers
// in memory. This is a read-only convenience aggregate, deliberately not
// analytics infrastructure (which CLAUDE.md defers to phase 2).
package stats

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/davidnguyen2205/opero/backend/internal/attendance"
	"github.com/davidnguyen2205/opero/backend/internal/identity"
	"github.com/davidnguyen2205/opero/backend/internal/roster"
)

// Narrow interfaces over the other modules' services (satisfied by
// *roster.Service, *attendance.Service, *identity.Service).
type myShiftLister interface {
	ListMyShifts(ctx context.Context, userID uuid.UUID, f roster.ShiftFilter) ([]roster.Shift, error)
}
type attendanceLister interface {
	List(ctx context.Context, f attendance.AttendanceFilter) ([]attendance.Record, error)
}
type employeeResolver interface {
	EmployeeIDByUserID(ctx context.Context, userID uuid.UUID) (uuid.UUID, bool, error)
	GetEmployee(ctx context.Context, id uuid.UUID) (identity.Employee, error)
}

// MyStats is the computed aggregate for one employee.
type MyStats struct {
	ShiftsThisMonth int
	HoursThisWeek   float64
	OnTimePct       int
	TenureDays      *int
}

type Service struct {
	shifts    myShiftLister
	attend    attendanceLister
	employees employeeResolver
}

func NewService(shifts myShiftLister, attend attendanceLister, employees employeeResolver) *Service {
	return &Service{shifts: shifts, attend: attend, employees: employees}
}

// MyStats computes the caller's stats relative to `now` (UTC week/month bounds).
// Returns an empty-but-valid result (on-time 100, zeros) when the caller has no
// linked employee or no data.
func (s *Service) MyStats(ctx context.Context, userID uuid.UUID, now time.Time) (MyStats, error) {
	out := MyStats{OnTimePct: 100}

	empID, found, err := s.employees.EmployeeIDByUserID(ctx, userID)
	if err != nil {
		return MyStats{}, err
	}
	if !found {
		return out, nil
	}

	now = now.UTC()
	monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
	monthEnd := monthStart.AddDate(0, 1, 0)
	weekStart := startOfWeek(now)

	// Shifts this month (the caller's own; published or not — it's their roster).
	monthShifts, err := s.shifts.ListMyShifts(ctx, userID, roster.ShiftFilter{From: &monthStart, To: &monthEnd})
	if err != nil {
		return MyStats{}, err
	}
	out.ShiftsThisMonth = len(monthShifts)

	// Attendance for the employee, used for hours-this-week and on-time rate.
	records, err := s.attend.List(ctx, attendance.AttendanceFilter{EmployeeID: &empID})
	if err != nil {
		return MyStats{}, err
	}

	// Hours this week: sum completed check-in/out pairs whose check-in is in the week.
	var weekMinutes float64
	for _, rec := range records {
		if rec.CheckInAt == nil || rec.CheckOutAt == nil {
			continue
		}
		ci := rec.CheckInAt.UTC()
		if ci.Before(weekStart) {
			continue
		}
		d := rec.CheckOutAt.Sub(*rec.CheckInAt)
		if d > 0 {
			weekMinutes += d.Minutes()
		}
	}
	out.HoursThisWeek = roundTo1(weekMinutes / 60.0)

	// On-time rate: of shifts with a linked check-in, the fraction where the
	// check-in was at or before the shift start. Index check-ins by shift id.
	checkInByShift := make(map[uuid.UUID]time.Time, len(records))
	for _, rec := range records {
		if rec.ShiftID == nil || rec.CheckInAt == nil {
			continue
		}
		// Keep the earliest check-in per shift.
		if existing, ok := checkInByShift[*rec.ShiftID]; !ok || rec.CheckInAt.Before(existing) {
			checkInByShift[*rec.ShiftID] = *rec.CheckInAt
		}
	}
	if len(checkInByShift) > 0 {
		// Need shift start times. Pull the caller's shifts over a wide window
		// (last 90 days) to resolve starts for any checked-in shift.
		lookback := now.AddDate(0, 0, -90)
		ahead := now.AddDate(0, 0, 1)
		allShifts, err := s.shifts.ListMyShifts(ctx, userID, roster.ShiftFilter{From: &lookback, To: &ahead})
		if err != nil {
			return MyStats{}, err
		}
		startByShift := make(map[uuid.UUID]time.Time, len(allShifts))
		for _, sh := range allShifts {
			startByShift[sh.ID] = sh.StartsAt
		}
		var considered, onTime int
		for shiftID, ci := range checkInByShift {
			start, ok := startByShift[shiftID]
			if !ok {
				continue // shift outside the window; skip rather than guess
			}
			considered++
			if !ci.After(start) {
				onTime++
			}
		}
		if considered > 0 {
			out.OnTimePct = int(float64(onTime) / float64(considered) * 100.0)
		}
	}

	// Tenure in days from hired_at.
	emp, err := s.employees.GetEmployee(ctx, empID)
	if err != nil {
		return MyStats{}, err
	}
	if emp.HiredAt != nil {
		days := int(now.Sub(emp.HiredAt.UTC()).Hours() / 24)
		if days < 0 {
			days = 0
		}
		out.TenureDays = &days
	}

	return out, nil
}

// startOfWeek returns 00:00 UTC on the Monday of now's week.
func startOfWeek(now time.Time) time.Time {
	d := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	// Go: Sunday=0..Saturday=6; shift so Monday is the start.
	offset := (int(d.Weekday()) + 6) % 7
	return d.AddDate(0, 0, -offset)
}

func roundTo1(v float64) float64 {
	return float64(int(v*10+0.5)) / 10.0
}
