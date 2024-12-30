package domain

import (
	"context"
	"github.com/golang-jwt/jwt/v5"
)

type TokenPair struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
}

type Claims struct {
	UserID string `json:"userId"`
	jwt.RegisteredClaims
}

type AuthClient interface {
	VerifyToken(token string) (*Claims, error)
}

type AuthUseCase interface {
	VerifyTokenFirebase(ctx context.Context, firebaseToken string) (*User, *TokenPair, error)
	RefreshToken(ctx context.Context, refreshToken string) (*TokenPair, error)
	RevokeRefreshToken(ctx context.Context, refreshToken string) error
	CreateTestToken(ctx context.Context, userID string) (*TokenPair, error)
}
