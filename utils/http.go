package utils

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
)

// FindQueryInt gets an integer value from query parameters with a default value
func FindQueryInt(c *fiber.Ctx, key string, defaultValue int) int {
	value := c.Query(key)
	if value == "" {
		return defaultValue
	}

	intValue, err := strconv.Atoi(value)
	if err != nil {
		return defaultValue
	}

	return intValue
}
