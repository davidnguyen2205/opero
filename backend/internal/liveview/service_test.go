package liveview

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/davidnguyen2205/opero/backend/gen/oapi"
	"github.com/davidnguyen2205/opero/backend/internal/attendance"
	"github.com/davidnguyen2205/opero/backend/internal/identity"
	"github.com/davidnguyen2205/opero/backend/internal/roster"
)

type fakeShifts struct {
	out     []roster.Shift
	gotFrom *time.Time
	gotTo   *time.Time
	gotStat *string
}

func (f *fakeShifts) ListShifts(_ context.Context, flt roster.ShiftFilter) ([]roster.Shift, error) {
	f.gotFrom, f.gotTo, f.gotStat = flt.From, flt.To, flt.Status
	return f.out, nil
}

type fakeAttend struct {
	out    []attendance.Record
	gotIDs []uuid.UUID
}

func (f *fakeAttend) ListByShiftIDs(_ context.Context, shiftIDs []uuid.UUID) ([]attendance.Record, error) {
	f.gotIDs = shiftIDs
	want := make(map[uuid.UUID]bool, len(shiftIDs))
	for _, id := range shiftIDs {
		want[id] = true
	}
	out := make([]attendance.Record, 0)
	for _, r := range f.out {
		if r.ShiftID != nil && want[*r.ShiftID] {
			out = append(out, r)
		}
	}
	return out, nil
}

type fakeEmps struct{ out []identity.Employee }

func (f *fakeEmps) ListEmployees(_ context.Context, _ identity.EmployeeFilter) ([]identity.Employee, error) {
	return f.out, nil
}

func TestLiveViewJoinAndStatusDerivation(t *testing.T) {
	emp1, emp2, emp3 := uuid.New(), uuid.New(), uuid.New()
	shift1, shift2, shift3 := uuid.New(), uuid.New(), uuid.New()
	now := time.Now().UTC()
	from, to := now.Add(-time.Minute), now.Add(time.Hour)
	// emp1 checked in BEFORE the window's `from` — the case the by-shift_id
	// join must still catch (the old check_in_at window would have dropped it).
	earlyCheckIn := from.Add(-2 * time.Hour)
	checkOut := now

	shifts := &fakeShifts{out: []roster.Shift{
		{ID: shift1, EmployeeID: emp1, StartsAt: now, EndsAt: now.Add(8 * time.Hour), Status: "published"},
		{ID: shift2, EmployeeID: emp2, StartsAt: now, EndsAt: now.Add(8 * time.Hour), Status: "published"},
		{ID: shift3, EmployeeID: emp3, StartsAt: now, EndsAt: now.Add(8 * time.Hour), Status: "published"},
	}}
	attend := &fakeAttend{out: []attendance.Record{
		// emp1: checked in early (before `from`) against shift1
		{ID: uuid.New(), EmployeeID: emp1, ShiftID: &shift1, CheckInAt: &earlyCheckIn, Status: "checked_in"},
		// emp3: checked in then out against shift3
		{ID: uuid.New(), EmployeeID: emp3, ShiftID: &shift3, CheckInAt: &earlyCheckIn, CheckOutAt: &checkOut, Status: "checked_out"},
		// a stray record with no shift link must be ignored by the join
		{ID: uuid.New(), EmployeeID: emp2, ShiftID: nil, CheckInAt: &earlyCheckIn, Status: "checked_in"},
	}}
	emps := &fakeEmps{out: []identity.Employee{
		{ID: emp1, FullName: "Ada Guide"},
		{ID: emp2, FullName: "Bo Driver"},
		{ID: emp3, FullName: "Cy Ops"},
	}}

	svc := NewService(shifts, attend, emps)
	entries, err := svc.LiveView(context.Background(), from, to)
	if err != nil {
		t.Fatalf("LiveView: %v", err)
	}
	if len(entries) != 3 {
		t.Fatalf("got %d entries, want 3", len(entries))
	}

	// shift order preserved from the shift lister
	if entries[0].Shift.ID != shift1 || entries[1].Shift.ID != shift2 || entries[2].Shift.ID != shift3 {
		t.Errorf("entry order not preserved")
	}
	// emp1: checked in (even though check-in preceded `from`), name + time populated
	if entries[0].AttendanceStatus != "checked_in" {
		t.Errorf("entry0 status = %q, want checked_in", entries[0].AttendanceStatus)
	}
	if entries[0].EmployeeName != "Ada Guide" {
		t.Errorf("entry0 name = %q", entries[0].EmployeeName)
	}
	if entries[0].CheckInAt == nil || !entries[0].CheckInAt.Equal(earlyCheckIn) {
		t.Errorf("entry0 check_in_at not populated: %v", entries[0].CheckInAt)
	}
	// emp2: no record linked to shift2 -> not_checked_in (stray no-shift record ignored)
	if entries[1].AttendanceStatus != "not_checked_in" {
		t.Errorf("entry1 status = %q, want not_checked_in", entries[1].AttendanceStatus)
	}
	if entries[1].CheckInAt != nil {
		t.Errorf("entry1 should have no check-in time")
	}
	// emp3: checked out, both timestamps propagated
	if entries[2].AttendanceStatus != "checked_out" {
		t.Errorf("entry2 status = %q, want checked_out", entries[2].AttendanceStatus)
	}
	if entries[2].CheckOutAt == nil || !entries[2].CheckOutAt.Equal(checkOut) {
		t.Errorf("entry2 check_out_at not populated: %v", entries[2].CheckOutAt)
	}

	// the published filter + window were passed through to the shift lister
	if shifts.gotStat == nil || *shifts.gotStat != "published" {
		t.Errorf("shift filter status = %v, want published", shifts.gotStat)
	}
	if shifts.gotFrom == nil || !shifts.gotFrom.Equal(from) || shifts.gotTo == nil || !shifts.gotTo.Equal(to) {
		t.Errorf("window not passed through: from=%v to=%v", shifts.gotFrom, shifts.gotTo)
	}
	// attendance was queried by the shift ids, not a time window
	if len(attend.gotIDs) != 3 {
		t.Errorf("attendance queried with %d shift ids, want 3", len(attend.gotIDs))
	}
}

func TestResolveWindowDefaultsToUTCDay(t *testing.T) {
	now := time.Date(2026, 6, 21, 15, 30, 0, 0, time.UTC)
	from, to := resolveWindow(oapi.GetLiveViewParams{}, now)
	wantFrom := time.Date(2026, 6, 21, 0, 0, 0, 0, time.UTC)
	if !from.Equal(wantFrom) {
		t.Errorf("from = %v, want start-of-UTC-day %v", from, wantFrom)
	}
	if !to.Equal(wantFrom.Add(24 * time.Hour)) {
		t.Errorf("to = %v, want +24h", to)
	}
}

func TestResolveWindowHonoursExplicitParams(t *testing.T) {
	f := time.Date(2026, 6, 20, 8, 0, 0, 0, time.UTC)
	tt := time.Date(2026, 6, 20, 20, 0, 0, 0, time.UTC)
	from, to := resolveWindow(oapi.GetLiveViewParams{From: &f, To: &tt}, time.Now())
	if !from.Equal(f) || !to.Equal(tt) {
		t.Errorf("explicit params not honoured: from=%v to=%v", from, to)
	}
}
