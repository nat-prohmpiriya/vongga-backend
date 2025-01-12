package middleware

import (
	"context"
	"fmt"
	"strings"

	"vongga-api/utils"

	"firebase.google.com/go/v4/auth"
	"github.com/gofiber/fiber/v2"
	"go.opentelemetry.io/otel/trace"
)

func FirebaseAuthMiddleware(auth *auth.Client, trcer trace.Tracer) fiber.Handler {
	return func(c *fiber.Ctx) error {
		_, span := trcer.Start(c.UserContext(), "middleware.FirebaseAuthMiddleware")
		defer span.End()
		logger := utils.NewTraceLogger(span)
		authHeader := c.Get("Authorization")
		logger.Input(authHeader)
		if authHeader == "" {
			logger.Output("missing authorization header", fmt.Errorf("missing authorization header"))
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "missing authorization header",
			})
		}

		idToken := strings.TrimPrefix(authHeader, "Bearer ")
		logger.Input(idToken)
		if idToken == authHeader {
			logger.Output("invalid token format", fmt.Errorf("invalid token format"))
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "invalid token format",
			})
		}

		token, err := auth.VerifyIDToken(context.Background(), idToken)
		logger.Input(idToken)
		if err != nil {
			logger.Output("invalid token", err)
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
