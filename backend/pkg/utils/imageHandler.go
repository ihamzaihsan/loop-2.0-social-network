package utils

import (
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
)

const (
	MaxImageSize = 5 * 1024 * 1024 // 5MB
)

var allowedImageTypes = map[string]bool{
	"image/jpeg": true,
	"image/jpg":  true,
	"image/png":  true,
	"image/gif":  true,
}

func ValidateImageFile(header *multipart.FileHeader) error {
	// Check file size
	if header.Size > MaxImageSize {
		return errors.New("image file size exceeds 5MB limit")
	}

	// Check file type by MIME type
	contentType := header.Header.Get("Content-Type")
	if !allowedImageTypes[contentType] {
		return errors.New("invalid image format. Only JPEG, PNG, and GIF files are allowed")
	}

	// Additional check by file extension
	ext := strings.ToLower(filepath.Ext(header.Filename))
	validExtensions := map[string]bool{
		".jpg":  true,
		".jpeg": true,
		".png":  true,
		".gif":  true,
	}

	if !validExtensions[ext] {
		return errors.New("invalid image format. Only JPEG, PNG, and GIF files are allowed")
	}

	return nil
}

func HandleImageUpload(file multipart.File, header *multipart.FileHeader) (string, error) {
	// Validate the image file first
	if err := ValidateImageFile(header); err != nil {
		return "", err
	}

	uploadDir := "./uploads"
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create upload directory: %v", err)
	}

	filename := filepath.Join(uploadDir, header.Filename)
	out, err := os.Create(filename)
	if err != nil {
		return "", fmt.Errorf("failed to create file: %v", err)
	}
	defer out.Close()

	_, err = io.Copy(out, file)
	if err != nil {
		return "", fmt.Errorf("failed to save file: %v", err)
	}

	// For new uploads, return a web-friendly path
	// This will store "/uploads/filename.jpg" in the database
	webPath := "/uploads/" + header.Filename
	return webPath, nil
}
