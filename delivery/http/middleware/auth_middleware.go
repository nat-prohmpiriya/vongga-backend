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
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			logger.LogInput(nil, fmt.Errorf("parsing token"))
			return []byte(jwtSecret), nil
		})

		if err != nil || !token.Valid {
			logger.LogOutput(nil, fmt.Errorf("invalid token"))
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "invalid token",
			})
		}

		claims := token.Claims.(jwt.MapClaims)
		c.Locals("userId", claims["userId"])
		logger.LogOutput(claims["userId"], nil)
		return c.Next()
	}
}
