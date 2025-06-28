package utils

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"
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

// Generate a unique filename
func generateUniqueFilename(originalFilename string) string {
	// Get file extension
	ext := strings.ToLower(filepath.Ext(originalFilename))

	// Generate timestamp
	timestamp := time.Now().Unix()

	// Generate random bytes
	randomBytes := make([]byte, 8)
	rand.Read(randomBytes)
	randomString := hex.EncodeToString(randomBytes)

	// Create unique filename: timestamp_randomstring.ext
	return fmt.Sprintf("%d_%s%s", timestamp, randomString, ext)
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

	// Generate unique filename instead of using original
	uniqueFilename := generateUniqueFilename(header.Filename)
	filepath := filepath.Join(uploadDir, uniqueFilename)

	out, err := os.Create(filepath)
	if err != nil {
		return "", fmt.Errorf("failed to create file: %v", err)
	}
	defer out.Close()

	_, err = io.Copy(out, file)
	if err != nil {
		return "", fmt.Errorf("failed to save file: %v", err)
	}

	// Return web-friendly path with the unique filename
	webPath := "/uploads/" + uniqueFilename
	return webPath, nil
}
