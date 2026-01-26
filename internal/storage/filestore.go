package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/flakerimi/inceptor/internal/core"
)

type LocalFileStore struct {
	basePath string
}

func NewLocalFileStore(basePath string) (*LocalFileStore, error) {
	// Ensure base directory exists
	if err := os.MkdirAll(basePath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create base directory: %w", err)
	}
	return &LocalFileStore{basePath: basePath}, nil
}

// SaveCrashLog saves the full crash payload to a file
// Returns the relative file path
func (fs *LocalFileStore) SaveCrashLog(ctx context.Context, crash *core.Crash) (string, error) {
	// Create directory structure: {basePath}/{app_id}/{YYYY-MM-DD}/
	dateDir := crash.CreatedAt.Format("2006-01-02")
	dirPath := filepath.Join(fs.basePath, crash.AppID, dateDir)

	if err := os.MkdirAll(dirPath, 0755); err != nil {
		return "", fmt.Errorf("failed to create directory: %w", err)
	}

	// File path: {crash_id}.json
	fileName := fmt.Sprintf("%s.json", crash.ID)
	filePath := filepath.Join(dirPath, fileName)
	relativePath := filepath.Join(crash.AppID, dateDir, fileName)

	// Marshal crash to JSON
	data, err := json.MarshalIndent(crash, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal crash: %w", err)
	}

	// Write file
	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return "", fmt.Errorf("failed to write file: %w", err)
	}

	return relativePath, nil
}

// GetCrashLog retrieves the full crash payload from a file
func (fs *LocalFileStore) GetCrashLog(ctx context.Context, relativePath string) (*core.Crash, error) {
	filePath := filepath.Join(fs.basePath, relativePath)

	data, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	var crash core.Crash
	if err := json.Unmarshal(data, &crash); err != nil {
		return nil, fmt.Errorf("failed to unmarshal crash: %w", err)
	}

	return &crash, nil
}

// DeleteCrashLog deletes a crash log file
func (fs *LocalFileStore) DeleteCrashLog(ctx context.Context, relativePath string) error {
	filePath := filepath.Join(fs.basePath, relativePath)

	if err := os.Remove(filePath); err != nil {
		if os.IsNotExist(err) {
			return nil // Already deleted
		}
		return fmt.Errorf("failed to delete file: %w", err)
	}

	// Try to clean up empty parent directories
	dirPath := filepath.Dir(filePath)
	fs.cleanEmptyDirs(dirPath)

	return nil
}

// DeleteOldLogs deletes all logs older than the specified date for an app
func (fs *LocalFileStore) DeleteOldLogs(ctx context.Context, appID string, before time.Time) (int, error) {
	appDir := filepath.Join(fs.basePath, appID)

	if _, err := os.Stat(appDir); os.IsNotExist(err) {
		return 0, nil
	}

	deleted := 0
	cutoffDate := before.Format("2006-01-02")

	// Walk through date directories
	entries, err := os.ReadDir(appDir)
	if err != nil {
		return 0, fmt.Errorf("failed to read app directory: %w", err)
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		dirName := entry.Name()
		// Check if directory name is a date and is older than cutoff
		if dirName < cutoffDate {
			dirPath := filepath.Join(appDir, dirName)

			// Count files before deletion
			files, err := os.ReadDir(dirPath)
			if err == nil {
				deleted += len(files)
			}

			// Remove entire directory
			if err := os.RemoveAll(dirPath); err != nil {
				return deleted, fmt.Errorf("failed to delete directory %s: %w", dirPath, err)
			}
		}
	}

	return deleted, nil
}

// GetStorageStats returns storage statistics for an app
func (fs *LocalFileStore) GetStorageStats(ctx context.Context, appID string) (*StorageStats, error) {
	stats := &StorageStats{}

	appDir := filepath.Join(fs.basePath, appID)
	if _, err := os.Stat(appDir); os.IsNotExist(err) {
		return stats, nil
	}

	err := filepath.Walk(appDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && filepath.Ext(path) == ".json" {
			stats.TotalFiles++
			stats.TotalSize += info.Size()
		}
		return nil
	})

	return stats, err
}

// cleanEmptyDirs removes empty parent directories up to the base path
func (fs *LocalFileStore) cleanEmptyDirs(dirPath string) {
	for dirPath != fs.basePath && dirPath != "." && dirPath != "/" {
		entries, err := os.ReadDir(dirPath)
		if err != nil || len(entries) > 0 {
			break
		}
		os.Remove(dirPath)
		dirPath = filepath.Dir(dirPath)
	}
}

// ListCrashFiles lists all crash files for an app within a date range
func (fs *LocalFileStore) ListCrashFiles(ctx context.Context, appID string, from, to time.Time) ([]string, error) {
	appDir := filepath.Join(fs.basePath, appID)
	var files []string

	if _, err := os.Stat(appDir); os.IsNotExist(err) {
		return files, nil
	}

	fromDate := from.Format("2006-01-02")
	toDate := to.Format("2006-01-02")

	entries, err := os.ReadDir(appDir)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		dirName := entry.Name()
		// Check if within date range
		if dirName >= fromDate && dirName <= toDate {
			dirPath := filepath.Join(appDir, dirName)
			crashFiles, err := os.ReadDir(dirPath)
			if err != nil {
				continue
			}

			for _, f := range crashFiles {
				if !f.IsDir() && filepath.Ext(f.Name()) == ".json" {
					files = append(files, filepath.Join(appID, dirName, f.Name()))
				}
			}
		}
	}

	return files, nil
}
