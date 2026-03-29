package repo

import (
	"database/sql"
	"fmt"

	"github.com/pressly/goose/v3"
)

func RunMigrations(db *sql.DB, dir string) error {
	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("set goose dialect %w", err)
	}

	if err := goose.Up(db, dir); err != nil {
		return fmt.Errorf("run goose migrations %w", err)
	}

	return nil
}
