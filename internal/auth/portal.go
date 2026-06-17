package auth

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// PortalClaims are the JWT claims for end-user portal sessions (separate from
// admin Claims). They carry a user ID instead of an admin ID.
type PortalClaims struct {
	UserID   uuid.UUID `json:"user_id"`
	Username string    `json:"username"`
	jwt.RegisteredClaims
}

// IssuePortal generates a portal-scoped JWT for an end-user.
func (iss *Issuer) IssuePortal(userID uuid.UUID, username string, ttl time.Duration) (string, error) {
	now := time.Now()
	claims := PortalClaims{
		UserID:   userID,
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "vortexui-portal",
			Subject:   userID.String(),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(iss.secret)
}

// VerifyPortal validates a portal JWT and returns its claims.
func (iss *Issuer) VerifyPortal(tokenStr string) (*PortalClaims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &PortalClaims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return iss.secret, nil
	})
	if err != nil {
		return nil, err
	}
	claims, ok := token.Claims.(*PortalClaims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid portal token")
	}
	return claims, nil
}
