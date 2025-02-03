package api

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/Dendyator/SQL-migration/internal/config"
	"github.com/Dendyator/SQL-migration/internal/migration"
	"github.com/Dendyator/SQL-migration/internal/store"
)

type MigratorAPI struct {
	cfg   *config.Config
	store *store.Store
}

func NewMigratorAPI(cfg *config.Config) *MigratorAPI {
	store, err := store.NewStore(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize store: %v", err)
	}
	return &MigratorAPI{cfg: cfg, store: store}
}

func (a *MigratorAPI) CreateMigration(name string) error {
	ext := ".sql"
	if a.cfg.MigrationType == "go" {
		ext = ".go"
	}
	filename := fmt.Sprintf("%s/%04d_%s%s", a.cfg.MigrationDir, getNextMigrationNumber(a.cfg.MigrationDir), name, ext)
	f, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create migration file: %w", err)
	}
	defer f.Close()

	if a.cfg.MigrationType == "sql" {
		if _, err := f.WriteString("-- +migrate Up\n\n-- +migrate Down\n"); err != nil {
			return fmt.Errorf("failed to write SQL migration template: %w", err)
		}
	} else {
		if _, err := f.WriteString("package migrations\n\n// Up_migration implements the up step of the migration.\nfunc Up_migration(o interface{}) {\n}\n\n// Down_migration implements the down step of the migration.\nfunc Down_migration(o interface{}) {\n}\n"); err != nil {
			return fmt.Errorf("failed to write Go migration template: %w", err)
		}
	}

	return nil
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

func (a *MigratorAPI) Up() error {
	migrations, err := migration.LoadMigrations(a.cfg)
	if err != nil {
		return fmt.Errorf("failed to load migrations: %w", err)
	}

	appliedMigrations, err := a.store.GetAppliedMigrations()
	if err != nil {
		return fmt.Errorf("failed to get applied migrations: %w", err)
	}

	for _, migration := range migrations {
		if contains(appliedMigrations, migration.Name) {
			continue
		}

		if err := migration.Apply(); err != nil {
			return fmt.Errorf("failed to apply migration %s: %w", migration.Name, err)
		}

		if err := a.store.ApplyMigration(migration.Name); err != nil {
			return fmt.Errorf("failed to mark migration as applied: %w", err)
		}
	}

	return nil
}

func (a *MigratorAPI) Down() error {
	latestMigration, err := a.store.GetLatestMigration()
	if err != nil {
		return fmt.Errorf("failed to get latest migration: %w", err)
	}
	if latestMigration == "" {
		return fmt.Errorf("no migrations to rollback")
	}

	migrations, err := migration.LoadMigrations(a.cfg)
	if err != nil {
		return fmt.Errorf("failed to load migrations: %w", err)
	}

	for _, migration := range migrations {
		if migration.Name == latestMigration {
			if err := migration.Rollback(); err != nil {
				return fmt.Errorf("failed to rollback migration %s: %w", migration.Name, err)
			}

			if err := a.store.RollbackMigration(migration.Name); err != nil {
				return fmt.Errorf("failed to mark migration as rolled back: %w", err)
			}
			break
		}
	}

	return nil
}

func (a *MigratorAPI) Redo() error {
	if err := a.Down(); err != nil {
		return err
	}
	if err := a.Up(); err != nil {
		return err
	}
	return nil
}

func (a *MigratorAPI) Status() (string, error) {
	appliedMigrations, err := a.store.GetAppliedMigrations()
	if err != nil {
		return "", fmt.Errorf("failed to get applied migrations: %w", err)
	}

	var status []string
	for _, migration := range appliedMigrations {
		status = append(status, fmt.Sprintf("Applied: %s", migration))
	}

	return strings.Join(status, "\n"), nil
}

func (a *MigratorAPI) DBVersion() (int, error) {
	latestMigration, err := a.store.GetLatestMigration()
	if err != nil {
		return 0, fmt.Errorf("failed to get latest migration: %w", err)
	}
	if latestMigration == "" {
		return 0, nil
	}

	parts := strings.Split(latestMigration, "_")
	if len(parts) == 0 {
		return 0, fmt.Errorf("invalid migration name format")
	}

	var version int
	fmt.Sscanf(parts[0], "%d", &version)
	return version, nil
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
