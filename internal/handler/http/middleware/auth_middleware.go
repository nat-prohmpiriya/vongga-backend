package middleware

import (
	"fmt"
	"strings"

	"vongga-api/utils"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"go.opentelemetry.io/otel/trace"
)

func AuthMiddleware(jwtSecret string, trcer trace.Tracer) fiber.Handler {
	return func(c *fiber.Ctx) error {
		_, span := trcer.Start(c.UserContext(), "middleware.AuthMiddleware")
		defer span.End()
		logger := utils.NewTraceLogger(span)

		authHeader := c.Get("Authorization")
		if authHeader == "" {
			logger.Output("missing authorization header", nil)
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "unauthorized",
			})
		}

		tokenString := strings.Replace(authHeader, "Bearer ", "", 1)
		logger.Info(tokenString)
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			logger.Output("parsing token", nil)
			return []byte(jwtSecret), nil
		})

		if err != nil || !token.Valid {
			logger.Output("invalid token", nil)
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "invalid token",
			})
		}

		claims := token.Claims.(jwt.MapClaims)
		logger.Input(map[string]interface{}{
			"claims":      claims,
			"userIdValue": claims["userId"],
			"userIdType":  fmt.Sprintf("%T", claims["userId"]),
		})
		// Convert userId to string before setting in context
		userID, ok := claims["userId"].(string)
		if !ok {
			logger.Output("Invalid token format", fmt.Errorf("invalid token format"))
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid token format",
			})
		}

		// Set userId as string in context
		c.Locals("userId", userID)
		logger.Output(userID, nil)
		return c.Next()
	}
}
