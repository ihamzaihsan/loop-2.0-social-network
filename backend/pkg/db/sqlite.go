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
	DB *sql.DB
}

var DBInstance Database

func InitDB() error {
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
	DBInstance.DB, err = sql.Open("sqlite3", dbPath+separator+"_foreign_keys=on&_busy_timeout=5000")
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

	// Check if comments table has image column
	var hasImageColumn bool
	err = DBInstance.DB.QueryRow(`
		SELECT COUNT(*) > 0
		FROM pragma_table_info('comments')
		WHERE name = 'image'
	`).Scan(&hasImageColumn)

	if err == nil && !hasImageColumn {
		// Add image column if it doesn't exist
		fmt.Println("Adding image column to comments table...")
		_, err = DBInstance.DB.Exec(`ALTER TABLE comments ADD COLUMN image TEXT`)
		if err != nil {
			fmt.Printf("Error adding image column: %v\n", err)
			// Continue anyway, as migrations might handle this
		} else {
			fmt.Println("Image column added successfully")
		}
	}

	// Check if users table has isprivate column
	var hasIsPrivateColumn bool
	err = DBInstance.DB.QueryRow(`
		SELECT COUNT(*) > 0
		FROM pragma_table_info('users')
		WHERE name = 'isprivate'
	`).Scan(&hasIsPrivateColumn)

	if err == nil && !hasIsPrivateColumn {
		// Add isprivate column if it doesn't exist
		fmt.Println("Adding isprivate column to users table...")
		_, err = DBInstance.DB.Exec(`ALTER TABLE users ADD COLUMN isprivate BOOLEAN DEFAULT 0`)
		if err != nil {
			fmt.Printf("Error adding isprivate column: %v\n", err)
			// Continue anyway, as migrations might handle this
		} else {
			fmt.Println("IsPrivate column added successfully")
		}
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
