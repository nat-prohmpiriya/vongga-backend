package middleware

import (
	"fmt"
	"strings"

	"firebase.google.com/go/v4/auth"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/prohmpiriya_phonumnuaisuk/vongga-platform/vongga-backend/utils"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func JWTAuthMiddleware(jwtSecret string, authClient *auth.Client) fiber.Handler {
	return func(c *fiber.Ctx) error {
		logger := utils.NewLogger("JWTAuthMiddleware")
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			logger.LogInput(authHeader)
			logger.LogOutput(nil, fmt.Errorf("missing authorization header"))
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "missing authorization header",
			})
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == authHeader {
			logger.LogInput(tokenString)
			logger.LogOutput(nil, fmt.Errorf("invalid token format"))
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "invalid token format",
			})
		}

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				logger.LogInput(tokenString)
				logger.LogOutput(nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"]))
				return nil, fiber.ErrUnauthorized
			}
			logger.LogInput(tokenString)
			return []byte(jwtSecret), nil
		})

		if err != nil {
			logger.LogInput(tokenString)
			logger.LogOutput(nil, err)
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "invalid token",
			})
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok || !token.Valid {
			logger.LogInput(tokenString)
			logger.LogOutput(nil, fmt.Errorf("invalid token claims"))
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "invalid token claims",
			})
		}

		// Store user ID in context
		userIDStr, ok := claims["userId"].(string)
		if !ok {
			logger.LogOutput(nil, fmt.Errorf("userId is not a string"))
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "invalid user ID format in token",
			})
		}

		userID, err := primitive.ObjectIDFromHex(userIDStr)
		if err != nil {
			logger.LogOutput(nil, err)
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "invalid user ID format",
			})
		}

		c.Locals("userId", userID)
		c.Locals("firebase_auth", authClient)
		logger.LogInput(userID)
		return c.Next()
	}
}
