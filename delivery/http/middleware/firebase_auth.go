package middleware

import (
	"context"
	"strings"

	"firebase.google.com/go/v4/auth"
	"github.com/gofiber/fiber/v2"
)

func FirebaseAuthMiddleware(auth *auth.Client) fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "missing authorization header",
			})
		}

		idToken := strings.TrimPrefix(authHeader, "Bearer ")
		if idToken == authHeader {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "invalid token format",
			})
		}

		token, err := auth.VerifyIDToken(context.Background(), idToken)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "invalid token",
			})
		}

		// Store Firebase UID and email in context
		c.Locals("firebase_uid", token.UID)
		if email, ok := token.Claims["email"].(string); ok {
			c.Locals("email", email)
		}

		return c.Next()
	}
}
