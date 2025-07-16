package service

import (
	"archive/zip"
	"download_service/internal/models"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

func Download(task *models.Task, storageDir string, baseUrl string) (string, error) {

	// dir
	taskDir := filepath.Join(storageDir, task.ID.String())
	if err := os.MkdirAll(taskDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create task directory: %w", err)
	}

	// download
	for i, link := range task.Links {

		isValid, err := ValidateURL(link)
		if err != nil {
			return "", fmt.Errorf("invalid URL %s: %w", link, err)
		}
		if !isValid {
			return "", fmt.Errorf("URL %s is not allowed", link)
		}

		filePath := filepath.Join(taskDir, fmt.Sprintf("file%d%s", i, filepath.Ext(link)))

		if err := downloadFile(link, filePath); err != nil {
			return "", fmt.Errorf("failed to download %s: %w", link, err)
		}

		isValid, err = ValidateFile(filePath)
		if err != nil {
			return "", fmt.Errorf("invalid URL %s: %w", link, err)
		}
		if !isValid {
			return "", fmt.Errorf("URL %s is not allowed", link)
		}

		task.Files = append(task.Files, models.FileInfo{
			OriginalURL: link,
			FilePath:    filePath,
		})
	}

	// zip
	zipPath := filepath.Join(storageDir, task.ID.String()+".zip")
	if err := createZip(task.Files, zipPath); err != nil {
		return "", fmt.Errorf("failed to create zip: %w", err)
	}

	//return filepath.Base(zipPath), nil
	downloadUrl := strings.TrimRight(baseUrl, "/") + "/download/" + task.ID.String() + ".zip"
	return downloadUrl, nil
}

func downloadFile(url, savePath string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	file, err := os.Create(savePath)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = io.Copy(file, resp.Body)
	return err
}

// create zip
func createZip(files []models.FileInfo, outputPath string) error {
	zipFile, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer zipFile.Close()

	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()

	for _, file := range files {
		srcFile, err := os.Open(file.FilePath)
		if err != nil {
			return err
		}
		defer srcFile.Close()

		entry, err := zipWriter.Create(filepath.Base(file.FilePath))
		if err != nil {
			return err
		}

		if _, err := io.Copy(entry, srcFile); err != nil {
			return err
		}
	}
	return nil
}
