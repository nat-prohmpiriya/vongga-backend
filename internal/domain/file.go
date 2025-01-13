// internal/domain/file.go
package domain

import (
	"context"
	"io"
)

type File struct {
	BaseModel
	UserID      string `json:"userId"`
	Name        string `json:"name"`
	Size        int64  `json:"size"`
	ContentType string `json:"contentType"`
	URL         string `json:"url"`
}

type FileRepository interface {
	Upload(ctx context.Context, file *File, data io.Reader) (*File, error)
	Delete(ctx context.Context, userID string, fileID string) error
	GetByID(ctx context.Context, fileID string) (*File, error)
	GetByUserID(ctx context.Context, userID string) ([]*File, error)
}

type FileUseCase interface {
	Upload(ctx context.Context, userID string, file io.Reader, filename string, size int64, contentType string) (*File, error)
	Delete(ctx context.Context, userID string, fileID string) error
	GetByID(ctx context.Context, fileID string) (*File, error)
	GetByUserID(ctx context.Context, userID string) ([]*File, error)
}
