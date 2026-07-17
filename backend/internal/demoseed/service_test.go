package demoseed

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/davidnguyen2205/opero/backend/internal/attendance"
	"github.com/davidnguyen2205/opero/backend/internal/identity"
	"github.com/davidnguyen2205/opero/backend/internal/roster"
	"github.com/davidnguyen2205/opero/backend/internal/tours"
)

type fakeEmployees struct{ emps []identity.Employee }

func (f *fakeEmployees) ListEmployees(_ context.Context, _ identity.EmployeeFilter) ([]identity.Employee, error) {
	return f.emps, nil
}

type fakeRoster struct {
	locs    []roster.Location
	oldIDs  []uuid.UUID
	deleted bool
	created []roster.Shift
}

func (f *fakeRoster) ListLocations(context.Context) ([]roster.Location, error) { return f.locs, nil }
func (f *fakeRoster) ShiftIDsByNote(_ context.Context, note string) ([]uuid.UUID, error) {
	if note != SeedNote {
		return nil, nil
	}
	return f.oldIDs, nil
}
func (f *fakeRoster) DeleteShiftsByNote(_ context.Context, _ string) (int64, error) {
	f.deleted = true
	return int64(len(f.oldIDs)), nil
}
func (f *fakeRoster) CreatePublishedShift(_ context.Context, in roster.CreateShiftInput) (roster.Shift, error) {
	s := roster.Shift{
		ID: uuid.New(), EmployeeID: in.EmployeeID, LocationID: in.LocationID,
		TourID: in.TourID, StartsAt: in.StartsAt, EndsAt: in.EndsAt,
		Notes: in.Notes, Status: "published",
	}
	f.created = append(f.created, s)
	return s, nil
}

type fakeAttendance struct {
	deletedShiftIDs []uuid.UUID
	records         []attendance.DemoRecordInput
}

func (f *fakeAttendance) DeleteByShiftIDs(_ context.Context, ids []uuid.UUID) (int64, error) {
	f.deletedShiftIDs = ids
	return int64(len(ids)), nil
}
func (f *fakeAttendance) SeedDemoRecord(_ context.Context, in attendance.DemoRecordInput) (attendance.Record, error) {
	f.records = append(f.records, in)
	return attendance.Record{ID: uuid.New(), Status: in.Status}, nil
}

type fakeTours struct{ trs []tours.Tour }

func (f *fakeTours) List(context.Context, tours.Filter) ([]tours.Tour, error) { return f.trs, nil }

func TestSeedReplacesAndCoversStates(t *testing.T) {
	emps := make([]identity.Employee, 9)
	for i := range emps {
		emps[i] = identity.Employee{ID: uuid.New(), Status: "active"}
	}
	lat, lng := 38.7, -9.1
	fr := &fakeRoster{
		locs:   []roster.Location{{ID: uuid.New(), Lat: &lat, Lng: &lng}},
		oldIDs: []uuid.UUID{uuid.New(), uuid.New()},
	}
	fa := &fakeAttendance{}
	svc := NewService(&fakeEmployees{emps: emps}, fr, fa, &fakeTours{trs: []tours.Tour{{ID: uuid.New()}, {ID: uuid.New()}}})
	now := time.Date(2026, 7, 16, 10, 30, 0, 0, time.UTC)
	svc.now = func() time.Time { return now }

	res, err := svc.Seed(context.Background())
	if err != nil {
		t.Fatalf("seed: %v", err)
	}

	// old data replaced: attendance deleted for the old shift ids, then shifts
	if len(fa.deletedShiftIDs) != 2 || !fr.deleted {
		t.Errorf("previous seed not replaced: attendance ids=%d shifts deleted=%v", len(fa.deletedShiftIDs), fr.deleted)
	}
	if res.Shifts != 9 || len(fr.created) != 9 {
		t.Fatalf("shifts created = %d (result %d), want 9", len(fr.created), res.Shifts)
	}
	if res.AttendanceRecords != 5 || len(fa.records) != 5 {
		t.Fatalf("attendance created = %d (result %d), want 5", len(fa.records), res.AttendanceRecords)
	}

	// state mix: checked_out with checkout time, on_break with break start,
	// three checked_in
	byStatus := map[string]int{}
	for _, r := range fa.records {
		byStatus[r.Status]++
	}
	if byStatus["checked_out"] != 1 || byStatus["on_break"] != 1 || byStatus["checked_in"] != 3 {
		t.Errorf("status mix = %v, want 1 checked_out / 1 on_break / 3 checked_in", byStatus)
	}
	if fa.records[0].CheckOutAt == nil {
		t.Error("checked_out record has no check_out_at")
	}
	if fa.records[1].BreakStartedAt == nil {
		t.Error("on_break record has no break_started_at")
	}

	// windows: first third started, last third upcoming; tours on first 6
	if !fr.created[0].StartsAt.Before(now) || !fr.created[8].StartsAt.After(now) {
		t.Errorf("window spread wrong: first starts %v, last starts %v (now %v)",
			fr.created[0].StartsAt, fr.created[8].StartsAt, now)
	}
	for i, s := range fr.created {
		if (i < 6) != (s.TourID != nil) {
			t.Errorf("shift %d tour assignment wrong (tour=%v)", i, s.TourID)
		}
		if s.Notes == nil || *s.Notes != SeedNote {
			t.Errorf("shift %d not tagged with seed note", i)
		}
	}
}

func TestSeedFailsWithoutEmployees(t *testing.T) {
	svc := NewService(&fakeEmployees{}, &fakeRoster{}, &fakeAttendance{}, &fakeTours{})
	if _, err := svc.Seed(context.Background()); err == nil {
		t.Fatal("expected error with no active employees")
	}
}
