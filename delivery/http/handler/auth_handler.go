package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/prohmpiriya_phonumnuaisuk/vongga-platform/vongga-backend/domain"
	"github.com/prohmpiriya_phonumnuaisuk/vongga-platform/vongga-backend/utils"
)

type AuthHandler struct {
	authUseCase domain.AuthUseCase
}

func NewAuthHandler(authUseCase domain.AuthUseCase) *AuthHandler {
	return &AuthHandler{
		authUseCase: authUseCase,
	}
}

// VerifyTokenFirebase verifies Firebase ID token and returns user data with JWT tokens
func (h *AuthHandler) VerifyTokenFirebase(c *fiber.Ctx) error {
	logger := utils.NewLogger("AuthHandler.VerifyTokenFirebase")
	
	var req struct {
		FirebaseToken string `json:"firebaseToken"`
	}

	if err := c.BodyParser(&req); err != nil {
		logger.LogInput(req)
		logger.LogOutput(nil, err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	logger.LogInput(req)
	user, tokenPair, err := h.authUseCase.VerifyTokenFirebase(c.Context(), req.FirebaseToken)
	if err != nil {
		logger.LogOutput(nil, err)
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	response := LoginResponse{
		User:         user,
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
	}

	logger.LogOutput(response, nil)
	return c.JSON(response)
}

// RefreshToken refreshes the access token using a refresh token
func (h *AuthHandler) RefreshToken(c *fiber.Ctx) error {
	logger := utils.NewLogger("AuthHandler.RefreshToken")
	
	var req RefreshTokenRequest
	if err := c.BodyParser(&req); err != nil {
		logger.LogInput(req)
		logger.LogOutput(nil, err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	logger.LogInput(req)
	tokenPair, err := h.authUseCase.RefreshToken(c.Context(), req.RefreshToken)
	if err != nil {
		logger.LogOutput(nil, err)
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	response := TokenResponse{
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
	}

	logger.LogOutput(response, nil)
	return c.JSON(response)
}

// Logout revokes the refresh token
func (h *AuthHandler) Logout(c *fiber.Ctx) error {
	logger := utils.NewLogger("AuthHandler.Logout")
	
	var req LogoutRequest
	if err := c.BodyParser(&req); err != nil {
		logger.LogInput(req)
		logger.LogOutput(nil, err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	logger.LogInput(req)
	err := h.authUseCase.RevokeRefreshToken(c.Context(), req.RefreshToken)
	if err != nil {
		logger.LogOutput(nil, err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	logger.LogOutput("Successfully logged out", nil)
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
