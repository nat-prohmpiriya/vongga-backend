package utils

import (
	"errors"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// FindUserIDFromContext retrieves the user ID from the Fiber context
func FindUserIDFromContext(c *fiber.Ctx) (primitive.ObjectID, error) {
	userIDValue := c.Locals("userId")
	if userIDValue == nil {
		return primitive.NilObjectID, errors.New("userId not found in context")
	}

	// Try to convert to ObjectID directly
	if userID, ok := userIDValue.(primitive.ObjectID); ok {
		return userID, nil
	}

	// If not ObjectID, try to convert from string
	if userIDStr, ok := userIDValue.(string); ok {
		userID, err := primitive.ObjectIDFromHex(userIDStr)
		if err != nil {
			return primitive.NilObjectID, errors.New("invalid userId format")
		}
		return userID, nil
	}

	return primitive.NilObjectID, errors.New("userId in context is not a valid format")
}
