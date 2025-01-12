// middleware/websocket_auth.go
package middleware

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
	"go.opentelemetry.io/otel/trace"

	"vongga-api/internal/domain"
	"vongga-api/utils"
)

func WebsocketAuthMiddleware(authUseCase domain.AuthUseCase, tracer trace.Tracer) fiber.Handler {
	return func(c *fiber.Ctx) error {
		ctx, span := tracer.Start(c.Context(), "Middleware.WebsocketAuthMiddleware")
		defer span.End()
		logger := utils.NewTraceLogger(span)

		// 1. ตรวจสอบว่าเป็น WebSocket request
		if !websocket.IsWebSocketUpgrade(c) {
			return fiber.ErrUpgradeRequired
		}

		// 2. ดึง token จาก query params
		token := c.Query("token")
		if token == "" {
			logger.Output("missing token", nil)
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "missing token",
			})
		}

		// 3. verify token
		claims, err := authUseCase.VerifyToken(ctx, token)
		if err != nil {
			logger.Output("invalid token", err)
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "invalid token",
			})
		}

		// 4. เก็บ claims และอนุญาต upgrade
		c.Locals("user", claims)
		return c.Next()
	}
}
