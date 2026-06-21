package identity

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"

	"github.com/google/uuid"
)

type fakeRepo struct {
	depts           map[uuid.UUID]Department
	emps            map[uuid.UUID]Employee
	roles           map[uuid.UUID]Role
	lastCreateEmp   CreateEmployeeInput
	createDeptCalls int
}

func newFakeRepo() *fakeRepo {
	return &fakeRepo{
		depts: map[uuid.UUID]Department{},
		emps:  map[uuid.UUID]Employee{},
		roles: map[uuid.UUID]Role{},
	}
}

func (f *fakeRepo) CreateDepartment(_ context.Context, in CreateDepartmentInput) (Department, error) {
	f.createDeptCalls++
	d := Department{ID: uuid.New(), Name: in.Name, ParentID: in.ParentID}
	f.depts[d.ID] = d
	return d, nil
}
func (f *fakeRepo) GetDepartment(_ context.Context, id uuid.UUID) (Department, error) {
	if d, ok := f.depts[id]; ok {
		return d, nil
	}
	return Department{}, ErrNotFound
}
func (f *fakeRepo) ListDepartments(context.Context) ([]Department, error) {
	out := make([]Department, 0, len(f.depts))
	for _, d := range f.depts {
		out = append(out, d)
	}
	return out, nil
}
func (f *fakeRepo) UpdateDepartment(_ context.Context, id uuid.UUID, in UpdateDepartmentInput) (Department, error) {
	d, ok := f.depts[id]
	if !ok {
		return Department{}, ErrNotFound
	}
	if in.Name != nil {
		d.Name = *in.Name
	}
	f.depts[id] = d
	return d, nil
}
func (f *fakeRepo) DeleteDepartment(_ context.Context, id uuid.UUID) error {
	if _, ok := f.depts[id]; !ok {
		return ErrNotFound
	}
	delete(f.depts, id)
	return nil
}
func (f *fakeRepo) CreateEmployee(_ context.Context, in CreateEmployeeInput) (Employee, error) {
	f.lastCreateEmp = in
	e := Employee{ID: uuid.New(), FullName: in.FullName, EmploymentType: in.EmploymentType, Status: in.Status}
	f.emps[e.ID] = e
	return e, nil
}
func (f *fakeRepo) GetEmployee(_ context.Context, id uuid.UUID) (Employee, error) {
	if e, ok := f.emps[id]; ok {
		return e, nil
	}
	return Employee{}, ErrNotFound
}
func (f *fakeRepo) ListEmployees(context.Context, EmployeeFilter) ([]Employee, error) {
	out := make([]Employee, 0, len(f.emps))
	for _, e := range f.emps {
		out = append(out, e)
	}
	return out, nil
}
func (f *fakeRepo) UpdateEmployee(_ context.Context, id uuid.UUID, _ UpdateEmployeeInput) (Employee, error) {
	if e, ok := f.emps[id]; ok {
		return e, nil
	}
	return Employee{}, ErrNotFound
}
func (f *fakeRepo) DeleteEmployee(_ context.Context, id uuid.UUID) error {
	if _, ok := f.emps[id]; !ok {
		return ErrNotFound
	}
	delete(f.emps, id)
	return nil
}

func (f *fakeRepo) GetEmployeeByUserID(_ context.Context, userID uuid.UUID) (Employee, error) {
	for _, e := range f.emps {
		if e.UserID != nil && *e.UserID == userID {
			return e, nil
		}
	}
	return Employee{}, ErrNotFound
}
func (f *fakeRepo) SetEmployeeUserID(_ context.Context, employeeID, userID uuid.UUID) (Employee, error) {
	e, ok := f.emps[employeeID]
	if !ok {
		return Employee{}, ErrNotFound
	}
	e.UserID = &userID
	f.emps[employeeID] = e
	return e, nil
}

type fakeUserCreator struct {
	created   map[uuid.UUID]bool
	createErr error
}

func (f *fakeUserCreator) CreateUser(_ context.Context, _ uuid.UUID, _, _, _ string) (uuid.UUID, error) {
	if f.createErr != nil {
		return uuid.Nil, f.createErr
	}
	id := uuid.New()
	if f.created == nil {
		f.created = map[uuid.UUID]bool{}
	}
	f.created[id] = true
	return id, nil
}

func (f *fakeUserCreator) DeleteUser(_ context.Context, id uuid.UUID) error {
	delete(f.created, id)
	return nil
}

func (f *fakeRepo) CreateRole(_ context.Context, in CreateRoleInput) (Role, error) {
	r := Role{ID: uuid.New(), Name: in.Name, Description: in.Description, Permissions: in.Permissions}
	f.roles[r.ID] = r
	return r, nil
}
func (f *fakeRepo) GetRole(_ context.Context, id uuid.UUID) (Role, error) {
	if r, ok := f.roles[id]; ok {
		return r, nil
	}
	return Role{}, ErrNotFound
}
func (f *fakeRepo) ListRoles(context.Context) ([]Role, error) {
	out := make([]Role, 0, len(f.roles))
	for _, r := range f.roles {
		out = append(out, r)
	}
	return out, nil
}
func (f *fakeRepo) UpdateRole(_ context.Context, id uuid.UUID, in UpdateRoleInput) (Role, error) {
	r, ok := f.roles[id]
	if !ok {
		return Role{}, ErrNotFound
	}
	if in.Name != nil {
		r.Name = *in.Name
	}
	f.roles[id] = r
	return r, nil
}
func (f *fakeRepo) DeleteRole(_ context.Context, id uuid.UUID) error {
	if _, ok := f.roles[id]; !ok {
		return ErrNotFound
	}
	delete(f.roles, id)
	return nil
}

func newSvcWithFake() (*Service, *fakeRepo) {
	f := newFakeRepo()
	svc := NewService(slog.New(slog.NewTextHandler(io.Discard, nil)), &fakeUserCreator{})
	svc.newStore = func(context.Context) (repo, error) { return f, nil }
	return svc, f
}

func TestCreateEmployeeLogin(t *testing.T) {
	fr := newFakeRepo()
	uc := &fakeUserCreator{}
	svc := NewService(slog.New(slog.NewTextHandler(io.Discard, nil)), uc)
	svc.newStore = func(context.Context) (repo, error) { return fr, nil }

	emp, err := fr.CreateEmployee(context.Background(), CreateEmployeeInput{
		FullName: "Field A", EmploymentType: "full_time", Status: "active",
	})
	if err != nil {
		t.Fatalf("seed employee: %v", err)
	}
	tid := uuid.New()

	res, err := svc.CreateEmployeeLogin(context.Background(), tid, emp.ID, "field@a.test", "password1", "")
	if err != nil {
		t.Fatalf("CreateEmployeeLogin: %v", err)
	}
	if res.Role != "employee" { // defaulted
		t.Errorf("role = %q, want employee", res.Role)
	}
	if len(uc.created) != 1 {
		t.Errorf("expected one user created, got %d", len(uc.created))
	}

	// second attempt: employee already has a login -> conflict
	if _, err := svc.CreateEmployeeLogin(context.Background(), tid, emp.ID, "z@a.test", "password1", "manager"); !errors.Is(err, ErrConflict) {
		t.Fatalf("err = %v, want ErrConflict", err)
	}
	// unknown employee -> not found
	if _, err := svc.CreateEmployeeLogin(context.Background(), tid, uuid.New(), "q@a.test", "password1", ""); !errors.Is(err, ErrNotFound) {
		t.Fatalf("err = %v, want ErrNotFound", err)
	}
}

func TestCreateEmployeeLoginCompensatesOnLinkFailure(t *testing.T) {
	// If linking the user to the employee fails, the created login is removed.
	fr := newFakeRepo()
	uc := &fakeUserCreator{}
	svc := NewService(slog.New(slog.NewTextHandler(io.Discard, nil)), uc)
	// store whose SetEmployeeUserID fails after GetEmployee succeeds
	svc.newStore = func(context.Context) (repo, error) { return &linkFailRepo{fakeRepo: fr}, nil }
	emp, _ := fr.CreateEmployee(context.Background(), CreateEmployeeInput{FullName: "B", EmploymentType: "full_time", Status: "active"})

	if _, err := svc.CreateEmployeeLogin(context.Background(), uuid.New(), emp.ID, "b@a.test", "password1", ""); err == nil {
		t.Fatal("expected link failure to surface")
	}
	if len(uc.created) != 0 {
		t.Errorf("orphan login not cleaned up: %d remain", len(uc.created))
	}
}

// linkFailRepo behaves like fakeRepo but fails SetEmployeeUserID.
type linkFailRepo struct{ *fakeRepo }

func (r *linkFailRepo) SetEmployeeUserID(context.Context, uuid.UUID, uuid.UUID) (Employee, error) {
	return Employee{}, errors.New("link failed")
}

func TestCreateDepartmentValidation(t *testing.T) {
	svc, f := newSvcWithFake()
	if _, err := svc.CreateDepartment(context.Background(), CreateDepartmentInput{Name: "  "}); !errors.Is(err, ErrValidation) {
		t.Fatalf("err = %v, want ErrValidation", err)
	}
	if f.createDeptCalls != 0 {
		t.Error("store should not be called when validation fails")
	}
}

func TestCreateEmployeeDefaultsStatusActive(t *testing.T) {
	svc, f := newSvcWithFake()
	_, err := svc.CreateEmployee(context.Background(), CreateEmployeeInput{FullName: "Ann", EmploymentType: "full_time"})
	if err != nil {
		t.Fatalf("CreateEmployee: %v", err)
	}
	if f.lastCreateEmp.Status != "active" {
		t.Errorf("status = %q, want active (default)", f.lastCreateEmp.Status)
	}
}

func TestCreateEmployeeValidation(t *testing.T) {
	svc, _ := newSvcWithFake()
	cases := []CreateEmployeeInput{
		{FullName: "", EmploymentType: "full_time"},
		{FullName: "Ann", EmploymentType: "bogus"},
		{FullName: "Ann", EmploymentType: "full_time", Status: "weird"},
	}
	for i, in := range cases {
		if _, err := svc.CreateEmployee(context.Background(), in); !errors.Is(err, ErrValidation) {
			t.Errorf("case %d: err = %v, want ErrValidation", i, err)
		}
	}
}

func TestListEmployeesStatusFilterValidation(t *testing.T) {
	svc, _ := newSvcWithFake()
	bad := "nope"
	if _, err := svc.ListEmployees(context.Background(), EmployeeFilter{Status: &bad}); !errors.Is(err, ErrValidation) {
		t.Fatalf("err = %v, want ErrValidation", err)
	}
}

func TestCreateRoleValidationAndSuccess(t *testing.T) {
	svc, _ := newSvcWithFake()
	if _, err := svc.CreateRole(context.Background(), CreateRoleInput{Name: "  "}); !errors.Is(err, ErrValidation) {
		t.Fatalf("empty name err = %v, want ErrValidation", err)
	}
	r, err := svc.CreateRole(context.Background(), CreateRoleInput{Name: "Manager", Permissions: []string{"roster.write"}})
	if err != nil {
		t.Fatalf("CreateRole: %v", err)
	}
	if r.Name != "Manager" || len(r.Permissions) != 1 {
		t.Errorf("role = %+v", r)
	}
}

func TestUpdateEmployeeValidation(t *testing.T) {
	svc, _ := newSvcWithFake()
	empty, bad, badStatus := "", "bogus", "weird"
	cases := []UpdateEmployeeInput{
		{FullName: &empty},
		{EmploymentType: &bad},
		{Status: &badStatus},
	}
	for i, in := range cases {
		if _, err := svc.UpdateEmployee(context.Background(), uuid.New(), in); !errors.Is(err, ErrValidation) {
			t.Errorf("case %d: err = %v, want ErrValidation", i, err)
		}
	}
}

func TestUpdateDepartmentValidation(t *testing.T) {
	svc, _ := newSvcWithFake()
	blank := "   "
	if _, err := svc.UpdateDepartment(context.Background(), uuid.New(), UpdateDepartmentInput{Name: &blank}); !errors.Is(err, ErrValidation) {
		t.Fatalf("err = %v, want ErrValidation", err)
	}
}

func TestGetEmployeeNotFoundPropagates(t *testing.T) {
	svc, _ := newSvcWithFake()
	if _, err := svc.GetEmployee(context.Background(), uuid.New()); !errors.Is(err, ErrNotFound) {
		t.Fatalf("err = %v, want ErrNotFound", err)
	}
}

// Default service (no fake) must surface ErrNoTenant when the context lacks a
// tenant pool — i.e. the route wasn't behind TenantMiddleware.
func TestNoTenantInContext(t *testing.T) {
	svc := NewService(slog.New(slog.NewTextHandler(io.Discard, nil)), &fakeUserCreator{})
	if _, err := svc.ListDepartments(context.Background()); !errors.Is(err, ErrNoTenant) {
		t.Fatalf("err = %v, want ErrNoTenant", err)
	}
}
