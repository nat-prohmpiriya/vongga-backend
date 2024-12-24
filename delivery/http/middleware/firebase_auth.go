package middleware

import (
	"context"
	"fmt"
	"strings"

	"firebase.google.com/go/v4/auth"
	"github.com/gofiber/fiber/v2"
	"github.com/prohmpiriya_phonumnuaisuk/vongga-platform/vongga-backend/utils"
)

func FirebaseAuthMiddleware(auth *auth.Client) fiber.Handler {
	return func(c *fiber.Ctx) error {
		logger := utils.NewLogger("FirebaseAuthMiddleware")
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			logger.LogInput(authHeader)
			logger.LogOutput(nil, fmt.Errorf("missing authorization header"))
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "missing authorization header",
			})
		}

		idToken := strings.TrimPrefix(authHeader, "Bearer ")
		if idToken == authHeader {
			logger.LogInput(idToken)
			logger.LogOutput(nil, fmt.Errorf("invalid token format"))
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "invalid token format",
			})
		}

		token, err := auth.VerifyIDToken(context.Background(), idToken)
		if err != nil {
			logger.LogInput(idToken)
			logger.LogOutput(nil, err)
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "invalid token",
			})
		}

		// Store Firebase UID and email in context
		c.Locals("firebase_uid", token.UID)
		if email, ok := token.Claims["email"].(string); ok {
			c.Locals("email", email)
		}
		logger.LogOutput(token.UID, nil)
		return c.Next()
	}
}
