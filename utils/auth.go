package utils

import (
	"errors"

	"vongga_api/internal/domain"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func FindUserIDFromContext(c *fiber.Ctx) (primitive.ObjectID, error) {
	userValue := c.Locals("userId")
	if userValue == nil {
		return primitive.NilObjectID, errors.New("user not found in context")
	}

	// ถ้าเป็น *domain.Claims
	if claims, ok := userValue.(*domain.Claims); ok {
		// แปลง claims.UserID (string) เป็น ObjectID
		userID, err := primitive.ObjectIDFromHex(claims.UserID)
		if err != nil {
			return primitive.NilObjectID, errors.New("invalid user ID format in claims")
		}
		return userID, nil
	}

	// fallback cases...
	if userID, ok := userValue.(primitive.ObjectID); ok {
		return userID, nil
	}

	if userIDStr, ok := userValue.(string); ok {
		userID, err := primitive.ObjectIDFromHex(userIDStr)
		if err != nil {
			return primitive.NilObjectID, errors.New("invalid user format")
		}
		return userID, nil
	}

	return primitive.NilObjectID, errors.New("user in context is not a valid format")
}
