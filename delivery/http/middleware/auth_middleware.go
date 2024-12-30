package middleware

import (
	"fmt"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/prohmpiriya_phonumnuaisuk/vongga-platform/vongga-backend/utils"
)

func AuthMiddleware(jwtSecret string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		logger := utils.NewLogger("AuthMiddleware")
		logger.LogInput(c)

		authHeader := c.Get("Authorization")
		if authHeader == "" {
			logger.LogOutput(nil, fmt.Errorf("missing authorization header"))
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "unauthorized",
			})
		}

		tokenString := strings.Replace(authHeader, "Bearer ", "", 1)
		logger.LogInfo(tokenString)
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			logger.LogOutput(nil, fmt.Errorf("parsing token"))
			return []byte(jwtSecret), nil
		})

		if err != nil || !token.Valid {
			logger.LogOutput(nil, fmt.Errorf("invalid token"))
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "invalid token",
			})
		}

		claims := token.Claims.(jwt.MapClaims)
		logger.LogInput(map[string]interface{}{
			"claims":      claims,
			"userIdValue": claims["userId"],
			"userIdType":  fmt.Sprintf("%T", claims["userId"]),
		}, nil)
		// Convert userId to string before setting in context
		userID, ok := claims["userId"].(string)
		if !ok {
			logger.LogOutput(nil, fmt.Errorf("userId is not a string"))
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid token format",
			})
		}

		// Set userId as string in context
		c.Locals("userId", userID)
		logger.LogOutput(userID, nil)
		return c.Next()
	}
}
