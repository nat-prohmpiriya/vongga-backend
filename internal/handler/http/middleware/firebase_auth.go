package middleware

import (
	"context"
	"fmt"
	"strings"

	"vongga-api/utils"

	"firebase.google.com/go/v4/auth"
	"github.com/gofiber/fiber/v2"
)

func FirebaseAuthMiddleware(auth *auth.Client) fiber.Handler {
	return func(c *fiber.Ctx) error {
		logger := utils.NewTraceLogger("FirebaseAuthMiddleware")
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			logger.Input(authHeader)
			logger.Output(nil, fmt.Errorf("missing authorization header"))
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "missing authorization header",
			})
		}

		idToken := strings.TrimPrefix(authHeader, "Bearer ")
		if idToken == authHeader {
			logger.Input(idToken)
			logger.Output(nil, fmt.Errorf("invalid token format"))
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "invalid token format",
			})
		}

		token, err := auth.VerifyIDToken(context.Background(), idToken)
		if err != nil {
			logger.Input(idToken)
			logger.Output(nil, err)
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "invalid token",
			})
		}

		// Store Firebase UID and email in context
		c.Locals("firebase_uid", token.UID)
		if email, ok := token.Claims["email"].(string); ok {
			c.Locals("email", email)
		}
		logger.Output(token.UID, nil)
		return c.Next()
	}
}
