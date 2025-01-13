package usecase

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"time"

	"vongga_api/internal/domain"
	"vongga_api/utils"

	"firebase.google.com/go/v4/auth"
	"github.com/golang-jwt/jwt/v5"
	"github.com/redis/go-redis/v9"
	"go.opentelemetry.io/otel/trace"
)

type authUseCase struct {
	userRepo           domain.UserRepository
	authClient         *auth.Client
	redisClient        *redis.Client
	jwtSecret          string
	refreshTokenSecret string
	tokenExpiry        time.Duration
	refreshTokenExpiry time.Duration
	tracer             trace.Tracer
}

func NewAuthUseCase(
	userRepo domain.UserRepository,
	authClient *auth.Client,
	redisClient *redis.Client,
	jwtSecret string,
	refreshTokenSecret string,
	tokenExpiry time.Duration,
	refreshTokenExpiry time.Duration,
	tracer trace.Tracer,
) domain.AuthUseCase {
	return &authUseCase{
		userRepo:           userRepo,
		authClient:         authClient,
		redisClient:        redisClient,
		jwtSecret:          jwtSecret,
		refreshTokenSecret: refreshTokenSecret,
		tokenExpiry:        tokenExpiry,
		refreshTokenExpiry: refreshTokenExpiry,
		tracer:             tracer,
	}
}

func (u *authUseCase) VerifyTokenFirebase(ctx context.Context, token string) (*domain.User, *domain.TokenPair, error) {
	ctx, span := u.tracer.Start(ctx, "AuthUseCase.VerifyTokenFirebase")
	defer span.End()
	logger := utils.NewTraceLogger(span)
	logger.Input(token)

	// Verify token with Firebase
	firebaseToken, err := u.authClient.VerifyIDToken(ctx, token)
	if err != nil {
		logger.Output("error verifying token 1", err)
		return nil, nil, fmt.Errorf("error verifying token: %v", err)
	}

	// Find existing user by Firebase UID
	user, err := u.userRepo.FindByFirebaseUID(ctx, firebaseToken.UID)
	if err != nil {
		logger.Output("error finding user 2", err)
		return nil, nil, fmt.Errorf("error finding user: %v", err)
	}

	// Create new user if not exists
	if user == nil {
		// Find user info from Firebase
		firebaseUser, err := u.authClient.GetUser(ctx, firebaseToken.UID)
		if err != nil {
			logger.Output("error getting firebase user 3", err)
			return nil, nil, fmt.Errorf("error getting firebase user: %v", err)
		}
		logger.Input(firebaseUser)

		user = &domain.User{
			FirebaseUID:   firebaseToken.UID,
			Role:          "user",
			Provider:      getProviderFromFirebase(firebaseUser.ProviderUserInfo[0].ProviderID, logger),
			Email:         firebaseUser.Email,
			DisplayName:   firebaseUser.DisplayName,
			EmailVerified: firebaseUser.EmailVerified,
			PhotoProfile:  firebaseUser.PhotoURL,
			// other user fields from firebaseToken
		}
		err = u.userRepo.Create(ctx, user)
		if err != nil {
			logger.Output("error creating user 3", err)
			return nil, nil, fmt.Errorf("error creating user: %v", err)
		}
	}

	// Generate token pair
	tokenPair, err := u.generateTokenPair(ctx, user.ID.Hex(), user.Role, user.Provider)
	if err != nil {
		logger.Output("error generating tokens 4", err)
		return nil, nil, fmt.Errorf("error generating tokens: %v", err)
	}

	logger.Output(map[string]interface{}{
		"tokenPair": tokenPair,
		"user":      user,
	}, nil)
	return user, tokenPair, nil
}

func (u *authUseCase) VerifyToken(ctx context.Context, token string) (*domain.Claims, error) {
	_, span := u.tracer.Start(ctx, "AuthUseCase.VerifyToken")
	defer span.End()
	logger := utils.NewTraceLogger(span)
	logger.Input(token)

	// 1. ตรวจสอบว่า token ว่างไหม
	if token == "" {
		logger.Output("token is empty", nil)
		return nil, fmt.Errorf("token is empty")
	}

	// 2. Parse และ verify token
	claims := &domain.Claims{}
	parsedToken, err := jwt.ParseWithClaims(token, claims, func(t *jwt.Token) (interface{}, error) {
		// ตรวจสอบ algorithm
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return []byte(u.jwtSecret), nil
	})

	if err != nil {
		if err == jwt.ErrSignatureInvalid {
			logger.Output("invalid signature", err)
			return nil, fmt.Errorf("invalid signature")
		}
		logger.Output("failed to parse token", err)
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	// 3. ตรวจสอบว่า token valid
	if !parsedToken.Valid {
		logger.Output("token is invalid 1", nil)
		return nil, fmt.Errorf("token is invalid")
	}

	// 4.การเช็ค userId ตรงนี้
	if claims.UserID == "" {
		logger.Output("token is invalid: missing userId", nil)
		return nil, fmt.Errorf("token is invalid: missing userId")
	}

	// 4. ตรวจสอบ expiration
	if claims.ExpiresAt.Before(time.Now()) {
		logger.Output("token is expired 2", nil)
		return nil, fmt.Errorf("token is expired")
	}

	// 5. ตรวจสอบว่า user ยังมีสิทธิ์ใช้งานไหม (optional)

	logger.Output("token verified successfully", nil)
	return claims, nil
}

func (u *authUseCase) RefreshToken(ctx context.Context, refreshToken string) (*domain.TokenPair, error) {
	ctx, span := u.tracer.Start(ctx, "AuthUseCase.RefreshToken")
	defer span.End()
	logger := utils.NewTraceLogger(span)
	logger.Input(map[string]string{
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
		logger.Output("invalid refresh token 1", err)
		return nil, err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		logger.Output("invalid refresh token", nil)
		return nil, fmt.Errorf("invalid refresh token")
	}

	// Check if refresh token is in Redis
	userIDInterface := claims["sub"]
	if userIDInterface == nil {
		logger.Output("userId not found in claims", nil)
		return nil, fmt.Errorf("invalid refresh token: userId not found")
	}

	userID, ok := userIDInterface.(string)
	if !ok {
		logger.Output("userId is not a string", nil)
		return nil, fmt.Errorf("invalid refresh token: invalid userId format")
	}

	key := fmt.Sprintf("refresh_token:%s:%s", userID, refreshToken)
	exists, err := u.redisClient.Exists(ctx, key).Result()
	if err != nil {
		logger.Output("error checking refresh token in Redis", err)
		return nil, err
	}
	if exists == 0 {
		logger.Output("refresh token has been revoked", nil)
		return nil, fmt.Errorf("refresh token has been revoked")
	}

	// Generate new token pair
	tokenPair, err := u.generateTokenPair(ctx, userID, "user", "")
	if err != nil {
		logger.Output("error generating new token pair 4", err)
		return nil, err
	}

	logger.Output(tokenPair, nil)
	return tokenPair, nil
}

func (u *authUseCase) RevokeRefreshToken(ctx context.Context, refreshToken string) error {
	ctx, span := u.tracer.Start(ctx, "AuthUseCase.RevokeRefreshToken")
	defer span.End()
	logger := utils.NewTraceLogger(span)
	logger.Input(map[string]string{
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
		logger.Output("invalid refresh token 1", err)
		return err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		logger.Output("invalid refresh token", nil)
		return fmt.Errorf("invalid refresh token")
	}

	// Remove refresh token from Redis
	userIDInterface := claims["sub"]
	if userIDInterface == nil {
		logger.Output("userId not found in claims", nil)
		return fmt.Errorf("invalid refresh token: userId not found")
	}

	userID, ok := userIDInterface.(string)
	if !ok {
		logger.Output("userId is not a string", nil)
		return fmt.Errorf("invalid refresh token: invalid userId format")
	}

	key := fmt.Sprintf("refresh_token:%s:%s", userID, refreshToken)
	err = u.redisClient.Del(ctx, key).Err()
	if err != nil {
		logger.Output("error revoking token 3", err)
		return err
	}

	logger.Output("Refresh token revoked successfully", nil)
	return nil
}

func (u *authUseCase) CreateTestToken(ctx context.Context, userID string) (*domain.User, *domain.TokenPair, error) {
	ctx, span := u.tracer.Start(ctx, "AuthUseCase.CreateTestToken")
	defer span.End()
	logger := utils.NewTraceLogger(span)
	logger.Input(userID)

	// Find user
	user, err := u.userRepo.FindByID(ctx, userID)
	if err != nil {
		logger.Output("error finding user", err)
		return nil, nil, err
	}
	if user == nil {
		err = fmt.Errorf("user not found")
		logger.Output("user not found", err)
		return nil, nil, err
	}

	// Check if user is admin
	logger.Info(user)
	if user.Role != "admin" {
		err = fmt.Errorf("unauthorized: user is not admin")
		logger.Output("user is not admin", err)
		return nil, nil, err
	}

	// Generate token pair
	tokenPair, err := u.generateTokenPair(ctx, user.ID.Hex(), user.Role, user.Provider)
	if err != nil {
		logger.Output("error generating tokens", err)
		return nil, nil, err
	}

	logger.Output(map[string]interface{}{
		"user":      user,
		"tokenPair": tokenPair,
	}, nil)
	return user, tokenPair, nil
}

func (u *authUseCase) generateTokenPair(ctx context.Context, userID string, role string, provider domain.AuthProvider) (*domain.TokenPair, error) {
	ctx, span := u.tracer.Start(ctx, "AuthUseCase.generateTokenPair")
	defer span.End()
	logger := utils.NewTraceLogger(span)
	logger.Input(map[string]interface{}{
		"userID":   userID,
		"role":     role,
		"provider": provider,
	})

	// Generate access token
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":       userID,
		"role":      role,
		"provider":  provider,
		"exp":       time.Now().Add(u.tokenExpiry).Unix(),
		"tokenType": "access",
	})

	accessTokenString, err := accessToken.SignedString([]byte(u.jwtSecret))
	if err != nil {
		logger.Output("error signing access token 1", err)
		return nil, err
	}

	// Generate refresh token
	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":       userID,
		"role":      role,
		"provider":  provider,
		"exp":       time.Now().Add(u.refreshTokenExpiry).Unix(),
		"tokenType": "refresh",
		"jti":       generateRandomString(32),
	})

	refreshTokenString, err := refreshToken.SignedString([]byte(u.refreshTokenSecret))
	if err != nil {
		logger.Output("error signing refresh token 2", err)
		return nil, err
	}

	// Store refresh token in Redis
	key := fmt.Sprintf("refresh_token:%s:%s", userID, refreshTokenString)
	err = u.redisClient.Set(ctx, key, "valid", u.refreshTokenExpiry).Err()
	if err != nil {
		logger.Output("error storing refresh token 3", err)
		return nil, err
	}

	tokenPair := &domain.TokenPair{
		AccessToken:  accessTokenString,
		RefreshToken: refreshTokenString,
	}
	logger.Output(tokenPair, nil)
	return tokenPair, nil
}

func generateRandomString(n int) string {
	b := make([]byte, n)
	rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)
}

func getProviderFromFirebase(providerID string, logger *utils.TraceLogger) domain.AuthProvider {
	logger.Input(providerID)
	switch providerID {
	case "google.com":
		logger.Output(domain.Google, nil)
		return domain.Google
	case "apple.com":
		logger.Output(domain.Apple, nil)
		return domain.Apple
	default:
		logger.Output(domain.Email, nil)
		return domain.Email
	}
}
