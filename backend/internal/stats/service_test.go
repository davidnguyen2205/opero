package stats

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/davidnguyen2205/opero/backend/internal/attendance"
	"github.com/davidnguyen2205/opero/backend/internal/identity"
	"github.com/davidnguyen2205/opero/backend/internal/roster"
)

type fakeShifts struct {
	byWindow func(from, to time.Time) []roster.Shift
}

func (f fakeShifts) ListMyShifts(_ context.Context, _ uuid.UUID, fl roster.ShiftFilter) ([]roster.Shift, error) {
	if f.byWindow == nil || fl.From == nil || fl.To == nil {
		return nil, nil
	}
	return f.byWindow(*fl.From, *fl.To), nil
}

type fakeAttend struct {
	records []attendance.Record
}

func (f fakeAttend) List(context.Context, attendance.AttendanceFilter) ([]attendance.Record, error) {
	return f.records, nil
}

type fakeEmployees struct {
	empID   uuid.UUID
	found   bool
	hiredAt *time.Time
}

func (f fakeEmployees) EmployeeIDByUserID(context.Context, uuid.UUID) (uuid.UUID, bool, error) {
	return f.empID, f.found, nil
}
func (f fakeEmployees) GetEmployee(context.Context, uuid.UUID) (identity.Employee, error) {
	return identity.Employee{ID: f.empID, HiredAt: f.hiredAt}, nil
}

func TestMyStatsNoEmployeeReturnsDefault(t *testing.T) {
	svc := NewService(fakeShifts{}, fakeAttend{}, fakeEmployees{found: false})
	got, err := svc.MyStats(context.Background(), uuid.New(), time.Now())
	if err != nil {
		t.Fatalf("MyStats: %v", err)
	}
	if got.OnTimePct != 100 || got.ShiftsThisMonth != 0 || got.HoursThisWeek != 0 || got.TenureDays != nil {
		t.Errorf("got %+v, want zeroed/default", got)
	}
}

func TestMyStatsComputesHoursOnTimeAndTenure(t *testing.T) {
	empID := uuid.New()
	// "now": Wed 24 Jun 2026 12:00 UTC. Week starts Mon 22 Jun.
	now := time.Date(2026, 6, 24, 12, 0, 0, 0, time.UTC)
	shiftID := uuid.New()
	shiftStart := time.Date(2026, 6, 23, 9, 0, 0, 0, time.UTC) // Mon-of-week+1, this month

	ci := time.Date(2026, 6, 23, 8, 55, 0, 0, time.UTC)   // 5 min early -> on time
	co := time.Date(2026, 6, 23, 12, 55, 0, 0, time.UTC)  // 4h worked
	hired := time.Date(2025, 6, 24, 0, 0, 0, 0, time.UTC) // ~1 year

	svc := NewService(
		fakeShifts{byWindow: func(from, to time.Time) []roster.Shift {
			// Return the shift whenever the window covers shiftStart.
			if !shiftStart.Before(from) && shiftStart.Before(to) {
				return []roster.Shift{{ID: shiftID, EmployeeID: empID, StartsAt: shiftStart, EndsAt: shiftStart.Add(4 * time.Hour), Status: "published"}}
			}
			return nil
		}},
		fakeAttend{records: []attendance.Record{
			{ID: uuid.New(), EmployeeID: empID, ShiftID: &shiftID, CheckInAt: &ci, CheckOutAt: &co, Status: "checked_out"},
		}},
		fakeEmployees{empID: empID, found: true, hiredAt: &hired},
	)

	got, err := svc.MyStats(context.Background(), uuid.New(), now)
	if err != nil {
		t.Fatalf("MyStats: %v", err)
	}
	if got.ShiftsThisMonth != 1 {
		t.Errorf("shifts_this_month = %d, want 1", got.ShiftsThisMonth)
	}
	if got.HoursThisWeek != 4.0 {
		t.Errorf("hours_this_week = %v, want 4.0", got.HoursThisWeek)
	}
	if got.OnTimePct != 100 {
		t.Errorf("on_time_pct = %d, want 100", got.OnTimePct)
	}
	if got.TenureDays == nil || *got.TenureDays < 360 || *got.TenureDays > 370 {
		t.Errorf("tenure_days = %v, want ~365", got.TenureDays)
	}
}

func TestMyStatsLateCheckIn(t *testing.T) {
	empID := uuid.New()
	now := time.Date(2026, 6, 24, 12, 0, 0, 0, time.UTC)
	shiftID := uuid.New()
	shiftStart := time.Date(2026, 6, 23, 9, 0, 0, 0, time.UTC)
	ci := time.Date(2026, 6, 23, 9, 15, 0, 0, time.UTC) // 15 min late

	svc := NewService(
		fakeShifts{byWindow: func(from, to time.Time) []roster.Shift {
			if !shiftStart.Before(from) && shiftStart.Before(to) {
				return []roster.Shift{{ID: shiftID, EmployeeID: empID, StartsAt: shiftStart, EndsAt: shiftStart.Add(4 * time.Hour)}}
			}
			return nil
		}},
		fakeAttend{records: []attendance.Record{
			{ID: uuid.New(), EmployeeID: empID, ShiftID: &shiftID, CheckInAt: &ci},
		}},
		fakeEmployees{empID: empID, found: true},
	)

	got, err := svc.MyStats(context.Background(), uuid.New(), now)
	if err != nil {
		t.Fatalf("MyStats: %v", err)
	}
	if got.OnTimePct != 0 {
		t.Errorf("on_time_pct = %d, want 0 (late)", got.OnTimePct)
	}
}
