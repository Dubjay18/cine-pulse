package storage

import (
	"database/sql"
	"embed"
	"fmt"
	"log"

	"github.com/pressly/goose/v3"
)

//go:embed migrations/*.sql
var embedMigrations embed.FS

type MigrationManager struct {
	db *sql.DB
}

func NewMigrationManager(db *sql.DB) *MigrationManager {
	return &MigrationManager{db: db}
}

func (m *MigrationManager) Initialize() error {
	// Set the base filesystem for migrations
	goose.SetBaseFS(embedMigrations)

	// Set the dialect to sqlite3
	if err := goose.SetDialect("sqlite3"); err != nil {
		return fmt.Errorf("failed to set goose dialect: %v", err)
	}

	return nil
}

func (m *MigrationManager) Up() error {
	if err := goose.Up(m.db, "migrations"); err != nil {
		return fmt.Errorf("failed to run migrations: %v", err)
	}
	log.Println("Database migrations completed successfully")
	return nil
}

func (m *MigrationManager) Down() error {
	if err := goose.Down(m.db, "migrations"); err != nil {
		return fmt.Errorf("failed to rollback migration: %v", err)
	}
	log.Println("Database migration rolled back successfully")
	return nil
}

func (m *MigrationManager) Status() error {
	if err := goose.Status(m.db, "migrations"); err != nil {
		return fmt.Errorf("failed to get migration status: %v", err)
	}
	return nil
}

func (m *MigrationManager) Version() (int64, error) {
	version, err := goose.GetDBVersion(m.db)
	if err != nil {
		return 0, fmt.Errorf("failed to get database version: %v", err)
	}
	return version, nil
}

func (m *MigrationManager) Reset() error {
	if err := goose.Reset(m.db, "migrations"); err != nil {
		return fmt.Errorf("failed to reset database: %v", err)
	}
	log.Println("Database reset completed successfully")
	return nil
}
