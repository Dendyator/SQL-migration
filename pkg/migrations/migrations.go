package migrations

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Dendyator/SQL-migration/internal/config"
	"github.com/Dendyator/SQL-migration/internal/migration"
)

// LoadMigrationsFromDir загружает все миграции из указанного каталога.
func LoadMigrationsFromDir(cfg *config.Config) ([]*migration.Migration, error) {
	// Убедимся, что директория миграций существует
	if err := ensureMigrationDirectoryExists(cfg.MigrationDir); err != nil {
		return nil, fmt.Errorf("failed to ensure migration directory exists: %w", err)
	}

	files, err := listMigrationFiles(cfg.MigrationDir, cfg.MigrationType)
	if err != nil {
		return nil, err
	}

	migrations := make([]*migration.Migration, 0, len(files))
	for _, file := range files {
		migration, err := migration.ParseMigrationFile(filepath.Join(cfg.MigrationDir, file), cfg.MigrationType)
		if err != nil {
			return nil, fmt.Errorf("failed to parse migration file %s: %w", file, err)
		}
		migrations = append(migrations, migration)
	}

	return migrations, nil
}

// ensureMigrationDirectoryExists проверяет существование директории миграций и создает её при необходимости.
func ensureMigrationDirectoryExists(dir string) error {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if err := os.MkdirAll(dir, os.ModePerm); err != nil {
			return fmt.Errorf("failed to create migration directory: %w", err)
		}
	}
	return nil
}

// listMigrationFiles возвращает список файлов миграций в указанной директории.
func listMigrationFiles(dir, migrationType string) ([]string, error) {
	files, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("failed to read migration directory: %w", err)
	}

	var migrationFiles []string
	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), "."+migrationType) {
			migrationFiles = append(migrationFiles, file.Name())
		}
	}

	return migrationFiles, nil
}
