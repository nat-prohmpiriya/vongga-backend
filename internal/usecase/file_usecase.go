package usecase

import (
	"context"
	"fmt"
	"io"

	"go.opentelemetry.io/otel/trace"

	"vongga_api/internal/domain"
	"vongga_api/utils"
)

type fileUseCase struct {
	repo   domain.FileRepository
	tracer trace.Tracer
}

func NewFileUseCase(repo domain.FileRepository, tracer trace.Tracer) domain.FileUseCase {
	return &fileUseCase{
		repo:   repo,
		tracer: tracer,
	}
}

func (uc *fileUseCase) Upload(ctx context.Context, userID string, file io.Reader, filename string, size int64, contentType string) (*domain.File, error) {
	ctx, span := uc.tracer.Start(ctx, "FileUseCase.Upload")
	defer span.End()
	logger := utils.NewTraceLogger(span)
	logger.Input(map[string]interface{}{
		"userID":      userID,
		"filename":    filename,
		"size":        size,
		"contentType": contentType,
	})

	// 1. Validate file
	if err := uc.validateFile(size, contentType); err != nil {
		logger.Output("file validation failed", err)
		return nil, err
	}

	// 2. Create file model
	fileModel := &domain.File{
		UserID:      userID,
		Name:        filename,
		Size:        size,
		ContentType: contentType,
	}

	// 3. Upload file
	uploadedFile, err := uc.repo.Upload(ctx, fileModel, file)
	if err != nil {
		logger.Output("failed to upload file", err)
		return nil, fmt.Errorf("failed to upload file: %w", err)
	}

	logger.Output(map[string]interface{}{
		"message": "file uploaded successfully",
		"file":    uploadedFile,
	}, nil)
	return uploadedFile, nil
}

func (uc *fileUseCase) Delete(ctx context.Context, userID string, fileID string) error {
	ctx, span := uc.tracer.Start(ctx, "FileUseCase.Delete")
	defer span.End()
	logger := utils.NewTraceLogger(span)
	logger.Input(map[string]interface{}{
		"userID": userID,
		"fileID": fileID,
	})

	err := uc.repo.Delete(ctx, userID, fileID)
	if err != nil {
		logger.Output("failed to delete file", err)
		return fmt.Errorf("failed to delete file: %w", err)
	}

	logger.Output("file deleted successfully", nil)
	return nil
}

func (uc *fileUseCase) GetByID(ctx context.Context, fileID string) (*domain.File, error) {
	ctx, span := uc.tracer.Start(ctx, "FileUseCase.GetByID")
	defer span.End()
	logger := utils.NewTraceLogger(span)
	logger.Input(map[string]interface{}{
		"fileID": fileID,
	})

	file, err := uc.repo.GetByID(ctx, fileID)
	if err != nil {
		logger.Output("failed to get file", err)
		return nil, fmt.Errorf("failed to get file: %w", err)
	}

	logger.Output(map[string]interface{}{
		"message": "file retrieved successfully",
		"file":    file,
	}, nil)
	return file, nil
}

func (uc *fileUseCase) GetByUserID(ctx context.Context, userID string) ([]*domain.File, error) {
	ctx, span := uc.tracer.Start(ctx, "FileUseCase.GetByUserID")
	defer span.End()
	logger := utils.NewTraceLogger(span)
	logger.Input(map[string]interface{}{
		"userID": userID,
	})

	files, err := uc.repo.GetByUserID(ctx, userID)
	if err != nil {
		logger.Output("failed to get files", err)
		return nil, fmt.Errorf("failed to get files: %w", err)
	}

	logger.Output(map[string]interface{}{
		"message":     "files retrieved successfully",
		"files_count": len(files),
	}, nil)
	return files, nil
}

func (uc *fileUseCase) validateFile(size int64, contentType string) error {
	// Add file validation logic here
	// For example:
	// 1. Check file size
	if size > 10*1024*1024 { // 10MB
		return fmt.Errorf("file size exceeds maximum limit of 10MB")
	}

	// 2. Check content type
	allowedTypes := map[string]bool{
		"image/jpeg": true,
		"image/png":  true,
		"image/gif":  true,
		"image/webp": true,
		"video/mp4":  true,
	}
	if !allowedTypes[contentType] {
		return fmt.Errorf("unsupported file type: %s", contentType)
	}

	return nil
}
