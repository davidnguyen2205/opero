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
	r.BreakStartedAt = nil // mirrors the SQL: checkout ends any open break
	f.byClient[in.ClientID] = r
	return r, nil
}
func (f *fakeRepo) SetStatus(_ context.Context, clientID uuid.UUID, status string) (Record, error) {
	r, ok := f.byClient[clientID]
	if !ok {
		return Record{}, ErrNotFound
	}
	r.Status = status
	// Mirror the SQL: stamp break_started_at on entering a break, clear it on
	// any other transition.
	if status == "on_break" {
		now := time.Now()
		r.BreakStartedAt = &now
	} else {
		r.BreakStartedAt = nil
	}
	f.byClient[clientID] = r
	return r, nil
}
func (f *fakeRepo) List(context.Context, AttendanceFilter) ([]Record, error) {
	out := make([]Record, 0, len(f.byClient))
	for _, r := range f.byClient {
		out = append(out, r)
	}
	return out, nil
}
func (f *fakeRepo) CreateDemoRecord(_ context.Context, in DemoRecordInput) (Record, error) {
	r := Record{
		ID: uuid.New(), EmployeeID: in.EmployeeID, ShiftID: in.ShiftID, ClientID: uuid.New(),
		CheckInAt: in.CheckInAt, CheckOutAt: in.CheckOutAt,
		BreakStartedAt: in.BreakStartedAt, Status: in.Status,
	}
	f.byClient[r.ClientID] = r
	return r, nil
}
func (f *fakeRepo) DeleteByShiftIDs(_ context.Context, shiftIDs []uuid.UUID) (int64, error) {
	want := make(map[uuid.UUID]bool, len(shiftIDs))
	for _, id := range shiftIDs {
		want[id] = true
	}
	var n int64
	for cid, r := range f.byClient {
		if r.ShiftID != nil && want[*r.ShiftID] {
			delete(f.byClient, cid)
			n++
		}
	}
	return n, nil
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

type shiftInfo struct {
	empID  uuid.UUID
	status string
}

type fakeShiftResolver struct {
	shifts map[uuid.UUID]shiftInfo
	calls  int
	err    error
}

func newFakeShiftResolver() *fakeShiftResolver {
	return &fakeShiftResolver{shifts: map[uuid.UUID]shiftInfo{}}
}

func (f *fakeShiftResolver) ShiftOwnerByID(_ context.Context, id uuid.UUID) (uuid.UUID, string, bool, error) {
	f.calls++
	if f.err != nil {
		return uuid.Nil, "", false, f.err
	}
	s, ok := f.shifts[id]
	if !ok {
		return uuid.Nil, "", false, nil
	}
	return s.empID, s.status, true, nil
}

func newSvc(res fakeResolver) (*Service, *fakeRepo, *fakeShiftResolver) {
	fr := newFakeRepo()
	fs := newFakeShiftResolver()
	svc := NewService(slog.New(slog.NewTextHandler(io.Discard, nil)), res, fs)
	svc.newStore = func(context.Context) (repo, error) { return fr, nil }
	return svc, fr, fs
}

func TestCheckInRequiresClientID(t *testing.T) {
	svc, _, _ := newSvc(fakeResolver{empID: uuid.New(), found: true})
	if _, _, err := svc.CheckIn(context.Background(), uuid.New(), CheckInInput{}); !errors.Is(err, ErrValidation) {
		t.Fatalf("err = %v, want ErrValidation", err)
	}
}

func TestCheckInNoEmployeeLinked(t *testing.T) {
	svc, _, _ := newSvc(fakeResolver{found: false})
	if _, _, err := svc.CheckIn(context.Background(), uuid.New(), CheckInInput{ClientID: uuid.New()}); !errors.Is(err, ErrValidation) {
		t.Fatalf("err = %v, want ErrValidation", err)
	}
}

func TestCheckInIdempotent(t *testing.T) {
	svc, _, _ := newSvc(fakeResolver{empID: uuid.New(), found: true})
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

	svc := NewService(slog.New(slog.NewTextHandler(io.Discard, nil)), fakeResolver{empID: uuid.New(), found: true}, newFakeShiftResolver())
	svc.newStore = func(context.Context) (repo, error) { return fr, nil }

	if _, _, err := svc.CheckIn(context.Background(), uuid.New(), CheckInInput{ClientID: cid}); !errors.Is(err, ErrConflict) {
		t.Fatalf("err = %v, want ErrConflict", err)
	}
}

func TestCheckInWithOwnPublishedShiftSucceeds(t *testing.T) {
	emp := uuid.New()
	svc, _, fs := newSvc(fakeResolver{empID: emp, found: true})
	shiftID := uuid.New()
	fs.shifts[shiftID] = shiftInfo{empID: emp, status: "published"}

	rec, created, err := svc.CheckIn(context.Background(), uuid.New(), CheckInInput{ClientID: uuid.New(), ShiftID: &shiftID})
	if err != nil || !created {
		t.Fatalf("check-in: created=%v err=%v", created, err)
	}
	if rec.ShiftID == nil || *rec.ShiftID != shiftID {
		t.Errorf("record shift_id = %v, want %v", rec.ShiftID, shiftID)
	}
}

func TestCheckInWithForeignShiftFailsValidation(t *testing.T) {
	svc, fr, fs := newSvc(fakeResolver{empID: uuid.New(), found: true})
	shiftID := uuid.New()
	fs.shifts[shiftID] = shiftInfo{empID: uuid.New(), status: "published"} // someone else's shift

	_, _, err := svc.CheckIn(context.Background(), uuid.New(), CheckInInput{ClientID: uuid.New(), ShiftID: &shiftID})
	if !errors.Is(err, ErrValidation) {
		t.Fatalf("err = %v, want ErrValidation", err)
	}
	if len(fr.byClient) != 0 {
		t.Errorf("record was created despite validation failure")
	}
}

func TestCheckInWithOwnDraftShiftFailsValidation(t *testing.T) {
	emp := uuid.New()
	svc, fr, fs := newSvc(fakeResolver{empID: emp, found: true})
	shiftID := uuid.New()
	fs.shifts[shiftID] = shiftInfo{empID: emp, status: "draft"}

	_, _, err := svc.CheckIn(context.Background(), uuid.New(), CheckInInput{ClientID: uuid.New(), ShiftID: &shiftID})
	if !errors.Is(err, ErrValidation) {
		t.Fatalf("err = %v, want ErrValidation", err)
	}
	if len(fr.byClient) != 0 {
		t.Errorf("record was created despite validation failure")
	}
}

func TestCheckInWithNonexistentShiftFailsIndistinguishably(t *testing.T) {
	emp := uuid.New()
	svc, fr, fs := newSvc(fakeResolver{empID: emp, found: true})

	missing := uuid.New()
	_, _, missErr := svc.CheckIn(context.Background(), uuid.New(), CheckInInput{ClientID: uuid.New(), ShiftID: &missing})
	if !errors.Is(missErr, ErrValidation) {
		t.Fatalf("nonexistent: err = %v, want ErrValidation", missErr)
	}

	foreign := uuid.New()
	fs.shifts[foreign] = shiftInfo{empID: uuid.New(), status: "published"}
	_, _, foreignErr := svc.CheckIn(context.Background(), uuid.New(), CheckInInput{ClientID: uuid.New(), ShiftID: &foreign})
	if !errors.Is(foreignErr, ErrValidation) {
		t.Fatalf("foreign: err = %v, want ErrValidation", foreignErr)
	}
	// The two failure modes must be indistinguishable (no existence oracle).
	if missErr.Error() != foreignErr.Error() {
		t.Errorf("error text differs:\n nonexistent: %q\n foreign:     %q", missErr, foreignErr)
	}
	if len(fr.byClient) != 0 {
		t.Errorf("record was created despite validation failure")
	}
}

func TestCheckInWithoutShiftSkipsShiftLookup(t *testing.T) {
	svc, _, fs := newSvc(fakeResolver{empID: uuid.New(), found: true})
	if _, created, err := svc.CheckIn(context.Background(), uuid.New(), CheckInInput{ClientID: uuid.New()}); err != nil || !created {
		t.Fatalf("check-in: created=%v err=%v", created, err)
	}
	if fs.calls != 0 {
		t.Errorf("shift resolver called %d times, want 0", fs.calls)
	}
}

func TestCheckInReplayWithShiftIDSkipsRevalidation(t *testing.T) {
	emp := uuid.New()
	svc, _, fs := newSvc(fakeResolver{empID: emp, found: true})
	shiftID := uuid.New()
	fs.shifts[shiftID] = shiftInfo{empID: emp, status: "published"}
	ctx, user, cid := context.Background(), uuid.New(), uuid.New()

	r1, created, err := svc.CheckIn(ctx, user, CheckInInput{ClientID: cid, ShiftID: &shiftID})
	if err != nil || !created {
		t.Fatalf("first check-in: created=%v err=%v", created, err)
	}
	fs.calls = 0
	fs.err = errors.New("resolver must not be called on replay")

	r2, created2, err := svc.CheckIn(ctx, user, CheckInInput{ClientID: cid, ShiftID: &shiftID})
	if err != nil || created2 {
		t.Fatalf("replay: created=%v err=%v (want created=false)", created2, err)
	}
	if r1.ID != r2.ID {
		t.Errorf("replay returned a different record: %v vs %v", r1.ID, r2.ID)
	}
	if fs.calls != 0 {
		t.Errorf("shift resolver called %d times on replay, want 0", fs.calls)
	}
}

func TestSetBreakTogglesStatus(t *testing.T) {
	emp := uuid.New()
	svc, _, _ := newSvc(fakeResolver{empID: emp, found: true})
	ctx, user, cid := context.Background(), uuid.New(), uuid.New()
	if _, _, err := svc.CheckIn(ctx, user, CheckInInput{ClientID: cid}); err != nil {
		t.Fatalf("check-in: %v", err)
	}
	r, err := svc.SetBreak(ctx, user, cid, true)
	if err != nil || r.Status != "on_break" {
		t.Fatalf("start break: status=%q err=%v", r.Status, err)
	}
	if r.BreakStartedAt == nil {
		t.Error("start break: break_started_at not set")
	}
	// idempotent on the same target — must not restart the break clock
	r2, err := svc.SetBreak(ctx, user, cid, true)
	if err != nil || r2.Status != "on_break" {
		t.Fatalf("re-break: status=%q err=%v", r2.Status, err)
	}
	if r2.BreakStartedAt == nil || !r2.BreakStartedAt.Equal(*r.BreakStartedAt) {
		t.Errorf("re-break moved break_started_at: %v vs %v", r2.BreakStartedAt, r.BreakStartedAt)
	}
	r3, err := svc.SetBreak(ctx, user, cid, false)
	if err != nil || r3.Status != "checked_in" {
		t.Fatalf("resume: status=%q err=%v", r3.Status, err)
	}
	if r3.BreakStartedAt != nil {
		t.Errorf("resume: break_started_at not cleared: %v", r3.BreakStartedAt)
	}
	// unknown client_id
	if _, err := svc.SetBreak(ctx, user, uuid.New(), true); !errors.Is(err, ErrNotFound) {
		t.Fatalf("unknown: err = %v, want ErrNotFound", err)
	}
}

func TestCheckOutIdempotentAndOwnership(t *testing.T) {
	emp := uuid.New()
	svc, fr, _ := newSvc(fakeResolver{empID: emp, found: true})
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
	other := NewService(slog.New(slog.NewTextHandler(io.Discard, nil)), fakeResolver{empID: uuid.New(), found: true}, newFakeShiftResolver())
	other.newStore = func(context.Context) (repo, error) { return fr, nil }
	if _, err := other.CheckOut(ctx, uuid.New(), CheckOutInput{ClientID: cid}); !errors.Is(err, ErrNotFound) {
		t.Fatalf("cross-employee checkout: err = %v, want ErrNotFound", err)
	}
}
