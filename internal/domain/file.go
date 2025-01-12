package domain

import (
	"context"
	"mime/multipart"
)

type File struct {
	FileName    string
	FileURL     string
	ContentType string
}

type FileRepository interface {
	Upload(ctx context.Context, file *File, fileData multipart.File) (*File, error)
}
