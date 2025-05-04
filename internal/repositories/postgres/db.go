// internal/repositories/postgres/db.go
package postgres

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/dukerupert/walking-drum/internal/config"
	_ "github.com/lib/pq" // PostgreSQL driver
)

// DB represents a PostgreSQL database connection pool
type DB struct {
	*sql.DB
}

// Connect establishes a connection to the PostgreSQL database
func Connect(dbConfig config.DBConfig) (*DB, error) {
	// Create the database connection
	db, err := sql.Open("postgres", dbConfig.DSN)
	if err != nil {
		return nil, fmt.Errorf("failed to open database connection: %w", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)
	db.SetConnMaxLifetime(5 * time.Minute)

	// Verify the connection is working
	if err = db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	log.Printf("Connected to PostgreSQL database: %s", dbConfig.Name)

	return &DB{db}, nil
}

// Close closes the database connection pool
func (db *DB) Close() error {
	return db.DB.Close()
}

// Transaction executes a function within a database transaction
func (db *DB) Transaction(txFunc func(*sql.Tx) error) error {
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer func() {
		if p := recover(); p != nil {
			// A panic occurred, rollback the transaction
			tx.Rollback()
			panic(p) // Re-throw panic after rollback
		} else if err != nil {
			// An error occurred, rollback the transaction
			tx.Rollback()
		} else {
			// All good, commit the transaction
			err = tx.Commit()
		}
	}()

	err = txFunc(tx)
	return err
}
