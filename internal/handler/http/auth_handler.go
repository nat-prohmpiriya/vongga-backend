package handler

import (
	"vongga-api/internal/domain"
	"vongga-api/utils"

	"github.com/gofiber/fiber/v2"
	"go.opentelemetry.io/otel/trace"
)

type AuthHandler struct {
	authUseCase domain.AuthUseCase
	tracer      trace.Tracer
}

func NewAuthHandler(authUseCase domain.AuthUseCase, tracer trace.Tracer) *AuthHandler {
	return &AuthHandler{
		authUseCase: authUseCase,
		tracer:      tracer,
	}
}

// CreateTestToken creates a test access token for development purposes
// @Summary Create test access token
// @Description Creates a test access token for development purposes. Should only be used in development environment.
// @Tags auth
// @Accept json
// @Produce json
// @Param request body CreateTestTokenRequest true "User ID for test token"
// @Success 200 {object} TokenResponse
// @Failure 400 {object} ErrorResponse
// @Router /auth/createTestToken [post]
func (h *AuthHandler) CreateTestToken(c *fiber.Ctx) error {
	ctx, span := h.tracer.Start(c.UserContext(), "AuthHandler.CreateTestToken")
	defer span.End()
	logger := utils.NewTraceLogger(span)

	var req CreateTestTokenRequest
	if err := c.BodyParser(&req); err != nil {
		logger.Output("error parsing request 1", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	logger.Input(req)
	tokenPair, err := h.authUseCase.CreateTestToken(ctx, req.UserID)
	if err != nil {
		logger.Output("error creating test token 2", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	logger.Output(tokenPair, nil)
	return c.JSON(tokenPair)
}

// VerifyTokenFirebase verifies Firebase ID token and returns user data with JWT tokens
func (h *AuthHandler) VerifyTokenFirebase(c *fiber.Ctx) error {
	ctx, span := h.tracer.Start(c.UserContext(), "AuthHandler.VerifyTokenFirebase")
	defer span.End()
	logger := utils.NewTraceLogger(span)

	var req struct {
		FirebaseToken string `json:"firebaseToken"`
	}

	if err := c.BodyParser(&req); err != nil {
		logger.Output("error parsing request 1", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	logger.Input(req)
	user, tokenPair, err := h.authUseCase.VerifyTokenFirebase(ctx, req.FirebaseToken)
	if err != nil {
		logger.Output("error verifying token 2", err)
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	response := LoginResponse{
		User:         user,
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
	}

	logger.Output(response, nil)
	return c.JSON(response)
}

// RefreshToken refreshes the access token using a refresh token
func (h *AuthHandler) RefreshToken(c *fiber.Ctx) error {
	ctx, span := h.tracer.Start(c.UserContext(), "AuthHandler.RefreshToken")
	defer span.End()
	logger := utils.NewTraceLogger(span)

	var req RefreshTokenRequest
	if err := c.BodyParser(&req); err != nil {
		logger.Output("error parsing request 1", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	logger.Input(req)
	tokenPair, err := h.authUseCase.RefreshToken(ctx, req.RefreshToken)
	if err != nil {
		logger.Output("error refreshing token 2", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	logger.Output(tokenPair, nil)
	return c.JSON(tokenPair)
}

// Logout revokes the refresh token
func (h *AuthHandler) Logout(c *fiber.Ctx) error {
	ctx, span := h.tracer.Start(c.UserContext(), "AuthHandler.Logout")
	defer span.End()
	logger := utils.NewTraceLogger(span)

	var req LogoutRequest
	if err := c.BodyParser(&req); err != nil {
		logger.Output("error parsing request 1", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	logger.Input(req)
	err := h.authUseCase.RevokeRefreshToken(ctx, req.RefreshToken)
	if err != nil {
		logger.Output("error logging out 2", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	logger.Output("Successfully logged out", nil)
	return c.SendStatus(fiber.StatusOK)
}

// Request/Response types
type LoginRequest struct {
	FirebaseToken string `json:"firebaseToken" example:"firebase_id_token_here"`
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refreshToken" example:"refresh_token_here"`
}

type LogoutRequest struct {
	RefreshToken string `json:"refreshToken" example:"refresh_token_here"`
}

type CreateTestTokenRequest struct {
	UserID string `json:"userId" example:"userId_here"`
}

type LoginResponse struct {
	User         *domain.User `json:"user"`
	AccessToken  string       `json:"accessToken" example:"access_token_here"`
	RefreshToken string       `json:"refreshToken" example:"refresh_token_here"`
}

type TokenResponse struct {
	AccessToken  string `json:"accessToken" example:"access_token_here"`
	RefreshToken string `json:"refreshToken" example:"refresh_token_here"`
}

type ErrorResponse struct {
	Error string `json:"error" example:"error message"`
}
