package auth

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// ErrInvalidToken is returned for any malformed, expired, or untrusted token.
var ErrInvalidToken = errors.New("invalid token")

// Claims is the JWT payload identifying an authenticated admin. Sudo and RoleID
// are embedded so the RBAC middleware can authorize without a DB round-trip on
// every request (the token is the trust boundary).
type Claims struct {
	AdminID uuid.UUID  `json:"aid"`
	Sudo    bool       `json:"sudo"`
	RoleID  *uuid.UUID `json:"rid,omitempty"`
	jwt.RegisteredClaims
}

// Issuer signs and verifies admin tokens with a shared HMAC secret.
type Issuer struct {
	secret []byte
	ttl    time.Duration
}

// NewIssuer builds an Issuer. The secret must be kept private; rotating it
// invalidates every outstanding token.
func NewIssuer(secret []byte, ttl time.Duration) *Issuer {
	return &Issuer{secret: secret, ttl: ttl}
}

// Issue mints a signed token for an admin.
func (i *Issuer) Issue(adminID uuid.UUID, sudo bool, roleID *uuid.UUID) (string, error) {
	now := time.Now()
	claims := Claims{
		AdminID: adminID,
		Sudo:    sudo,
		RoleID:  roleID,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   adminID.String(),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(i.ttl)),
		},
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(i.secret)
}

// Verify parses and validates a token, returning its claims. It pins the signing
// method to HS256 to defend against algorithm-confusion attacks.
func (i *Issuer) Verify(token string) (*Claims, error) {
	parsed, err := jwt.ParseWithClaims(token, &Claims{}, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("%w: unexpected signing method %v", ErrInvalidToken, t.Header["alg"])
		}
		return i.secret, nil
	})
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidToken, err)
	}
	claims, ok := parsed.Claims.(*Claims)
	if !ok || !parsed.Valid {
		return nil, ErrInvalidToken
	}
	return claims, nil
}
