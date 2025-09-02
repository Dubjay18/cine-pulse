package main

import (
	"cine-pulse/model"
	"cine-pulse/scheduler"
	"cine-pulse/scraper"
	"cine-pulse/storage"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/joho/godotenv/autoload"
)

func main() {
	// Initialize storage
	dataPath := os.Getenv("DATA_PATH")
	if dataPath == "" {
		dataPath = "./data"
	}

	// Initialize logger with timestamp
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
	log.Println("Starting Cine Pulse application...")

	// Initialize storage
	sqliteStorage := storage.NewSQLiteStorage(dataPath)
	if err := sqliteStorage.Initialize(); err != nil {
		log.Fatalf("Failed to initialize storage: %v", err)
	}
	defer sqliteStorage.Close()

	// Initialize scraper and model manager
	webScraper := scraper.NewScraper()
	modelManager := model.NewModelManager()

	// Get configuration
	runMode := os.Getenv("RUN_MODE")
	sourceURLs := getSourceURLs()

	if runMode == "scheduler" || runMode == "" {
		log.Println("Starting in scheduler mode")

		// Initialize scheduler
		sched := scheduler.NewScheduler()

		// Create content scraper job
		scraperJob := scheduler.NewContentScraperJob(webScraper, sqliteStorage, modelManager, sourceURLs)

		// Add job to run at 10am and 5pm
		if err := sched.AddMorningEveningJob(scraperJob); err != nil {
			log.Fatalf("Failed to schedule content scraper job: %v", err)
		}

		// Start the scheduler
		sched.Start()
		log.Println("Scheduler started. Content will be scraped at 10:00 AM and 5:00 PM daily")

		// Run the job once at startup if specified
		if os.Getenv("RUN_AT_STARTUP") == "true" {
			log.Println("Running initial content scrape at startup")
			if err := sched.RunJobNow(scraperJob.Name()); err != nil {
				log.Printf("Error running initial job: %v", err)
			}
		}

		// Display database stats
		displayDatabaseStats(sqliteStorage)

		// Set up signal handling for graceful shutdown
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

		log.Println("Application running. Press Ctrl+C to exit")

		// Wait for termination signal
		sig := <-quit
		log.Printf("Received signal %s, shutting down...", sig)

		// Gracefully stop the scheduler
		sched.Stop()

	} else if runMode == "once" {
		log.Println("Running in single execution mode")

		// Create the job
		scraperJob := scheduler.NewContentScraperJob(webScraper, sqliteStorage, modelManager, sourceURLs)

		// Run it once with a timeout
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
		defer cancel()

		if err := scraperJob.Run(ctx); err != nil {
			log.Fatalf("Error running job: %v", err)
		}

		// Display database stats
		displayDatabaseStats(sqliteStorage)
	}

	log.Println("Application exiting")
}

// getSourceURLs returns the source URLs to scrape from environment variables
func getSourceURLs() []string {
	// Default source
	sourceURLs := []string{"https://nkiri.com/"}

	// Check for additional sources in environment variables
	if sources := os.Getenv("SOURCE_URLS"); sources != "" {
		// Parse comma-separated list
		var additionalSources []string
		if err := json.Unmarshal([]byte(sources), &additionalSources); err != nil {
			log.Printf("Error parsing SOURCE_URLS: %v", err)
		} else {
			sourceURLs = additionalSources
		}
	}

	return sourceURLs
}

// displayDatabaseStats shows database statistics
func displayDatabaseStats(db *storage.SQLiteStorage) {
	log.Println("Database Statistics")

	// Get database stats
	stats, err := db.GetStats()
	if err != nil {
		log.Printf("Error getting database stats: %v", err)
		return
	}

	log.Printf("Total content: %d", stats["total"])
	log.Printf("Movies: %d", stats["movies"])
	log.Printf("Series: %d", stats["series"])

	// Show recent content
	allContent, err := db.GetAllContent()
	if err != nil {
		log.Printf("Error getting content: %v", err)
		return
	}

	limit := 5
	if len(allContent) < limit {
		limit = len(allContent)
	}

	log.Printf("Recent Content (last %d):", limit)
	for i := 0; i < limit; i++ {
		content := allContent[i]
		year := ""
		if content.Year != nil {
			year = fmt.Sprintf(" (%d)", *content.Year)
		}
		log.Printf("- %s%s [%s] - %s", content.Title, year, content.Type, content.Category)
	}
}
