package middleware

import (
	"fmt"
	"time"
	"vongga_api/utils"

	"github.com/gofiber/fiber/v2"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

func OtelMiddleware(serviceName string) fiber.Handler {
	tracer := otel.Tracer(serviceName)

	return func(c *fiber.Ctx) error {
		// สร้าง span และ context
		spanName := c.Path()
		ctx, span := tracer.Start(c.Context(), spanName)
		defer span.End()
		requestInfo := map[string]interface{}{
			// Request Basic Info
			"http.method":   c.Method(),
			"http.url":      c.OriginalURL(),
			"http.path":     c.Path(),
			"http.host":     c.Hostname(),
			"http.protocol": c.Protocol(),

			// Client Info
			"http.client_ip":  c.IP(),
			"http.user_agent": c.Get("User-Agent"),
			"http.referer":    c.Get("Referer"),

			// Request ID & Correlation
			"request.id": c.Get("X-Request-ID"),
			"trace.id":   c.Get("X-Trace-ID"),

			// Performance
			"request.start_time": time.Now(),

			// Request Size
			"request.content_length": len(c.Body()),
			"request.content_type":   c.Get("Content-Type"),
		}

		// สร้าง span attributes
		span = trace.SpanFromContext(ctx)
		for key, value := range requestInfo {
			span.SetAttributes(attribute.String(key, fmt.Sprint(value)))
		}

		span.AddEvent(fmt.Sprintf("INPUT # $%s", utils.ToJSONString(requestInfo)))

		// เก็บไว้ใช้ต่อใน handler
		c.Locals("requestInfo", requestInfo)

		// Execute handler
		err := c.Next()

		// Response Info
		responseInfo := map[string]interface{}{
			"response.status":         c.Response().StatusCode(),
			"response.content_type":   c.GetRespHeader("Content-Type"),
			"response.content_length": len(c.Response().Body()),
			"response.time":           time.Since(requestInfo["request.start_time"].(time.Time)),
		}
		span.AddEvent(fmt.Sprintf("OUTPUT # $%s", utils.ToJSONString(responseInfo)))

		// Add response attributes to span
		for key, value := range responseInfo {
			span.SetAttributes(attribute.String(key, fmt.Sprint(value)))
		}

		return err
	}
}
