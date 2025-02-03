package config

import (
	"errors"
	"os"
)

type Config struct {
	DSN           string
	MigrationDir  string
	MigrationType string
}

func (c *Config) Validate() error {
	if c.DSN == "" {
		c.DSN = os.Getenv("DB_DSN")
	}
	if c.DSN == "" {
		return errors.New("database connection string is required")
	}
	if c.MigrationDir == "" {
		return errors.New("migration directory is required")
	}
	if c.MigrationType != "go" && c.MigrationType != "sql" {
		return errors.New("invalid migration type")
	}
	return nil
}
