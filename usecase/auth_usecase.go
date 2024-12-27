package usecase

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"time"

	"firebase.google.com/go/v4/auth"
	"github.com/golang-jwt/jwt/v5"
	"github.com/prohmpiriya_phonumnuaisuk/vongga-platform/vongga-backend/domain"
	"github.com/prohmpiriya_phonumnuaisuk/vongga-platform/vongga-backend/utils"
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

func (u *authUseCase) VerifyTokenFirebase(ctx context.Context, firebaseToken string) (*domain.User, *domain.TokenPair, error) {
	logger := utils.NewLogger("AuthUseCase.VerifyTokenFirebase")
	logger.LogInput(map[string]string{
		"firebaseToken": firebaseToken,
	})

	// Verify Firebase token
	token, err := u.authClient.VerifyIDToken(ctx, firebaseToken)
	if err != nil {
		logger.LogOutput(nil, fmt.Errorf("invalid firebase token: %v", err))
		return nil, nil, fmt.Errorf("invalid firebase token: %v", err)
	}

	// Get or create user
	user, err := u.userRepo.FindByFirebaseUID(token.UID)
	if err != nil {
		logger.LogOutput(nil, fmt.Errorf("error finding user: %v", err))
		return nil, nil, fmt.Errorf("error finding user: %v", err)
	}

	if user == nil {
		// Get user info from Firebase
		firebaseUser, err := u.authClient.GetUser(ctx, token.UID)
		if err != nil {
			logger.LogOutput(nil, fmt.Errorf("error getting firebase user: %v", err))
			return nil, nil, fmt.Errorf("error getting firebase user: %v", err)
		}
		logger.LogInput("user from firebase", firebaseUser)

		// Create new user
		user = &domain.User{
			FirebaseUID: token.UID,
			Email:       firebaseUser.Email,
			Provider:    getProviderFromFirebase(firebaseUser.ProviderUserInfo[0].ProviderID),
		}

		err = u.userRepo.Create(user)
		if err != nil {
			logger.LogOutput(nil, fmt.Errorf("error creating user: %v", err))
			return nil, nil, fmt.Errorf("error creating user: %v", err)
		}

		// Get the created user from database to get the generated ID
		user, err = u.userRepo.FindByFirebaseUID(token.UID)
		if err != nil {
			logger.LogOutput(nil, fmt.Errorf("error getting created user: %v", err))
			return nil, nil, fmt.Errorf("error getting created user: %v", err)
		}
	}

	// Generate token pair
	tokenPair, err := u.generateTokenPair(ctx, user.ID.Hex())
	if err != nil {
		logger.LogOutput(nil, fmt.Errorf("error generating tokens: %v", err))
		return nil, nil, fmt.Errorf("error generating tokens: %v", err)
	}

	result := struct {
		User      *domain.User
		TokenPair *domain.TokenPair
	}{
		User:      user,
		TokenPair: tokenPair,
	}
	logger.LogOutput(result, nil)
	return user, tokenPair, nil
}

func (u *authUseCase) RefreshToken(ctx context.Context, refreshToken string) (*domain.TokenPair, error) {
	logger := utils.NewLogger("AuthUseCase.RefreshToken")
	logger.LogInput(map[string]string{
		"refreshToken": refreshToken,
	})

	// Verify refresh token
	token, err := jwt.Parse(refreshToken, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(u.refreshTokenSecret), nil
	})

	if err != nil {
		logger.LogOutput(nil, fmt.Errorf("error parsing refresh token: %v", err))
		return nil, err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		logger.LogOutput(nil, fmt.Errorf("invalid refresh token"))
		return nil, fmt.Errorf("invalid refresh token")
	}

	// Check if refresh token is in Redis
	userIDInterface := claims["userId"]
	if userIDInterface == nil {
		logger.LogOutput(nil, fmt.Errorf("userId not found in claims"))
		return nil, fmt.Errorf("invalid refresh token: userId not found")
	}

	userID, ok := userIDInterface.(string)
	if !ok {
		logger.LogOutput(nil, fmt.Errorf("userId is not a string"))
		return nil, fmt.Errorf("invalid refresh token: invalid userId format")
	}

	key := fmt.Sprintf("refresh_token:%s:%s", userID, refreshToken)
	exists, err := u.redisClient.Exists(ctx, key).Result()
	if err != nil {
		logger.LogOutput(nil, fmt.Errorf("error checking refresh token in Redis: %v", err))
		return nil, err
	}
	if exists == 0 {
		logger.LogOutput(nil, fmt.Errorf("refresh token has been revoked"))
		return nil, fmt.Errorf("refresh token has been revoked")
	}

	// Generate new token pair
	tokenPair, err := u.generateTokenPair(ctx, userID)
	if err != nil {
		logger.LogOutput(nil, fmt.Errorf("error generating new token pair: %v", err))
		return nil, err
	}

	logger.LogOutput(tokenPair, nil)
	return tokenPair, nil
}

func (u *authUseCase) RevokeRefreshToken(ctx context.Context, refreshToken string) error {
	logger := utils.NewLogger("AuthUseCase.RevokeRefreshToken")
	logger.LogInput(map[string]string{
		"refreshToken": refreshToken,
	})

	// Verify refresh token
	token, err := jwt.Parse(refreshToken, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(u.refreshTokenSecret), nil
	})

	if err != nil {
		logger.LogOutput(nil, fmt.Errorf("error parsing refresh token: %v", err))
		return err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		logger.LogOutput(nil, fmt.Errorf("invalid refresh token"))
		return fmt.Errorf("invalid refresh token")
	}

	// Remove refresh token from Redis
	userIDInterface := claims["userId"]
	if userIDInterface == nil {
		logger.LogOutput(nil, fmt.Errorf("userId not found in claims"))
		return fmt.Errorf("invalid refresh token: userId not found")
	}

	userID, ok := userIDInterface.(string)
	if !ok {
		logger.LogOutput(nil, fmt.Errorf("userId is not a string"))
		return fmt.Errorf("invalid refresh token: invalid userId format")
	}

	key := fmt.Sprintf("refresh_token:%s:%s", userID, refreshToken)
	err = u.redisClient.Del(ctx, key).Err()
	if err != nil {
		logger.LogOutput(nil, fmt.Errorf("error revoking refresh token: %v", err))
		return err
	}

	logger.LogOutput("Refresh token revoked successfully", nil)
	return nil
}

func (u *authUseCase) CreateTestToken(ctx context.Context, userID string) (*domain.TokenPair, error) {
	logger := utils.NewLogger("AuthUseCase.CreateTestToken")
	logger.LogInput(userID)

	// Find user
	user, err := u.userRepo.FindByID(userID)
	if err != nil {
		logger.LogOutput(nil, err)
		return nil, err
	}
	if user == nil {
		err = fmt.Errorf("user not found")
		logger.LogOutput(nil, err)
		return nil, err
	}

	// Create access token
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": user.ID.Hex(),
		"exp": time.Now().Add(u.tokenExpiry).Unix(),
	})
	accessTokenString, err := accessToken.SignedString([]byte(u.jwtSecret))
	if err != nil {
		logger.LogOutput(nil, err)
		return nil, err
	}

	// Create refresh token
	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": user.ID.Hex(),
		"exp": time.Now().Add(u.refreshTokenExpiry).Unix(),
	})
	refreshTokenString, err := refreshToken.SignedString([]byte(u.refreshTokenSecret))
	if err != nil {
		logger.LogOutput(nil, err)
		return nil, err
	}

	// Store refresh token in Redis
	err = u.redisClient.Set(ctx, fmt.Sprintf("refresh_token:%s", refreshTokenString), user.ID.Hex(), u.refreshTokenExpiry).Err()
	if err != nil {
		logger.LogOutput(nil, err)
		return nil, err
	}

	tokenPair := &domain.TokenPair{
		AccessToken:  accessTokenString,
		RefreshToken: refreshTokenString,
	}

	logger.LogOutput(tokenPair, nil)
	return tokenPair, nil
}

func (u *authUseCase) generateTokenPair(ctx context.Context, userID string) (*domain.TokenPair, error) {
	logger := utils.NewLogger("AuthUseCase.generateTokenPair")
	logger.LogInput(userID)

	// Generate access token
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"userId": userID,
		"exp":    time.Now().Add(u.tokenExpiry).Unix(),
		"type":   "access",
	})

	accessTokenString, err := accessToken.SignedString([]byte(u.jwtSecret))
	if err != nil {
		logger.LogOutput(nil, fmt.Errorf("error generating access token: %v", err))
		return nil, err
	}

	// Generate refresh token
	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"userId": userID,
		"exp":    time.Now().Add(u.refreshTokenExpiry).Unix(),
		"type":   "refresh",
		"jti":    generateRandomString(32),
	})

	refreshTokenString, err := refreshToken.SignedString([]byte(u.refreshTokenSecret))
	if err != nil {
		logger.LogOutput(nil, fmt.Errorf("error generating refresh token: %v", err))
		return nil, err
	}

	// Store refresh token in Redis
	key := fmt.Sprintf("refresh_token:%s:%s", userID, refreshTokenString)
	err = u.redisClient.Set(ctx, key, "valid", u.refreshTokenExpiry).Err()
	if err != nil {
		logger.LogOutput(nil, fmt.Errorf("error storing refresh token in Redis: %v", err))
		return nil, err
	}

	tokenPair := &domain.TokenPair{
		AccessToken:  accessTokenString,
		RefreshToken: refreshTokenString,
	}
	logger.LogOutput(tokenPair, nil)
	return tokenPair, nil
}

func generateRandomString(n int) string {
	b := make([]byte, n)
	rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)
}

func getProviderFromFirebase(providerID string) domain.AuthProvider {
	logger := utils.NewLogger("AuthUseCase.getProviderFromFirebase")
	logger.LogInput(providerID)
	switch providerID {
	case "google.com":
		return domain.Google
	case "apple.com":
		return domain.Apple
	default:
		return domain.Email
	}
}
