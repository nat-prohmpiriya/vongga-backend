package handler

import (
	"fmt"

	"vongga-api/internal/domain"
	"vongga-api/utils"

	"github.com/gofiber/fiber/v2"
)

type FileHandler struct {
	fileRepo domain.FileRepository
}

func NewFileHandler(router fiber.Router, fileRepo domain.FileRepository) *FileHandler {
	logger := utils.NewTraceLogger("FileHandler.NewFileHandler")
	logger.Input(map[string]interface{}{
		"fileRepo": fileRepo,
	})
	handler := &FileHandler{
		fileRepo: fileRepo,
	}

	router.Post("/upload", handler.Upload)

	return handler
}

func (h *FileHandler) Upload(c *fiber.Ctx) error {
	logger := utils.NewTraceLogger("FileHandler.Upload")

	// Find file from request
	file, err := c.FormFile("file")
	if err != nil {
		logger.Output(nil, fmt.Errorf("error getting file from request: %v", err))
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "file is required",
		})
	}

	logger.Input(map[string]interface{}{
		"filename":    file.Filename,
		"size":        file.Size,
		"header":      file.Header,
		"contentType": file.Header.Find("Content-Type"),
	})

	// Validate file type
	contentType := file.Header.Find("Content-Type")
	if !isValidFileType(contentType) {
		err := fmt.Errorf("invalid file type: %s", contentType)
		logger.Output(nil, err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	// Validate file size (max 10MB)
	if file.Size > 10*1024*1024 {
		err := fmt.Errorf("file size too large: %d bytes", file.Size)
		logger.Output(nil, err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	// Open file
	fileData, err := file.Open()
	if err != nil {
		logger.Output(nil, fmt.Errorf("error opening file: %v", err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "error opening file",
		})
	}
	defer fileData.Close()

	// Create file model
	fileModel := &domain.File{
		FileName:    file.Filename,
		ContentType: contentType,
	}

	// Upload file
	uploadedFile, err := h.fileRepo.Upload(fileModel, fileData)
	if err != nil {
		logger.Output(nil, fmt.Errorf("error uploading file: %v", err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "error uploading file",
		})
	}

	logger.Output(map[string]interface{}{
		"fileURL":  uploadedFile.FileURL,
		"fileName": uploadedFile.FileName,
	}, nil)

	return c.JSON(fiber.Map{
		"url":      uploadedFile.FileURL,
		"fileName": uploadedFile.FileName,
	})
}

func isValidFileType(contentType string) bool {
	validTypes := map[string]bool{
		"image/jpeg": true,
		"image/png":  true,
		"image/gif":  true,
		"image/webp": true,
	}

	return validTypes[contentType]
}
