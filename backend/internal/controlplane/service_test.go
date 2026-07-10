package controlplane

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"maps"
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
	platformUsersByID  map[uuid.UUID]PlatformUser
	platformUsersEmail map[string]PlatformUser
	subscriptionsByID  map[uuid.UUID]PlatformSubscription
	auditEvents        []SuperAdminAuditEvent
	// auditErr, when set, makes CreateSuperAdminAuditEvent fail — used to test
	// that a mutation rolls back when its audit write fails.
	auditErr error
}

func newFakeRepo() *fakeRepo {
	return &fakeRepo{
		tenantsByID:        map[uuid.UUID]Tenant{},
		tenantsBySlug:      map[string]Tenant{},
		usersByID:          map[uuid.UUID]User{},
		usersByTenantEmail: map[string]User{},
		platformUsersByID:  map[uuid.UUID]PlatformUser{},
		platformUsersEmail: map[string]PlatformUser{},
		subscriptionsByID:  map[uuid.UUID]PlatformSubscription{},
	}
}

// InTx emulates a transaction: it snapshots the mutable state the platform
// mutations touch, runs fn, and restores the snapshot if fn returns an error —
// so a failed audit write leaves no partial state, mirroring the real store.
func (f *fakeRepo) InTx(_ context.Context, fn func(repo) error) error {
	tenants := maps.Clone(f.tenantsByID)
	tenantsSlug := maps.Clone(f.tenantsBySlug)
	users := maps.Clone(f.usersByID)
	usersEmail := maps.Clone(f.usersByTenantEmail)
	subs := maps.Clone(f.subscriptionsByID)
	audit := append([]SuperAdminAuditEvent(nil), f.auditEvents...)
	if err := fn(f); err != nil {
		f.tenantsByID = tenants
		f.tenantsBySlug = tenantsSlug
		f.usersByID = users
		f.usersByTenantEmail = usersEmail
		f.subscriptionsByID = subs
		f.auditEvents = audit
		return err
	}
	return nil
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

func (f *fakeRepo) ListTenants(context.Context) ([]Tenant, error) {
	out := make([]Tenant, 0, len(f.tenantsByID))
	for _, tenant := range f.tenantsByID {
		out = append(out, tenant)
	}
	return out, nil
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

func (f *fakeRepo) UpdateTenantPlatform(_ context.Context, id uuid.UUID, name, status, plan *string) (Tenant, error) {
	t, ok := f.tenantsByID[id]
	if !ok {
		return Tenant{}, ErrNotFound
	}
	if name != nil {
		t.Name = *name
	}
	if status != nil {
		t.Status = *status
	}
	if plan != nil {
		t.Plan = *plan
	}
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

func (f *fakeRepo) ListUsersPlatform(_ context.Context, tenantID *uuid.UUID, role, status *string) ([]PlatformTenantUser, error) {
	var out []PlatformTenantUser
	for _, user := range f.usersByID {
		if tenantID != nil && user.TenantID != *tenantID {
			continue
		}
		if role != nil && user.Role != *role {
			continue
		}
		if status != nil && user.Status != *status {
			continue
		}
		tenant := f.tenantsByID[user.TenantID]
		out = append(out, PlatformTenantUser{
			ID: user.ID, TenantID: user.TenantID, TenantName: tenant.Name,
			TenantSlug: tenant.Slug, Email: user.Email, Role: user.Role, Status: user.Status,
		})
	}
	return out, nil
}

func (f *fakeRepo) UpdateUserStatusPlatform(_ context.Context, id uuid.UUID, status string) (User, error) {
	user, ok := f.usersByID[id]
	if !ok {
		return User{}, ErrNotFound
	}
	user.Status = status
	f.usersByID[id] = user
	f.usersByTenantEmail[userKey(user.TenantID, user.Email)] = user
	return user, nil
}

func (f *fakeRepo) GetUserByTenantAndEmail(_ context.Context, tenantID uuid.UUID, email string) (User, error) {
	if u, ok := f.usersByTenantEmail[userKey(tenantID, email)]; ok {
		return u, nil
	}
	return User{}, ErrNotFound
}

func (f *fakeRepo) CreatePlatformUser(_ context.Context, email, hash, role, status string) (PlatformUser, error) {
	key := strings.ToLower(email)
	if _, ok := f.platformUsersEmail[key]; ok {
		return PlatformUser{}, ErrConflict
	}
	user := PlatformUser{ID: uuid.New(), Email: email, Role: role, Status: status, PasswordHash: hash}
	f.platformUsersByID[user.ID] = user
	f.platformUsersEmail[key] = user
	return user, nil
}

func (f *fakeRepo) GetPlatformUserByID(_ context.Context, id uuid.UUID) (PlatformUser, error) {
	if user, ok := f.platformUsersByID[id]; ok {
		return user, nil
	}
	return PlatformUser{}, ErrNotFound
}

func (f *fakeRepo) GetPlatformUserByEmail(_ context.Context, email string) (PlatformUser, error) {
	if user, ok := f.platformUsersEmail[strings.ToLower(email)]; ok {
		return user, nil
	}
	return PlatformUser{}, ErrNotFound
}

func (f *fakeRepo) ListSubscriptionsPlatform(_ context.Context, tenantID *uuid.UUID, plan, status *string) ([]PlatformSubscription, error) {
	var out []PlatformSubscription
	for _, sub := range f.subscriptionsByID {
		if tenantID != nil && sub.TenantID != *tenantID {
			continue
		}
		if plan != nil && sub.Plan != *plan {
			continue
		}
		if status != nil && sub.Status != *status {
			continue
		}
		out = append(out, sub)
	}
	return out, nil
}

func (f *fakeRepo) UpdateSubscriptionPlatform(_ context.Context, id uuid.UUID, plan, status *string) (PlatformSubscription, error) {
	sub, ok := f.subscriptionsByID[id]
	if !ok {
		return PlatformSubscription{}, ErrNotFound
	}
	if plan != nil {
		sub.Plan = *plan
	}
	if status != nil {
		sub.Status = *status
	}
	f.subscriptionsByID[id] = sub
	return sub, nil
}

func (f *fakeRepo) CreateSuperAdminAuditEvent(_ context.Context, actorID uuid.UUID, action, targetType string, targetID, tenantID *uuid.UUID, metadata map[string]any) error {
	if f.auditErr != nil {
		return f.auditErr
	}
	f.auditEvents = append(f.auditEvents, SuperAdminAuditEvent{
		ID: uuid.New(), ActorPlatformUserID: actorID, Action: action,
		TargetType: targetType, TargetID: targetID, TenantID: tenantID, Metadata: metadata,
	})
	return nil
}

func (f *fakeRepo) ListSuperAdminAuditEvents(context.Context, *uuid.UUID, *uuid.UUID, *string, int32) ([]SuperAdminAuditEvent, error) {
	return f.auditEvents, nil
}

func (f *fakeRepo) CountTenantsByStatus(context.Context) (map[string]int, error) {
	counts := map[string]int{}
	for _, tenant := range f.tenantsByID {
		counts[tenant.Status]++
	}
	return counts, nil
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

func TestPlatformLogin(t *testing.T) {
	svc, repo, _ := newTestService()
	id, err := svc.CreatePlatformUser(context.Background(), "ops@opero.test", "long-password", "super_admin")
	if err != nil {
		t.Fatalf("CreatePlatformUser: %v", err)
	}

	res, err := svc.PlatformLogin(context.Background(), PlatformLoginInput{
		Email:    "ops@opero.test",
		Password: "long-password",
	})
	if err != nil {
		t.Fatalf("PlatformLogin: %v", err)
	}
	if res.Token == "" {
		t.Fatal("expected platform token")
	}
	if res.User.ID != id {
		t.Fatalf("user id = %v, want %v", res.User.ID, id)
	}

	user := repo.platformUsersByID[id]
	user.Status = "disabled"
	repo.platformUsersByID[id] = user
	repo.platformUsersEmail[strings.ToLower(user.Email)] = user
	if _, err := svc.PlatformLogin(context.Background(), PlatformLoginInput{Email: user.Email, Password: "long-password"}); !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf("disabled platform user err = %v, want ErrInvalidCredentials", err)
	}
}

func TestPlatformUpdateTenantWritesAudit(t *testing.T) {
	svc, repo, _ := newTestService()
	res, err := svc.Signup(context.Background(), signupInput())
	if err != nil {
		t.Fatalf("Signup: %v", err)
	}
	actorID, err := svc.CreatePlatformUser(context.Background(), "ops@opero.test", "long-password", "ops")
	if err != nil {
		t.Fatalf("CreatePlatformUser: %v", err)
	}

	status := "suspended"
	tenant, err := svc.PlatformUpdateTenant(context.Background(), actorID, res.Tenant.ID, nil, &status, nil)
	if err != nil {
		t.Fatalf("PlatformUpdateTenant: %v", err)
	}
	if tenant.Status != "suspended" {
		t.Fatalf("status = %q, want suspended", tenant.Status)
	}
	if len(repo.auditEvents) != 1 {
		t.Fatalf("audit event count = %d, want 1", len(repo.auditEvents))
	}
	if repo.auditEvents[0].Action != "tenant.updated" {
		t.Fatalf("audit action = %q", repo.auditEvents[0].Action)
	}
}

// seedPlatformActor creates a platform user and returns its id for use as an
// audit actor.
func seedPlatformActor(t *testing.T, svc *Service) uuid.UUID {
	t.Helper()
	actorID, err := svc.CreatePlatformUser(context.Background(), "ops@opero.test", "long-password", "ops")
	if err != nil {
		t.Fatalf("CreatePlatformUser: %v", err)
	}
	return actorID
}

func TestPlatformUpdateUserWritesAudit(t *testing.T) {
	svc, repo, _ := newTestService()
	res, err := svc.Signup(context.Background(), signupInput())
	if err != nil {
		t.Fatalf("Signup: %v", err)
	}
	actorID := seedPlatformActor(t, svc)

	user, err := svc.PlatformUpdateUser(context.Background(), actorID, res.User.ID, "disabled")
	if err != nil {
		t.Fatalf("PlatformUpdateUser: %v", err)
	}
	if user.Status != "disabled" {
		t.Fatalf("status = %q, want disabled", user.Status)
	}
	if len(repo.auditEvents) != 1 || repo.auditEvents[0].Action != "user.status_updated" {
		t.Fatalf("audit events = %+v, want one user.status_updated", repo.auditEvents)
	}
	if repo.auditEvents[0].Metadata["status"] != "disabled" {
		t.Fatalf("audit metadata = %+v, want status=disabled", repo.auditEvents[0].Metadata)
	}
}

func TestPlatformUpdateUserRejectsInvalidStatus(t *testing.T) {
	svc, repo, _ := newTestService()
	res, err := svc.Signup(context.Background(), signupInput())
	if err != nil {
		t.Fatalf("Signup: %v", err)
	}
	actorID := seedPlatformActor(t, svc)

	if _, err := svc.PlatformUpdateUser(context.Background(), actorID, res.User.ID, "banished"); !errors.Is(err, ErrValidation) {
		t.Fatalf("err = %v, want ErrValidation", err)
	}
	if len(repo.auditEvents) != 0 {
		t.Fatalf("no audit event should be written on validation failure, got %d", len(repo.auditEvents))
	}
}

func TestPlatformUpdateUserRollsBackWhenAuditFails(t *testing.T) {
	svc, repo, _ := newTestService()
	res, err := svc.Signup(context.Background(), signupInput())
	if err != nil {
		t.Fatalf("Signup: %v", err)
	}
	actorID := seedPlatformActor(t, svc)
	repo.auditErr = errors.New("audit insert failed")

	if _, err := svc.PlatformUpdateUser(context.Background(), actorID, res.User.ID, "disabled"); err == nil {
		t.Fatal("expected error when audit write fails")
	}
	// The status change must have rolled back with the failed audit write.
	got, err := repo.GetUserByID(context.Background(), res.User.ID)
	if err != nil {
		t.Fatalf("GetUserByID: %v", err)
	}
	if got.Status == "disabled" {
		t.Fatal("user status persisted despite audit failure — mutation was not atomic")
	}
	if len(repo.auditEvents) != 0 {
		t.Fatalf("audit events = %d, want 0 after rollback", len(repo.auditEvents))
	}
}

func TestPlatformUpdateSubscriptionWritesAudit(t *testing.T) {
	svc, repo, _ := newTestService()
	res, err := svc.Signup(context.Background(), signupInput())
	if err != nil {
		t.Fatalf("Signup: %v", err)
	}
	actorID := seedPlatformActor(t, svc)
	subID := uuid.New()
	repo.subscriptionsByID[subID] = PlatformSubscription{
		ID: subID, TenantID: res.Tenant.ID, Plan: "free", Status: "active",
	}

	plan := "pro"
	sub, err := svc.PlatformUpdateSubscription(context.Background(), actorID, subID, &plan, nil)
	if err != nil {
		t.Fatalf("PlatformUpdateSubscription: %v", err)
	}
	if sub.Plan != "pro" {
		t.Fatalf("plan = %q, want pro", sub.Plan)
	}
	if sub.TenantName != res.Tenant.Name {
		t.Fatalf("tenant name = %q, want %q", sub.TenantName, res.Tenant.Name)
	}
	if len(repo.auditEvents) != 1 || repo.auditEvents[0].Action != "subscription.updated" {
		t.Fatalf("audit events = %+v, want one subscription.updated", repo.auditEvents)
	}
}

func TestPlatformUpdateSubscriptionRequiresAField(t *testing.T) {
	svc, _, _ := newTestService()
	actorID := seedPlatformActor(t, svc)
	if _, err := svc.PlatformUpdateSubscription(context.Background(), actorID, uuid.New(), nil, nil); !errors.Is(err, ErrValidation) {
		t.Fatalf("err = %v, want ErrValidation", err)
	}
}
