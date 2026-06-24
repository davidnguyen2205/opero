package leave

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
	reqs       map[uuid.UUID]Request
	entitled   map[int]int // year -> entitled (presence = row exists)
	approved   int         // value SumApprovedDays returns
	lastFilter Filter
}

func newFakeRepo() *fakeRepo {
	return &fakeRepo{reqs: map[uuid.UUID]Request{}, entitled: map[int]int{}}
}

func (f *fakeRepo) Create(_ context.Context, employeeID uuid.UUID, in CreateInput) (Request, error) {
	r := Request{
		ID: uuid.New(), EmployeeID: employeeID, Type: in.Type,
		StartDate: in.StartDate, EndDate: in.EndDate, Note: in.Note, Status: StatusPending,
	}
	f.reqs[r.ID] = r
	return r, nil
}

func (f *fakeRepo) Get(_ context.Context, id uuid.UUID) (Request, error) {
	if r, ok := f.reqs[id]; ok {
		return r, nil
	}
	return Request{}, ErrNotFound
}

func (f *fakeRepo) List(_ context.Context, filter Filter) ([]Request, error) {
	f.lastFilter = filter
	out := make([]Request, 0, len(f.reqs))
	for _, r := range f.reqs {
		if filter.EmployeeID != nil && r.EmployeeID != *filter.EmployeeID {
			continue
		}
		if filter.Status != nil && r.Status != *filter.Status {
			continue
		}
		out = append(out, r)
	}
	return out, nil
}

func (f *fakeRepo) SetStatus(_ context.Context, id uuid.UUID, status string, reviewedBy *uuid.UUID) (Request, error) {
	r, ok := f.reqs[id]
	if !ok {
		return Request{}, ErrNotFound
	}
	r.Status = status
	r.ReviewedBy = reviewedBy
	now := time.Now()
	r.ReviewedAt = &now
	f.reqs[id] = r
	return r, nil
}

func (f *fakeRepo) SumApprovedDays(context.Context, uuid.UUID, time.Time, time.Time) (int, error) {
	return f.approved, nil
}

func (f *fakeRepo) Entitlement(_ context.Context, _ uuid.UUID, year int) (int, bool, error) {
	v, ok := f.entitled[year]
	return v, ok, nil
}

type fakeResolver struct {
	empID uuid.UUID
	found bool
}

func (f fakeResolver) EmployeeIDByUserID(context.Context, uuid.UUID) (uuid.UUID, bool, error) {
	return f.empID, f.found, nil
}

func newSvc(res fakeResolver) (*Service, *fakeRepo) {
	f := newFakeRepo()
	svc := NewService(slog.New(slog.NewTextHandler(io.Discard, nil)), res)
	svc.newStore = func(context.Context) (repo, error) { return f, nil }
	return svc, f
}

func validInput() CreateInput {
	d := time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC)
	return CreateInput{Type: "holiday", StartDate: d, EndDate: d.AddDate(0, 0, 2)}
}

func TestCreateMyLeaveValidation(t *testing.T) {
	svc, _ := newSvc(fakeResolver{empID: uuid.New(), found: true})
	d := time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC)
	cases := []CreateInput{
		{Type: "vacation", StartDate: d, EndDate: d},                  // bad type
		{Type: "holiday", StartDate: d, EndDate: d.AddDate(0, 0, -1)}, // end before start
		{Type: "holiday"}, // zero dates
	}
	for i, in := range cases {
		if _, err := svc.CreateMyLeave(context.Background(), uuid.New(), in); !errors.Is(err, ErrValidation) {
			t.Errorf("case %d: err = %v, want ErrValidation", i, err)
		}
	}
}

func TestCreateMyLeaveNoEmployeeIsValidationError(t *testing.T) {
	svc, _ := newSvc(fakeResolver{found: false})
	if _, err := svc.CreateMyLeave(context.Background(), uuid.New(), validInput()); !errors.Is(err, ErrValidation) {
		t.Fatalf("err = %v, want ErrValidation", err)
	}
}

func TestCreateMyLeaveDefaultsPending(t *testing.T) {
	svc, _ := newSvc(fakeResolver{empID: uuid.New(), found: true})
	r, err := svc.CreateMyLeave(context.Background(), uuid.New(), validInput())
	if err != nil {
		t.Fatalf("CreateMyLeave: %v", err)
	}
	if r.Status != StatusPending {
		t.Errorf("status = %q, want pending", r.Status)
	}
}

func TestListMyLeaveUnlinkedReturnsEmpty(t *testing.T) {
	svc, _ := newSvc(fakeResolver{found: false})
	got, err := svc.ListMyLeave(context.Background(), uuid.New())
	if err != nil {
		t.Fatalf("ListMyLeave: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("len = %d, want 0", len(got))
	}
}

func TestListMyLeaveScopesToEmployee(t *testing.T) {
	empID := uuid.New()
	svc, f := newSvc(fakeResolver{empID: empID, found: true})
	if _, err := svc.ListMyLeave(context.Background(), uuid.New()); err != nil {
		t.Fatalf("ListMyLeave: %v", err)
	}
	if f.lastFilter.EmployeeID == nil || *f.lastFilter.EmployeeID != empID {
		t.Errorf("filter employee = %v, want %v", f.lastFilter.EmployeeID, empID)
	}
}

func TestApproveThenRejectIsBlocked(t *testing.T) {
	empID := uuid.New()
	svc, _ := newSvc(fakeResolver{empID: empID, found: true})
	r, err := svc.CreateMyLeave(context.Background(), uuid.New(), validInput())
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	reviewer := uuid.New()
	app, err := svc.Approve(context.Background(), r.ID, reviewer)
	if err != nil {
		t.Fatalf("approve: %v", err)
	}
	if app.Status != StatusApproved || app.ReviewedBy == nil || *app.ReviewedBy != reviewer {
		t.Fatalf("approve result = %+v", app)
	}
	// Re-approve is idempotent.
	if _, err := svc.Approve(context.Background(), r.ID, reviewer); err != nil {
		t.Fatalf("re-approve should be idempotent: %v", err)
	}
	// Rejecting an already-approved request is blocked.
	if _, err := svc.Reject(context.Background(), r.ID, reviewer); !errors.Is(err, ErrValidation) {
		t.Errorf("reject after approve: err = %v, want ErrValidation", err)
	}
}

func TestApproveNotFound(t *testing.T) {
	svc, _ := newSvc(fakeResolver{empID: uuid.New(), found: true})
	if _, err := svc.Approve(context.Background(), uuid.New(), uuid.New()); !errors.Is(err, ErrNotFound) {
		t.Errorf("err = %v, want ErrNotFound", err)
	}
}

func TestMyBalanceUsesDefaultWhenNoRow(t *testing.T) {
	svc, f := newSvc(fakeResolver{empID: uuid.New(), found: true})
	f.approved = 5
	bal, err := svc.MyBalance(context.Background(), uuid.New(), time.Date(2026, 6, 24, 0, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("MyBalance: %v", err)
	}
	if bal.EntitledDays != DefaultEntitledDays {
		t.Errorf("entitled = %d, want %d", bal.EntitledDays, DefaultEntitledDays)
	}
	if bal.UsedDays != 5 || bal.RemainingDays != DefaultEntitledDays-5 {
		t.Errorf("used=%d remaining=%d", bal.UsedDays, bal.RemainingDays)
	}
	if bal.Year != 2026 {
		t.Errorf("year = %d, want 2026", bal.Year)
	}
}

func TestMyBalanceUsesEntitlementRow(t *testing.T) {
	svc, f := newSvc(fakeResolver{empID: uuid.New(), found: true})
	f.entitled[2026] = 30
	f.approved = 2
	bal, err := svc.MyBalance(context.Background(), uuid.New(), time.Date(2026, 6, 24, 0, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("MyBalance: %v", err)
	}
	if bal.EntitledDays != 30 || bal.RemainingDays != 28 {
		t.Errorf("entitled=%d remaining=%d, want 30/28", bal.EntitledDays, bal.RemainingDays)
	}
}
