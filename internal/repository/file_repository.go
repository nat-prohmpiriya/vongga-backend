// internal/repository/file_repository.go
package repository

import (
	"context"
	"fmt"
	"io"
	"path"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"

	"vongga_api/internal/domain"

	"go.opentelemetry.io/otel/trace"

	"vongga_api/utils"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type fileRepository struct {
	db         *mongo.Collection
	s3Client   *s3.Client
	bucketName string
	baseURL    string
	tracer     trace.Tracer
}

func NewFileRepository(db *mongo.Database, s3Client *s3.Client, bucketName, baseURL string, tracer trace.Tracer) domain.FileRepository {
	return &fileRepository{
		db:         db.Collection("files"),
		s3Client:   s3Client,
		bucketName: bucketName,
		baseURL:    baseURL,
		tracer:     tracer,
	}
}

func (r *fileRepository) Upload(ctx context.Context, file *domain.File, data io.Reader) (*domain.File, error) {
	ctx, span := r.tracer.Start(ctx, "FileRepository.Upload")
	defer span.End()
	logger := utils.NewTraceLogger(span)

	// create
	file.ID = primitive.NewObjectID()

	// 1. Generate unique key for S3
	key := fmt.Sprintf("%s/%s/%s%s",
		strings.Split(file.ContentType, "/")[0], // e.g., "image" or "video"
		file.UserID,
		file.ID.Hex(),
		path.Ext(file.Name),
	)

	logger.Input(map[string]interface{}{
		"filename":    file.Name,
		"size":        file.Size,
		"contentType": file.ContentType,
		"key":         key,
	})

	// 2. Upload to S3
	_, err := r.s3Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(r.bucketName),
		Key:         aws.String(key),
		Body:        data,
		ContentType: aws.String(file.ContentType),
	})
	if err != nil {
		logger.Output("failed to upload file to storage", err)
		return nil, fmt.Errorf("failed to upload file: %w", err)
	}

	// 3. Set file URL
	file.URL = fmt.Sprintf("%s/%s", r.baseURL, key)

	// 4. Save to MongoDB
	file.CreatedAt = time.Now()
	file.UpdatedAt = time.Now()

	_, err = r.db.InsertOne(ctx, file)
	if err != nil {
		logger.Output("failed to save file metadata", err)
		return nil, fmt.Errorf("failed to save file metadata: %w", err)
	}

	logger.Output(map[string]interface{}{
		"mesasge": "file uploaded successfully",
		"file":    file,
	}, nil)
	return file, nil
}

func (r *fileRepository) Delete(ctx context.Context, userID, fileID string) error {
	ctx, span := r.tracer.Start(ctx, "FileRepository.Delete")
	defer span.End()
	logger := utils.NewTraceLogger(span)
	logger.Input(map[string]interface{}{
		"userID": userID,
		"fileID": fileID,
	})

	// 1. Get file info
	file, err := r.GetByID(ctx, fileID)
	if err != nil {
		logger.Output("failed to get file", err)
		return err
	}

	// 2. Check ownership
	if file.UserID != userID {
		logger.Output("unauthorized", nil)
		return fmt.Errorf("unauthorized")
	}

	// 3. Soft delete in MongoDB
	now := time.Now()
	update := bson.M{
		"$set": bson.M{
			"deletedAt": now,
			"isActive":  false,
			"updatedAt": now,
		},
	}
	_, err = r.db.UpdateOne(ctx, bson.M{"_id": fileID}, update)
	if err != nil {
		logger.Output("failed to soft delete file", err)
		return fmt.Errorf("failed to soft delete file: %w", err)
	}

	logger.Output(map[string]interface{}{
		"message": "file soft deleted successfully",
		"file":    file,
	}, nil)
	return nil
}

func (r *fileRepository) GetByID(ctx context.Context, fileID string) (*domain.File, error) {
	ctx, span := r.tracer.Start(ctx, "FileRepository.GetByID")
	defer span.End()
	logger := utils.NewTraceLogger(span)
	logger.Input(map[string]interface{}{
		"fileID": fileID,
	})

	var file domain.File
	err := r.db.FindOne(ctx, bson.M{
		"_id":       fileID,
		"isActive":  true,
		"deletedAt": nil,
	}).Decode(&file)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			logger.Output("file not found", err)
			return nil, fmt.Errorf("file not found")
		}
		logger.Output("failed to get file", err)
		return nil, fmt.Errorf("failed to get file: %w", err)
	}

	logger.Output(map[string]interface{}{
		"message": "file found",
		"file":    file,
	}, nil)
	return &file, nil
}

func (r *fileRepository) GetByUserID(ctx context.Context, userID string) ([]*domain.File, error) {
	ctx, span := r.tracer.Start(ctx, "FileRepository.GetByUserID")
	defer span.End()
	logger := utils.NewTraceLogger(span)
	logger.Input(map[string]interface{}{
		"userID": userID,
	})

	cursor, err := r.db.Find(ctx, bson.M{
		"userId":    userID,
		"isActive":  true,
		"deletedAt": nil,
	})
	if err != nil {
		logger.Output("failed to get files", err)
		return nil, fmt.Errorf("failed to get files: %w", err)
	}
	defer cursor.Close(ctx)

	var files []*domain.File
	if err := cursor.All(ctx, &files); err != nil {
		logger.Output("failed to decode files", err)
		return nil, fmt.Errorf("failed to decode files: %w", err)
	}

	logger.Output(map[string]interface{}{
		"message": "files found",
		"count":   len(files),
	}, nil)
	return files, nil
}
