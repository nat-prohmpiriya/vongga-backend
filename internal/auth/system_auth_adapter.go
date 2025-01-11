package auth

import (
	"errors"

	"vongga-api/internal/domain"

	"github.com/golang-jwt/jwt/v5"
)

type SystemAuthAdapter struct {
	jwtSecret string
}

func NewSystemAuthAdapter(jwtSecret string) domain.AuthClient {
	return &SystemAuthAdapter{
		jwtSecret: jwtSecret,
	}
}

func (a *SystemAuthAdapter) VerifyToken(token string) (*domain.Claims, error) {
	// Parse token
	claims := &domain.Claims{}
	parsedToken, err := jwt.ParseWithClaims(token, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(a.jwtSecret), nil
	})
	if err != nil {
		return nil, err
	}

	// Check if token is valid
	if !parsedToken.Valid {
		return nil, errors.New("invalid token")
	}

	return claims, nil
}
