package controlplane

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/davidnguyen2205/opero/backend/internal/platform/auth"
)

// --- fakes ---

type fakeRepo struct {
	tenantsByID        map[uuid.UUID]Tenant
	tenantsBySlug      map[string]Tenant
	usersByID          map[uuid.UUID]User
	usersByTenantEmail map[string]User
}

func newFakeRepo() *fakeRepo {
	return &fakeRepo{
		tenantsByID:        map[uuid.UUID]Tenant{},
		tenantsBySlug:      map[string]Tenant{},
		usersByID:          map[uuid.UUID]User{},
		usersByTenantEmail: map[string]User{},
	}
}

func userKey(tenantID uuid.UUID, email string) string {
	return tenantID.String() + "|" + strings.ToLower(email)
}

func (f *fakeRepo) CreateTenant(_ context.Context, name, slug, _, plan string) (Tenant, error) {
	if _, ok := f.tenantsBySlug[slug]; ok {
		return Tenant{}, ErrConflict
	}
	t := Tenant{ID: uuid.New(), Name: name, Slug: slug, Status: "provisioning", Plan: plan}
	f.tenantsByID[t.ID] = t
	f.tenantsBySlug[slug] = t
	return t, nil
}

func (f *fakeRepo) GetTenantByID(_ context.Context, id uuid.UUID) (Tenant, error) {
	if t, ok := f.tenantsByID[id]; ok {
		return t, nil
	}
	return Tenant{}, ErrNotFound
}

func (f *fakeRepo) GetTenantBySlug(_ context.Context, slug string) (Tenant, error) {
	if t, ok := f.tenantsBySlug[slug]; ok {
		return t, nil
	}
	return Tenant{}, ErrNotFound
}

func (f *fakeRepo) SetTenantStatus(_ context.Context, id uuid.UUID, status string) (Tenant, error) {
	t, ok := f.tenantsByID[id]
	if !ok {
		return Tenant{}, ErrNotFound
	}
	t.Status = status
	f.tenantsByID[id] = t
	f.tenantsBySlug[t.Slug] = t
	return t, nil
}

func (f *fakeRepo) DeleteTenant(_ context.Context, id uuid.UUID) error {
	if t, ok := f.tenantsByID[id]; ok {
		delete(f.tenantsBySlug, t.Slug)
		delete(f.tenantsByID, id)
	}
	return nil
}

func (f *fakeRepo) CreateUser(_ context.Context, tenantID uuid.UUID, email, hash, role, status string) (User, error) {
	key := userKey(tenantID, email)
	if _, ok := f.usersByTenantEmail[key]; ok {
		return User{}, ErrConflict
	}
	u := User{ID: uuid.New(), TenantID: tenantID, Email: email, Role: role, Status: status, PasswordHash: hash}
	f.usersByID[u.ID] = u
	f.usersByTenantEmail[key] = u
	return u, nil
}

func (f *fakeRepo) DeleteUser(_ context.Context, id uuid.UUID) error {
	if u, ok := f.usersByID[id]; ok {
		delete(f.usersByID, id)
		delete(f.usersByTenantEmail, userKey(u.TenantID, u.Email))
	}
	return nil
}

func (f *fakeRepo) GetUserByID(_ context.Context, id uuid.UUID) (User, error) {
	if u, ok := f.usersByID[id]; ok {
		return u, nil
	}
	return User{}, ErrNotFound
}

func (f *fakeRepo) GetUserByTenantAndEmail(_ context.Context, tenantID uuid.UUID, email string) (User, error) {
	if u, ok := f.usersByTenantEmail[userKey(tenantID, email)]; ok {
		return u, nil
	}
	return User{}, ErrNotFound
}

type fakeProvisioner struct {
	created   map[string]bool
	createErr error
}

func (f *fakeProvisioner) Create(_ context.Context, dbName string) error {
	if f.createErr != nil {
		return f.createErr
	}
	if f.created == nil {
		f.created = map[string]bool{}
	}
	f.created[dbName] = true
	return nil
}

func (f *fakeProvisioner) Drop(_ context.Context, dbName string) error {
	delete(f.created, dbName)
	return nil
}

func newTestService() (*Service, *fakeRepo, *fakeProvisioner) {
	repo := newFakeRepo()
	prov := &fakeProvisioner{}
	tokens := auth.NewTokenManager("test-secret", "opero", time.Hour)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	return NewService(repo, prov, tokens, "opero_tenant_", logger), repo, prov
}

func signupInput() SignupInput {
	return SignupInput{
		CompanyName:   "Acme Travel",
		AdminFullName: "Ada Admin",
		AdminEmail:    "ada@acme.test",
		AdminPassword: "password1",
	}
}

// --- tests ---

func TestSignupSuccess(t *testing.T) {
	svc, _, prov := newTestService()
	res, err := svc.Signup(context.Background(), signupInput())
	if err != nil {
		t.Fatalf("Signup: %v", err)
	}
	if res.Token == "" {
		t.Error("expected a token")
	}
	if res.Tenant.Slug != "acme-travel" {
		t.Errorf("slug = %q, want acme-travel", res.Tenant.Slug)
	}
	if res.Tenant.Status != "active" {
		t.Errorf("status = %q, want active", res.Tenant.Status)
	}
	if res.User.Role != "admin" {
		t.Errorf("role = %q, want admin", res.User.Role)
	}
	if !prov.created["opero_tenant_acme_travel"] {
		t.Error("expected tenant database to be provisioned")
	}
}

func TestSignupDuplicateSlugIsConflict(t *testing.T) {
	svc, _, _ := newTestService()
	if _, err := svc.Signup(context.Background(), signupInput()); err != nil {
		t.Fatalf("first signup: %v", err)
	}
	dup := signupInput()
	dup.AdminEmail = "other@acme.test"
	if _, err := svc.Signup(context.Background(), dup); !errors.Is(err, ErrConflict) {
		t.Fatalf("err = %v, want ErrConflict", err)
	}
}

func TestSignupValidation(t *testing.T) {
	svc, _, _ := newTestService()
	bad := signupInput()
	bad.CompanyName = ""
	bad.Slug = ""
	if _, err := svc.Signup(context.Background(), bad); !errors.Is(err, ErrValidation) {
		t.Errorf("empty company: err = %v, want ErrValidation", err)
	}
	short := signupInput()
	short.AdminPassword = "short"
	if _, err := svc.Signup(context.Background(), short); !errors.Is(err, ErrValidation) {
		t.Errorf("short password: err = %v, want ErrValidation", err)
	}
}

func TestSignupProvisionFailureCleansUp(t *testing.T) {
	svc, repo, _ := newTestService()
	svc.prov = &fakeProvisioner{createErr: errors.New("disk full")}
	if _, err := svc.Signup(context.Background(), signupInput()); err == nil {
		t.Fatal("expected error when provisioning fails")
	}
	if len(repo.tenantsBySlug) != 0 {
		t.Errorf("expected tenant row cleaned up, found %d", len(repo.tenantsBySlug))
	}
}

func TestLogin(t *testing.T) {
	svc, _, _ := newTestService()
	if _, err := svc.Signup(context.Background(), signupInput()); err != nil {
		t.Fatalf("signup: %v", err)
	}

	ok, err := svc.Login(context.Background(), LoginInput{TenantSlug: "acme-travel", Email: "ada@acme.test", Password: "password1"})
	if err != nil {
		t.Fatalf("login: %v", err)
	}
	if ok.Token == "" {
		t.Error("expected token on successful login")
	}

	cases := []LoginInput{
		{TenantSlug: "acme-travel", Email: "ada@acme.test", Password: "wrong"},
		{TenantSlug: "acme-travel", Email: "nobody@acme.test", Password: "password1"},
		{TenantSlug: "no-such-tenant", Email: "ada@acme.test", Password: "password1"},
	}
	for _, in := range cases {
		if _, err := svc.Login(context.Background(), in); !errors.Is(err, ErrInvalidCredentials) {
			t.Errorf("Login(%+v) err = %v, want ErrInvalidCredentials", in, err)
		}
	}
}

func TestCurrentUser(t *testing.T) {
	svc, _, _ := newTestService()
	res, err := svc.Signup(context.Background(), signupInput())
	if err != nil {
		t.Fatalf("signup: %v", err)
	}
	cu, err := svc.CurrentUser(context.Background(), res.User.ID)
	if err != nil {
		t.Fatalf("CurrentUser: %v", err)
	}
	if cu.User.Email != "ada@acme.test" {
		t.Errorf("email = %q", cu.User.Email)
	}
	if cu.Tenant.Slug != "acme-travel" {
		t.Errorf("tenant slug = %q", cu.Tenant.Slug)
	}

	if _, err := svc.CurrentUser(context.Background(), uuid.New()); !errors.Is(err, ErrNotFound) {
		t.Errorf("unknown user err = %v, want ErrNotFound", err)
	}
}
