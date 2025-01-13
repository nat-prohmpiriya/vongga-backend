package domain

import (
	"context"
	"mime/multipart"
)

type StorageProvider interface {
	UploadImage(ctx context.Context, file *File, data multipart.File) (*File, error)
	UploadVideo(ctx context.Context, file *File, data multipart.File) (*File, error)
}
