package utils

import (
	"errors"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// GetUserIDFromContext retrieves the user ID from the Fiber context
func GetUserIDFromContext(c *fiber.Ctx) (primitive.ObjectID, error) {
	userIDStr := c.Locals("userId")
	if userIDStr == nil {
		return primitive.NilObjectID, errors.New("userId not found in context")
	}

	// Convert to string
	userIDString, ok := userIDStr.(string)
	if !ok {
		return primitive.NilObjectID, errors.New("userId in context is not a string")
	}

	// Convert string to ObjectID
	userID, err := primitive.ObjectIDFromHex(userIDString)
	if err != nil {
		return primitive.NilObjectID, errors.New("invalid userId format")
	}

	return userID, nil
}
