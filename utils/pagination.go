package utils

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
)

const (
	DefaultLimit  = 10
	DefaultOffset = 0
)

// FindPaginationParams extracts limit and offset from query parameters
func FindPaginationParams(c *fiber.Ctx) (limit, offset int) {
	// Find limit from query parameter, default to DefaultLimit if not provided
	limitStr := c.Query("limit", strconv.Itoa(DefaultLimit))
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 0 {
		limit = DefaultLimit
	}

	// Find offset from query parameter, default to DefaultOffset if not provided
	offsetStr := c.Query("offset", strconv.Itoa(DefaultOffset))
	offset, err = strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		offset = DefaultOffset
	}

	return limit, offset
}
