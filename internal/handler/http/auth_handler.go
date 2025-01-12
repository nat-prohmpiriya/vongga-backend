package handler

import (
	"vongga-api/internal/domain"
	"vongga-api/utils"

	"github.com/gofiber/fiber/v2"
)

type AuthHandler struct {
	authUseCase domain.AuthUseCase
}

func NewAuthHandler(authUseCase domain.AuthUseCase) *AuthHandler {
	return &AuthHandler{
		authUseCase: authUseCase,
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
	logger := utils.NewTraceLogger("AuthHandler.CreateTestToken")

	var req CreateTestTokenRequest
	if err := c.BodyParser(&req); err != nil {
		logger.Input(req)
		logger.Output(nil, err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	logger.Input(req)
	tokenPair, err := h.authUseCase.CreateTestToken(c.Context(), req.UserID)
	if err != nil {
		logger.Output(nil, err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	response := TokenResponse{
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
	}

	logger.Output(response, nil)
	return c.JSON(response)
}

// VerifyTokenFirebase verifies Firebase ID token and returns user data with JWT tokens
func (h *AuthHandler) VerifyTokenFirebase(c *fiber.Ctx) error {
	logger := utils.NewTraceLogger("AuthHandler.VerifyTokenFirebase")

	var req struct {
		FirebaseToken string `json:"firebaseToken"`
	}

	if err := c.BodyParser(&req); err != nil {
		logger.Input(req)
		logger.Output(nil, err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	logger.Input(req)
	user, tokenPair, err := h.authUseCase.VerifyTokenFirebase(c.Context(), req.FirebaseToken)
	if err != nil {
		logger.Output(nil, err)
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
	logger := utils.NewTraceLogger("AuthHandler.RefreshToken")

	var req RefreshTokenRequest
	if err := c.BodyParser(&req); err != nil {
		logger.Input(req)
		logger.Output(nil, err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	logger.Input(req)
	tokenPair, err := h.authUseCase.RefreshToken(c.Context(), req.RefreshToken)
	if err != nil {
		logger.Output(nil, err)
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	response := TokenResponse{
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
	}

	logger.Output(response, nil)
	return c.JSON(response)
}

// Logout revokes the refresh token
func (h *AuthHandler) Logout(c *fiber.Ctx) error {
	logger := utils.NewTraceLogger("AuthHandler.Logout")

	var req LogoutRequest
	if err := c.BodyParser(&req); err != nil {
		logger.Input(req)
		logger.Output(nil, err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	logger.Input(req)
	err := h.authUseCase.RevokeRefreshToken(c.Context(), req.RefreshToken)
	if err != nil {
		logger.Output(nil, err)
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
