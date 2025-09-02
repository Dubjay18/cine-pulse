package main

import (
	"cine-pulse/storage"
	"flag"
	"fmt"
	"log"
	"os"
)

func main() {
	var (
		dataPath = flag.String("data", "./data", "Path to database directory")
		command  = flag.String("cmd", "up", "Migration command: up, down, status, version, reset")
	)
	flag.Parse()

	// Initialize storage
	sqliteStorage := storage.NewSQLiteStorage(*dataPath)
	if err := sqliteStorage.Initialize(); err != nil {
		log.Fatalf("Failed to initialize storage: %v", err)
	}
	defer sqliteStorage.Close()

	// Execute command
	switch *command {
	case "up":
		if err := sqliteStorage.RunMigrations(); err != nil {
			log.Fatalf("Failed to run migrations: %v", err)
		}
		fmt.Println("Migrations completed successfully")

	case "down":
		if err := sqliteStorage.RollbackMigration(); err != nil {
			log.Fatalf("Failed to rollback migration: %v", err)
		}
		fmt.Println("Migration rolled back successfully")

	case "status":
		migrationManager := sqliteStorage.GetMigrationManager()
		if err := migrationManager.Initialize(); err != nil {
			log.Fatalf("Failed to initialize migration manager: %v", err)
		}
		if err := migrationManager.Status(); err != nil {
			log.Fatalf("Failed to get migration status: %v", err)
		}

	case "version":
		version, err := sqliteStorage.GetDatabaseVersion()
		if err != nil {
			log.Fatalf("Failed to get database version: %v", err)
		}
		fmt.Printf("Database version: %d\n", version)

	case "reset":
		if err := sqliteStorage.ResetDatabase(); err != nil {
			log.Fatalf("Failed to reset database: %v", err)
		}
		fmt.Println("Database reset completed successfully")

	default:
		fmt.Printf("Unknown command: %s\n", *command)
		fmt.Println("Available commands: up, down, status, version, reset")
		os.Exit(1)
	}
}
