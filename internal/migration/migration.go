package migration

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/Dendyator/SQL-migration/internal/config"
	"github.com/Dendyator/SQL-migration/internal/store"
	"github.com/Dendyator/SQL-migration/pkg/migrations"
)

type Migration struct {
	Name string
	Up   func() error
	Down func() error
}

func NewMigration(name string, up, down func() error) *Migration {
	return &Migration{
		Name: name,
		Up:   up,
		Down: down,
	}
}

func (m *Migration) Apply() error {
	if err := m.Up(); err != nil {
		return fmt.Errorf("failed to apply migration %s: %w", m.Name, err)
	}
	return nil
}

func (m *Migration) Rollback() error {
	if err := m.Down(); err != nil {
		return fmt.Errorf("failed to rollback migration %s: %w", m.Name, err)
	}
	return nil
}

// LoadMigrations загружает все миграции из указанного каталога.
func LoadMigrations(cfg *config.Config) ([]*Migration, error) {
	return migrations.LoadMigrationsFromDir(cfg)
}

// ParseGoMigration парсит Go-файл миграции и возвращает объект миграции.
func ParseGoMigration(data []byte, name string) (*Migration, error) {
	log.Println("Go migrations are not implemented yet.")
	// Здесь можно добавить логику для парсинга Go-миграций.
	return nil, fmt.Errorf("Go migrations are not implemented yet")
}

// GetStore возвращает экземпляр хранилища.
func GetStore() *store.Store {
	return store.GetStore()
}

// ParseMigrationFile парсит файл миграции и возвращает объект миграции.
func ParseMigrationFile(path, migrationType string) (*Migration, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read migration file: %w", err)
	}

	name := filepath.Base(path)
	switch migrationType {
	case "sql":
		return parseSQLMigration(data, name)
	case "go":
		return ParseGoMigration(data, name)
	default:
		return nil, fmt.Errorf("unsupported migration type: %s", migrationType)
	}
}
