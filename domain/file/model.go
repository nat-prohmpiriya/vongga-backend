package file

import "mime/multipart"

type File struct {
	FileName    string `json:"file_name"`
	FileURL     string `json:"file_url"`
	FileSize    int64  `json:"file_size"`
	ContentType string `json:"content_type"`
}

type FileRepository interface {
	Upload(file *File, fileData multipart.File) error
	GetURL(fileName string) (string, error)
}
