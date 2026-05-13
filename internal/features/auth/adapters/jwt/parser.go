package authjwt

import (
	"fmt"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// Parser validates HS256 JWTs and extracts the user ID.
type Parser struct {
	secret []byte
}

// NewParser creates a JWT parser for the given secret.
func NewParser(secret string) *Parser {
	return &Parser{secret: []byte(secret)}
}

// Parse validates the token string and returns the user ID.
func (p *Parser) Parse(tokenString string) (uuid.UUID, error) {
	token, err := jwt.ParseWithClaims(tokenString, &jwt.RegisteredClaims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return p.secret, nil
	})
	if err != nil {
		return uuid.Nil, fmt.Errorf("parse jwt: %w", err)
	}
	claims, ok := token.Claims.(*jwt.RegisteredClaims)
	if !ok || !token.Valid {
		return uuid.Nil, fmt.Errorf("invalid jwt claims")
	}
	id, err := uuid.Parse(claims.Subject)
	if err != nil {
		return uuid.Nil, fmt.Errorf("parse subject uuid: %w", err)
	}
	return id, nil
}
