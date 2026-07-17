// Package demoseed re-seeds the demo tenant's live-view data on demand. It
// owns no tables: it composes identity, roster, attendance, and tours through
// their exported interfaces, exactly like the liveview module composes reads.
package demoseed

import (
	"context"
	"fmt"
	"math/rand/v2"
	"time"

	"github.com/google/uuid"

	"github.com/davidnguyen2205/opero/backend/internal/attendance"
	"github.com/davidnguyen2205/opero/backend/internal/identity"
	"github.com/davidnguyen2205/opero/backend/internal/roster"
	"github.com/davidnguyen2205/opero/backend/internal/tours"
)

// SeedNote tags every shift the seeder creates; re-seeding replaces only
// shifts carrying this exact note.
const SeedNote = "Seeded demo shift"

type employeeLister interface {
	ListEmployees(ctx context.Context, f identity.EmployeeFilter) ([]identity.Employee, error)
}

type rosterSeeder interface {
	ListLocations(ctx context.Context) ([]roster.Location, error)
	ShiftIDsByNote(ctx context.Context, note string) ([]uuid.UUID, error)
	DeleteShiftsByNote(ctx context.Context, note string) (int64, error)
	CreatePublishedShift(ctx context.Context, in roster.CreateShiftInput) (roster.Shift, error)
}

type attendanceSeeder interface {
	DeleteByShiftIDs(ctx context.Context, shiftIDs []uuid.UUID) (int64, error)
	SeedDemoRecord(ctx context.Context, in attendance.DemoRecordInput) (attendance.Record, error)
}

type tourLister interface {
	List(ctx context.Context, f tours.Filter) ([]tours.Tour, error)
}

type Service struct {
	employees  employeeLister
	roster     rosterSeeder
	attendance attendanceSeeder
	tours      tourLister
	now        func() time.Time // injectable for tests
}

func NewService(e employeeLister, r rosterSeeder, a attendanceSeeder, t tourLister) *Service {
	return &Service{employees: e, roster: r, attendance: a, tours: t, now: time.Now}
}

// Result summarizes what a seeding run created.
type Result struct {
	Shifts            int
	AttendanceRecords int
}

// Seed replaces the tagged demo shifts (and their attendance) with a fresh
// set anchored to the current time: for N active employees ordered by name,
// the first third started 6h ago, the second third 3h ago, the rest start in
// 2h. Attendance covers a realistic mix — one checked out (20m ago), one on
// break (started 25m ago), the remaining started rows checked in, and one
// started row deliberately left without a record (a no-show). The first six
// shifts rotate across the tenant's tours so tour grouping has content.
func (s *Service) Seed(ctx context.Context) (Result, error) {
	active := "active"
	emps, err := s.employees.ListEmployees(ctx, identity.EmployeeFilter{Status: &active})
	if err != nil {
		return Result{}, fmt.Errorf("list employees: %w", err)
	}
	if len(emps) == 0 {
		return Result{}, fmt.Errorf("demo tenant has no active employees to seed")
	}
	locs, err := s.roster.ListLocations(ctx)
	if err != nil {
		return Result{}, fmt.Errorf("list locations: %w", err)
	}
	trs, err := s.tours.List(ctx, tours.Filter{})
	if err != nil {
		return Result{}, fmt.Errorf("list tours: %w", err)
	}

	// Replace previous seeded data: attendance first (the shift FK is ON
	// DELETE SET NULL, so deleting shifts first would orphan the records).
	oldIDs, err := s.roster.ShiftIDsByNote(ctx, SeedNote)
	if err != nil {
		return Result{}, fmt.Errorf("find previous seeded shifts: %w", err)
	}
	if _, err := s.attendance.DeleteByShiftIDs(ctx, oldIDs); err != nil {
		return Result{}, fmt.Errorf("delete previous seeded attendance: %w", err)
	}
	if _, err := s.roster.DeleteShiftsByNote(ctx, SeedNote); err != nil {
		return Result{}, fmt.Errorf("delete previous seeded shifts: %w", err)
	}

	anchor := s.now().Truncate(time.Hour)
	note := SeedNote
	third := (len(emps) + 2) / 3

	var res Result
	for i, emp := range emps {
		var startsAt, endsAt time.Time
		switch {
		case i < third:
			startsAt, endsAt = anchor.Add(-6*time.Hour), anchor.Add(2*time.Hour)
		case i < 2*third:
			startsAt, endsAt = anchor.Add(-3*time.Hour), anchor.Add(5*time.Hour)
		default:
			startsAt, endsAt = anchor.Add(2*time.Hour), anchor.Add(10*time.Hour)
		}

		in := roster.CreateShiftInput{
			EmployeeID: emp.ID,
			StartsAt:   startsAt,
			EndsAt:     endsAt,
			Notes:      &note,
		}
		var loc *roster.Location
		if len(locs) > 0 {
			loc = &locs[i%len(locs)]
			in.LocationID = &loc.ID
		}
		if i < 6 && len(trs) > 0 {
			in.TourID = &trs[i%len(trs)].ID
		}

		shift, err := s.roster.CreatePublishedShift(ctx, in)
		if err != nil {
			return Result{}, fmt.Errorf("create seeded shift: %w", err)
		}
		res.Shifts++

		rec := demoRecordFor(i, shift, loc, s.now())
		if rec == nil {
			continue
		}
		if _, err := s.attendance.SeedDemoRecord(ctx, *rec); err != nil {
			return Result{}, fmt.Errorf("create seeded attendance: %w", err)
		}
		res.AttendanceRecords++
	}
	return res, nil
}

// demoRecordFor fabricates the attendance state for the i-th seeded shift:
// 0 → checked out, 1 → on break, 2..4 → checked in, everything else (the
// no-show and the upcoming rows) gets none. Only started shifts get records.
func demoRecordFor(i int, shift roster.Shift, loc *roster.Location, now time.Time) *attendance.DemoRecordInput {
	if i > 4 || shift.StartsAt.After(now) {
		return nil
	}
	checkIn := shift.StartsAt.Add(time.Duration(rand.IntN(10)) * time.Minute)
	rec := attendance.DemoRecordInput{
		EmployeeID: shift.EmployeeID,
		ShiftID:    &shift.ID,
		CheckInAt:  &checkIn,
		Status:     "checked_in",
	}
	if loc != nil && loc.Lat != nil && loc.Lng != nil {
		lat := *loc.Lat + (rand.Float64()-0.5)/1000
		lng := *loc.Lng + (rand.Float64()-0.5)/1000
		rec.CheckInLat, rec.CheckInLng = &lat, &lng
	}
	switch i {
	case 0:
		out := now.Add(-20 * time.Minute)
		rec.Status = "checked_out"
		rec.CheckOutAt = &out
		rec.CheckOutLat, rec.CheckOutLng = rec.CheckInLat, rec.CheckInLng
	case 1:
		breakStart := now.Add(-25 * time.Minute)
		rec.Status = "on_break"
		rec.BreakStartedAt = &breakStart
	}
	return &rec
}
