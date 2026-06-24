package auth

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// Claims are the custom JWT claims. Tenant tokens use subject as the tenant
// user id and include tenant_id. Platform tokens use subject as the
// platform_user id and do not select a tenant database.
type Claims struct {
	Kind     string    `json:"kind"`
	TenantID uuid.UUID `json:"tenant_id"`
	Role     string    `json:"role"`
	jwt.RegisteredClaims
}

// UserID returns the token subject parsed as a UUID.
func (c *Claims) UserID() (uuid.UUID, error) {
	return uuid.Parse(c.Subject)
}

// PlatformUserID returns the platform token subject parsed as a UUID.
func (c *Claims) PlatformUserID() (uuid.UUID, error) {
	return uuid.Parse(c.Subject)
}

// TokenManager issues and verifies HS256 JWTs.
type TokenManager struct {
	secret []byte
	issuer string
	ttl    time.Duration
}

// NewTokenManager builds a TokenManager. secret must be non-empty (enforced by
// config). ttl is the access-token lifetime.
func NewTokenManager(secret, issuer string, ttl time.Duration) *TokenManager {
	return &TokenManager{secret: []byte(secret), issuer: issuer, ttl: ttl}
}

// Issue creates a signed token for the user and returns it with its expiry.
// now is passed explicitly so callers control the clock (and tests are stable).
func (m *TokenManager) Issue(userID, tenantID uuid.UUID, role string, now time.Time) (string, time.Time, error) {
	return m.issue("tenant", userID, tenantID, role, now)
}

// IssuePlatform creates a signed token for a platform user and returns it with
// its expiry. Platform tokens never include a tenant id.
func (m *TokenManager) IssuePlatform(platformUserID uuid.UUID, role string, now time.Time) (string, time.Time, error) {
	return m.issue("platform", platformUserID, uuid.Nil, role, now)
}

func (m *TokenManager) issue(kind string, subjectID, tenantID uuid.UUID, role string, now time.Time) (string, time.Time, error) {
	expiresAt := now.Add(m.ttl)
	claims := Claims{
		Kind:     kind,
		TenantID: tenantID,
		Role:     role,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    m.issuer,
			Subject:   subjectID.String(),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(expiresAt),
		},
	}
	signed, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(m.secret)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("sign token: %w", err)
	}
	return signed, expiresAt, nil
}

// Parse verifies the token signature, method, issuer, and expiry, returning
// the claims on success.
func (m *TokenManager) Parse(tokenString string) (*Claims, error) {
	claims := &Claims{}
	_, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return m.secret, nil
	}, jwt.WithIssuer(m.issuer), jwt.WithValidMethods([]string{"HS256"}))
	if err != nil {
		return nil, fmt.Errorf("parse token: %w", err)
	}
	return claims, nil
}
