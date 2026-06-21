package attendance

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
	byClient map[uuid.UUID]Record
}

func newFakeRepo() *fakeRepo { return &fakeRepo{byClient: map[uuid.UUID]Record{}} }

func (f *fakeRepo) GetByClientID(_ context.Context, clientID uuid.UUID) (Record, error) {
	if r, ok := f.byClient[clientID]; ok {
		return r, nil
	}
	return Record{}, ErrNotFound
}
func (f *fakeRepo) CreateCheckIn(_ context.Context, empID uuid.UUID, in CheckInInput) (Record, error) {
	if _, ok := f.byClient[in.ClientID]; ok {
		return Record{}, ErrConflict
	}
	now := time.Now()
	r := Record{ID: uuid.New(), EmployeeID: empID, ClientID: in.ClientID, ShiftID: in.ShiftID, CheckInAt: &now, Status: "checked_in"}
	f.byClient[in.ClientID] = r
	return r, nil
}
func (f *fakeRepo) CheckOut(_ context.Context, in CheckOutInput) (Record, error) {
	r, ok := f.byClient[in.ClientID]
	if !ok {
		return Record{}, ErrNotFound
	}
	now := time.Now()
	r.CheckOutAt = &now
	r.Status = "checked_out"
	f.byClient[in.ClientID] = r
	return r, nil
}
func (f *fakeRepo) List(context.Context, AttendanceFilter) ([]Record, error) {
	out := make([]Record, 0, len(f.byClient))
	for _, r := range f.byClient {
		out = append(out, r)
	}
	return out, nil
}
func (f *fakeRepo) ListByShiftIDs(_ context.Context, shiftIDs []uuid.UUID) ([]Record, error) {
	want := make(map[uuid.UUID]bool, len(shiftIDs))
	for _, id := range shiftIDs {
		want[id] = true
	}
	out := make([]Record, 0)
	for _, r := range f.byClient {
		if r.ShiftID != nil && want[*r.ShiftID] {
			out = append(out, r)
		}
	}
	return out, nil
}

type fakeResolver struct {
	empID uuid.UUID
	found bool
	err   error
}

func (f fakeResolver) EmployeeIDByUserID(context.Context, uuid.UUID) (uuid.UUID, bool, error) {
	return f.empID, f.found, f.err
}

func newSvc(res fakeResolver) (*Service, *fakeRepo) {
	fr := newFakeRepo()
	svc := NewService(slog.New(slog.NewTextHandler(io.Discard, nil)), res)
	svc.newStore = func(context.Context) (repo, error) { return fr, nil }
	return svc, fr
}

func TestCheckInRequiresClientID(t *testing.T) {
	svc, _ := newSvc(fakeResolver{empID: uuid.New(), found: true})
	if _, _, err := svc.CheckIn(context.Background(), uuid.New(), CheckInInput{}); !errors.Is(err, ErrValidation) {
		t.Fatalf("err = %v, want ErrValidation", err)
	}
}

func TestCheckInNoEmployeeLinked(t *testing.T) {
	svc, _ := newSvc(fakeResolver{found: false})
	if _, _, err := svc.CheckIn(context.Background(), uuid.New(), CheckInInput{ClientID: uuid.New()}); !errors.Is(err, ErrValidation) {
		t.Fatalf("err = %v, want ErrValidation", err)
	}
}

func TestCheckInIdempotent(t *testing.T) {
	svc, _ := newSvc(fakeResolver{empID: uuid.New(), found: true})
	ctx, user, cid := context.Background(), uuid.New(), uuid.New()

	r1, created, err := svc.CheckIn(ctx, user, CheckInInput{ClientID: cid})
	if err != nil || !created {
		t.Fatalf("first check-in: created=%v err=%v", created, err)
	}
	r2, created2, err := svc.CheckIn(ctx, user, CheckInInput{ClientID: cid})
	if err != nil || created2 {
		t.Fatalf("replay: created=%v err=%v (want created=false)", created2, err)
	}
	if r1.ID != r2.ID {
		t.Errorf("replay returned a different record: %v vs %v", r1.ID, r2.ID)
	}
}

func TestCheckInClientIDReuseByAnotherEmployeeConflicts(t *testing.T) {
	fr := newFakeRepo()
	cid := uuid.New()
	owner := uuid.New()
	fr.byClient[cid] = Record{ID: uuid.New(), EmployeeID: owner, ClientID: cid, Status: "checked_in"}

	svc := NewService(slog.New(slog.NewTextHandler(io.Discard, nil)), fakeResolver{empID: uuid.New(), found: true})
	svc.newStore = func(context.Context) (repo, error) { return fr, nil }

	if _, _, err := svc.CheckIn(context.Background(), uuid.New(), CheckInInput{ClientID: cid}); !errors.Is(err, ErrConflict) {
		t.Fatalf("err = %v, want ErrConflict", err)
	}
}

func TestCheckOutIdempotentAndOwnership(t *testing.T) {
	emp := uuid.New()
	svc, fr := newSvc(fakeResolver{empID: emp, found: true})
	ctx, user, cid := context.Background(), uuid.New(), uuid.New()
	if _, _, err := svc.CheckIn(ctx, user, CheckInInput{ClientID: cid}); err != nil {
		t.Fatalf("check-in: %v", err)
	}

	r, err := svc.CheckOut(ctx, user, CheckOutInput{ClientID: cid})
	if err != nil || r.Status != "checked_out" {
		t.Fatalf("check-out: status=%q err=%v", r.Status, err)
	}
	// idempotent re-checkout
	if r2, err := svc.CheckOut(ctx, user, CheckOutInput{ClientID: cid}); err != nil || r2.Status != "checked_out" {
		t.Fatalf("re-checkout: status=%q err=%v", r2.Status, err)
	}
	// unknown client_id
	if _, err := svc.CheckOut(ctx, user, CheckOutInput{ClientID: uuid.New()}); !errors.Is(err, ErrNotFound) {
		t.Fatalf("unknown: err = %v, want ErrNotFound", err)
	}
	// another employee can't check out this record
	other := NewService(slog.New(slog.NewTextHandler(io.Discard, nil)), fakeResolver{empID: uuid.New(), found: true})
	other.newStore = func(context.Context) (repo, error) { return fr, nil }
	if _, err := other.CheckOut(ctx, uuid.New(), CheckOutInput{ClientID: cid}); !errors.Is(err, ErrNotFound) {
		t.Fatalf("cross-employee checkout: err = %v, want ErrNotFound", err)
	}
}
