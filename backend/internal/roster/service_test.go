package roster

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/google/uuid"
)

type fakeRepo struct {
	locs            map[uuid.UUID]Location
	shifts          map[uuid.UUID]Shift
	lastShiftStatus string
	lastListFilter  ShiftFilter
}

func newFakeRepo() *fakeRepo {
	return &fakeRepo{locs: map[uuid.UUID]Location{}, shifts: map[uuid.UUID]Shift{}}
}

func (f *fakeRepo) CreateLocation(_ context.Context, in CreateLocationInput) (Location, error) {
	l := Location{ID: uuid.New(), Name: in.Name, Address: in.Address, Lat: in.Lat, Lng: in.Lng}
	f.locs[l.ID] = l
	return l, nil
}
func (f *fakeRepo) GetLocation(_ context.Context, id uuid.UUID) (Location, error) {
	if l, ok := f.locs[id]; ok {
		return l, nil
	}
	return Location{}, ErrNotFound
}
func (f *fakeRepo) ListLocations(context.Context) ([]Location, error) {
	out := make([]Location, 0, len(f.locs))
	for _, l := range f.locs {
		out = append(out, l)
	}
	return out, nil
}
func (f *fakeRepo) UpdateLocation(_ context.Context, id uuid.UUID, in UpdateLocationInput) (Location, error) {
	l, ok := f.locs[id]
	if !ok {
		return Location{}, ErrNotFound
	}
	if in.Name != nil {
		l.Name = *in.Name
	}
	f.locs[id] = l
	return l, nil
}
func (f *fakeRepo) DeleteLocation(_ context.Context, id uuid.UUID) error {
	if _, ok := f.locs[id]; !ok {
		return ErrNotFound
	}
	delete(f.locs, id)
	return nil
}
func (f *fakeRepo) CreateShift(_ context.Context, in CreateShiftInput, status string) (Shift, error) {
	f.lastShiftStatus = status
	s := Shift{ID: uuid.New(), EmployeeID: in.EmployeeID, LocationID: in.LocationID, StartsAt: in.StartsAt, EndsAt: in.EndsAt, Notes: in.Notes, Status: status}
	f.shifts[s.ID] = s
	return s, nil
}
func (f *fakeRepo) GetShift(_ context.Context, id uuid.UUID) (Shift, error) {
	if s, ok := f.shifts[id]; ok {
		return s, nil
	}
	return Shift{}, ErrNotFound
}
func (f *fakeRepo) ListShifts(_ context.Context, flt ShiftFilter) ([]Shift, error) {
	f.lastListFilter = flt
	out := make([]Shift, 0, len(f.shifts))
	for _, s := range f.shifts {
		out = append(out, s)
	}
	return out, nil
}
func (f *fakeRepo) UpdateShift(_ context.Context, id uuid.UUID, _ UpdateShiftInput) (Shift, error) {
	if s, ok := f.shifts[id]; ok {
		return s, nil
	}
	return Shift{}, ErrNotFound
}
func (f *fakeRepo) PublishShift(_ context.Context, id uuid.UUID) (Shift, error) {
	s, ok := f.shifts[id]
	if !ok {
		return Shift{}, ErrNotFound
	}
	s.Status = "published"
	f.shifts[id] = s
	return s, nil
}
func (f *fakeRepo) DeleteShift(_ context.Context, id uuid.UUID) error {
	if _, ok := f.shifts[id]; !ok {
		return ErrNotFound
	}
	delete(f.shifts, id)
	return nil
}

type fakeResolver struct {
	empID uuid.UUID
	found bool
}

func (f fakeResolver) EmployeeIDByUserID(context.Context, uuid.UUID) (uuid.UUID, bool, error) {
	return f.empID, f.found, nil
}

func newSvcWithFake() (*Service, *fakeRepo) {
	f := newFakeRepo()
	svc := NewService(slog.New(slog.NewTextHandler(io.Discard, nil)), fakeResolver{})
	svc.newStore = func(context.Context) (repo, error) { return f, nil }
	return svc, f
}

func TestCreateLocationValidation(t *testing.T) {
	svc, _ := newSvcWithFake()
	if _, err := svc.CreateLocation(context.Background(), CreateLocationInput{Name: "  "}); !errors.Is(err, ErrValidation) {
		t.Fatalf("err = %v, want ErrValidation", err)
	}
}

func TestCreateShiftValidation(t *testing.T) {
	svc, _ := newSvcWithFake()
	now := time.Now()
	cases := []CreateShiftInput{
		{EmployeeID: uuid.Nil, StartsAt: now, EndsAt: now.Add(time.Hour)},    // no employee
		{EmployeeID: uuid.New(), StartsAt: now, EndsAt: now},                 // ends == starts
		{EmployeeID: uuid.New(), StartsAt: now, EndsAt: now.Add(-time.Hour)}, // ends < starts
		{EmployeeID: uuid.New(), EndsAt: now.Add(time.Hour)},                 // zero starts
	}
	for i, in := range cases {
		if _, err := svc.CreateShift(context.Background(), in); !errors.Is(err, ErrValidation) {
			t.Errorf("case %d: err = %v, want ErrValidation", i, err)
		}
	}
}

func TestCreateShiftDefaultsDraftAndPublish(t *testing.T) {
	svc, f := newSvcWithFake()
	now := time.Now()
	s, err := svc.CreateShift(context.Background(), CreateShiftInput{
		EmployeeID: uuid.New(), StartsAt: now, EndsAt: now.Add(2 * time.Hour),
	})
	if err != nil {
		t.Fatalf("CreateShift: %v", err)
	}
	if f.lastShiftStatus != "draft" {
		t.Errorf("created status = %q, want draft", f.lastShiftStatus)
	}
	pub, err := svc.PublishShift(context.Background(), s.ID)
	if err != nil {
		t.Fatalf("PublishShift: %v", err)
	}
	if pub.Status != "published" {
		t.Errorf("published status = %q, want published", pub.Status)
	}
}

func TestUpdateShiftTimeOrderValidation(t *testing.T) {
	svc, _ := newSvcWithFake()
	now := time.Now()
	earlier := now.Add(-time.Hour)
	if _, err := svc.UpdateShift(context.Background(), uuid.New(), UpdateShiftInput{StartsAt: &now, EndsAt: &earlier}); !errors.Is(err, ErrValidation) {
		t.Fatalf("err = %v, want ErrValidation", err)
	}
}

func TestGetShiftNotFoundPropagates(t *testing.T) {
	svc, _ := newSvcWithFake()
	if _, err := svc.GetShift(context.Background(), uuid.New()); !errors.Is(err, ErrNotFound) {
		t.Fatalf("err = %v, want ErrNotFound", err)
	}
}

func TestNoTenantInContext(t *testing.T) {
	svc := NewService(slog.New(slog.NewTextHandler(io.Discard, nil)), fakeResolver{})
	if _, err := svc.ListShifts(context.Background(), ShiftFilter{}); !errors.Is(err, ErrNoTenant) {
		t.Fatalf("err = %v, want ErrNoTenant", err)
	}
}

func TestListMyShiftsUnlinkedUserReturnsEmpty(t *testing.T) {
	f := newFakeRepo()
	svc := NewService(slog.New(slog.NewTextHandler(io.Discard, nil)), fakeResolver{found: false})
	svc.newStore = func(context.Context) (repo, error) { return f, nil }
	got, err := svc.ListMyShifts(context.Background(), uuid.New(), ShiftFilter{})
	if err != nil {
		t.Fatalf("ListMyShifts: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("want empty for unlinked user, got %d", len(got))
	}
}

func TestListMyShiftsScopesToEmployee(t *testing.T) {
	emp := uuid.New()
	f := newFakeRepo()
	svc := NewService(slog.New(slog.NewTextHandler(io.Discard, nil)), fakeResolver{empID: emp, found: true})
	svc.newStore = func(context.Context) (repo, error) { return f, nil }
	if _, err := svc.ListMyShifts(context.Background(), uuid.New(), ShiftFilter{}); err != nil {
		t.Fatalf("ListMyShifts: %v", err)
	}
	if f.lastListFilter.EmployeeID == nil || *f.lastListFilter.EmployeeID != emp {
		t.Errorf("ListShifts not scoped to resolved employee: %v", f.lastListFilter.EmployeeID)
	}
}
