package repository

import (
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"path/filepath"
	"time"

	"cloud.google.com/go/storage"
	firebase "firebase.google.com/go/v4"
	"github.com/prohmpiriya_phonumnuaisuk/vongga-platform/vongga-backend/domain"
	"github.com/prohmpiriya_phonumnuaisuk/vongga-platform/vongga-backend/utils"
	"google.golang.org/api/option"
)

type fileStorage struct {
	bucket     *storage.BucketHandle
	bucketName string
}

func NewFileStorage(credentialsFile string, bucketName string) (domain.FileRepository, error) {
	logger := utils.NewLogger("FileRepository.NewFileStorage")
	logger.LogInput(map[string]string{
		"credentialsFile": credentialsFile,
		"bucketName":      bucketName,
	})

	if bucketName == "" {
		err := fmt.Errorf("bucket name is required")
		logger.LogOutput(nil, err)
		return nil, err
	}

	ctx := context.Background()
	config := &firebase.Config{
		StorageBucket: bucketName,
	}

	opt := option.WithCredentialsFile(credentialsFile)
	app, err := firebase.NewApp(ctx, config, opt)
	if err != nil {
		logger.LogOutput(nil, fmt.Errorf("error initializing firebase app: %v", err))
		return nil, fmt.Errorf("error initializing firebase app: %v", err)
	}

	client, err := app.Storage(ctx)
	if err != nil {
		logger.LogOutput(nil, fmt.Errorf("error getting storage client: %v", err))
		return nil, fmt.Errorf("error getting storage client: %v", err)
	}

	bucket, err := client.DefaultBucket()
	if err != nil {
		logger.LogOutput(nil, fmt.Errorf("error getting bucket: %v", err))
		return nil, fmt.Errorf("error getting bucket: %v", err)
	}

	storage := &fileStorage{
		bucket:     bucket,
		bucketName: bucketName,
	}

	logger.LogOutput(storage, nil)
	return storage, nil
}

func (fs *fileStorage) Upload(file *domain.File, fileData multipart.File) (*domain.File, error) {
	logger := utils.NewLogger("FileRepository.Upload")
	logger.LogInput(map[string]interface{}{
		"fileName":    file.FileName,
		"contentType": file.ContentType,
	})

	ctx := context.Background()

	// Generate unique filename using timestamp
	timestamp := time.Now().UnixNano()
	ext := filepath.Ext(file.FileName)
	uniqueFileName := fmt.Sprintf("%d%s", timestamp, ext)

	obj := fs.bucket.Object(uniqueFileName)
	writer := obj.NewWriter(ctx)

	// Set content type
	writer.ContentType = file.ContentType

	if _, err := io.Copy(writer, fileData); err != nil {
		logger.LogOutput(nil, fmt.Errorf("error copying file to storage: %v", err))
		return nil, fmt.Errorf("error copying file to storage: %v", err)
	}

	if err := writer.Close(); err != nil {
		logger.LogOutput(nil, fmt.Errorf("error closing writer: %v", err))
		return nil, fmt.Errorf("error closing writer: %v", err)
	}

	// Get object attributes
	attrs, err := obj.Attrs(ctx)
	if err != nil {
		logger.LogOutput(nil, fmt.Errorf("error getting object attributes: %v", err))
		return nil, fmt.Errorf("error getting object attributes: %v", err)
	}
	logger.LogOutput(map[string]interface{}{
		"attrs": attrs,
	}, nil)

	// Create file model with URL from upload
	fileModel := &domain.File{
		FileURL:     fmt.Sprintf("https://firebasestorage.googleapis.com/v0/b/%s/o/%s?alt=media", fs.bucketName, uniqueFileName),
		FileName:    uniqueFileName,
		ContentType: file.ContentType,
	}

	logger.LogOutput(map[string]string{
		"fileURL":  fileModel.FileURL,
		"fileName": fileModel.FileName,
	}, nil)
	return fileModel, nil
}
