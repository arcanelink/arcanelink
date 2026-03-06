package models

import "time"

// FileMetadata represents file metadata stored in database
type FileMetadata struct {
	FileID      string    `json:"file_id" db:"file_id"`
	Uploader    string    `json:"uploader" db:"uploader"`
	Filename    string    `json:"filename" db:"filename"`
	ContentType string    `json:"content_type" db:"content_type"`
	FileSize    int64     `json:"file_size" db:"file_size"`
	StoragePath string    `json:"-" db:"storage_path"` // Don't expose storage path to clients
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
}

// UploadFileResponse represents the response after uploading a file
type UploadFileResponse struct {
	FileID      string `json:"file_id"`
	Filename    string `json:"filename"`
	ContentType string `json:"content_type"`
	FileSize    int64  `json:"file_size"`
	URL         string `json:"url"` // Download URL
}
