package middleware

import (
	"fmt"
	"strings"

	"vongga-api/utils"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

func AuthMiddleware(jwtSecret string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		logger := utils.NewTraceLogger("AuthMiddleware")
		logger.Input(c)

		authHeader := c.Get("Authorization")
		if authHeader == "" {
			logger.Output(nil, fmt.Errorf("missing authorization header"))
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "unauthorized",
			})
		}

		tokenString := strings.Replace(authHeader, "Bearer ", "", 1)
		logger.LogInfo(tokenString)
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			logger.Output(nil, fmt.Errorf("parsing token"))
			return []byte(jwtSecret), nil
		})

		if err != nil || !token.Valid {
			logger.Output(nil, fmt.Errorf("invalid token"))
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "invalid token",
			})
		}

		claims := token.Claims.(jwt.MapClaims)
		logger.Input(map[string]interface{}{
			"claims":      claims,
			"userIdValue": claims["userId"],
			"userIdType":  fmt.Sprintf("%T", claims["userId"]),
		}, nil)
		// Convert userId to string before setting in context
		userID, ok := claims["userId"].(string)
		if !ok {
			logger.Output(nil, fmt.Errorf("userId is not a string"))
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
