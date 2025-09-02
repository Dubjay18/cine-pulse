package storage

import (
	"database/sql"
	"testing"
)

func TestMigrations(t *testing.T) {
	// Create temporary directory for test
	tempDir := t.TempDir()

	// Initialize storage
	storage := NewSQLiteStorage(tempDir)
	err := storage.Initialize()
	if err != nil {
		t.Fatalf("Failed to initialize storage: %v", err)
	}
	defer storage.Close()

	// Test getting database version
	version, err := storage.GetDatabaseVersion()
	if err != nil {
		t.Fatalf("Failed to get database version: %v", err)
	}

	if version < 1 {
		t.Errorf("Expected database version >= 1, got %d", version)
	}

	// Test that migrations created the content table
	db, err := storage.GetDB()
	if err != nil {
		t.Fatalf("Failed to get database: %v", err)
	}

	// Check if content table exists
	var tableName string
	err = db.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name='content'").Scan(&tableName)
	if err != nil {
		t.Fatalf("Content table was not created: %v", err)
	}

	if tableName != "content" {
		t.Errorf("Expected table name 'content', got '%s'", tableName)
	}

	// Test running migrations again (should be idempotent)
	err = storage.RunMigrations()
	if err != nil {
		t.Fatalf("Failed to run migrations again: %v", err)
	}

	// Version should be the same or higher
	newVersion, err := storage.GetDatabaseVersion()
	if err != nil {
		t.Fatalf("Failed to get database version after re-running migrations: %v", err)
	}

	if newVersion < version {
		t.Errorf("Database version went backwards: %d -> %d", version, newVersion)
	}
}

func TestMigrationManager(t *testing.T) {
	// Create in-memory database for testing
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open in-memory database: %v", err)
	}
	defer db.Close()

	// Create migration manager
	migrationManager := NewMigrationManager(db)
	err = migrationManager.Initialize()
	if err != nil {
		t.Fatalf("Failed to initialize migration manager: %v", err)
	}

	// Test getting version before any migrations
	version, err := migrationManager.Version()
	if err != nil {
		t.Fatalf("Failed to get initial version: %v", err)
	}

	if version != 0 {
		t.Errorf("Expected initial version 0, got %d", version)
	}

	// Run migrations
	err = migrationManager.Up()
	if err != nil {
		t.Fatalf("Failed to run migrations: %v", err)
	}

	// Check version after migrations
	version, err = migrationManager.Version()
	if err != nil {
		t.Fatalf("Failed to get version after migrations: %v", err)
	}

	if version < 1 {
		t.Errorf("Expected version >= 1 after migrations, got %d", version)
	}

	// Test rollback
	err = migrationManager.Down()
	if err != nil {
		t.Fatalf("Failed to rollback migration: %v", err)
	}

	// Version should be decremented
	newVersion, err := migrationManager.Version()
	if err != nil {
		t.Fatalf("Failed to get version after rollback: %v", err)
	}

	if newVersion >= version {
		t.Errorf("Expected version to decrease after rollback: %d -> %d", version, newVersion)
	}
}
