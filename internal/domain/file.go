package domain

import "mime/multipart"

type File struct {
	FileName    string 
	FileURL     string 
	ContentType string 
}

type FileRepository interface {
	Upload(file *File, fileData multipart.File) (*File, error)
}
