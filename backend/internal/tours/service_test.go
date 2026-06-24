package tours

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"

	"github.com/google/uuid"
)

type fakeRepo struct {
	tours      map[uuid.UUID]Tour
	lastFilter Filter
}

func newFakeRepo() *fakeRepo { return &fakeRepo{tours: map[uuid.UUID]Tour{}} }

func (f *fakeRepo) Create(_ context.Context, in CreateInput) (Tour, error) {
	t := Tour{ID: uuid.New(), Name: in.Name, Category: in.Category, Active: in.Active, DepartureTimes: in.DepartureTimes}
	f.tours[t.ID] = t
	return t, nil
}
func (f *fakeRepo) Get(_ context.Context, id uuid.UUID) (Tour, error) {
	if t, ok := f.tours[id]; ok {
		return t, nil
	}
	return Tour{}, ErrNotFound
}
func (f *fakeRepo) List(_ context.Context, filter Filter) ([]Tour, error) {
	f.lastFilter = filter
	out := make([]Tour, 0, len(f.tours))
	for _, t := range f.tours {
		if filter.Category != nil && t.Category != *filter.Category {
			continue
		}
		if filter.Active != nil && t.Active != *filter.Active {
			continue
		}
		out = append(out, t)
	}
	return out, nil
}
func (f *fakeRepo) Update(_ context.Context, id uuid.UUID, in UpdateInput) (Tour, error) {
	t, ok := f.tours[id]
	if !ok {
		return Tour{}, ErrNotFound
	}
	if in.Name != nil {
		t.Name = *in.Name
	}
	if in.Category != nil {
		t.Category = *in.Category
	}
	f.tours[id] = t
	return t, nil
}
func (f *fakeRepo) Delete(_ context.Context, id uuid.UUID) error {
	if _, ok := f.tours[id]; !ok {
		return ErrNotFound
	}
	delete(f.tours, id)
	return nil
}

func newSvc() (*Service, *fakeRepo) {
	f := newFakeRepo()
	svc := NewService(slog.New(slog.NewTextHandler(io.Discard, nil)))
	svc.newStore = func(context.Context) (repo, error) { return f, nil }
	return svc, f
}

func TestCreateValidation(t *testing.T) {
	svc, _ := newSvc()
	cases := []CreateInput{
		{Name: "  ", Category: "walking"},   // blank name
		{Name: "Alfama", Category: "scuba"}, // bad category
	}
	for i, in := range cases {
		if _, err := svc.Create(context.Background(), in); !errors.Is(err, ErrValidation) {
			t.Errorf("case %d: err = %v, want ErrValidation", i, err)
		}
	}
}

func TestCreateThenGet(t *testing.T) {
	svc, _ := newSvc()
	created, err := svc.Create(context.Background(), CreateInput{Name: "Alfama Walking Tour", Category: "walking", Active: true})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	got, err := svc.Get(context.Background(), created.ID)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if got.Name != "Alfama Walking Tour" || got.Category != "walking" {
		t.Errorf("got %+v", got)
	}
}

func TestUpdateValidation(t *testing.T) {
	svc, f := newSvc()
	t0, _ := svc.Create(context.Background(), CreateInput{Name: "X", Category: "walking", Active: true})
	bad := "nope"
	if _, err := svc.Update(context.Background(), t0.ID, UpdateInput{Category: &bad}); !errors.Is(err, ErrValidation) {
		t.Errorf("bad category: err = %v, want ErrValidation", err)
	}
	blank := "  "
	if _, err := svc.Update(context.Background(), t0.ID, UpdateInput{Name: &blank}); !errors.Is(err, ErrValidation) {
		t.Errorf("blank name: err = %v, want ErrValidation", err)
	}
	_ = f
}

func TestListFilterPassthrough(t *testing.T) {
	svc, f := newSvc()
	active := true
	cat := "food"
	if _, err := svc.List(context.Background(), Filter{Active: &active, Category: &cat}); err != nil {
		t.Fatalf("List: %v", err)
	}
	if f.lastFilter.Active == nil || !*f.lastFilter.Active || f.lastFilter.Category == nil || *f.lastFilter.Category != "food" {
		t.Errorf("filter not passed through: %+v", f.lastFilter)
	}
}

func TestDeleteNotFound(t *testing.T) {
	svc, _ := newSvc()
	if err := svc.Delete(context.Background(), uuid.New()); !errors.Is(err, ErrNotFound) {
		t.Errorf("err = %v, want ErrNotFound", err)
	}
}
