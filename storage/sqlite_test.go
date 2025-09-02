package storage

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSQLiteStorage(t *testing.T) {
	// Create temporary directory for test
	tempDir := t.TempDir()

	// Initialize storage
	storage := NewSQLiteStorage(tempDir)
	err := storage.Initialize()
	if err != nil {
		t.Fatalf("Failed to initialize storage: %v", err)
	}
	defer storage.Close()

	// Test saving content
	testContent := Content{
		Title:     "Test Movie",
		Year:      &[]int{2023}[0], // Pointer to int
		Category:  "Hollywood",
		ExtraInfo: "Test movie description",
		Type:      "movie",
	}

	err = storage.SaveContent(testContent)
	if err != nil {
		t.Fatalf("Failed to save content: %v", err)
	}

	// Test retrieving all content
	contents, err := storage.GetAllContent()
	if err != nil {
		t.Fatalf("Failed to get all content: %v", err)
	}

	if len(contents) != 1 {
		t.Fatalf("Expected 1 content, got %d", len(contents))
	}

	if contents[0].Title != testContent.Title {
		t.Errorf("Expected title %s, got %s", testContent.Title, contents[0].Title)
	}

	// Test retrieving content by type
	movies, err := storage.GetContentByType("movie")
	if err != nil {
		t.Fatalf("Failed to get movies: %v", err)
	}

	if len(movies) != 1 {
		t.Fatalf("Expected 1 movie, got %d", len(movies))
	}

	// Test search
	searchResults, err := storage.SearchContent("Test")
	if err != nil {
		t.Fatalf("Failed to search content: %v", err)
	}

	if len(searchResults) != 1 {
		t.Fatalf("Expected 1 search result, got %d", len(searchResults))
	}

	// Test stats
	stats, err := storage.GetStats()
	if err != nil {
		t.Fatalf("Failed to get stats: %v", err)
	}

	if stats["total"] != 1 {
		t.Errorf("Expected total 1, got %d", stats["total"])
	}

	if stats["movies"] != 1 {
		t.Errorf("Expected movies 1, got %d", stats["movies"])
	}

	if stats["series"] != 0 {
		t.Errorf("Expected series 0, got %d", stats["series"])
	}
}

func TestSQLiteStorageInit(t *testing.T) {
	tempDir := t.TempDir()

	storage := NewSQLiteStorage(tempDir)
	err := storage.Initialize()
	if err != nil {
		t.Fatalf("Failed to initialize storage: %v", err)
	}
	defer storage.Close()

	// Check if database file was created
	dbPath := filepath.Join(tempDir, "cine_pulse.db")
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		t.Fatalf("Database file was not created")
	}
}
