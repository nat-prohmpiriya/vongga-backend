package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"go.opentelemetry.io/otel/trace"

	"vongga_api/internal/domain"
	"vongga_api/utils"
)

func HttpAuthMiddleware(authUseCase domain.AuthUseCase, tracer trace.Tracer) fiber.Handler {
	return func(c *fiber.Ctx) error {
		ctx, span := tracer.Start(c.Context(), "Middleware.HttpAuthMiddleware")
		defer span.End()
		logger := utils.NewTraceLogger(span)

		// 1. ตรวจสอบ header
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			logger.Output("missing authorization header", nil)
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "unauthorized",
			})
		}

		// 2. แยก token
		if !strings.HasPrefix(authHeader, "Bearer ") {
			logger.Output("invalid token format", nil)
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "invalid token format",
			})
		}
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")

		// 3. verify token
		claims, err := authUseCase.VerifyToken(ctx, tokenString)
		if err != nil {
			logger.Output("invalid token", err)
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "invalid token",
			})
		}

		// 4. เก็บ claims
		c.Locals("user", claims)

		return c.Next()
	}
}
