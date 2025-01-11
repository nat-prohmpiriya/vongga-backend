package auth

import (
	"context"
	"strings"

	"vongga-api/internal/domain"

	firebase "firebase.google.com/go/v4/auth"
)

type FirebaseAuthAdapter struct {
	client *firebase.Client
}

func NewFirebaseAuthAdapter(client *firebase.Client) domain.AuthClient {
	return &FirebaseAuthAdapter{
		client: client,
	}
}

func (a *FirebaseAuthAdapter) VerifyToken(token string) (*domain.Claims, error) {
	// Remove "Bearer " prefix if present
	token = strings.TrimPrefix(token, "Bearer ")

	// Verify the Firebase token
	tokenData, err := a.client.VerifyIDToken(context.Background(), token)
	if err != nil {
		return nil, err
	}

	// Extract user ID from token claims
	userID := tokenData.UID

	return &domain.Claims{
		UserID: userID,
	}, nil
}
