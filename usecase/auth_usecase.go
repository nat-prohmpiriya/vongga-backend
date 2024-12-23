package usecase

import (
	"context"
	"crypto/rand"
	"encoding/base64"

	// "errors"
	"fmt"
	"time"

	"firebase.google.com/go/v4/auth"
	"github.com/golang-jwt/jwt/v5"
	"github.com/prohmpiriya_phonumnuaisuk/vongga-platform/vongga-backend/domain"
	"github.com/redis/go-redis/v9"
)

type authUseCase struct {
	userRepo           domain.UserRepository
	authClient         *auth.Client
	redisClient        *redis.Client
	jwtSecret          string
	refreshTokenSecret string
	tokenExpiry        time.Duration
	refreshTokenExpiry time.Duration
}

func NewAuthUseCase(
	userRepo domain.UserRepository,
	authClient *auth.Client,
	redisClient *redis.Client,
	jwtSecret string,
	refreshTokenSecret string,
	tokenExpiry time.Duration,
	refreshTokenExpiry time.Duration,
) domain.AuthUseCase {
	return &authUseCase{
		userRepo:           userRepo,
		authClient:         authClient,
		redisClient:        redisClient,
		jwtSecret:          jwtSecret,
		refreshTokenSecret: refreshTokenSecret,
		tokenExpiry:        tokenExpiry,
		refreshTokenExpiry: refreshTokenExpiry,
	}
}

func (u *authUseCase) Login(ctx context.Context, firebaseToken string) (*domain.User, *domain.TokenPair, error) {
	// Verify Firebase token
	token, err := u.authClient.VerifyIDToken(ctx, firebaseToken)
	if err != nil {
		return nil, nil, err
	}

	// Get or create user
	user, err := u.userRepo.FindByFirebaseUID(token.UID)
	if err != nil {
		return nil, nil, err
	}

	if user == nil {
		// Get user info from Firebase
		firebaseUser, err := u.authClient.GetUser(ctx, token.UID)
		if err != nil {
			return nil, nil, err
		}

		// Create new user
		user = &domain.User{
			FirebaseUID: token.UID,
			Email:       firebaseUser.Email,
			FirstName:   firebaseUser.DisplayName,
			PhotoURL:    firebaseUser.PhotoURL,
			Provider:    getProviderFromFirebase(firebaseUser.ProviderID),
		}

		err = u.userRepo.Create(user)
		if err != nil {
			return nil, nil, err
		}
	}

	// Generate token pair
	tokenPair, err := u.generateTokenPair(ctx, user.ID.Hex())
	if err != nil {
		return nil, nil, err
	}

	return user, tokenPair, nil
}

func (u *authUseCase) RefreshToken(ctx context.Context, refreshToken string) (*domain.TokenPair, error) {
	// Verify refresh token
	token, err := jwt.Parse(refreshToken, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(u.refreshTokenSecret), nil
	})

	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid refresh token")
	}

	// Check if refresh token is in Redis
	userID := claims["user_id"].(string)
	key := fmt.Sprintf("refresh_token:%s:%s", userID, refreshToken)
	exists, err := u.redisClient.Exists(ctx, key).Result()
	if err != nil {
		return nil, err
	}
	if exists == 0 {
		return nil, fmt.Errorf("refresh token has been revoked")
	}

	// Generate new token pair
	return u.generateTokenPair(ctx, userID)
}

func (u *authUseCase) RevokeRefreshToken(ctx context.Context, refreshToken string) error {
	// Verify refresh token
	token, err := jwt.Parse(refreshToken, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(u.refreshTokenSecret), nil
	})

	if err != nil {
		return err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return fmt.Errorf("invalid refresh token")
	}

	// Remove refresh token from Redis
	userID := claims["user_id"].(string)
	key := fmt.Sprintf("refresh_token:%s:%s", userID, refreshToken)
	return u.redisClient.Del(ctx, key).Err()
}

func (u *authUseCase) generateTokenPair(ctx context.Context, userID string) (*domain.TokenPair, error) {
	// Generate access token
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(u.tokenExpiry).Unix(),
		"type":    "access",
	})

	accessTokenString, err := accessToken.SignedString([]byte(u.jwtSecret))
	if err != nil {
		return nil, err
	}

	// Generate refresh token
	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(u.refreshTokenExpiry).Unix(),
		"type":    "refresh",
		"jti":     generateRandomString(32),
	})

	refreshTokenString, err := refreshToken.SignedString([]byte(u.refreshTokenSecret))
	if err != nil {
		return nil, err
	}

	// Store refresh token in Redis
	key := fmt.Sprintf("refresh_token:%s:%s", userID, refreshTokenString)
	err = u.redisClient.Set(ctx, key, "valid", u.refreshTokenExpiry).Err()
	if err != nil {
		return nil, err
	}

	return &domain.TokenPair{
		AccessToken:  accessTokenString,
		RefreshToken: refreshTokenString,
	}, nil
}

func generateRandomString(n int) string {
	b := make([]byte, n)
	rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)
}

func getProviderFromFirebase(providerID string) domain.AuthProvider {
	switch providerID {
	case "google.com":
		return domain.Google
	case "apple.com":
		return domain.Apple
	default:
		return domain.Email
	}
}
