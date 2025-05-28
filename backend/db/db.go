package db

import (
	"errors"
	"fmt"
	"os"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

var DB Database

type Database struct {
	*sqlx.DB
}

func NewDB() error {
	connStr := os.Getenv("DATABASE_URL")
	var err error
	DB.DB, err = sqlx.Connect("postgres", connStr)
	if err != nil {
		return fmt.Errorf("opening new DB connection: %w", err)
	}

	return nil
}

func (db *Database) MigrateUp() error {
	filePath := "file:///app/migrations"
	connStr := os.Getenv("DATABASE_URL")

	m, err := migrate.New(filePath, connStr)
	if err != nil {
		return fmt.Errorf("create migration err: %w", err)
	}

	err = m.Up()
	if err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			return nil
		}

		return fmt.Errorf("migration up err: %w", err)
	}

	return nil
}
