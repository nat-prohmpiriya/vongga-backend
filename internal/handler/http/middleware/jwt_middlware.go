package middleware

import (
	"fmt"
	"strings"

	"vongga-api/utils"

	"firebase.google.com/go/v4/auth"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.opentelemetry.io/otel/trace"
)

func JWTAuthMiddleware(jwtSecret string, authClient *auth.Client, trcer trace.Tracer) fiber.Handler {
	return func(c *fiber.Ctx) error {
		_, span := trcer.Start(c.UserContext(), "middleware.JWTAuthMiddleware")
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

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		logger.Input(tokenString)
		if tokenString == authHeader {
			logger.Output("invalid token format", fmt.Errorf("invalid token format"))
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "invalid token format",
			})
		}

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			logger.Input(tokenString)
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				logger.Output("unexpected signing method", fmt.Errorf("unexpected signing method: %v", token.Header["alg"]))
				return nil, fiber.ErrUnauthorized
			}
			return []byte(jwtSecret), nil
		})

		if err != nil {
			logger.Output(nil, err)
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "invalid token",
			})
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		logger.Input(claims)
		if !ok || !token.Valid {
			logger.Output("invalid token claims", fmt.Errorf("invalid token claims"))
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
				logger.Output("userId is not a valid string", fmt.Errorf("userId is not a valid string: %T", v))
				return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
					"error": "invalid user ID format in token",
				})
			}
		default:
			logger.Output("userId is not a valid format", fmt.Errorf("userId is not a valid format: %T", v))
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
