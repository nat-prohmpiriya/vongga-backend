package middleware

import (
	"fmt"
	"strings"

	"vongga-api/utils"

	"firebase.google.com/go/v4/auth"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func JWTAuthMiddleware(jwtSecret string, authClient *auth.Client) fiber.Handler {
	return func(c *fiber.Ctx) error {
		logger := utils.NewTraceLogger("JWTAuthMiddleware")
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			logger.Input(authHeader)
			logger.Output(nil, fmt.Errorf("missing authorization header"))
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "missing authorization header",
			})
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == authHeader {
			logger.Input(tokenString)
			logger.Output(nil, fmt.Errorf("invalid token format"))
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "invalid token format",
			})
		}

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				logger.Input(tokenString)
				logger.Output(nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"]))
				return nil, fiber.ErrUnauthorized
			}
			logger.Input(tokenString)
			return []byte(jwtSecret), nil
		})

		if err != nil {
			logger.Input(tokenString)
			logger.Output(nil, err)
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "invalid token",
			})
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok || !token.Valid {
			logger.Input(tokenString)
			logger.Output(nil, fmt.Errorf("invalid token claims"))
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "invalid token claims",
			})
		}

		// Store user ID in context
		userIDValue := claims["userId"]
		var userIDStr string

		// Try to convert userId to string based on its type
		switch v := userIDValue.(type) {
		case string:
			userIDStr = v
		case interface{}:
			// Try to convert to string directly
			if str, ok := v.(string); ok {
				userIDStr = str
			} else {
				logger.Output(nil, fmt.Errorf("userId is not a valid string: %T", v))
				return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
					"error": "invalid user ID format in token",
				})
			}
		default:
			logger.Output(nil, fmt.Errorf("userId is not a valid format: %T", v))
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "invalid user ID format in token",
			})
		}

		userID, err := primitive.ObjectIDFromHex(userIDStr)
		if err != nil {
			logger.Output(nil, err)
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "invalid user ID format",
			})
		}

		c.Locals("userId", userID)
		c.Locals("firebase_auth", authClient)
		logger.Input(userID)
		return c.Next()
	}
}
