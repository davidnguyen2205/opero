package auth

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestPasswordRoundTrip(t *testing.T) {
	hash, err := HashPassword("s3cret-pw")
	if err != nil {
		t.Fatalf("HashPassword: %v", err)
	}
	if !CheckPassword(hash, "s3cret-pw") {
		t.Error("CheckPassword returned false for correct password")
	}
	if CheckPassword(hash, "wrong") {
		t.Error("CheckPassword returned true for wrong password")
	}
}

func TestTokenIssueParse(t *testing.T) {
	tm := NewTokenManager("a-secret", "opero", time.Hour)
	userID, tenantID := uuid.New(), uuid.New()
	now := time.Now()

	tok, exp, err := tm.Issue(userID, tenantID, "admin", now)
	if err != nil {
		t.Fatalf("Issue: %v", err)
	}
	if !exp.After(now) {
		t.Errorf("expiry %v not after now %v", exp, now)
	}

	claims, err := tm.Parse(tok)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if claims.TenantID != tenantID {
		t.Errorf("TenantID = %v, want %v", claims.TenantID, tenantID)
	}
	if claims.Role != "admin" {
		t.Errorf("Role = %q, want admin", claims.Role)
	}
	gotUser, err := claims.UserID()
	if err != nil {
		t.Fatalf("UserID: %v", err)
	}
	if gotUser != userID {
		t.Errorf("UserID = %v, want %v", gotUser, userID)
	}
}

func TestParseRejectsWrongSecret(t *testing.T) {
	issuer := NewTokenManager("secret-a", "opero", time.Hour)
	verifier := NewTokenManager("secret-b", "opero", time.Hour)
	tok, _, err := issuer.Issue(uuid.New(), uuid.New(), "admin", time.Now())
	if err != nil {
		t.Fatalf("Issue: %v", err)
	}
	if _, err := verifier.Parse(tok); err == nil {
		t.Fatal("expected parse to fail with wrong secret")
	}
}

func TestParseRejectsExpired(t *testing.T) {
	tm := NewTokenManager("secret", "opero", time.Hour)
	tok, _, err := tm.Issue(uuid.New(), uuid.New(), "admin", time.Now().Add(-2*time.Hour))
	if err != nil {
		t.Fatalf("Issue: %v", err)
	}
	if _, err := tm.Parse(tok); err == nil {
		t.Fatal("expected parse to fail for expired token")
	}
}
