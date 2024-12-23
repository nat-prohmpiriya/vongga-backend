package handler

import (
	"github.com/gofiber/fiber/v2"
	domain "github.com/prohmpiriya_phonumnuaisuk/vongga-platform/vongga-backend/domain/file"
)

type FileHandler struct {
	fileRepo domain.FileRepository
}

func NewFileHandler(repo domain.FileRepository) *FileHandler {
	return &FileHandler{
		fileRepo: repo,
	}
}

func (h *FileHandler) Upload(c *fiber.Ctx) error {
	// Get file from form
	file, err := c.FormFile("file")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "No file uploaded",
		})
	}

	// Open the file
	src, err := file.Open()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Error opening file",
		})
	}
	defer src.Close()

	// Create file model
	fileModel := &domain.File{
		FileName:    file.Filename,
		FileSize:    file.Size,
		ContentType: file.Header.Get("Content-Type"),
	}

	// Upload file to Firebase Storage
	err = h.fileRepo.Upload(fileModel, src)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "File uploaded successfully",
		"file":    fileModel,
	})
}
