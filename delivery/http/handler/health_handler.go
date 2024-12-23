package handler

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/mongo"
)

type HealthHandler struct {
	mongoDB     *mongo.Database
	redisClient *redis.Client
}

// NewHealthHandler creates a new health handler
func NewHealthHandler(mongoDB *mongo.Database, redisClient *redis.Client) *HealthHandler {
	return &HealthHandler{
		mongoDB:     mongoDB,
		redisClient: redisClient,
	}
}

// Health godoc
// @Summary Check service health
// @Description Check the health of the service and its dependencies
// @Tags health
// @Accept json
// @Produce json
// @Success 200 {object} HealthResponse
// @Failure 503 {object} ErrorResponse
// @Router /health [get]
func (h *HealthHandler) Health(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(c.Context(), 5*time.Second)
	defer cancel()

	// Check MongoDB connection
	mongoStatus := "up"
	if err := h.mongoDB.Client().Ping(ctx, nil); err != nil {
		mongoStatus = "down"
	}

	// Check Redis connection
	redisStatus := "up"
	if err := h.redisClient.Ping(ctx).Err(); err != nil {
		redisStatus = "down"
	}

	// Overall status is down if any dependency is down
	status := "healthy"
	statusCode := fiber.StatusOK
	if mongoStatus == "down" || redisStatus == "down" {
		status = "unhealthy"
		statusCode = fiber.StatusServiceUnavailable
	}

	return c.Status(statusCode).JSON(fiber.Map{
		"status": status,
		"timestamp": time.Now().Format(time.RFC3339),
		"services": fiber.Map{
			"mongodb": mongoStatus,
			"redis":   redisStatus,
		},
	})
}

// HealthResponse represents the health check response
type HealthResponse struct {
	Status    string            `json:"status" example:"healthy"`
	Timestamp string            `json:"timestamp" example:"2024-12-23T07:02:21Z"`
	Services  map[string]string `json:"services"`
}
