package storage

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"time"

	"github.com/arcane/arcanelink/pkg/models"
	"github.com/google/uuid"
)

// FileStorage handles file upload and download operations
type FileStorage struct {
	db          *sql.DB
	storagePath string // Base directory for file storage
}

// NewFileStorage creates a new FileStorage instance
func NewFileStorage(db *sql.DB, storagePath string) (*FileStorage, error) {
	// Create storage directory if it doesn't exist
	if err := os.MkdirAll(storagePath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create storage directory: %w", err)
	}

	return &FileStorage{
		db:          db,
		storagePath: storagePath,
	}, nil
}

// UploadFile saves an uploaded file and stores its metadata
func (fs *FileStorage) UploadFile(uploader string, file multipart.File, header *multipart.FileHeader) (*models.FileMetadata, error) {
	// Generate unique file ID
	fileID := uuid.New().String()

	// Calculate file hash for deduplication (optional)
	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return nil, fmt.Errorf("failed to calculate file hash: %w", err)
	}
	fileHash := hex.EncodeToString(hash.Sum(nil))

	// Reset file pointer
	if _, err := file.Seek(0, 0); err != nil {
		return nil, fmt.Errorf("failed to reset file pointer: %w", err)
	}

	// Create subdirectory based on date (YYYY/MM/DD)
	now := time.Now()
	dateDir := filepath.Join(fs.storagePath, fmt.Sprintf("%d/%02d/%02d", now.Year(), now.Month(), now.Day()))
	if err := os.MkdirAll(dateDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create date directory: %w", err)
	}

	// Save file to disk with hash as filename to avoid duplicates
	storagePath := filepath.Join(dateDir, fileHash+filepath.Ext(header.Filename))
	destFile, err := os.Create(storagePath)
	if err != nil {
		return nil, fmt.Errorf("failed to create destination file: %w", err)
	}
	defer destFile.Close()

	// Copy file content
	if _, err := io.Copy(destFile, file); err != nil {
		os.Remove(storagePath) // Clean up on error
		return nil, fmt.Errorf("failed to save file: %w", err)
	}

	// Store metadata in database
	metadata := &models.FileMetadata{
		FileID:      fileID,
		Uploader:    uploader,
		Filename:    header.Filename,
		ContentType: header.Header.Get("Content-Type"),
		FileSize:    header.Size,
		StoragePath: storagePath,
		CreatedAt:   now,
	}

	query := `INSERT INTO file_storage (file_id, uploader, filename, content_type, file_size, storage_path, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`
	_, err = fs.db.Exec(query, metadata.FileID, metadata.Uploader, metadata.Filename,
		metadata.ContentType, metadata.FileSize, metadata.StoragePath, metadata.CreatedAt)
	if err != nil {
		os.Remove(storagePath) // Clean up on error
		return nil, fmt.Errorf("failed to store file metadata: %w", err)
	}

	return metadata, nil
}

// GetFileMetadata retrieves file metadata by file ID
func (fs *FileStorage) GetFileMetadata(fileID string) (*models.FileMetadata, error) {
	var metadata models.FileMetadata
	query := `SELECT file_id, uploader, filename, content_type, file_size, storage_path, created_at
		FROM file_storage WHERE file_id = $1`
	err := fs.db.QueryRow(query, fileID).Scan(
		&metadata.FileID,
		&metadata.Uploader,
		&metadata.Filename,
		&metadata.ContentType,
		&metadata.FileSize,
		&metadata.StoragePath,
		&metadata.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get file metadata: %w", err)
	}
	return &metadata, nil
}

// GetFilePath returns the storage path for a file
func (fs *FileStorage) GetFilePath(fileID string) (string, error) {
	metadata, err := fs.GetFileMetadata(fileID)
	if err != nil {
		return "", err
	}
	return metadata.StoragePath, nil
}

// DeleteFile removes a file and its metadata
func (fs *FileStorage) DeleteFile(fileID, userID string) error {
	// Get metadata first to check ownership and get storage path
	metadata, err := fs.GetFileMetadata(fileID)
	if err != nil {
		return err
	}

	// Check if user is the uploader
	if metadata.Uploader != userID {
		return fmt.Errorf("unauthorized: user is not the file uploader")
	}

	// Delete from database
	query := `DELETE FROM file_storage WHERE file_id = $1`
	_, err = fs.db.Exec(query, fileID)
	if err != nil {
		return fmt.Errorf("failed to delete file metadata: %w", err)
	}

	// Delete physical file
	if err := os.Remove(metadata.StoragePath); err != nil {
		// Log error but don't fail the operation
		// File might already be deleted or not exist
		fmt.Printf("Warning: failed to delete physical file %s: %v\n", metadata.StoragePath, err)
	}

	return nil
}
