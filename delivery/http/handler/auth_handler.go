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

// Login godoc
// @Summary Login with Firebase token
// @Description Login with Firebase ID token and get JWT tokens
// @Tags auth
// @Accept json
// @Produce json
// @Param request body LoginRequest true "Login request"
// @Success 200 {object} LoginResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Router /auth/login [post]
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

// RefreshToken godoc
// @Summary Refresh access token
// @Description Get new access token using refresh token
// @Tags auth
// @Accept json
// @Produce json
// @Param request body RefreshTokenRequest true "Refresh token request"
// @Success 200 {object} TokenResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Router /auth/refresh [post]
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

// Logout godoc
// @Summary Logout user
// @Description Revoke refresh token
// @Tags auth
// @Accept json
// @Produce json
// @Param request body LogoutRequest true "Logout request"
// @Success 200 "OK"
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /auth/logout [post]
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

	logger.LogOutput("Logout successful", nil)
	return c.SendStatus(fiber.StatusOK)
}

// LoginRequest represents the login request body
type LoginRequest struct {
	FirebaseToken string `json:"firebaseToken" example:"firebase_id_token_here"`
}

// RefreshTokenRequest represents the refresh token request body
type RefreshTokenRequest struct {
	RefreshToken string `json:"refreshToken" example:"refresh_token_here"`
}

// LogoutRequest represents the logout request body
type LogoutRequest struct {
	RefreshToken string `json:"refreshToken" example:"refresh_token_here"`
}

// LoginResponse represents the login response
type LoginResponse struct {
	User         *domain.User `json:"user"`
	AccessToken  string       `json:"accessToken" example:"access_token_here"`
	RefreshToken string       `json:"refreshToken" example:"refresh_token_here"`
}

// TokenResponse represents the token response
type TokenResponse struct {
	AccessToken  string `json:"accessToken" example:"access_token_here"`
	RefreshToken string `json:"refreshToken" example:"refresh_token_here"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error string `json:"error" example:"error message"`
}
