package utils

import (
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
)

func HandleImageUpload(file multipart.File, header *multipart.FileHeader) (string, error) {
	uploadDir := "./uploads"
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		return "", err
	}

	filename := filepath.Join(uploadDir, header.Filename)
	out, err := os.Create(filename)
	if err != nil {
		return "", err
	}
	defer out.Close()

	_, err = io.Copy(out, file)
	if err != nil {
		return "", err
	}

	// For new uploads, return a web-friendly path
	// This will store "/uploads/filename.jpg" in the database
	webPath := "/uploads/" + header.Filename
	return webPath, nil
}
