package utils

import (
	"fmt"
	"os"
	"path/filepath"
)

// EnsureDirectoryExists проверяет существование директории и создает её при необходимости.
func EnsureDirectoryExists(dir string) error {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if err := os.MkdirAll(dir, os.ModePerm); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}
	return nil
}

// GetMigrationNameFromPath извлекает имя файла миграции из пути.
func GetMigrationNameFromPath(path string) string {
	return filepath.Base(path)
}
