package utils

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"socialNetwork/pkg/cloud"
	"socialNetwork/pkg/db"
	"strings"
	"time"
)

const (
	MaxImageSize = 5 * 1024 * 1024 // 5MB
)

func ImageSizeLimit() int64 {
	if os.Getenv("VERCEL") != "" {
		return 4 * 1024 * 1024
	}
	return MaxImageSize
}

var allowedImageTypes = map[string]bool{
	"image/jpeg": true,
	"image/jpg":  true,
	"image/png":  true,
	"image/gif":  true,
}

func ValidateImageFile(header *multipart.FileHeader) error {
	// Check file size
	if header.Size > ImageSizeLimit() {
		return fmt.Errorf("image file size exceeds %dMB limit", ImageSizeLimit()/(1024*1024))
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

func HandleImageUpload(file multipart.File, header *multipart.FileHeader, owner ...int) (string, error) {
	// Validate the image file first
	if err := ValidateImageFile(header); err != nil {
		return "", err
	}

	config, _, err := image.DecodeConfig(file)
	if err != nil || config.Width <= 0 || config.Height <= 0 || config.Width > 16000 || config.Height > 16000 || config.Width*config.Height > 40000000 {
		return "", errors.New("invalid or oversized image dimensions")
	}
	if _, err = file.Seek(0, 0); err != nil {
		return "", err
	}
	uniqueFilename := generateUniqueFilename(header.Filename)
	webPath := "/uploads/" + uniqueFilename
	var cleanup func()
	if cloud.Enabled() {
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()
		if err = cloud.Upload(ctx, uniqueFilename, header.Header.Get("Content-Type"), io.LimitReader(file, MaxImageSize+1)); err != nil {
			return "", err
		}
		cleanup = func() { _ = cloud.Delete(context.Background(), uniqueFilename) }
	} else {
		if err = os.MkdirAll("uploads", 0755); err != nil {
			return "", err
		}
		destination := filepath.Join("uploads", uniqueFilename)
		out, openErr := os.Create(destination)
		if openErr != nil {
			return "", openErr
		}
		_, err = io.Copy(out, io.LimitReader(file, MaxImageSize+1))
		closeErr := out.Close()
		if err == nil {
			err = closeErr
		}
		if err != nil {
			os.Remove(destination)
			return "", err
		}
		cleanup = func() { _ = os.Remove(destination) }
	}
	if len(owner) > 0 && owner[0] > 0 {
		if _, err = db.DBInstance.DB.Exec(`INSERT INTO media_uploads(path,owner_id) VALUES (?,?)`, webPath, owner[0]); err != nil {
			cleanup()
			return "", err
		}
	}

	return webPath, nil
}
