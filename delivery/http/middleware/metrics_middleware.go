package middleware

import (
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/prohmpiriya_phonumnuaisuk/vongga-platform/vongga-backend/utils"
)

func MetricsMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()

		// ทำ request handling
		err := c.Next()

		// วัด duration หลังจบ request
		duration := time.Since(start).Seconds()

		// เก็บ metrics
		statusCode := c.Response().StatusCode()
		path := c.Path()
		method := c.Method()

		// Request counter
		utils.HttpRequestTotal.WithLabelValues(
			method,
			path,
			strconv.Itoa(statusCode),
		).Inc()

		// Duration histogram
		utils.HttpRequestDuration.WithLabelValues(
			method,
			path,
		).Observe(duration)

		// Error counter
		if statusCode >= 400 {
			utils.ErrorTotal.WithLabelValues("http_error").Inc()
		}

		return err
	}
}
