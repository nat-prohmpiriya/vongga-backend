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
	"github.com/prohmpiriya_phonumnuaisuk/vongga-platform/vongga-backend/domain/file"
	"google.golang.org/api/option"
)

type firebaseStorage struct {
	bucket     *storage.BucketHandle
	bucketName string
}

func NewFirebaseStorage(credentialsFile string, bucketName string) (file.FileRepository, error) {
	ctx := context.Background()
	config := &firebase.Config{
		StorageBucket: bucketName,
	}
	
	opt := option.WithCredentialsFile(credentialsFile)
	app, err := firebase.NewApp(ctx, config, opt)
	if err != nil {
		return nil, fmt.Errorf("error initializing firebase app: %v", err)
	}

	client, err := app.Storage(ctx)
	if err != nil {
		return nil, fmt.Errorf("error getting storage client: %v", err)
	}

	bucket, err := client.DefaultBucket()
	if err != nil {
		return nil, fmt.Errorf("error getting bucket: %v", err)
	}

	return &firebaseStorage{
		bucket:     bucket,
		bucketName: bucketName,
	}, nil
}

func (fs *firebaseStorage) Upload(file *file.File, fileData multipart.File) error {
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
		return fmt.Errorf("error copying file to storage: %v", err)
	}
	
	if err := writer.Close(); err != nil {
		return fmt.Errorf("error closing writer: %v", err)
	}

	// Update file URL
	file.FileURL = fmt.Sprintf("https://storage.googleapis.com/%s/%s", fs.bucketName, uniqueFileName)
	file.FileName = uniqueFileName
	
	return nil
}

func (fs *firebaseStorage) GetURL(fileName string) (string, error) {
	return fmt.Sprintf("https://storage.googleapis.com/%s/%s", fs.bucketName, fileName), nil
}
