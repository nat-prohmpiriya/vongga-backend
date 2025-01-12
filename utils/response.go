package utils

import (
	"vongga_api/internal/domain"

	"github.com/gofiber/fiber/v2"
)

// ErrorResponse represents the structure of error responses
type ErrorResponse struct {
	Error string `json:"error"`
}

// SuccessResponse represents the structure of success responses
type SuccessResponse struct {
	Message string `json:"message"`
}

// SendError sends an error response with the given status code and message
func SendError(c *fiber.Ctx, status int, message string) error {
	return c.Status(status).JSON(ErrorResponse{Error: message})
}

// SendSuccess sends a success response with the given message
func SendSuccess(c *fiber.Ctx, message string) error {
	return c.Status(fiber.StatusOK).JSON(SuccessResponse{Message: message})
}

// HandleError handles different types of errors and sends appropriate responses
func HandleError(c *fiber.Ctx, err error) error {
	var status int
	var message string

	switch {
	case err == domain.ErrNotFound:
		status = fiber.StatusNotFound
		message = err.Error()
	case err == domain.ErrUnauthorized:
		status = fiber.StatusUnauthorized
		message = err.Error()
	case err == domain.ErrInvalidInput:
		status = fiber.StatusBadRequest
		message = err.Error()
	case err == domain.ErrInternalError:
		status = fiber.StatusInternalServerError
		message = err.Error()
	default:
		status = fiber.StatusInternalServerError
		message = "Internal server error"
	}

	return SendError(c, status, message)
}
