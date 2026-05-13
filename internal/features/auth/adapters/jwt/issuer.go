package authjwt

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// Issuer signs JWT access tokens (HS256).
type Issuer struct {
	secret []byte
	ttl    time.Duration
}

// NewIssuer creates a JWT issuer with secret and TTL.
func NewIssuer(secret string, ttl time.Duration) *Issuer {
	return &Issuer{
		secret: []byte(secret),
		ttl:    ttl,
	}
}

// Generate emits a JWT with subject = user ID.
func (i *Issuer) Generate(userID uuid.UUID) (string, error) {
	now := time.Now()
	claims := jwt.RegisteredClaims{
		Subject:   userID.String(),
		IssuedAt:  jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(now.Add(i.ttl)),
	}
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, &claims)
	signed, err := t.SignedString(i.secret)
	if err != nil {
		return "", fmt.Errorf("sign jwt: %w", err)
	}
	return signed, nil
}
