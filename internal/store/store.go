package store

import (
	"database/sql"
	"fmt"

	"github.com/Dendyator/SQL-migration/internal/config"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

var instance *Store

func GetStore() *Store {
	return instance
}

func NewStore(cfg *config.Config) (*Store, error) {
	db, err := sqlx.Connect("postgres", cfg.DSN)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	if _, err := db.Exec(`CREATE TABLE IF NOT EXISTS migrations (
        id SERIAL PRIMARY KEY,
        name VARCHAR(255) NOT NULL UNIQUE,
        applied_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
    )`); err != nil {
		return nil, fmt.Errorf("failed to create migrations table: %w", err)
	}

	instance = &Store{db: db}
	return instance, nil
}

func (s *Store) Exec(query string, args ...interface{}) (sql.Result, error) {
	return s.db.Exec(query, args...)
}

type Store struct {
	db *sqlx.DB
}

func (s *Store) ApplyMigration(name string) error {
	tx, err := s.db.Beginx()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	_, err = tx.Exec("SELECT pg_advisory_lock(1)")
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to acquire lock: %w", err)
	}

	_, err = tx.Exec("INSERT INTO migrations (name) VALUES ($1)", name)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to insert migration record: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (s *Store) RollbackMigration(name string) error {
	tx, err := s.db.Beginx()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	_, err = tx.Exec("SELECT pg_advisory_lock(1)")
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to acquire lock: %w", err)
	}

	_, err = tx.Exec("DELETE FROM migrations WHERE name = $1", name)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to delete migration record: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (s *Store) GetAppliedMigrations() ([]string, error) {
	var names []string
	if err := s.db.Select(&names, "SELECT name FROM migrations ORDER BY applied_at ASC"); err != nil {
		return nil, fmt.Errorf("failed to get applied migrations: %w", err)
	}
	return names, nil
}

func (s *Store) IsMigrationApplied(name string) (bool, error) {
	var count int
	if err := s.db.Get(&count, "SELECT COUNT(*) FROM migrations WHERE name = $1", name); err != nil {
		return false, fmt.Errorf("failed to check migration status: %w", err)
	}
	return count > 0, nil
}

func (s *Store) GetLatestMigration() (string, error) {
	var name string
	if err := s.db.Get(&name, "SELECT name FROM migrations ORDER BY applied_at DESC LIMIT 1"); err != nil {
		if err == sql.ErrNoRows {
			return "", nil
		}
		return "", fmt.Errorf("failed to get latest migration: %w", err)
	}
	return name, nil
}
