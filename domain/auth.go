package domain

import "context"

type TokenPair struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
}

type Claims struct {
	UserID string `json:"user_id"`
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
