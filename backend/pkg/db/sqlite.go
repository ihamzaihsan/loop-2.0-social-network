package db

import (
	"database/sql"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/sqlite3"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/mattn/go-sqlite3"
)

type Database struct {
	DB *sql.DB
}

var DBInstance Database

func InitDB() error {
	var err error
	DBInstance.DB, err = sql.Open("sqlite3", "social-network.db")
	if err != nil {
		return fmt.Errorf("error opening database: %v", err)
	}

	err = DBInstance.DB.Ping()
	if err != nil {
		return fmt.Errorf("error pinging database: %v", err)
	}

	_, err = DBInstance.DB.Exec("PRAGMA foreign_keys = ON;")
	if err != nil {
		return fmt.Errorf("error enabling foreign keys: %v", err)
	}

	// Run migrations
	err = RunMigrations(DBInstance.DB)
	if err != nil {
		return fmt.Errorf("error running migrations: %v", err)
	}

	return nil
}

func RunMigrations(db *sql.DB) error {
	driver, err := sqlite3.WithInstance(db, &sqlite3.Config{})
	if err != nil {
		return fmt.Errorf("could not create driver: %v", err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://pkg/db/migrations/sqlite",
		"sqlite3",
		driver,
	)
	if err != nil {
		return fmt.Errorf("could not create migration instance: %v", err)
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("could not run migrations: %v", err)
	}

	return nil
}

// CloseDB closes the database connection
func CloseDB() error {
	if DBInstance.DB != nil {
		return DBInstance.DB.Close()
	}
	return nil
}

