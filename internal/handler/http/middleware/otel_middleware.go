package middleware

import (
	"github.com/gofiber/fiber/v2"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
)

func OtelMiddleware(serviceName string) fiber.Handler {
	tracer := otel.Tracer(serviceName)

	return func(c *fiber.Ctx) error {
		spanName := c.Path()
		ctx, span := tracer.Start(c.Context(), spanName)
		defer span.End()

		// Add basic HTTP attributes to the span
		span.SetAttributes(
			attribute.String("http.method", c.Method()),
			attribute.String("http.route", c.Path()),
			attribute.String("http.url", c.OriginalURL()),
			attribute.String("http.host", c.Hostname()),
		)

		// Store the span context in fiber context
		c.Locals("ctx", ctx)

		// Continue with the next middleware/handler
		err := c.Next()

		// Add response status to span
		span.SetAttributes(attribute.Int("http.status_code", c.Response().StatusCode()))

		// If there was an error, record it
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
		} else {
			span.SetStatus(codes.Ok, "")
		}

		return err
	}
}
