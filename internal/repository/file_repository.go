package repository

import (
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"path/filepath"
	"time"

	"vongga-api/internal/domain"
	"vongga-api/utils"

	"cloud.google.com/go/storage"
	firebase "firebase.google.com/go/v4"
	"google.golang.org/api/option"

	"go.opentelemetry.io/otel/trace"
)

type fileStorage struct {
	bucket     *storage.BucketHandle
	bucketName string
	tracer     trace.Tracer
}

func NewFileStorage(credentialsFile string, bucketName string, tracer trace.Tracer) (domain.FileRepository, error) {
	ctx, span := tracer.Start(context.Background(), "FileRepository.NewFileStorage")
	defer span.End()
	logger := utils.NewTraceLogger(span)
	logger.Input(map[string]string{
		"credentialsFile": credentialsFile,
		"bucketName":      bucketName,
	})

	if bucketName == "" {
		err := fmt.Errorf("bucket name is required")
		logger.Output("bucket name is required 0", err)
		return nil, err
	}

	config := &firebase.Config{
		StorageBucket: bucketName,
	}

	opt := option.WithCredentialsFile(credentialsFile)
	app, err := firebase.NewApp(ctx, config, opt)
	if err != nil {
		logger.Output("error initializing firebase app 1", err)
		return nil, fmt.Errorf("error initializing firebase app: %v", err)
	}

	client, err := app.Storage(ctx)
	if err != nil {
		logger.Output("error getting storage clien 2", err)
		return nil, fmt.Errorf("error getting storage client: %v", err)
	}

	bucket, err := client.DefaultBucket()
	if err != nil {
		logger.Output("error getting bucket 3", err)
		return nil, fmt.Errorf("error getting bucket: %v", err)
	}

	storage := &fileStorage{
		bucket:     bucket,
		bucketName: bucketName,
	}

	logger.Output(storage, nil)
	return storage, nil
}

func (fs *fileStorage) Upload(ctx context.Context, file *domain.File, fileData multipart.File) (*domain.File, error) {
	ctx, span := fs.tracer.Start(ctx, "FileRepository.Upload")
	defer span.End()
	logger := utils.NewTraceLogger(span)
	logger.Input(map[string]interface{}{
		"fileName":    file.FileName,
		"contentType": file.ContentType,
	})

	// Generate unique filename using timestamp
	timestamp := time.Now().UnixNano()
	ext := filepath.Ext(file.FileName)
	uniqueFileName := fmt.Sprintf("%d%s", timestamp, ext)

	obj := fs.bucket.Object(uniqueFileName)
	writer := obj.NewWriter(ctx)

	// Set content type
	writer.ContentType = file.ContentType

	if _, err := io.Copy(writer, fileData); err != nil {
		logger.Output("error copying file to storage 1", err)
		return nil, fmt.Errorf("error copying file to storage: %v", err)
	}

	if err := writer.Close(); err != nil {
		logger.Output("error closing writer 2", err)
		return nil, fmt.Errorf("error closing writer: %v", err)
	}

	// Find object attributes
	attrs, err := obj.Attrs(ctx)
	if err != nil {
		logger.Output("error getting object attributes 3", err)
		return nil, fmt.Errorf("error getting object attributes: %v", err)
	}
	logger.Output(map[string]interface{}{
		"attrs": attrs,
	}, nil)

	// Create file model with URL from upload
	fileModel := &domain.File{
		FileURL:     fmt.Sprintf("https://firebasestorage.googleapis.com/v0/b/%s/o/%s?alt=media", fs.bucketName, uniqueFileName),
		FileName:    uniqueFileName,
		ContentType: file.ContentType,
	}

	logger.Output(map[string]string{
		"fileURL":  fileModel.FileURL,
		"fileName": fileModel.FileName,
	}, nil)
	return fileModel, nil
}
