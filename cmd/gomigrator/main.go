package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/Dendyator/SQL-migration/internal/config"
	"github.com/Dendyator/SQL-migration/internal/utils"
	"github.com/Dendyator/SQL-migration/pkg/api"
)

func main() {
	var dsn, migrationDir, migrationType string
	flag.StringVar(&dsn, "dsn", "", "Database connection string")
	flag.StringVar(&migrationDir, "dir", "./migrations", "Directory with migration files")
	flag.StringVar(&migrationType, "type", "sql", "Migration type (go/sql)")
	flag.Parse()

	cfg := &config.Config{
		DSN:           dsn,
		MigrationDir:  migrationDir,
		MigrationType: migrationType,
	}

	if err := cfg.Validate(); err != nil {
		log.Fatalf("Invalid configuration: %v", err)
	}

	// Убедимся, что директория миграций существует
	if err := utils.EnsureDirectoryExists(cfg.MigrationDir); err != nil {
		log.Fatalf("Failed to ensure migration directory exists: %v", err)
	}

	migratorAPI := api.NewMigratorAPI(cfg)
	args := flag.Args()
	if len(args) == 0 {
		fmt.Println("No command provided")
		return
	}

	switch args[0] {
	case "create":
		if len(args) < 2 {
			fmt.Println("Migration name is required")
			return
		}
		migrationName := args[1]
		migrationFileName := fmt.Sprintf("%04d_%s.%s", getNextMigrationNumber(cfg.MigrationDir), migrationName, cfg.MigrationType)
		migrationFilePath := filepath.Join(cfg.MigrationDir, migrationFileName)
		if err := migratorAPI.CreateMigration(migrationFilePath); err != nil {
			log.Fatalf("Failed to create migration: %v", err)
		}
		fmt.Printf("Migration '%s' created successfully\n", utils.GetMigrationNameFromPath(migrationFilePath))
	case "up":
		if err := migratorAPI.Up(); err != nil {
			log.Fatalf("Failed to apply migrations: %v", err)
		}
		fmt.Println("All migrations applied successfully")
	case "down":
		if err := migratorAPI.Down(); err != nil {
			log.Fatalf("Failed to rollback migration: %v", err)
		}
		fmt.Println("Latest migration rolled back successfully")
	case "redo":
		if err := migratorAPI.Redo(); err != nil {
			log.Fatalf("Failed to redo migration: %v", err)
		}
		fmt.Println("Latest migration redone successfully")
	case "status":
		status, err := migratorAPI.Status()
		if err != nil {
			log.Fatalf("Failed to get migration status: %v", err)
		}
		fmt.Println(status)
	case "dbversion":
		version, err := migratorAPI.DBVersion()
		if err != nil {
			log.Fatalf("Failed to get database version: %v", err)
		}
		fmt.Printf("Current DB version: %d\n", version)
	default:
		fmt.Println("Unknown command")
	}
}

func getNextMigrationNumber(dir string) int {
	files, err := os.ReadDir(dir)
	if err != nil {
		return 1
	}

	maxNum := 0
	for _, file := range files {
		parts := strings.Split(file.Name(), "_")
		if len(parts) > 0 {
			num := 0
			fmt.Sscanf(parts[0], "%d", &num)
			if num > maxNum {
				maxNum = num
			}
		}
	}

	return maxNum + 1
}
