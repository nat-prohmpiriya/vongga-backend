package http

import (
	"github.com/gofiber/fiber/v2"
	"github.com/prohmpiriya_phonumnuaisuk/vongga-platform/vongga-backend/domain/file"
)

type FileHandler struct {
	fileRepo file.FileRepository
}

func NewFileHandler(repo file.FileRepository) *FileHandler {
	return &FileHandler{
		fileRepo: repo,
	}
}

func (h *FileHandler) Upload(c *fiber.Ctx) error {
	// Get file from form
	uploadedFile, err := c.FormFile("file")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "No file uploaded",
		})
	}

	// Open the file
	src, err := uploadedFile.Open()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Error opening file",
		})
	}
	defer src.Close()

	// Create file model
	fileModel := &file.File{
		FileName:    uploadedFile.Filename,
		FileSize:    uploadedFile.Size,
		ContentType: uploadedFile.Header.Get("Content-Type"),
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
