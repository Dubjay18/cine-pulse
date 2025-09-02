package storage

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
)

type SQLiteStorage struct {
	db       *sql.DB
	dbPath   string
	dataPath string
}

type StorageInterface interface {
	Initialize() error
	SaveContent(content Content) error
	GetAllContent() ([]Content, error)
	GetContentByType(contentType string) ([]Content, error)
	SearchContent(title string) ([]Content, error)
	Close() error
}

func NewSQLiteStorage(dataPath string) *SQLiteStorage {
	dbPath := filepath.Join(dataPath, "cine_pulse.db")
	return &SQLiteStorage{
		dbPath:   dbPath,
		dataPath: dataPath,
	}
}

func (s *SQLiteStorage) Initialize() error {
	// Create data directory if it doesn't exist
	if err := os.MkdirAll(s.dataPath, 0755); err != nil {
		return fmt.Errorf("failed to create data directory: %v", err)
	}

	// Open database connection
	db, err := sql.Open("sqlite3", s.dbPath)
	if err != nil {
		return fmt.Errorf("failed to open database: %v", err)
	}

	s.db = db

	// Initialize and run migrations using Goose
	migrationManager := NewMigrationManager(s.db)
	if err := migrationManager.Initialize(); err != nil {
		return fmt.Errorf("failed to initialize migrations: %v", err)
	}

	if err := migrationManager.Up(); err != nil {
		return fmt.Errorf("failed to run migrations: %v", err)
	}

	log.Printf("SQLite database initialized at: %s", s.dbPath)
	return nil
}

func (s *SQLiteStorage) SaveContent(content Content) error {
	// First check if this is an existing record
	var exists bool
	err := s.db.QueryRow(`SELECT EXISTS(SELECT 1 FROM content WHERE title = ? AND type = ?)`,
		content.Title, content.Type).Scan(&exists)

	if err != nil {
		return fmt.Errorf("failed to check if content exists: %v", err)
	}

	if exists {
		// For existing records, only update fields but keep original scraped_at
		query := `
		UPDATE content
		SET year = ?, category = ?, extra_info = ?, rating = ?, source_url = ?, updated_at = CURRENT_TIMESTAMP
		WHERE title = ? AND type = ?
		`

		_, err := s.db.Exec(query, content.Year, content.Category, content.ExtraInfo,
			content.Rating, content.SourceURL, content.Title, content.Type)
		if err != nil {
			return fmt.Errorf("failed to update content: %v", err)
		}
	} else {
		// For new records, insert everything including scraped_at timestamp
		query := `
		INSERT INTO content (title, year, category, extra_info, type, rating, source_url, 
			scraped_at, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
		`

		_, err := s.db.Exec(query, content.Title, content.Year, content.Category, content.ExtraInfo,
			content.Type, content.Rating, content.SourceURL)
		if err != nil {
			return fmt.Errorf("failed to insert content: %v", err)
		}
	}

	return nil
}

func (s *SQLiteStorage) GetAllContent() ([]Content, error) {
	query := `
	SELECT title, year, category, extra_info, type, rating, source_url
	FROM content
	ORDER BY created_at DESC
	`

	rows, err := s.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query content: %v", err)
	}
	defer rows.Close()

	var contents []Content
	for rows.Next() {
		var content Content
		err := rows.Scan(&content.Title, &content.Year, &content.Category, &content.ExtraInfo, &content.Type, &content.Rating, &content.SourceURL)
		if err != nil {
			return nil, fmt.Errorf("failed to scan content: %v", err)
		}
		contents = append(contents, content)
	}

	return contents, nil
}

func (s *SQLiteStorage) GetContentByType(contentType string) ([]Content, error) {
	query := `
	SELECT title, year, category, extra_info, type, rating, source_url
	FROM content
	WHERE type = ?
	ORDER BY created_at DESC
	`

	rows, err := s.db.Query(query, contentType)
	if err != nil {
		return nil, fmt.Errorf("failed to query content by type: %v", err)
	}
	defer rows.Close()

	var contents []Content
	for rows.Next() {
		var content Content
		err := rows.Scan(&content.Title, &content.Year, &content.Category, &content.ExtraInfo, &content.Type, &content.Rating, &content.SourceURL)
		if err != nil {
			return nil, fmt.Errorf("failed to scan content: %v", err)
		}
		contents = append(contents, content)
	}

	return contents, nil
}

func (s *SQLiteStorage) SearchContent(title string) ([]Content, error) {
	query := `
	SELECT title, year, category, extra_info, type, rating, source_url
	FROM content
	WHERE title LIKE ?
	ORDER BY created_at DESC
	`

	rows, err := s.db.Query(query, "%"+title+"%")
	if err != nil {
		return nil, fmt.Errorf("failed to search content: %v", err)
	}
	defer rows.Close()

	var contents []Content
	for rows.Next() {
		var content Content
		err := rows.Scan(&content.Title, &content.Year, &content.Category, &content.ExtraInfo, &content.Type, &content.Rating, &content.SourceURL)
		if err != nil {
			return nil, fmt.Errorf("failed to scan content: %v", err)
		}
		contents = append(contents, content)
	}

	return contents, nil
}

func (s *SQLiteStorage) Close() error {
	if s.db != nil {
		return s.db.Close()
	}
	return nil
}

func (s *SQLiteStorage) GetDB() (*sql.DB, error) {
	if s.db == nil {
		// Open database connection if not already open
		db, err := sql.Open("sqlite3", s.dbPath)
		if err != nil {
			return nil, fmt.Errorf("failed to open database: %v", err)
		}
		s.db = db
	}
	return s.db, nil
}

func (s *SQLiteStorage) GetStats() (map[string]int, error) {
	stats := make(map[string]int)

	// Total content
	var total int
	err := s.db.QueryRow("SELECT COUNT(*) FROM content").Scan(&total)
	if err != nil {
		return nil, fmt.Errorf("failed to get total count: %v", err)
	}
	stats["total"] = total

	// Movies count
	var movies int
	err = s.db.QueryRow("SELECT COUNT(*) FROM content WHERE type = 'movie'").Scan(&movies)
	if err != nil {
		return nil, fmt.Errorf("failed to get movies count: %v", err)
	}
	stats["movies"] = movies

	// Series count
	var series int
	err = s.db.QueryRow("SELECT COUNT(*) FROM content WHERE type = 'series'").Scan(&series)
	if err != nil {
		return nil, fmt.Errorf("failed to get series count: %v", err)
	}
	stats["series"] = series

	return stats, nil
}

// Migration management methods
func (s *SQLiteStorage) GetMigrationManager() *MigrationManager {
	return NewMigrationManager(s.db)
}

func (s *SQLiteStorage) GetDatabaseVersion() (int64, error) {
	migrationManager := s.GetMigrationManager()
	if err := migrationManager.Initialize(); err != nil {
		return 0, err
	}
	return migrationManager.Version()
}

func (s *SQLiteStorage) RunMigrations() error {
	migrationManager := s.GetMigrationManager()
	if err := migrationManager.Initialize(); err != nil {
		return err
	}
	return migrationManager.Up()
}

func (s *SQLiteStorage) RollbackMigration() error {
	migrationManager := s.GetMigrationManager()
	if err := migrationManager.Initialize(); err != nil {
		return err
	}
	return migrationManager.Down()
}

func (s *SQLiteStorage) ResetDatabase() error {
	migrationManager := s.GetMigrationManager()
	if err := migrationManager.Initialize(); err != nil {
		return err
	}
	return migrationManager.Reset()
}
