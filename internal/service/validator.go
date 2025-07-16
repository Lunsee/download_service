package service

import (
	"download_service/internal/models"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

var (
	allowedTypes = initAllowedTypes()
	maxFileSize  = int64(models.MaxFileSize)
)

func initAllowedTypes() map[string]bool {
	types := strings.Split(models.AllowedTypes, ",")
	allowed := make(map[string]bool)
	for _, t := range types {
		allowed[strings.TrimSpace(t)] = true
	}
	return allowed
}

// validate URL
func ValidateURL(url string) (bool, error) {
	ext := strings.ToLower(filepath.Ext(url))
	if ext == "" {
		return false, fmt.Errorf("URL does not contain file extension")
	}

	if !allowedTypes[ext] {
		return false, fmt.Errorf("file type '%s' is not allowed", ext)
	}

	return true, nil
}

// Validate File
func ValidateFile(path string) (bool, error) {
	ext := strings.ToLower(filepath.Ext(path))
	if ext == "" {
		return false, fmt.Errorf("file has no extension")
	}

	if !allowedTypes[ext] {
		return false, fmt.Errorf("file type '%s' is not allowed", ext)
	}

	fileInfo, err := os.Stat(path)
	if err != nil {
		return false, fmt.Errorf("file stat error: %v", err)
	}

	if fileInfo.Size() > maxFileSize {
		return false, fmt.Errorf("file size %d exceeds limit %d", fileInfo.Size(), maxFileSize)
	}

	return true, nil
}
