package middleware

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type ObsvMiddleware struct {
	requestID string
	traceID   string
	start     time.Time
}

func info(obsv ObsvMiddleware) {

}
func warn(obsv ObsvMiddleware) {

}
func error(obsv ObsvMiddleware) {

}
func start(obsv ObsvMiddleware) {

}

func OBSVMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		requestID := c.Get("X-Request-ID")
		c.Locals("requestID", requestID)
		id := uuid.New()
		c.Locals("traceID", id.String())

		return c.Next()
	}
}
