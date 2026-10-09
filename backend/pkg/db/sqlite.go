package db

import (
	"database/sql"
	"fmt"
	"os"
	"strings"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/sqlite3"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/mattn/go-sqlite3"
)

type Database struct {
	DB *Store
}

var DBInstance Database

func InitDB() error {
	if databaseURL := os.Getenv("DATABASE_URL"); databaseURL != "" {
		store, err := openPostgres(databaseURL)
		if err == nil {
			DBInstance.DB = store
		}
		return err
	}
	if os.Getenv("VERCEL") != "" {
		return fmt.Errorf("DATABASE_URL is required on Vercel")
	}
	var err error

	// Get database path from environment variable or use default
	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "social-network.db"
	}

	separator := "?"
	if strings.Contains(dbPath, "?") {
		separator = "&"
	}
	// Driver options apply to every pooled connection, including new connections.
	raw, err := sql.Open("sqlite3", dbPath+separator+"_foreign_keys=on&_busy_timeout=5000")
	if err != nil {
		return fmt.Errorf("error opening database: %v", err)
	}

	DBInstance.DB = &Store{DB: raw}
	err = DBInstance.DB.Ping()
	if err != nil {
		return fmt.Errorf("error pinging database: %v", err)
	}

	_, err = DBInstance.DB.Exec("PRAGMA foreign_keys = ON;")
	if err != nil {
		return fmt.Errorf("error enabling foreign keys: %v", err)
	}

	// Run migrations
	err = RunMigrations(DBInstance.DB.DB)
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
