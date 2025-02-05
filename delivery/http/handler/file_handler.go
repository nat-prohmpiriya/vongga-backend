package handler

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/prohmpiriya_phonumnuaisuk/vongga-platform/vongga-backend/domain"
	"github.com/prohmpiriya_phonumnuaisuk/vongga-platform/vongga-backend/utils"
)

type FileHandler struct {
	fileRepo domain.FileRepository
}

func NewFileHandler(router fiber.Router, fileRepo domain.FileRepository) *FileHandler {
	logger := utils.NewLogger("FileHandler.NewFileHandler")
	logger.LogInput(map[string]interface{}{
		"fileRepo": fileRepo,
	})
	handler := &FileHandler{
		fileRepo: fileRepo,
	}

	router.Post("/upload", handler.Upload)

	return handler
}

func (h *FileHandler) Upload(c *fiber.Ctx) error {
	logger := utils.NewLogger("FileHandler.Upload")

	// Get file from request
	file, err := c.FormFile("file")
	if err != nil {
		logger.LogOutput(nil, fmt.Errorf("error getting file from request: %v", err))
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "file is required",
		})
	}

	logger.LogInput(map[string]interface{}{
		"filename":    file.Filename,
		"size":        file.Size,
		"header":      file.Header,
		"contentType": file.Header.Get("Content-Type"),
	})

	// Validate file type
	contentType := file.Header.Get("Content-Type")
	if !isValidFileType(contentType) {
		err := fmt.Errorf("invalid file type: %s", contentType)
		logger.LogOutput(nil, err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	// Validate file size (max 10MB)
	if file.Size > 10*1024*1024 {
		err := fmt.Errorf("file size too large: %d bytes", file.Size)
		logger.LogOutput(nil, err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	// Open file
	fileData, err := file.Open()
	if err != nil {
		logger.LogOutput(nil, fmt.Errorf("error opening file: %v", err))
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
		logger.LogOutput(nil, fmt.Errorf("error uploading file: %v", err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "error uploading file",
		})
	}

	logger.LogOutput(map[string]interface{}{
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
