package parser

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"awesomeProject/internal/models"
)

// ParseArchive extracts a zip archive with log files and parses the first
// db_csv file plus optional sharp_an_info file found inside it.
func ParseArchive(path string) (*models.ParsedLog, error) {
	tempDir, err := os.MkdirTemp("", "log-parser-*")
	if err != nil {
		return nil, fmt.Errorf("create temp dir: %w", err)
	}
	defer os.RemoveAll(tempDir)

	if err := unzip(path, tempDir); err != nil {
		return nil, err
	}

	csvPath, sharpPath, err := findLogFiles(tempDir)
	if err != nil {
		return nil, err
	}

	return ParseFiles(csvPath, sharpPath)
}

func unzip(archivePath string, destDir string) error {
	reader, err := zip.OpenReader(archivePath)
	if err != nil {
		return fmt.Errorf("open zip archive: %w", err)
	}
	defer reader.Close()

	for _, file := range reader.File {
		targetPath, err := safeArchivePath(destDir, file.Name)
		if err != nil {
			return err
		}

		if file.FileInfo().IsDir() {
			if err := os.MkdirAll(targetPath, 0o755); err != nil {
				return fmt.Errorf("create archive dir: %w", err)
			}
			continue
		}

		if err := os.MkdirAll(filepath.Dir(targetPath), 0o755); err != nil {
			return fmt.Errorf("create archive parent dir: %w", err)
		}

		if err := extractFile(file, targetPath); err != nil {
			return err
		}
	}

	return nil
}

func safeArchivePath(destDir string, archiveName string) (string, error) {
	cleanName := filepath.Clean(archiveName)
	if filepath.IsAbs(cleanName) || cleanName == "." || strings.HasPrefix(cleanName, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("unsafe archive path: %s", archiveName)
	}

	targetPath := filepath.Join(destDir, cleanName)
	relativePath, err := filepath.Rel(destDir, targetPath)
	if err != nil {
		return "", fmt.Errorf("check archive path: %w", err)
	}
	if relativePath == ".." || strings.HasPrefix(relativePath, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("unsafe archive path: %s", archiveName)
	}

	return targetPath, nil
}

func extractFile(file *zip.File, targetPath string) error {
	source, err := file.Open()
	if err != nil {
		return fmt.Errorf("open archive file %s: %w", file.Name, err)
	}
	defer source.Close()

	target, err := os.OpenFile(targetPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, file.Mode())
	if err != nil {
		return fmt.Errorf("create archive file %s: %w", file.Name, err)
	}
	defer target.Close()

	if _, err := io.Copy(target, source); err != nil {
		return fmt.Errorf("extract archive file %s: %w", file.Name, err)
	}

	return nil
}

func findLogFiles(root string) (string, string, error) {
	var csvPath string
	var sharpPath string

	err := filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			return nil
		}

		name := strings.ToLower(entry.Name())
		switch {
		case strings.HasSuffix(name, ".db_csv"):
			if csvPath != "" {
				return fmt.Errorf("archive contains more than one db_csv file")
			}
			csvPath = path
		case strings.HasSuffix(name, ".sharp_an_info"):
			if sharpPath != "" {
				return fmt.Errorf("archive contains more than one sharp_an_info file")
			}
			sharpPath = path
		}

		return nil
	})
	if err != nil {
		return "", "", fmt.Errorf("find log files in archive: %w", err)
	}

	if csvPath == "" {
		return "", "", fmt.Errorf("archive does not contain db_csv file")
	}

	return csvPath, sharpPath, nil
}
