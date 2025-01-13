// internal/handler/file_handler.go
package handler

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
	"go.opentelemetry.io/otel/trace"

	"vongga_api/internal/domain"
	"vongga_api/utils"
)

type fileHandler struct {
	useCase domain.FileUseCase
	tracer  trace.Tracer
}

func NewFileHandler(router fiber.Router, useCase domain.FileUseCase, tracer trace.Tracer) *fileHandler {

	handler := &fileHandler{
		useCase: useCase,
		tracer:  tracer,
	}

	router.Post("/upload", handler.Upload)
	router.Delete("/:id", handler.Delete)
	router.Get("/:id", handler.GetByID)
	router.Get("/user", handler.GetByUserID)

	return handler
}

func (h *fileHandler) Upload(c *fiber.Ctx) error {
	ctx, span := h.tracer.Start(c.Context(), "FileHandler.Upload")
	defer span.End()
	logger := utils.NewTraceLogger(span)

	// 1. Get user ID from context
	userID := c.Locals("userId").(string)

	// 2. Get file from request
	file, err := c.FormFile("file")
	if err != nil {
		logger.Output("failed to get file from request", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fmt.Sprintf("failed to get file: %v", err),
		})
	}

	// 3. Open file
	fileContent, err := file.Open()
	if err != nil {
		logger.Output("failed to open file", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("failed to open file: %v", err),
		})
	}
	defer fileContent.Close()

	// 4. Upload file
	uploadedFile, err := h.useCase.Upload(ctx, userID, fileContent, file.Filename, file.Size, file.Header.Get("Content-Type"))
	if err != nil {
		logger.Output("failed to upload file", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("failed to upload file: %v", err),
		})
	}

	logger.Output(map[string]interface{}{
		"message": "file uploaded successfully",
		"file":    uploadedFile,
	}, nil)

	return c.Status(fiber.StatusCreated).JSON(uploadedFile)
}

func (h *fileHandler) Delete(c *fiber.Ctx) error {
	ctx, span := h.tracer.Start(c.Context(), "FileHandler.Delete")
	defer span.End()
	logger := utils.NewTraceLogger(span)

	// 1. Get user ID from context
	userID := c.Locals("userId").(string)

	// 2. Get file ID from params
	fileID := c.Params("id")
	if fileID == "" {
		logger.Output("missing file id", nil)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "missing file id",
		})
	}

	// 3. Delete file
	err := h.useCase.Delete(ctx, userID, fileID)
	if err != nil {
		logger.Output("failed to delete file", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("failed to delete file: %v", err),
		})
	}

	logger.Output("file deleted successfully", nil)
	return c.SendStatus(fiber.StatusNoContent)
}

func (h *fileHandler) GetByID(c *fiber.Ctx) error {
	ctx, span := h.tracer.Start(c.Context(), "FileHandler.GetByID")
	defer span.End()
	logger := utils.NewTraceLogger(span)

	// 1. Get file ID from params
	fileID := c.Params("id")
	if fileID == "" {
		logger.Output("missing file id", nil)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "missing file id",
		})
	}

	// 2. Get file
	file, err := h.useCase.GetByID(ctx, fileID)
	if err != nil {
		logger.Output("failed to get file", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("failed to get file: %v", err),
		})
	}

	logger.Output(map[string]interface{}{
		"message": "file found",
		"file":    file,
	}, nil)

	return c.JSON(file)
}

func (h *fileHandler) GetByUserID(c *fiber.Ctx) error {
	ctx, span := h.tracer.Start(c.Context(), "FileHandler.GetByUserID")
	defer span.End()
	logger := utils.NewTraceLogger(span)

	// 1. Get user ID from context
	userID := c.Locals("userId").(string)

	// 2. Get files
	files, err := h.useCase.GetByUserID(ctx, userID)
	if err != nil {
		logger.Output("failed to get files", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("failed to get files: %v", err),
		})
	}

	logger.Output(map[string]interface{}{
		"message": "files found",
		"count":   len(files),
	}, nil)

	return c.JSON(files)
}
